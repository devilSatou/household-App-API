# API仕様書

## 概要

家計簿アプリのREST API仕様書です。

- **ベースURL**: `http://localhost:8080`
- **APIバージョン**: v1
- **認証**: 現在は `X-User-ID` ヘッダーまたは環境変数 `USER_ID` で指定（JWT認証は今後実装予定）

---

## 共通仕様

### リクエストヘッダー

| ヘッダー | 必須 | 説明 |
|----------|------|------|
| `Content-Type` | POST/PUT時 | `application/json` |
| `X-User-ID` | 任意 | ユーザーID（指定しない場合はデフォルトユーザー1） |

### レスポンス形式

すべてのレスポンスはJSON形式です。

**成功時**:
```json
{
  "data": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

**エラー時**:
```json
{
  "error": "エラーメッセージ"
}
```

### HTTPステータスコード

| コード | 説明 |
|--------|------|
| 200 | 成功 |
| 201 | 作成成功 |
| 400 | リクエスト不正 |
| 404 | リソースが見つからない |
| 500 | サーバーエラー |

---

## エンドポイント一覧

### ヘルスチェック

#### GET /health

サーバーの稼働状態を確認します。

**リクエスト例**:
```bash
curl http://localhost:8080/health
```

**レスポンス例**:
```json
{
  "status": "ok"
}
```

---

## 収支記録（Transactions）

### GET /api/v1/transactions

収支記録の一覧を取得します。

**クエリパラメータ**:

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|------------|------|------|------------|------|
| `limit` | int | 任意 | 20 | 取得件数（1以上） |
| `offset` | int | 任意 | 0 | 取得開始位置 |
| `type` | string | 任意 | - | `income` または `expense` でフィルタ |
| `start_date` | string | 任意 | - | 開始日（YYYY-MM-DD） |
| `end_date` | string | 任意 | - | 終了日（YYYY-MM-DD） |

**リクエスト例**:
```bash
# 基本
curl http://localhost:8080/api/v1/transactions

# ページネーション
curl "http://localhost:8080/api/v1/transactions?limit=10&offset=20"

# 支出のみ
curl "http://localhost:8080/api/v1/transactions?type=expense"

# 期間指定
curl "http://localhost:8080/api/v1/transactions?start_date=2024-01-01&end_date=2024-01-31"
```

**レスポンス例**:
```json
{
  "data": [
    {
      "id": 1,
      "category_id": 1,
      "amount": 1500,
      "type": "expense",
      "date": "2024-01-15T00:00:00Z",
      "memo": "ランチ",
      "created_at": "2024-01-15T12:30:00Z",
      "updated_at": "2024-01-15T12:30:00Z",
      "category": {
        "id": 1,
        "name": "食費",
        "type": "expense",
        "created_at": "2024-01-01T00:00:00Z"
      }
    }
  ],
  "total": 150,
  "limit": 20,
  "offset": 0
}
```

---

### POST /api/v1/transactions

収支記録を新規作成します。

**リクエストボディ**:

| フィールド | 型 | 必須 | 説明 |
|------------|------|------|------|
| `category_id` | uint | 必須 | カテゴリID |
| `amount` | int | 必須 | 金額（1以上の正の整数） |
| `type` | string | 必須 | `income`（収入）または `expense`（支出） |
| `date` | string | 必須 | 日付（YYYY-MM-DD形式） |
| `memo` | string | 任意 | メモ |

**リクエスト例**:
```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": 1,
    "amount": 1500,
    "type": "expense",
    "date": "2024-01-15",
    "memo": "ランチ"
  }'
