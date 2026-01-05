package models

import (
	"testing"
	"time"

)

func TestTransactionRepository_Create(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := GetTestCategory(t, user.ID, "expense")

	repo := NewTransactionRepository(user.ID)
	tx := &Transaction{
		CategoryID: category.ID,
		Amount:     1500,
		Type:       TypeExpense,
		Date:       time.Now(),
		Memo:       "テスト取引",
	}

	err := repo.Create(tx)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	if tx.ID == 0 {
		t.Error("Transaction ID should not be 0")
	}
}

func TestTransactionRepository_GetByID(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := GetTestCategory(t, user.ID, "expense")

	repo := NewTransactionRepository(user.ID)
	tx := &Transaction{
		CategoryID: category.ID,
		Amount:     1500,
		Type:       TypeExpense,
		Date:       time.Now(),
		Memo:       "テスト取引",
	}

	if err := repo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	retrieved, err := repo.GetByID(tx.ID)
	if err != nil {
		t.Fatalf("Failed to get transaction by ID: %v", err)
	}

	if retrieved.ID != tx.ID {
		t.Errorf("Expected ID %d, got %d", tx.ID, retrieved.ID)
	}

	if retrieved.Amount != tx.Amount {
		t.Errorf("Expected amount %d, got %d", tx.Amount, retrieved.Amount)
	}

	// カテゴリがプリロードされているか確認
	if retrieved.Category == nil {
		t.Error("Category should be preloaded")
	} else if retrieved.Category.ID != category.ID {
		t.Errorf("Expected category ID %d, got %d", category.ID, retrieved.Category.ID)
	}
}

func TestTransactionRepository_List(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	expenseCategory := GetTestCategory(t, user.ID, "expense")
	incomeCategory := GetTestCategory(t, user.ID, "income")

	repo := NewTransactionRepository(user.ID)

	// 複数の取引を作成
	transactions := []Transaction{
		{CategoryID: expenseCategory.ID, Amount: 1000, Type: TypeExpense, Date: time.Now(), Memo: "支出1"},
		{CategoryID: expenseCategory.ID, Amount: 2000, Type: TypeExpense, Date: time.Now(), Memo: "支出2"},
		{CategoryID: incomeCategory.ID, Amount: 5000, Type: TypeIncome, Date: time.Now(), Memo: "収入1"},
	}

	for i := range transactions {
		if err := repo.Create(&transactions[i]); err != nil {
			t.Fatalf("Failed to create transaction: %v", err)
		}
	}

	// すべての取引を取得
	allTxs, total, err := repo.List(ListParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list transactions: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	if len(allTxs) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(allTxs))
	}

	// 支出のみを取得
	expenseTxs, expenseTotal, err := repo.List(ListParams{Limit: 10, Offset: 0, Type: TypeExpense})
	if err != nil {
		t.Fatalf("Failed to list expense transactions: %v", err)
	}

	if expenseTotal != 2 {
		t.Errorf("Expected 2 expense transactions, got %d", expenseTotal)
	}

	if len(expenseTxs) != 2 {
		t.Errorf("Expected 2 expense transactions, got %d", len(expenseTxs))
	}

	// 収入のみを取得
	incomeTxs, incomeTotal, err := repo.List(ListParams{Limit: 10, Offset: 0, Type: TypeIncome})
	if err != nil {
		t.Fatalf("Failed to list income transactions: %v", err)
	}

	if incomeTotal != 1 {
		t.Errorf("Expected 1 income transaction, got %d", incomeTotal)
	}

	if len(incomeTxs) != 1 {
		t.Errorf("Expected 1 income transaction, got %d", len(incomeTxs))
	}
}

func TestTransactionRepository_List_Pagination(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := GetTestCategory(t, user.ID, "expense")

	repo := NewTransactionRepository(user.ID)

	// 5件の取引を作成
	for i := 0; i < 5; i++ {
		tx := &Transaction{
			CategoryID: category.ID,
			Amount:     1000 * (i + 1),
			Type:       TypeExpense,
			Date:       time.Now(),
			Memo:       "取引",
		}
		if err := repo.Create(tx); err != nil {
			t.Fatalf("Failed to create transaction: %v", err)
		}
	}

	// 最初の2件を取得
	txs, total, err := repo.List(ListParams{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list transactions: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs))
	}

	// 次の2件を取得
	txs2, _, err := repo.List(ListParams{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("Failed to list transactions with offset: %v", err)
	}

	if len(txs2) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs2))
	}
}

func TestTransactionRepository_List_DateRange(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := GetTestCategory(t, user.ID, "expense")

	repo := NewTransactionRepository(user.ID)

	now := time.Now()
	oldDate := now.AddDate(0, 0, -10)
	recentDate := now.AddDate(0, 0, -2)

	// 異なる日付の取引を作成
	transactions := []Transaction{
		{CategoryID: category.ID, Amount: 1000, Type: TypeExpense, Date: oldDate, Memo: "古い取引"},
		{CategoryID: category.ID, Amount: 2000, Type: TypeExpense, Date: recentDate, Memo: "最近の取引"},
		{CategoryID: category.ID, Amount: 3000, Type: TypeExpense, Date: now, Memo: "今日の取引"},
	}

	for i := range transactions {
		if err := repo.Create(&transactions[i]); err != nil {
			t.Fatalf("Failed to create transaction: %v", err)
		}
	}

	// 最近の取引のみを取得
	startDate := now.AddDate(0, 0, -5)
	txs, total, err := repo.List(ListParams{
		Limit:     10,
		Offset:    0,
		StartDate: startDate,
	})
	if err != nil {
		t.Fatalf("Failed to list transactions with date range: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 transactions in date range, got %d", total)
	}

	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs))
	}
}

func TestTransactionRepository_Update(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := GetTestCategory(t, user.ID, "expense")

	repo := NewTransactionRepository(user.ID)
	tx := &Transaction{
		CategoryID: category.ID,
		Amount:     1500,
		Type:       TypeExpense,
		Date:       time.Now(),
		Memo:       "元のメモ",
	}

	if err := repo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	tx.Amount = 2000
	tx.Memo = "更新されたメモ"

	if err := repo.Update(tx); err != nil {
		t.Fatalf("Failed to update transaction: %v", err)
	}

	retrieved, err := repo.GetByID(tx.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated transaction: %v", err)
	}

	if retrieved.Amount != 2000 {
		t.Errorf("Expected amount 2000, got %d", retrieved.Amount)
	}

	if retrieved.Memo != "更新されたメモ" {
		t.Errorf("Expected memo '更新されたメモ', got '%s'", retrieved.Memo)
	}
}

func TestTransactionRepository_Delete(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := GetTestCategory(t, user.ID, "expense")

	repo := NewTransactionRepository(user.ID)
	tx := &Transaction{
		CategoryID: category.ID,
		Amount:     1500,
		Type:       TypeExpense,
		Date:       time.Now(),
		Memo:       "削除する取引",
	}

	if err := repo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	if err := repo.Delete(tx.ID); err != nil {
		t.Fatalf("Failed to delete transaction: %v", err)
	}

	// 削除された取引を取得しようとするとエラーになることを確認
	_, err := repo.GetByID(tx.ID)
	if err == nil {
		t.Error("Expected error when getting deleted transaction, got nil")
	}
}
