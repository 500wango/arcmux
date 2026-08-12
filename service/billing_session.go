package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/logger"
	"github.com/500wango/arcmux/model"
	relaycommon "github.com/500wango/arcmux/relay/common"
	"github.com/500wango/arcmux/relaykit/types"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// BillingSession — 统一计费会话
// ---------------------------------------------------------------------------

// BillingSession 封装单次请求的预扣费/结算/退款生命周期。
// 实现 relaycommon.BillingSettler 接口。
type BillingSession struct {
	relayInfo        *relaycommon.RelayInfo
	funding          FundingSource
	preConsumedQuota int  // 实际预扣额度（信任用户可能为 0）
	tokenConsumed    int  // 令牌额度实际扣减量
	extraReserved    int  // 发送前补充预扣的额度（订阅退款时需要单独回滚）
	trusted          bool // 是否命中信任额度旁路
	settled          bool // Settle 全部完成（资金 + 令牌）
	refunded         bool // Refund 已调用
	refunding        bool // Refund 异步任务正在执行
	mu               sync.Mutex
}

// Settle 根据实际消耗额度进行结算。
// 资金来源和令牌额度通过同一事务提交。
func (s *BillingSession) Settle(actualQuota int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if actualQuota < 0 {
		return fmt.Errorf("actual quota cannot be negative: %d", actualQuota)
	}
	if s.settled {
		return nil
	}
	if s.refunded || s.refunding {
		return errors.New("billing session refund is in progress or completed")
	}
	delta := actualQuota - s.preConsumedQuota
	if delta == 0 {
		s.settled = true
		return nil
	}
	adjustment := model.BillingAdjustment{
		Key:      s.relayInfo.RequestId + ":settle",
		UserId:   s.relayInfo.UserId,
		TokenId:  s.relayInfo.TokenId,
		TokenKey: s.relayInfo.TokenKey,
	}
	if s.funding.Source() == BillingSourceSubscription {
		adjustment.SubscriptionId = s.relayInfo.SubscriptionId
		adjustment.SubscriptionDelta = int64(delta)
	} else {
		adjustment.WalletDelta = -delta
	}
	if !s.relayInfo.IsPlayground {
		adjustment.TokenDelta = -delta
	}
	if _, err := model.ApplyBillingAdjustment(adjustment); err != nil {
		return err
	}
	// 更新 relayInfo 上的订阅 PostDelta（用于日志）
	if s.funding.Source() == BillingSourceSubscription {
		s.relayInfo.SubscriptionPostDelta += int64(delta)
	}
	s.settled = true
	return nil
}

// Refund 退还所有预扣费，幂等安全，并在返回前确认事务结果。
func (s *BillingSession) Refund(c *gin.Context) {
	s.mu.Lock()
	if s.settled || s.refunded || s.refunding || !s.needsRefundLocked() {
		s.mu.Unlock()
		return
	}
	s.refunding = true
	// Copy state while holding the lifecycle lock. Settle and Reserve reject a
	// session once refunding begins, so the async adjustment has a stable view.
	tokenId := s.relayInfo.TokenId
	tokenKey := s.relayInfo.TokenKey
	isPlayground := s.relayInfo.IsPlayground
	tokenConsumed := s.tokenConsumed
	preConsumedQuota := s.preConsumedQuota
	extraReserved := s.extraReserved
	subscriptionId := s.relayInfo.SubscriptionId
	fundingSource := s.funding.Source()
	requestId := s.relayInfo.RequestId
	userId := s.relayInfo.UserId
	s.mu.Unlock()

	logger.LogInfo(c, fmt.Sprintf("用户 %d 请求失败, 返还预扣费（token_quota=%s, funding=%s）",
		userId,
		logger.FormatQuota(tokenConsumed),
		fundingSource,
	))

	adjustment := model.BillingAdjustment{
		Key:      requestId + ":refund",
		UserId:   userId,
		TokenId:  tokenId,
		TokenKey: tokenKey,
	}
	if fundingSource == BillingSourceSubscription {
		adjustment.SubscriptionId = subscriptionId
		adjustment.SubscriptionRefundPreConsumeId = requestId
		adjustment.SubscriptionRefundExtraQuota = -int64(extraReserved)
	} else {
		adjustment.WalletDelta = preConsumedQuota
	}
	if tokenConsumed > 0 && !isPlayground {
		adjustment.TokenDelta = tokenConsumed
	}

	err := refundWithRetry(func() error {
		_, err := model.ApplyBillingAdjustment(adjustment)
		return err
	})
	s.mu.Lock()
	s.refunding = false
	if err == nil {
		s.refunded = true
	}
	s.mu.Unlock()
	if err != nil {
		common.SysLog("error refunding billing session: " + err.Error())
	}
}

