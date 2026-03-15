package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nao317/tsu_hack/backend/internal/model"
	"github.com/nao317/tsu_hack/backend/internal/service"
)

type AIHandler struct {
	svc *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

func (h *AIHandler) Recommend(c *gin.Context) {
	var req model.AIRecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	resp, err := h.svc.Recommend(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AIエラー", "code": "AI_ERROR"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
