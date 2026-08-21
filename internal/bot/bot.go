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
	ChannelID          string
)

func GetPersonality() string {
	mu.RLock()
	defer mu.RUnlock()
	return CurrentPersonality
}

func SetPersonality(p string) {
	mu.Lock()
	defer mu.Unlock()
	CurrentPersonality = p
}

func InitBot() (*discordgo.Session, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	ChannelID = os.Getenv("DISCORD_CHANNEL_ID")

	if token == "" || ChannelID == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN または DISCORD_CHANNEL_ID が設定されていません")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("Discord セッション作成エラー: %w", err)
	}

	// イベントハンドラー追加
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("🤖 Discord Bot [%s#%s] が正常にオンラインになりました！", s.State.User.Username, s.State.User.Discriminator)
		// コマンド登録
		if err := RegisterCommands(s); err != nil {
			log.Printf("⚠️ コマンド登録エラー: %v", err)
		}
	})

	// スラッシュコマンド受信ハンドラー
	dg.AddHandler(CommandHandler)

	err = dg.Open()
	if err != nil {
		return nil, fmt.Errorf("Discord 接続エラー: %w", err)
	}

	Session = dg
	return dg, nil
}

func SendEmbed(title, description string, color int, fields []*discordgo.MessageEmbedField) error {
	if Session == nil || ChannelID == "" {
		return fmt.Errorf("Discord Bot が初期化されていません")
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
	}

	_, err := Session.ChannelMessageSendEmbed(ChannelID, embed)
	return err
}