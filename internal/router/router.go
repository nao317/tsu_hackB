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

func Setup(r *gin.Engine, h *Handlers, allowedOrigins string, authMW *middleware.AuthMiddleware) {
	r.Use(middleware.CORS(allowedOrigins))

	v1 := r.Group("/api/v1")

	// 認証
	auth := v1.Group("/auth")
	{
		auth.POST("/signup", h.Auth.Signup)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)

		authPrivate := auth.Group("")
		authPrivate.Use(authMW.RequireAuth())
		authPrivate.GET("/me", h.Auth.Me)
		authPrivate.POST("/logout", h.Auth.Logout)
	}

	// ゲスト可
	v1.GET("/locations/nearby",    h.Location.Nearby)
	v1.GET("/locations",           h.Location.List)
	v1.GET("/locations/:id/cards", h.Location.GetCards)
	v1.GET("/cards/daily",         h.Card.Daily)
	v1.POST("/ai/recommend",       h.AI.Recommend)

	protected := v1.Group("")
	protected.Use(authMW.RequireAuth())
	protected.GET("/user/locations", h.UserLocation.List)
	protected.POST("/user/locations", h.UserLocation.Create)
	protected.PUT("/user/locations/:id", h.UserLocation.Update)
	protected.DELETE("/user/locations/:id", h.UserLocation.Delete)
	protected.GET("/user/locations/:id/cards", h.UserLocation.GetCards)
	protected.POST("/user/cards", h.Card.Create)
	protected.POST("/user/locations/:id/cards", h.Card.AddToLocation)
	protected.DELETE("/user/locations/:id/cards/:card_id", h.Card.RemoveFromLocation)
	protected.PUT("/user/locations/:id/cards/reorder", h.Card.Reorder)
}
