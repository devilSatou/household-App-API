package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/models"
)

// TestSummaryHandler_GetMonthlySummary は月次集計取得の単体テスト
func TestSummaryHandler_GetMonthlySummary(t *testing.T) {
	t.Run("successful get with default params", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/monthly", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetMonthlySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/monthly", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.MonthlySummary
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// デフォルトは今月
		now := time.Now()
		if response.Year != now.Year() {
			t.Errorf("Expected year %d, got %d", now.Year(), response.Year)
		}

		if response.Month != int(now.Month()) {
			t.Errorf("Expected month %d, got %d", int(now.Month()), response.Month)
		}
	})

	t.Run("successful get with custom params", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/monthly", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetMonthlySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/monthly?year=2024&month=6", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.MonthlySummary
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Year != 2024 {
			t.Errorf("Expected year 2024, got %d", response.Year)
		}

		if response.Month != 6 {
			t.Errorf("Expected month 6, got %d", response.Month)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		mockRepo.ListErr = errors.New("database error")
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			// エラーを返すカスタムモック
			return &errorMonthlySummaryRepo{}
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/monthly", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetMonthlySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/monthly", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestSummaryHandler_GetCategorySummary はカテゴリ別集計取得の単体テスト
func TestSummaryHandler_GetCategorySummary(t *testing.T) {
	t.Run("successful get with default params", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/category", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetCategorySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/category", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := response["data"]; !ok {
			t.Error("Response should contain 'data'")
		}

		if _, ok := response["start_date"]; !ok {
			t.Error("Response should contain 'start_date'")
		}

		if _, ok := response["end_date"]; !ok {
			t.Error("Response should contain 'end_date'")
		}
	})

	t.Run("successful get with custom date range", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/category", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetCategorySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/category?start_date=2024-01-01&end_date=2024-06-30", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["start_date"] != "2024-01-01" {
			t.Errorf("Expected start_date '2024-01-01', got '%s'", response["start_date"])
		}

		if response["end_date"] != "2024-06-30" {
			t.Errorf("Expected end_date '2024-06-30', got '%s'", response["end_date"])
		}
	})

	t.Run("error handling", func(t *testing.T) {
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return &errorCategorySummaryRepo{}
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/category", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetCategorySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/category", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestSummaryHandler_GetYearlySummary は年次集計取得の単体テスト
func TestSummaryHandler_GetYearlySummary(t *testing.T) {
	t.Run("successful get with default year", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/yearly", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetYearlySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/yearly", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.YearlySummary
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// デフォルトは今年
		now := time.Now()
		if response.Year != now.Year() {
			t.Errorf("Expected year %d, got %d", now.Year(), response.Year)
		}
	})

	t.Run("successful get with custom year", func(t *testing.T) {
		mockRepo := NewMockTransactionRepository()
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/yearly", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetYearlySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/yearly?year=2023", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.YearlySummary
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Year != 2023 {
			t.Errorf("Expected year 2023, got %d", response.Year)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		handler := NewSummaryHandler(func(userID uint) models.TransactionRepositoryInterface {
			return &errorYearlySummaryRepo{}
		})

		router := setupTestRouter()
		router.GET("/api/v1/summary/yearly", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetYearlySummary)

		req, _ := http.NewRequest("GET", "/api/v1/summary/yearly", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// エラーを返すモックリポジトリ
type errorMonthlySummaryRepo struct {
	MockTransactionRepository
}

func (r *errorMonthlySummaryRepo) GetMonthlySummary(year, month int) (*models.MonthlySummary, error) {
	return nil, errors.New("database error")
}

type errorCategorySummaryRepo struct {
	MockTransactionRepository
}

func (r *errorCategorySummaryRepo) GetCategorySummary(startDate, endDate time.Time) ([]models.CategorySummary, error) {
	return nil, errors.New("database error")
}

type errorYearlySummaryRepo struct {
	MockTransactionRepository
}

func (r *errorYearlySummaryRepo) GetYearlySummary(year int) (*models.YearlySummary, error) {
	return nil, errors.New("database error")
}
