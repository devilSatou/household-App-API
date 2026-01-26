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
