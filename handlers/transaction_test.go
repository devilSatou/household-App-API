package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"household-app/middleware"
	"household-app/models"
)

func TestListTransactions(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := models.GetTestCategory(t, user.ID, "expense")

	// テストトランザクションを作成
	repo := models.NewTransactionRepository(user.ID)
	for i := 0; i < 3; i++ {
		tx := &models.Transaction{
			CategoryID: category.ID,
			Amount:     1000 * (i + 1),
			Type:       models.TypeExpense,
			Date:       time.Now(),
			Memo:       fmt.Sprintf("取引 %d", i+1),
		}
		if err := repo.Create(tx); err != nil {
			t.Fatalf("Failed to create transaction: %v", err)
		}
	}

	router := setupTestRouter()
	router.GET("/api/v1/transactions", middleware.UserContext(), ListTransactions)

	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Response should contain 'data' array")
	}

	if len(data) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(data))
	}

	total, ok := response["total"].(float64)
	if !ok {
		t.Fatal("Response should contain 'total'")
	}

	if int(total) != 3 {
		t.Errorf("Expected total 3, got %d", int(total))
	}
}

func TestCreateTransaction(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := models.GetTestCategory(t, user.ID, "expense")

	router := setupTestRouter()
	router.POST("/api/v1/transactions", middleware.UserContext(), CreateTransaction)

	reqBody := map[string]interface{}{
		"category_id": category.ID,
		"amount":      1500,
		"type":        "expense",
		"date":        time.Now().Format("2006-01-02"),
		"memo":        "テスト取引",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID == 0 {
		t.Error("Transaction ID should not be 0")
	}

	if response.Amount != 1500 {
		t.Errorf("Expected amount 1500, got %d", response.Amount)
	}

	if response.Memo != "テスト取引" {
		t.Errorf("Expected memo 'テスト取引', got '%s'", response.Memo)
	}
}

func TestCreateTransaction_InvalidRequest(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")

	router := setupTestRouter()
	router.POST("/api/v1/transactions", middleware.UserContext(), CreateTransaction)

	tests := []struct {
		name    string
		reqBody map[string]interface{}
	}{
		{
			name: "missing category_id",
			reqBody: map[string]interface{}{
				"amount": 1500,
				"type":   "expense",
				"date":   time.Now().Format("2006-01-02"),
			},
		},
		{
			name: "missing amount",
			reqBody: map[string]interface{}{
				"category_id": 1,
				"type":        "expense",
				"date":        time.Now().Format("2006-01-02"),
			},
		},
		{
			name: "invalid type",
			reqBody: map[string]interface{}{
				"category_id": 1,
				"amount":      1500,
				"type":        "invalid",
				"date":        time.Now().Format("2006-01-02"),
			},
		},
		{
			name: "negative amount",
			reqBody: map[string]interface{}{
				"category_id": 1,
				"amount":      -1500,
				"type":        "expense",
				"date":        time.Now().Format("2006-01-02"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestGetTransaction(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := models.GetTestCategory(t, user.ID, "expense")

	repo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: category.ID,
		Amount:     1500,
		Type:       models.TypeExpense,
		Date:       time.Now(),
		Memo:       "テスト取引",
	}

	if err := repo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	router := setupTestRouter()
	router.GET("/api/v1/transactions/:id", middleware.UserContext(), GetTransaction)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/transactions/%d", tx.ID), nil)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != tx.ID {
		t.Errorf("Expected ID %d, got %d", tx.ID, response.ID)
	}

	if response.Amount != tx.Amount {
		t.Errorf("Expected amount %d, got %d", tx.Amount, response.Amount)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")

	router := setupTestRouter()
	router.GET("/api/v1/transactions/:id", middleware.UserContext(), GetTransaction)

	req, _ := http.NewRequest("GET", "/api/v1/transactions/9999", nil)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateTransaction(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := models.GetTestCategory(t, user.ID, "expense")

	repo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: category.ID,
		Amount:     1500,
		Type:       models.TypeExpense,
		Date:       time.Now(),
		Memo:       "元のメモ",
	}

	if err := repo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	router := setupTestRouter()
	router.PUT("/api/v1/transactions/:id", middleware.UserContext(), UpdateTransaction)

	reqBody := map[string]interface{}{
		"category_id": category.ID,
		"amount":      2000,
		"type":        "expense",
		"date":        time.Now().Format("2006-01-02"),
		"memo":        "更新されたメモ",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/transactions/%d", tx.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Amount != 2000 {
		t.Errorf("Expected amount 2000, got %d", response.Amount)
	}

	if response.Memo != "更新されたメモ" {
		t.Errorf("Expected memo '更新されたメモ', got '%s'", response.Memo)
	}
}

func TestDeleteTransaction(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Test User", "test@example.com")
	category := models.GetTestCategory(t, user.ID, "expense")

	repo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: category.ID,
		Amount:     1500,
		Type:       models.TypeExpense,
		Date:       time.Now(),
		Memo:       "削除する取引",
	}

	if err := repo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	router := setupTestRouter()
	router.DELETE("/api/v1/transactions/:id", middleware.UserContext(), DeleteTransaction)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/transactions/%d", tx.ID), nil)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// 削除されたことを確認
	_, err := repo.GetByID(tx.ID)
	if err == nil {
		t.Error("Transaction should be deleted")
	}
}
