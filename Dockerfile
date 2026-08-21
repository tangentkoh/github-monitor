# ビルドステージ
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server cmd/server/main.go

# 実行ステージ
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Tokyo

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]