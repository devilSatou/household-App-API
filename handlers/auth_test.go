package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return r
}

func TestRegister(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	router := setupTestRouter()
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

	var response AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("Token should not be empty")
	}

	if response.User == nil {
		t.Fatal("User should not be nil")
	}

	if response.User.Email != reqBody.Email {
		t.Errorf("Expected email %s, got %s", reqBody.Email, response.User.Email)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// 既存ユーザーを作成
	models.CreateTestUser(t, "Existing User", "existing@example.com", "password123")

	router := setupTestRouter()
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

func TestRegister_InvalidRequest(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := setupTestRouter()
	router.POST("/api/v1/auth/register", Register)

	tests := []struct {
		name    string
		reqBody RegisterRequest
	}{
		{
			name: "missing email",
			reqBody: RegisterRequest{
				Password: "password123",
				Name:     "Test User",
			},
		},
		{
			name: "invalid email",
			reqBody: RegisterRequest{
				Email:    "invalid-email",
				Password: "password123",
				Name:     "Test User",
			},
		},
		{
			name: "short password",
			reqBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
				Name:     "Test User",
			},
		},
		{
			name: "missing name",
			reqBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// テストユーザーを作成
	email := "test@example.com"
	password := "password123"
	models.CreateTestUser(t, "Test User", email, password)

	router := setupTestRouter()
	router.POST("/api/v1/auth/login", Login)

	reqBody := LoginRequest{
		Email:    email,
		Password: password,
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

	if response.Token == "" {
		t.Error("Token should not be empty")
	}

	if response.User == nil {
		t.Fatal("User should not be nil")
	}

	if response.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, response.User.Email)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// テストユーザーを作成
	models.CreateTestUser(t, "Test User", "test@example.com", "password123")

	router := setupTestRouter()
	router.POST("/api/v1/auth/login", Login)

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"wrong password", "test@example.com", "wrongpassword"},
		{"non-existent email", "nonexistent@example.com", "password123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := LoginRequest{
				Email:    tt.email,
				Password: tt.password,
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
	}
}

func TestGetProfile(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// テストユーザーを作成
	user := models.CreateTestUser(t, "Test User", "test@example.com", "password123")

	// トークンを生成
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := setupTestRouter()
	router.GET("/api/v1/auth/me", middleware.AuthRequired(), GetProfile)

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

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
}

func TestGetProfile_Unauthorized(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	router := setupTestRouter()
	router.GET("/api/v1/auth/me", middleware.AuthRequired(), GetProfile)

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpdateProfile(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// テストユーザーを作成
	user := models.CreateTestUser(t, "Test User", "test@example.com", "password123")

	// トークンを生成
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := setupTestRouter()
	router.PUT("/api/v1/auth/me", middleware.AuthRequired(), UpdateProfile)

	reqBody := UpdateProfileRequest{
		Name: "Updated Name",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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
}

func TestChangePassword(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// テストユーザーを作成
	currentPassword := "password123"
	user := models.CreateTestUser(t, "Test User", "test@example.com", currentPassword)

	// トークンを生成
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := setupTestRouter()
	router.PUT("/api/v1/auth/password", middleware.AuthRequired(), ChangePassword)

	reqBody := ChangePasswordRequest{
		CurrentPassword: currentPassword,
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// パスワードが変更されたか確認
	updatedUser, err := models.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if !updatedUser.CheckPassword("newpassword456") {
		t.Error("New password should be valid")
	}

	if updatedUser.CheckPassword(currentPassword) {
		t.Error("Old password should no longer be valid")
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// テストユーザーを作成
	user := models.CreateTestUser(t, "Test User", "test@example.com", "password123")

	// トークンを生成
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := setupTestRouter()
	router.PUT("/api/v1/auth/password", middleware.AuthRequired(), ChangePassword)

	reqBody := ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
