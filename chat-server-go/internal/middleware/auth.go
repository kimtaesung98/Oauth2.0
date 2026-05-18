// internal/middleware/auth.go
package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthRequired — Node.js의 JWT 검증 미들웨어와 동일
// Fastify: app.addHook('preHandler', verifyJWT)
// Gin:     r.Use(middleware.AuthRequired())
func AuthRequired() gin.HandlerFunc {
	// gin.HandlerFunc는 func(*gin.Context) 타입의 별칭
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// "Bearer <token>" 형태 확인
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 토큰이 없습니다"})
			c.Abort() // 다음 핸들러로 넘기지 않고 여기서 중단
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// JWT 검증
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 토큰"})
			c.Abort()
			return
		}

		// claims에서 userId 꺼내서 context에 저장
		// Node.js: req.user = decoded  와 동일
		claims := token.Claims.(jwt.MapClaims)
		c.Set("userId", uint(claims["userId"].(float64)))

		c.Next() // 다음 핸들러로 진행
	}
}
