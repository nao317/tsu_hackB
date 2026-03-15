package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserIDKey = "userID"
	ContextClaimsKey = "jwtClaims"
)

type AuthConfig struct {
	SecretKey   []byte
	Issuer      string
	TTLMinutes  int
	TokenHeader string // default: Authorization
	TokenParam  string // default: token
	AllowParam  bool
	AuthOptions bool // OPTIONS を認証スキップするかどうか
}

type AuthMiddleware struct {
	config AuthConfig
}

type CustomClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthMiddleware(cfg AuthConfig) *AuthMiddleware {
	if cfg.TokenHeader == "" {
		cfg.TokenHeader = "Authorization"
	}
	if cfg.TokenParam == "" {
		cfg.TokenParam = "token"
	}
	if cfg.TTLMinutes <= 0 {
		cfg.TTLMinutes = 60
	}

	return &AuthMiddleware{
		config: cfg,
	}
}

func (a *AuthMiddleware) GenerateToken(userID uint) (string, error) {
	now := time.Now()

	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(a.config.TTLMinutes) * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.config.SecretKey)
}

func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions && !a.config.AuthOptions {
			c.Next()
			return
		}

		tokenString, err := a.extractToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid token",
			})
			return
		}

		claims := &CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// alg
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return a.config.SecretKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextClaimsKey, claims)
		c.Next()
	}
}

func (a *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader(a.config.TokenHeader)
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1], nil
		}
	}

	if a.config.AllowParam {
		if token := c.Query(a.config.TokenParam); token != "" {
			return token, nil
		}
	}

	return "", errors.New("token not found")
}
