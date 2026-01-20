package models

import (
	"testing"
)

func TestSeedDefaultCategories(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	// ユーザーを作成
	user, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デフォルトカテゴリをシード
	if err := SeedDefaultCategories(user.ID); err != nil {
		t.Fatalf("Failed to seed default categories: %v", err)
	}

	// カテゴリが作成されたことを確認
	repo := NewCategoryRepository(user.ID)
	categories, err := repo.List("")
	if err != nil {
		t.Fatalf("Failed to list categories: %v", err)
	}

	if len(categories) != len(DefaultCategories) {
		t.Errorf("Expected %d categories, got %d", len(DefaultCategories), len(categories))
	}

	// 支出カテゴリを確認
	expenseCategories, err := repo.List("expense")
	if err != nil {
		t.Fatalf("Failed to list expense categories: %v", err)
	}

	expectedExpenseCount := 0
	for _, c := range DefaultCategories {
		if c.Type == "expense" {
			expectedExpenseCount++
		}
	}

	if len(expenseCategories) != expectedExpenseCount {
		t.Errorf("Expected %d expense categories, got %d", expectedExpenseCount, len(expenseCategories))
	}

	// 収入カテゴリを確認
	incomeCategories, err := repo.List("income")
	if err != nil {
		t.Fatalf("Failed to list income categories: %v", err)
	}

	expectedIncomeCount := 0
	for _, c := range DefaultCategories {
		if c.Type == "income" {
			expectedIncomeCount++
		}
	}

	if len(incomeCategories) != expectedIncomeCount {
		t.Errorf("Expected %d income categories, got %d", expectedIncomeCount, len(incomeCategories))
	}
}

func TestSeedDefaultCategories_SkipIfExists(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	// ユーザーを作成
	user, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 最初のシード
	if err := SeedDefaultCategories(user.ID); err != nil {
		t.Fatalf("Failed to seed default categories: %v", err)
	}

	// 2回目のシード（スキップされるはず）
	if err := SeedDefaultCategories(user.ID); err != nil {
		t.Fatalf("Failed on second seed: %v", err)
	}

	// カテゴリ数が変わっていないことを確認
	repo := NewCategoryRepository(user.ID)
	categories, err := repo.List("")
	if err != nil {
		t.Fatalf("Failed to list categories: %v", err)
	}

	if len(categories) != len(DefaultCategories) {
		t.Errorf("Expected %d categories after double seed, got %d", len(DefaultCategories), len(categories))
	}
}

func TestSeedTestUser(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user, err := SeedTestUser()
	if err != nil {
		t.Fatalf("Failed to seed test user: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	// パスワードが設定されていることを確認
	if !user.CheckPassword("password123") {
		t.Error("Password verification failed")
	}

	// カテゴリが作成されていることを確認
	repo := NewCategoryRepository(user.ID)
	categories, err := repo.List("")
	if err != nil {
		t.Fatalf("Failed to list categories: %v", err)
	}

	if len(categories) == 0 {
		t.Error("Expected categories to be created")
	}
}

func TestSeedTestUser_DeletesExisting(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	// 最初のユーザー作成
	user1, err := SeedTestUser()
	if err != nil {
		t.Fatalf("Failed to seed first test user: %v", err)
	}

	firstID := user1.ID

	// 2回目の作成（既存ユーザーが削除されるはず）
	user2, err := SeedTestUser()
	if err != nil {
		t.Fatalf("Failed to seed second test user: %v", err)
	}

	// 新しいIDが割り当てられている
	if user2.ID == firstID {
		t.Error("Expected new user ID after re-seeding")
	}

	// 古いユーザーは存在しない
	_, err = GetUserByID(firstID)
	if err == nil {
		t.Error("Old user should have been deleted")
	}
}

func TestSeedTestTransactions(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	// テストユーザーを作成
	user, err := SeedTestUser()
	if err != nil {
		t.Fatalf("Failed to seed test user: %v", err)
	}

	// テストトランザクションをシード
	if err := SeedTestTransactions(user.ID); err != nil {
		t.Fatalf("Failed to seed test transactions: %v", err)
	}

	// トランザクションが作成されていることを確認
	repo := NewTransactionRepository(user.ID)
	transactions, total, err := repo.List(ListParams{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list transactions: %v", err)
	}

	if total == 0 {
		t.Error("Expected transactions to be created")
	}

	if len(transactions) == 0 {
		t.Error("Expected at least one transaction")
	}

	// 収入と支出の両方が含まれていることを確認
	hasIncome := false
	hasExpense := false
	for _, tx := range transactions {
		if tx.Type == TypeIncome {
			hasIncome = true
		}
		if tx.Type == TypeExpense {
			hasExpense = true
		}
	}

	if !hasIncome {
		t.Error("Expected at least one income transaction")
	}

	if !hasExpense {
		t.Error("Expected at least one expense transaction")
	}
}

func TestSeedTestTransactions_WithDefaultCategories(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	// ユーザーを作成（GetUserDBがデフォルトカテゴリを自動シードする）
	user, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デフォルトカテゴリが自動的に作成されるため、トランザクションのシードは成功する
	err = SeedTestTransactions(user.ID)
	if err != nil {
		t.Fatalf("Expected no error with auto-seeded categories: %v", err)
	}

	// トランザクションが作成されたことを確認
	repo := NewTransactionRepository(user.ID)
	_, total, err := repo.List(ListParams{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list transactions: %v", err)
	}

	if total == 0 {
		t.Error("Expected transactions to be created")
	}
}

func TestSeedAll(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	if err := SeedAll(); err != nil {
		t.Fatalf("Failed to seed all: %v", err)
	}

	// テストユーザーが存在することを確認
	user, err := GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Test user should exist: %v", err)
	}

	// カテゴリが存在することを確認
	catRepo := NewCategoryRepository(user.ID)
	categories, err := catRepo.List("")
	if err != nil {
		t.Fatalf("Failed to list categories: %v", err)
	}

	if len(categories) == 0 {
		t.Error("Categories should be created")
	}

	// トランザクションが存在することを確認
	txRepo := NewTransactionRepository(user.ID)
	_, total, err := txRepo.List(ListParams{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list transactions: %v", err)
	}

	if total == 0 {
		t.Error("Transactions should be created")
	}
}
