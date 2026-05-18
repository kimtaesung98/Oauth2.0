// internal/handler/message.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"chat-server-go/internal/db"
	"chat-server-go/internal/model"
)

// ─────────────────────────────────────────
// 메시지 저장
// POST /messages
// Node.js: prisma.message.create({ data: ... })
// ─────────────────────────────────────────

type SendMessageInput struct {
	ReceiverID uint   `json:"receiverId" binding:"required"`
	Content    string `json:"content"    binding:"required"`
}

func SendMessage(c *gin.Context) {
	// JWT 미들웨어가 저장해둔 userId 꺼내기
	// Node.js: req.user.userId
	senderID := c.MustGet("userId").(uint)

	var input SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := model.Message{
		Content:    input.Content,
		SenderID:   senderID,
		ReceiverID: input.ReceiverID,
	}

	if err := db.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "메시지 저장 실패"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// ─────────────────────────────────────────
// 메시지 조회
// GET /messages/:userId
// 나(senderID) ↔ 상대방(userId) 사이의 대화 전체
// ─────────────────────────────────────────

func GetMessages(c *gin.Context) {
	myID := c.MustGet("userId").(uint)

	// URL 파라미터 꺼내기 — Node.js의 req.params.userId
	otherID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 userId"})
		return
	}

	var messages []model.Message

	// 나→상대 OR 상대→나 메시지 모두 조회
	// Node.js Prisma:
	// prisma.message.findMany({
	//   where: { OR: [
	//     { senderId: myID, receiverId: otherID },
	//     { senderId: otherID, receiverId: myID }
	//   ]},
	//   orderBy: { createdAt: 'asc' }
	// })
	db.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		myID, otherID, otherID, myID,
	).Order("created_at asc").Find(&messages)

	c.JSON(http.StatusOK, messages)
}

// ─────────────────────────────────────────
// 읽음 처리
// PATCH /messages/read
// 상대방이 나에게 보낸 메시지 중 안 읽은 것 → isRead = true
// ─────────────────────────────────────────

type MarkReadInput struct {
	SenderID uint `json:"senderId" binding:"required"`
}

func MarkAsRead(c *gin.Context) {
	myID := c.MustGet("userId").(uint)

	var input MarkReadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Node.js Prisma:
	// prisma.message.updateMany({
	//   where: { senderId: input.senderId, receiverId: myID, isRead: false },
	//   data: { isRead: true }
	// })
	db.DB.Model(&model.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = false", input.SenderID, myID).
		Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{"message": "읽음 처리 완료"})
}

// ─────────────────────────────────────────
// 채팅 목록 조회
// GET /chats
// 내가 참여한 대화 목록 + 마지막 메시지 + 안 읽은 수
// ─────────────────────────────────────────

// ChatRoom — 응답 전용 struct
// DB 테이블이 아니라 조회 결과를 담는 그릇
type ChatRoom struct {
	PartnerID   uint   `json:"partnerId"`
	PartnerName string `json:"partnerName"`
	LastMessage string `json:"lastMessage"`
	UnreadCount int64  `json:"unreadCount"`
}

func GetChats(c *gin.Context) {
	myID := c.MustGet("userId").(uint)

	// 나와 대화한 상대방 ID 목록
	var partnerIDs []uint
	db.DB.Model(&model.Message{}).
		Where("sender_id = ? OR receiver_id = ?", myID, myID).
		Select("DISTINCT CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS partner_id", myID).
		Pluck("partner_id", &partnerIDs)

	var chatRooms []ChatRoom

	for _, partnerID := range partnerIDs {
		var partner model.User
		db.DB.First(&partner, partnerID)

		// 마지막 메시지
		var lastMsg model.Message
		db.DB.Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			myID, partnerID, partnerID, myID,
		).Order("created_at desc").First(&lastMsg)

		// 안 읽은 수
		var unreadCount int64
		db.DB.Model(&model.Message{}).
			Where("sender_id = ? AND receiver_id = ? AND is_read = false", partnerID, myID).
			Count(&unreadCount)

		chatRooms = append(chatRooms, ChatRoom{
			PartnerID:   partnerID,
			PartnerName: partner.Name,
			LastMessage: lastMsg.Content,
			UnreadCount: unreadCount,
		})
	}

	// chatRooms가 nil이면 빈 배열로 응답
	if chatRooms == nil {
		chatRooms = []ChatRoom{}
	}

	c.JSON(http.StatusOK, chatRooms)
}
