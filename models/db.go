package models

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	appDB     *gorm.DB
	appDBOnce sync.Once

	// ユーザーDBのキャッシュ
	userDBCache = make(map[uint]*gorm.DB)
	userDBMutex sync.RWMutex
)

// GetAppDB は共通DB（ユーザー管理用）の接続を返す
func GetAppDB() (*gorm.DB, error) {
	var err error
	appDBOnce.Do(func() {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "data/app.db"
		}

		// ディレクトリ作成
		dir := filepath.Dir(dbPath)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}

		appDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			return
		}

		// マイグレーション（共通テーブル）
		err = appDB.AutoMigrate(&User{})
	})

	return appDB, err
}

// GetUserDB はユーザー専用DBの接続を返す
func GetUserDB(userID uint) (*gorm.DB, error) {
	// キャッシュ確認
	userDBMutex.RLock()
	if db, exists := userDBCache[userID]; exists {
		userDBMutex.RUnlock()
		return db, nil
	}
	userDBMutex.RUnlock()

	// 新規接続
	userDBMutex.Lock()
	defer userDBMutex.Unlock()

	// ダブルチェック
	if db, exists := userDBCache[userID]; exists {
		return db, nil
	}

	basePath := os.Getenv("USER_DB_BASE_PATH")
	if basePath == "" {
		basePath = "data/users"
	}

	userDir := filepath.Join(basePath, fmt.Sprintf("%d", userID))
	dbPath := filepath.Join(userDir, "household.db")

	// ディレクトリ作成
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open user database: %w", err)
	}

	// マイグレーション（ユーザー専用テーブル）
	if err := db.AutoMigrate(&Transaction{}, &Category{}); err != nil {
		return nil, fmt.Errorf("failed to migrate user database: %w", err)
	}

	// 初期カテゴリをシード
	if err := seedDefaultCategories(db); err != nil {
		return nil, fmt.Errorf("failed to seed categories: %w", err)
	}

	userDBCache[userID] = db
	return db, nil
}

// CloseAllDBs は全てのDB接続を閉じる
func CloseAllDBs() error {
	userDBMutex.Lock()
	defer userDBMutex.Unlock()

	for _, db := range userDBCache {
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}
		sqlDB.Close()
	}
	userDBCache = make(map[uint]*gorm.DB)

	if appDB != nil {
		sqlDB, err := appDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}

	return nil
}

// seedDefaultCategories は初期カテゴリを作成する
func seedDefaultCategories(db *gorm.DB) error {
	var count int64
	db.Model(&Category{}).Count(&count)
	if count > 0 {
		return nil // 既にデータがある場合はスキップ
	}

	defaultCategories := []Category{
		// 支出カテゴリ
		{Name: "食費", Type: TypeExpense},
		{Name: "日用品", Type: TypeExpense},
		{Name: "交通費", Type: TypeExpense},
		{Name: "光熱費", Type: TypeExpense},
		{Name: "通信費", Type: TypeExpense},
		{Name: "住居費", Type: TypeExpense},
		{Name: "医療費", Type: TypeExpense},
		{Name: "衣服", Type: TypeExpense},
		{Name: "娯楽", Type: TypeExpense},
		{Name: "交際費", Type: TypeExpense},
		{Name: "教育", Type: TypeExpense},
		{Name: "その他支出", Type: TypeExpense},
		// 収入カテゴリ
		{Name: "給与", Type: TypeIncome},
		{Name: "副業", Type: TypeIncome},
		{Name: "投資", Type: TypeIncome},
		{Name: "その他収入", Type: TypeIncome},
	}

	return db.Create(&defaultCategories).Error
}
