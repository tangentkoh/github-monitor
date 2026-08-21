package bot

import (
	"fmt"
	"log"

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
		Name:        "setchannel",
		Description: "このチャンネルを GitHub の通知先に設定します",
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

// グローバルコマンドとして登録（第2引数を空文字にする）
func RegisterCommands(s *discordgo.Session) error {
	for _, cmd := range Commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("コマンド登録失敗 [%s]: %w", cmd.Name, err)
		}
	}
	log.Println("🌍 Discord グローバルスラッシュコマンドを登録しました")
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
		targetCh := GetChannelID()
		msg := fmt.Sprintf("📊 **現在のステータス**\n• 性格モード: **%s**\n• 監視チャンネル: <#%s>\n• システム: 稼働中 (Render常時デプロイ)", current, targetCh)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: msg},
		})

	case "setchannel":
		SetChannelID(i.ChannelID)
		msg := fmt.Sprintf("📢 通知先チャンネルを <#%s> に更新しました！", i.ChannelID)
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
		customMsg := "feat: 部活プロジェクト機能追加"
		if len(data.Options) > 1 && data.Options[1].StringValue() != "" {
			customMsg = data.Options[1].StringValue()
		}

		personality := GetPersonality()

		switch eventType {
		case "push":
			aiComment := AIServiceInstance.GenerateComment(personality, "Push", "DemoUser", "club-project", customMsg, "cmd/main.go, README.md")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "club-project (Demo)", Inline: true},
				{Name: "👤 Author", Value: "DemoUser", Inline: true},
				{Name: "📁 変更ファイル", Value: "`cmd/main.go, README.md`", Inline: false},
				{Name: "💬 Commit Message", Value: fmt.Sprintf("```\n%s\n```", customMsg), Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("🚀 新しいコードが Push されました！ (Simulated)", "", 0x00FF88, fields)

		case "pr_opened":
			aiComment := AIServiceInstance.GenerateComment(personality, "PR_Opened", "DemoUser", "club-project", customMsg, "新規PR")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "club-project (Demo)", Inline: true},
				{Name: "👤 Created by", Value: "DemoUser", Inline: true},
				{Name: "📑 PR Title", Value: customMsg, Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("📢 新しい Pull Request が作成されました！ (Simulated)", "", 0x3498DB, fields)

		case "pr_merged":
			aiComment := AIServiceInstance.GenerateComment(personality, "PR_Merged", "DemoUser", "club-project", customMsg, "PRマージ")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "club-project (Demo)", Inline: true},
				{Name: "👤 Merged by", Value: "DemoUser", Inline: true},
				{Name: "🎉 PR Title", Value: customMsg, Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("🎉 Pull Request がマージされました！ (Simulated)", "", 0x9B59B6, fields)

		case "issue_opened":
			aiComment := AIServiceInstance.GenerateComment(personality, "Issue_Opened", "DemoUser", "club-project", customMsg, "Issue作成")
			fields := []*discordgo.MessageEmbedField{
				{Name: "📦 Repository", Value: "club-project (Demo)", Inline: true},
				{Name: "👤 Author", Value: "DemoUser", Inline: true},
				{Name: "📌 Issue Title", Value: fmt.Sprintf("#42 %s", customMsg), Inline: false},
				{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
			}
			_ = SendEmbed("🎫 新しい Issue が作成されました！ (Simulated)", "", 0xE67E22, fields)
		}
	}
}