package service

import (
	"context"
	"fmt"
	"time"

	"fishing-game/config"
	"fishing-game/model"

	"github.com/redis/go-redis/v9"
)

const (
	GlobalRankingKey = "leaderboard:global_ranklist"
)

type RankingService struct {
	redisClient *redis.Client
	UserService *UserService
}

// NewRankingService 创建榜单服务
func NewRankingService(UserService *UserService) *RankingService {
	return &RankingService{
		redisClient: config.GetRedisClient(),
		UserService: UserService,
	}
}

// IncrementScore 增加用户积分
func (rs *RankingService) IncrementScore(ctx context.Context, req *model.RankingIncrementRequest) (*model.RankingIncrementResponse, error) {
	userKey := fmt.Sprintf("user:%s", req.UserID)

	// 使用 ZINCRBY 增加积分
	newScore, err := rs.redisClient.ZIncrBy(ctx, GlobalRankingKey, float64(req.Delta), userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to increment score: %w", err)
	}

	// 获取用户排名（从0开始，需要+1）
	rank, err := rs.redisClient.ZRevRank(ctx, GlobalRankingKey, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get rank: %w", err)
	}

	return &model.RankingIncrementResponse{
		UserID:    req.UserID,
		NewScore:  int(newScore),
		Rank:      int(rank) + 1, // Redis rank从0开始，转换为从1开始
		UpdatedAt: time.Now(),
	}, nil
}

// GetTopRanking 获取Top N排行榜
func (rs *RankingService) GetTopRanking(ctx context.Context, req *model.RankingTopRequest) (*model.RankingTopResponse, error) {
	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize
	stop := offset + req.PageSize - 1

	// 使用 ZREVRANGE 获取排行榜数据（从高到低）
	results, err := rs.redisClient.ZRevRangeWithScores(ctx, GlobalRankingKey, int64(offset), int64(stop)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get top ranking: %w", err)
	}

	// 提取所有用户ID（wxID）
	wxIDs := make([]string, 0, len(results))
	for _, result := range results {
		// 提取用户ID（去掉"user:"前缀）
		userID := result.Member.(string)[5:]
		wxIDs = append(wxIDs, userID)
	}

	// 批量获取用户信息
	users, err := rs.UserService.BatchGetUsers(ctx, wxIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	entries := make([]model.RankingEntry, 0, len(results))
	for i, result := range results {
		userID := result.Member.(string)[5:] // 去掉"user:"前缀
		user := users.Users[userID]

		username := userID // 默认使用userID作为用户名
		if user != nil {
			username = user.Username
		}

		entry := model.RankingEntry{
			Rank:     offset + i + 1, // 计算实际排名
			UserID:   userID,
			Username: username,
			Score:    int(result.Score),
		}
		entries = append(entries, entry)
	}

	return &model.RankingTopResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
		Entries:  entries,
	}, nil
}

// GetUserRanking 获取单个用户的排名和积分
func (rs *RankingService) GetUserRanking(ctx context.Context, userID string) (*model.RankingUserResponse, error) {
	userKey := fmt.Sprintf("user:%s", userID)

	// 获取用户信息
	userResp, err := rs.UserService.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	username := userID // 默认使用userID作为用户名
	if userResp.Found {
		username = userResp.Username
	}

	// 获取用户积分
	score, err := rs.redisClient.ZScore(ctx, GlobalRankingKey, userKey).Result()
	if err != nil {
		if err == redis.Nil {
			// 用户不在排行榜中，返回默认值
			return &model.RankingUserResponse{
				UserID:   userID,
				Username: username,
				Score:    0,
				Rank:     0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get user score: %w", err)
	}

	// 获取用户排名
	rank, err := rs.redisClient.ZRevRank(ctx, GlobalRankingKey, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user rank: %w", err)
	}

	return &model.RankingUserResponse{
		UserID:   userID,
		Username: username,
		Score:    int(score),
		Rank:     int(rank) + 1, // Redis rank从0开始，转换为从1开始
	}, nil
}
