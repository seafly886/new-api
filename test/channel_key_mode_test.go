package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"one-api/constant"
	"one-api/controller"
	"one-api/model"
)

// TestToggleChannelKeyMode 测试渠道Key模式切换功能
func TestToggleChannelKeyMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 模拟测试用的Channel
	testChannel := &model.Channel{
		Id:   1,
		Name: "Test Channel",
		Type: 1,
		Key:  "test-key-1\ntest-key-2\ntest-key-3",
		ChannelInfo: model.ChannelInfo{
			IsMultiKey:   true,
			MultiKeySize: 3,
			MultiKeyMode: constant.MultiKeyModeRandom,
		},
	}

	tests := []struct {
		name           string
		channelId      string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedMode   constant.MultiKeyMode
		expectError    bool
	}{
		{
			name:      "Enable polling mode",
			channelId: "1", 
			requestBody: map[string]interface{}{
				"enabled": true,
			},
			expectedStatus: http.StatusOK,
			expectedMode:   constant.MultiKeyModePolling,
			expectError:    false,
		},
		{
			name:      "Disable polling mode",
			channelId: "1",
			requestBody: map[string]interface{}{
				"enabled": false,
			},
			expectedStatus: http.StatusOK,
			expectedMode:   constant.MultiKeyModeRandom,
			expectError:    false,
		},
		{
			name:      "Invalid channel ID",
			channelId: "invalid",
			requestBody: map[string]interface{}{
				"enabled": true,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			r := gin.New()
			r.PATCH("/api/channel/:id/key-mode", controller.ToggleChannelKeyMode)

			// 准备请求体
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PATCH", "/api/channel/"+tt.channelId+"/key-mode", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			r.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code)

			// 解析响应
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectError {
				assert.False(t, response["success"].(bool))
			} else {
				assert.True(t, response["success"].(bool))
				
				// 验证返回的数据
				data := response["data"].(map[string]interface{})
				assert.Equal(t, string(tt.expectedMode), data["key_mode"])
				assert.Equal(t, tt.requestBody["enabled"], data["enabled"])
			}
		})
	}
}

// TestChannelGetNextEnabledKey 测试Key轮询选择逻辑
func TestChannelGetNextEnabledKey(t *testing.T) {
	tests := []struct {
		name        string
		channel     *model.Channel
		expectKey   string
		expectError bool
	}{
		{
			name: "Single key mode",
			channel: &model.Channel{
				Key: "single-key",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey: false,
				},
			},
			expectKey:   "single-key",
			expectError: false,
		},
		{
			name: "Multi key random mode",
			channel: &model.Channel{
				Key: "key1\nkey2\nkey3",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey:   true,
					MultiKeySize: 3,
					MultiKeyMode: constant.MultiKeyModeRandom,
				},
			},
			expectError: false,
		},
		{
			name: "Multi key polling mode",
			channel: &model.Channel{
				Key: "key1\nkey2\nkey3",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey:           true,
					MultiKeySize:         3,
					MultiKeyMode:         constant.MultiKeyModePolling,
					MultiKeyPollingIndex: 0,
				},
			},
			expectKey:   "key1",
			expectError: false,
		},
		{
			name: "Empty keys",
			channel: &model.Channel{
				Key: "",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey: true,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, index, err := tt.channel.GetNextEnabledKey()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, key)
				assert.GreaterOrEqual(t, index, 0)

				if tt.expectKey != "" {
					assert.Equal(t, tt.expectKey, key)
				}

				// 对于轮询模式，验证索引递增
				if tt.channel.ChannelInfo.MultiKeyMode == constant.MultiKeyModePolling && tt.channel.ChannelInfo.IsMultiKey {
					expectedNextIndex := (index + 1) % len(tt.channel.GetKeys())
					assert.Equal(t, expectedNextIndex, tt.channel.ChannelInfo.MultiKeyPollingIndex)
				}
			}
		})
	}
}

// TestMultiKeyModeConstants 测试MultiKeyMode常量
func TestMultiKeyModeConstants(t *testing.T) {
	assert.Equal(t, "random", string(constant.MultiKeyModeRandom))
	assert.Equal(t, "polling", string(constant.MultiKeyModePolling))
}