// internal/db/db.go
package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"chat-server-go/internal/model"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB 연결 실패:", err)
	}

	// AutoMigrate — Prisma의 'prisma migrate dev' 에 해당
	// struct 구조를 보고 테이블이 없으면 생성, 컬럼이 없으면 추가
	err = database.AutoMigrate(&model.User{}, &model.Message{}) // Message 추가
	if err != nil {
		log.Fatal("마이그레이션 실패:", err)
	}

	DB = database
	log.Println("PostgreSQL 연결 + 마이그레이션 완료")
}
