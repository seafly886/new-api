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

// TestKeyStrategyConstants 测试新的Key策略常量
func TestKeyStrategyConstants(t *testing.T) {
	assert.Equal(t, "random", string(constant.KeyStrategyRandom))
	assert.Equal(t, "sequential", string(constant.KeyStrategySequential))
}

// TestChannelKeyStrategySelection 测试Key策略选择逻辑
func TestChannelKeyStrategySelection(t *testing.T) {
	tests := []struct {
		name        string
		channel     *model.Channel
		expectError bool
		description string
	}{
		{
			name: "Polling disabled - should use random",
			channel: &model.Channel{
				Id:  1,
				Key: "key1\nkey2\nkey3\nkey4",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey:      true,
					MultiKeySize:    4,
					PollingEnabled:  false,
					PollingStrategy: constant.KeyStrategyRandom,
				},
			},
			expectError: false,
			description: "When polling is disabled, should use random selection",
		},
		{
			name: "Random strategy enabled",
			channel: &model.Channel{
				Id:  2,
				Key: "key1\nkey2\nkey3\nkey4",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey:      true,
					MultiKeySize:    4,
					PollingEnabled:  true,
					PollingStrategy: constant.KeyStrategyRandom,
				},
			},
			expectError: false,
			description: "Random strategy should work correctly",
		},
		{
			name: "Sequential strategy enabled",
			channel: &model.Channel{
				Id:  3,
				Key: "key1\nkey2\nkey3\nkey4",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey:       true,
					MultiKeySize:     4,
					PollingEnabled:   true,
					PollingStrategy:  constant.KeyStrategySequential,
					SequentialIndex:  0,
				},
			},
			expectError: false,
			description: "Sequential strategy should work correctly",
		},
		{
			name: "Single key mode",
			channel: &model.Channel{
				Id:  4,
				Key: "single-key",
				ChannelInfo: model.ChannelInfo{
					IsMultiKey: false,
				},
			},
			expectError: false,
			description: "Single key mode should work regardless of strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, index, err := tt.channel.GetNextEnabledKey()

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, key, "Key should not be empty")
				assert.GreaterOrEqual(t, index, 0, "Index should be non-negative")
			}
		})
	}
}

// TestSequentialKeySelection 测试顺序循环Key选择
func TestSequentialKeySelection(t *testing.T) {
	channel := &model.Channel{
		Id:  5,
		Key: "key1\nkey2\nkey3",
		ChannelInfo: model.ChannelInfo{
			IsMultiKey:       true,
			MultiKeySize:     3,
			PollingEnabled:   true,
			PollingStrategy:  constant.KeyStrategySequential,
			SequentialIndex:  0,
		},
	}

	expectedKeys := []string{"key1", "key2", "key3", "key1", "key2"}
	
	for i, expectedKey := range expectedKeys {
		key, index, err := channel.GetNextEnabledKey()
		
		assert.NoError(t, err, "Iteration %d should not error", i)
		assert.Equal(t, expectedKey, key, "Iteration %d should return expected key", i)
		assert.Equal(t, i%3, index, "Iteration %d should return expected index", i)
		
		// 验证索引循环更新
		expectedNextIndex := (i + 1) % 3
		assert.Equal(t, expectedNextIndex, channel.ChannelInfo.SequentialIndex, 
			"Sequential index should be updated correctly after iteration %d", i)
	}
}

// TestRandomKeyDistribution 测试随机Key分布
func TestRandomKeyDistribution(t *testing.T) {
	channel := &model.Channel{
		Id:  6,
		Key: "key1\nkey2\nkey3\nkey4",
		ChannelInfo: model.ChannelInfo{
			IsMultiKey:      true,
			MultiKeySize:    4,
			PollingEnabled:  true,
			PollingStrategy: constant.KeyStrategyRandom,
		},
	}

	keyDistribution := make(map[string]int)
	totalRuns := 1000

	for i := 0; i < totalRuns; i++ {
		key, _, err := channel.GetNextEnabledKey()
		assert.NoError(t, err)
		keyDistribution[key]++
	}

	// 验证所有key都被选择过
	assert.Len(t, keyDistribution, 4, "All 4 keys should be selected")
	
	// 验证分布的随机性（每个key至少被选择了10%的次数）
	for key, count := range keyDistribution {
		minExpected := totalRuns / 10 // 至少10%
		assert.GreaterOrEqual(t, count, minExpected, 
			"Key %s should be selected at least %d times, got %d", key, minExpected, count)
	}
}

