package main

import (
	"log"
	"os"

	"github-monitor/internal/bot"
	"github-monitor/internal/handlers"
	"github-monitor/internal/services"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../../.env")
	}

	// Discord Bot 起動
	dg, err := bot.InitBot()
	if err != nil {
		log.Fatalf("❌ Discord Bot 起動失敗: %v", err)
	}
	defer dg.Close()

	// サービス & ハンドラー初期化
	aiService := services.NewAIService()
	bot.AIServiceInstance = aiService
	webhookHandler := handlers.NewWebhookHandler(aiService)

	// Echo サーバー初期化
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Webhook エンドポイント
	e.POST("/api/webhook", webhookHandler.HandleGitHubWebhook)
	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	e.Match([]string{"GET", "HEAD"}, "/", func(c echo.Context) error {
		return c.String(200, "GitHub Monitor is Running!")
	})
	e.Match([]string{"GET", "HEAD"}, "/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 HTTP サーバーがポート %s で待機開始...", port)
	e.Logger.Fatal(e.Start(":" + port))
}