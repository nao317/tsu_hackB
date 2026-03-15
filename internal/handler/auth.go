package handler

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/nao317/tsu_hack/backend/internal/model"
    "github.com/nao317/tsu_hack/backend/internal/service"
)

type AuthHandler struct {
    svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Signup(c *gin.Context) {
    var req model.SignupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    resp, err := h.svc.Signup(c.Request.Context(), &req)
    if errors.Is(err, service.ErrEmailAlreadyExists) {
        c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "EMAIL_ALREADY_EXISTS"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    resp, err := h.svc.Login(c.Request.Context(), &req)
    if errors.Is(err, service.ErrInvalidCredentials) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "INVALID_CREDENTIALS"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
    var req model.RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    resp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
    if errors.Is(err, service.ErrInvalidToken) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "INVALID_TOKEN"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
    var req model.LogoutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }
    h.svc.Logout(c.Request.Context(), req.RefreshToken)
    c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
    userID := c.GetString("user_id")
    me, err := h.svc.GetMe(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "code": "NOT_FOUND"})
        return
    }
    c.JSON(http.StatusOK, me)
}