// internal/handler/websocket.go
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"chat-server-go/internal/db"
	"chat-server-go/internal/model"
)

// ─────────────────────────────────────────
// 1.2.1 연결 Map 관리
// Node.js: const onlineUsers = new Map()  와 동일
// key: userID, value: WebSocket 연결
// ─────────────────────────────────────────

var (
	// clients — 현재 연결된 유저들
	clients = make(map[uint]*websocket.Conn)

	// mu — 동시성 보호용 뮤텍스
	// goroutine 여러 개가 동시에 map을 읽고 쓰면 충돌남
	// Node.js는 싱글스레드라 이 문제가 없었음
	mu sync.Mutex
)

// upgrader — HTTP → WebSocket 업그레이드 설정
// Node.js: new Server(httpServer, { cors: ... })  와 유사
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 개발 환경: 모든 origin 허용
	},
}

// ─────────────────────────────────────────
// WebSocket 메시지 구조
// 클라이언트가 보내는 JSON 형태 정의
// ─────────────────────────────────────────

type WSMessage struct {
	ReceiverID uint   `json:"receiverId"`
	Content    string `json:"content"`
}

// ─────────────────────────────────────────
// 1.2.0 소켓 인증 + 1.2.2 메시지 전달
// GET /ws?token=JWT토큰
// ─────────────────────────────────────────

func HandleWebSocket(c *gin.Context) {
	// 쿼리 파라미터로 JWT 받기
	// Node.js socket.io: socket.handshake.auth.token  와 동일
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "토큰 없음"})
		return
	}

	// JWT 검증 — middleware와 동일한 로직
	userID, err := validateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 토큰"})
		return
	}

	// HTTP → WebSocket 업그레이드
	// Node.js: io.on('connection', (socket) => { ... })  진입점
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 업그레이드 실패:", err)
		return
	}

	// 연결 등록
	mu.Lock()
	clients[userID] = conn
	mu.Unlock()

	// 연결 시 모든 유저에게 온라인 목록 브로드캐스트
	broadcastOnlineUsers()

	log.Printf("유저 %d 연결됨 (현재 접속자: %d명)", userID, len(clients))

	// 연결 종료 시 정리
	// Node.js: socket.on('disconnect', () => { ... })  와 동일
	defer func() {
		mu.Lock()
		delete(clients, userID)
		mu.Unlock()
		// 종료 시 모든 유저에게 온라인 목록 브로드캐스트
		broadcastOnlineUsers()
		conn.Close()
		log.Printf("유저 %d 연결 끊김", userID)
	}()

	// ─────────────────────────────────────────
	// 메시지 수신 루프
	// Node.js: socket.on('message', async (data) => { ... })  와 동일
	// goroutine이 이 루프를 독립적으로 실행 — 블로킹해도 다른 연결에 영향 없음
	// ─────────────────────────────────────────
	for {
		// 메시지 읽기 — 클라이언트가 보낼 때까지 여기서 대기(블로킹)
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			// 연결이 끊기면 ReadMessage가 에러 반환 → 루프 종료
			break
		}

		// JSON 파싱
		var msg WSMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Println("메시지 파싱 실패:", err)
			continue
		}

		// DB 저장
		newMessage := model.Message{
			Content:    msg.Content,
			SenderID:   userID,
			ReceiverID: msg.ReceiverID,
		}
		db.DB.Create(&newMessage)

		// 수신자가 현재 연결되어 있으면 실시간 전송
		// Node.js: io.to(receiverSocketId).emit('message', data)  와 동일
		mu.Lock()
		receiverConn, online := clients[msg.ReceiverID]
		mu.Unlock()

		if online {
			receiverConn.WriteJSON(newMessage)
		}
	}
}

// GetOnlineUsers — 현재 온라인 유저 ID 목록 반환
// GET /users/online
func GetOnlineUsers(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock() // 함수 끝나면 자동 잠금 해제

	// map의 key(userID) 목록만 뽑아서 슬라이스로 만들기
	// Node.js: [...onlineUsers.keys()]  와 동일
	onlineIDs := make([]uint, 0, len(clients))
	for userID := range clients {
		onlineIDs = append(onlineIDs, userID)
	}

	c.JSON(http.StatusOK, gin.H{
		"onlineUsers": onlineIDs,
		"count":       len(onlineIDs),
	})
}

// broadcastOnlineUsers — 현재 온라인 목록을 모든 연결에 전송
func broadcastOnlineUsers() {
	mu.Lock()
	defer mu.Unlock()

	onlineIDs := make([]uint, 0, len(clients))
	for id := range clients {
		onlineIDs = append(onlineIDs, id)
	}

	msg, _ := json.Marshal(gin.H{
		"type":        "online_users",
		"onlineUsers": onlineIDs,
	})

	// 모든 연결에 전송
	for _, conn := range clients {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}
