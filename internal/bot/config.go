package bot

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const configFilePath = "config.json"

// BotConfig はBot全体の設定を管理する構造体
type BotConfig struct {
	// GuildID (サーバーID) -> ChannelID のマッピング
	ChannelMapping map[string]string `json:"channel_mapping"`
	// グローバルなデフォルトチャンネルID（設定がない場合）
	DefaultChannelID string `json:"default_channel_id"`
	// 現在の性格
	Personality string `json:"personality"`
}

var (
	config *BotConfig
	cfgMu  sync.RWMutex
)

// LoadConfig は config.json ファイルから設定を読み込む
// ファイルが存在しない場合は、デフォルト設定で初期化される
func LoadConfig() error {
	cfgMu.Lock()
	defer cfgMu.Unlock()

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// ファイルが存在しない場合は、環境変数から初期化
			log.Printf("⚠️ %s が見つかりません。環境変数から初期設定を作成します", configFilePath)
			config = &BotConfig{
				ChannelMapping:   make(map[string]string),
				DefaultChannelID: os.Getenv("DISCORD_CHANNEL_ID"),
				Personality:      "tsundere",
			}
			// 環境変数に値がある場合は保存
			if config.DefaultChannelID != "" {
				if err := SaveConfigLocked(); err != nil {
					log.Printf("⚠️ 初期設定の保存に失敗: %v", err)
				}
			}
			return nil
		}
		return err
	}

	config = &BotConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		log.Printf("⚠️ config.json のパース失敗: %v", err)
		// パース失敗時も環境変数から復旧を試みる
		config = &BotConfig{
			ChannelMapping:   make(map[string]string),
			DefaultChannelID: os.Getenv("DISCORD_CHANNEL_ID"),
			Personality:      "tsundere",
		}
		return nil
	}

	log.Println("✅ config.json から設定を読み込みました")
	return nil
}

// SaveConfig は設定をファイルに保存
func SaveConfig() error {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	return SaveConfigLocked()
}

// SaveConfigLocked はロック済みの状態で設定を保存（内部用）
func SaveConfigLocked() error {
	if config == nil {
		return nil
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("❌ 設定のJSON化に失敗: %v", err)
		return err
	}

	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		log.Printf("❌ config.json への書き込み失敗: %v", err)
		return err
	}

	log.Println("💾 設定をファイルに保存しました")
	return nil
}

// GetChannelID は GuildID に対応するチャンネルIDを取得
// 見つからない場合はデフォルトチャンネルIDを返す
func GetChannelIDByGuild(guildID string) string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	if guildID != "" {
		if chID, ok := config.ChannelMapping[guildID]; ok && chID != "" {
			return chID
		}
	}
	return config.DefaultChannelID
}

// SetChannelIDForGuild はGuildIDに対応するチャンネルIDを設定して保存
func SetChannelIDForGuild(guildID, channelID string) error {
	cfgMu.Lock()
	defer cfgMu.Unlock()

	if config.ChannelMapping == nil {
		config.ChannelMapping = make(map[string]string)
	}

	config.ChannelMapping[guildID] = channelID
	// デフォルトチャンネルIDも更新（互換性のため）
	config.DefaultChannelID = channelID

	return SaveConfigLocked()
}

// GetPersonality は現在の性格を取得
func GetConfigPersonality() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if config == nil {
		return "tsundere"
	}
	return config.Personality
}

// SetPersonalityConfig は性格を設定して保存
func SetPersonalityConfig(personality string) error {
	cfgMu.Lock()
	defer cfgMu.Unlock()

	if config == nil {
		config = &BotConfig{
			ChannelMapping: make(map[string]string),
			Personality:    personality,
		}
	} else {
		config.Personality = personality
	}

	return SaveConfigLocked()
}
