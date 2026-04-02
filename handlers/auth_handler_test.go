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

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return r
}

// TestAuthHandler_Register はユーザー登録の単体テスト
func TestAuthHandler_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.POST("/api/v1/auth/register", handler.Register)

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

		var response AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Token != mockTokenGen.Token {
			t.Errorf("Expected token %s, got %s", mockTokenGen.Token, response.Token)
		}

		if response.User == nil {
			t.Fatal("User should not be nil")
		}

		if response.User.Email != reqBody.Email {
			t.Errorf("Expected email %s, got %s", reqBody.Email, response.User.Email)
		}

		// モックリポジトリにユーザーが追加されたことを確認
		if len(mockUserRepo.Users) != 1 {
			t.Errorf("Expected 1 user in repository, got %d", len(mockUserRepo.Users))
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		// 既存ユーザーを追加
		existingUser := &models.User{
			ID:    1,
			Name:  "Existing User",
			Email: "existing@example.com",
		}
		mockUserRepo.AddUser(existingUser)

		router := setupTestRouter()
		router.POST("/api/v1/auth/register", handler.Register)

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
	})

	t.Run("invalid request - missing email", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.POST("/api/v1/auth/register", handler.Register)

		reqBody := RegisterRequest{
			Password: "password123",
			Name:     "Test User",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid request - short password", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.POST("/api/v1/auth/register", handler.Register)

		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "short",
			Name:     "Test User",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("user creation error", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockUserRepo.CreateErr = errors.New("database error")
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.POST("/api/v1/auth/register", handler.Register)

		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("token generation error", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		mockTokenGen.GenerateErr = errors.New("token error")
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.POST("/api/v1/auth/register", handler.Register)

		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestAuthHandler_Login はログインの単体テスト
func TestAuthHandler_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		// テストユーザーを作成
		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.POST("/api/v1/auth/login", handler.Login)

		reqBody := LoginRequest{
			Email:    "test@example.com",
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

		var response AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Token != mockTokenGen.Token {
			t.Errorf("Expected token %s, got %s", mockTokenGen.Token, response.Token)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.POST("/api/v1/auth/login", handler.Login)

		reqBody := LoginRequest{
			Email:    "test@example.com",
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
	})

	t.Run("non-existent user", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.POST("/api/v1/auth/login", handler.Login)

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
	})

	t.Run("invalid request - bad json", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.POST("/api/v1/auth/login", handler.Login)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("token generation error", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		mockTokenGen.GenerateErr = errors.New("token generation failed")
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.POST("/api/v1/auth/login", handler.Login)

		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

// TestAuthHandler_GetProfile はプロファイル取得の単体テスト
func TestAuthHandler_GetProfile(t *testing.T) {
	t.Run("successful profile retrieval", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		// ユーザーIDをコンテキストに設定するミドルウェア
		router.GET("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.GetProfile)

		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]*models.User
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["user"] == nil {
			t.Fatal("User should not be nil")
		}

		if response["user"].Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, response["user"].Email)
		}
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.GET("/api/v1/auth/me", handler.GetProfile)

		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.GET("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(999))
			c.Next()
		}, handler.GetProfile)

		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// TestAuthHandler_UpdateProfile はプロファイル更新の単体テスト
func TestAuthHandler_UpdateProfile(t *testing.T) {
	t.Run("successful profile update", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

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

		var response map[string]*models.User
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["user"].Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got '%s'", response["user"].Name)
		}
	})

	t.Run("update error", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockUserRepo.UpdateErr = errors.New("update error")
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

		reqBody := UpdateProfileRequest{
			Name: "Updated Name",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("invalid request - bad json", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", handler.UpdateProfile)

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
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(999))
			c.Next()
		}, handler.UpdateProfile)

		reqBody := UpdateProfileRequest{
			Name: "Updated Name",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// TestAuthHandler_ChangePassword はパスワード変更の単体テスト
func TestAuthHandler_ChangePassword(t *testing.T) {
	t.Run("successful password change", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/password", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ChangePassword)

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

		// パスワードが変更されたことを確認
		updatedUser, _ := mockUserRepo.GetByID(1)
		if !updatedUser.CheckPassword("newpassword456") {
			t.Error("New password should be valid")
		}
	})

	t.Run("wrong current password", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/password", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ChangePassword)

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
	})

	t.Run("invalid request - bad json", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/password", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ChangePassword)

		req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/password", handler.ChangePassword)

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
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/password", func(c *gin.Context) {
			c.Set("userID", uint(999))
			c.Next()
		}, handler.ChangePassword)

		reqBody := ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "newpassword456",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("update error", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockUserRepo.UpdateErr = errors.New("update error")
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/password", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ChangePassword)

		reqBody := ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "newpassword456",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("short new password", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		user.SetPassword("password123")
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/password", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.ChangePassword)

		reqBody := ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "short",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestAuthHandler_UpdateProfile_Email はメールアドレス更新の単体テスト
func TestAuthHandler_UpdateProfile_Email(t *testing.T) {
	t.Run("successful email update", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

		reqBody := UpdateProfileRequest{
			Email: "new@example.com",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]*models.User
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["user"].Email != "new@example.com" {
			t.Errorf("Expected email 'new@example.com', got '%s'", response["user"].Email)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user1 := &models.User{
			ID:    1,
			Name:  "User 1",
			Email: "user1@example.com",
		}
		user2 := &models.User{
			ID:    2,
			Name:  "User 2",
			Email: "user2@example.com",
		}
		mockUserRepo.AddUser(user1)
		mockUserRepo.AddUser(user2)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

		reqBody := UpdateProfileRequest{
			Email: "user2@example.com",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
		}
	})

	t.Run("same email - no change", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

		reqBody := UpdateProfileRequest{
			Email: "test@example.com",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("update both name and email", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

		reqBody := UpdateProfileRequest{
			Name:  "Updated Name",
			Email: "updated@example.com",
		}

		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]*models.User
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["user"].Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got '%s'", response["user"].Name)
		}

		if response["user"].Email != "updated@example.com" {
			t.Errorf("Expected email 'updated@example.com', got '%s'", response["user"].Email)
		}
	})

	t.Run("invalid email format", func(t *testing.T) {
		mockUserRepo := NewMockUserRepository()
		mockTokenGen := NewMockTokenGenerator()
		handler := NewAuthHandler(mockUserRepo, mockTokenGen)

		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockUserRepo.AddUser(user)

		router := setupTestRouter()
		router.PUT("/api/v1/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}, handler.UpdateProfile)

		body := []byte(`{"email": "not-an-email"}`)
		req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestPasswordReset_Integration はパスワードリセットの統合テスト
func TestPasswordReset_Integration(t *testing.T) {
	t.Run("successful password reset flow", func(t *testing.T) {
		models.SetupTestDB(t)
		defer models.TeardownTestDB(t)

		// ユーザーを作成
		user := models.CreateTestUser(t, "Test User", "test@example.com", "oldpassword123")

		router := setupTestRouter()
		router.POST("/api/v1/auth/password-reset", RequestPasswordReset)
		router.POST("/api/v1/auth/password-reset/confirm", ConfirmPasswordReset)
		router.POST("/api/v1/auth/login", Login)

		// Step 1: パスワードリセットをリクエスト
		resetReqBody, _ := json.Marshal(PasswordResetRequest{
			Email: "test@example.com",
			Name:  "Test User",
		})
		resetReq, _ := http.NewRequest("POST", "/api/v1/auth/password-reset", bytes.NewBuffer(resetReqBody))
		resetReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, resetReq)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var resetResponse map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resetResponse); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		token, ok := resetResponse["reset_token"].(string)
		if !ok || token == "" {
			t.Fatal("reset_token should be present in response")
		}

		// Step 2: トークンで新パスワードを設定
		confirmReqBody, _ := json.Marshal(PasswordResetConfirmRequest{
			Token:       token,
			NewPassword: "newpassword456",
		})
		confirmReq, _ := http.NewRequest("POST", "/api/v1/auth/password-reset/confirm", bytes.NewBuffer(confirmReqBody))
		confirmReq.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, confirmReq)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		// Step 3: 新しいパスワードでログインできることを確認
		loginReqBody, _ := json.Marshal(LoginRequest{
			Email:    "test@example.com",
			Password: "newpassword456",
		})
		loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginReqBody))
		loginReq.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, loginReq)

		if w.Code != http.StatusOK {
			t.Errorf("Expected login with new password to succeed, got status %d. Body: %s", w.Code, w.Body.String())
		}

		// Step 4: 古いパスワードではログインできないことを確認
		oldLoginReqBody, _ := json.Marshal(LoginRequest{
			Email:    "test@example.com",
			Password: "oldpassword123",
		})
		oldLoginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(oldLoginReqBody))
		oldLoginReq.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, oldLoginReq)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected old password login to fail with %d, got %d", http.StatusUnauthorized, w.Code)
		}

		_ = user // suppress unused warning
	})

	t.Run("wrong name - no token issued", func(t *testing.T) {
		models.SetupTestDB(t)
		defer models.TeardownTestDB(t)

		models.CreateTestUser(t, "Test User", "test@example.com", "password123")

		router := setupTestRouter()
		router.POST("/api/v1/auth/password-reset", RequestPasswordReset)

		reqBody, _ := json.Marshal(PasswordResetRequest{
			Email: "test@example.com",
			Name:  "Wrong Name",
		})
		req, _ := http.NewRequest("POST", "/api/v1/auth/password-reset", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if _, exists := response["reset_token"]; exists {
			t.Error("reset_token should NOT be present when name doesn't match")
		}
	})

	t.Run("non-existent email - same response", func(t *testing.T) {
		models.SetupTestDB(t)
		defer models.TeardownTestDB(t)

		router := setupTestRouter()
		router.POST("/api/v1/auth/password-reset", RequestPasswordReset)

		reqBody, _ := json.Marshal(PasswordResetRequest{
			Email: "nonexistent@example.com",
			Name:  "Nobody",
		})
		req, _ := http.NewRequest("POST", "/api/v1/auth/password-reset", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if _, exists := response["reset_token"]; exists {
			t.Error("reset_token should NOT be present for non-existent email")
		}
	})

	t.Run("invalid request - missing fields", func(t *testing.T) {
		models.SetupTestDB(t)
		defer models.TeardownTestDB(t)

		router := setupTestRouter()
		router.POST("/api/v1/auth/password-reset", RequestPasswordReset)

		reqBody, _ := json.Marshal(map[string]string{
			"email": "test@example.com",
		})
		req, _ := http.NewRequest("POST", "/api/v1/auth/password-reset", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("confirm - invalid token", func(t *testing.T) {
		models.SetupTestDB(t)
		defer models.TeardownTestDB(t)

		router := setupTestRouter()
		router.POST("/api/v1/auth/password-reset/confirm", ConfirmPasswordReset)

		reqBody, _ := json.Marshal(PasswordResetConfirmRequest{
			Token:       "invalid-token",
			NewPassword: "newpassword456",
		})
		req, _ := http.NewRequest("POST", "/api/v1/auth/password-reset/confirm", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("confirm - used token", func(t *testing.T) {
		models.SetupTestDB(t)
		defer models.TeardownTestDB(t)

		user := models.CreateTestUser(t, "Test User", "test@example.com", "password123")

		// トークンを作成して使用済みにする
		resetToken, _ := models.CreatePasswordResetToken(user.ID)
		resetToken.MarkAsUsed()

		router := setupTestRouter()
		router.POST("/api/v1/auth/password-reset/confirm", ConfirmPasswordReset)

		reqBody, _ := json.Marshal(PasswordResetConfirmRequest{
			Token:       resetToken.Token,
			NewPassword: "newpassword456",
		})
		req, _ := http.NewRequest("POST", "/api/v1/auth/password-reset/confirm", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("confirm - short password", func(t *testing.T) {
		models.SetupTestDB(t)
		defer models.TeardownTestDB(t)

		router := setupTestRouter()
		router.POST("/api/v1/auth/password-reset/confirm", ConfirmPasswordReset)

		reqBody, _ := json.Marshal(PasswordResetConfirmRequest{
			Token:       "some-token",
			NewPassword: "short",
		})
		req, _ := http.NewRequest("POST", "/api/v1/auth/password-reset/confirm", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
