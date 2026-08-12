package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/constant"
	"github.com/500wango/arcmux/logger"
	"github.com/500wango/arcmux/model"
	relaycommon "github.com/500wango/arcmux/relay/common"
	"github.com/500wango/arcmux/setting/ratio_setting"
	"github.com/500wango/arcmux/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LogTaskConsumption 记录已成功结算的任务消费日志和统计信息。
func LogTaskConsumption(c *gin.Context, info *relaycommon.RelayInfo, settledQuota int) {
	tokenName := c.GetString("token_name")
	logContent := fmt.Sprintf("操作 %s", info.Action)
	// 支持任务仅按次计费
	if common.StringsContains(constant.TaskPricePatches, info.OriginModelName) {
		logContent = fmt.Sprintf("%s，按次计费", logContent)
	} else {
		if otherRatios := info.PriceData.OtherRatios(); len(otherRatios) > 0 {
			var contents []string
			for key, ra := range otherRatios {
				if 1.0 != ra {
					contents = append(contents, fmt.Sprintf("%s: %.2f", key, ra))
				}
			}
			if len(contents) > 0 {
				logContent = fmt.Sprintf("%s, 计算参数：%s", logContent, strings.Join(contents, ", "))
			}
		}
	}
	other := make(map[string]interface{})
	other["is_task"] = true
	other["request_path"] = c.Request.URL.Path
	other["model_price"] = info.PriceData.ModelPrice
	if info.PriceData.ModelRatio > 0 {
		other["model_ratio"] = info.PriceData.ModelRatio
	}
	other["group_ratio"] = info.PriceData.GroupRatioInfo.GroupRatio
	if info.PriceData.GroupRatioInfo.HasSpecialRatio {
		other["user_group_ratio"] = info.PriceData.GroupRatioInfo.GroupSpecialRatio
	}
	if info.IsModelMapped {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = info.UpstreamModelName
	}
	attachQuotaSaturation(c, info, other)
	model.RecordConsumeLog(c, info.UserId, model.RecordConsumeLogParams{
		ChannelId: info.ChannelId,
		ModelName: info.OriginModelName,
		TokenName: tokenName,
		Quota:     settledQuota,
		Content:   logContent,
		TokenId:   info.TokenId,
		Group:     info.UsingGroup,
		Other:     other,
	})
	model.UpdateUserUsedQuotaAndRequestCount(info.UserId, settledQuota)
	model.UpdateChannelUsedQuota(info.ChannelId, settledQuota)
}

// ---------------------------------------------------------------------------
// 异步任务计费辅助函数
// ---------------------------------------------------------------------------

// taskIsSubscription 判断任务是否通过订阅计费。
func taskIsSubscription(task *model.Task) bool {
	return task.PrivateData.BillingSource == BillingSourceSubscription && task.PrivateData.SubscriptionId > 0
}

// taskToken returns the current token key needed to keep its cache in sync.
// A deleted token no longer has a balance to adjust and must not block the
// wallet or subscription side of a task refund.
func taskToken(ctx context.Context, task *model.Task, delta int) (string, bool, error) {
	if task.PrivateData.TokenId <= 0 || delta == 0 {
		return "", false, nil
	}
	token, err := model.GetTokenById(task.PrivateData.TokenId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LogWarn(ctx, fmt.Sprintf("令牌已删除，跳过任务令牌额度调整 (tokenId=%d, task=%s)", task.PrivateData.TokenId, task.TaskID))
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get token %d for task %s: %w", task.PrivateData.TokenId, task.TaskID, err)
	}
	return token.Key, true, nil
}

// taskBillingOther 从 task 的 BillingContext 构建日志 Other 字段。
func taskBillingOther(task *model.Task) map[string]interface{} {
	other := make(map[string]interface{})
	if bc := task.PrivateData.BillingContext; bc != nil {
		other["model_price"] = bc.ModelPrice
		if bc.ModelRatio > 0 {
			other["model_ratio"] = bc.ModelRatio
		}
		other["group_ratio"] = bc.GroupRatio
		if priceData := taskBillingContextPriceData(bc); priceData != nil {
			for k, v := range priceData.OtherRatios() {
				other[k] = v
			}
		}
	}
	props := task.Properties
	if props.UpstreamModelName != "" && props.UpstreamModelName != props.OriginModelName {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = props.UpstreamModelName
	}
	return other
}

