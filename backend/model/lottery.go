package model

import "time"

// LotteryPool 奖池配置
type LotteryPool struct {
	Items []LotteryItem `json:"items"`
}

// LotteryItem 奖品配置
type LotteryItem struct {
	ID          int    `json:"id"`          // 奖品ID
	Name        string `json:"name"`        // 奖品名称
	Description string `json:"description"` // 奖品描述
	Points      int    `json:"points"`      // 积分奖励
}

// LotteryStrategies 抽奖策略配置
type LotteryStrategies struct {
	Strategies map[string]LotteryStrategy `json:"strategies"`
}

// LotteryStrategy 单个抽奖策略
type LotteryStrategy struct {
	Description string         `json:"description"` // 策略描述
	Weights     map[string]int `json:"weights"`     // 奖品ID到权重的映射（基于1000000）
}

// LotteryDrawRequest 抽奖请求
type LotteryDrawRequest struct {
	UserID  string                 `json:"user_id" binding:"required"`
	Context map[string]interface{} `json:"context,omitempty"`
	TraceID string                 `json:"trace_id,omitempty"`
}

// LotteryDrawResponse 抽奖响应
type LotteryDrawResponse struct {
	DrawID    string        `json:"draw_id"`
	UserID    string        `json:"user_id"`
	Result    LotteryResult `json:"result"`
	CreatedAt time.Time     `json:"created_at"`
}

// LotteryResult 抽奖结果
type LotteryResult struct {
	Win         bool   `json:"win"`
	ItemID      int    `json:"item_id"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	Points      int    `json:"points"`
}

// LotteryRecord 抽奖记录（存储在Redis中）
type LotteryRecord struct {
	ItemID      int       `json:"item_id"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	Points      int       `json:"points"`
	Strategy    string    `json:"strategy"`
	Timestamp   time.Time `json:"timestamp"`
}
