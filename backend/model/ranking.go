package model

import "time"

// RankingIncrementRequest 增加积分请求
type RankingIncrementRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	Delta   int    `json:"delta" binding:"required"`
	Reason  string `json:"reason,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// RankingIncrementResponse 增加积分响应
type RankingIncrementResponse struct {
	UserID    string    `json:"user_id"`
	NewScore  int       `json:"new_score"`
	Rank      int       `json:"rank"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RankingTopRequest 获取Top N请求
type RankingTopRequest struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// RankingTopResponse 获取Top N响应
type RankingTopResponse struct {
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Entries  []RankingEntry `json:"entries"`
}

// RankingUserResponse 获取用户排名响应
type RankingUserResponse struct {
	UserID string `json:"user_id"`
	Score  int    `json:"score"`
	Rank   int    `json:"rank"`
}

// RankingEntry 排行榜条目
type RankingEntry struct {
	Rank   int    `json:"rank"`
	UserID string `json:"user_id"`
	Score  int    `json:"score"`
}
