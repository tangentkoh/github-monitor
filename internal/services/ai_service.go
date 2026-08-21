package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIService struct {
	client *genai.Client
}

func NewAIService() *AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("⚠️ GEMINI_API_KEY が未設定です。オフライン定型文モードで動作します。")
		return &AIService{client: nil}
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("⚠️ Gemini クライアント初期化失敗: %v (オフラインモードで動作)", err)
		return &AIService{client: nil}
	}

	log.Println("🤖 Gemini API に正常に接続しました！")
	return &AIService{client: client}
}

// イベントに応じた AI コメント生成
func (s *AIService) GenerateComment(personality, eventType, username, repoName, detail string) string {
	if s.client == nil {
		return s.getFallbackMessage(personality, eventType, username)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	model := s.client.GenerativeModel("gemini-2.5-flash")
	model.SetTemperature(0.7)

	prompt := fmt.Sprintf(`あなたは開発者をサポートするAIバディです。
以下の状況に合わせて、指定された性格になりきって1〜2文程度の簡潔なメッセージ（日本語）を返してください。

【設定】
- 性格: %s
  - tsundere: ツンデレ。素直になれないが本当は進捗を喜んでいる。
  - strict: 厳しい鬼上司。妥協を許さずコードの質にシビアだが期待している。
  - relaxed: のんびり脱力系。まったりと優しく労う。
  - passionate: 熱血コーチ。ハイテンションで情熱的に応援する。
  - gentle: 温和・臆病。控えめで優しく丁寧にフォローする。
- イベント種別: %s (Push / PR作成 / PRマージ)
- ユーザー名: %s
- リポジトリ名: %s
- 詳細情報: %s

【制約】
- 挨拶や余計な前置きは不要。セリフのみを出力すること。
- 100文字以内で簡潔に。`, personality, eventType, username, repoName, detail)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil || len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Printf("⚠️ Gemini API 呼び出し失敗: %v -> オフライン定型文を使用", err)
		return s.getFallbackMessage(personality, eventType, username)
	}

	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(textPart)
	}

	return s.getFallbackMessage(personality, eventType, username)
}

// オフライン・エラー時のローカル定型文（性格 x イベント）
func (s *AIService) getFallbackMessage(personality, eventType, username string) string {
	fallbacks := map[string]map[string][]string{
		"tsundere": {
			"Push": {
				fmt.Sprintf("べ、別に%sのコミットなんて待ってなかったんだからね！…でも進んだのは認めてあげるわ。", username),
				"ふん、やっとPushしたの？次はもっとスマートなコードにしなさいよね！",
			},
			"PR_Opened": {
				fmt.Sprintf("%s、PR出したのね。バグだらけじゃないか私がチェックしてあげるわ！", username),
			},
			"PR_Merged": {
				"無事にマージされたじゃない！…ちょっとだけ見直したんだから感謝しなさいよね！",
			},
		},
		"strict": {
			"Push": {
				fmt.Sprintf("%s、Pushを確認した。だがここで気を抜くな、テストは通っているな？", username),
				"進捗は出たようだな。次はリファクタリングを忘れるなよ。",
			},
			"PR_Opened": {
				fmt.Sprintf("PRを確認した。%s、妥協のないコードになっているかレビューを行う。", username),
			},
			"PR_Merged": {
				"マージ完了を確認した。すぐに次のタスクの設計に取り掛かれ。",
			},
		},
		"relaxed": {
			"Push": {
				fmt.Sprintf("おっ、%sくんコミットお疲れさま〜。お茶でも飲んで一服しよ〜。", username),
				"いい感じに進んでるね〜。無理せずマイペースにいこう〜。",
			},
			"PR_Opened": {
				"PRできたんだね〜。のんびりレビュー待とうか〜。",
			},
			"PR_Merged": {
				"わ〜いマージおめでとう！今日はいっぱい休んでいいよ〜。",
			},
		},
		"passionate": {
			"Push": {
				fmt.Sprintf("うおぉぉーっ！%s、熱いコミットが届いたぜ！その調子で限界を突破しろ！！", username),
				"ナイスPushだ！！お前のコードに対する情熱、確かに受け取ったぜ！！",
			},
			"PR_Opened": {
				fmt.Sprintf("PRキタァァーッ！%s、最高のコードを全員に見せつけてやれ！！", username),
			},
			"PR_Merged": {
				"マージ完了だぁぁーッ！！偉大な一歩を踏み出したぞ！次も燃えていこうぜ！！",
			},
		},
		"gentle": {
			"Push": {
				fmt.Sprintf("%sさん、コミットありがとうございます…！無理のないペースで頑張ってくださいね。", username),
				"着実に進んでいて素敵です…！お体には気をつけてくださいね。",
			},
			"PR_Opened": {
				fmt.Sprintf("%sさんがPRを作成してくれました…！どなたか確認をお願いできますでしょうか…？", username),
			},
			"PR_Merged": {
				"無事にマージされましたね…！本当にお疲れ様でした…！",
			},
		},
	}

	eventMap, ok := fallbacks[personality]
	if !ok {
		eventMap = fallbacks["tsundere"]
	}

	messages, ok := eventMap[eventType]
	if !ok || len(messages) == 0 {
		return fmt.Sprintf("進捗を確認しました！（%sさん: %s）", username, eventType)
	}

	return messages[rand.Intn(len(messages))]
}