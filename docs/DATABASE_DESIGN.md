# データベース設計書

## 概要

本システムではSQLiteを使用し、**スキーマレベルでのユーザー分離**を採用しています。
これにより、各ユーザーのデータは物理的に別ファイルとして管理され、データ漏洩リスクを最小化しています。

## アーキテクチャ

### DB分離方式

```
data/
├── app.db                    # 共通DB（ユーザー管理のみ）
└── users/
    ├── 1/
    │   └── household.db      # ユーザー1の家計簿データ
    ├── 2/
    │   └── household.db      # ユーザー2の家計簿データ
    └── .../
```

### 設計理由

| 観点 | 説明 |
|------|------|
| **セキュリティ** | WHERE句でのフィルタ漏れによるデータ混在リスクを排除 |
| **プライバシー** | ポートフォリオ公開時に個人データを完全分離可能 |
| **バックアップ** | ユーザー単位でのバックアップ・リストアが容易 |
| **削除** | ユーザー削除 = フォルダ削除でシンプル |
| **パフォーマンス** | 各DBが小さくなり、クエリが高速 |

---

## テーブル定義

### 共通DB（app.db）

#### users テーブル

ユーザー情報を管理する。

| カラム名 | 型 | 制約 | 説明 |
|----------|------|------|------|
| id | INTEGER | PRIMARY KEY, AUTO INCREMENT | ユーザーID |
| name | TEXT | NOT NULL | ユーザー名 |
| email | TEXT | UNIQUE | メールアドレス（任意） |
| created_at | DATETIME | | 作成日時 |
| updated_at | DATETIME | | 更新日時 |

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

### ユーザー専用DB（users/{id}/household.db）

#### transactions テーブル

収支記録を管理する。**user_id は不要**（DBファイル自体がユーザーに紐づく）。

| カラム名 | 型 | 制約 | 説明 |
|----------|------|------|------|
| id | INTEGER | PRIMARY KEY, AUTO INCREMENT | 取引ID |
| category_id | INTEGER | INDEX, FOREIGN KEY | カテゴリID |
| amount | INTEGER | NOT NULL | 金額（正の整数） |
| type | TEXT | NOT NULL | 種別（income / expense） |
| date | DATE | NOT NULL, INDEX | 取引日 |
| memo | TEXT | | メモ |
| created_at | DATETIME | | 作成日時 |
| updated_at | DATETIME | | 更新日時 |

```sql
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER,
    amount INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    date DATE NOT NULL,
    memo TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
```

#### categories テーブル

カテゴリマスタを管理する。ユーザーごとにカスタマイズ可能。

| カラム名 | 型 | 制約 | 説明 |
|----------|------|------|------|
| id | INTEGER | PRIMARY KEY, AUTO INCREMENT | カテゴリID |
| name | TEXT | UNIQUE, NOT NULL | カテゴリ名 |
| type | TEXT | NOT NULL | 種別（income / expense） |
| created_at | DATETIME | | 作成日時 |

```sql
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 初期データ（シード）

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

## ER図

```
┌─────────────────────────────────────────────────────────────┐
│                      共通DB (app.db)                         │
│  ┌─────────────┐                                            │
│  │   users     │                                            │
│  ├─────────────┤                                            │
│  │ id (PK)     │──────┐                                     │
│  │ name        │      │                                     │
│  │ email       │      │                                     │
│  │ created_at  │      │                                     │
│  │ updated_at  │      │                                     │
│  └─────────────┘      │                                     │
└───────────────────────┼─────────────────────────────────────┘
                        │
                        │ ユーザーIDでDBファイルを特定
                        ▼
┌─────────────────────────────────────────────────────────────┐
│            ユーザー専用DB (users/{id}/household.db)          │
│                                                             │
│  ┌─────────────────┐       ┌─────────────┐                 │
│  │  transactions   │       │ categories  │                 │
│  ├─────────────────┤       ├─────────────┤                 │
│  │ id (PK)         │       │ id (PK)     │                 │
│  │ category_id (FK)│──────►│ name        │                 │
│  │ amount          │       │ type        │                 │
│  │ type            │       │ created_at  │                 │
│  │ date            │       └─────────────┘                 │
│  │ memo            │                                        │
│  │ created_at      │                                        │
│  │ updated_at      │                                        │
│  └─────────────────┘                                        │
└─────────────────────────────────────────────────────────────┘
```

---

## DB接続フロー

```
1. リクエスト受信
       │
       ▼
2. 認証/ユーザー特定
   └─► userID を取得
       │
       ▼
3. GetUserDB(userID) 呼び出し
   ├─► キャッシュにあれば返却
   └─► なければ新規接続
       ├─► ディレクトリ作成 (data/users/{id}/)
       ├─► DB接続 (household.db)
       ├─► マイグレーション実行
       ├─► 初期カテゴリをシード
       └─► キャッシュに保存
           │
           ▼
4. Repository経由でCRUD操作
       │
       ▼
5. レスポンス返却
```

---

## 将来の拡張

### 認証機能追加時

`users` テーブルに以下のカラムを追加予定：

```sql
ALTER TABLE users ADD COLUMN password_hash TEXT;
ALTER TABLE users ADD COLUMN last_login_at DATETIME;
```

### 共有機能追加時（家族での共有など）

新規テーブルを追加：

```sql
CREATE TABLE user_shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    shared_user_id INTEGER NOT NULL,
    permission TEXT NOT NULL CHECK (permission IN ('read', 'write')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id),
    FOREIGN KEY (shared_user_id) REFERENCES users(id)
);
```

---

## 注意事項

1. **バックアップ**: `data/` ディレクトリ全体を定期的にバックアップすること
2. **削除**: ユーザー削除時は `data/users/{id}/` フォルダを物理削除すること
3. **同時接続**: SQLiteは書き込み時にロックがかかるため、高負荷環境では注意
4. **マイグレーション**: スキーマ変更時は全ユーザーのDBに対して実行が必要
