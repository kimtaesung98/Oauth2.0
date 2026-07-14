// main.go
package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"chat-server-go/internal/db"
	"chat-server-go/internal/handler"
	"chat-server-go/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env 파일 없음")
	}

	db.Connect()

	r := gin.Default()

	// CORS 설정 — 라우트 등록 전에 먼저 와야 함
	// Node.js: app.use(cors(...))  와 동일 — 미들웨어는 항상 라우트보다 위
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 인증 라우트 그룹
	auth := r.Group("/auth")
	{
		auth.POST("/register",       handler.Register)
		auth.POST("/login",          handler.Login)
		auth.POST("/google",         handler.GoogleLogin)        // 방식 1 기존
		auth.GET("/google/login",    handler.GoogleOAuthLogin)   // 방식 2 ← 추가
		auth.GET("/google/callback", handler.GoogleOAuthCallback) // 방식 2 ← 추가
	}

	// JWT 미들웨어 적용 라우트
	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		protected.POST("/messages", handler.SendMessage)
		protected.GET("/messages/:userId", handler.GetMessages)
		protected.PATCH("/messages/read", handler.MarkAsRead)
		protected.GET("/chats", handler.GetChats)
		protected.GET("/users/online", handler.GetOnlineUsers)
	}

	// 실습용 라우트
	practice := r.Group("/practice")
	{
		practice.POST("/users", handler.CreateUser)
		practice.GET("/users", handler.GetAllUsers)
		practice.GET("/users/:id", handler.GetUserByID)
		practice.PATCH("/users/:id", handler.UpdateUser)
		practice.DELETE("/users/:id", handler.DeleteUser)
		practice.GET("/users/:id/messages", handler.GetUserMessages)
	}

	// WebSocket
	r.GET("/ws", handler.HandleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("서버 시작: http://localhost:%s", port)
	r.Run(":" + port)
}
