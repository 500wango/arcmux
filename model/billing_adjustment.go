package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/500wango/arcmux/common"
	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BillingAdjustmentRecord makes a multi-row billing adjustment idempotent.
// The adjustment key is derived from a request or task plus its lifecycle step.
type BillingAdjustmentRecord struct {
	Id             int64  `json:"id" gorm:"primaryKey"`
	AdjustmentKey  string `json:"adjustment_key" gorm:"type:varchar(191);uniqueIndex"`
	AdjustmentHash string `json:"adjustment_hash" gorm:"type:varchar(40);not null"`
	CreatedAt      int64  `json:"created_at" gorm:"index"`
}

type BillingAdjustment struct {
	Key                            string
	UserId                         int
	WalletDelta                    int
	WalletRequireSufficient        bool
	TokenId                        int
	TokenKey                       string
	TokenDelta                     int
	TokenRequireSufficient         bool
	SubscriptionId                 int
	SubscriptionDelta              int64
	SubscriptionRefundPreConsumeId string
	SubscriptionRefundExtraQuota   int64
	TaskId                         int64
	ExpectedTaskQuota              *int
	NewTaskQuota                   int
}

// ApplyBillingAdjustment atomically updates every persistent balance involved
// in one billing lifecycle step. Reusing Key is a successful no-op.
func ApplyBillingAdjustment(adjustment BillingAdjustment) (bool, error) {
	if strings.TrimSpace(adjustment.Key) == "" {
		return false, errors.New("billing adjustment key is empty")
	}
	if adjustment.WalletDelta != 0 && adjustment.UserId <= 0 {
		return false, errors.New("billing adjustment user id is invalid")
	}
	if adjustment.TokenDelta != 0 && adjustment.TokenId <= 0 {
		return false, errors.New("billing adjustment token id is invalid")
	}
	if adjustment.SubscriptionRefundPreConsumeId == "" &&
		(adjustment.SubscriptionDelta != 0 || adjustment.SubscriptionRefundExtraQuota != 0) &&
		adjustment.SubscriptionId <= 0 {
		return false, errors.New("billing adjustment subscription id is invalid")
	}
	if adjustment.ExpectedTaskQuota != nil && adjustment.TaskId <= 0 {
		return false, errors.New("billing adjustment task id is invalid")
	}
	if adjustment.WalletDelta < -common.MaxQuota || adjustment.WalletDelta > common.MaxQuota {
		return false, errors.New("billing adjustment wallet delta exceeds quota limit")
	}
	if adjustment.TokenDelta < -common.MaxQuota || adjustment.TokenDelta > common.MaxQuota {
		return false, errors.New("billing adjustment token delta exceeds quota limit")
	}
	if adjustment.SubscriptionDelta < -int64(common.MaxQuota) || adjustment.SubscriptionDelta > int64(common.MaxQuota) {
		return false, errors.New("billing adjustment subscription delta exceeds quota limit")
	}
	if adjustment.SubscriptionRefundExtraQuota < -int64(common.MaxQuota) || adjustment.SubscriptionRefundExtraQuota > int64(common.MaxQuota) {
		return false, errors.New("billing adjustment subscription refund exceeds quota limit")
	}
	if adjustment.NewTaskQuota < 0 || adjustment.NewTaskQuota > common.MaxQuota {
		return false, errors.New("billing adjustment task quota exceeds quota limit")
	}
	if adjustment.ExpectedTaskQuota != nil && (*adjustment.ExpectedTaskQuota < 0 || *adjustment.ExpectedTaskQuota > common.MaxQuota) {
		return false, errors.New("billing adjustment expected task quota exceeds quota limit")
	}
	expectedTaskQuota := "nil"
	if adjustment.ExpectedTaskQuota != nil {
		expectedTaskQuota = fmt.Sprintf("%d", *adjustment.ExpectedTaskQuota)
	}
	adjustmentHash := common.Sha1([]byte(fmt.Sprintf(
		"user=%d;wallet=%d;wallet_sufficient=%t;token=%d;token_delta=%d;token_sufficient=%t;subscription=%d;subscription_delta=%d;subscription_refund=%s;subscription_extra=%d;task=%d;expected_task=%s;new_task=%d",
		adjustment.UserId,
		adjustment.WalletDelta,
		adjustment.WalletRequireSufficient,
		adjustment.TokenId,
		adjustment.TokenDelta,
		adjustment.TokenRequireSufficient,
		adjustment.SubscriptionId,
		adjustment.SubscriptionDelta,
		adjustment.SubscriptionRefundPreConsumeId,
		adjustment.SubscriptionRefundExtraQuota,
		adjustment.TaskId,
		expectedTaskQuota,
		adjustment.NewTaskQuota,
	)))

	applied := false
	userAdjusted := false
	tokenAdjusted := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		record := &BillingAdjustmentRecord{
			AdjustmentKey:  adjustment.Key,
			AdjustmentHash: adjustmentHash,
			CreatedAt:      common.GetTimestamp(),
		}
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(record)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			var existing BillingAdjustmentRecord
			if err := tx.Where("adjustment_key = ?", adjustment.Key).First(&existing).Error; err != nil {
				return err
			}
			if existing.AdjustmentHash != adjustmentHash {
				return ErrBillingStateConflict
			}
			return nil
		}

		if adjustment.ExpectedTaskQuota != nil {
			updated := tx.Model(&Task{}).
				Where("id = ? AND quota = ?", adjustment.TaskId, *adjustment.ExpectedTaskQuota).
				Update("quota", adjustment.NewTaskQuota)
			if updated.Error != nil {
				return updated.Error
			}
			if updated.RowsAffected == 0 {
				var current int
				if err := tx.Model(&Task{}).Where("id = ?", adjustment.TaskId).Select("quota").Scan(&current).Error; err != nil {
					return err
				}
				return ErrBillingStateConflict
			}
		}
		applied = true

		if adjustment.WalletDelta != 0 {
			query := tx.Model(&User{}).Where("id = ?", adjustment.UserId)
			if adjustment.WalletRequireSufficient && adjustment.WalletDelta < 0 {
				query = query.Where("quota >= ?", -adjustment.WalletDelta)
			}
			updated := query.
				Update("quota", gorm.Expr("quota + ?", adjustment.WalletDelta))
			if updated.Error != nil {
				return updated.Error
			}
			if updated.RowsAffected == 0 {
				if adjustment.WalletRequireSufficient {
					return ErrInsufficientQuota
				}
				return fmt.Errorf("billing user %d not found", adjustment.UserId)
			}
			userAdjusted = true
		}

		subscriptionId := adjustment.SubscriptionId
		subscriptionDelta := adjustment.SubscriptionDelta + adjustment.SubscriptionRefundExtraQuota
		if adjustment.SubscriptionRefundPreConsumeId != "" {
			var preConsume SubscriptionPreConsumeRecord
			if err := lockForUpdate(tx).Where("request_id = ?", adjustment.SubscriptionRefundPreConsumeId).
				First(&preConsume).Error; err != nil {
				return err
			}
			if adjustment.UserId > 0 && preConsume.UserId != adjustment.UserId {
				return ErrBillingStateConflict
			}
			if subscriptionId > 0 && preConsume.UserSubscriptionId != subscriptionId {
				return ErrBillingStateConflict
			}
			subscriptionId = preConsume.UserSubscriptionId
			if preConsume.Status == "refunded" {
				subscriptionDelta = 0
			} else {
				subscriptionDelta -= preConsume.PreConsumed
				preConsume.Status = "refunded"
				if err := tx.Save(&preConsume).Error; err != nil {
					return err
				}
			}
		}
		if subscriptionDelta != 0 {
			if subscriptionDelta < -int64(common.MaxQuota) || subscriptionDelta > int64(common.MaxQuota) {
				return errors.New("combined subscription delta exceeds quota limit")
			}
			if subscriptionId <= 0 {
				return errors.New("billing adjustment subscription id is invalid")
			}
			if err := postConsumeUserSubscriptionDeltaTx(tx, subscriptionId, subscriptionDelta); err != nil {
				return err
			}
		}

		if adjustment.TokenId > 0 && adjustment.TokenDelta != 0 {
			if err := applyTokenBillingDeltaTx(tx, adjustment.TokenId, adjustment.TokenDelta, adjustment.TokenRequireSufficient); err != nil {
				return err
			}
			tokenAdjusted = true
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	if userAdjusted {
		gopool.Go(func() {
			if err := cacheIncrUserQuota(adjustment.UserId, int64(adjustment.WalletDelta)); err != nil {
				common.SysLog("failed to update user quota cache after billing adjustment: " + err.Error())
			}
		})
	}
	if tokenAdjusted && common.RedisEnabled && adjustment.TokenKey != "" {
		gopool.Go(func() {
			if err := cacheIncrTokenQuota(adjustment.TokenKey, int64(adjustment.TokenDelta)); err != nil {
				common.SysLog("failed to update token quota cache after billing adjustment: " + err.Error())
			}
		})
	}
	return applied, nil
}

func applyTokenBillingDeltaTx(tx *gorm.DB, tokenId int, delta int, requireSufficient bool) error {
	query := tx.Model(&Token{}).Where("id = ?", tokenId)
	if requireSufficient && delta < 0 {
		query = query.Where("remain_quota >= ?", -delta)
	}
	updated := query.Updates(map[string]interface{}{
		"remain_quota":  gorm.Expr("remain_quota + ?", delta),
		"used_quota":    gorm.Expr("used_quota - ?", delta),
		"accessed_time": common.GetTimestamp(),
	})
	if updated.Error != nil {
		return updated.Error
	}
	if updated.RowsAffected == 0 {
		if requireSufficient {
			return ErrInsufficientQuota
		}
		return fmt.Errorf("billing token %d not found", tokenId)
	}
	return nil
}

func CleanupBillingAdjustmentRecords(olderThanSeconds int64) (int64, error) {
	if olderThanSeconds <= 0 {
		olderThanSeconds = 90 * 24 * 3600
	}
	cutoff := GetDBTimestamp() - olderThanSeconds
	result := DB.Where("created_at < ?", cutoff).Delete(&BillingAdjustmentRecord{})
	return result.RowsAffected, result.Error
}
