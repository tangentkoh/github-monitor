# GitHub Monitor Bot 🤖

GitHubのイベント（Push, PR, Issue）をWebhookで受信し、**Gemini API**を用いて「設定された性格」に基づいたリアクション・助言コメントを生成して**Discordチャンネルに通知**するGo言語製のBot。

## 📋 主な特徴

- **複数性格対応**: ツンデレ、熱血、のんびり、温和、厳しい など5つの性格から選択可能
- **GitHub Webhookに対応**: Push, PR作成・マージ, Issue作成 の3種類のイベントに対応
- **AI生成コメント**: Gemini APIで性格に合わせた自動コメント生成
- **Discord統合**: Embed形式で見栄えの良い通知
- **設定の永続化**: チャンネルIDと性格設定をJSONで保存（再起動後も保持）
- **複数サーバー対応**: GuildID別のチャンネルマッピング

## 🛠️ 技術スタック

| 項目             | 詳細                                                          |
| ---------------- | ------------------------------------------------------------- |
| **言語**         | Go 1.25                                                       |
| **HTTPサーバー** | Echo v4                                                       |
| **Discord連携**  | discordgo v0.29.0                                             |
| **AI API**       | Google Generative AI (`gemini-3.6-flash`, `gemini-3.5-flash`) |
| **デプロイ**     | Docker on Render                                              |
| **監視**         | UptimeRobot (KeepAlive用)                                     |

## 📂 ディレクトリ構成

```
github-monitor/
├── cmd/
│   └── server/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── bot/
│   │   ├── bot.go               # Discord初期化・設定管理
│   │   ├── config.go            # 設定の永続化・GuildID管理
│   │   └── commands.go          # スラッシュコマンド定義
│   ├── handlers/
│   │   └── webhook.go           # GitHub Webhook処理
│   ├── models/
│   │   └── github.go            # Webhookペイロード定義
│   └── services/
│       └── ai_service.go        # Gemini API統合
├── Dockerfile                   # マルチステージビルド
├── go.mod / go.sum
├── .env                         # 環境変数（Git追跡外）
├── config.json                  # 設定ファイル（自動生成、Git追跡外）
└── README.md

```

## 🚀 セットアップ & 実行

### 前提条件

- Go 1.25以上
- Discord Bot Token
- Google AI Studio API Key
- GitHub Webhook Secret

### インストール

```bash
git clone https://github.com/kousuke-koh/github-monitor.git
cd github-monitor

# 依存パッケージのインストール
go mod tidy

# 環境変数の設定
cp .env.example .env  # または .env を作成
# .env に以下を記入:
# PORT=8080
# DISCORD_BOT_TOKEN=your_token
# DISCORD_CHANNEL_ID=your_channel_id
# GEMINI_API_KEY=your_api_key
# GITHUB_WEBHOOK_SECRET=your_secret
```

### ローカル実行

```bash
go run ./cmd/server
```

### Docker実行

```bash
docker build -t github-monitor .
docker run -p 8080:8080 \
  -e DISCORD_BOT_TOKEN=your_token \
  -e DISCORD_CHANNEL_ID=your_channel_id \
  -e GEMINI_API_KEY=your_api_key \
  -e GITHUB_WEBHOOK_SECRET=your_secret \
  github-monitor
```

## 📱 Discord コマンド

Bot がオンラインになると、以下のスラッシュコマンドが利用可能になります：

### `/personality [mode]`

AI バディの性格を変更します。

**選択肢:**

- `tsundere` - ツンデレ（デレ要素あり）
- `strict` - 厳しい / 鬼上司
- `relaxed` - のんびり / 脱力
- `passionate` - 熱血 / コーチ
- `gentle` - 温和 / 臆病

```
/personality tsundere
→ "🎭 AI バディの性格を [tsundere] に変更しました！"
```

### `/status`

現在の AI バディの状態を確認します。

```
/status
→ 📊 **現在のステータス**
  • 性格モード: **tsundere**
  • 監視チャンネル: <#123456789>
  • システム: 稼働中
```

### `/setchannel`

このコマンドを実行したチャンネルを GitHub 通知先に設定します。

```
/setchannel
→ "📢 通知先チャンネルを <#987654321> に更新しました！（永続化済み）"
```

**特徴:**

- GuildID別にチャンネルを管理
- 設定は `config.json` に自動保存
- 再起動後も設定が保持される

### `/simulate [event] [message?]`

【デモ用】GitHubイベントを擬似的に発生させて通知をテストします。

**イベントタイプ:**

- `push` - コード送信
- `pr_opened` - プルリク作成
- `pr_merged` - マージ達成
- `issue_opened` - 課題作成

```
/simulate push "feat: 新機能実装"
→ Embedメッセージが送信される
```

## 🔌 GitHub Webhook設定

### セットアップ手順

1. **GitHub リポジトリ** → Settings → Webhooks
2. **Add webhook** をクリック
3. 以下を入力：
   - **Payload URL**: `https://your-render-url.onrender.com/api/webhook`
   - **Content type**: `application/json`
   - **Secret**: `GITHUB_WEBHOOK_SECRET` の値
   - **Events**:
     - ✅ Push events
     - ✅ Pull requests
     - ✅ Issues

