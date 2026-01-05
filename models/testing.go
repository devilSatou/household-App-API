package models

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// SetupTestDB はテスト用のデータベースをセットアップする
func SetupTestDB(t *testing.T) {
	t.Helper()

	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 環境変数を設定
	os.Setenv("DB_PATH", filepath.Join(tmpDir, "test_app.db"))
	os.Setenv("USER_DB_BASE_PATH", filepath.Join(tmpDir, "users"))

	// 共通DBを初期化
	_, err := GetAppDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

// TeardownTestDB はテスト後のクリーンアップを行う
func TeardownTestDB(t *testing.T) {
	t.Helper()
	CloseAllDBs()

	// appDBOnceをリセット（次のテストで再初期化できるようにする）
	appDBOnce = sync.Once{}
	appDB = nil
}

// CreateTestUser はテスト用のユーザーを作成する
func CreateTestUser(t *testing.T, name, email, password string) *User {
	t.Helper()

	user, err := CreateUserWithPassword(name, email, password)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestUserWithCategories はテスト用のユーザーとデフォルトカテゴリを作成する
func CreateTestUserWithCategories(t *testing.T, name, email string) *User {
	t.Helper()

	user := CreateTestUser(t, name, email, "password123")

	// デフォルトカテゴリを作成
	if err := SeedDefaultCategories(user.ID); err != nil {
		t.Fatalf("Failed to seed default categories: %v", err)
	}

	return user
}

// GetTestCategory はテスト用のカテゴリを取得する
func GetTestCategory(t *testing.T, userID uint, categoryType string) *Category {
	t.Helper()

	repo := NewCategoryRepository(userID)
	categories, err := repo.List(categoryType)
	if err != nil {
		t.Fatalf("Failed to get categories: %v", err)
	}

	if len(categories) == 0 {
		t.Fatal("No categories found")
	}

	return &categories[0]
}
