// internal/handler/oauth.go
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"chat-server-go/internal/db"
	"chat-server-go/internal/model"
)

// oauthConfig — Google OAuth 설정
// .env에서 읽어야 하므로 함수로 감쌈
// (패키지 초기화 시점에는 godotenv가 아직 로드 안 됐을 수 있음)
func getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:3000/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// ─────────────────────────────────────────
// ① GET /auth/google/login
// 브라우저를 Google 로그인 페이지로 리디렉트
// Node.js: res.redirect(oauth2Client.generateAuthUrl(...)) 와 동일
// ─────────────────────────────────────────
func GoogleOAuthLogin(c *gin.Context) {
	config := getOAuthConfig()

	// state — CSRF 방지용 값
	// 실제 서비스에서는 crypto/rand로 생성 후 세션에 저장해야 함
	state := "random-state-string"

	// Google 로그인 URL 생성 후 리디렉트
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// ─────────────────────────────────────────
// ④ GET /auth/google/callback
// Google이 Authorization Code를 들고 여기로 돌아옴
// ─────────────────────────────────────────
func GoogleOAuthCallback(c *gin.Context) {
	config := getOAuthConfig()

	// Google이 보내준 code 꺼내기
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code 없음"})
		return
	}

	// ⑤ code → Access Token 교환
	// 브라우저는 이 과정을 모름 — Go 서버 ↔ Google 서버끼리만 통신
	oauthToken, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 교환 실패: " + err.Error()})
		return
	}

	// Access Token으로 Google 유저 정보 가져오기
	client := config.Client(context.Background(), oauthToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "유저 정보 조회 실패"})
		return
	}
	defer resp.Body.Close()

	// 유저 정보 파싱
	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "유저 정보 파싱 실패"})
		return
	}

	// 유저 조회 또는 자동 생성
	var user model.User
	result := db.DB.Where("email = ?", googleUser.Email).First(&user)
	if result.Error != nil {
		user = model.User{
			Email:    googleUser.Email,
			Name:     googleUser.Name,
			Provider: "google",
			GoogleID: googleUser.ID,
			Password: nil,
		}
		if err := db.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "유저 생성 실패"})
			return
		}
	} else {
		db.DB.Model(&user).Updates(map[string]any{
			"google_id": googleUser.ID,
			"provider":  "google",
		})
	}

	// JWT 발급
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 발급 실패"})
		return
	}

	// 프론트로 리디렉트하면서 토큰 전달
	c.Redirect(http.StatusTemporaryRedirect,
		"http://localhost:5173/chat.html?token="+tokenString)
}
