package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nao317/tsu_hack/backend/internal/model"
	"github.com/nao317/tsu_hack/backend/internal/service"
)

type CardHandler struct {
	svc *service.CardService
}

func NewCardHandler(svc *service.CardService) *CardHandler {
	return &CardHandler{svc: svc}
}

func (h *CardHandler) Daily(c *gin.Context) {
	cards, err := h.svc.GetDailyCards(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusOK, cards)
}

func (h *CardHandler) Create(c *gin.Context) {
	var req model.CreateCardRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil && err.Error() != "http: no such file" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ファイルの読み込みエラー", "code": "INVALID_FILE"})
		return
	}
	if file != nil {
		defer file.Close()
	}

	userID := c.GetString("user_id")
	card, err := h.svc.CreateCard(c.Request.Context(), userID, &req, file, header)
	if errors.Is(err, service.ErrInvalidImageType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_IMAGE_TYPE"})
		return
	}
	if errors.Is(err, service.ErrImageTooLarge) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "IMAGE_TOO_LARGE"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusCreated, card)
}

func (h *CardHandler) AddToLocation(c *gin.Context) {
	locationID := c.Param("id")
	var req model.AddCardToLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	userID := c.GetString("user_id")
	err := h.svc.AddToLocation(c.Request.Context(), locationID, userID, &req)
	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "code": "NOT_FOUND"})
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *CardHandler) RemoveFromLocation(c *gin.Context) {
	locationID := c.Param("id")
	cardID := c.Param("card_id")
	userID := c.GetString("user_id")

	err := h.svc.RemoveFromLocation(c.Request.Context(), locationID, cardID, userID)
	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "code": "NOT_FOUND"})
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CardHandler) Reorder(c *gin.Context) {
	locationID := c.Param("id")
	var req model.ReorderCardsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	userID := c.GetString("user_id")
	err := h.svc.ReorderCards(c.Request.Context(), locationID, userID, &req)
	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "code": "NOT_FOUND"})
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
		return
	}
	c.Status(http.StatusNoContent)
}