func taskBillingContextPriceData(bc *model.TaskBillingContext) *types.PriceData {
	if bc == nil || len(bc.OtherRatios) == 0 {
		return nil
	}
	priceData := &types.PriceData{}
	if !priceData.ReplaceOtherRatios(bc.OtherRatios) {
		return nil
	}
	return priceData
}

// taskModelName 从 BillingContext 或 Properties 中获取模型名称。
func taskModelName(task *model.Task) string {
	if bc := task.PrivateData.BillingContext; bc != nil && bc.OriginModelName != "" {
		return bc.OriginModelName
	}
	return task.Properties.OriginModelName
}

// RefundTaskQuota 统一的任务失败退款逻辑。
// 当异步任务失败时，将预扣的 quota 退还给用户（支持钱包和订阅），并退还令牌额度。
// 返回资金来源是否已成功退还；失败时保留 quota，供显式重试或人工对账。
func RefundTaskQuota(ctx context.Context, task *model.Task, reason string) bool {
	quota := task.Quota
	if quota == 0 {
		return true
	}

	tokenKey, tokenExists, err := taskToken(ctx, task, -quota)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("退还令牌额度准备失败 task %s: %s", task.TaskID, err.Error()))
		return false
	}
	adjustment := model.BillingAdjustment{
		Key:          fmt.Sprintf("task:%d:%s:refund", task.ID, task.TaskID),
		UserId:       task.UserId,
		TokenId:      task.PrivateData.TokenId,
		TokenKey:     tokenKey,
		TaskId:       task.ID,
		NewTaskQuota: 0,
	}
	if tokenExists {
		adjustment.TokenDelta = quota
	} else {
		adjustment.TokenId = 0
	}
	if task.ID > 0 {
		adjustment.ExpectedTaskQuota = &quota
	}
	if taskIsSubscription(task) {
		adjustment.SubscriptionId = task.PrivateData.SubscriptionId
		adjustment.SubscriptionDelta = -int64(quota)
	} else {
		adjustment.WalletDelta = quota
	}
	applied, err := model.ApplyBillingAdjustment(adjustment)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("退还任务额度失败 task %s: %s", task.TaskID, err.Error()))
		return false
	}
	task.Quota = 0
	if !applied {
		return true
	}

	// 3. 记录日志
	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["reason"] = reason
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:    task.UserId,
		LogType:   model.LogTypeRefund,
		Content:   "",
		ChannelId: task.ChannelId,
		ModelName: taskModelName(task),
		Quota:     quota,
		TokenId:   task.PrivateData.TokenId,
		Group:     task.Group,
		Other:     other,
	})

	return true
}

