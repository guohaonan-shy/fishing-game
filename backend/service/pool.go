package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"

	"fishing-game/config"
	"fishing-game/model"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis keys
	PoolItemsKey   = "lottery:pool:items"
	PoolWeightsKey = "lottery:pool:weights"

	// 权重配置
	TotalWeight      = 1000000
	BaseWeight       = 2500
	WeightMultiplier = 100
	MaxDescLength    = 25

	// 下限保护
	EmptyMinWeight  = 100000 // 空军最低10%
	SmallMinWeight  = 50000  // 小鱼最低5%
	MediumMinWeight = 25000  // 中鱼最低2.5%
	LargeMinWeight  = 10000  // 大鱼最低1%

	// 图片资源配置
	BaseAssetURL    = "/assets" // 基础资源URL
	SystemImagePath = ""        // 系统鱼图片路径（直接在assets根目录）
	UserImagePath   = ""        // 用户鱼图片路径（直接在assets根目录）
	ImageExtension  = ".png"    // 图片扩展名
)

// 系统鱼类ID（固定UUID）
var (
	EmptyFishID  = "00000000-0000-0000-0000-000000000001"
	SmallFishID  = "00000000-0000-0000-0000-000000000002"
	MediumFishID = "00000000-0000-0000-0000-000000000003"
	LargeFishID  = "00000000-0000-0000-0000-000000000004"
	RareFishID   = "00000000-0000-0000-0000-000000000005"
)

// 用户鱼图片资源池（fish_1.jpg到fish_15.jpg）
var UserFishImages = []string{
	"fish_1", "fish_2", "fish_3", "fish_4", "fish_5",
	"fish_7", "fish_8", "fish_9", "fish_10",
}

// BorrowInfo 权重借用信息
type BorrowInfo struct {
	FishID    string
	MinWeight int
}

// 权重借用顺序（全局配置，便于后续修改）
var BorrowOrder = []BorrowInfo{
	{EmptyFishID, EmptyMinWeight},   // 优先从空军借用
	{SmallFishID, SmallMinWeight},   // 其次从小鱼借用
	{MediumFishID, MediumMinWeight}, // 再从中鱼借用
	{LargeFishID, LargeMinWeight},   // 最后从大鱼借用
	// 注意：稀有鱼不参与借用，保持权重不变
}

type PoolService struct {
	redisClient *redis.Client
}

// NewPoolService 创建奖池服务
func NewPoolService() *PoolService {
	return &PoolService{
		redisClient: config.GetRedisClient(),
	}
}

// InitializePool 初始化奖池（迁移现有数据）
func (ps *PoolService) InitializePool(ctx context.Context) error {
	// 检查是否已经初始化
	exists, err := ps.redisClient.Exists(ctx, PoolItemsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check pool existence: %w", err)
	}

	if exists > 0 {
		return nil // 已经初始化过了
	}

	// 初始化系统鱼类
	systemFishes := []*model.LotteryItem{
		{
			ID:          EmptyFishID,
			Name:        "空军",
			Description: "今天运气不太好，鱼儿们都在睡觉，只钓到了一堆水草和无尽的等待时光",
			Points:      0,
			IsUserFish:  false,
			ImageURL:    "",
		},
		{
			ID:          SmallFishID,
			Name:        "小鱼",
			Description: "一条活泼可爱的小鱼，虽然个头不大但充满活力，游来游去像个调皮的孩子",
			Points:      5,
			IsUserFish:  false,
			ImageURL:    fmt.Sprintf("%s/small%s", BaseAssetURL, ImageExtension),
		},
		{
			ID:          MediumFishID,
			Name:        "中鱼",
			Description: "体型适中的鱼儿，肉质鲜美，正好够一顿美餐，是钓鱼人最喜欢的收获",
			Points:      20,
			IsUserFish:  false,
			ImageURL:    fmt.Sprintf("%s/medium%s", BaseAssetURL, ImageExtension),
		},
		{
			ID:          LargeFishID,
			Name:        "大鱼",
			Description: "一条威武的大鱼，力大无穷，上钩时差点把鱼竿都拉断了，绝对是今日最佳战利品",
			Points:      100,
			IsUserFish:  false,
			ImageURL:    fmt.Sprintf("%s/large%s", BaseAssetURL, ImageExtension),
		},
		{
			ID:          RareFishID,
			Name:        "稀有鱼",
			Description: "传说中的神秘鱼类，全身闪闪发光，据说一生只能遇到一次，是所有钓鱼人梦寐以求的终极目标",
			Points:      500,
			IsUserFish:  false,
			ImageURL:    fmt.Sprintf("%s/rare%s", BaseAssetURL, ImageExtension),
		},
	}

	// 初始权重分布
	initialWeights := map[string]int{
		EmptyFishID:  500000, // 50%
		SmallFishID:  300000, // 30%
		MediumFishID: 150000, // 15%
		LargeFishID:  40000,  // 4%
		RareFishID:   10000,  // 1%
	}

	// 保存到Redis
	pipe := ps.redisClient.Pipeline()

	// 保存鱼类信息
	for _, fish := range systemFishes {
		fishJSON, err := json.Marshal(fish)
		if err != nil {
			return fmt.Errorf("failed to marshal fish %s: %w", fish.ID, err)
		}
		pipe.HSet(ctx, PoolItemsKey, fish.ID, fishJSON)
	}

	// 保存权重信息
	for fishID, weight := range initialWeights {
		pipe.HSet(ctx, PoolWeightsKey, fishID, weight)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize pool: %w", err)
	}

	return nil
}