// NeedsRefund 返回是否存在需要退还的预扣状态。
func (s *BillingSession) NeedsRefund() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.needsRefundLocked()
}

func (s *BillingSession) needsRefundLocked() bool {
	if s.settled || s.refunded {
		return false
	}
	return s.preConsumedQuota > 0
}

// GetPreConsumedQuota 返回实际预扣的额度。
func (s *BillingSession) GetPreConsumedQuota() int {
	return s.preConsumedQuota
}

func (s *BillingSession) Reserve(targetQuota int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.settled || s.refunded || s.refunding {
		return errors.New("billing session is already settled or refunding")
	}
	if s.trusted || targetQuota <= s.preConsumedQuota {
		return nil
	}

	delta := targetQuota - s.preConsumedQuota
	if delta <= 0 {
		return nil
	}

	adjustment := model.BillingAdjustment{
		Key:                    fmt.Sprintf("%s:reserve:%d", s.relayInfo.RequestId, targetQuota),
		UserId:                 s.relayInfo.UserId,
		TokenId:                s.relayInfo.TokenId,
		TokenKey:               s.relayInfo.TokenKey,
		TokenDelta:             -delta,
		TokenRequireSufficient: !s.relayInfo.TokenUnlimited && !s.relayInfo.IsPlayground,
	}
	if s.relayInfo.IsPlayground {
		adjustment.TokenId = 0
		adjustment.TokenDelta = 0
	}
	if s.funding.Source() == BillingSourceSubscription {
		adjustment.SubscriptionId = s.relayInfo.SubscriptionId
		adjustment.SubscriptionDelta = int64(delta)
	} else {
		adjustment.WalletDelta = -delta
		adjustment.WalletRequireSufficient = true
	}
	if _, err := model.ApplyBillingAdjustment(adjustment); err != nil {
		return err
	}

	s.preConsumedQuota += delta
	if !s.relayInfo.IsPlayground {
		s.tokenConsumed += delta
	}
	if s.funding.Source() == BillingSourceSubscription {
		s.extraReserved += delta
	}
	s.syncRelayInfo()
	return nil
}

// ---------------------------------------------------------------------------
// PreConsume — 统一预扣费入口（含信任额度旁路）
// ---------------------------------------------------------------------------

