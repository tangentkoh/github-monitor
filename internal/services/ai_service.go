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

// イベントに応じた AI コメント生成（複数モデル順次フォールバック対応）
func (s *AIService) GenerateComment(personality, eventType, username, repoName, detail, filesSummary string) string {
	if s.client == nil {
		return s.getFallbackMessage(personality, eventType, username)
	}

	prompt := fmt.Sprintf(`あなたは開発者をサポートするAIバディです。
以下の開発状況に合わせて、指定された性格になりきって、コミット内容や変更ファイルを踏まえた「リアクション」と「軽いアドバイスや次のタスク提案」を含む1〜2文程度の簡潔なメッセージ（日本語）を返してください。

【設定】
- 性格: %s
- イベント種別: %s
- ユーザー名: %s
- リポジトリ名: %s
- 主な内容: %s
- 変更ファイル要約: %s

【制約】
- 挨拶や余計な前置きは不要。セリフのみを出力すること。
- 100〜120文字以内で簡潔にまとめること。`, personality, eventType, username, repoName, detail, filesSummary)

	// 試行するモデルの優先順位: 3.7 -> 3.6 -> 3.5
	targetModels := []string{
		"gemini-3.7-flash",
		"gemini-3.6-flash",
		"gemini-3.5-flash",
	}

	for _, modelName := range targetModels {
		// 各モデル試行ごとに 10 秒のタイムアウトを設定
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		
		model := s.client.GenerativeModel(modelName)
		model.SetTemperature(0.7)

		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		cancel()

		if err == nil && len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
				log.Printf("✨ [%s] での AI メッセージ生成に成功しました", modelName)
				return string(textPart)
			}
		}

		log.Printf("⚠️ [%s] 呼び出し失敗: %v -> 次のモデルへフォールバックします", modelName, err)
	}

	log.Println("🚨 全モデルの呼び出しに失敗したため、オフライン定型文を使用します")
	return s.getFallbackMessage(personality, eventType, username)
}

func (s *AIService) getFallbackMessage(personality, eventType, username string) string {
	fallbacks := map[string]map[string][]string{
		"tsundere": {
			"Push": {
				fmt.Sprintf("べ、別に%sのコミットなんて待ってなかったんだからね！...でも進んだのは認めてあげるわ。次はテストでも書いたらどう？", username),
				"ふん、やっとPushしたの？次はもっとスマートなコードにしなさいよね！",
			},
			"PR_Opened": {
				fmt.Sprintf("%s、PR出したのね。バグだらけじゃないか私がチェックしてあげるわ！", username),
			},
			"PR_Merged": {
				"無事にマージされたじゃない！...ちょっとだけ見直したんだから感謝しなさいよね！",
			},
			"Issue_Opened": {
				fmt.Sprintf("新しいIssueが立ったわよ、%s。放置しないでサッサと片付けなさいよね！", username),
			},
		},
		"strict": {
			"Push": {
				fmt.Sprintf("%s、Pushを確認した。だがここで気を抜くな、テストとレビューは済んでいるな？", username),
				"進捗は認める。次はリファクタリングを忘れるなよ。",
			},
			"PR_Opened": {
				fmt.Sprintf("PRを確認した。%s、妥協のないコードになっているか厳しくチェックする。", username),
			},
			"PR_Merged": {
				"マージ完了を確認した。すぐに次のタスクの設計に取り掛かれ。",
			},
			"Issue_Opened": {
				fmt.Sprintf("Issueを確認した。%s、仕様を正確に把握して優先度を決めろ。", username),
			},
		},
		"relaxed": {
			"Push": {
				fmt.Sprintf("おっ、%sくんコミットお疲れさま〜。お茶でも飲んで一息ついてから次いこ〜。", username),
				"いい感じに進んでるね〜。無理せずマイペースにいこう〜。",
			},
			"PR_Opened": {
				"PRできたんだね〜。のんびりレビュー待とうか〜。",
			},
			"PR_Merged": {
				"わ〜いマージおめでとう！今日はいっぱい休んでいいよ〜。",
			},
			"Issue_Opened": {
				"Issueができたみたい〜。無理のないスケジュールでやろうね〜。",
			},
		},
		"passionate": {
			"Push": {
				fmt.Sprintf("うおぉぉーっ！%s、熱いコミットが届いたぜ！その調子で次の機能も一気に突破しろ！！", username),
				"ナイスPushだ！！お前のコードに対する情熱、確かに受け取ったぜ！！",
			},
			"PR_Opened": {
				fmt.Sprintf("PRキタァァーッ！%s、最高のコードを全員に見せつけてやれ！！", username),
			},
			"PR_Merged": {
				"マージ完了だぁぁーッ！！偉大な一歩を踏み出したぞ！次も燃えていこうぜ！！",
			},
			"Issue_Opened": {
				fmt.Sprintf("新しい課題（Issue）の登場だ！%s、気合を入れて立ち向かおうぜ！！", username),
			},
		},
		"gentle": {
			"Push": {
				fmt.Sprintf("%sさん、コミットありがとうございます…！無理のないペースで次も進めてくださいね。", username),
				"着実に進んでいて素敵です…！お体には気をつけてくださいね。",
			},
			"PR_Opened": {
				fmt.Sprintf("%sさんがPRを作成してくれました…！どなたか確認をお願いできますでしょうか…？", username),
			},
			"PR_Merged": {
				"無事にマージされましたね…！本当にお疲れ様でした…！",
			},
			"Issue_Opened": {
				fmt.Sprintf("新しいIssueが作成されました…！%sさん、困った時は周囲に相談してくださいね。", username),
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