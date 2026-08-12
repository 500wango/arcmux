package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/500wango/arcmux/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertUserForPaymentGuardTest(t *testing.T, id int, quota int) {
	t.Helper()
	user := &User{
		Id:       id,
		Username: "payment_guard_user",
		Status:   common.UserStatusEnabled,
		Quota:    quota,
	}
	require.NoError(t, DB.Create(user).Error)
}

func insertSubscriptionPlanForPaymentGuardTest(t *testing.T, id int) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         "Guard Plan",
		PriceAmount:   9.99,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	return plan
}

func insertSubscriptionOrderForPaymentGuardTest(t *testing.T, tradeNo string, userID int, planID int, paymentProvider string) {
	t.Helper()
	order := &SubscriptionOrder{
		UserId:          userID,
		PlanId:          planID,
		Money:           9.99,
		TradeNo:         tradeNo,
		PaymentMethod:   paymentProvider,
		PaymentProvider: paymentProvider,
		Status:          common.TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}
	require.NoError(t, order.Insert())
}

func insertTopUpForPaymentGuardTest(t *testing.T, tradeNo string, userID int, paymentProvider string) {
	t.Helper()
	topUp := &TopUp{
		UserId:          userID,
		Amount:          2,
		Money:           9.99,
		TradeNo:         tradeNo,
		PaymentMethod:   paymentProvider,
		PaymentProvider: paymentProvider,
		Status:          common.TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}
	require.NoError(t, topUp.Insert())
}

func getTopUpStatusForPaymentGuardTest(t *testing.T, tradeNo string) string {
	t.Helper()
	topUp := GetTopUpByTradeNo(tradeNo)
	require.NotNil(t, topUp)
	return topUp.Status
}

func countUserSubscriptionsForPaymentGuardTest(t *testing.T, userID int) int64 {
	t.Helper()
	var count int64
	require.NoError(t, DB.Model(&UserSubscription{}).Where("user_id = ?", userID).Count(&count).Error)
	return count
}

func getUserQuotaForPaymentGuardTest(t *testing.T, userID int) int {
	t.Helper()
	var user User
	require.NoError(t, DB.Select("quota").Where("id = ?", userID).First(&user).Error)
	return user.Quota
}

func TestRechargeWaffoPancake_RejectsMismatchedPaymentMethod(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 101, 0)
	insertTopUpForPaymentGuardTest(t, "waffo-pancake-guard", 101, PaymentProviderStripe)

	err := RechargeWaffoPancake("waffo-pancake-guard", "9.99")
	require.Error(t, err)

	topUp := GetTopUpByTradeNo("waffo-pancake-guard")
	require.NotNil(t, topUp)
	assert.Equal(t, common.TopUpStatusPending, topUp.Status)
	assert.Equal(t, 0, getUserQuotaForPaymentGuardTest(t, 101))
}

func TestCompleteEpayTopUpIsAtomicAndIdempotent(t *testing.T) {
	truncateTables(t)

	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500_000
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	insertUserForPaymentGuardTest(t, 102, 100)
	insertTopUpForPaymentGuardTest(t, "epay-idempotent", 102, PaymentProviderEpay)

	topUp, quota, credited, err := CompleteEpayTopUp("epay-idempotent", "alipay", "9.99")
	require.NoError(t, err)
	require.NotNil(t, topUp)
	assert.True(t, credited)
	assert.Equal(t, 1_000_000, quota)
	assert.Equal(t, 1_000_100, getUserQuotaForPaymentGuardTest(t, 102))
	assert.Equal(t, common.TopUpStatusSuccess, getTopUpStatusForPaymentGuardTest(t, "epay-idempotent"))

	_, quota, credited, err = CompleteEpayTopUp("epay-idempotent", "alipay", "9.99")
	require.NoError(t, err)
	assert.False(t, credited)
	assert.Zero(t, quota)
	assert.Equal(t, 1_000_100, getUserQuotaForPaymentGuardTest(t, 102))
}

func TestCompleteEpayTopUpRejectsOverflowWithoutChangingOrder(t *testing.T) {
	truncateTables(t)

	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 1e20
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	insertUserForPaymentGuardTest(t, 103, 100)
	insertTopUpForPaymentGuardTest(t, "epay-overflow", 103, PaymentProviderEpay)

	_, _, _, err := CompleteEpayTopUp("epay-overflow", "alipay", "9.99")
	require.Error(t, err)
	assert.Equal(t, common.TopUpStatusPending, getTopUpStatusForPaymentGuardTest(t, "epay-overflow"))
	assert.Equal(t, 100, getUserQuotaForPaymentGuardTest(t, 103))
}

func TestCompleteEpayTopUpRejectsAmountMismatch(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 104, 100)
	insertTopUpForPaymentGuardTest(t, "epay-amount-mismatch", 104, PaymentProviderEpay)

	_, _, _, err := CompleteEpayTopUp("epay-amount-mismatch", "alipay", "0.01")
	require.ErrorContains(t, err, "支付金额不匹配")
	assert.Equal(t, common.TopUpStatusPending, getTopUpStatusForPaymentGuardTest(t, "epay-amount-mismatch"))
	assert.Equal(t, 100, getUserQuotaForPaymentGuardTest(t, 104))
}

