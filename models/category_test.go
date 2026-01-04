package models

import (
	"testing"

)

func TestCategoryRepository_Create(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")
	repo := NewCategoryRepository(user.ID)

	category := &Category{
		Name: "テストカテゴリ",
		Type: "expense",
	}

	err := repo.Create(category)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	if category.ID == 0 {
		t.Error("Category ID should not be 0")
	}
}

func TestCategoryRepository_GetByID(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")
	repo := NewCategoryRepository(user.ID)

	category := &Category{
		Name: "テストカテゴリ",
		Type: "expense",
	}

	if err := repo.Create(category); err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	retrieved, err := repo.GetByID(category.ID)
	if err != nil {
		t.Fatalf("Failed to get category by ID: %v", err)
	}

	if retrieved.ID != category.ID {
		t.Errorf("Expected ID %d, got %d", category.ID, retrieved.ID)
	}

	if retrieved.Name != category.Name {
		t.Errorf("Expected name %s, got %s", category.Name, retrieved.Name)
	}
}

func TestCategoryRepository_List(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	repo := NewCategoryRepository(user.ID)

	// すべてのカテゴリを取得
	allCategories, err := repo.List("")
	if err != nil {
		t.Fatalf("Failed to list all categories: %v", err)
	}

	if len(allCategories) != len(DefaultCategories) {
		t.Errorf("Expected %d categories, got %d", len(DefaultCategories), len(allCategories))
	}

	// 支出カテゴリのみを取得
	expenseCategories, err := repo.List("expense")
	if err != nil {
		t.Fatalf("Failed to list expense categories: %v", err)
	}

	for _, cat := range expenseCategories {
		if cat.Type != "expense" {
			t.Errorf("Expected type 'expense', got '%s'", cat.Type)
		}
	}

	// 収入カテゴリのみを取得
	incomeCategories, err := repo.List("income")
	if err != nil {
		t.Fatalf("Failed to list income categories: %v", err)
	}

	for _, cat := range incomeCategories {
		if cat.Type != "income" {
			t.Errorf("Expected type 'income', got '%s'", cat.Type)
		}
	}

	// 合計が一致することを確認
	if len(expenseCategories)+len(incomeCategories) != len(allCategories) {
		t.Error("Sum of expense and income categories should equal total categories")
	}
}

func TestCategoryRepository_GetByName(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	repo := NewCategoryRepository(user.ID)

	category, err := repo.GetByName("食費")
	if err != nil {
		t.Fatalf("Failed to get category by name: %v", err)
	}

	if category.Name != "食費" {
		t.Errorf("Expected name '食費', got '%s'", category.Name)
	}

	if category.Type != "expense" {
		t.Errorf("Expected type 'expense', got '%s'", category.Type)
	}
}

func TestCategoryRepository_Update(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")
	repo := NewCategoryRepository(user.ID)

	category := &Category{
		Name: "テストカテゴリ",
		Type: "expense",
	}

	if err := repo.Create(category); err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	category.Name = "更新されたカテゴリ"
	if err := repo.Update(category); err != nil {
		t.Fatalf("Failed to update category: %v", err)
	}

	retrieved, err := repo.GetByID(category.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated category: %v", err)
	}

	if retrieved.Name != "更新されたカテゴリ" {
		t.Errorf("Expected name '更新されたカテゴリ', got '%s'", retrieved.Name)
	}
}

func TestCategoryRepository_Delete(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")
	repo := NewCategoryRepository(user.ID)

	category := &Category{
		Name: "削除するカテゴリ",
		Type: "expense",
	}

	if err := repo.Create(category); err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	if err := repo.Delete(category.ID); err != nil {
		t.Fatalf("Failed to delete category: %v", err)
	}

	// 削除されたカテゴリを取得しようとするとエラーになることを確認
	_, err := repo.GetByID(category.ID)
	if err == nil {
		t.Error("Expected error when getting deleted category, got nil")
	}
}

func TestSeedDefaultCategories(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")

	// デフォルトカテゴリをシード
	if err := SeedDefaultCategories(user.ID); err != nil {
		t.Fatalf("Failed to seed default categories: %v", err)
	}

	repo := NewCategoryRepository(user.ID)
	categories, err := repo.List("")
	if err != nil {
		t.Fatalf("Failed to list categories: %v", err)
	}

	if len(categories) != len(DefaultCategories) {
		t.Errorf("Expected %d categories, got %d", len(DefaultCategories), len(categories))
	}

	// 2回目の実行では何もしないことを確認
	if err := SeedDefaultCategories(user.ID); err != nil {
		t.Fatalf("Failed on second seed: %v", err)
	}

	categories2, err := repo.List("")
	if err != nil {
		t.Fatalf("Failed to list categories after second seed: %v", err)
	}

	if len(categories2) != len(categories) {
		t.Error("Second seed should not create duplicate categories")
	}
}
