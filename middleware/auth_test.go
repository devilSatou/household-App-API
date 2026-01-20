package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateToken(t *testing.T) {
	// JWT_SECRETを設定
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	token, err := GenerateToken(1, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}
}

func TestValidateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	userID := uint(123)
	email := "test@example.com"

	// トークンを生成
	token, err := GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// トークンを検証
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 無効なトークン
	invalidToken := "invalid.token.here"

	_, err := ValidateToken(invalidToken)
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	// 異なるシークレットでトークンを生成
	os.Setenv("JWT_SECRET", "secret1")
	token, err := GenerateToken(1, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 異なるシークレットで検証
	os.Setenv("JWT_SECRET", "secret2")
	_, err = ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating with wrong secret, got nil")
	}
}

func TestJWTExpiry(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("JWT_EXPIRY_HOURS", "24")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRY_HOURS")
	}()

	token, err := GenerateToken(1, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// 有効期限が24時間後に設定されているか確認
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// 許容誤差1分
	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Expected expiry around %v, got %v", expectedExpiry, actualExpiry)
	}
}

func TestGetUserID(t *testing.T) {
	t.Run("returns userID when set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(123))

		userID := GetUserID(c)
		if userID != 123 {
			t.Errorf("Expected userID 123, got %d", userID)
		}
	})

	t.Run("returns 0 when not set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		userID := GetUserID(c)
		if userID != 0 {
			t.Errorf("Expected userID 0, got %d", userID)
		}
	})
}

func TestGetUserEmail(t *testing.T) {
	t.Run("returns email when set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("email", "test@example.com")

		email, ok := GetUserEmail(c)
		if !ok {
			t.Error("Expected ok to be true")
		}
		if email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", email)
		}
	})

	t.Run("returns false when not set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		email, ok := GetUserEmail(c)
		if ok {
			t.Error("Expected ok to be false")
		}
		if email != "" {
			t.Errorf("Expected empty email, got '%s'", email)
		}
	})
}

func TestAuthRequired(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// テストユーザーを作成
	user := models.CreateTestUser(t, "Test User", "test@example.com", "password123")

	t.Run("successful authentication", func(t *testing.T) {
		token, _ := GenerateToken(user.ID, user.Email)

		router := gin.New()
		router.GET("/protected", AuthRequired(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("missing authorization header", func(t *testing.T) {
		router := gin.New()
		router.GET("/protected", AuthRequired(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		router := gin.New()
		router.GET("/protected", AuthRequired(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token123")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		router := gin.New()
		router.GET("/protected", AuthRequired(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestUserContext(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	t.Run("with X-User-ID header", func(t *testing.T) {
		// テストユーザーを作成
		user := models.CreateTestUser(t, "Test User", "header@example.com", "password123")

		router := gin.New()
		router.GET("/test", UserContext(), func(c *gin.Context) {
			userID := GetUserID(c)
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		// クリーンアップ
		_ = user
	})

	t.Run("with USER_ID env var", func(t *testing.T) {
		os.Setenv("USER_ID", "1")
		defer os.Unsetenv("USER_ID")

		router := gin.New()
		router.GET("/test", UserContext(), func(c *gin.Context) {
			userID := GetUserID(c)
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req, _ := http.NewRequest("GET", "/test", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("default user creation", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", UserContext(), func(c *gin.Context) {
			userID := GetUserID(c)
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req, _ := http.NewRequest("GET", "/test", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// デフォルトユーザー (ID: 1) が使用される
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})
}

func TestGetJWTSecret(t *testing.T) {
	t.Run("returns env var value", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "my-custom-secret")
		defer os.Unsetenv("JWT_SECRET")

		secret := getJWTSecret()
		if string(secret) != "my-custom-secret" {
			t.Errorf("Expected 'my-custom-secret', got '%s'", string(secret))
		}
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET")

		secret := getJWTSecret()
		expected := "default-secret-key-change-in-production"
		if string(secret) != expected {
			t.Errorf("Expected '%s', got '%s'", expected, string(secret))
		}
	})
}

func TestGetJWTExpiry(t *testing.T) {
	t.Run("returns env var value", func(t *testing.T) {
		os.Setenv("JWT_EXPIRY_HOURS", "48")
		defer os.Unsetenv("JWT_EXPIRY_HOURS")

		expiry := getJWTExpiry()
		expected := 48 * time.Hour
		if expiry != expected {
			t.Errorf("Expected %v, got %v", expected, expiry)
		}
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Unsetenv("JWT_EXPIRY_HOURS")

		expiry := getJWTExpiry()
		expected := 24 * time.Hour
		if expiry != expected {
			t.Errorf("Expected %v, got %v", expected, expiry)
		}
	})

	t.Run("returns default on invalid value", func(t *testing.T) {
		os.Setenv("JWT_EXPIRY_HOURS", "invalid")
		defer os.Unsetenv("JWT_EXPIRY_HOURS")

		expiry := getJWTExpiry()
		expected := 24 * time.Hour
		if expiry != expected {
			t.Errorf("Expected %v, got %v", expected, expiry)
		}
	})
}
