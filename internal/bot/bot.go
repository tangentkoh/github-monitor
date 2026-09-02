package bot

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	CurrentPersonality = "tsundere"
	mu                 sync.RWMutex
	Session            *discordgo.Session
)

func GetPersonality() string {
	return GetConfigPersonality()
}

func SetPersonality(p string) {
	if err := SetPersonalityConfig(p); err != nil {
		log.Printf("⚠️ 性格の設定保存に失敗: %v", err)
	}
	mu.Lock()
	CurrentPersonality = p
	mu.Unlock()
}

func GetChannelID() string {
	return GetChannelIDByGuild("")
}

func GetChannelIDForGuild(guildID string) string {
	return GetChannelIDByGuild(guildID)
}

func InitBot() (*discordgo.Session, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN が設定されていません")
	}

	if err := LoadConfig(); err != nil {
		log.Printf("⚠️ 設定読み込み警告: %v", err)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("Discord セッション作成エラー: %w", err)
	}

	// 必要なインテントを設定
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// ハンドラー登録
	dg.AddHandler(CommandHandler)
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("🤖 Discord Bot [%s#%s] がオンラインになりました！", s.State.User.Username, s.State.User.Discriminator)
		if err := RegisterCommands(s); err != nil {
			log.Printf("⚠️ コマンド登録エラー: %v", err)
		}
	})

	if err := dg.Open(); err != nil {
		return nil, fmt.Errorf("Discord 接続エラー: %w", err)
	}

	Session = dg
	return dg, nil
}

// targetCh に config から最新のチャンネルIDを取得して送信
func SendEmbed(title, description string, color int, fields []*discordgo.MessageEmbedField) error {
	targetCh := GetChannelID()
	if Session == nil || targetCh == "" {
		return fmt.Errorf("Discord 未接続、または通知先チャンネルが未設定です (ChannelID: %s)", targetCh)
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
	}

	_, err := Session.ChannelMessageSendEmbed(targetCh, embed)
	if err != nil {
		log.Printf("🚨 Discord Embed 送信失敗 [Channel: %s]: %v", targetCh, err)
	}
	return err
}