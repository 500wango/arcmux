package service

import (
	"github.com/500wango/arcmux/setting/operation_setting"
	"github.com/500wango/arcmux/setting/system_setting"
)

func GetCallbackAddress() string {
	if operation_setting.CustomCallbackAddress == "" {
		return system_setting.ServerAddress
	}
	return operation_setting.CustomCallbackAddress
}
