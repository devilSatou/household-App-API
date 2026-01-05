package middleware

import (
	"os"
	"testing"
	"time"

	"household-app/models"
)

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
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	// Ginコンテキストのモックを作成する必要があるため、
	// このテストは統合テストとして handlers のテストで実施
	t.Skip("This is tested in integration tests")
}

func TestAuthRequired(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	// Ginコンテキストのモックを作成する必要があるため、
	// このテストは統合テストとして handlers のテストで実施
	t.Skip("This is tested in integration tests")
}

func TestUserContext(t *testing.T) {
	models.SetupTestDB(t)
	defer models.TeardownTestDB(t)

	// Ginコンテキストのモックを作成する必要があるため、
	// このテストは統合テストとして handlers のテストで実施
	t.Skip("This is tested in integration tests")
}