// preConsume 执行预扣费，资金来源与令牌额度在同一事务内提交。
func (s *BillingSession) preConsume(c *gin.Context, quota int) *types.NewAPIError {
	effectiveQuota := quota

	// ---- 信任额度旁路 ----
	if s.shouldTrust(c) {
		s.trusted = true
		effectiveQuota = 0
		logger.LogInfo(c, fmt.Sprintf("用户 %d 额度充足, 信任且不需要预扣费 (funding=%s)", s.relayInfo.UserId, s.funding.Source()))
	} else if effectiveQuota > 0 {
		logger.LogInfo(c, fmt.Sprintf("用户 %d 需要预扣费 %s (funding=%s)", s.relayInfo.UserId, logger.FormatQuota(effectiveQuota), s.funding.Source()))
	}

	var err error
	switch funding := s.funding.(type) {
	case *WalletFunding:
		adjustment := model.BillingAdjustment{
			Key:                     s.relayInfo.RequestId + ":preconsume",
			UserId:                  s.relayInfo.UserId,
			WalletDelta:             -effectiveQuota,
			WalletRequireSufficient: true,
			TokenId:                 s.relayInfo.TokenId,
			TokenKey:                s.relayInfo.TokenKey,
			TokenDelta:              -effectiveQuota,
			TokenRequireSufficient:  !s.relayInfo.TokenUnlimited,
		}
		if s.relayInfo.IsPlayground {
			adjustment.TokenId = 0
			adjustment.TokenDelta = 0
		}
		_, err = model.ApplyBillingAdjustment(adjustment)
		if err == nil {
			funding.consumed = effectiveQuota
		}
	case *SubscriptionFunding:
		var result *model.SubscriptionPreConsumeResult
		if s.relayInfo.IsPlayground {
			result, err = model.PreConsumeUserSubscription(funding.requestId, funding.userId, funding.modelName, 0, funding.amount)
		} else {
			result, err = model.PreConsumeUserSubscriptionAndToken(
				funding.requestId,
				funding.userId,
				funding.modelName,
				0,
				funding.amount,
				s.relayInfo.TokenId,
				s.relayInfo.TokenKey,
				!s.relayInfo.TokenUnlimited,
			)
		}
		if err == nil {
			funding.applyPreConsumeResult(result)
		}
	default:
		err = fmt.Errorf("unsupported funding source: %s", s.funding.Source())
	}
	if err != nil {
		// TODO: model 层应定义哨兵错误（如 ErrNoActiveSubscription），用 errors.Is 替代字符串匹配
		if errors.Is(err, model.ErrInsufficientQuota) {
			return types.NewErrorWithStatusCode(err, types.ErrorCodeInsufficientUserQuota, http.StatusForbidden, types.ErrOptionWithSkipRetry(), types.ErrOptionWithNoRecordErrorLog())
		}
		errMsg := err.Error()
		if strings.Contains(errMsg, "no active subscription") || strings.Contains(errMsg, "subscription quota insufficient") {
			return types.NewErrorWithStatusCode(fmt.Errorf("订阅额度不足或未配置订阅: %s", errMsg), types.ErrorCodeInsufficientUserQuota, http.StatusForbidden, types.ErrOptionWithSkipRetry(), types.ErrOptionWithNoRecordErrorLog())
		}
		return types.NewError(err, types.ErrorCodeUpdateDataError, types.ErrOptionWithSkipRetry())
	}

	s.preConsumedQuota = effectiveQuota
	if effectiveQuota > 0 && !s.relayInfo.IsPlayground {
		s.tokenConsumed = effectiveQuota
	}

	// ---- 同步 RelayInfo 兼容字段 ----
	s.syncRelayInfo()

	return nil
}

// shouldTrust 统一信任额度检查，适用于钱包和订阅。
func (s *BillingSession) shouldTrust(c *gin.Context) bool {
	// 异步任务（ForcePreConsume=true）必须预扣全额，不允许信任旁路
	if s.relayInfo.ForcePreConsume {
		return false
	}

	trustQuota := common.GetTrustQuota()
	if trustQuota <= 0 {
		return false
	}

	// 检查令牌是否充足
	tokenTrusted := s.relayInfo.TokenUnlimited
	if !tokenTrusted {
		tokenQuota := c.GetInt("token_quota")
		tokenTrusted = tokenQuota > trustQuota
	}
	if !tokenTrusted {
		return false
	}

	switch s.funding.Source() {
	case BillingSourceWallet:
		return s.relayInfo.UserQuota > trustQuota
	case BillingSourceSubscription:
		// 订阅不能启用信任旁路。原因：
		// 1. PreConsumeUserSubscription 要求 amount>0 来创建预扣记录并锁定订阅
		// 2. SubscriptionFunding.PreConsume 忽略参数，始终用 s.amount 预扣
		// 3. 若信任旁路将 effectiveQuota 设为 0，会导致 preConsumedQuota 与实际订阅预扣不一致
		return false
	default:
		return false
	}
}

// syncRelayInfo 将 BillingSession 的状态同步到 RelayInfo 的兼容字段上。
func (s *BillingSession) syncRelayInfo() {
	info := s.relayInfo
	info.FinalPreConsumedQuota = s.preConsumedQuota
	info.BillingSource = s.funding.Source()

	if sub, ok := s.funding.(*SubscriptionFunding); ok {
		info.SubscriptionId = sub.subscriptionId
		info.SubscriptionPreConsumed = sub.preConsumed + int64(s.extraReserved)
		info.SubscriptionPostDelta = 0
		info.SubscriptionAmountTotal = sub.AmountTotal
		info.SubscriptionAmountUsedAfterPreConsume = sub.AmountUsedAfter + int64(s.extraReserved)
		info.SubscriptionPlanId = sub.PlanId
		info.SubscriptionPlanTitle = sub.PlanTitle
	} else {
		info.SubscriptionId = 0
		info.SubscriptionPreConsumed = 0
	}
}

// ---------------------------------------------------------------------------
// NewBillingSession 工厂 — 根据计费偏好创建会话并处理回退
// ---------------------------------------------------------------------------

