package dto

import "one-api/constant"

// ChannelKeyViewRequest 渠道密钥查看请求
type ChannelKeyViewRequest struct {
	ViewMode string `json:"view_mode" form:"view_mode"` // masked: 脱敏显示, count: 仅显示密钥数量
}

// ChannelKeyInfo 单个密钥信息
type ChannelKeyInfo struct {
	Index       int    `json:"index"`        // 密钥索引
	MaskedKey   string `json:"masked_key"`   // 脱敏后的密钥
	Status      string `json:"status"`       // 密钥状态: active, disabled, unknown
	LastUsed    string `json:"last_used"`    // 最后使用时间
	ErrorMessage *string `json:"error_message,omitempty"` // 错误信息
}

// ChannelKeyViewResponse 渠道密钥查看响应
type ChannelKeyViewResponse struct {
	ChannelId    int                      `json:"channel_id"`     // 渠道ID
	KeyType      string                   `json:"key_type"`       // 密钥类型: single, multi_line, multi_json
	KeyCount     int                      `json:"key_count"`      // 密钥数量
	Keys         []ChannelKeyInfo         `json:"keys"`           // 密钥列表
	MultiKeyMode constant.MultiKeyMode    `json:"multi_key_mode"` // 多密钥模式: random, polling
	IsMultiKey   bool                     `json:"is_multi_key"`   // 是否多密钥模式
}

// ChannelKeyError 密钥查看错误
type ChannelKeyError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}