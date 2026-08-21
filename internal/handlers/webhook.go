package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github-monitor/internal/bot"
	"github-monitor/internal/models"
	"github-monitor/internal/services"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	aiService *services.AIService
}

func NewWebhookHandler(aiService *services.AIService) *WebhookHandler {
	return &WebhookHandler{aiService: aiService}
}

// HMAC-SHA256 署名検証
func verifySignature(secret string, body []byte, signature string) bool {
	if signature == "" || secret == "" {
		return true // シークレット未設定時はスキップ（ローカルテスト用）
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

func (h *WebhookHandler) HandleGitHubWebhook(c echo.Context) error {
	event := c.Request().Header.Get("X-GitHub-Event")
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエストボディ読み込み失敗"})
	}

	// 署名検証
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	signature := c.Request().Header.Get("X-Hub-Signature-256")
	if !verifySignature(secret, bodyBytes, signature) {
		log.Println("🚨 [Security Alert] 不正なWebhook署名を検知してブロックしました！")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid signature"})
	}

	personality := bot.GetPersonality()

	switch event {
	case "push":
		return h.handlePushEvent(bodyBytes, personality, c)
	case "pull_request":
		return h.handlePREvent(bodyBytes, personality, c)
	case "ping":
		log.Println("🏓 GitHub Ping イベントを受信しました")
		_ = bot.SendEmbed("🏓 GitHub Webhook 疎通確認", "GitHub との Webhook 接続が正常に確認されました！", 0x00AAFF, nil)
		return c.JSON(http.StatusOK, map[string]string{"message": "pong"})
	default:
		return c.JSON(http.StatusOK, map[string]string{"message": "Event ignored"})
	}
}

// Push イベント処理
func (h *WebhookHandler) handlePushEvent(body []byte, personality string, c echo.Context) error {
	var payload models.GitHubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "JSONパース失敗"})
	}

	username := payload.Pusher.Name
	repoName := payload.Repository.Name
	commitMsg := "進捗の更新"
	if len(payload.Commits) > 0 {
		commitMsg = payload.Commits[len(payload.Commits)-1].Message
	}

	aiComment := h.aiService.GenerateComment(personality, "Push", username, repoName, commitMsg)

	fields := []*discordgo.MessageEmbedField{
		{Name: "📦 Repository", Value: repoName, Inline: true},
		{Name: "👤 Author", Value: username, Inline: true},
		{Name: "💬 Latest Commit", Value: fmt.Sprintf("`%s`", commitMsg), Inline: false},
		{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
	}

	_ = bot.SendEmbed("🚀 新しいコードが Push されました！", "", 0x00FF88, fields)
	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// Pull Request イベント処理
func (h *WebhookHandler) handlePREvent(body []byte, personality string, c echo.Context) error {
	var payload models.GitHubPRPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "JSONパース失敗"})
	}

	action := payload.Action
	repoName := payload.Repository.Name
	prTitle := payload.PullRequest.Title

	if action == "opened" {
		username := payload.PullRequest.User.Login
		aiComment := h.aiService.GenerateComment(personality, "PR_Opened", username, repoName, prTitle)

		fields := []*discordgo.MessageEmbedField{
			{Name: "📦 Repository", Value: repoName, Inline: true},
			{Name: "👤 Created by", Value: username, Inline: true},
			{Name: "📑 PR Title", Value: fmt.Sprintf("[%s](%s)", prTitle, payload.PullRequest.HTMLURL), Inline: false},
			{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
		}
		_ = bot.SendEmbed("📢 新しい Pull Request が作成されました！", "", 0x3498DB, fields)

	} else if action == "closed" && payload.PullRequest.Merged {
		username := payload.PullRequest.MergedBy.Login
		aiComment := h.aiService.GenerateComment(personality, "PR_Merged", username, repoName, prTitle)

		fields := []*discordgo.MessageEmbedField{
			{Name: "📦 Repository", Value: repoName, Inline: true},
			{Name: "👤 Merged by", Value: username, Inline: true},
			{Name: "🎉 PR Title", Value: prTitle, Inline: false},
			{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
		}
		_ = bot.SendEmbed("🎉 Pull Request がマージされました！", "", 0x9B59B6, fields)
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}