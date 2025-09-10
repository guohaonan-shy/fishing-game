package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"fishing-game/config"
	"fishing-game/model"

	"github.com/redis/go-redis/v9"
)

type LotteryService struct {
	redisClient       *redis.Client
	rankingService    *RankingService
	lotteryPool       *model.LotteryPool
	lotteryStrategies *model.LotteryStrategies
	itemsMap          map[int]*model.LotteryItem // ID到奖品的映射
}

// NewLotteryService 创建抽奖服务
func NewLotteryService(rankingService *RankingService) (*LotteryService, error) {
	ls := &LotteryService{
		redisClient:    config.GetRedisClient(),
		rankingService: rankingService,
	}

	// 加载奖池配置
	if err := ls.loadLotteryPool(); err != nil {
		return nil, fmt.Errorf("failed to load lottery pool: %w", err)
	}

	// 加载策略配置
	if err := ls.loadLotteryStrategies(); err != nil {
		return nil, fmt.Errorf("failed to load lottery strategies: %w", err)
	}

	// 构建奖品映射
	ls.buildItemsMap()

	return ls, nil
}

// loadLotteryPool 加载奖池配置
func (ls *LotteryService) loadLotteryPool() error {
	configFile := "configs/lottery_pool.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read pool config file: %w", err)
	}

	ls.lotteryPool = &model.LotteryPool{}
	if err := json.Unmarshal(data, ls.lotteryPool); err != nil {
		return fmt.Errorf("failed to parse pool config file: %w", err)
	}

	return nil
}

// loadLotteryStrategies 加载策略配置
func (ls *LotteryService) loadLotteryStrategies() error {
	configFile := "configs/lottery_strategies.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read strategies config file: %w", err)
	}

	ls.lotteryStrategies = &model.LotteryStrategies{}
	if err := json.Unmarshal(data, ls.lotteryStrategies); err != nil {
		return fmt.Errorf("failed to parse strategies config file: %w", err)
	}

	// 验证所有策略的权重总和
	for strategyName, strategy := range ls.lotteryStrategies.Strategies {
		totalWeight := 0
		for _, weight := range strategy.Weights {
			totalWeight += weight
		}
		if totalWeight != 1000000 {
			return fmt.Errorf("strategy %s total weight must be 1000000, got %d", strategyName, totalWeight)
		}
	}

	return nil
}

// buildItemsMap 构建奖品ID到奖品对象的映射
func (ls *LotteryService) buildItemsMap() {
	ls.itemsMap = make(map[int]*model.LotteryItem)
	for i := range ls.lotteryPool.Items {
		item := &ls.lotteryPool.Items[i]
		ls.itemsMap[item.ID] = item
	}
}

// selectStrategyByDuration 根据duration选择策略
func (ls *LotteryService) selectStrategyByDuration(duration float64) string {
	if duration <= 1000 {
		return "high"
	} else if duration <= 3000 {
		return "medium"
	} else if duration <= 5000 {
		return "random"
	} else {
		return "empty"
	}
}

// getDurationFromContext 从context中提取duration
func (ls *LotteryService) getDurationFromContext(ctx map[string]interface{}) float64 {
	if ctx == nil {
		return 10000 // 默认超过5000ms，使用empty策略
	}

	durationValue, exists := ctx["duration"]
	if !exists {
		return 10000 // 没有duration字段，使用empty策略
	}

	// 尝试转换为float64
	switch v := durationValue.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}

	return 10000 // 无法解析，使用empty策略
}

// Draw 执行抽奖
func (ls *LotteryService) Draw(ctx context.Context, req *model.LotteryDrawRequest) (*model.LotteryDrawResponse, error) {
	// 从context中获取duration并选择策略
	duration := ls.getDurationFromContext(req.Context)
	strategyName := ls.selectStrategyByDuration(duration)

	// 根据策略随机选择奖品
	selectedItem, err := ls.weightedRandomSelectByStrategy(strategyName)
	if err != nil {
		return nil, fmt.Errorf("failed to select item with strategy %s: %w", strategyName, err)
	}

	// 生成抽奖ID
	drawID := fmt.Sprintf("d_%s_%d", time.Now().Format("20060102"), time.Now().UnixNano())

	// 创建抽奖记录
	record := &model.LotteryRecord{
		ItemID:      selectedItem.ID,
		ItemName:    selectedItem.Name,
		Description: selectedItem.Description,
		Points:      selectedItem.Points,
		Strategy:    strategyName,
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
		},
		CreatedAt: time.Now(),
	}

	return response, nil
}

// weightedRandomSelectByStrategy 基于策略的权重随机选择
func (ls *LotteryService) weightedRandomSelectByStrategy(strategyName string) (*model.LotteryItem, error) {
	strategy, exists := ls.lotteryStrategies.Strategies[strategyName]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", strategyName)
	}

	// 生成随机数 [0, 1000000)
	randomBig, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}
	randomNum := int(randomBig.Int64())

	// 累计权重选择
	cumulativeWeight := 0
	for itemIDStr, weight := range strategy.Weights {
		cumulativeWeight += weight
		if randomNum < cumulativeWeight {
			// 将字符串ID转换为整数
			itemID, err := strconv.Atoi(itemIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid item ID %s: %w", itemIDStr, err)
			}

			// 查找对应的奖品
			item, exists := ls.itemsMap[itemID]
			if !exists {
				return nil, fmt.Errorf("item with ID %d not found", itemID)
			}

			return item, nil
		}
	}

	// 理论上不应该到达这里，但为了安全起见返回空军
	return ls.itemsMap[1], nil // ID 1 是空军
}

// 注释：保持向后兼容的旧方法已移除，现在全部使用策略选择

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

// GetUserDrawHistory 获取用户抽奖历史（可选功能）
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