// TestUpdateChannelKeyStrategy 测试更新渠道Key策略API
func TestUpdateChannelKeyStrategy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		channelId      string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
		description    string
	}{
		{
			name:      "Enable polling with random strategy",
			channelId: "1",
			requestBody: map[string]interface{}{
				"polling_enabled":  true,
				"polling_strategy": "random",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should successfully enable polling with random strategy",
		},
		{
			name:      "Enable polling with sequential strategy",
			channelId: "1",
			requestBody: map[string]interface{}{
				"polling_enabled":  true,
				"polling_strategy": "sequential",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should successfully enable polling with sequential strategy",
		},
		{
			name:      "Disable polling",
			channelId: "1",
			requestBody: map[string]interface{}{
				"polling_enabled":  false,
				"polling_strategy": "random",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should successfully disable polling",
		},
		{
			name:      "Invalid strategy",
			channelId: "1",
			requestBody: map[string]interface{}{
				"polling_enabled":  true,
				"polling_strategy": "invalid_strategy",
			},
			expectedStatus: http.StatusOK,
			expectError:    true,
			description:    "Should reject invalid strategy",
		},
		{
			name:      "Invalid channel ID",
			channelId: "invalid",
			requestBody: map[string]interface{}{
				"polling_enabled":  true,
				"polling_strategy": "random",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should reject invalid channel ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			r := gin.New()
			r.PUT("/api/channel/:id/key-strategy", controller.UpdateChannelKeyStrategy)

			// 准备请求体
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/api/channel/"+tt.channelId+"/key-strategy", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			r.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			// 解析响应
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Should be able to parse response")

			if tt.expectError {
				assert.False(t, response["success"].(bool), tt.description)
			} else {
				assert.True(t, response["success"].(bool), tt.description)
				
				// 验证返回的数据
				if data, ok := response["data"].(map[string]interface{}); ok {
					assert.Equal(t, tt.requestBody["polling_enabled"], data["polling_enabled"], 
						"Should return correct polling_enabled value")
					assert.Equal(t, tt.requestBody["polling_strategy"], data["polling_strategy"], 
						"Should return correct polling_strategy value")
				}
			}
		})
	}
}

// TestBatchUpdateChannelKeyStrategy 测试批量更新渠道Key策略API
func TestBatchUpdateChannelKeyStrategy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
		description    string
	}{
		{
			name: "Batch enable polling",
			requestBody: map[string]interface{}{
				"channel_ids":      []int{1, 2, 3},
				"polling_enabled":  true,
				"polling_strategy": "sequential",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should successfully batch enable polling",
		},
		{
			name: "Empty channel IDs",
			requestBody: map[string]interface{}{
				"channel_ids":      []int{},
				"polling_enabled":  true,
				"polling_strategy": "random",
			},
			expectedStatus: http.StatusOK,
			expectError:    true,
			description:    "Should reject empty channel IDs",
		},
		{
			name: "Invalid strategy in batch update",
			requestBody: map[string]interface{}{
				"channel_ids":      []int{1, 2},
				"polling_enabled":  true,
				"polling_strategy": "invalid",
			},
			expectedStatus: http.StatusOK,
			expectError:    true,
			description:    "Should reject invalid strategy in batch update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			r := gin.New()
			r.PATCH("/api/channels/key-strategy", controller.BatchUpdateChannelKeyStrategy)

			// 准备请求体
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PATCH", "/api/channels/key-strategy", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			r.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			// 解析响应
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Should be able to parse response")

			if tt.expectError {
				assert.False(t, response["success"].(bool), tt.description)
			} else {
				assert.True(t, response["success"].(bool), tt.description)
			}
		})
	}
}

// TestChannelInfoExtensions 测试ChannelInfo结构体扩展
func TestChannelInfoExtensions(t *testing.T) {
	channelInfo := model.ChannelInfo{
		IsMultiKey:       true,
		MultiKeySize:     3,
		PollingEnabled:   true,
		PollingStrategy:  constant.KeyStrategySequential,
		SequentialIndex:  1,
	}

	// 验证新字段的存在和默认值
	assert.True(t, channelInfo.PollingEnabled, "PollingEnabled should be set correctly")
	assert.Equal(t, constant.KeyStrategySequential, channelInfo.PollingStrategy, "PollingStrategy should be set correctly")
	assert.Equal(t, 1, channelInfo.SequentialIndex, "SequentialIndex should be set correctly")
}