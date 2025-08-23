package constant

type MultiKeyMode string

const (
	MultiKeyModeRandom     MultiKeyMode = "random"     // 随机
	MultiKeyModePolling    MultiKeyMode = "polling"    // 轮询
	MultiKeyModeSequential MultiKeyMode = "sequential" // 顺序循环
)

// KeyStrategy 定义Key使用策略类型
type KeyStrategy string

const (
	KeyStrategyRandom     KeyStrategy = "random"     // 随机Key模式
	KeyStrategySequential KeyStrategy = "sequential" // 顺序循环模式
)
