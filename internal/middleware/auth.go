package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// JWTAuth は Authorization: Bearer <token> を検証し、
// 成功時に c.Set("user_id", userID) をセットするミドルウェア。
func JWTAuth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        if !strings.HasPrefix(header, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "認証が必要です",
                "code":  "UNAUTHORIZED",
            })
            return
        }

        tokenStr := strings.TrimPrefix(header, "Bearer ")

        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(jwtSecret), nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "無効なトークンです",
                "code":  "INVALID_TOKEN",
            })
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "トークンの解析に失敗しました",
                "code":  "INVALID_TOKEN",
            })
            return
        }

        userID, ok := claims["sub"].(string)
        if !ok || userID == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "ユーザーIDが取得できません",
                "code":  "INVALID_TOKEN",
            })
            return
        }

        c.Set("user_id", userID)
        c.Next()
    }
}