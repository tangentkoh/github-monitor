package bot

import (
	"fmt"
	"log"
	"os"

	"github-monitor/internal/services"

	"github.com/bwmarrin/discordgo"
)

var AIServiceInstance *services.AIService

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "personality",
		Description: "AI バディの性格を変更します",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "mode",
				Description: "性格を選択してください",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "ツンデレ (Tsundere)", Value: "tsundere"},
					{Name: "厳しい / 鬼上司 (Strict)", Value: "strict"},
					{Name: "のんびり / 脱力 (Relaxed)", Value: "relaxed"},
					{Name: "熱血 / コーチ (Passionate)", Value: "passionate"},
					{Name: "温和 / 臆病 (Gentle)", Value: "gentle"},
				},
			},
		},
	},
	{
		Name:        "status",
		Description: "現在の AI バディの状態を確認します",
	},
	{
		Name:        "simulate",
		Description: "【デモ用】GitHubイベントを擬似的に発生させます",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "event",
				Description: "シミュレートするイベント",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Push (コード送信)", Value: "push"},
					{Name: "PR Opened (プルリク作成)", Value: "pr_opened"},
					{Name: "PR Merged (マージ達成)", Value: "pr_merged"},
					{Name: "Issue Opened (課題作成)", Value: "issue_opened"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "コミットメッセージやPRタイトル（任意）",
				Required:    false,
			},
		},
	},
}

func RegisterCommands(s *discordgo.Session) error {
	guildID := os.Getenv("DISCORD_GUILD_ID")

	for _, cmd := range Commands {
		// 🎯 guildID を渡すことでサーバー限定コマンドとして即時反映
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			return fmt.Errorf("コマンド登録失敗 [%s]: %w", cmd.Name, err)
		}
	}
	log.Printf("✅ Discord スラッシュコマンドを即時登録しました (Guild: %s)", guildID)
	return nil
}

func CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	switch data.Name {
	case "personality":
		newMode := data.Options[0].StringValue()
		SetPersonality(newMode)

		msg := fmt.Sprintf("🎭 AI バディの性格を **[%s]** に変更しました！", newMode)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: msg},
		})

	case "status":
		current := GetPersonality()
		msg := fmt.Sprintf("📊 **現在のステータス**\n• 性格モード: **%s**\n• 監視チャンネル: <#%s>\n• システム: 稼働中 (DBレス/オンメモリ)", current, ChannelID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: msg},
		})

	case "simulate":
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "⚡ 擬似 GitHub イベントを発火しました！"},
		})

		eventType := data.Options[0].StringValue()
		customMsg := "feat: ユーザー認証の実装"
		if len(data.Options) > 1 && data.Options[1].StringValue() != "" {
			customMsg = data.Options[1].StringValue()
		}

		personality := GetPersonality()

		switch eventType {
		case "push":
			aiComment := AIServiceInstance.GenerateComment(personality, "Push", "DemoUser", "github-monitor", customMsg, "cmd/server/main.go, internal/auth.go")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "github-monitor (Demo)", Inline: true},
				{Name: "👤 Author", Value: "DemoUser", Inline: true},
				{Name: "📁 変更ファイル", Value: "`cmd/server/main.go, internal/auth.go`", Inline: false},
				{Name: "💬 Commit Message", Value: fmt.Sprintf("```\n%s\n```", customMsg), Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("🚀 新しいコードが Push されました！ (Simulated)", "", 0x00FF88, fields)

		case "pr_opened":
			aiComment := AIServiceInstance.GenerateComment(personality, "PR_Opened", "DemoUser", "github-monitor", customMsg, "新規PR")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "github-monitor (Demo)", Inline: true},
				{Name: "👤 Created by", Value: "DemoUser", Inline: true},
				{Name: "📑 PR Title", Value: customMsg, Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("📢 新しい Pull Request が作成されました！ (Simulated)", "", 0x3498DB, fields)

		case "pr_merged":
			aiComment := AIServiceInstance.GenerateComment(personality, "PR_Merged", "DemoUser", "github-monitor", customMsg, "PRマージ")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "github-monitor (Demo)", Inline: true},
				{Name: "👤 Merged by", Value: "DemoUser", Inline: true},
				{Name: "🎉 PR Title", Value: customMsg, Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("🎉 Pull Request がマージされました！ (Simulated)", "", 0x9B59B6, fields)

		case "issue_opened":
			aiComment := AIServiceInstance.GenerateComment(personality, "Issue_Opened", "DemoUser", "github-monitor", customMsg, "Issue作成")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "github-monitor (Demo)", Inline: true},
				{Name: "👤 Author", Value: "DemoUser", Inline: true},
				{Name: "📌 Issue Title", Value: fmt.Sprintf("#12 %s", customMsg), Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("🎫 新しい Issue が作成されました！ (Simulated)", "", 0xE67E22, fields)
		}
	}
}