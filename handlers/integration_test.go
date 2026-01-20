package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

// setupIntegrationRouter はDB付きの統合テスト用ルーターを設定する
func setupIntegrationRouter(t *testing.T) (*gin.Engine, *models.User, func()) {
	t.Helper()

	models.SetupTestDB(t)

	user := models.CreateTestUserWithCategories(t, "Integration Test User", "integration@example.com")

	router := gin.New()
	router.Use(gin.Recovery())

	cleanup := func() {
		models.TeardownTestDB(t)
	}

	return router, user, cleanup
}

// 旧ハンドラーの統合テスト

func TestListCategories_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ListCategories)

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

	if _, ok := response["data"]; !ok {
		t.Error("Response should contain 'data'")
	}
}

func TestListCategories_ByType_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ListCategories)

	req, _ := http.NewRequest("GET", "/api/v1/categories?type=expense", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCreateCategory_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.POST("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, CreateCategory)

	reqBody := CategoryRequest{
		Name: "新カテゴリ統合テスト",
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
}

func TestCreateCategory_InvalidRequest_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.POST("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, CreateCategory)

	// 不正なリクエスト
	reqBody := map[string]string{
		"name": "テスト",
		// type が不正
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
}

func TestExportCSV_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	// テストデータを作成
	repo := models.NewTransactionRepository(user.ID)
	catRepo := models.NewCategoryRepository(user.ID)

	categories, _ := catRepo.List("expense")
	if len(categories) > 0 {
		tx := &models.Transaction{
			CategoryID: categories[0].ID,
			Amount:     1500,
			Type:       models.TypeExpense,
			Date:       time.Now(),
			Memo:       "CSV出力テスト",
		}
		repo.Create(tx)
	}

	router.GET("/api/v1/export/csv", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ExportCSV)

	req, _ := http.NewRequest("GET", "/api/v1/export/csv", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/csv; charset=utf-8', got '%s'", contentType)
	}
}

func TestExportCSV_WithFilters_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/export/csv", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ExportCSV)

	req, _ := http.NewRequest("GET", "/api/v1/export/csv?type=expense&start_date=2024-01-01&end_date=2024-12-31", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetMonthlySummary_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/monthly", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetMonthlySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/monthly", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetMonthlySummary_WithParams_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/monthly", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetMonthlySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/monthly?year=2024&month=6", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetCategorySummary_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/category", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetCategorySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/category", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetCategorySummary_WithDateRange_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/category", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetCategorySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/category?start_date=2024-01-01&end_date=2024-12-31", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetYearlySummary_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/yearly", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetYearlySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/yearly", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetYearlySummary_WithYear_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/yearly", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetYearlySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/yearly?year=2024", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestListTransactions_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/transactions", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ListTransactions)

	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestListTransactions_WithFilters_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/transactions", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ListTransactions)

	req, _ := http.NewRequest("GET", "/api/v1/transactions?type=expense&limit=5&offset=0&start_date=2024-01-01&end_date=2024-12-31", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCreateTransaction_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	// カテゴリを取得
	catRepo := models.NewCategoryRepository(user.ID)
	categories, _ := catRepo.List("expense")
	if len(categories) == 0 {
		t.Fatal("No expense categories available")
	}

	router.POST("/api/v1/transactions", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, CreateTransaction)

	reqBody := TransactionRequest{
		CategoryID: categories[0].ID,
		Amount:     1500,
		Type:       "expense",
		Date:       time.Now().Format("2006-01-02"),
		Memo:       "統合テスト取引",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestCreateTransaction_InvalidRequest_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.POST("/api/v1/transactions", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, CreateTransaction)

	// 不正なリクエスト（amountが負）
	reqBody := map[string]interface{}{
		"category_id": 1,
		"amount":      -100,
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
}

func TestGetTransaction_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	// テストデータを作成
	catRepo := models.NewCategoryRepository(user.ID)
	categories, _ := catRepo.List("expense")
	if len(categories) == 0 {
		t.Fatal("No expense categories available")
	}

	repo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: categories[0].ID,
		Amount:     1500,
		Type:       models.TypeExpense,
		Date:       time.Now(),
		Memo:       "取得テスト",
	}
	repo.Create(tx)

	router.GET("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetTransaction)

	req, _ := http.NewRequest("GET", "/api/v1/transactions/1", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetTransaction_NotFound_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetTransaction)

	req, _ := http.NewRequest("GET", "/api/v1/transactions/99999", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetTransaction_InvalidID_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetTransaction)

	req, _ := http.NewRequest("GET", "/api/v1/transactions/invalid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateTransaction_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	// テストデータを作成
	catRepo := models.NewCategoryRepository(user.ID)
	categories, _ := catRepo.List("expense")
	if len(categories) == 0 {
		t.Fatal("No expense categories available")
	}

	repo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: categories[0].ID,
		Amount:     1500,
		Type:       models.TypeExpense,
		Date:       time.Now(),
		Memo:       "更新前",
	}
	repo.Create(tx)

	router.PUT("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, UpdateTransaction)

	reqBody := TransactionRequest{
		CategoryID: categories[0].ID,
		Amount:     2000,
		Type:       "expense",
		Date:       time.Now().Format("2006-01-02"),
		Memo:       "更新後",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/transactions/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestUpdateTransaction_NotFound_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.PUT("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, UpdateTransaction)

	reqBody := TransactionRequest{
		CategoryID: 1,
		Amount:     2000,
		Type:       "expense",
		Date:       time.Now().Format("2006-01-02"),
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/transactions/99999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteTransaction_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	// テストデータを作成
	catRepo := models.NewCategoryRepository(user.ID)
	categories, _ := catRepo.List("expense")
	if len(categories) == 0 {
		t.Fatal("No expense categories available")
	}

	repo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: categories[0].ID,
		Amount:     1500,
		Type:       models.TypeExpense,
		Date:       time.Now(),
		Memo:       "削除テスト",
	}
	repo.Create(tx)

	router.DELETE("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, DeleteTransaction)

	req, _ := http.NewRequest("DELETE", "/api/v1/transactions/1", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteTransaction_InvalidID_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.DELETE("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, DeleteTransaction)

	req, _ := http.NewRequest("DELETE", "/api/v1/transactions/invalid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// 認証ハンドラーの統合テスト

func TestRegister_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := gin.New()
	router.POST("/api/v1/auth/register", Register)

	reqBody := RegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "New User",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestRegister_DuplicateEmail_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	// 既存ユーザーを作成
	models.CreateTestUser(t, "Existing User", "existing@example.com", "password123")

	router := gin.New()
	router.POST("/api/v1/auth/register", Register)

	reqBody := RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "New User",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestLogin_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	// テストユーザーを作成
	models.CreateTestUser(t, "Test User", "login@example.com", "password123")

	router := gin.New()
	router.POST("/api/v1/auth/login", Login)

	reqBody := LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestLogin_WrongPassword_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	models.CreateTestUser(t, "Test User", "wrong@example.com", "password123")

	router := gin.New()
	router.POST("/api/v1/auth/login", Login)

	reqBody := LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetProfile_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUser(t, "Profile User", "profile@example.com", "password123")

	router := gin.New()
	router.GET("/api/v1/auth/me", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetProfile)

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetProfile_Unauthorized_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := gin.New()
	router.GET("/api/v1/auth/me", GetProfile)

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpdateProfile_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUser(t, "Update User", "update@example.com", "password123")

	router := gin.New()
	router.PUT("/api/v1/auth/me", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, UpdateProfile)

	reqBody := UpdateProfileRequest{
		Name: "Updated Name",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestChangePassword_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUser(t, "Password User", "password@example.com", "password123")

	router := gin.New()
	router.PUT("/api/v1/auth/password", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ChangePassword)

	reqBody := ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestChangePassword_WrongCurrent_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUser(t, "Password User", "wrongcurrent@example.com", "password123")

	router := gin.New()
	router.PUT("/api/v1/auth/password", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ChangePassword)

	reqBody := ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// middleware.GetUserIDのテスト
func TestMiddlewareGetUserID_Integration(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(42))

	userID := middleware.GetUserID(c)
	if userID != 42 {
		t.Errorf("Expected userID 42, got %d", userID)
	}
}

// 追加のエッジケーステスト

func TestRegister_InvalidJSON_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := gin.New()
	router.POST("/api/v1/auth/register", Register)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_InvalidJSON_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := gin.New()
	router.POST("/api/v1/auth/login", Login)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_UserNotFound_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := gin.New()
	router.POST("/api/v1/auth/login", Login)

	reqBody := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpdateProfile_InvalidJSON_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUser(t, "Test User", "invalidjson@example.com", "password123")

	router := gin.New()
	router.PUT("/api/v1/auth/me", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, UpdateProfile)

	req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateProfile_NoUserID_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := gin.New()
	router.PUT("/api/v1/auth/me", UpdateProfile)

	reqBody := UpdateProfileRequest{
		Name: "Updated Name",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestChangePassword_InvalidJSON_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	user := models.CreateTestUser(t, "Test User", "changepassjson@example.com", "password123")

	router := gin.New()
	router.PUT("/api/v1/auth/password", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ChangePassword)

	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestChangePassword_NoUserID_Integration(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := gin.New()
	router.PUT("/api/v1/auth/password", ChangePassword)

	reqBody := ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateTransaction_InvalidJSON_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.POST("/api/v1/transactions", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, CreateTransaction)

	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateTransaction_InvalidJSON_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	// トランザクションを作成
	cat := models.GetTestCategory(t, user.ID, "expense")
	txRepo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: cat.ID,
		Amount:     1000,
		Type:       "expense",
		Date:       time.Now(),
		Memo:       "Test",
	}
	if err := txRepo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	router.PUT("/api/v1/transactions/:id", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, UpdateTransaction)

	req, _ := http.NewRequest("PUT", "/api/v1/transactions/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateCategory_InvalidJSON_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.POST("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, CreateCategory)

	req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}


func TestGetMonthlySummary_InvalidParams_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/monthly", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetMonthlySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/monthly?year=invalid&month=13", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 無効なパラメータの場合は現在の年月を使用するため200
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetCategorySummary_InvalidParams_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/category", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetCategorySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/category?start_date=invalid&end_date=invalid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 無効なパラメータの場合はデフォルト期間を使用するため200
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetYearlySummary_InvalidYear_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	router.GET("/api/v1/summary/yearly", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, GetYearlySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/yearly?year=invalid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 無効なパラメータの場合は現在の年を使用するため200
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExportCSV_WithDateRange_Integration(t *testing.T) {
	router, user, cleanup := setupIntegrationRouter(t)
	defer cleanup()

	// テストデータを作成
	cat := models.GetTestCategory(t, user.ID, "expense")
	txRepo := models.NewTransactionRepository(user.ID)
	tx := &models.Transaction{
		CategoryID: cat.ID,
		Amount:     1000,
		Type:       "expense",
		Date:       time.Now(),
		Memo:       "Test",
	}
	if err := txRepo.Create(tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	router.GET("/api/v1/export/csv", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, ExportCSV)

	today := time.Now().Format("2006-01-02")
	req, _ := http.NewRequest("GET", "/api/v1/export/csv?start_date="+today+"&end_date="+today+"&type=expense", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// Default*Handler ファクトリ関数のテスト

func TestDefaultTransactionHandler(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	handler := DefaultTransactionHandler()
	if handler == nil {
		t.Fatal("DefaultTransactionHandler should return non-nil handler")
	}

	// ハンドラーが正常に動作することを確認
	user := models.CreateTestUserWithCategories(t, "Test User", "defaulttx@example.com")

	router := gin.New()
	router.GET("/api/v1/transactions", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, handler.ListTransactions)

	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDefaultCategoryHandler(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	handler := DefaultCategoryHandler()
	if handler == nil {
		t.Fatal("DefaultCategoryHandler should return non-nil handler")
	}

	// ハンドラーが正常に動作することを確認
	user := models.CreateTestUserWithCategories(t, "Test User", "defaultcat@example.com")

	router := gin.New()
	router.GET("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, handler.ListCategories)

	req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDefaultSummaryHandler(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	handler := DefaultSummaryHandler()
	if handler == nil {
		t.Fatal("DefaultSummaryHandler should return non-nil handler")
	}

	// ハンドラーが正常に動作することを確認
	user := models.CreateTestUserWithCategories(t, "Test User", "defaultsum@example.com")

	router := gin.New()
	router.GET("/api/v1/summary/monthly", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, handler.GetMonthlySummary)

	req, _ := http.NewRequest("GET", "/api/v1/summary/monthly", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDefaultExportHandler(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	handler := DefaultExportHandler()
	if handler == nil {
		t.Fatal("DefaultExportHandler should return non-nil handler")
	}

	// ハンドラーが正常に動作することを確認
	user := models.CreateTestUserWithCategories(t, "Test User", "defaultexp@example.com")

	router := gin.New()
	router.GET("/api/v1/export/csv", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, handler.ExportCSV)

	req, _ := http.NewRequest("GET", "/api/v1/export/csv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDefaultAuthHandler(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	handler := DefaultAuthHandler()
	if handler == nil {
		t.Fatal("DefaultAuthHandler should return non-nil handler")
	}

	// Registerが正常に動作することを確認
	router := gin.New()
	router.POST("/api/v1/auth/register", handler.Register)

	reqBody := RegisterRequest{
		Email:    "defaultauth@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}
