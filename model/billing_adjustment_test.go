package model

import (
	"testing"
	"time"

	"github.com/500wango/arcmux/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyBillingAdjustmentRollsBackWhenTokenIsMissing(t *testing.T) {
	truncateTables(t)

	user := User{Id: 501, Username: "adjustment-rollback", Quota: 100}
	require.NoError(t, DB.Create(&user).Error)

	applied, err := ApplyBillingAdjustment(BillingAdjustment{
		Key:         "adjustment:missing-token",
		UserId:      user.Id,
		WalletDelta: -20,
		TokenId:     999,
		TokenDelta:  -20,
	})
	require.Error(t, err)
	assert.False(t, applied)

	require.NoError(t, DB.Select("quota").First(&user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
	var recordCount int64
	require.NoError(t, DB.Model(&BillingAdjustmentRecord{}).
		Where("adjustment_key = ?", "adjustment:missing-token").Count(&recordCount).Error)
	assert.Zero(t, recordCount)
}

func TestApplyBillingAdjustmentTaskConflictRollsBackBalances(t *testing.T) {
	truncateTables(t)

	user := User{Id: 502, Username: "adjustment-task-conflict", Quota: 100}
	token := Token{Id: 502, UserId: user.Id, Key: "adjustment-task-token", RemainQuota: 100}
	task := Task{TaskID: "adjustment-task", UserId: user.Id, Quota: 50, Status: TaskStatusInProgress}
	require.NoError(t, DB.Create(&user).Error)
	require.NoError(t, DB.Create(&token).Error)
	require.NoError(t, DB.Create(&task).Error)

	expectedQuota := 40
	applied, err := ApplyBillingAdjustment(BillingAdjustment{
		Key:               "adjustment:task-conflict",
		UserId:            user.Id,
		WalletDelta:       -10,
		TokenId:           token.Id,
		TokenDelta:        -10,
		TaskId:            task.ID,
		ExpectedTaskQuota: &expectedQuota,
		NewTaskQuota:      60,
	})
	require.ErrorIs(t, err, ErrBillingStateConflict)
	assert.False(t, applied)

	require.NoError(t, DB.Select("quota").First(&user, user.Id).Error)
	require.NoError(t, DB.Select("remain_quota", "used_quota").First(&token, token.Id).Error)
	require.NoError(t, DB.Select("quota").First(&task, task.ID).Error)
	assert.Equal(t, 100, user.Quota)
	assert.Equal(t, 100, token.RemainQuota)
	assert.Zero(t, token.UsedQuota)
	assert.Equal(t, 50, task.Quota)
}

func TestRefundSubscriptionPreConsumeIsIdempotent(t *testing.T) {
	truncateTables(t)

	subscription := UserSubscription{
		Id:          503,
		UserId:      503,
		AmountTotal: 1_000,
		AmountUsed:  70,
		Status:      "active",
	}
	require.NoError(t, DB.Create(&subscription).Error)
	record := SubscriptionPreConsumeRecord{
		RequestId:          "subscription-refund-idempotent",
		UserId:             subscription.UserId,
		UserSubscriptionId: subscription.Id,
		PreConsumed:        20,
		Status:             "consumed",
	}
	require.NoError(t, DB.Create(&record).Error)

	require.NoError(t, RefundSubscriptionPreConsume(record.RequestId))
	require.NoError(t, RefundSubscriptionPreConsume(record.RequestId))

	require.NoError(t, DB.Select("amount_used").First(&subscription, subscription.Id).Error)
	require.NoError(t, DB.Where("request_id = ?", record.RequestId).First(&record).Error)
	assert.EqualValues(t, 50, subscription.AmountUsed)
	assert.Equal(t, "refunded", record.Status)
}

func TestApplyBillingAdjustmentRejectsNegativeChargeByInvalidIdentity(t *testing.T) {
	_, err := ApplyBillingAdjustment(BillingAdjustment{
		Key:         "adjustment:invalid-user",
		WalletDelta: -common.MaxQuota,
	})
	require.Error(t, err)
}

func TestApplyBillingAdjustmentRejectsOutOfRangeDelta(t *testing.T) {
	_, err := ApplyBillingAdjustment(BillingAdjustment{
		Key:         "adjustment:out-of-range",
		UserId:      1,
		WalletDelta: common.MaxQuota + 1,
	})
	require.ErrorContains(t, err, "exceeds quota limit")
}

func TestApplyBillingAdjustmentRejectsReusedKeyWithDifferentValues(t *testing.T) {
	truncateTables(t)

	user := User{Id: 509, Username: "adjustment-key-conflict", Quota: 100}
	require.NoError(t, DB.Create(&user).Error)
	first := BillingAdjustment{Key: "adjustment:key-conflict", UserId: user.Id, WalletDelta: -10}
	applied, err := ApplyBillingAdjustment(first)
	require.NoError(t, err)
	assert.True(t, applied)

	first.WalletDelta = -20
	applied, err = ApplyBillingAdjustment(first)
	require.ErrorIs(t, err, ErrBillingStateConflict)
	assert.False(t, applied)
	require.NoError(t, DB.Select("quota").First(&user, user.Id).Error)
	assert.Equal(t, 90, user.Quota)
}

func TestApplyBillingAdjustmentRejectsMismatchedSubscriptionRefund(t *testing.T) {
	truncateTables(t)

	first := UserSubscription{Id: 504, UserId: 504, AmountTotal: 1_000, AmountUsed: 40, Status: "active"}
	second := UserSubscription{Id: 505, UserId: 505, AmountTotal: 1_000, AmountUsed: 60, Status: "active"}
	require.NoError(t, DB.Create(&first).Error)
	require.NoError(t, DB.Create(&second).Error)
	record := SubscriptionPreConsumeRecord{
		RequestId:          "subscription-refund-mismatch",
		UserId:             first.UserId,
		UserSubscriptionId: first.Id,
		PreConsumed:        20,
		Status:             "consumed",
	}
	require.NoError(t, DB.Create(&record).Error)

	applied, err := ApplyBillingAdjustment(BillingAdjustment{
		Key:                            "adjustment:subscription-refund-mismatch",
		UserId:                         second.UserId,
		SubscriptionId:                 second.Id,
		SubscriptionRefundPreConsumeId: record.RequestId,
	})
	require.ErrorIs(t, err, ErrBillingStateConflict)
	assert.False(t, applied)

	require.NoError(t, DB.Select("amount_used").First(&first, first.Id).Error)
	require.NoError(t, DB.Select("amount_used").First(&second, second.Id).Error)
	require.NoError(t, DB.Where("request_id = ?", record.RequestId).First(&record).Error)
	assert.EqualValues(t, 40, first.AmountUsed)
	assert.EqualValues(t, 60, second.AmountUsed)
	assert.Equal(t, "consumed", record.Status)
}

func TestPreConsumeSubscriptionAndTokenRollsBackWhenTokenIsMissing(t *testing.T) {
	truncateTables(t)

	plan := SubscriptionPlan{Id: 506, Title: "atomic pre-consume", TotalAmount: 1_000}
	subscription := UserSubscription{
		Id:          506,
		UserId:      506,
		PlanId:      plan.Id,
		AmountTotal: 1_000,
		AmountUsed:  100,
		StartTime:   time.Now().Add(-time.Hour).Unix(),
		EndTime:     time.Now().Add(time.Hour).Unix(),
		Status:      "active",
	}
	require.NoError(t, DB.Create(&plan).Error)
	require.NoError(t, DB.Create(&subscription).Error)

	_, err := PreConsumeUserSubscriptionAndToken(
		"subscription-token-rollback",
		subscription.UserId,
		"test-model",
		0,
		50,
		999,
		"missing-token",
		true,
	)
	require.Error(t, err)

	require.NoError(t, DB.Select("amount_used").First(&subscription, subscription.Id).Error)
	assert.EqualValues(t, 100, subscription.AmountUsed)
	var preConsumeCount int64
	require.NoError(t, DB.Model(&SubscriptionPreConsumeRecord{}).
		Where("request_id = ?", "subscription-token-rollback").Count(&preConsumeCount).Error)
	assert.Zero(t, preConsumeCount)
	var adjustmentCount int64
	require.NoError(t, DB.Model(&BillingAdjustmentRecord{}).
		Where("adjustment_key = ?", "subscription-token-rollback:preconsume").Count(&adjustmentCount).Error)
	assert.Zero(t, adjustmentCount)
}

func TestPreConsumeSubscriptionAndTokenIsIdempotent(t *testing.T) {
	truncateTables(t)

	plan := SubscriptionPlan{Id: 507, Title: "idempotent pre-consume", TotalAmount: 1_000}
	subscription := UserSubscription{
		Id:          507,
		UserId:      507,
		PlanId:      plan.Id,
		AmountTotal: 1_000,
		AmountUsed:  100,
		StartTime:   time.Now().Add(-time.Hour).Unix(),
		EndTime:     time.Now().Add(time.Hour).Unix(),
		Status:      "active",
	}
	token := Token{Id: 507, UserId: subscription.UserId, Key: "subscription-token-idempotent", RemainQuota: 500}
	require.NoError(t, DB.Create(&plan).Error)
	require.NoError(t, DB.Create(&subscription).Error)
	require.NoError(t, DB.Create(&token).Error)

	for range 2 {
		result, err := PreConsumeUserSubscriptionAndToken(
			"subscription-token-idempotent",
			subscription.UserId,
			"test-model",
			0,
			50,
			token.Id,
			token.Key,
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, subscription.Id, result.UserSubscriptionId)
		assert.EqualValues(t, 50, result.PreConsumed)
	}

	require.NoError(t, DB.Select("amount_used").First(&subscription, subscription.Id).Error)
	require.NoError(t, DB.Select("remain_quota", "used_quota").First(&token, token.Id).Error)
	assert.EqualValues(t, 150, subscription.AmountUsed)
	assert.Equal(t, 450, token.RemainQuota)
	assert.Equal(t, 50, token.UsedQuota)
}

func TestPreConsumeSubscriptionRejectsReusedRequestWithDifferentAmount(t *testing.T) {
	truncateTables(t)

	plan := SubscriptionPlan{Id: 508, Title: "conflicting pre-consume", TotalAmount: 1_000}
	subscription := UserSubscription{
		Id:          508,
		UserId:      508,
		PlanId:      plan.Id,
		AmountTotal: 1_000,
		StartTime:   time.Now().Add(-time.Hour).Unix(),
		EndTime:     time.Now().Add(time.Hour).Unix(),
		Status:      "active",
	}
	require.NoError(t, DB.Create(&plan).Error)
	require.NoError(t, DB.Create(&subscription).Error)
	_, err := PreConsumeUserSubscription("subscription-request-conflict", subscription.UserId, "test-model", 0, 50)
	require.NoError(t, err)

	_, err = PreConsumeUserSubscription("subscription-request-conflict", subscription.UserId, "test-model", 0, 60)
	require.ErrorIs(t, err, ErrBillingStateConflict)
	require.NoError(t, DB.Select("amount_used").First(&subscription, subscription.Id).Error)
	assert.EqualValues(t, 50, subscription.AmountUsed)
}