func TestRechargeCallbacksRejectAmountMismatch(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		recharge func(string) error
	}{
		{
			name:     "creem",
			provider: PaymentProviderCreem,
			recharge: func(tradeNo string) error {
				return RechargeCreem(tradeNo, "0.01", "", "", "127.0.0.1")
			},
		},
		{
			name:     "waffo",
			provider: PaymentProviderWaffo,
			recharge: func(tradeNo string) error {
				return RechargeWaffo(tradeNo, "0.01", "127.0.0.1")
			},
		},
		{
			name:     "waffo pancake",
			provider: PaymentProviderWaffoPancake,
			recharge: func(tradeNo string) error {
				return RechargeWaffoPancake(tradeNo, "0.01")
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateTables(t)
			userID := 110 + i
			tradeNo := fmt.Sprintf("%s-amount-mismatch", tt.provider)
			insertUserForPaymentGuardTest(t, userID, 100)
			insertTopUpForPaymentGuardTest(t, tradeNo, userID, tt.provider)

			require.Error(t, tt.recharge(tradeNo))
			assert.Equal(t, common.TopUpStatusPending, getTopUpStatusForPaymentGuardTest(t, tradeNo))
			assert.Equal(t, 100, getUserQuotaForPaymentGuardTest(t, userID))
		})
	}
}

func TestRechargeStripeIsIdempotent(t *testing.T) {
	truncateTables(t)

	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500_000
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	insertUserForPaymentGuardTest(t, 105, 100)
	insertTopUpForPaymentGuardTest(t, "stripe-idempotent", 105, PaymentProviderStripe)

	require.NoError(t, Recharge("stripe-idempotent", "cus_test", "127.0.0.1"))
	require.NoError(t, Recharge("stripe-idempotent", "cus_test", "127.0.0.1"))
	assert.Equal(t, 4_995_100, getUserQuotaForPaymentGuardTest(t, 105))
}

func TestRechargeCreemIsIdempotent(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 106, 100)
	insertTopUpForPaymentGuardTest(t, "creem-idempotent", 106, PaymentProviderCreem)

	require.NoError(t, RechargeCreem("creem-idempotent", "9.99", "", "", "127.0.0.1"))
	require.NoError(t, RechargeCreem("creem-idempotent", "9.99", "", "", "127.0.0.1"))
	assert.Equal(t, 102, getUserQuotaForPaymentGuardTest(t, 106))
}

func TestManualCompleteCreemUsesFinalQuotaAmount(t *testing.T) {
	truncateTables(t)

	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500_000
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	insertUserForPaymentGuardTest(t, 107, 100)
	insertTopUpForPaymentGuardTest(t, "creem-manual-complete", 107, PaymentProviderCreem)

	require.NoError(t, ManualCompleteTopUp("creem-manual-complete", "127.0.0.1"))
	assert.Equal(t, 102, getUserQuotaForPaymentGuardTest(t, 107))
}

func TestUpdatePendingTopUpStatus_RejectsMismatchedPaymentProvider(t *testing.T) {
	testCases := []struct {
		name                    string
		tradeNo                 string
		storedPaymentProvider   string
		expectedPaymentProvider string
		targetStatus            string
	}{
		{
			name:                    "stripe expire",
			tradeNo:                 "stripe-expire-guard",
			storedPaymentProvider:   PaymentProviderCreem,
			expectedPaymentProvider: PaymentProviderStripe,
			targetStatus:            common.TopUpStatusExpired,
		},
		{
			name:                    "waffo failed",
			tradeNo:                 "waffo-failed-guard",
			storedPaymentProvider:   PaymentProviderStripe,
			expectedPaymentProvider: PaymentProviderWaffo,
			targetStatus:            common.TopUpStatusFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			truncateTables(t)
			insertUserForPaymentGuardTest(t, 150, 0)
			insertTopUpForPaymentGuardTest(t, tc.tradeNo, 150, tc.storedPaymentProvider)

			err := UpdatePendingTopUpStatus(tc.tradeNo, tc.expectedPaymentProvider, tc.targetStatus)
			require.ErrorIs(t, err, ErrPaymentMethodMismatch)
			assert.Equal(t, common.TopUpStatusPending, getTopUpStatusForPaymentGuardTest(t, tc.tradeNo))
		})
	}
}

func TestCompleteSubscriptionOrder_RejectsMismatchedPaymentProvider(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 202, 0)
	plan := insertSubscriptionPlanForPaymentGuardTest(t, 301)
	insertSubscriptionOrderForPaymentGuardTest(t, "sub-guard-order", 202, plan.Id, PaymentProviderStripe)

	err := CompleteSubscriptionOrder("sub-guard-order", `{"provider":"epay"}`, PaymentProviderEpay, "alipay")
	require.ErrorIs(t, err, ErrPaymentMethodMismatch)

	order := GetSubscriptionOrderByTradeNo("sub-guard-order")
	require.NotNil(t, order)
	assert.Equal(t, common.TopUpStatusPending, order.Status)
	assert.Zero(t, countUserSubscriptionsForPaymentGuardTest(t, 202))

	topUp := GetTopUpByTradeNo("sub-guard-order")
	assert.Nil(t, topUp)
}

func TestExpireSubscriptionOrder_RejectsMismatchedPaymentProvider(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 303, 0)
	plan := insertSubscriptionPlanForPaymentGuardTest(t, 401)
	insertSubscriptionOrderForPaymentGuardTest(t, "sub-expire-guard", 303, plan.Id, PaymentProviderStripe)

	err := ExpireSubscriptionOrder("sub-expire-guard", PaymentProviderCreem)
	require.ErrorIs(t, err, ErrPaymentMethodMismatch)

	order := GetSubscriptionOrderByTradeNo("sub-expire-guard")
	require.NotNil(t, order)
	assert.Equal(t, common.TopUpStatusPending, order.Status)
}
