package handler

import (
	"net/http"
	"strconv"

	"fishing-game/model"
	"fishing-game/service"

	"github.com/gin-gonic/gin"
)

type LotteryHandler struct {
	lotteryService *service.LotteryService
}

// NewLotteryHandler 创建抽奖处理器
func NewLotteryHandler(lotteryService *service.LotteryService) *LotteryHandler {
	return &LotteryHandler{
		lotteryService: lotteryService,
	}
}

// Draw 执行抽奖
// POST /fishing/lotteries/draw
func (lh *LotteryHandler) Draw(c *gin.Context) {
	var req model.LotteryDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	response, err := lh.lotteryService.Draw(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// GetUserDrawHistory 获取用户抽奖历史（可选功能）
// GET /fishing/lotteries/history/{user_id}?limit=10
func (lh *LotteryHandler) GetUserDrawHistory(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	// 获取limit参数，默认为10
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // 最多返回100条
	}

	records, err := lh.lotteryService.GetUserDrawHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(records))
}
