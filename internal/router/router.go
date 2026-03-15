package router

import (
    "github.com/gin-gonic/gin"
    "github.com/nao317/tsu_hack/backend/internal/handler"
    "github.com/nao317/tsu_hack/backend/internal/middleware"
)

type Handlers struct {
    Auth         *handler.AuthHandler
    Location     *handler.LocationHandler
    UserLocation *handler.UserLocationHandler
    Card         *handler.CardHandler
    AI           *handler.AIHandler
}

func Setup(r *gin.Engine, h *Handlers, jwtSecret string, allowedOrigins string) {
    r.Use(middleware.CORS(allowedOrigins))

    v1 := r.Group("/api/v1")

    // 認証不要
    auth := v1.Group("/auth")
    {
        auth.POST("/signup",  h.Auth.Signup)
        auth.POST("/login",   h.Auth.Login)
        auth.POST("/refresh", h.Auth.Refresh)
    }

    // ゲスト可（認証オプション）
    v1.GET("/locations/nearby",        h.Location.Nearby)
    v1.GET("/locations",               h.Location.List)
    v1.GET("/locations/:id/cards",     h.Location.GetCards)
    v1.GET("/cards/daily",             h.Card.Daily)
    v1.POST("/ai/recommend",           h.AI.Recommend)

    // 認証必須
    authed := v1.Group("")
    authed.Use(middleware.JWTAuth(jwtSecret))
    {
        authed.POST("/auth/logout", h.Auth.Logout)
        authed.GET("/auth/me",      h.Auth.Me)

        // ユーザーロケーション
        authed.GET("/user/locations",               h.UserLocation.List)
        authed.POST("/user/locations",              h.UserLocation.Create)
        authed.PUT("/user/locations/:id",           h.UserLocation.Update)
        authed.DELETE("/user/locations/:id",        h.UserLocation.Delete)
        authed.GET("/user/locations/:id/cards",     h.UserLocation.GetCards)

        // カード
        authed.POST("/user/cards",                              h.Card.Create)
        authed.POST("/user/locations/:id/cards",                h.Card.AddToLocation)
        authed.DELETE("/user/locations/:id/cards/:card_id",     h.Card.RemoveFromLocation)
        authed.PUT("/user/locations/:id/cards/reorder",         h.Card.Reorder)
    }
}