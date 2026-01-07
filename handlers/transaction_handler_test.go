package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/models"
)

// TestTransactionHandler_ListTransactions は一覧取得の単体テスト
func TestTransactionHandler_ListTransactions(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		// テストデータを追加
		for i := 0; i < 3; i++ {
			mockRepo.AddTransaction(&models.Transaction{
				CategoryID: 1,
				Amount:     1000 * (i + 1),
				Type:       models.TypeExpense,
				Date:       time.Now(),
				Memo:       fmt.Sprintf("取引 %d", i+1),
			})
		}

		router := setupTestRouter()
		router.GET("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ListTransactions)

		req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)

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
	})

	t.Run("list with pagination", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		for i := 0; i < 10; i++ {
			mockRepo.AddTransaction(&models.Transaction{
				CategoryID: 1,
				Amount:     1000,
				Type:       models.TypeExpense,
				Date:       time.Now(),
			})
		}

		router := setupTestRouter()
		router.GET("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ListTransactions)

		req, _ := http.NewRequest("GET", "/api/v1/transactions?limit=5&offset=0", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		if len(data) != 5 {
			t.Errorf("Expected 5 transactions, got %d", len(data))
		}
	})

	t.Run("list error", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		mockRepo.ListErr = errors.New("database error")
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ListTransactions)

		req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestTransactionHandler_CreateTransaction は新規作成の単体テスト
func TestTransactionHandler_CreateTransaction(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      1500,
			"type":        "expense",
			"date":        time.Now().Format("2006-01-02"),
			"memo":        "テスト取引",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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

		// モックリポジトリにトランザクションが追加されたことを確認
		if len(mockRepo.Transactions) != 1 {
			t.Errorf("Expected 1 transaction in repository, got %d", len(mockRepo.Transactions))
		}
	})

	t.Run("invalid request - missing category_id", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateTransaction)

		reqBody := map[string]interface{}{
			"amount": 1500,
			"type":   "expense",
			"date":   time.Now().Format("2006-01-02"),
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid request - negative amount", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      -1500,
			"type":        "expense",
			"date":        time.Now().Format("2006-01-02"),
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid request - invalid type", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      1500,
			"type":        "invalid",
			"date":        time.Now().Format("2006-01-02"),
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid date format", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      1500,
			"type":        "expense",
			"date":        "invalid-date",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("creation error", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		mockRepo.CreateErr = errors.New("database error")
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/transactions", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      1500,
			"type":        "expense",
			"date":        time.Now().Format("2006-01-02"),
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestTransactionHandler_GetTransaction は詳細取得の単体テスト
func TestTransactionHandler_GetTransaction(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		mockRepo.AddTransaction(&models.Transaction{
			CategoryID: 1,
			Amount:     1500,
			Type:       models.TypeExpense,
			Date:       time.Now(),
			Memo:       "テスト取引",
		})

		router := setupTestRouter()
		router.GET("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetTransaction)

		req, _ := http.NewRequest("GET", "/api/v1/transactions/1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.Transaction
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != 1 {
			t.Errorf("Expected ID 1, got %d", response.ID)
		}

		if response.Amount != 1500 {
			t.Errorf("Expected amount 1500, got %d", response.Amount)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetTransaction)

		req, _ := http.NewRequest("GET", "/api/v1/transactions/9999", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetTransaction)

		req, _ := http.NewRequest("GET", "/api/v1/transactions/invalid", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestTransactionHandler_UpdateTransaction は更新の単体テスト
func TestTransactionHandler_UpdateTransaction(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		mockRepo.AddTransaction(&models.Transaction{
			CategoryID: 1,
			Amount:     1500,
			Type:       models.TypeExpense,
			Date:       time.Now(),
			Memo:       "元のメモ",
		})

		router := setupTestRouter()
		router.PUT("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      2000,
			"type":        "expense",
			"date":        time.Now().Format("2006-01-02"),
			"memo":        "更新されたメモ",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/transactions/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

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
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.PUT("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      2000,
			"type":        "expense",
			"date":        time.Now().Format("2006-01-02"),
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/transactions/9999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("update error", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		mockRepo.UpdateErr = errors.New("update error")
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		mockRepo.AddTransaction(&models.Transaction{
			CategoryID: 1,
			Amount:     1500,
			Type:       models.TypeExpense,
			Date:       time.Now(),
		})

		router := setupTestRouter()
		router.PUT("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateTransaction)

		reqBody := map[string]interface{}{
			"category_id": 1,
			"amount":      2000,
			"type":        "expense",
			"date":        time.Now().Format("2006-01-02"),
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/transactions/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestTransactionHandler_DeleteTransaction は削除の単体テスト
func TestTransactionHandler_DeleteTransaction(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		mockRepo.AddTransaction(&models.Transaction{
			CategoryID: 1,
			Amount:     1500,
			Type:       models.TypeExpense,
			Date:       time.Now(),
		})

		router := setupTestRouter()
		router.DELETE("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.DeleteTransaction)

		req, _ := http.NewRequest("DELETE", "/api/v1/transactions/1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		// 削除されたことを確認
		if len(mockRepo.Transactions) != 0 {
			t.Errorf("Expected 0 transactions, got %d", len(mockRepo.Transactions))
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.DELETE("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.DeleteTransaction)

		req, _ := http.NewRequest("DELETE", "/api/v1/transactions/invalid", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		mockRepo.DeleteErr = errors.New("delete error")
		handler := NewTransactionHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.DELETE("/api/v1/transactions/:id", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.DeleteTransaction)

		req, _ := http.NewRequest("DELETE", "/api/v1/transactions/1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}
