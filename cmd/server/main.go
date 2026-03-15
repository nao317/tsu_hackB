package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nao317/tsu_hack/backend/internal/config"
	"github.com/nao317/tsu_hack/backend/internal/db"
	"github.com/nao317/tsu_hack/backend/internal/handler"
	"github.com/nao317/tsu_hack/backend/internal/router"
	"github.com/nao317/tsu_hack/backend/internal/service"
	"github.com/nao317/tsu_hack/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer database.Close()

	// ストレージ
	imageStorage := storage.NewSupabaseStorage(
		cfg.SupabaseURL,
		cfg.SupabaseServiceKey,
		cfg.SupabaseStorageBucket,
	)

	// サービス
	authSvc     := service.NewAuthService(database)
	locationSvc := service.NewLocationService(database)
	cardSvc     := service.NewCardService(database, imageStorage)
	aiSvc       := service.NewAIService(cfg.GeminiAPIKey)

	// ハンドラ
	handlers := &router.Handlers{
		Auth:         handler.NewAuthHandler(authSvc),
		Location:     handler.NewLocationHandler(locationSvc),
		UserLocation: handler.NewUserLocationHandler(locationSvc),
		Card:         handler.NewCardHandler(cardSvc),
		AI:           handler.NewAIHandler(aiSvc),
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.Setup(r, handlers, cfg.CORSAllowedOrigins)

	log.Printf("サーバー起動: :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("サーバー起動失敗: %v", err)
	}
}
