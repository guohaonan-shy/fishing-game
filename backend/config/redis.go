package config

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis() {
	// 从环境变量获取Redis地址，默认为localhost:6379
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // 本地无密码
		DB:       0,  // 默认数据库
	})

	// 测试连接
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}
}

// GetRedisClient 获取Redis客户端
func GetRedisClient() *redis.Client {
	return RedisClient
}
