package model

import "time"

// LotteryPool 奖池配置（Redis存储）
type LotteryPool struct {
	Items   map[string]*LotteryItem `json:"items"`   // item_id -> LotteryItem
	Weights map[string]int          `json:"weights"` // item_id -> weight
}

// LotteryItem 奖品配置
type LotteryItem struct {
	ID          string `json:"id"`              // 奖品UUID
	Name        string `json:"name"`            // 奖品名称
	Description string `json:"description"`     // 奖品描述
	Points      int    `json:"points"`          // 积分奖励
	IsUserFish  bool   `json:"is_user_fish"`    // 是否为用户添加的鱼
	WxID        string `json:"wx_id,omitempty"` // 微信ID（仅用户添加的鱼有值）
	ImageURL    string `json:"image_url"`       // 图片URL
}

// AddFishRequest 添加新鱼请求
type AddFishRequest struct {
	Name        string `json:"name" binding:"required"`        // 鱼的名称
	Description string `json:"description" binding:"required"` // 鱼的描述
	WxID        string `json:"wx_id" binding:"required"`       // 微信ID
	ImageName   string `json:"image_name,omitempty"`           // 指定的图片名称（可选，如fish_1, fish_2等）
}

// AddFishResponse 添加新鱼响应
type AddFishResponse struct {
	ID          string `json:"id"`          // 生成的UUID
	Name        string `json:"name"`        // 鱼的名称
	Description string `json:"description"` // 鱼的描述
	ImageURL    string `json:"image_url"`   // 图片URL
}

// PoolInfoResponse 奖池信息响应
type PoolInfoResponse struct {
	TotalItems  int                     `json:"total_items"`  // 总鱼类数量
	Items       map[string]*LotteryItem `json:"items"`        // 所有鱼类信息
	Weights     map[string]int          `json:"weights"`      // 权重分布
	TotalWeight int                     `json:"total_weight"` // 总权重（应该是1000000）
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
	ItemID      string `json:"item_id"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	Points      int    `json:"points"`
	WxID        string `json:"wx_id,omitempty"` // 微信ID（仅用户添加的鱼有值）
	ImageURL    string `json:"image_url"`       // 图片URL
}

// LotteryRecord 抽奖记录（存储在Redis中）
type LotteryRecord struct {
	ItemID      string    `json:"item_id"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	Points      int       `json:"points"`
	Strategy    string    `json:"strategy"`
	Timestamp   time.Time `json:"timestamp"`
}
