package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/google/uuid"
)

const (
	UpstreamOAuthProviderCodex       = "codex"
	UpstreamOAuthProviderClaude      = "claude"
	UpstreamOAuthProviderGeminiCLI   = "gemini-cli"
	UpstreamOAuthProviderAntigravity = "antigravity"
	UpstreamOAuthProviderKimi        = "kimi"
	UpstreamOAuthProviderXAI         = "xai"
	UpstreamOAuthFlowBrowser         = "browser"
	UpstreamOAuthFlowDevice          = "device"

	codexOAuthAuthorizeURL        = "https://auth.openai.com/oauth/authorize"
	codexOAuthRedirectURI         = "http://localhost:1455/auth/callback"
	codexDeviceUserCodeURL        = "https://auth.openai.com/api/accounts/deviceauth/usercode"
	codexDeviceTokenURL           = "https://auth.openai.com/api/accounts/deviceauth/token"
	codexDeviceVerificationURL    = "https://auth.openai.com/codex/device"
	codexDeviceExchangeRedirect   = "https://auth.openai.com/deviceauth/callback"
	claudeOAuthAuthorizeURL       = "https://claude.ai/oauth/authorize"
	claudeOAuthTokenURL           = "https://platform.claude.com/v1/oauth/token"
	claudeOAuthClientID           = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	claudeOAuthRedirectURI        = "http://localhost:54545/callback"
	claudeOAuthScope              = "user:profile user:inference user:sessions:claude_code user:mcp_servers user:file_upload"
	geminiCLIClientID             = "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com"
	geminiCLIRedirectURI          = "http://localhost:8085/oauth2callback"
	geminiCLIAuthorizeURL         = "https://accounts.google.com/o/oauth2/v2/auth"
	antigravityClientID           = "1071006060591-tmhssin2h21lcre235vtolojh4g403ep.apps.googleusercontent.com"
	antigravityRedirectURI        = "http://localhost:51121/oauth-callback"
	kimiClientID                  = "17e5f671-d194-4dfb-9706-5516cb48c098"
	kimiDeviceCodeURL             = "https://auth.kimi.com/api/oauth/device_authorization"
	kimiTokenURL                  = "https://auth.kimi.com/api/oauth/token"
	kimiVerificationURL           = "https://auth.kimi.com/device"
	xaiClientID                   = "b1a00492-073a-47ea-816f-4c329264a828"
	xaiDiscoveryURL               = "https://auth.x.ai/.well-known/openid-configuration"
	xaiVerificationURL            = "https://accounts.x.ai/oauth2/device"
	upstreamOAuthSessionTTL       = 15 * time.Minute
	upstreamOAuthRefreshThreshold = 10 * time.Minute
)

func upstreamOAuthClientSecret(provider string) string {
	var environmentVariable string
	switch provider {
	case UpstreamOAuthProviderGeminiCLI:
		environmentVariable = "UPSTREAM_OAUTH_GEMINI_CLIENT_SECRET"
	case UpstreamOAuthProviderAntigravity:
		environmentVariable = "UPSTREAM_OAUTH_ANTIGRAVITY_CLIENT_SECRET"
	default:
		return ""
	}
	return strings.TrimSpace(os.Getenv(environmentVariable))
}

var ErrUpstreamOAuthPending = errors.New("upstream OAuth authorization is pending")

type UpstreamOAuthTokenPayload struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	Scope            string `json:"scope,omitempty"`
	AccountID        string `json:"account_id"`
	Email            string `json:"email,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	ExpiresAt        int64  `json:"expires_at"`
	DeviceID         string `json:"device_id,omitempty"`
	ProjectID        string `json:"project_id,omitempty"`
	BaseURL          string `json:"base_url,omitempty"`
	TokenEndpoint    string `json:"token_endpoint,omitempty"`
	AuthKind         string `json:"auth_kind,omitempty"`
}

type UpstreamOAuthStartResult struct {
	SessionID        string `json:"session_id"`
	Provider         string `json:"provider"`
	FlowType         string `json:"flow_type"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
	VerificationURL  string `json:"verification_url,omitempty"`
	UserCode         string `json:"user_code,omitempty"`
	ExpiresAt        int64  `json:"expires_at"`
	PollInterval     int    `json:"poll_interval,omitempty"`
}

