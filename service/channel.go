package service

import (
	"fmt"
	"strings"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/model"
	"github.com/500wango/arcmux/relaykit/dto"
	"github.com/500wango/arcmux/relaykit/types"
	"github.com/500wango/arcmux/setting/operation_setting"
)

func formatNotifyType(channelId int, status int) string {
	return fmt.Sprintf("%s_%d_%d", dto.NotifyTypeChannelUpdate, channelId, status)
}

// disable & notify
func DisableChannel(channelError types.ChannelError, reason string) {
	common.SysLog(fmt.Sprintf("通道「%s」（#%d）发生错误，准备禁用，原因：%s", channelError.ChannelName, channelError.ChannelId, common.LocalLogPreview(reason)))

	// 检查是否启用自动禁用功能
	if !channelError.AutoBan {
		common.SysLog(fmt.Sprintf("通道「%s」（#%d）未启用自动禁用功能，跳过禁用操作", channelError.ChannelName, channelError.ChannelId))
		return
	}

	success := model.UpdateChannelStatus(channelError.ChannelId, channelError.UsingKey, common.ChannelStatusAutoDisabled, reason)
	if success {
		subject := fmt.Sprintf("通道「%s」（#%d）已被禁用", channelError.ChannelName, channelError.ChannelId)
		content := fmt.Sprintf("通道「%s」（#%d）已被禁用，原因：%s", channelError.ChannelName, channelError.ChannelId, reason)
		NotifyRootUser(formatNotifyType(channelError.ChannelId, common.ChannelStatusAutoDisabled), subject, content)
	}
}

func EnableChannel(channelId int, usingKey string, channelName string) {
	success := model.UpdateChannelStatus(channelId, usingKey, common.ChannelStatusEnabled, "")
	if success {
		subject := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
		content := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
		NotifyRootUser(formatNotifyType(channelId, common.ChannelStatusEnabled), subject, content)
	}
}

func ShouldDisableChannel(err *types.NewAPIError) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if types.IsChannelError(err) {
		return true
	}
	if types.IsSkipRetryError(err) {
		return false
	}
	if operation_setting.ShouldDisableByStatusCode(err.StatusCode) {
		return true
	}

	lowerMessage := strings.ToLower(err.Error())
	search, _ := AcSearch(lowerMessage, operation_setting.AutomaticDisableKeywords, true)
	return search
}

// IsChannelUnavailableError reports failures that indicate the selected
// upstream channel cannot serve this request.  Some providers report an
// exhausted balance as HTTP 400, which is intentionally excluded from the
// generic retry status-code list because most 400 responses are client errors.
// The same configured disable keywords are therefore also treated as a
// channel-level failover signal, independent of whether auto-disable is on.
func IsChannelUnavailableError(err *types.NewAPIError) bool {
	if err == nil || types.IsSkipRetryError(err) {
		return false
	}
	if types.IsChannelError(err) || operation_setting.ShouldDisableByStatusCode(err.StatusCode) {
		return true
	}
	return IsChannelUnavailableMessage(err.Error(), string(err.GetErrorCode()))
}

// IsChannelUnavailableMessage applies provider-independent channel availability
// signals to both normal relay errors and task relay errors.
func IsChannelUnavailableMessage(message string, errorCode string) bool {
	// Provider error codes are more stable than localized messages.
	switch strings.ToLower(strings.TrimSpace(errorCode)) {
	case "insufficient_quota", "quota_exceeded", "resource_exhausted", "credit_balance_too_low", "billing_hard_limit_reached":
		return true
	}

	lowerMessage := strings.ToLower(message)
	for _, fragment := range []string{
		"credit balance is too low",
		"insufficient quota",
		"quota exceeded",
		"out of credits",
		"not enough credits",
		"billing hard limit",
	} {
		if strings.Contains(lowerMessage, fragment) {
			return true
		}
	}
	search, _ := AcSearch(lowerMessage, operation_setting.AutomaticDisableKeywords, true)
	return search
}

func ShouldEnableChannel(newAPIError *types.NewAPIError, status int) bool {
	if !common.AutomaticEnableChannelEnabled {
		return false
	}
	if newAPIError != nil {
		return false
	}
	if status != common.ChannelStatusAutoDisabled {
		return false
	}
	return true
}
