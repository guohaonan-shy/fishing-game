package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"fishing-game/config"
	"fishing-game/model"

	"github.com/redis/go-redis/v9"
)

type LotteryService struct {
	redisClient    *redis.Client
	rankingService *RankingService
	poolService    *PoolService
}

// NewLotteryService 创建抽奖服务
func NewLotteryService(rankingService *RankingService, poolService *PoolService) (*LotteryService, error) {
	ls := &LotteryService{
		redisClient:    config.GetRedisClient(),
		rankingService: rankingService,
		poolService:    poolService,
	}

	return ls, nil
}

// Draw 执行抽奖
func (ls *LotteryService) Draw(ctx context.Context, req *model.LotteryDrawRequest) (*model.LotteryDrawResponse, error) {
	// 从Redis获取奖池信息
	pool, err := ls.poolService.GetPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool: %w", err)
	}

	// 根据权重随机选择奖品
	_, selectedItem, err := ls.weightedRandomSelect(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to select item: %w", err)
	}

	// 生成抽奖ID
	drawID := fmt.Sprintf("d_%s_%d", time.Now().Format("20060102"), time.Now().UnixNano())

	// 创建抽奖记录
	record := &model.LotteryRecord{
		ItemID:      selectedItem.ID,
		ItemName:    selectedItem.Name,
		Description: selectedItem.Description,
		Points:      selectedItem.Points,
		Strategy:    "default", // 简化为单一策略
		Timestamp:   time.Now(),
	}

	// 保存抽奖记录到Redis
	if err := ls.saveLotteryRecord(ctx, req.UserID, record); err != nil {
		return nil, fmt.Errorf("failed to save lottery record: %w", err)
	}

	// 如果中奖且有积分，调用榜单服务增加积分
	if selectedItem.Points > 0 {
		incrementReq := &model.RankingIncrementRequest{
			UserID:  req.UserID,
			Delta:   selectedItem.Points,
			Reason:  fmt.Sprintf("lottery_win_%s", selectedItem.Name),
			TraceID: req.TraceID,
		}

		_, err := ls.rankingService.IncrementScore(ctx, incrementReq)
		if err != nil {
			return nil, fmt.Errorf("failed to increment score: %w", err)
		}
	}

	// 构造响应
	response := &model.LotteryDrawResponse{
		DrawID: drawID,
		UserID: req.UserID,
		Result: model.LotteryResult{
			Win:         selectedItem.Points > 0,
			ItemID:      selectedItem.ID,
			ItemName:    selectedItem.Name,
			Description: selectedItem.Description,
			Points:      selectedItem.Points,
			WxID:        selectedItem.WxID,     // 透出微信ID
			ImageURL:    selectedItem.ImageURL, // 透出图片URL
		},
		CreatedAt: time.Now(),
	}

	return response, nil
}

// weightedRandomSelect 基于权重的随机选择
func (ls *LotteryService) weightedRandomSelect(pool *model.PoolInfoResponse) (string, *model.LotteryItem, error) {
	if pool.TotalWeight != 1000000 {
		return "", nil, fmt.Errorf("invalid total weight: %d, expected 1000000", pool.TotalWeight)
	}

	// 生成随机数 [0, 1000000)
	randomBig, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate random number: %w", err)
	}
	randomNum := int(randomBig.Int64())

	// 累计权重选择
	cumulativeWeight := 0
	for fishID, weight := range pool.Weights {
		cumulativeWeight += weight
		if randomNum < cumulativeWeight {
			item, exists := pool.Items[fishID]
			if !exists {
				return "", nil, fmt.Errorf("item %s not found in pool", fishID)
			}
			return fishID, item, nil
		}
	}

	// 理论上不应该到达这里，但为了安全起见返回空军
	emptyFishID := "00000000-0000-0000-0000-000000000001"
	if item, exists := pool.Items[emptyFishID]; exists {
		return emptyFishID, item, nil
	}

	return "", nil, fmt.Errorf("no valid item found in pool")
}

// saveLotteryRecord 保存抽奖记录到Redis
func (ls *LotteryService) saveLotteryRecord(ctx context.Context, userID string, record *model.LotteryRecord) error {
	key := fmt.Sprintf("lottery:draws:%s", userID)

	// 将记录序列化为JSON
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// 使用LPUSH添加到列表头部（最新的在前面）
	if err := ls.redisClient.LPush(ctx, key, recordJSON).Err(); err != nil {
		return fmt.Errorf("failed to save to redis: %w", err)
	}

	// 可选：限制历史记录数量，只保留最近100条
	if err := ls.redisClient.LTrim(ctx, key, 0, 99).Err(); err != nil {
		return fmt.Errorf("failed to trim history: %w", err)
	}

	return nil
}

// GetUserDrawHistory 获取用户抽奖历史
func (ls *LotteryService) GetUserDrawHistory(ctx context.Context, userID string, limit int) ([]*model.LotteryRecord, error) {
	key := fmt.Sprintf("lottery:draws:%s", userID)

	// 获取历史记录
	results, err := ls.redisClient.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	records := make([]*model.LotteryRecord, 0, len(results))
	for _, result := range results {
		var record model.LotteryRecord
		if err := json.Unmarshal([]byte(result), &record); err != nil {
			continue // 跳过无法解析的记录
		}
		records = append(records, &record)
	}

	return records, nil
}
