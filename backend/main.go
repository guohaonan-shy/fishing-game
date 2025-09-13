package main

import (
	"context"
	"log"

	"fishing-game/config"
	"fishing-game/handler"
	"fishing-game/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化Redis连接
	config.InitRedis()
	log.Println("Redis connection initialized")

	// 初始化服务层
	rankingService := service.NewRankingService()
	poolService := service.NewPoolService()

	// 初始化奖池数据
	if err := poolService.InitializePool(context.Background()); err != nil {
		log.Fatalf("Failed to initialize pool: %v", err)
	}
	log.Println("Pool initialized")

	lotteryService, err := service.NewLotteryService(rankingService, poolService)
	if err != nil {
		log.Fatalf("Failed to initialize lottery service: %v", err)
	}
	log.Println("Services initialized")

	// 初始化处理器
	rankingHandler := handler.NewRankingHandler(rankingService)
	lotteryHandler := handler.NewLotteryHandler(lotteryService)
	poolHandler := handler.NewPoolHandler(poolService)

	// 创建Gin路由器
	r := gin.Default()

	// 设置路由
	setupRoutes(r, rankingHandler, lotteryHandler, poolHandler)

	// 启动服务器
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine, rankingHandler *handler.RankingHandler, lotteryHandler *handler.LotteryHandler, poolHandler *handler.PoolHandler) {
	// 创建API组
	api := r.Group("/fishing")

	// 榜单相关路由
	leaderboards := api.Group("/leaderboards")
	{
		// 增加积分
		leaderboards.POST("/scores/increment", rankingHandler.IncrementScore)
		// 获取Top N排行榜
		leaderboards.GET("/top", rankingHandler.GetTopRanking)
		// 获取单个用户的排名和积分
		leaderboards.GET("/users/:user_id", rankingHandler.GetUserRanking)
	}

	// 抽奖相关路由
	lotteries := api.Group("/lotteries")
	{
		// 执行抽奖
		lotteries.POST("/draw", lotteryHandler.Draw)
		// 获取用户抽奖历史
		lotteries.GET("/history/:user_id", lotteryHandler.GetUserDrawHistory)
	}

	// 奖池相关路由
	lottery := api.Group("/lottery")
	{
		// 添加新鱼到奖池
		lottery.POST("/items", poolHandler.AddFish)
		// 获取奖池信息
		lottery.GET("/pool", poolHandler.GetPool)
	}

	// 健康检查
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "fishing-game-backend",
		})
	})
}