// RecalculateTaskQuota 通用的异步差额结算。
// actualQuota 是任务完成后的实际应扣额度，与预扣额度 (task.Quota) 做差额结算。
// reason 用于日志记录（例如 "token重算" 或 "adaptor调整"）。
// clamps 可选：若计算 actualQuota 时发生额度饱和，将其记入日志 admin_info（仅管理员可见）。
func RecalculateTaskQuota(ctx context.Context, task *model.Task, actualQuota int, reason string, clamps ...*common.QuotaClamp) {
	if actualQuota < 0 {
		logger.LogError(ctx, fmt.Sprintf("拒绝负数任务结算 task %s: actualQuota=%d", task.TaskID, actualQuota))
		return
	}
	preConsumedQuota := task.Quota
	quotaDelta := actualQuota - preConsumedQuota

	if quotaDelta == 0 {
		logger.LogInfo(ctx, fmt.Sprintf("任务 %s 预扣费准确（%s，%s）",
			task.TaskID, logger.LogQuota(actualQuota), reason))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("任务 %s 差额结算：delta=%s（实际：%s，预扣：%s，%s）",
		task.TaskID,
		logger.LogQuota(quotaDelta),
		logger.LogQuota(actualQuota),
		logger.LogQuota(preConsumedQuota),
		reason,
	))

	tokenKey, tokenExists, err := taskToken(ctx, task, quotaDelta)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("差额结算令牌准备失败 task %s: %s", task.TaskID, err.Error()))
		return
	}
	adjustment := model.BillingAdjustment{
		Key:          fmt.Sprintf("task:%d:%s:settle:%d", task.ID, task.TaskID, actualQuota),
		UserId:       task.UserId,
		TokenId:      task.PrivateData.TokenId,
		TokenKey:     tokenKey,
		TaskId:       task.ID,
		NewTaskQuota: actualQuota,
	}
	if tokenExists {
		adjustment.TokenDelta = -quotaDelta
	} else {
		adjustment.TokenId = 0
	}
	if task.ID > 0 {
		adjustment.ExpectedTaskQuota = &preConsumedQuota
	}
	if taskIsSubscription(task) {
		adjustment.SubscriptionId = task.PrivateData.SubscriptionId
		adjustment.SubscriptionDelta = int64(quotaDelta)
	} else {
		adjustment.WalletDelta = -quotaDelta
	}
	applied, err := model.ApplyBillingAdjustment(adjustment)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("差额结算失败 task %s: %s", task.TaskID, err.Error()))
		return
	}
	task.Quota = actualQuota
	if !applied {
		return
	}

	var logType int
	var logQuota int
	if quotaDelta > 0 {
		logType = model.LogTypeConsume
		logQuota = quotaDelta
		model.UpdateUserUsedQuotaAndRequestCount(task.UserId, quotaDelta)
		model.UpdateChannelUsedQuota(task.ChannelId, quotaDelta)
	} else {
		logType = model.LogTypeRefund
		logQuota = -quotaDelta
	}
	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["pre_consumed_quota"] = preConsumedQuota
	other["actual_quota"] = actualQuota
	for _, clamp := range clamps {
		attachQuotaSaturationToOther(other, clamp)
	}
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:    task.UserId,
		LogType:   logType,
		Content:   reason,
		ChannelId: task.ChannelId,
		ModelName: taskModelName(task),
		Quota:     logQuota,
		TokenId:   task.PrivateData.TokenId,
		Group:     task.Group,
		Other:     other,
		NodeName:  task.PrivateData.NodeName,
	})
}

// RecalculateTaskQuotaByTokens 根据实际 token 消耗重新计费（异步差额结算）。
// 当任务成功且返回了 totalTokens 时，根据模型倍率和分组倍率重新计算实际扣费额度，
// 与预扣费的差额进行补扣或退还。支持钱包和订阅计费来源。
func RecalculateTaskQuotaByTokens(ctx context.Context, task *model.Task, totalTokens int) {
	if totalTokens <= 0 {
		return
	}

	modelName := taskModelName(task)

	// 获取模型价格和倍率
	modelRatio, hasRatioSetting, _ := ratio_setting.GetModelRatio(modelName)
	// 只有配置了倍率(非固定价格)时才按 token 重新计费
	if !hasRatioSetting || modelRatio <= 0 {
		return
	}

	// 获取用户和组的倍率信息
	group := task.Group
	if group == "" {
		user, err := model.GetUserById(task.UserId, false)
		if err == nil {
			group = user.Group
		}
	}
	if group == "" {
		return
	}

	groupRatio := ratio_setting.GetGroupRatio(group)
	userGroupRatio, hasUserGroupRatio := ratio_setting.GetGroupGroupRatio(group, group)

	var finalGroupRatio float64
	if hasUserGroupRatio {
		finalGroupRatio = userGroupRatio
	} else {
		finalGroupRatio = groupRatio
	}

	// 计算 OtherRatios 乘积（视频折扣、时长等）
	otherMultiplier := 1.0
	if priceData := taskBillingContextPriceData(task.PrivateData.BillingContext); priceData != nil {
		otherMultiplier = priceData.OtherRatioMultiplier()
	}

	// 计算实际应扣费额度: totalTokens * modelRatio * groupRatio * otherMultiplier（饱和转换，防止溢出成负数）
	actualQuota, clamp := common.QuotaFromFloatChecked(float64(totalTokens) * modelRatio * finalGroupRatio * otherMultiplier)

	reason := fmt.Sprintf("token重算：tokens=%d, modelRatio=%.2f, groupRatio=%.2f, otherMultiplier=%.4f", totalTokens, modelRatio, finalGroupRatio, otherMultiplier)
	RecalculateTaskQuota(ctx, task, actualQuota, reason, clamp)
}