// getRandomUserImage 随机选择用户鱼图片
func getRandomUserImage() (string, error) {
	if len(UserFishImages) == 0 {
		return "", fmt.Errorf("no user fish images available")
	}

	// 生成随机索引
	randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(UserFishImages))))
	if err != nil {
		return "", fmt.Errorf("failed to generate random index: %w", err)
	}

	// 构建完整的图片URL
	imageName := UserFishImages[randomIndex.Int64()]
	imageURL := fmt.Sprintf("%s/%s%s", BaseAssetURL, imageName, ImageExtension)

	return imageURL, nil
}

// getUserImageURL 获取用户鱼图片URL（支持指定或随机）
func getUserImageURL(imageName string) (string, error) {
	// 如果指定了图片名称，验证并使用指定的图片
	if imageName != "" {
		// 验证指定的图片名称是否在可用列表中
		found := false
		for _, availableName := range UserFishImages {
			if availableName == imageName {
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("invalid image name: %s, available images: %v", imageName, UserFishImages)
		}

		imageURL := fmt.Sprintf("%s/%s%s", BaseAssetURL, imageName, ImageExtension)
		return imageURL, nil
	}

	// 如果没有指定图片名称，使用随机策略
	return getRandomUserImage()
}

// AddFish 添加新鱼到奖池
func (ps *PoolService) AddFish(ctx context.Context, req *model.AddFishRequest) (*model.AddFishResponse, error) {
	// 生成UUID
	fishID := uuid.New().String()

	// 计算权重
	weight := ps.calculateWeight(req.Description)

	// 获取用户图片URL（指定或随机）
	imageURL, err := getUserImageURL(req.ImageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user image: %w", err)
	}

	// 创建新鱼
	newFish := &model.LotteryItem{
		ID:          fishID,
		Name:        req.Name,
		Description: req.Description,
		Points:      250, // 用户鱼固定250分
		IsUserFish:  true,
		WxID:        req.WxID, // 保存微信ID
		ImageURL:    imageURL, // 随机分配的图片URL
	}

	// 借用权重
	if err := ps.borrowWeight(ctx, weight); err != nil {
		return nil, fmt.Errorf("failed to borrow weight: %w", err)
	}

	// 保存新鱼到Redis
	pipe := ps.redisClient.Pipeline()

	fishJSON, err := json.Marshal(newFish)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new fish: %w", err)
	}

	pipe.HSet(ctx, PoolItemsKey, fishID, fishJSON)
	pipe.HSet(ctx, PoolWeightsKey, fishID, weight)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to save new fish: %w", err)
	}

	return &model.AddFishResponse{
		ID:          fishID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    newFish.ImageURL,
	}, nil
}

