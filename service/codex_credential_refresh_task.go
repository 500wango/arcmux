package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/constant"
	"github.com/500wango/arcmux/logger"
	"github.com/500wango/arcmux/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	codexCredentialRefreshTickInterval = 10 * time.Minute
	codexCredentialRefreshThreshold    = 24 * time.Hour
	codexCredentialRefreshBatchSize    = 200
	codexCredentialRefreshTimeout      = 15 * time.Second
)

var (
	codexCredentialRefreshOnce    sync.Once
	codexCredentialRefreshRunning atomic.Bool
)

func shouldAutoRefreshCodexChannelStatus(status int) bool {
	return status == common.ChannelStatusEnabled || status == common.ChannelStatusAutoDisabled
}

func StartCodexCredentialAutoRefreshTask() {
	codexCredentialRefreshOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}

		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("codex credential auto-refresh task started: tick=%s threshold=%s", codexCredentialRefreshTickInterval, codexCredentialRefreshThreshold))
			migrateLegacyCodexOAuthCredentials()

			ticker := time.NewTicker(codexCredentialRefreshTickInterval)
			defer ticker.Stop()

			runCodexCredentialAutoRefreshOnce()
			for range ticker.C {
				runCodexCredentialAutoRefreshOnce()
			}
		})
	})
}

func runCodexCredentialAutoRefreshOnce() {
	if !codexCredentialRefreshRunning.CompareAndSwap(false, true) {
		return
	}
	defer codexCredentialRefreshRunning.Store(false)

	ctx := context.Background()
	now := time.Now()
	runUpstreamOAuthCredentialRefresh(ctx, now)

	var refreshed int
	var scanned int

	offset := 0
	for {
		var channels []*model.Channel
		err := model.DB.
			Select("id", "name", "key", "status", "channel_info").
			Where("type = ? AND (status = ? OR status = ?)",
				constant.ChannelTypeCodex,
				common.ChannelStatusEnabled,
				common.ChannelStatusAutoDisabled,
			).
			Order("id asc").
			Limit(codexCredentialRefreshBatchSize).
			Offset(offset).
			Find(&channels).Error
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("codex credential auto-refresh: query channels failed: %v", err))
			return
		}
		if len(channels) == 0 {
			break
		}
		offset += codexCredentialRefreshBatchSize

		for _, ch := range channels {
			if ch == nil {
				continue
			}
			scanned++
			if ch.ChannelInfo.IsMultiKey {
				continue
			}

			rawKey := strings.TrimSpace(ch.Key)
			if rawKey == "" {
				continue
			}

			oauthKey, err := parseCodexOAuthKey(rawKey)
			if err != nil {
				continue
			}

			refreshToken := strings.TrimSpace(oauthKey.RefreshToken)
			if refreshToken == "" {
				continue
			}

			expiredAtRaw := strings.TrimSpace(oauthKey.Expired)
			expiredAt, err := time.Parse(time.RFC3339, expiredAtRaw)
			if err == nil && !expiredAt.IsZero() && expiredAt.Sub(now) > codexCredentialRefreshThreshold {
				continue
			}

			refreshCtx, cancel := context.WithTimeout(ctx, codexCredentialRefreshTimeout)
			newKey, _, err := RefreshCodexChannelCredential(refreshCtx, ch.Id, CodexCredentialRefreshOptions{ResetCaches: false})
			cancel()
			if err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("codex credential auto-refresh: channel_id=%d name=%s refresh failed: %v", ch.Id, ch.Name, err))
				continue
			}

			refreshed++
			logger.LogInfo(ctx, fmt.Sprintf("codex credential auto-refresh: channel_id=%d name=%s refreshed, expires_at=%s", ch.Id, ch.Name, newKey.Expired))
		}
	}

	if refreshed > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.LogWarn(ctx, fmt.Sprintf("codex credential auto-refresh: InitChannelCache panic: %v", r))
				}
			}()
			model.InitChannelCache()
		}()
	}

	if common.DebugEnabled {
		logger.LogDebug(ctx, "codex credential auto-refresh: scanned=%d refreshed=%d", scanned, refreshed)
	}
}

