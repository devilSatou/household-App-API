# テストとシードデータ

## テストの実行

### 全てのテストを実行

```bash
go test ./...
```

### 特定のパッケージのテストを実行

```bash
# modelsのテスト
go test ./models -v

# middlewareのテスト
go test ./middleware -v

# handlersのテスト
go test ./handlers -v
```

### 特定のテストを実行

```bash
go test ./models -run TestCreateUser -v
```

## テストの構成

### Models層のテスト

- `models/user_test.go` - ユーザー関連のテスト
- `models/category_test.go` - カテゴリ関連のテスト
- `models/transaction_test.go` - トランザクション関連のテスト

**テストカバレッジ:**
- ユーザーの作成、取得、更新、削除
- パスワードのハッシュ化と検証
- カテゴリのCRUD操作
- トランザクションのCRUD操作
- ページネーション
- 日付範囲フィルタリング

### Middleware層のテスト

- `middleware/auth_test.go` - JWT認証のテスト

**テストカバレッジ:**
- JWTトークンの生成
- JWTトークンの検証
- 無効なトークンのハンドリング
- トークンの有効期限

### Handlers層のテスト

- `handlers/auth_test.go` - 認証ハンドラーのテスト
- `handlers/transaction_test.go` - トランザクションハンドラーのテスト

**テストカバレッジ:**
- ユーザー登録
- ログイン
- プロファイル取得・更新
- パスワード変更
- トランザクションの作成、取得、更新、削除
- バリデーションエラーのハンドリング

## シードデータ

### シードデータの作成

開発用のテストデータを作成するには、以下のコマンドを実行します：

```bash
go run cmd/seed/main.go
```

これにより、以下が作成されます：
- テストユーザー（email: test@example.com, password: password123）
- デフォルトカテゴリ（食費、日用品、交通費など）
- サンプルトランザクションデータ

### デフォルトカテゴリ

新規ユーザー作成時に自動的に以下のカテゴリが作成されます：

**支出カテゴリ:**
- 食費
- 日用品
- 交通費
- 光熱費
- 通信費
- 住居費
- 医療費
- 衣服
- 娯楽
- 交際費
- 教育
- その他支出

**収入カテゴリ:**
- 給与
- 副業
- 投資
- その他収入

### プログラムからシードデータを作成

```go
import "household-app/models"

// テストユーザーとデフォルトカテゴリを作成
user, err := models.SeedTestUser()
if err != nil {
    log.Fatal(err)
}

// サンプルトランザクションを作成
err = models.SeedTestTransactions(user.ID)
if err != nil {
    log.Fatal(err)
}

// すべてのシードデータを一度に作成
err = models.SeedAll()
if err != nil {
    log.Fatal(err)
}
```

## テストヘルパー

テストで使用できるヘルパー関数は `models/testing.go` に定義されています：

```go
// テスト用DBのセットアップ
models.SetupTestDB(t)
defer models.TeardownTestDB(t)

// テストユーザーの作成
user := models.CreateTestUser(t, "Test User", "test@example.com", "password123")

// テストユーザー + デフォルトカテゴリの作成
user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")

// テストカテゴリの取得
category := models.GetTestCategory(t, user.ID, "expense")
```

## CI/CDでのテスト実行

GitHub Actionsなどで自動テストを実行する場合：

```yaml
- name: Run tests
  run: go test ./... -v -cover
```

## カバレッジレポート

テストカバレッジを確認するには：

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
