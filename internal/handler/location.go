package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/nao317/tsu_hack/backend/internal/service"
    "github.com/nao317/tsu_hack/backend/internal/model"
)

type LocationHandler struct {
    svc *service.LocationService
}

func NewLocationHandler(svc *service.LocationService) *LocationHandler {
    return &LocationHandler{svc: svc}
}

func (h *LocationHandler) Nearby(c *gin.Context) {
    var q model.NearbyQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }
    if q.RadiusM == 0 {
        q.RadiusM = 500 // デフォルト500m
    }

    // 認証済みの場合はユーザーロケーションも含める（任意）
    userID, _ := c.Get("user_id")
    uid, _ := userID.(string)

    locs, err := h.svc.GetNearby(c.Request.Context(), q.Lat, q.Lng, q.RadiusM, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, locs)
}

func (h *LocationHandler) List(c *gin.Context) {
    locs, err := h.svc.ListShared(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, locs)
}

func (h *LocationHandler) GetCards(c *gin.Context) {
    locationID := c.Param("id")
    cards, err := h.svc.GetCards(c.Request.Context(), locationID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, cards)
}