type upstreamDeviceMetadata struct {
	DeviceAuthID string `json:"device_auth_id"`
	UserCode     string `json:"user_code"`
	DeviceCode   string `json:"device_code,omitempty"`
	TokenURL     string `json:"token_url,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Provider     string `json:"provider,omitempty"`
	DeviceID     string `json:"device_id,omitempty"`
}

func StartUpstreamOAuth(ctx context.Context, adminUserId int, channelId int, provider string, flowType string, commercialAcknowledged bool) (*UpstreamOAuthStartResult, error) {
	if !commercialAcknowledged {
		return nil, errors.New("commercial-use policy acknowledgement is required")
	}
	if !common.UpstreamCredentialEncryptionConfigured() {
		return nil, errors.New("UPSTREAM_CREDENTIAL_ENCRYPTION_KEY is not configured")
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	flowType = strings.ToLower(strings.TrimSpace(flowType))
	if !IsUpstreamOAuthProviderEnabled(provider) {
		return nil, fmt.Errorf("upstream OAuth provider %s is not enabled by deployment policy", provider)
	}
	channel, err := model.GetChannelById(channelId, true)
	if err != nil {
		return nil, err
	}
	if err = validateUpstreamOAuthChannel(channel, provider); err != nil {
		return nil, err
	}
	if flowType == "" {
		flowType = UpstreamOAuthFlowBrowser
	}
	if flowType == UpstreamOAuthFlowDevice && provider != UpstreamOAuthProviderCodex && provider != UpstreamOAuthProviderKimi && provider != UpstreamOAuthProviderXAI {
		return nil, errors.New("device authorization is only supported for Codex, Kimi, and xAI")
	}
	if flowType != UpstreamOAuthFlowBrowser && flowType != UpstreamOAuthFlowDevice {
		return nil, errors.New("unsupported OAuth flow type")
	}

	state, err := randomURLToken(32)
	if err != nil {
		return nil, err
	}
	session := &model.UpstreamOAuthSession{
		Id: uuid.NewString(), ChannelId: channelId, AdminUserId: adminUserId, Provider: provider,
		FlowType: flowType, StateHash: hashOAuthState(state), Status: model.UpstreamOAuthSessionPending,
		CreatedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(upstreamOAuthSessionTTL).Unix(),
	}
	result := &UpstreamOAuthStartResult{SessionID: session.Id, Provider: provider, FlowType: flowType, ExpiresAt: session.ExpiresAt}
	client, err := GetHttpClientWithProxy(channel.GetSetting().Proxy)
	if err != nil {
		return nil, err
	}

	if flowType == UpstreamOAuthFlowDevice {
		metadata, interval, err := startCodexDeviceAuthorization(ctx, client)
		if provider == UpstreamOAuthProviderKimi || provider == UpstreamOAuthProviderXAI {
			metadata, interval, err = startGenericDeviceAuthorization(ctx, client, provider)
		}
		if err != nil {
			return nil, err
		}
		raw, err := common.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		session.EncryptedMetadata, err = common.EncryptUpstreamCredential(raw)
		if err != nil {
			return nil, err
		}
		result.VerificationURL = codexDeviceVerificationURL
		if provider == UpstreamOAuthProviderKimi {
			result.VerificationURL = kimiVerificationURL
		}
		if provider == UpstreamOAuthProviderXAI {
			result.VerificationURL = xaiVerificationURL
		}
		result.UserCode = metadata.UserCode
		result.PollInterval = interval
	} else {
		verifier, challenge, err := generatePKCE()
		if err != nil {
			return nil, err
		}
		session.EncryptedVerifier, err = common.EncryptUpstreamCredential([]byte(verifier))
		if err != nil {
			return nil, err
		}
		result.AuthorizationURL = buildUpstreamAuthorizationURL(provider, state, challenge)
	}
	if err = model.CreateUpstreamOAuthSession(session); err != nil {
		return nil, err
	}
	return result, nil
}

func IsUpstreamOAuthProviderEnabled(provider string) bool {
	provider = strings.ToLower(strings.TrimSpace(provider))
	for _, configured := range strings.Split(os.Getenv("UPSTREAM_OAUTH_ENABLED_PROVIDERS"), ",") {
		if strings.ToLower(strings.TrimSpace(configured)) == provider {
			return slices.Contains([]string{UpstreamOAuthProviderCodex, UpstreamOAuthProviderClaude, UpstreamOAuthProviderGeminiCLI, UpstreamOAuthProviderAntigravity, UpstreamOAuthProviderKimi, UpstreamOAuthProviderXAI}, provider)
		}
	}
	return false
}

func EnabledUpstreamOAuthProviders() []string {
	providers := make([]string, 0, 6)
	for _, provider := range []string{UpstreamOAuthProviderCodex, UpstreamOAuthProviderClaude, UpstreamOAuthProviderGeminiCLI, UpstreamOAuthProviderAntigravity, UpstreamOAuthProviderKimi, UpstreamOAuthProviderXAI} {
		if IsUpstreamOAuthProviderEnabled(provider) {
			providers = append(providers, provider)
		}
	}
	return providers
}

func CompleteUpstreamOAuth(ctx context.Context, adminUserId int, channelId int, sessionId string, callbackInput string) (*model.UpstreamCredential, error) {
	session, err := model.GetUpstreamOAuthSession(sessionId, adminUserId)
	if err != nil {
		return nil, err
	}
	if err = validatePendingOAuthSession(session); err != nil {
		return nil, err
	}
	if session.ChannelId != channelId {
		return nil, errors.New("OAuth session does not belong to channel")
	}
	if session.FlowType != UpstreamOAuthFlowBrowser {
		return nil, errors.New("this OAuth session uses device authorization")
	}
	code, state, oauthError, err := parseOAuthCallbackInput(callbackInput)
	if err != nil {
		return nil, err
	}
	if oauthError != "" {
		_ = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionFailed, oauthError)
		return nil, fmt.Errorf("upstream authorization failed: %s", oauthError)
	}
	if hashOAuthState(state) != session.StateHash {
		return nil, errors.New("OAuth state does not match the authorization session")
	}
	verifier, err := common.DecryptUpstreamCredential(session.EncryptedVerifier)
	if err != nil {
		return nil, err
	}
	channel, err := model.GetChannelById(session.ChannelId, true)
	if err != nil {
		return nil, err
	}
	payload, err := exchangeUpstreamOAuthCode(ctx, channel, session.Provider, code, state, string(verifier), "")
	if err != nil {
		_ = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionFailed, sanitizeUpstreamOAuthError(err))
		return nil, err
	}
	credential, err := saveUpstreamOAuthCredential(session.ChannelId, session.Provider, payload)
	if err != nil {
		return nil, err
	}
	if err = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionCompleted, ""); err != nil {
		return nil, err
	}
	return credential, nil
}

func PollUpstreamOAuth(ctx context.Context, adminUserId int, channelId int, sessionId string) (*model.UpstreamCredential, error) {
	session, err := model.GetUpstreamOAuthSession(sessionId, adminUserId)
	if err != nil {
		return nil, err
	}
	if err = validatePendingOAuthSession(session); err != nil {
		return nil, err
	}
	if session.ChannelId != channelId {
		return nil, errors.New("OAuth session does not belong to channel")
	}
	if session.FlowType != UpstreamOAuthFlowDevice || !slices.Contains([]string{UpstreamOAuthProviderCodex, UpstreamOAuthProviderKimi, UpstreamOAuthProviderXAI}, session.Provider) {
		return nil, errors.New("this OAuth session does not support device polling")
	}
	raw, err := common.DecryptUpstreamCredential(session.EncryptedMetadata)
	if err != nil {
		return nil, err
	}
	var metadata upstreamDeviceMetadata
	if err = common.Unmarshal(raw, &metadata); err != nil {
		return nil, err
	}
	channel, err := model.GetChannelById(session.ChannelId, true)
	if err != nil {
		return nil, err
	}
	client, err := GetHttpClientWithProxy(channel.GetSetting().Proxy)
	if err != nil {
		return nil, err
	}
	if session.Provider != UpstreamOAuthProviderCodex {
		payload, pollErr := pollGenericDeviceAuthorization(ctx, client, metadata)
		if errors.Is(pollErr, ErrUpstreamOAuthPending) {
			return nil, pollErr
		}
		if pollErr != nil {
			_ = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionFailed, sanitizeUpstreamOAuthError(pollErr))
			return nil, pollErr
		}
		credential, saveErr := saveUpstreamOAuthCredential(session.ChannelId, session.Provider, payload)
		if saveErr != nil {
			return nil, saveErr
		}
		if saveErr = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionCompleted, ""); saveErr != nil {
			return nil, saveErr
		}
		return credential, nil
	}
	authorizationCode, verifier, err := pollCodexDeviceAuthorization(ctx, client, metadata)
	if errors.Is(err, ErrUpstreamOAuthPending) {
		return nil, err
	}
	if err != nil {
		_ = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionFailed, sanitizeUpstreamOAuthError(err))
		return nil, err
	}
	payload, err := exchangeUpstreamOAuthCode(ctx, channel, session.Provider, authorizationCode, "", verifier, codexDeviceExchangeRedirect)
	if err != nil {
		_ = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionFailed, sanitizeUpstreamOAuthError(err))
		return nil, err
	}
	credential, err := saveUpstreamOAuthCredential(session.ChannelId, session.Provider, payload)
	if err != nil {
		return nil, err
	}
	if err = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionCompleted, ""); err != nil {
		return nil, err
	}
	return credential, nil
}

func RefreshUpstreamCredential(ctx context.Context, credential *model.UpstreamCredential) error {
	releaseRedis, acquired, err := acquireUpstreamCredentialRefreshLock(ctx, credential.Id)
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("OAuth credential refresh is already in progress")
	}
	defer releaseRedis()

	now := time.Now()
	claimed, err := model.ClaimUpstreamCredentialRefresh(credential.Id, credential.Version, now.Unix(), now.Add(time.Minute).Unix())
	if err != nil {
		return err
	}
	if !claimed {
		return errors.New("OAuth credential refresh is already in progress")
	}
	refreshSucceeded := false
	defer func() {
		if !refreshSucceeded {
			_ = model.ReleaseUpstreamCredentialRefresh(credential.Id)
		}
	}()

	raw, err := common.DecryptUpstreamCredential(credential.EncryptedPayload)
	if err != nil {
		return err
	}
	var payload UpstreamOAuthTokenPayload
	if err = common.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if strings.TrimSpace(payload.RefreshToken) == "" {
		return errors.New("OAuth credential has no refresh token")
	}
	channel, err := model.GetChannelById(credential.ChannelId, true)
	if err != nil {
		return err
	}
	refreshed, err := refreshUpstreamOAuthToken(ctx, channel, credential.Provider, payload.RefreshToken, payload.TokenEndpoint)
	if err != nil {
		return err
	}
	mergeRotatedUpstreamOAuthToken(refreshed, &payload)
	_, err = saveUpstreamOAuthCredential(credential.ChannelId, credential.Provider, refreshed)
	refreshSucceeded = err == nil
	return err
}

func mergeRotatedUpstreamOAuthToken(refreshed *UpstreamOAuthTokenPayload, previous *UpstreamOAuthTokenPayload) {
	if refreshed.AccountID == "" {
		refreshed.AccountID = previous.AccountID
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = previous.RefreshToken
	}
	if refreshed.IDToken == "" {
		refreshed.IDToken = previous.IDToken
	}
	if refreshed.Scope == "" {
		refreshed.Scope = previous.Scope
	}
	if refreshed.Email == "" {
		refreshed.Email = previous.Email
	}
	if refreshed.OrganizationID == "" {
		refreshed.OrganizationID = previous.OrganizationID
	}
	if refreshed.OrganizationName == "" {
		refreshed.OrganizationName = previous.OrganizationName
	}
	if refreshed.DeviceID == "" {
		refreshed.DeviceID = previous.DeviceID
	}
	if refreshed.ProjectID == "" {
		refreshed.ProjectID = previous.ProjectID
	}
	if refreshed.BaseURL == "" {
		refreshed.BaseURL = previous.BaseURL
	}
	if refreshed.TokenEndpoint == "" {
		refreshed.TokenEndpoint = previous.TokenEndpoint
	}
	if refreshed.AuthKind == "" {
		refreshed.AuthKind = previous.AuthKind
	}
}

func acquireUpstreamCredentialRefreshLock(ctx context.Context, credentialId int64) (func(), bool, error) {
	if !common.RedisEnabled || common.RDB == nil {
		return func() {}, true, nil
	}
	key := fmt.Sprintf("upstream_oauth:refresh:%d", credentialId)
	token := uuid.NewString()
	acquired, err := common.RDB.SetNX(ctx, key, token, time.Minute).Result()
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("upstream OAuth Redis refresh lock unavailable, using database lock: %v", err))
		return func() {}, true, nil
	}
	if !acquired {
		return func() {}, false, nil
	}
	return func() {
		const releaseScript = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
		if err := common.RDB.Eval(context.Background(), releaseScript, []string{key}, token).Err(); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("upstream OAuth Redis refresh lock release failed: %v", err))
		}
	}, true, nil
}

func SelectUpstreamOAuthCredential(ctx context.Context, channel *model.Channel, modelName string) (string, int64, string, bool, error) {
	providers := providersForChannelType(channel.Type)
	if len(providers) == 0 {
		return "", 0, "", false, nil
	}
	var lastErr error
	anyConfigured := false
	for _, provider := range providers {
		key, credentialID, selectedProvider, exists, err := selectUpstreamOAuthCredentialForProvider(ctx, channel, provider, modelName)
		anyConfigured = anyConfigured || exists
		if err == nil && exists {
			return key, credentialID, selectedProvider, true, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	return "", 0, providers[0], anyConfigured, lastErr
}

// GetUpstreamOAuthRequestMetadata returns only non-secret routing metadata for a credential.
func GetUpstreamOAuthRequestMetadata(credentialID int64, channelID int) (*UpstreamOAuthTokenPayload, error) {
	if credentialID <= 0 {
		return nil, nil
	}
	credential, err := model.GetUpstreamCredential(credentialID, channelID)
	if err != nil {
		return nil, err
	}
	raw, err := common.DecryptUpstreamCredential(credential.EncryptedPayload)
	if err != nil {
		return nil, err
	}
	var payload UpstreamOAuthTokenPayload
	if err = common.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	payload.AccessToken, payload.RefreshToken, payload.IDToken = "", "", ""
	return &payload, nil
}

func selectUpstreamOAuthCredentialForProvider(ctx context.Context, channel *model.Channel, provider string, modelName string) (string, int64, string, bool, error) {
	credential, exists, err := model.SelectUpstreamCredential(channel.Id, provider, modelName, time.Now().Unix())
	if err != nil || !exists {
		return "", 0, provider, exists, err
	}
	raw, err := common.DecryptUpstreamCredential(credential.EncryptedPayload)
	if err != nil {
		return "", credential.Id, provider, true, err
	}
	var payload UpstreamOAuthTokenPayload
	if err = common.Unmarshal(raw, &payload); err != nil {
		return "", credential.Id, provider, true, err
	}
	// Imported access-token-only credentials are valid until the upstream
	// rejects them. Refresh proactively only when a refresh token is present;
	// otherwise selection must not turn a token import into a login/refresh
	// failure.
	if credential.ExpiresAt > 0 && credential.ExpiresAt <= time.Now().Add(upstreamOAuthRefreshThreshold).Unix() && strings.TrimSpace(payload.RefreshToken) != "" {
		if err = RefreshUpstreamCredential(ctx, credential); err != nil {
			return "", credential.Id, provider, true, err
		}
		credential, err = model.GetUpstreamCredential(credential.Id, channel.Id)
		if err != nil {
			return "", credential.Id, provider, true, err
		}
		raw, err = common.DecryptUpstreamCredential(credential.EncryptedPayload)
		if err != nil {
			return "", credential.Id, provider, true, err
		}
		if err = common.Unmarshal(raw, &payload); err != nil {
			return "", credential.Id, provider, true, err
		}
	}
	if provider == UpstreamOAuthProviderCodex {
		key := CodexOAuthKey{IDToken: payload.IDToken, AccessToken: payload.AccessToken, RefreshToken: payload.RefreshToken, AccountID: payload.AccountID, Email: payload.Email, Type: "codex", Expired: time.Unix(payload.ExpiresAt, 0).Format(time.RFC3339)}
		encoded, err := common.Marshal(&key)
		return string(encoded), credential.Id, provider, true, err
	}
	return payload.AccessToken, credential.Id, provider, true, nil
}

func providerForChannelType(channelType int) string {
	providers := providersForChannelType(channelType)
	if len(providers) > 0 {
		return providers[0]
	}
	return ""
}

func providersForChannelType(channelType int) []string {
	switch channelType {
	case constant.ChannelTypeCodex:
		return []string{UpstreamOAuthProviderCodex}
	case constant.ChannelTypeAnthropic:
		return []string{UpstreamOAuthProviderClaude}
	case constant.ChannelTypeGemini:
		return []string{UpstreamOAuthProviderGeminiCLI, UpstreamOAuthProviderAntigravity}
	case constant.ChannelTypeMoonshot:
		return []string{UpstreamOAuthProviderKimi}
	case constant.ChannelTypeXai:
		return []string{UpstreamOAuthProviderXAI}
	default:
		return nil
	}
}

func ProvidersForChannelType(channelType int) []string { return providersForChannelType(channelType) }

func validateUpstreamOAuthChannel(channel *model.Channel, provider string) error {
	if channel == nil {
		return errors.New("channel not found")
	}
	if !slices.Contains(providersForChannelType(channel.Type), provider) {
		return errors.New("OAuth provider does not match channel type")
	}
	return nil
}

func validatePendingOAuthSession(session *model.UpstreamOAuthSession) error {
	if session.Status != model.UpstreamOAuthSessionPending {
		return fmt.Errorf("OAuth session is %s", session.Status)
	}
	if session.ExpiresAt <= time.Now().Unix() {
		_ = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionExpired, "authorization session expired")
		return errors.New("OAuth authorization session expired")
	}
	return nil
}

func saveUpstreamOAuthCredential(channelId int, provider string, payload *UpstreamOAuthTokenPayload) (*model.UpstreamCredential, error) {
	if payload == nil || payload.AccessToken == "" || payload.AccountID == "" {
		return nil, errors.New("OAuth token response is missing account information")
	}
	raw, err := common.Marshal(payload)
	if err != nil {
		return nil, err
	}
	encrypted, err := common.EncryptUpstreamCredential(raw)
	if err != nil {
		return nil, err
	}
	credential := &model.UpstreamCredential{ChannelId: channelId, Provider: provider, AccountId: payload.AccountID, AccountEmail: payload.Email, DisplayName: payload.Email, EncryptedPayload: encrypted, ExpiresAt: payload.ExpiresAt}
	if err = model.UpsertUpstreamCredential(credential); err != nil {
		return nil, err
	}
	return credential, nil
}

func jwtSubject(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Subject string `json:"sub"`
	}
	if common.Unmarshal(raw, &claims) != nil {
		return ""
	}
	return strings.TrimSpace(claims.Subject)
}

func discoverGoogleCodeAssistProject(ctx context.Context, client *http.Client, accessToken string) (string, error) {
	body, err := common.Marshal(map[string]any{"metadata": map[string]string{
		"ideType": "IDE_UNSPECIFIED", "platform": "PLATFORM_UNSPECIFIED", "pluginType": "GEMINI",
	}})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gemini-cli/0.1")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Project string `json:"cloudaicompanionProject"`
	}
	if err = common.DecodeJson(resp.Body, &result); err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Code Assist project discovery failed with status %d", resp.StatusCode)
	}
	project := strings.TrimSpace(result.Project)
	if project == "" {
		return "", errors.New("Code Assist project discovery returned no project")
	}
	return project, nil
}

func buildUpstreamAuthorizationURL(provider string, state string, challenge string) string {
	params := url.Values{"client_id": {codexOAuthClientID}, "response_type": {"code"}, "redirect_uri": {codexOAuthRedirectURI}, "scope": {"openid email profile offline_access"}, "state": {state}, "code_challenge": {challenge}, "code_challenge_method": {"S256"}, "id_token_add_organizations": {"true"}, "codex_cli_simplified_flow": {"true"}}
	endpoint := codexOAuthAuthorizeURL
	if provider == UpstreamOAuthProviderClaude {
		endpoint = claudeOAuthAuthorizeURL
		params = url.Values{"code": {"true"}, "client_id": {claudeOAuthClientID}, "response_type": {"code"}, "redirect_uri": {claudeOAuthRedirectURI}, "scope": {claudeOAuthScope}, "state": {state}, "code_challenge": {challenge}, "code_challenge_method": {"S256"}}
	}
	if provider == UpstreamOAuthProviderAntigravity {
		params = url.Values{"client_id": {antigravityClientID}, "response_type": {"code"}, "redirect_uri": {antigravityRedirectURI}, "scope": {"https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/cclog https://www.googleapis.com/auth/experimentsandconfigs"}, "state": {state}, "access_type": {"offline"}, "prompt": {"consent"}, "code_challenge": {challenge}, "code_challenge_method": {"S256"}}
		endpoint = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	if provider == UpstreamOAuthProviderGeminiCLI {
		params = url.Values{"client_id": {geminiCLIClientID}, "response_type": {"code"}, "redirect_uri": {geminiCLIRedirectURI}, "scope": {"openid email profile https://www.googleapis.com/auth/cloud-platform"}, "state": {state}, "access_type": {"offline"}, "prompt": {"consent"}, "code_challenge": {challenge}, "code_challenge_method": {"S256"}}
		endpoint = geminiCLIAuthorizeURL
	}
	return endpoint + "?" + params.Encode()
}

func exchangeUpstreamOAuthCode(ctx context.Context, channel *model.Channel, provider string, code string, state string, verifier string, redirectOverride string) (*UpstreamOAuthTokenPayload, error) {
	client, err := GetHttpClientWithProxy(channel.GetSetting().Proxy)
	if err != nil {
		return nil, err
	}
	if provider == UpstreamOAuthProviderClaude {
		return exchangeClaudeOAuthCode(ctx, client, code, state, verifier)
	}
	if provider == UpstreamOAuthProviderAntigravity {
		payload, err := requestGoogleOAuthToken(ctx, client, code, verifier, antigravityClientID, upstreamOAuthClientSecret(provider), antigravityRedirectURI, "https://oauth2.googleapis.com/token")
		if err != nil {
			return nil, err
		}
		payload.ProjectID, err = discoverGoogleCodeAssistProject(ctx, client, payload.AccessToken)
		return payload, err
	}
	if provider == UpstreamOAuthProviderGeminiCLI {
		payload, err := requestGoogleOAuthToken(ctx, client, code, verifier, geminiCLIClientID, upstreamOAuthClientSecret(provider), geminiCLIRedirectURI, "https://oauth2.googleapis.com/token")
		if err != nil {
			return nil, err
		}
		payload.ProjectID, err = discoverGoogleCodeAssistProject(ctx, client, payload.AccessToken)
		return payload, err
	}
	redirectURI := codexOAuthRedirectURI
	if redirectOverride != "" {
		redirectURI = redirectOverride
	}
	form := url.Values{"grant_type": {"authorization_code"}, "client_id": {codexOAuthClientID}, "code": {code}, "redirect_uri": {redirectURI}, "code_verifier": {verifier}}
	return requestCodexToken(ctx, client, form)
}

func requestGoogleOAuthToken(ctx context.Context, client *http.Client, code, verifier, clientID, clientSecret, redirectURI, tokenURL string) (*UpstreamOAuthTokenPayload, error) {
	form := url.Values{"code": {code}, "client_id": {clientID}, "redirect_uri": {redirectURI}, "grant_type": {"authorization_code"}, "code_verifier": {verifier}}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err = common.DecodeJson(resp.Body, &token); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.AccessToken == "" {
		return nil, fmt.Errorf("Google OAuth token exchange failed with status %d", resp.StatusCode)
	}
	accountID, email := jwtSubject(token.IDToken), ""
	if accountID == "" {
		return nil, errors.New("OAuth token response has no stable subject")
	}
	if token.IDToken != "" {
		email, _ = ExtractEmailFromJWT(token.IDToken)
	}
	return &UpstreamOAuthTokenPayload{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, IDToken: token.IDToken, TokenType: token.TokenType, Scope: token.Scope, AccountID: accountID, Email: email, AuthKind: "oauth", ExpiresAt: time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix()}, nil
}

func requestCodexToken(ctx context.Context, client *http.Client, form url.Values) (*UpstreamOAuthTokenPayload, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, codexOAuthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err = common.DecodeJson(resp.Body, &token); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.AccessToken == "" {
		return nil, fmt.Errorf("Codex OAuth token exchange failed with status %d", resp.StatusCode)
	}
	accountID, _ := ExtractCodexAccountIDFromJWT(token.IDToken)
	email, _ := ExtractEmailFromJWT(token.IDToken)
	return &UpstreamOAuthTokenPayload{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, IDToken: token.IDToken, TokenType: token.TokenType, Scope: token.Scope, AccountID: accountID, Email: email, ExpiresAt: time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix()}, nil
}

func exchangeClaudeOAuthCode(ctx context.Context, client *http.Client, code string, state string, verifier string) (*UpstreamOAuthTokenPayload, error) {
	parts := strings.SplitN(code, "#", 2)
	code = parts[0]
	if len(parts) == 2 && parts[1] != "" {
		state = parts[1]
	}
	body, err := common.Marshal(struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		RedirectURI  string `json:"redirect_uri"`
		ClientID     string `json:"client_id"`
		CodeVerifier string `json:"code_verifier"`
		State        string `json:"state"`
	}{"authorization_code", code, claudeOAuthRedirectURI, claudeOAuthClientID, verifier, state})
	if err != nil {
		return nil, err
	}
	return requestClaudeToken(ctx, client, body)
}

func requestClaudeToken(ctx context.Context, client *http.Client, body []byte) (*UpstreamOAuthTokenPayload, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeOAuthTokenURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "axios/1.15.2")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Account      struct {
			UUID  string `json:"uuid"`
			Email string `json:"email_address"`
		} `json:"account"`
		Organization struct {
			UUID string `json:"uuid"`
			Name string `json:"name"`
		} `json:"organization"`
	}
	if err = common.DecodeJson(resp.Body, &token); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.AccessToken == "" {
		return nil, fmt.Errorf("Claude OAuth token exchange failed with status %d", resp.StatusCode)
	}
	return &UpstreamOAuthTokenPayload{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, TokenType: token.TokenType, AccountID: token.Account.UUID, Email: token.Account.Email, OrganizationID: token.Organization.UUID, OrganizationName: token.Organization.Name, ExpiresAt: time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix()}, nil
}

func refreshUpstreamOAuthToken(ctx context.Context, channel *model.Channel, provider string, refreshToken string, tokenEndpoint string) (*UpstreamOAuthTokenPayload, error) {
	client, err := GetHttpClientWithProxy(channel.GetSetting().Proxy)
	if err != nil {
		return nil, err
	}
	if provider == UpstreamOAuthProviderClaude {
		body, err := common.Marshal(map[string]string{"client_id": claudeOAuthClientID, "grant_type": "refresh_token", "refresh_token": refreshToken, "scope": claudeOAuthScope})
		if err != nil {
			return nil, err
		}
		return requestClaudeToken(ctx, client, body)
	}
	if provider == UpstreamOAuthProviderAntigravity {
		return refreshGoogleOAuthToken(ctx, client, refreshToken, antigravityClientID, upstreamOAuthClientSecret(provider))
	}
	if provider == UpstreamOAuthProviderGeminiCLI {
		return refreshGoogleOAuthToken(ctx, client, refreshToken, geminiCLIClientID, upstreamOAuthClientSecret(provider))
	}
	if provider == UpstreamOAuthProviderKimi {
		return refreshGenericOAuthToken(ctx, client, refreshToken, kimiTokenURL, kimiClientID)
	}
	if provider == UpstreamOAuthProviderXAI {
		if strings.TrimSpace(tokenEndpoint) == "" {
			tokenEndpoint = "https://auth.x.ai/oauth2/token"
		}
		return refreshGenericOAuthToken(ctx, client, refreshToken, tokenEndpoint, xaiClientID)
	}
	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}, "client_id": {codexOAuthClientID}, "scope": {"openid profile email"}}
	return requestCodexToken(ctx, client, form)
}

func refreshGoogleOAuthToken(ctx context.Context, client *http.Client, refreshToken, clientID, clientSecret string) (*UpstreamOAuthTokenPayload, error) {
	return refreshGenericOAuthTokenWithSecret(ctx, client, refreshToken, "https://oauth2.googleapis.com/token", clientID, clientSecret)
}

func refreshGenericOAuthToken(ctx context.Context, client *http.Client, refreshToken, tokenURL, clientID string) (*UpstreamOAuthTokenPayload, error) {
	return refreshGenericOAuthTokenWithSecret(ctx, client, refreshToken, tokenURL, clientID, "")
}

func refreshGenericOAuthTokenWithSecret(ctx context.Context, client *http.Client, refreshToken, tokenURL, clientID, clientSecret string) (*UpstreamOAuthTokenPayload, error) {
	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}, "client_id": {clientID}}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err = common.DecodeJson(resp.Body, &token); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.AccessToken == "" {
		return nil, fmt.Errorf("OAuth refresh failed with status %d", resp.StatusCode)
	}
	accountID, email := jwtSubject(token.IDToken), ""
	if token.IDToken != "" {
		email, _ = ExtractEmailFromJWT(token.IDToken)
	}
	return &UpstreamOAuthTokenPayload{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, IDToken: token.IDToken, TokenType: token.TokenType, Scope: token.Scope, AccountID: accountID, Email: email, AuthKind: "oauth", ExpiresAt: time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix()}, nil
}

func startCodexDeviceAuthorization(ctx context.Context, client *http.Client) (*upstreamDeviceMetadata, int, error) {
	body, err := common.Marshal(map[string]string{"client_id": codexOAuthClientID})
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, codexDeviceUserCodeURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var result struct {
		DeviceAuthID string `json:"device_auth_id"`
		UserCode     string `json:"user_code"`
		UserCodeAlt  string `json:"usercode"`
		Interval     any    `json:"interval"`
	}
	if err = common.DecodeJson(resp.Body, &result); err != nil {
		return nil, 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf("Codex device authorization failed with status %d", resp.StatusCode)
	}
	if result.UserCode == "" {
		result.UserCode = result.UserCodeAlt
	}
	if result.DeviceAuthID == "" || result.UserCode == "" {
		return nil, 0, errors.New("Codex device authorization response is incomplete")
	}
	interval := 5
	switch value := result.Interval.(type) {
	case float64:
		if int(value) > 0 {
			interval = int(value)
		}
	case string:
		if parsed, parseErr := time.ParseDuration(value + "s"); parseErr == nil && parsed > 0 {
			interval = int(parsed.Seconds())
		}
	}
	return &upstreamDeviceMetadata{DeviceAuthID: result.DeviceAuthID, UserCode: result.UserCode}, interval, nil
}

func startGenericDeviceAuthorization(ctx context.Context, client *http.Client, provider string) (*upstreamDeviceMetadata, int, error) {
	metadata := &upstreamDeviceMetadata{Provider: provider}
	interval := 5
	if provider == UpstreamOAuthProviderKimi {
		metadata.DeviceID = uuid.NewString()
		form := url.Values{"client_id": {kimiClientID}}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, kimiDeviceCodeURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		setKimiOAuthHeaders(req.Header, metadata.DeviceID)
		resp, err := client.Do(req)
		if err != nil {
			return nil, 0, err
		}
		defer resp.Body.Close()
		var result struct {
			DeviceCode      string `json:"device_code"`
			UserCode        string `json:"user_code"`
			VerificationURI string `json:"verification_uri"`
			ExpiresIn       int    `json:"expires_in"`
			Interval        int    `json:"interval"`
		}
		if err = common.DecodeJson(resp.Body, &result); err != nil {
			return nil, 0, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 || result.DeviceCode == "" || result.UserCode == "" {
			return nil, 0, fmt.Errorf("Kimi device authorization failed with status %d", resp.StatusCode)
		}
		metadata.DeviceCode, metadata.UserCode, metadata.TokenURL, metadata.ClientID, metadata.Scope = result.DeviceCode, result.UserCode, kimiTokenURL, kimiClientID, "openid profile email offline_access"
		if result.Interval > 0 {
			interval = result.Interval
		}
		return metadata, interval, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, xaiDiscoveryURL, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	var discovery struct {
		DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
		TokenEndpoint               string `json:"token_endpoint"`
	}
	decodeErr := common.DecodeJson(resp.Body, &discovery)
	resp.Body.Close()
	if decodeErr != nil {
		return nil, 0, decodeErr
	}
	if discovery.DeviceAuthorizationEndpoint == "" || discovery.TokenEndpoint == "" {
		return nil, 0, errors.New("xAI OAuth discovery response is incomplete")
	}
	form := url.Values{"client_id": {xaiClientID}, "scope": {"openid profile email offline_access grok-cli:access api:access"}}
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, discovery.DeviceAuthorizationEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if metadata.Provider == UpstreamOAuthProviderKimi {
		setKimiOAuthHeaders(req.Header, metadata.DeviceID)
	}
	resp, err = client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var result struct {
		DeviceCode string `json:"device_code"`
		UserCode   string `json:"user_code"`
		Interval   int    `json:"interval"`
	}
	if err = common.DecodeJson(resp.Body, &result); err != nil {
		return nil, 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || result.DeviceCode == "" || result.UserCode == "" {
		return nil, 0, fmt.Errorf("xAI device authorization failed with status %d", resp.StatusCode)
	}
	metadata.DeviceCode, metadata.UserCode, metadata.TokenURL, metadata.ClientID, metadata.Scope = result.DeviceCode, result.UserCode, discovery.TokenEndpoint, xaiClientID, "openid profile email offline_access grok-cli:access api:access"
	if result.Interval > 0 {
		interval = result.Interval
	}
	return metadata, interval, nil
}

func setKimiOAuthHeaders(header http.Header, deviceID string) {
	header.Set("Accept", "application/json")
	header.Set("X-Msh-Platform", "new-api")
	header.Set("X-Msh-Version", "1")
	header.Set("X-Msh-Device-Name", "new-api")
	header.Set("X-Msh-Device-Model", "server")
	header.Set("X-Msh-Device-Id", deviceID)
}

func pollGenericDeviceAuthorization(ctx context.Context, client *http.Client, metadata upstreamDeviceMetadata) (*UpstreamOAuthTokenPayload, error) {
	form := url.Values{"device_code": {metadata.DeviceCode}, "client_id": {metadata.ClientID}, "grant_type": {"urn:ietf:params:oauth:grant-type:device_code"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metadata.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
	}
	if err = common.DecodeJson(resp.Body, &token); err != nil {
		return nil, err
	}
	if token.Error == "authorization_pending" || token.Error == "slow_down" || resp.StatusCode == http.StatusForbidden {
		return nil, ErrUpstreamOAuthPending
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.AccessToken == "" {
		return nil, fmt.Errorf("%s OAuth token exchange failed with status %d", metadata.Provider, resp.StatusCode)
	}
	accountID, email := jwtSubject(token.IDToken), ""
	if accountID == "" && metadata.Provider == UpstreamOAuthProviderKimi {
		accountID = metadata.DeviceID
	}
	if accountID == "" {
		return nil, errors.New("OAuth device token response has no stable account identity")
	}
	if token.IDToken != "" {
		email, _ = ExtractEmailFromJWT(token.IDToken)
	}
	baseURL := ""
	if metadata.Provider == UpstreamOAuthProviderKimi {
		baseURL = "https://api.kimi.com/coding"
	}
	if metadata.Provider == UpstreamOAuthProviderXAI {
		baseURL = "https://cli-chat-proxy.grok.com/v1"
	}
	return &UpstreamOAuthTokenPayload{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, IDToken: token.IDToken, TokenType: token.TokenType, Scope: token.Scope, AccountID: accountID, Email: email, DeviceID: metadata.DeviceID, BaseURL: baseURL, TokenEndpoint: metadata.TokenURL, AuthKind: "oauth", ExpiresAt: time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix()}, nil
}

func oauthMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func pollCodexDeviceAuthorization(ctx context.Context, client *http.Client, metadata upstreamDeviceMetadata) (string, string, error) {
	body, err := common.Marshal(map[string]string{"device_auth_id": metadata.DeviceAuthID, "user_code": metadata.UserCode})
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, codexDeviceTokenURL, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
		return "", "", ErrUpstreamOAuthPending
	}
	var result struct {
		AuthorizationCode string `json:"authorization_code"`
		CodeVerifier      string `json:"code_verifier"`
	}
	if err = common.DecodeJson(resp.Body, &result); err != nil {
		return "", "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("Codex device polling failed with status %d", resp.StatusCode)
	}
	if result.AuthorizationCode == "" || result.CodeVerifier == "" {
		return "", "", errors.New("Codex device token response is incomplete")
	}
	return result.AuthorizationCode, result.CodeVerifier, nil
}

func parseOAuthCallbackInput(input string) (code string, state string, oauthError string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", "", errors.New("callback URL or authorization code is required")
	}
	parsed, parseErr := url.Parse(input)
	if parseErr == nil && parsed.Scheme != "" {
		values := parsed.Query()
		return values.Get("code"), values.Get("state"), values.Get("error"), nil
	}
	return "", "", "", errors.New("paste the complete OAuth callback URL")
}

func generatePKCE() (string, string, error) {
	verifier, err := randomURLToken(64)
	if err != nil {
		return "", "", err
	}
	digest := sha256.Sum256([]byte(verifier))
	return verifier, base64.RawURLEncoding.EncodeToString(digest[:]), nil
}

func randomURLToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashOAuthState(state string) string {
	digest := sha256.Sum256([]byte(state))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func sanitizeUpstreamOAuthError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}
