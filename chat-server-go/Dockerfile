# chat-server-go/Dockerfile

# ─────────────────────────────────────────
# 1단계 — 빌드
# Go 소스를 컴파일해서 바이너리 생성
# ─────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 의존성 먼저 복사 (캐시 활용)
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사 + 빌드
COPY . .
RUN go build -o server main.go

# ─────────────────────────────────────────
# 2단계 — 실행
# 빌드 결과물만 가져와서 실행
# 이미지 크기를 최소화 (golang 이미지는 크기가 큼)
# ─────────────────────────────────────────
FROM alpine:latest

WORKDIR /app

# 빌드된 바이너리만 복사
COPY --from=builder /app/server .

EXPOSE 3000

CMD ["./server"]
