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
		// GuildID を取得してチャンネルIDを永続化
		guildID := i.GuildID
		if guildID == "" {
			guildID = "default"
		}
		
		if err := SetChannelIDForGuild(guildID, i.ChannelID); err != nil {
			log.Printf("⚠️ チャンネル設定の保存に失敗: %v", err)
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("⚠️ チャンネル設定の保存に失敗しました: %v", err),
				},
			})
			return
		}

		msg := fmt.Sprintf("📢 通知先チャンネルを <#%s> に更新しました！（永続化済み）", i.ChannelID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: msg},
		})

	case "simulate":
		// 1. 3秒タイムアウトを防ぐため即座に Deferred 応答
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			log.Printf("⚠️ InteractionRespond エラー: %v", err)
			return
		}

		// 2. バックグラウンドで生成して上書き更新
		go func() {
			eventType := data.Options[0].StringValue()
			customMsg := "feat: テスト機能の実装"
			if len(data.Options) > 1 && data.Options[1].StringValue() != "" {
				customMsg = data.Options[1].StringValue()
			}

			personality := GetPersonality()
			var embed *discordgo.MessageEmbed

			switch eventType {
			case "push":
				aiComment := AIServiceInstance.GenerateComment(personality, "Push", "DemoUser", "test-repo", customMsg, "cmd/server/main.go")
				fields := []*discordgo.MessageEmbedField{
					{Name: "📦 Repository", Value: "test-repo (Demo)", Inline: true},
					{Name: "👤 Author", Value: "DemoUser", Inline: true},
					{Name: "📁 変更ファイル", Value: "`cmd/server/main.go`", Inline: false},
					{Name: "💬 Commit Message", Value: fmt.Sprintf("```\n%s\n```", customMsg), Inline: false},
					{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
				}
				embed = &discordgo.MessageEmbed{
					Title:  "🚀 新しいコードが Push されました！ (Simulated)",
					Color:  0x00FF88,
					Fields: fields,
				}

			case "pr_opened":
				aiComment := AIServiceInstance.GenerateComment(personality, "PR_Opened", "DemoUser", "test-repo", customMsg, "新規PR")
				fields := []*discordgo.MessageEmbedField{
					{Name: "📦 Repository", Value: "test-repo (Demo)", Inline: true},
					{Name: "👤 Created by", Value: "DemoUser", Inline: true},
					{Name: "📑 PR Title", Value: customMsg, Inline: false},
					{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
				}
				embed = &discordgo.MessageEmbed{
					Title:  "📢 新しい Pull Request が作成されました！ (Simulated)",
					Color:  0x3498DB,
					Fields: fields,
				}

			case "pr_merged":
				aiComment := AIServiceInstance.GenerateComment(personality, "PR_Merged", "DemoUser", "test-repo", customMsg, "PRマージ")
				fields := []*discordgo.MessageEmbedField{
					{Name: "📦 Repository", Value: "test-repo (Demo)", Inline: true},
					{Name: "👤 Merged by", Value: "DemoUser", Inline: true},
					{Name: "🎉 PR Title", Value: customMsg, Inline: false},
					{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
				}
				embed = &discordgo.MessageEmbed{
					Title:  "🎉 Pull Request がマージされました！ (Simulated)",
					Color:  0x9B59B6,
					Fields: fields,
				}

			case "issue_opened":
				aiComment := AIServiceInstance.GenerateComment(personality, "Issue_Opened", "DemoUser", "test-repo", customMsg, "Issue作成")
				fields := []*discordgo.MessageEmbedField{
					{Name: "📦 Repository", Value: "test-repo (Demo)", Inline: true},
					{Name: "👤 Author", Value: "DemoUser", Inline: true},
					{Name: "📌 Issue Title", Value: fmt.Sprintf("#1 %s", customMsg), Inline: false},
					{Name: fmt.Sprintf("🤖 AI バディ (%s)", personality), Value: aiComment, Inline: false},
				}
				embed = &discordgo.MessageEmbed{
					Title:  "🎫 新しい Issue が作成されました！ (Simulated)",
					Color:  0xE67E22,
					Fields: fields,
				}
			}

			// 「考え中...」を生成結果のEmbedで上書き更新
			_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				log.Printf("🚨 InteractionResponseEdit 失敗: %v", err)
			}
		}()
	}
}