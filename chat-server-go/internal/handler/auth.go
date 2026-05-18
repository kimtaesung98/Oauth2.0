// internal/handler/auth.go
package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"

	"chat-server-go/internal/db"
	"chat-server-go/internal/model"
)

// ─────────────────────────────────────────
// 1.1.1 회원가입
// ─────────────────────────────────────────

type RegisterInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"     binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing model.User
	if result := db.DB.Where("email = ?", input.Email).First(&existing); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "이미 사용 중인 이메일입니다"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "서버 오류"})
		return
	}

	hashedStr := string(hashed)
	user := model.User{
		Email:    input.Email,
		Password: &hashedStr,
		Name:     input.Name,
		Provider: "local",
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "회원가입 실패"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "회원가입 성공",
		"user":    user,
	})
}

func validateToken(tokenStr string) (uint, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	claims := token.Claims.(jwt.MapClaims)
	return uint(claims["userId"].(float64)), nil
}

// ─────────────────────────────────────────
// 1.1.2 로그인
// ─────────────────────────────────────────

type LoginInput struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 틀렸습니다"})
		return
	}

	if user.Password == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "소셜 로그인 계정입니다"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 틀렸습니다"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 발급 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

// ─────────────────────────────────────────
// 1.1.3 Google 소셜 로그인
// ─────────────────────────────────────────

type GoogleLoginInput struct {
	IDToken string `json:"idToken" binding:"required"`
}

func GoogleLogin(c *gin.Context) {
	var input GoogleLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	// ← 추가: 실제 에러 내용 확인용
	log.Println("GOOGLE_CLIENT_ID:", clientID)
	log.Println("IDToken 앞 20자:", input.IDToken[:20])

	payload, err := idtoken.Validate(context.Background(), input.IDToken, clientID)
	if err != nil {
		// ← 수정: 실제 에러 내용 출력
		log.Println("Google 토큰 검증 실패:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	googleID := payload.Subject

	var user model.User
	result := db.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		user = model.User{
			Email:    email,
			Name:     name,
			Provider: "google",
			GoogleID: googleID,
			Password: nil,
		}
		if err := db.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "유저 생성 실패"})
			return
		}
	} else {
		db.DB.Model(&user).Updates(map[string]any{
			"google_id": googleID,
			"provider":  "google",
		})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 발급 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}
