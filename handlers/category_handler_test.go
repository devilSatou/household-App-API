package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"household-app/models"
)

// TestCategoryHandler_ListCategories は一覧取得の単体テスト
func TestCategoryHandler_ListCategories(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockRepo := NewMockCategoryRepository()
		handler := NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
			return mockRepo
		})

		// テストデータを追加
		mockRepo.AddCategory(&models.Category{Name: "食費", Type: "expense"})
		mockRepo.AddCategory(&models.Category{Name: "給与", Type: "income"})
		mockRepo.AddCategory(&models.Category{Name: "交通費", Type: "expense"})

		router := setupTestRouter()
		router.GET("/api/v1/categories", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ListCategories)

		req, _ := http.NewRequest("GET", "/api/v1/categories", nil)

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
			t.Errorf("Expected 3 categories, got %d", len(data))
		}
	})

	t.Run("list by type", func(t *testing.T) {
		mockRepo := NewMockCategoryRepository()
		handler := NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
			return mockRepo
		})

		mockRepo.AddCategory(&models.Category{Name: "食費", Type: "expense"})
		mockRepo.AddCategory(&models.Category{Name: "給与", Type: "income"})
		mockRepo.AddCategory(&models.Category{Name: "交通費", Type: "expense"})

		router := setupTestRouter()
		router.GET("/api/v1/categories", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ListCategories)

		req, _ := http.NewRequest("GET", "/api/v1/categories?type=expense", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		if len(data) != 2 {
			t.Errorf("Expected 2 expense categories, got %d", len(data))
		}
	})

	t.Run("list error", func(t *testing.T) {
		mockRepo := NewMockCategoryRepository()
		mockRepo.ListErr = errors.New("database error")
		handler := NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.GET("/api/v1/categories", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ListCategories)

		req, _ := http.NewRequest("GET", "/api/v1/categories", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestCategoryHandler_CreateCategory は新規作成の単体テスト
func TestCategoryHandler_CreateCategory(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockRepo := NewMockCategoryRepository()
		handler := NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/categories", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateCategory)

		reqBody := CategoryRequest{
			Name: "新カテゴリ",
			Type: "expense",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response models.Category
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID == 0 {
			t.Error("Category ID should not be 0")
		}

		if response.Name != "新カテゴリ" {
			t.Errorf("Expected name '新カテゴリ', got '%s'", response.Name)
		}

		// モックリポジトリにカテゴリが追加されたことを確認
		if len(mockRepo.Categories) != 1 {
			t.Errorf("Expected 1 category in repository, got %d", len(mockRepo.Categories))
		}
	})

	t.Run("invalid request - missing name", func(t *testing.T) {
		mockRepo := NewMockCategoryRepository()
		handler := NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/categories", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateCategory)

		reqBody := map[string]interface{}{
			"type": "expense",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid request - invalid type", func(t *testing.T) {
		mockRepo := NewMockCategoryRepository()
		handler := NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/categories", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateCategory)

		reqBody := map[string]interface{}{
			"name": "テスト",
			"type": "invalid",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("creation error", func(t *testing.T) {
		mockRepo := NewMockCategoryRepository()
		mockRepo.CreateErr = errors.New("database error")
		handler := NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
			return mockRepo
		})

		router := setupTestRouter()
		router.POST("/api/v1/categories", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.CreateCategory)

		reqBody := CategoryRequest{
			Name: "新カテゴリ",
			Type: "expense",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}
