package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserIDKey = "user_id"
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
	TokenType string `json:"type"`
	jwt.RegisteredClaims
}

func NewAuthMiddleware(cfg AuthConfig) *AuthMiddleware {
	if cfg.TokenHeader == "" {
		cfg.TokenHeader = "Authorization"
	}
	if cfg.TokenParam == "" {
		cfg.TokenParam = "token"
	}
	return &AuthMiddleware{
		config: cfg,
	}
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
				"error": "トークンが不正です",
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
		if err != nil || !token.Valid || claims.TokenType != "access" || claims.Subject == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "認証に失敗しました",
			})
			return
		}

		c.Set(ContextUserIDKey, claims.Subject)
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