// calculateWeight 计算鱼的权重
func (ps *PoolService) calculateWeight(description string) int {
	descLength := len([]rune(strings.TrimSpace(description)))
	effectiveLength := int(math.Min(float64(descLength), float64(MaxDescLength)))
	return BaseWeight + effectiveLength*WeightMultiplier
}

// borrowWeight 从系统鱼类借用权重
func (ps *PoolService) borrowWeight(ctx context.Context, needWeight int) error {
	// 获取当前权重
	weights, err := ps.redisClient.HGetAll(ctx, PoolWeightsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get current weights: %w", err)
	}

	// 转换为int map
	currentWeights := make(map[string]int)
	for fishID, weightStr := range weights {
		var weight int
		if err := json.Unmarshal([]byte(weightStr), &weight); err != nil {
			// 如果JSON解析失败，尝试直接转换字符串
			if w, parseErr := parseWeight(weightStr); parseErr == nil {
				weight = w
			} else {
				return fmt.Errorf("failed to parse weight for %s: %w", fishID, err)
			}
		}
		currentWeights[fishID] = weight
	}

	// 按全局配置的优先级借用权重

	remainingNeed := needWeight
	newWeights := make(map[string]int)

	// 复制当前权重
	for fishID, weight := range currentWeights {
		newWeights[fishID] = weight
	}

	// 按顺序借用
	for _, borrowInfo := range BorrowOrder {
		if remainingNeed <= 0 {
			break
		}

		currentWeight := newWeights[borrowInfo.FishID]
		availableWeight := currentWeight - borrowInfo.MinWeight

		if availableWeight > 0 {
			borrowAmount := int(math.Min(float64(remainingNeed), float64(availableWeight)))
			newWeights[borrowInfo.FishID] -= borrowAmount
			remainingNeed -= borrowAmount
		}
	}

	// 检查是否成功借到足够权重
	if remainingNeed > 0 {
		return fmt.Errorf("insufficient weight available, still need %d", remainingNeed)
	}

	// 更新Redis中的权重
	pipe := ps.redisClient.Pipeline()
	for fishID, weight := range newWeights {
		if currentWeights[fishID] != weight {
			pipe.HSet(ctx, PoolWeightsKey, fishID, weight)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update weights: %w", err)
	}

	return nil
}

// GetPool 获取完整奖池信息
func (ps *PoolService) GetPool(ctx context.Context) (*model.PoolInfoResponse, error) {
	// 获取所有鱼类
	itemsData, err := ps.redisClient.HGetAll(ctx, PoolItemsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	// 获取所有权重
	weightsData, err := ps.redisClient.HGetAll(ctx, PoolWeightsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get weights: %w", err)
	}

	// 解析鱼类信息
	items := make(map[string]*model.LotteryItem)
	for fishID, itemJSON := range itemsData {
		var item model.LotteryItem
		if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item %s: %w", fishID, err)
		}
		items[fishID] = &item
	}

	// 解析权重信息
	weights := make(map[string]int)
	totalWeight := 0
	for fishID, weightStr := range weightsData {
		var weight int
		if err := json.Unmarshal([]byte(weightStr), &weight); err != nil {
			// 如果JSON解析失败，尝试直接转换
			if w, parseErr := parseWeight(weightStr); parseErr == nil {
				weight = w
			} else {
				return nil, fmt.Errorf("failed to parse weight for %s: %w", fishID, err)
			}
		}
		weights[fishID] = weight
		totalWeight += weight
	}

	return &model.PoolInfoResponse{
		TotalItems:  len(items),
		Items:       items,
		Weights:     weights,
		TotalWeight: totalWeight,
	}, nil
}

// parseWeight 辅助函数：解析权重字符串
func parseWeight(weightStr string) (int, error) {
	// 尝试多种解析方式
	var weight int

	// 方式1：直接解析整数
	if _, err := fmt.Sscanf(weightStr, "%d", &weight); err == nil {
		return weight, nil
	}

	return 0, fmt.Errorf("unable to parse weight: %s", weightStr)
}
