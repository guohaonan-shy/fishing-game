package handler

import (
	"net/http"

	"fishing-game/model"
	"fishing-game/service"

	"github.com/gin-gonic/gin"
)

type RankingHandler struct {
	rankingService *service.RankingService
}

// NewRankingHandler 创建榜单处理器
func NewRankingHandler(rankingService *service.RankingService) *RankingHandler {
	return &RankingHandler{
		rankingService: rankingService,
	}
}

// IncrementScore 增加积分
// POST /fishing/leaderboards/scores/increment
func (rh *RankingHandler) IncrementScore(c *gin.Context) {
	var req model.RankingIncrementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	response, err := rh.rankingService.IncrementScore(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// GetTopRanking 获取Top N排行榜
// GET /fishing/leaderboards/top?page=1&page_size=50
func (rh *RankingHandler) GetTopRanking(c *gin.Context) {
	var req model.RankingTopRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 50

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	response, err := rh.rankingService.GetTopRanking(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// GetUserRanking 获取单个用户的排名和积分
// GET /fishing/leaderboards/users/{user_id}
func (rh *RankingHandler) GetUserRanking(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	response, err := rh.rankingService.GetUserRanking(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}
