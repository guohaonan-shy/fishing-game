package handler

import (
	"net/http"

	"fishing-game/model"
	"fishing-game/service"

	"github.com/gin-gonic/gin"
)

type PoolHandler struct {
	poolService *service.PoolService
}

// NewPoolHandler 创建奖池处理器
func NewPoolHandler(poolService *service.PoolService) *PoolHandler {
	return &PoolHandler{
		poolService: poolService,
	}
}

// AddFish 添加新鱼到奖池
// POST /fishing/lottery/items
func (ph *PoolHandler) AddFish(c *gin.Context) {
	var req model.AddFishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	response, err := ph.poolService.AddFish(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// GetPool 获取奖池信息
// GET /fishing/lottery/pool
func (ph *PoolHandler) GetPool(c *gin.Context) {
	response, err := ph.poolService.GetPool(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}
