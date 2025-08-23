# 渠道Key轮询模式功能测试文档

## 测试概述

本文档描述了渠道Key轮询模式功能的测试计划和测试用例。

## 功能描述

在渠道列表(/console/channel)中增加"Key模式"列，提供开关控制：
- **开启状态**：该渠道在调用API时使用轮询模式（每个Key使用一次，循环使用）
- **关闭状态**：使用随机模式选择Key

## 测试环境要求

- Go 1.23.4+
- Node.js + npm/bun
- MySQL/PostgreSQL/SQLite数据库
- Redis（可选，用于缓存）

## 测试用例

### 1. 后端API测试

#### 1.1 渠道Key模式切换接口测试

**接口**: `PATCH /api/channel/{id}/key-mode`

**测试用例1**: 启用轮询模式
```json
请求:
{
  "enabled": true
}

期望响应:
{
  "success": true,
  "message": "Key模式切换成功",
  "data": {
    "channel_id": 1,
    "key_mode": "polling",
    "enabled": true
  }
}
```

**测试用例2**: 禁用轮询模式
```json
请求:
{
  "enabled": false
}

期望响应:
{
  "success": true,
  "message": "Key模式切换成功", 
  "data": {
    "channel_id": 1,
    "key_mode": "random",
    "enabled": false
  }
}
```

**测试用例3**: 单Key渠道切换
```json
期望响应:
{
  "success": false,
  "message": "该渠道不是多Key模式，无法切换Key模式"
}
```

**测试用例4**: 无效渠道ID
```json
期望响应:
{
  "success": false,
  "message": "参数错误"
}
```

#### 1.2 渠道更新接口测试

**接口**: `PUT /api/channel/`

**测试用例**: 通过multi_key_mode字段更新
```json
请求:
{
  "id": 1,
  "multi_key_mode": "polling"
}

期望: 渠道的MultiKeyMode被正确更新
```

### 2. Key选择逻辑测试

#### 2.1 单Key模式测试
- **输入**: 单Key渠道
- **期望**: 始终返回该Key

#### 2.2 多Key随机模式测试
- **输入**: 多Key渠道，mode="random"
- **期望**: 随机选择一个可用Key

#### 2.3 多Key轮询模式测试
- **输入**: 多Key渠道，mode="polling"
- **期望**: 按顺序轮询选择Key，到末尾后回到开头

#### 2.4 Key状态过滤测试
- **输入**: 部分Key被禁用的多Key渠道
- **期望**: 只从可用Key中选择

#### 2.5 轮询索引管理测试
- **输入**: 连续多次调用GetNextEnabledKey
- **期望**: 轮询索引正确递增并循环

### 3. 前端组件测试

#### 3.1 显示测试
- **单Key渠道**: 显示"单Key"标签，无开关
- **多Key随机模式**: 显示开关（关闭状态）+ "随机"标签
- **多Key轮询模式**: 显示开关（开启状态）+ "轮询"标签

#### 3.2 交互测试
- **开关点击**: 正确调用API并更新状态
- **加载状态**: API调用期间显示loading状态
- **错误处理**: API错误时显示错误消息
- **成功反馈**: 操作成功时显示成功消息

#### 3.3 状态同步测试
- **本地状态**: 操作后正确更新本地渠道数据
- **界面刷新**: 状态变化后界面正确重新渲染

### 4. 集成测试

#### 4.1 API调用流程测试
1. 客户端请求 `/v1/chat/completions`
2. middleware.Distribute() 选择渠道
3. SetupContextForSelectedChannel() 调用 GetNextEnabledKey()
4. 根据MultiKeyMode选择Key
5. Key设置到context中供relay使用

#### 4.2 轮询一致性测试
- **并发请求**: 多个并发请求的Key选择是否正确轮询
- **缓存同步**: Redis缓存环境下的状态同步
- **故障恢复**: 服务重启后轮询状态恢复

#### 4.3 性能测试
- **响应时间**: Key选择是否增加明显延迟
- **并发处理**: 高并发情况下的性能表现
- **内存使用**: 轮询锁和状态管理的内存占用

### 5. 数据库测试

#### 5.1 ChannelInfo存储测试
- **JSON字段**: MultiKeyMode正确存储和读取
- **向下兼容**: 现有渠道的默认值处理
- **数据迁移**: 升级时的数据兼容性

#### 5.2 并发更新测试
- **锁机制**: 轮询索引更新的线程安全性
- **数据一致性**: 多实例环境下的状态同步

### 6. 用户界面测试

#### 6.1 渠道列表页面
- **列显示**: Key模式列正确显示
- **列设置**: 可以隐藏/显示Key模式列
- **标签聚合**: 标签聚合模式下不显示Key模式切换

#### 6.2 响应式设计
- **移动端**: 小屏幕设备上的显示效果
- **桌面端**: 大屏幕设备上的布局

### 7. 错误处理测试

#### 7.1 边界条件
- **空Key列表**: 无可用Key时的处理
- **无效Key格式**: 错误的Key格式处理
- **网络错误**: API调用失败的处理

#### 7.2 权限测试
- **管理员权限**: 只有管理员可以切换Key模式
- **普通用户**: 普通用户无法看到/操作Key模式开关

## 测试执行步骤

### 准备工作
1. 启动数据库（MySQL/PostgreSQL/SQLite）
2. 启动Redis（如果使用缓存）
3. 创建测试渠道数据

### 后端测试
```bash
# 运行Go单元测试
go test ./test -v

# 运行特定测试
go test ./test -run TestToggleChannelKeyMode -v
go test ./test -run TestChannelGetNextEnabledKey -v
```

### 前端测试
```bash
# 安装依赖
cd web && npm install --legacy-peer-deps

# 运行前端测试
npm test ChannelKeyModeToggle.test.js
```

### 手动测试
1. 启动完整应用
2. 登录管理员账号
3. 访问 `/console/channel`
4. 验证Key模式列显示
5. 测试开关操作

## 预期结果

### 成功标准
- ✅ 所有单元测试通过
- ✅ API接口正确响应
- ✅ 前端组件正确显示和交互
- ✅ Key轮询逻辑工作正常
- ✅ 并发场景下状态一致
- ✅ 错误场景正确处理

### 性能要求
- Key选择延迟 < 10ms
- 并发处理能力 > 1000 QPS
- 内存增长 < 10MB

## 已知限制

1. **前端构建**: 当前存在依赖冲突问题，需要使用 `--legacy-peer-deps`
2. **Go环境**: 需要确保Go 1.23.4已正确安装
3. **权限控制**: 功能仅对管理员用户开放

## 回归测试

在每次代码更改后，应执行以下回归测试：
1. 基本Key选择功能
2. 现有随机模式兼容性
3. 缓存机制正常工作
4. API接口向下兼容

## 故障排除

### 常见问题
1. **API调用失败**: 检查路由配置和权限
2. **Key选择异常**: 检查渠道配置和Key格式
3. **前端显示错误**: 检查channel_info数据结构
4. **并发问题**: 检查锁机制和缓存同步