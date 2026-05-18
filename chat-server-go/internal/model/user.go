// internal/model/user.go
package model

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null"     json:"email"`
	Password  *string   `gorm:"default:null"             json:"-"` // ← string → *string (포인터, nil 허용)
	Name      string    `gorm:"not null"                 json:"name"`
	Provider  string    `gorm:"default:'local'"          json:"provider"` // 'local' or 'google'
	GoogleID  string    `gorm:"default:null"             json:"googleId,omitempty"`
	CreatedAt time.Time `                                json:"createdAt"`
	UpdatedAt time.Time `                                json:"updatedAt"`
}
