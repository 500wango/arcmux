package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/model"
)

const (
	MaxUpstreamOAuthImportBytes    = 2 << 20
	MaxUpstreamOAuthImportAccounts = 500
)

type UpstreamOAuthImportResult struct {
	Imported int `json:"imported"`
}

func ImportUpstreamOAuthCredentials(ctx context.Context, channelId int, provider string, content string, commercialAcknowledged bool) (*UpstreamOAuthImportResult, error) {
	if !commercialAcknowledged {
		return nil, errors.New("commercial-use policy acknowledgement is required")
	}
	if !common.UpstreamCredentialEncryptionConfigured() {
		return nil, errors.New("UPSTREAM_CREDENTIAL_ENCRYPTION_KEY is not configured")
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
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

	payloads, err := parseUpstreamOAuthImport(content)
	if err != nil {
		return nil, err
	}
	credentialsByAccount := make(map[string]*model.UpstreamCredential, len(payloads))
	accountOrder := make([]string, 0, len(payloads))
	for index, payload := range payloads {
		if payload.AccessToken == "" {
			refreshed, refreshErr := refreshUpstreamOAuthToken(ctx, channel, provider, payload.RefreshToken, payload.TokenEndpoint)
			if refreshErr != nil {
				return nil, fmt.Errorf("OAuth credential %d refresh failed: %w", index+1, refreshErr)
			}
			mergeRotatedUpstreamOAuthToken(refreshed, payload)
			payload = refreshed
		}
		normalizeImportedUpstreamOAuthPayload(provider, payload)
		if payload.AccessToken == "" || payload.AccountID == "" {
			return nil, fmt.Errorf("OAuth credential %d is missing access_token or stable account identity", index+1)
		}
		raw, marshalErr := common.Marshal(payload)
		if marshalErr != nil {
			return nil, marshalErr
		}
		encrypted, encryptErr := common.EncryptUpstreamCredential(raw)
		if encryptErr != nil {
			return nil, encryptErr
		}
		credential := &model.UpstreamCredential{
			ChannelId:        channelId,
			Provider:         provider,
			AccountId:        payload.AccountID,
			AccountEmail:     payload.Email,
			DisplayName:      payload.Email,
			EncryptedPayload: encrypted,
			ExpiresAt:        payload.ExpiresAt,
		}
		if _, exists := credentialsByAccount[payload.AccountID]; !exists {
			accountOrder = append(accountOrder, payload.AccountID)
		}
		credentialsByAccount[payload.AccountID] = credential
	}

	credentials := make([]*model.UpstreamCredential, 0, len(accountOrder))
	for _, accountID := range accountOrder {
		credentials = append(credentials, credentialsByAccount[accountID])
	}
	if err = model.UpsertUpstreamCredentials(credentials); err != nil {
		return nil, err
	}
	return &UpstreamOAuthImportResult{Imported: len(credentials)}, nil
}

func parseUpstreamOAuthImport(content string) ([]*UpstreamOAuthTokenPayload, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("OAuth credential JSON is empty")
	}
	if len(content) > MaxUpstreamOAuthImportBytes {
		return nil, errors.New("OAuth credential JSON exceeds 2 MiB")
	}
	var root json.RawMessage = []byte(content)
	items := make([]*UpstreamOAuthTokenPayload, 0)
	var collect func(json.RawMessage) error
	collect = func(raw json.RawMessage) error {
		if len(items) >= MaxUpstreamOAuthImportAccounts {
			return fmt.Errorf("OAuth credential JSON exceeds %d accounts", MaxUpstreamOAuthImportAccounts)
		}
		switch common.GetJsonType(raw) {
		case "array":
			var entries []json.RawMessage
			if err := common.Unmarshal(raw, &entries); err != nil {
				return err
			}
			for _, entry := range entries {
				if err := collect(entry); err != nil {
					return err
				}
			}
			return nil
		case "object":
			var record map[string]json.RawMessage
			if err := common.Unmarshal(raw, &record); err != nil {
				return err
			}
			for _, key := range []string{"credentials", "accounts", "items"} {
				if nested, ok := record[key]; ok {
					return collect(nested)
				}
			}
			payload, err := decodeUpstreamOAuthImportRecord(record)
			if err != nil {
				return err
			}
			items = append(items, payload)
			return nil
		default:
			return errors.New("OAuth credential JSON must be an object or array")
		}
	}
	if err := collect(root); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("OAuth credential JSON contains no accounts")
	}
	return items, nil
}

