package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	"one-api/setting/operation_setting"
	"one-api/types"
	"strings"
	"time"
)

func formatNotifyType(channelId int, status int) string {
	return fmt.Sprintf("%s_%d_%d", dto.NotifyTypeChannelUpdate, channelId, status)
}

// disable & notify
func DisableChannel(channelError types.ChannelError, reason string) {
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

func ShouldDisableChannel(channelType int, err *types.NewAPIError) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if types.IsChannelError(err) {
		return true
	}
	if types.IsLocalError(err) {
		return false
	}
	if err.StatusCode == http.StatusUnauthorized {
		return true
	}
	if err.StatusCode == http.StatusForbidden {
		switch channelType {
		case constant.ChannelTypeGemini:
			return true
		}
	}
	oaiErr := err.ToOpenAIError()
	switch oaiErr.Code {
	case "invalid_api_key":
		return true
	case "account_deactivated":
		return true
	case "billing_not_active":
		return true
	case "pre_consume_token_quota_failed":
		return true
	}
	switch oaiErr.Type {
	case "insufficient_quota":
		return true
	case "insufficient_user_quota":
		return true
	// https://docs.anthropic.com/claude/reference/errors
	case "authentication_error":
		return true
	case "permission_error":
		return true
	case "forbidden":
		return true
	}

	lowerMessage := strings.ToLower(err.Error())
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

// MaskKey 对密钥进行脱敏处理
func MaskKey(key string) string {
	if len(key) <= 7 {
		return "***"
	}
	// 保留前3位和后4位，中间用***代替
	return key[:3] + "***" + key[len(key)-4:]
}

// DetectKeyType 检测密钥类型
func DetectKeyType(key string) string {
	if key == "" {
		return "single"
	}
	
	trimmed := strings.TrimSpace(key)
	
	// 检查是否为JSON格式
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(trimmed), &jsonData); err == nil {
			return "multi_json"
		}
	}
	
	// 检查是否包含换行符（多行模式）
	if strings.Contains(key, "\n") {
		return "multi_line"
	}
	
	return "single"
}

// ParseKeys 解析密钥字符串为密钥数组
func ParseKeys(key string) []string {
	if key == "" {
		return []string{}
	}
	
	trimmed := strings.TrimSpace(key)
	
	// 如果是JSON数组格式
	if strings.HasPrefix(trimmed, "[") {
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			res := make([]string, len(arr))
			for i, v := range arr {
				res[i] = string(v)
			}
			return res
		}
	}
	
	// 按换行符分割
	keys := strings.Split(strings.Trim(key, "\n"), "\n")
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			result = append(result, k)
		}
	}
	
	return result
}

// GetChannelKeyStatus 获取密钥状态
func GetChannelKeyStatus(channel *model.Channel, keyIndex int) string {
	if !channel.ChannelInfo.IsMultiKey {
		// 单密钥模式，根据渠道状态判断
		if channel.Status == common.ChannelStatusEnabled {
			return "active"
		}
		return "disabled"
	}
	
	// 多密钥模式，检查具体密钥状态
	if channel.ChannelInfo.MultiKeyStatusList != nil {
		if status, ok := channel.ChannelInfo.MultiKeyStatusList[keyIndex]; ok {
			if status == common.ChannelStatusEnabled {
				return "active"
			}
			return "disabled"
		}
	}
	
	// 默认为活跃状态
	return "active"
}

// GetChannelKeyView 获取渠道密钥查看信息
func GetChannelKeyView(channel *model.Channel, viewMode string) (*dto.ChannelKeyViewResponse, error) {
	if channel == nil {
		return nil, fmt.Errorf("渠道不存在")
	}
	
	keys := ParseKeys(channel.Key)
	keyType := DetectKeyType(channel.Key)
	
	// 如果是仅显示数量模式
	if viewMode == "count" {
		return &dto.ChannelKeyViewResponse{
			ChannelId:    channel.Id,
			KeyType:      keyType,
			KeyCount:     len(keys),
			Keys:         []dto.ChannelKeyInfo{},
			MultiKeyMode: channel.ChannelInfo.MultiKeyMode,
			IsMultiKey:   channel.ChannelInfo.IsMultiKey,
		}, nil
	}
	
	// 构建密钥信息列表
	keyInfos := make([]dto.ChannelKeyInfo, len(keys))
	for i, key := range keys {
		keyInfo := dto.ChannelKeyInfo{
			Index:     i,
			MaskedKey: MaskKey(strings.TrimSpace(key)),
			Status:    GetChannelKeyStatus(channel, i),
			LastUsed:  time.Unix(channel.TestTime, 0).Format("2006-01-02 15:04:05"),
		}
		
		// 如果有错误信息，添加到响应中
		if channel.Status != common.ChannelStatusEnabled {
			errorMsg := "渠道已禁用"
			keyInfo.ErrorMessage = &errorMsg
		}
		
		keyInfos[i] = keyInfo
	}
	
	return &dto.ChannelKeyViewResponse{
		ChannelId:    channel.Id,
		KeyType:      keyType,
		KeyCount:     len(keys),
		Keys:         keyInfos,
		MultiKeyMode: channel.ChannelInfo.MultiKeyMode,
		IsMultiKey:   channel.ChannelInfo.IsMultiKey,
	}, nil
}
