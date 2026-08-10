package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupUpstreamOAuthModelTest(t *testing.T) *gorm.DB {
	t.Helper()
	originalDB := DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&UpstreamOAuthSession{}, &UpstreamCredential{}, &UpstreamCredentialModelState{}))
	DB = db
	t.Cleanup(func() {
		DB = originalDB
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			require.NoError(t, sqlDB.Close())
		}
	})
	return db
}

func TestSelectUpstreamCredentialRotatesStableAccountIDs(t *testing.T) {
	setupUpstreamOAuthModelTest(t)
	first := &UpstreamCredential{ChannelId: 11, Provider: "codex", AccountId: "account-a", EncryptedPayload: "a"}
	second := &UpstreamCredential{ChannelId: 11, Provider: "codex", AccountId: "account-b", EncryptedPayload: "b"}
	require.NoError(t, UpsertUpstreamCredential(first))
	require.NoError(t, UpsertUpstreamCredential(second))

	selected, exists, err := SelectUpstreamCredential(11, "codex", "gpt-test", time.Now().Unix())
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "account-a", selected.AccountId)

	selected, exists, err = SelectUpstreamCredential(11, "codex", "gpt-test", time.Now().Add(time.Second).Unix())
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "account-b", selected.AccountId)
}

func TestSelectUpstreamCredentialSkipsDisabledAndCoolingAccounts(t *testing.T) {
	db := setupUpstreamOAuthModelTest(t)
	now := time.Now().Unix()
	credentials := []*UpstreamCredential{
		{ChannelId: 12, Provider: "claude", AccountId: "disabled", EncryptedPayload: "a"},
		{ChannelId: 12, Provider: "claude", AccountId: "cooling", EncryptedPayload: "b"},
		{ChannelId: 12, Provider: "claude", AccountId: "ready", EncryptedPayload: "c"},
	}
	for _, credential := range credentials {
		require.NoError(t, UpsertUpstreamCredential(credential))
	}
	require.NoError(t, UpdateUpstreamCredentialStatus(credentials[0].Id, 12, UpstreamCredentialDisabled, "manual"))
	require.NoError(t, db.Model(&UpstreamCredential{}).Where("id = ?", credentials[1].Id).Update("cooldown_until", now+300).Error)

	selected, exists, err := SelectUpstreamCredential(12, "claude", "claude-test", now)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "ready", selected.AccountId)
}

func TestSelectUpstreamCredentialDistinguishesNoPoolFromUnavailablePool(t *testing.T) {
	setupUpstreamOAuthModelTest(t)
	selected, exists, err := SelectUpstreamCredential(13, "codex", "gpt-test", time.Now().Unix())
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, selected)

	credential := &UpstreamCredential{ChannelId: 13, Provider: "codex", AccountId: "disabled", EncryptedPayload: "a"}
	require.NoError(t, UpsertUpstreamCredential(credential))
	require.NoError(t, UpdateUpstreamCredentialStatus(credential.Id, 13, UpstreamCredentialDisabled, "manual"))
	selected, exists, err = SelectUpstreamCredential(13, "codex", "gpt-test", time.Now().Unix())
	require.Error(t, err)
	assert.True(t, exists)
	assert.Nil(t, selected)
}

func TestClaimUpstreamCredentialRefreshAllowsOnlyOneVersionOwner(t *testing.T) {
	setupUpstreamOAuthModelTest(t)
	credential := &UpstreamCredential{ChannelId: 14, Provider: "codex", AccountId: "account", EncryptedPayload: "payload"}
	require.NoError(t, UpsertUpstreamCredential(credential))
	now := time.Now().Unix()

	claimed, err := ClaimUpstreamCredentialRefresh(credential.Id, credential.Version, now, now+60)
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = ClaimUpstreamCredentialRefresh(credential.Id, credential.Version, now, now+60)
	require.NoError(t, err)
	assert.False(t, claimed)
}

func TestSelectUpstreamCredentialSkipsOnlyTheBlockedModel(t *testing.T) {
	setupUpstreamOAuthModelTest(t)
	now := time.Now().Unix()
	first := &UpstreamCredential{ChannelId: 15, Provider: "codex", AccountId: "account-a", EncryptedPayload: "a"}
	second := &UpstreamCredential{ChannelId: 15, Provider: "codex", AccountId: "account-b", EncryptedPayload: "b"}
	require.NoError(t, UpsertUpstreamCredential(first))
	require.NoError(t, UpsertUpstreamCredential(second))
	require.NoError(t, MarkUpstreamCredentialModelFailure(first.Id, "gpt-blocked", "model unavailable", now+600))

	selected, exists, err := SelectUpstreamCredential(15, "codex", "gpt-blocked", now)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, second.Id, selected.Id)

	selected, exists, err = SelectUpstreamCredential(15, "codex", "gpt-available", now+1)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, first.Id, selected.Id)
}

func TestDeleteUpstreamCredentialDoesNotTouchAnotherChannel(t *testing.T) {
	setupUpstreamOAuthModelTest(t)
	credential := &UpstreamCredential{ChannelId: 16, Provider: "codex", AccountId: "account", EncryptedPayload: "payload"}
	require.NoError(t, UpsertUpstreamCredential(credential))
	require.NoError(t, MarkUpstreamCredentialModelFailure(credential.Id, "gpt-test", "blocked", time.Now().Add(time.Minute).Unix()))

	err := DeleteUpstreamCredential(credential.Id, 17)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var credentialCount int64
	require.NoError(t, DB.Model(&UpstreamCredential{}).Where("id = ?", credential.Id).Count(&credentialCount).Error)
	assert.Equal(t, int64(1), credentialCount)
	var stateCount int64
	require.NoError(t, DB.Model(&UpstreamCredentialModelState{}).Where("credential_id = ?", credential.Id).Count(&stateCount).Error)
	assert.Equal(t, int64(1), stateCount)
}

func TestDeleteUpstreamOAuthForChannelsRemovesCredentialsSessionsAndModelState(t *testing.T) {
	setupUpstreamOAuthModelTest(t)
	credential := &UpstreamCredential{ChannelId: 18, Provider: "claude", AccountId: "account", EncryptedPayload: "payload"}
	require.NoError(t, UpsertUpstreamCredential(credential))
	require.NoError(t, MarkUpstreamCredentialModelFailure(credential.Id, "claude-test", "blocked", time.Now().Add(time.Minute).Unix()))
	require.NoError(t, CreateUpstreamOAuthSession(&UpstreamOAuthSession{
		Id: "session", ChannelId: 18, AdminUserId: 1, Provider: "claude", FlowType: "browser",
		StateHash: "state", Status: UpstreamOAuthSessionPending, CreatedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}))

	require.NoError(t, deleteUpstreamOAuthForChannels(DB, []int{18}))
	for _, value := range []any{&UpstreamCredential{}, &UpstreamCredentialModelState{}, &UpstreamOAuthSession{}} {
		var count int64
		require.NoError(t, DB.Model(value).Count(&count).Error)
		assert.Zero(t, count)
	}
}
