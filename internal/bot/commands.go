package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// スラッシュコマンドの定義
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
}

// スラッシュコマンドの登録
func RegisterCommands(s *discordgo.Session) error {
	for _, cmd := range Commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("コマンド登録失敗 [%s]: %w", cmd.Name, err)
		}
	}
	log.Println("✅ Discord スラッシュコマンドを登録しました (/personality, /status)")
	return nil
}

// スラッシュコマンドのイベントハンドラ
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
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})

	case "status":
		current := GetPersonality()
		msg := fmt.Sprintf("📊 **現在のステータス**\n• 性格モード: **%s**\n• 監視チャンネル: <#%s>\n• システム: 稼働中 (DBレス/オンメモリ)", current, ChannelID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
	}
}