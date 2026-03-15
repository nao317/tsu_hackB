package middleware

import (
    "strings"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func CORS(allowedOrigins string) gin.HandlerFunc {
    origins := strings.Split(allowedOrigins, ",")
    for i := range origins {
        origins[i] = strings.TrimSpace(origins[i])
    }

    return cors.New(cors.Config{
        AllowOrigins:     origins,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}