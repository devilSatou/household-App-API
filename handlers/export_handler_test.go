package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/models"
)

// TestExportHandler_ExportCSV はCSVエクスポートの単体テスト
func TestExportHandler_ExportCSV(t *testing.T) {
	t.Run("successful export empty data", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewExportHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/export/csv", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ExportCSV)

		req, _ := http.NewRequest("GET", "/api/v1/export/csv", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		// Content-Type の確認
		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/csv") {
			t.Errorf("Expected Content-Type to contain 'text/csv', got '%s'", contentType)
		}

		// Content-Disposition の確認
		disposition := w.Header().Get("Content-Disposition")
		if !strings.Contains(disposition, "attachment") {
			t.Errorf("Expected Content-Disposition to contain 'attachment', got '%s'", disposition)
		}

		// CSVヘッダーの確認（BOMを除去して確認）
		body := strings.TrimPrefix(w.Body.String(), "\xef\xbb\xbf")
		if !strings.Contains(body, "ID,日付,種別,カテゴリ,金額,メモ,作成日時") {
			t.Error("CSV should contain header row")
		}
	})

	t.Run("successful export with data", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewExportHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		// テストデータを追加
		category := &models.Category{ID: 1, Name: "食費", Type: "expense"}
		mockRepo.AddTransaction(&models.Transaction{
			CategoryID: 1,
			Amount:     1500,
			Type:       models.TypeExpense,
			Date:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
			Memo:       "ランチ",
			Category:   category,
		})

		router := setupTestRouter()
		router.GET("/api/v1/export/csv", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ExportCSV)

		req, _ := http.NewRequest("GET", "/api/v1/export/csv", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()

		// データ行の確認
		if !strings.Contains(body, "支出") {
			t.Error("CSV should contain expense type label")
		}

		if !strings.Contains(body, "食費") {
			t.Error("CSV should contain category name")
		}

		if !strings.Contains(body, "1500") {
			t.Error("CSV should contain amount")
		}

		if !strings.Contains(body, "ランチ") {
			t.Error("CSV should contain memo")
		}
	})

	t.Run("export with income type", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewExportHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		mockRepo.AddTransaction(&models.Transaction{
			CategoryID: 1,
			Amount:     300000,
			Type:       models.TypeIncome,
			Date:       time.Now(),
			Memo:       "給与",
		})

		router := setupTestRouter()
		router.GET("/api/v1/export/csv", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ExportCSV)

		req, _ := http.NewRequest("GET", "/api/v1/export/csv", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()

		if !strings.Contains(body, "収入") {
			t.Error("CSV should contain income type label")
		}
	})

	t.Run("export with date filter", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewExportHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/export/csv", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ExportCSV)

		req, _ := http.NewRequest("GET", "/api/v1/export/csv?start_date=2024-01-01&end_date=2024-12-31", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("export with type filter", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewExportHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/export/csv", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ExportCSV)

		req, _ := http.NewRequest("GET", "/api/v1/export/csv?type=expense", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("export error", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		mockRepo.ListErr = errors.New("database error")
		handler := NewExportHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/export/csv", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ExportCSV)

		req, _ := http.NewRequest("GET", "/api/v1/export/csv", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}
