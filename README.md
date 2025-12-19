# 家計簿アプリ（Household App）

Go + Gin + HTMX + Alpine.js + SQLite で作る家計簿Webアプリケーション。

## 技術スタック

### バックエンド
- **Go 1.21+** - プログラミング言語
- **Gin** - Webフレームワーク
- **GORM** - ORM
- **SQLite3** - データベース

### フロントエンド
- **HTMX** - SPA風UI（CDN経由）
- **Alpine.js** - リアクティブUI（CDN経由）
- **Bootstrap 5** - CSSフレームワーク（CDN経由）
- **Chart.js** - グラフ表示（CDN経由）

### インフラ
- **Docker / Docker Compose** - 開発環境
- **DevContainer** - VS Code開発環境

## セットアップ

### 前提条件
- Docker / Docker Compose がインストールされていること
- （オプション）VS Code + Dev Containers 拡張機能

### 起動方法

#### Docker Compose（本番ビルド）
```bash
docker-compose up --build
```

#### Docker Compose（開発モード）
```bash
docker-compose --profile dev up
```

#### DevContainer（VS Code）
1. VS Code でプロジェクトを開く
2. コマンドパレットから「Dev Containers: Reopen in Container」を選択
3. ターミナルで `go run main.go` を実行

### アクセス
- http://localhost:8080

## 環境変数

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `GIN_MODE` | Ginの動作モード（debug/release） | debug |
| `PORT` | サーバーポート | 8080 |
| `DB_PATH` | SQLiteファイルパス | /app/data/household.db |

## API エンドポイント

### ヘルスチェック
```
GET /health
```

### 収支記録（Transactions）
```
GET    /api/v1/transactions          # 一覧取得（ページネーション対応）
POST   /api/v1/transactions          # 新規作成
GET    /api/v1/transactions/:id      # 詳細取得
PUT    /api/v1/transactions/:id      # 更新
DELETE /api/v1/transactions/:id      # 削除
```

### カテゴリ（Categories）
```
GET    /api/v1/categories            # カテゴリ一覧
POST   /api/v1/categories            # カテゴリ追加
```

### 集計（Summary）
```
GET    /api/v1/summary/monthly       # 月次集計
GET    /api/v1/summary/category      # カテゴリ別集計
GET    /api/v1/summary/yearly        # 年次集計
```

### エクスポート
```
GET    /api/v1/export/csv            # CSV出力
```

## ディレクトリ構成

```
household-app/
├── main.go                 # エントリーポイント
├── go.mod / go.sum         # Go依存管理
├── Dockerfile              # コンテナ定義
├── docker-compose.yml      # 開発環境
├── .devcontainer/          # DevContainer設定
├── handlers/               # ルーティング処理
├── models/                 # データモデル
├── middleware/             # ミドルウェア
├── templates/              # HTMLテンプレート
├── static/                 # 静的ファイル
└── data/                   # SQLiteファイル（gitignore）
```

## 今後の実装予定

- [ ] 認証機能（JWT）
- [ ] フロントエンド詳細実装
- [ ] テストコード
- [ ] CI/CD設定

## ライセンス

MIT License