```

**レスポンス例** (201 Created):
```json
{
  "id": 1,
  "category_id": 1,
  "amount": 1500,
  "type": "expense",
  "date": "2024-01-15T00:00:00Z",
  "memo": "ランチ",
  "created_at": "2024-01-15T12:30:00Z",
  "updated_at": "2024-01-15T12:30:00Z"
}
```

---

### GET /api/v1/transactions/:id

収支記録の詳細を取得します。

**パスパラメータ**:

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | uint | 収支記録ID |

**リクエスト例**:
```bash
curl http://localhost:8080/api/v1/transactions/1
```

**レスポンス例**:
```json
{
  "id": 1,
  "category_id": 1,
  "amount": 1500,
  "type": "expense",
  "date": "2024-01-15T00:00:00Z",
  "memo": "ランチ",
  "created_at": "2024-01-15T12:30:00Z",
  "updated_at": "2024-01-15T12:30:00Z",
  "category": {
    "id": 1,
    "name": "食費",
    "type": "expense",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### PUT /api/v1/transactions/:id

収支記録を更新します。

**パスパラメータ**:

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | uint | 収支記録ID |

**リクエストボディ**: POST と同じ

**リクエスト例**:
```bash
curl -X PUT http://localhost:8080/api/v1/transactions/1 \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": 1,
    "amount": 2000,
    "type": "expense",
    "date": "2024-01-15",
    "memo": "ランチ（修正）"
  }'
```

**レスポンス例**:
```json
{
  "id": 1,
  "category_id": 1,
  "amount": 2000,
  "type": "expense",
  "date": "2024-01-15T00:00:00Z",
  "memo": "ランチ（修正）",
  "created_at": "2024-01-15T12:30:00Z",
  "updated_at": "2024-01-15T13:00:00Z"
}
```

---

### DELETE /api/v1/transactions/:id

収支記録を削除します。

**パスパラメータ**:

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | uint | 収支記録ID |

**リクエスト例**:
```bash
curl -X DELETE http://localhost:8080/api/v1/transactions/1
```

**レスポンス例**:
```json
{
  "message": "Transaction deleted"
}
```

---

## カテゴリ（Categories）

### GET /api/v1/categories

カテゴリ一覧を取得します。

**クエリパラメータ**:

| パラメータ | 型 | 必須 | 説明 |
|------------|------|------|------|
| `type` | string | 任意 | `income` または `expense` でフィルタ |

**リクエスト例**:
```bash
# 全カテゴリ
curl http://localhost:8080/api/v1/categories

# 支出カテゴリのみ
curl "http://localhost:8080/api/v1/categories?type=expense"
```

**レスポンス例**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "食費",
      "type": "expense",
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": 13,
      "name": "給与",
      "type": "income",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST /api/v1/categories

カテゴリを新規作成します。

**リクエストボディ**:

| フィールド | 型 | 必須 | 説明 |
|------------|------|------|------|
| `name` | string | 必須 | カテゴリ名（ユニーク） |
| `type` | string | 必須 | `income`（収入）または `expense`（支出） |

**リクエスト例**:
```bash
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{
    "name": "サブスク",
    "type": "expense"
  }'
```

**レスポンス例** (201 Created):
```json
{
  "id": 17,
  "name": "サブスク",
  "type": "expense",
  "created_at": "2024-01-15T12:00:00Z"
}
```

---

## 集計（Summary）

### GET /api/v1/summary/monthly

月次集計を取得します。

**クエリパラメータ**:

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|------------|------|------|------------|------|
| `year` | int | 任意 | 今年 | 年（例: 2024） |
| `month` | int | 任意 | 今月 | 月（1-12） |

**リクエスト例**:
```bash
# 今月
curl http://localhost:8080/api/v1/summary/monthly

# 指定月
curl "http://localhost:8080/api/v1/summary/monthly?year=2024&month=1"
```

**レスポンス例**:
```json
{
  "year": 2024,
  "month": 1,
  "income": 300000,
  "expense": 150000,
  "balance": 150000
}
```

---

### GET /api/v1/summary/category

カテゴリ別集計を取得します。

**クエリパラメータ**:

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|------------|------|------|------------|------|
| `start_date` | string | 任意 | 今月1日 | 開始日（YYYY-MM-DD） |
| `end_date` | string | 任意 | 今月末日 | 終了日（YYYY-MM-DD） |

**リクエスト例**:
```bash
# 今月
curl http://localhost:8080/api/v1/summary/category

# 期間指定
curl "http://localhost:8080/api/v1/summary/category?start_date=2024-01-01&end_date=2024-12-31"
```

**レスポンス例**:
```json
{
  "data": [
    {
      "category_id": 1,
      "category_name": "食費",
      "type": "expense",
      "total": 45000
    },
    {
      "category_id": 13,
      "category_name": "給与",
      "type": "income",
      "total": 300000
    }
  ],
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

---

### GET /api/v1/summary/yearly

年次集計を取得します。月別の内訳も含みます。

**クエリパラメータ**:

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|------------|------|------|------------|------|
| `year` | int | 任意 | 今年 | 年（例: 2024） |

**リクエスト例**:
```bash
# 今年
curl http://localhost:8080/api/v1/summary/yearly

# 指定年
curl "http://localhost:8080/api/v1/summary/yearly?year=2024"
```

**レスポンス例**:
```json
{
  "year": 2024,
  "income": 3600000,
  "expense": 1800000,
  "balance": 1800000,
  "monthly": [
    {
      "year": 2024,
      "month": 1,
      "income": 300000,
      "expense": 150000,
      "balance": 150000
    },
    {
      "year": 2024,
      "month": 2,
      "income": 300000,
      "expense": 140000,
      "balance": 160000
    }
  ]
}
```

---

## エクスポート（Export）

### GET /api/v1/export/csv

収支記録をCSV形式でダウンロードします。

**クエリパラメータ**:

| パラメータ | 型 | 必須 | 説明 |
|------------|------|------|------|
| `start_date` | string | 任意 | 開始日（YYYY-MM-DD） |
| `end_date` | string | 任意 | 終了日（YYYY-MM-DD） |
| `type` | string | 任意 | `income` または `expense` でフィルタ |

**リクエスト例**:
```bash
# 全件ダウンロード
curl -O http://localhost:8080/api/v1/export/csv

# 期間指定
curl -o transactions.csv "http://localhost:8080/api/v1/export/csv?start_date=2024-01-01&end_date=2024-12-31"

# 支出のみ
curl -o expenses.csv "http://localhost:8080/api/v1/export/csv?type=expense"
```

**レスポンス**:

- Content-Type: `text/csv; charset=utf-8`
- Content-Disposition: `attachment; filename=transactions_YYYYMMDD.csv`

**CSVフォーマット**:
```csv
ID,日付,種別,カテゴリ,金額,メモ,作成日時
1,2024-01-15,支出,食費,1500,ランチ,2024-01-15 12:30:00
2,2024-01-25,収入,給与,300000,1月分給与,2024-01-25 09:00:00
```

---

## デフォルトカテゴリ

新規ユーザー作成時に自動で以下のカテゴリが作成されます。

### 支出カテゴリ（expense）

| ID | カテゴリ名 |
|----|-----------|
| 1 | 食費 |
| 2 | 日用品 |
| 3 | 交通費 |
| 4 | 光熱費 |
| 5 | 通信費 |
| 6 | 住居費 |
| 7 | 医療費 |
| 8 | 衣服 |
| 9 | 娯楽 |
| 10 | 交際費 |
| 11 | 教育 |
| 12 | その他支出 |

### 収入カテゴリ（income）

| ID | カテゴリ名 |
|----|-----------|
| 13 | 給与 |
| 14 | 副業 |
| 15 | 投資 |
| 16 | その他収入 |

---

## エラーレスポンス例

### 400 Bad Request

```json
{
  "error": "Key: 'TransactionRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag"
}
```

### 404 Not Found

```json
{
  "error": "Transaction not found"
}
```

### 500 Internal Server Error

```json
{
  "error": "Failed to get user database"
}
```
