package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/model"
	"github.com/500wango/arcmux/service"

	"github.com/gin-gonic/gin"
)

type startUpstreamOAuthRequest struct {
	Provider               string `json:"provider"`
	FlowType               string `json:"flow_type"`
	CommercialAcknowledged bool   `json:"commercial_acknowledged"`
}

type completeUpstreamOAuthRequest struct {
	CallbackURL string `json:"callback_url"`
}

type updateUpstreamCredentialRequest struct {
	Enabled bool `json:"enabled"`
}

type importUpstreamOAuthRequest struct {
	Provider               string `json:"provider"`
	Content                string `json:"content"`
	CommercialAcknowledged bool   `json:"commercial_acknowledged"`
}

const maxUpstreamOAuthImportRequestBytes = service.MaxUpstreamOAuthImportBytes + 1<<20

func StartChannelUpstreamOAuth(c *gin.Context) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return
	}
	var request startUpstreamOAuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	result, err := service.StartUpstreamOAuth(ctx, c.GetInt("id"), channelId, request.Provider, request.FlowType, request.CommercialAcknowledged)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_start", map[string]interface{}{"id": channelId, "provider": result.Provider, "flow": result.FlowType})
	common.ApiSuccess(c, result)
}

func CompleteChannelUpstreamOAuth(c *gin.Context) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return
	}
	var request completeUpstreamOAuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	credential, err := service.CompleteUpstreamOAuth(ctx, c.GetInt("id"), channelId, c.Param("session_id"), request.CallbackURL)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_complete", map[string]interface{}{"id": channelId, "provider": credential.Provider, "credential_id": credential.Id})
	common.ApiSuccess(c, credential)
}

func PollChannelUpstreamOAuth(c *gin.Context) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	credential, err := service.PollUpstreamOAuth(ctx, c.GetInt("id"), channelId, c.Param("session_id"))
	if errors.Is(err, service.ErrUpstreamOAuthPending) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": model.UpstreamOAuthSessionPending}})
		return
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_complete", map[string]interface{}{"id": channelId, "provider": credential.Provider, "credential_id": credential.Id})
	common.ApiSuccess(c, gin.H{"status": model.UpstreamOAuthSessionCompleted, "credential": credential})
}

func GetChannelUpstreamOAuthSession(c *gin.Context) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return
	}
	session, err := model.GetUpstreamOAuthSession(c.Param("session_id"), c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if session.ChannelId != channelId {
		common.ApiError(c, errors.New("OAuth session does not belong to channel"))
		return
	}
	if session.Status == model.UpstreamOAuthSessionPending && session.ExpiresAt <= time.Now().Unix() {
		_ = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionExpired, "authorization session expired")
		session.Status = model.UpstreamOAuthSessionExpired
	}
	common.ApiSuccess(c, session)
}

func CancelChannelUpstreamOAuth(c *gin.Context) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return
	}
	session, err := model.GetUpstreamOAuthSession(c.Param("session_id"), c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if session.ChannelId != channelId {
		common.ApiError(c, errors.New("OAuth session does not belong to channel"))
		return
	}
	if err = model.UpdateUpstreamOAuthSession(session.Id, model.UpstreamOAuthSessionCancelled, "cancelled by administrator"); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_cancel", map[string]interface{}{"id": channelId, "provider": session.Provider})
	common.ApiSuccess(c, nil)
}

func ListChannelUpstreamCredentials(c *gin.Context) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return
	}
	if _, err := model.GetChannelById(channelId, false); err != nil {
		common.ApiError(c, err)
		return
	}
	credentials, err := model.ListUpstreamCredentials(channelId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"items": credentials, "encryption_configured": common.UpstreamCredentialEncryptionConfigured(),
		"enabled_providers":   service.EnabledUpstreamOAuthProviders(),
		"import_max_bytes":    service.MaxUpstreamOAuthImportBytes,
		"import_max_accounts": service.MaxUpstreamOAuthImportAccounts,
	})
}

func ImportChannelUpstreamCredentials(c *gin.Context) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUpstreamOAuthImportRequestBytes)
	var request importUpstreamOAuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()
	result, err := service.ImportUpstreamOAuthCredentials(ctx, channelId, request.Provider, request.Content, request.CommercialAcknowledged)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_import", map[string]interface{}{"id": channelId, "provider": request.Provider, "count": result.Imported})
	common.ApiSuccess(c, result)
}

func RefreshChannelUpstreamCredential(c *gin.Context) {
	channelId, credentialId, ok := parseCredentialIds(c)
	if !ok {
		return
	}
	credential, err := model.GetUpstreamCredential(credentialId, channelId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	if err = service.RefreshUpstreamCredential(ctx, credential); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_refresh", map[string]interface{}{"id": channelId, "provider": credential.Provider, "credential_id": credential.Id})
	common.ApiSuccess(c, nil)
}

func UpdateChannelUpstreamCredential(c *gin.Context) {
	channelId, credentialId, ok := parseCredentialIds(c)
	if !ok {
		return
	}
	var request updateUpstreamCredentialRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	status := model.UpstreamCredentialDisabled
	reason := "disabled by administrator"
	if request.Enabled {
		status = model.UpstreamCredentialEnabled
		reason = ""
	}
	if err := model.UpdateUpstreamCredentialStatus(credentialId, channelId, status, reason); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_status", map[string]interface{}{"id": channelId, "credential_id": credentialId, "enabled": request.Enabled})
	common.ApiSuccess(c, nil)
}

func DeleteChannelUpstreamCredential(c *gin.Context) {
	channelId, credentialId, ok := parseCredentialIds(c)
	if !ok {
		return
	}
	credential, err := model.GetUpstreamCredential(credentialId, channelId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err = model.DeleteUpstreamCredential(credentialId, channelId); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "channel.oauth_delete", map[string]interface{}{"id": channelId, "provider": credential.Provider, "credential_id": credentialId})
	common.ApiSuccess(c, nil)
}

func parseChannelId(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param("id")))
	if err != nil || id <= 0 {
		common.ApiError(c, errors.New("invalid channel id"))
		return 0, false
	}
	return id, true
}

func parseCredentialIds(c *gin.Context) (int, int64, bool) {
	channelId, ok := parseChannelId(c)
	if !ok {
		return 0, 0, false
	}
	credentialId, err := strconv.ParseInt(strings.TrimSpace(c.Param("credential_id")), 10, 64)
	if err != nil || credentialId <= 0 {
		common.ApiError(c, errors.New("invalid credential id"))
		return 0, 0, false
	}
	return channelId, credentialId, true
}
