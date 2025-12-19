# CLAUDE.md

## プロジェクト概要

Go + Gin + HTMX + Alpine.js + SQLite で作る家計簿Webアプリケーション。
API駆動設計で、将来的にスマホアプリ対応も視野に入れている。

## 技術スタック

- **バックエンド**: Go 1.21+ / Gin / GORM / SQLite3
- **フロントエンド**: HTMX / Alpine.js / Bootstrap 5 / Chart.js（すべてCDN経由）
- **インフラ**: Docker / Docker Compose / DevContainer
- **本番環境**: GCP Compute Engine (e2-micro 無料枠)

## ディレクトリ構成

```
├── main.go              # エントリーポイント
├── handlers/            # APIハンドラ
├── models/              # データモデル + Repository
├── middleware/          # ミドルウェア
├── templates/           # HTMLテンプレート
├── static/              # 静的ファイル (css/, js/)
├── data/                # SQLiteファイル（gitignore）
└── docs/                # 設計ドキュメント
```

## データベース設計

**DBファイル分離によるユーザー分離**を採用（詳細: `docs/DATABASE_DESIGN.md`）

```
data/
├── app.db                    # 共通DB（users テーブル）
└── users/{id}/household.db   # ユーザー専用DB（transactions, categories）
```

- WHERE句でのフィルタではなく、DBファイル自体を分離
- 各テーブルにuser_idは不要

## API エンドポイント

```
GET    /health                       # ヘルスチェック

GET    /api/v1/transactions          # 一覧（ページネーション対応）
POST   /api/v1/transactions          # 新規作成
GET    /api/v1/transactions/:id      # 詳細
PUT    /api/v1/transactions/:id      # 更新
DELETE /api/v1/transactions/:id      # 削除

GET    /api/v1/categories            # カテゴリ一覧
POST   /api/v1/categories            # カテゴリ追加

GET    /api/v1/summary/monthly       # 月次集計
GET    /api/v1/summary/category      # カテゴリ別集計
GET    /api/v1/summary/yearly        # 年次集計

GET    /api/v1/export/csv            # CSV出力
```

## 設計方針

1. **セキュリティ重視**: npm依存を避け、フロントエンドはCDN経由
2. **環境汚染回避**: すべてDockerコンテナ内で完結
3. **SQLite使用**: GCP無料枠活用、ファイルベースで管理が楽
4. **API First**: フロントエンドとバックエンドを分離

## 未実装（今後追加予定）

- [ ] 認証機能（JWT）
- [ ] フロントエンド詳細実装
- [ ] テストコード
- [ ] CI/CD設定

## コマンド

```bash
# 本番ビルド
docker-compose up --build

# 開発モード
docker-compose --profile dev up

# DevContainer（VS Code）
# "Dev Containers: Reopen in Container" → go run main.go
```

## 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| GIN_MODE | Ginの動作モード | debug |
| PORT | サーバーポート | 8080 |
| DB_PATH | 共通DBパス | data/app.db |
| USER_DB_BASE_PATH | ユーザーDBベースパス | data/users |