func migrateLegacyCodexOAuthCredentials() {
	if !common.UpstreamCredentialEncryptionConfigured() {
		return
	}
	var channels []*model.Channel
	if err := model.DB.Select("id", "key").Where("type = ?", constant.ChannelTypeCodex).Find(&channels).Error; err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("legacy Codex OAuth migration query failed: %v", err))
		return
	}
	for _, channel := range channels {
		if channel == nil || strings.TrimSpace(channel.Key) == "" {
			continue
		}
		key, err := parseCodexOAuthKey(channel.Key)
		if err != nil || strings.TrimSpace(key.AccessToken) == "" || strings.TrimSpace(key.AccountID) == "" {
			continue
		}
		var existing int64
		if err = model.DB.Model(&model.UpstreamCredential{}).
			Where("channel_id = ? AND provider = ? AND account_id = ?", channel.Id, UpstreamOAuthProviderCodex, key.AccountID).
			Count(&existing).Error; err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("legacy Codex OAuth migration lookup failed for channel_id=%d: %v", channel.Id, err))
			continue
		}
		if existing > 0 {
			continue
		}
		expiresAt := int64(0)
		if parsed, parseErr := time.Parse(time.RFC3339, key.Expired); parseErr == nil {
			expiresAt = parsed.Unix()
		}
		_, err = saveUpstreamOAuthCredential(channel.Id, UpstreamOAuthProviderCodex, &UpstreamOAuthTokenPayload{
			AccessToken: key.AccessToken, RefreshToken: key.RefreshToken, IDToken: key.IDToken,
			AccountID: key.AccountID, Email: key.Email, ExpiresAt: expiresAt,
		})
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("legacy Codex OAuth migration failed for channel_id=%d: %v", channel.Id, err))
		}
	}
}

func runUpstreamOAuthCredentialRefresh(ctx context.Context, now time.Time) {
	if !common.UpstreamCredentialEncryptionConfigured() {
		return
	}
	var credentials []*model.UpstreamCredential
	if err := model.DB.Where("status = ? AND expires_at > 0 AND expires_at <= ?", model.UpstreamCredentialEnabled, now.Add(codexCredentialRefreshThreshold).Unix()).Order("id asc").Limit(codexCredentialRefreshBatchSize).Find(&credentials).Error; err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("upstream OAuth credential refresh query failed: %v", err))
		return
	}
	for _, credential := range credentials {
		if credential == nil {
			continue
		}
		raw, decryptErr := common.DecryptUpstreamCredential(credential.EncryptedPayload)
		if decryptErr != nil {
			logger.LogWarn(ctx, fmt.Sprintf("upstream OAuth credential payload decrypt failed: credential_id=%d provider=%s: %v", credential.Id, credential.Provider, decryptErr))
			continue
		}
		var payload UpstreamOAuthTokenPayload
		if unmarshalErr := common.Unmarshal(raw, &payload); unmarshalErr != nil {
			logger.LogWarn(ctx, fmt.Sprintf("upstream OAuth credential payload decode failed: credential_id=%d provider=%s: %v", credential.Id, credential.Provider, unmarshalErr))
			continue
		}
		// An imported access-token-only credential must remain usable without
		// starting a refresh/login flow when its access token nears expiry.
		if strings.TrimSpace(payload.RefreshToken) == "" {
			continue
		}
		refreshCtx, cancel := context.WithTimeout(ctx, codexCredentialRefreshTimeout)
		err := RefreshUpstreamCredential(refreshCtx, credential)
		cancel()
		if err != nil {
			_ = model.MarkUpstreamCredentialFailure(credential.Id, err.Error(), now.Add(time.Minute).Unix())
			logger.LogWarn(ctx, fmt.Sprintf("upstream OAuth credential refresh failed: credential_id=%d provider=%s: %v", credential.Id, credential.Provider, err))
		}
	}
}
