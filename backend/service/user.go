package service

import (
	"context"
	"encoding/json"
	"fmt"

	"fishing-game/config"
	"fishing-game/model"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	redisClient *redis.Client
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		redisClient: config.GetRedisClient(),
	}
}

// CreateUser 创建用户信息
func (us *UserService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.CreateUserResponse, error) {
	user := &model.User{
		Username: req.Username,
		WxID:     req.WxID,
	}

	// 将用户信息序列化为JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	// 存储到Redis，key为 user:{wx_id}
	key := fmt.Sprintf("user:%s", req.WxID)
	err = us.redisClient.Set(ctx, key, userJSON, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to save user to redis: %w", err)
	}

	return &model.CreateUserResponse{
		Username: req.Username,
		WxID:     req.WxID,
		Success:  true,
	}, nil
}

// GetUser 获取用户信息
func (us *UserService) GetUser(ctx context.Context, wxID string) (*model.GetUserResponse, error) {
	key := fmt.Sprintf("user:%s", wxID)

	userJSON, err := us.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 用户不存在
			return &model.GetUserResponse{
				WxID:  wxID,
				Found: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	var user model.User
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &model.GetUserResponse{
		Username: user.Username,
		WxID:     user.WxID,
		Found:    true,
	}, nil
}

// BatchGetUsers 批量获取用户信息
func (us *UserService) BatchGetUsers(ctx context.Context, wxIDs []string) (*model.BatchGetUsersResponse, error) {
	if len(wxIDs) == 0 {
		return &model.BatchGetUsersResponse{
			Users: make(map[string]*model.User),
		}, nil
	}

	// 构建所有的key
	keys := make([]string, len(wxIDs))
	for i, wxID := range wxIDs {
		keys[i] = fmt.Sprintf("user:%s", wxID)
	}

	// 批量获取
	results, err := us.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to batch get users: %w", err)
	}

	users := make(map[string]*model.User)
	for i, result := range results {
		wxID := wxIDs[i]

		if result == nil {
			// 用户不存在，跳过
			continue
		}

		var user model.User
		err = json.Unmarshal([]byte(result.(string)), &user)
		if err != nil {
			// 跳过解析失败的用户
			continue
		}

		users[wxID] = &user
	}

	return &model.BatchGetUsersResponse{
		Users: users,
	}, nil
}

// UserExists 检查用户是否存在
func (us *UserService) UserExists(ctx context.Context, wxID string) (bool, error) {
	key := fmt.Sprintf("user:%s", wxID)

	exists, err := us.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists > 0, nil
}
