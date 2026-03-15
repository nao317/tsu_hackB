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

func Setup(r *gin.Engine, h *Handlers, allowedOrigins string) {
	r.Use(middleware.CORS(allowedOrigins))

	v1 := r.Group("/api/v1")

	// 認証
	auth := v1.Group("/auth")
	{
		auth.POST("/signup", h.Auth.Signup)
		auth.POST("/login",  h.Auth.Login)
		auth.GET("/me",      h.Auth.Me)
	}

	// ゲスト可
	v1.GET("/locations/nearby",    h.Location.Nearby)
	v1.GET("/locations",           h.Location.List)
	v1.GET("/locations/:id/cards", h.Location.GetCards)
	v1.GET("/cards/daily",         h.Card.Daily)
	v1.POST("/ai/recommend",       h.AI.Recommend)

	// TODO: JWT実装後に認証ミドルウェアを追加する
	v1.GET("/user/locations",                           h.UserLocation.List)
	v1.POST("/user/locations",                          h.UserLocation.Create)
	v1.PUT("/user/locations/:id",                       h.UserLocation.Update)
	v1.DELETE("/user/locations/:id",                    h.UserLocation.Delete)
	v1.GET("/user/locations/:id/cards",                 h.UserLocation.GetCards)
	v1.POST("/user/cards",                              h.Card.Create)
	v1.POST("/user/locations/:id/cards",                h.Card.AddToLocation)
	v1.DELETE("/user/locations/:id/cards/:card_id",     h.Card.RemoveFromLocation)
	v1.PUT("/user/locations/:id/cards/reorder",         h.Card.Reorder)
}
