package models

import (
	"fmt"
	"time"
)

// DefaultCategories はデフォルトカテゴリのリスト
var DefaultCategories = []struct {
	Name string
	Type string
}{
	// 支出カテゴリ
	{"食費", "expense"},
	{"日用品", "expense"},
	{"交通費", "expense"},
	{"光熱費", "expense"},
	{"通信費", "expense"},
	{"住居費", "expense"},
	{"医療費", "expense"},
	{"衣服", "expense"},
	{"娯楽", "expense"},
	{"交際費", "expense"},
	{"教育", "expense"},
	{"その他支出", "expense"},
	// 収入カテゴリ
	{"給与", "income"},
	{"副業", "income"},
	{"投資", "income"},
	{"その他収入", "income"},
}

// SeedDefaultCategories はユーザー専用DBにデフォルトカテゴリを作成する
func SeedDefaultCategories(userID uint) error {
	db, err := GetUserDB(userID)
	if err != nil {
		return err
	}

	// 既存のカテゴリ数を確認
	var count int64
	if err := db.Model(&Category{}).Count(&count).Error; err != nil {
		return err
	}

	// すでにカテゴリが存在する場合はスキップ
	if count > 0 {
		return nil
	}

	// デフォルトカテゴリを作成
	for _, cat := range DefaultCategories {
		category := Category{
			Name: cat.Name,
			Type: cat.Type,
		}
		if err := db.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to create category %s: %w", cat.Name, err)
		}
	}

	return nil
}

// SeedTestUser はテスト用ユーザーを作成する
func SeedTestUser() (*User, error) {
	email := "test@example.com"

	// 既存ユーザーがいれば削除
	if user, err := GetUserByEmail(email); err == nil {
		if err := user.Delete(); err != nil {
			return nil, err
		}
	}

	// テストユーザーを作成
	user, err := CreateUserWithPassword("テストユーザー", email, "password123")
	if err != nil {
		return nil, err
	}

	// デフォルトカテゴリを作成
	if err := SeedDefaultCategories(user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// SeedTestTransactions はテスト用の取引データを作成する
func SeedTestTransactions(userID uint) error {
	repo := NewTransactionRepository(userID)
	catRepo := NewCategoryRepository(userID)

	// カテゴリを取得
	categories, err := catRepo.List("")
	if err != nil {
		return err
	}

	if len(categories) == 0 {
		return fmt.Errorf("no categories found")
	}

	// カテゴリマップを作成
	catMap := make(map[string]*Category)
	for i := range categories {
		catMap[categories[i].Name] = &categories[i]
	}

	// サンプルトランザクションデータ
	now := time.Now()
	testData := []struct {
		CategoryName string
		Amount       int
		Type         string
		Date         time.Time
		Memo         string
	}{
		{"給与", 300000, "income", now.AddDate(0, 0, -25), "1月分給与"},
		{"食費", 1500, "expense", now.AddDate(0, 0, -20), "ランチ"},
		{"食費", 3000, "expense", now.AddDate(0, 0, -18), "ディナー"},
		{"交通費", 500, "expense", now.AddDate(0, 0, -15), "電車代"},
		{"光熱費", 8000, "expense", now.AddDate(0, 0, -10), "電気代"},
		{"通信費", 5000, "expense", now.AddDate(0, 0, -8), "携帯代"},
		{"娯楽", 2000, "expense", now.AddDate(0, 0, -5), "映画"},
		{"食費", 800, "expense", now.AddDate(0, 0, -3), "コーヒー"},
		{"副業", 50000, "income", now.AddDate(0, 0, -2), "フリーランス収入"},
		{"交際費", 4000, "expense", now.AddDate(0, 0, -1), "飲み会"},
	}

	// トランザクションを作成
	for _, data := range testData {
		cat, ok := catMap[data.CategoryName]
		if !ok {
			continue
		}

		tx := &Transaction{
			CategoryID: cat.ID,
			Amount:     data.Amount,
			Type:       data.Type,
			Date:       data.Date,
			Memo:       data.Memo,
		}

		if err := repo.Create(tx); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	return nil
}

// SeedAll はすべてのシードデータを作成する
func SeedAll() error {
	// テストユーザーを作成
	user, err := SeedTestUser()
	if err != nil {
		return fmt.Errorf("failed to seed test user: %w", err)
	}

	// テストトランザクションを作成
	if err := SeedTestTransactions(user.ID); err != nil {
		return fmt.Errorf("failed to seed test transactions: %w", err)
	}

	fmt.Printf("Seed completed successfully!\n")
	fmt.Printf("Test user: %s (ID: %d)\n", user.Email, user.ID)
	fmt.Printf("Password: password123\n")

	return nil
}
