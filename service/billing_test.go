package service

import (
	"testing"

	relaycommon "github.com/500wango/arcmux/relay/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettleBillingRejectsNegativeQuota(t *testing.T) {
	err := SettleBilling(nil, &relaycommon.RelayInfo{}, -1)
	require.ErrorContains(t, err, "cannot be negative")
}

func TestBillingSessionPlaygroundWalletPreConsumeNeedsRefund(t *testing.T) {
	session := &BillingSession{
		relayInfo:        &relaycommon.RelayInfo{IsPlayground: true},
		funding:          &WalletFunding{userId: 1, consumed: 100},
		preConsumedQuota: 100,
	}

	assert.True(t, session.NeedsRefund())
	assert.Zero(t, session.tokenConsumed)
}
