package service

import (
	"context"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartUpstreamOAuthRequiresCommercialAcknowledgement(t *testing.T) {
	result, err := StartUpstreamOAuth(context.Background(), 1, 1, UpstreamOAuthProviderCodex, UpstreamOAuthFlowBrowser, false)
	require.ErrorContains(t, err, "commercial-use policy acknowledgement is required")
	assert.Nil(t, result)
}

func TestUpstreamOAuthClientSecretUsesProviderEnvironment(t *testing.T) {
	t.Setenv("UPSTREAM_OAUTH_GEMINI_CLIENT_SECRET", " gemini-secret ")
	t.Setenv("UPSTREAM_OAUTH_ANTIGRAVITY_CLIENT_SECRET", "antigravity-secret")

	assert.Equal(t, "gemini-secret", upstreamOAuthClientSecret(UpstreamOAuthProviderGeminiCLI))
	assert.Equal(t, "antigravity-secret", upstreamOAuthClientSecret(UpstreamOAuthProviderAntigravity))
	assert.Empty(t, upstreamOAuthClientSecret(UpstreamOAuthProviderCodex))
}

func TestBuildUpstreamAuthorizationURLUsesProviderPKCEContract(t *testing.T) {
	tests := []struct {
		provider string
		clientID string
		redirect string
	}{
		{provider: UpstreamOAuthProviderCodex, clientID: codexOAuthClientID, redirect: codexOAuthRedirectURI},
		{provider: UpstreamOAuthProviderClaude, clientID: claudeOAuthClientID, redirect: claudeOAuthRedirectURI},
		{provider: UpstreamOAuthProviderGeminiCLI, clientID: geminiCLIClientID, redirect: geminiCLIRedirectURI},
		{provider: UpstreamOAuthProviderAntigravity, clientID: antigravityClientID, redirect: antigravityRedirectURI},
	}
	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			parsed, err := url.Parse(buildUpstreamAuthorizationURL(test.provider, "state-value", "challenge-value"))
			require.NoError(t, err)
			assert.Equal(t, test.clientID, parsed.Query().Get("client_id"))
			assert.Equal(t, test.redirect, parsed.Query().Get("redirect_uri"))
			assert.Equal(t, "state-value", parsed.Query().Get("state"))
			assert.Equal(t, "challenge-value", parsed.Query().Get("code_challenge"))
			assert.Equal(t, "S256", parsed.Query().Get("code_challenge_method"))
		})
	}
}

func TestParseOAuthCallbackInputRequiresCompleteURL(t *testing.T) {
	code, state, oauthError, err := parseOAuthCallbackInput("http://localhost:1455/auth/callback?code=code-value&state=state-value")
	require.NoError(t, err)
	assert.Equal(t, "code-value", code)
	assert.Equal(t, "state-value", state)
	assert.Empty(t, oauthError)

	_, _, _, err = parseOAuthCallbackInput("code-value")
	require.Error(t, err)
}

func TestGeneratePKCEProducesMatchingChallenge(t *testing.T) {
	verifier, challenge, err := generatePKCE()
	require.NoError(t, err)
	assert.NotEmpty(t, verifier)
	assert.NotEmpty(t, challenge)
	assert.NotEqual(t, verifier, challenge)
}

func TestMergeRotatedUpstreamOAuthTokenPreservesOmittedFields(t *testing.T) {
	previous := &UpstreamOAuthTokenPayload{
		RefreshToken: "old-refresh", IDToken: "old-id", Scope: "old-scope",
		AccountID: "account", Email: "user@example.com", OrganizationID: "org",
	}
	refreshed := &UpstreamOAuthTokenPayload{AccessToken: "new-access", ExpiresAt: 123}

	mergeRotatedUpstreamOAuthToken(refreshed, previous)

	assert.Equal(t, "old-refresh", refreshed.RefreshToken)
	assert.Equal(t, "old-id", refreshed.IDToken)
	assert.Equal(t, "old-scope", refreshed.Scope)
	assert.Equal(t, "account", refreshed.AccountID)
	assert.Equal(t, "user@example.com", refreshed.Email)
	assert.Equal(t, "org", refreshed.OrganizationID)
}

func TestProvidersForChannelTypeIncludesOAuthVariants(t *testing.T) {
	assert.Equal(t, []string{UpstreamOAuthProviderGeminiCLI, UpstreamOAuthProviderAntigravity}, providersForChannelType(24))
	assert.Equal(t, []string{UpstreamOAuthProviderKimi}, providersForChannelType(25))
	assert.Equal(t, []string{UpstreamOAuthProviderXAI}, providersForChannelType(48))
}

func TestJWTSubjectUsesStableSubClaim(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"account-123","email":"user@example.com"}`))
	token := "header." + payload + ".signature"
	assert.Equal(t, "account-123", jwtSubject(token))
	assert.Empty(t, jwtSubject("not-a-jwt"))
}

func TestParseUpstreamOAuthImportSupportsArraysAndWrappedAccounts(t *testing.T) {
	items, err := parseUpstreamOAuthImport(`{
		"accounts": [
			{"access_token":"access-a","refresh_token":"refresh-a","account_id":"account-a","email":"a@example.com","expires_at":1893456000},
			{"accessToken":"access-b","refreshToken":"refresh-b","accountId":"account-b","email":"b@example.com","expired":"2030-01-01T00:00:00Z"}
		]
	}`)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "access-a", items[0].AccessToken)
	assert.Equal(t, "account-a", items[0].AccountID)
	assert.Equal(t, int64(1893456000), items[0].ExpiresAt)
	assert.Equal(t, "access-b", items[1].AccessToken)
	assert.Equal(t, "account-b", items[1].AccountID)
	assert.Equal(t, time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), items[1].ExpiresAt)
}

func TestParseUpstreamOAuthImportSupportsCodexAuthJSON(t *testing.T) {
	items, err := parseUpstreamOAuthImport(`{
		"tokens": {
			"id_token":"id-token",
			"access_token":"access-token",
			"refresh_token":"refresh-token",
			"account_id":"account-123"
		},
		"email":"user@example.com"
	}`)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "id-token", items[0].IDToken)
	assert.Equal(t, "access-token", items[0].AccessToken)
	assert.Equal(t, "refresh-token", items[0].RefreshToken)
	assert.Equal(t, "account-123", items[0].AccountID)
	assert.Equal(t, "user@example.com", items[0].Email)
}

func TestParseUpstreamOAuthImportSupportsNestedCodexAuthAliases(t *testing.T) {
	items, err := parseUpstreamOAuthImport(`{
		"auth": {
			"accessToken": "access-token",
			"refreshToken": "refresh-token",
			"chatgptAccountId": "account-456"
		}
	}`)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "access-token", items[0].AccessToken)
	assert.Equal(t, "refresh-token", items[0].RefreshToken)
	assert.Equal(t, "account-456", items[0].AccountID)
}

func TestExtractCodexAccountIDFromJWTSupportsTopLevelClaim(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"account_id":"account-789"}`))
	accountID, ok := ExtractCodexAccountIDFromJWT("header." + payload + ".signature")
	require.True(t, ok)
	assert.Equal(t, "account-789", accountID)
}

func TestParseUpstreamOAuthImportRejectsObjectsWithoutTokens(t *testing.T) {
	_, err := parseUpstreamOAuthImport(`{"account_id":"missing-token"}`)
	require.ErrorContains(t, err, "access_token or refresh_token")
}
