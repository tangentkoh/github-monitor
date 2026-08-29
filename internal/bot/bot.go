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
	ChannelID          string
	mu                 sync.RWMutex
	Session            *discordgo.Session
)

func GetPersonality() string {
	mu.RLock()
	defer mu.RUnlock()
	return GetConfigPersonality()
}

func SetPersonality(p string) {
	mu.Lock()
	defer mu.Unlock()
	if err := SetPersonalityConfig(p); err != nil {
		log.Printf("⚠️ 性格の設定保存に失敗: %v", err)
	}
	CurrentPersonality = p
}

func GetChannelID() string {
	mu.RLock()
	defer mu.RUnlock()
	return ChannelID
}

// GetChannelIDForGuild は GuildID 別のチャンネルIDを取得
func GetChannelIDForGuild(guildID string) string {
	mu.RLock()
	defer mu.RUnlock()
	return GetChannelIDByGuild(guildID)
}

func SetChannelID(ch string) {
	mu.Lock()
	defer mu.Unlock()
	ChannelID = ch
}

func InitBot() (*discordgo.Session, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN が設定されていません")
	}

	// 設定ファイルから読み込み
	if err := LoadConfig(); err != nil {
		log.Printf("⚠️ 設定の読み込みに失敗: %v（環境変数で続行）", err)
	}

	// ローカルメモリ変数も同期
	ChannelID = GetChannelIDByGuild("")
	if ChannelID == "" {
		ChannelID = os.Getenv("DISCORD_CHANNEL_ID")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("Discord セッション作成エラー: %w", err)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("🤖 Discord Bot [%s#%s] が正常にオンラインになりました！", s.State.User.Username, s.State.User.Discriminator)
		if err := RegisterCommands(s); err != nil {
			log.Printf("⚠️ コマンド登録エラー: %v", err)
		}
	})

	dg.AddHandler(CommandHandler)

	err = dg.Open()
	if err != nil {
		return nil, fmt.Errorf("Discord 接続エラー: %w", err)
	}

	Session = dg
	return dg, nil
}

func SendEmbed(title, description string, color int, fields []*discordgo.MessageEmbedField) error {
	targetCh := GetChannelID()
	if Session == nil || targetCh == "" {
		return fmt.Errorf("Discord Bot が初期化されていないか、通知先チャンネルが未設定です")
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
	}

	_, err := Session.ChannelMessageSendEmbed(targetCh, embed)
	return err
}