4. **Add webhook** をクリック

### テスト

Webhook が正常に動作しているか確認：

```bash
# Renderダッシュボード → Logs でリアルタイムログを確認
# または、GitHub リポジトリで Push / PR を作成してテスト
```

## 📊 通知フォーマット

### Push イベント例

```
🚀 新しいコードが Push されました！

📦 Repository: example-repo
👤 Author: kousuke-koh
📁 変更ファイル: src/main.go, go.mod
💬 Commit Message: feat: 新機能実装

🤖 AI バディ (tsundere)
べ、別にkousuke-kohのコミットなんて待ってなかったんだからね！
...でも進んだのは認めてあげるわ。次はテストでも書いたらどう？
```

### PR / Issue イベント

類似のフォーマットで、イベントの種類に応じた情報を表示します。

## ⚙️ 環境変数

| 変数                    | 説明                                     | 必須 |
| ----------------------- | ---------------------------------------- | ---- |
| `PORT`                  | HTTPサーバーのポート（デフォルト: 8080） | ❌   |
| `DISCORD_BOT_TOKEN`     | Discord Bot トークン                     | ✅   |
| `DISCORD_CHANNEL_ID`    | 初期通知先チャンネルID                   | ✅   |
| `GEMINI_API_KEY`        | Google AI Studio API キー                | ✅   |
| `GITHUB_WEBHOOK_SECRET` | GitHub Webhook署名検証用シークレット     | ✅   |

## 📁 設定の永続化

設定は自動的に `config.json` に保存されます：

```json
{
  "channel_mapping": {
    "12345678": "98765432",
    "default": "11111111"
  },
  "default_channel_id": "11111111",
  "personality": "tsundere"
}
```

**特徴:**

- `channel_mapping`: GuildID ごとのチャンネルマッピング
- `default_channel_id`: デフォルトチャンネル（複数サーバー未設定時の代替）
- `personality`: 現在の性格設定

**注意:** `config.json` は `.gitignore` に含まれているため、リポジトリに含まれません。

## 🐛 トラブルシューティング

### Gemini API が応答しない

**症状**: AI コメントが常に定型文（オフラインモード）

**原因と対策:**

1. API キーが正しいか確認
   ```bash
   echo $GEMINI_API_KEY
   ```
2. ネットワーク接続を確認
3. ログを確認
   ```
   ⚠️ [gemini-3.6-flash] 呼び出し失敗: ...
   ```

### チャンネル設定が反映されない

**症状**: `/setchannel` で設定しても通知が別チャンネルに飛ぶ

**原因と対策:**

1. `config.json` が存在するか確認
   ```bash
   ls -la config.json
   ```
2. ディレクトリの書き込み権限を確認
3. `/status` で現在の設定を確認

### Discord Bot がオフラインになる

**症状**: Bot がDiscordに表示されない

**原因と対策:**

1. Bot Token が正しいか確認
2. ログでエラーメッセージを確認
3. Render のログをチェック

## 🚀 デプロイメント（Render）

### 手順

1. **GitHub にプッシュ**

   ```bash
   git add .
   git commit -m "Update: Gemini API 最新モデル対応"
   git push
   ```

2. **Render にデプロイ**
   - Render のダッシュボード → 該当 Web Service
   - 自動デプロイが実行される

3. **環境変数を設定**
   - Settings → Environment Variables
   - 上記の必須変数を入力

4. **サービスが起動したら**
   - Logs でスタートアップメッセージを確認
   - Discord で `/status` コマンドを実行

### Render Free Tier の制限

- インスタンス再起動時に `/tmp` が削除される可能性
  - 本番環境では SQLite や外部ストレージの導入を検討
- メモリ: 512 MB
- CPU: 共有

## 📝 修正履歴

### v2.0 (2025年08月)

✅ **Gemini API モデル更新**

- 古いモデルID (`gemini-2.0/1.5-flash`) を廃止
- 最新のモデル (`gemini-3.6-flash`, `gemini-3.5-flash`) に対応
- エラーログを詳細化

✅ **設定永続化機能追加**

- `config.json` でチャンネルID と性格設定を保存
- 再起動後も設定が保持される
- GuildID別のマッピング対応

### v1.0 (初期版)

- GitHub Webhook 受信
- Discord Embed 通知
- AI コメント生成

## 📚 参考資料

- [Google Generative AI - Models](https://ai.google.dev/models)
- [Discord.go Documentation](https://pkg.go.dev/github.com/bwmarrin/discordgo)
- [Echo Web Framework](https://echo.labstack.com/)
- [GitHub Webhooks](https://docs.github.com/en/developers/webhooks-and-events/webhooks/creating-webhooks)

## 📄 ライセンス

MIT License

## 🤝 貢献

改善提案・バグ報告は Issue を作成してください。

---

**作成者**: kousuke-koh  
**最終更新**: 2025年08月
