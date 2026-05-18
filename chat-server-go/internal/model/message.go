// internal/model/message.go
package model

import "time"

type Message struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Content    string    `gorm:"not null"                 json:"content"`
	IsRead     bool      `gorm:"default:false"            json:"isRead"`
	CreatedAt  time.Time `                                json:"createdAt"`
	SenderID   uint      `gorm:"not null"                 json:"senderId"`
	ReceiverID uint      `gorm:"not null"                 json:"receiverId"`

	// 연관 관계 — Prisma의 @relation에 해당
	// GORM이 JOIN할 때 자동으로 채워줌, JSON 응답엔 포함 안 할 거라 omitempty
	Sender   User `gorm:"foreignKey:SenderID"   json:"sender,omitempty"`
	Receiver User `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
}