func decodeUpstreamOAuthImportRecord(record map[string]json.RawMessage) (*UpstreamOAuthTokenPayload, error) {
	payload := &UpstreamOAuthTokenPayload{}
	// Codex auth.json files have appeared with token data under tokens, auth,
	// oauth, or credential. Merge those objects first, then let top-level
	// fields override them. This keeps imports compatible with CLI variants.
	for _, nestedKey := range []string{"tokens", "auth", "oauth", "credential", "token"} {
		nested, ok := record[nestedKey]
		if !ok || common.GetJsonType(nested) != "object" {
			continue
		}
		var nestedRecord map[string]json.RawMessage
		if err := common.Unmarshal(nested, &nestedRecord); err != nil {
			return nil, err
		}
		decoded, err := decodeUpstreamOAuthImportRecord(nestedRecord)
		if err != nil {
			return nil, err
		}
		mergeImportedOAuthPayload(payload, decoded)
	}

	readString := func(keys ...string) string {
		for _, key := range keys {
			raw, ok := record[key]
			if !ok {
				continue
			}
			var value string
			if common.Unmarshal(raw, &value) == nil && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
		return ""
	}
	assign := func(target *string, keys ...string) {
		if value := readString(keys...); value != "" {
			*target = value
		}
	}
	assign(&payload.AccessToken, "access_token", "accessToken")
	assign(&payload.RefreshToken, "refresh_token", "refreshToken")
	assign(&payload.IDToken, "id_token", "idToken")
	assign(&payload.TokenType, "token_type", "tokenType")
	assign(&payload.Scope, "scope")
	assign(&payload.AccountID, "account_id", "accountId", "chatgpt_account_id", "chatgptAccountId")
	assign(&payload.Email, "email", "account_email", "accountEmail")
	assign(&payload.OrganizationID, "organization_id", "organizationId")
	assign(&payload.OrganizationName, "organization_name", "organizationName")
	assign(&payload.DeviceID, "device_id", "deviceId")
	assign(&payload.ProjectID, "project_id", "projectId")
	assign(&payload.BaseURL, "base_url", "baseUrl")
	assign(&payload.TokenEndpoint, "token_endpoint", "tokenEndpoint")
	assign(&payload.AuthKind, "auth_kind", "authKind")

	for _, key := range []string{"expires_at", "expiresAt", "expired"} {
		raw, ok := record[key]
		if !ok {
			continue
		}
		if expiresAt, ok := parseUpstreamOAuthImportExpiry(raw); ok {
			payload.ExpiresAt = expiresAt
			break
		}
	}
	if payload.ExpiresAt == 0 {
		for _, key := range []string{"expires_in", "expiresIn"} {
			raw, ok := record[key]
			if !ok {
				continue
			}
			var seconds int64
			if common.Unmarshal(raw, &seconds) == nil && seconds > 0 {
				payload.ExpiresAt = time.Now().Add(time.Duration(seconds) * time.Second).Unix()
				break
			}
		}
	}
	if payload.AccessToken == "" && payload.RefreshToken == "" {
		return nil, errors.New("OAuth credential must contain access_token or refresh_token")
	}
	return payload, nil
}

func mergeImportedOAuthPayload(dst *UpstreamOAuthTokenPayload, src *UpstreamOAuthTokenPayload) {
	if dst.AccessToken == "" {
		dst.AccessToken = src.AccessToken
	}
	if dst.RefreshToken == "" {
		dst.RefreshToken = src.RefreshToken
	}
	if dst.IDToken == "" {
		dst.IDToken = src.IDToken
	}
	if dst.TokenType == "" {
		dst.TokenType = src.TokenType
	}
	if dst.Scope == "" {
		dst.Scope = src.Scope
	}
	if dst.AccountID == "" {
		dst.AccountID = src.AccountID
	}
	if dst.Email == "" {
		dst.Email = src.Email
	}
	if dst.OrganizationID == "" {
		dst.OrganizationID = src.OrganizationID
	}
	if dst.OrganizationName == "" {
		dst.OrganizationName = src.OrganizationName
	}
	if dst.DeviceID == "" {
		dst.DeviceID = src.DeviceID
	}
	if dst.ProjectID == "" {
		dst.ProjectID = src.ProjectID
	}
	if dst.BaseURL == "" {
		dst.BaseURL = src.BaseURL
	}
	if dst.TokenEndpoint == "" {
		dst.TokenEndpoint = src.TokenEndpoint
	}
	if dst.AuthKind == "" {
		dst.AuthKind = src.AuthKind
	}
	if dst.ExpiresAt == 0 {
		dst.ExpiresAt = src.ExpiresAt
	}
}

func parseUpstreamOAuthImportExpiry(raw json.RawMessage) (int64, bool) {
	var timestamp int64
	if common.Unmarshal(raw, &timestamp) == nil && timestamp > 0 {
		if timestamp > 1_000_000_000_000 {
			timestamp /= 1000
		}
		return timestamp, true
	}
	var value string
	if common.Unmarshal(raw, &value) != nil {
		return 0, false
	}
	value = strings.TrimSpace(value)
	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil && parsed > 0 {
		if parsed > 1_000_000_000_000 {
			parsed /= 1000
		}
		return parsed, true
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed.Unix(), err == nil
}

func normalizeImportedUpstreamOAuthPayload(provider string, payload *UpstreamOAuthTokenPayload) {
	if payload.AccountID == "" && provider == UpstreamOAuthProviderCodex {
		payload.AccountID, _ = ExtractCodexAccountIDFromJWT(payload.IDToken)
		if payload.AccountID == "" {
			payload.AccountID, _ = ExtractCodexAccountIDFromJWT(payload.AccessToken)
		}
	}
	if payload.AccountID == "" {
		payload.AccountID = jwtSubject(payload.IDToken)
	}
	if payload.AccountID == "" {
		payload.AccountID = jwtSubject(payload.AccessToken)
	}
	if payload.Email == "" {
		payload.Email, _ = ExtractEmailFromJWT(payload.IDToken)
	}
	if payload.Email == "" {
		payload.Email, _ = ExtractEmailFromJWT(payload.AccessToken)
	}
	if payload.AccountID == "" {
		payload.AccountID = payload.Email
	}
	if payload.ExpiresAt == 0 {
		for _, token := range []string{payload.AccessToken, payload.IDToken} {
			if claims, ok := decodeJWTClaims(token); ok {
				if expiresAt, ok := claims["exp"].(float64); ok && expiresAt > 0 {
					payload.ExpiresAt = int64(expiresAt)
					break
				}
			}
		}
	}
	if payload.AuthKind == "" {
		payload.AuthKind = "oauth"
	}
	if provider == UpstreamOAuthProviderKimi && payload.BaseURL == "" {
		payload.BaseURL = "https://api.kimi.com/coding"
	}
	if provider == UpstreamOAuthProviderXAI {
		if payload.BaseURL == "" {
			payload.BaseURL = "https://cli-chat-proxy.grok.com/v1"
		}
		if payload.TokenEndpoint == "" {
			payload.TokenEndpoint = "https://auth.x.ai/oauth2/token"
		}
	}
}
