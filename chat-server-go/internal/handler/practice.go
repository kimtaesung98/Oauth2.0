// internal/handler/practice.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"chat-server-go/internal/db"
	"chat-server-go/internal/model"
)

// ─────────────────────────────────────────
// C — Create
// POST /practice/users
// Node.js: prisma.user.create({ data: ... })
// ─────────────────────────────────────────

func CreateUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"    binding:"required"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	hashedStr := string(hashed)

	user := model.User{
		Email:    input.Email,
		Password: &hashedStr,
		Name:     input.Name,
		Provider: "local",
	}

	// Create — INSERT INTO users (...) VALUES (...)
	result := db.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// result.RowsAffected — 실제로 몇 행이 바뀌었는지
	c.JSON(http.StatusCreated, gin.H{
		"user":         user,
		"rowsAffected": result.RowsAffected, // 1
	})
}

// ─────────────────────────────────────────
// R — Read 전체
// GET /practice/users
// Node.js: prisma.user.findMany()
// ─────────────────────────────────────────

func GetAllUsers(c *gin.Context) {
	var users []model.User

	// Find — SELECT * FROM users
	result := db.DB.Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": result.RowsAffected,
	})
}

// ─────────────────────────────────────────
// R — Read 단건
// GET /practice/users/:id
// Node.js: prisma.user.findUnique({ where: { id } })
// ─────────────────────────────────────────

func GetUserByID(c *gin.Context) {
	id := c.Param("id")

	var user model.User

	// First — SELECT * FROM users WHERE id = ? LIMIT 1
	// 못 찾으면 gorm.ErrRecordNotFound 반환
	result := db.DB.First(&user, id)

	if result.Error != nil {
		// errors.Is로 에러 종류를 구분
		// Node.js: err instanceof PrismaClientKnownRequestError 와 유사
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "유저를 찾을 수 없습니다"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// ─────────────────────────────────────────
// U — Update
// PATCH /practice/users/:id
// Node.js: prisma.user.update({ where: { id }, data: { name } })
// ─────────────────────────────────────────

func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	// 먼저 존재 확인
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "유저를 찾을 수 없습니다"})
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Updates — UPDATE users SET name = ?, updated_at = ? WHERE id = ?
	// map을 쓰면 zero value(빈 문자열, 0 등)도 업데이트 가능
	result := db.DB.Model(&user).Updates(map[string]any{
		"name": input.Name,
	})

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"rowsAffected": result.RowsAffected,
	})
}

// ─────────────────────────────────────────
// D — Delete
// DELETE /practice/users/:id
// Node.js: prisma.user.delete({ where: { id } })
// ─────────────────────────────────────────

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// 먼저 존재 확인
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "유저를 찾을 수 없습니다"})
		return
	}

	// Delete — DELETE FROM users WHERE id = ?
	result := db.DB.Delete(&user)

	c.JSON(http.StatusOK, gin.H{
		"message":      "삭제 완료",
		"rowsAffected": result.RowsAffected,
	})
}

// ─────────────────────────────────────────
// R — 관계 조회 (Preload)
// GET /practice/users/:id/messages
// Node.js: prisma.user.findUnique({ include: { sentMessages: true } })
// ─────────────────────────────────────────

// UserWithMessages — 응답 전용 struct
type UserWithMessages struct {
	model.User
	SentMessages     []model.Message `json:"sentMessages"`
	ReceivedMessages []model.Message `json:"receivedMessages"`
}

func GetUserMessages(c *gin.Context) {
	id := c.Param("id")

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "유저를 찾을 수 없습니다"})
		return
	}

	// 보낸 메시지
	var sentMessages []model.Message
	db.DB.Where("sender_id = ?", id).
		Order("created_at desc").
		Find(&sentMessages)

	// 받은 메시지
	var receivedMessages []model.Message
	db.DB.Where("receiver_id = ?", id).
		Order("created_at desc").
		Find(&receivedMessages)

	c.JSON(http.StatusOK, UserWithMessages{
		User:             user,
		SentMessages:     sentMessages,
		ReceivedMessages: receivedMessages,
	})
}