// NewBillingSession 根据用户计费偏好创建 BillingSession，处理 subscription_first / wallet_first 的回退。
func NewBillingSession(c *gin.Context, relayInfo *relaycommon.RelayInfo, preConsumedQuota int) (*BillingSession, *types.NewAPIError) {
	if relayInfo == nil {
		return nil, types.NewError(fmt.Errorf("relayInfo is nil"), types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}

	pref := common.NormalizeBillingPreference(relayInfo.UserSetting.BillingPreference)

	// 钱包路径需要先检查用户额度
	tryWallet := func() (*BillingSession, *types.NewAPIError) {
		userQuota, err := model.GetUserQuota(relayInfo.UserId, false)
		if err != nil {
			return nil, types.NewError(err, types.ErrorCodeQueryDataError, types.ErrOptionWithSkipRetry())
		}
		if userQuota <= 0 {
			return nil, types.NewErrorWithStatusCode(
				fmt.Errorf("用户额度不足, 剩余额度: %s", logger.FormatQuota(userQuota)),
				types.ErrorCodeInsufficientUserQuota, http.StatusForbidden,
				types.ErrOptionWithSkipRetry(), types.ErrOptionWithNoRecordErrorLog())
		}
		if userQuota-preConsumedQuota < 0 {
			return nil, types.NewErrorWithStatusCode(
				fmt.Errorf("预扣费额度失败, 用户剩余额度: %s, 需要预扣费额度: %s", logger.FormatQuota(userQuota), logger.FormatQuota(preConsumedQuota)),
				types.ErrorCodeInsufficientUserQuota, http.StatusForbidden,
				types.ErrOptionWithSkipRetry(), types.ErrOptionWithNoRecordErrorLog())
		}
		relayInfo.UserQuota = userQuota

		session := &BillingSession{
			relayInfo: relayInfo,
			funding:   &WalletFunding{userId: relayInfo.UserId},
		}
		if apiErr := session.preConsume(c, preConsumedQuota); apiErr != nil {
			return nil, apiErr
		}
		return session, nil
	}

	trySubscription := func() (*BillingSession, *types.NewAPIError) {
		subConsume := int64(preConsumedQuota)
		if subConsume <= 0 {
			subConsume = 1
		}
		session := &BillingSession{
			relayInfo: relayInfo,
			funding: &SubscriptionFunding{
				requestId: relayInfo.RequestId,
				userId:    relayInfo.UserId,
				modelName: relayInfo.OriginModelName,
				amount:    subConsume,
			},
		}
		// 必须传 subConsume 而非 preConsumedQuota，保证 SubscriptionFunding.amount、
		// preConsume 参数和 FinalPreConsumedQuota 三者一致，避免订阅多扣费。
		if apiErr := session.preConsume(c, int(subConsume)); apiErr != nil {
			return nil, apiErr
		}
		return session, nil
	}

	switch pref {
	case "subscription_only":
		return trySubscription()
	case "wallet_only":
		return tryWallet()
	case "wallet_first":
		session, err := tryWallet()
		if err != nil {
			if err.GetErrorCode() == types.ErrorCodeInsufficientUserQuota {
				return trySubscription()
			}
			return nil, err
		}
		return session, nil
	case "subscription_first":
		fallthrough
	default:
		hasSub, subCheckErr := model.HasActiveUserSubscription(relayInfo.UserId)
		if subCheckErr != nil {
			return nil, types.NewError(subCheckErr, types.ErrorCodeQueryDataError, types.ErrOptionWithSkipRetry())
		}
		if !hasSub {
			return tryWallet()
		}
		session, apiErr := trySubscription()
		if apiErr != nil {
			if apiErr.GetErrorCode() == types.ErrorCodeInsufficientUserQuota {
				// 仅当用户的活跃订阅允许钱包回退时才回退到钱包，否则返回订阅额度不足错误
				allowOverflow, overflowErr := model.UserActiveSubscriptionsAllowWalletOverflow(relayInfo.UserId)
				if overflowErr != nil {
					return nil, types.NewError(overflowErr, types.ErrorCodeQueryDataError, types.ErrOptionWithSkipRetry())
				}
				if allowOverflow {
					return tryWallet()
				}
				return nil, apiErr
			}
			return nil, apiErr
		}
		return session, nil
	}
}
