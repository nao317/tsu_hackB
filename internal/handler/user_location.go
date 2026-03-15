package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nao317/tsu_hack/backend/internal/model"
	"github.com/nao317/tsu_hack/backend/internal/service"
)

type UserLocationHandler struct {
	svc *service.LocationService
}

func NewUserLocationHandler(svc *service.LocationService) *UserLocationHandler {
	return &UserLocationHandler{svc: svc}
}

func (h *UserLocationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	locs, err := h.svc.ListUserLocations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusOK, locs)
}

func (h *UserLocationHandler) Create(c *gin.Context) {
	var req model.CreateUserLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	userID := c.GetString("user_id")
	loc, err := h.svc.CreateUserLocation(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusCreated, loc)
}

func (h *UserLocationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateUserLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	userID := c.GetString("user_id")
	loc, err := h.svc.UpdateUserLocation(c.Request.Context(), id, userID, &req)
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
	c.JSON(http.StatusOK, loc)
}

func (h *UserLocationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	err := h.svc.DeleteUserLocation(c.Request.Context(), id, userID)
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

func (h *UserLocationHandler) GetCards(c *gin.Context) {
	locationID := c.Param("id")
	userID := c.GetString("user_id")

	cards, err := h.svc.GetUserLocationCards(c.Request.Context(), locationID, userID)
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
	c.JSON(http.StatusOK, cards)
}
