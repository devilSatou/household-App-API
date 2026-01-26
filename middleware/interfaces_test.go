package middleware

import (
	"testing"
)

func TestNewTokenGenerator(t *testing.T) {
	generator := NewTokenGenerator()

	if generator == nil {
		t.Error("Expected non-nil generator")
	}

	// 型チェック
	_, ok := generator.(*TokenGeneratorImpl)
	if !ok {
		t.Error("Expected *TokenGeneratorImpl type")
	}
}

func TestTokenGeneratorImpl_Generate(t *testing.T) {
	generator := NewTokenGenerator()

	token, err := generator.Generate(1, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestNewTokenValidator(t *testing.T) {
	validator := NewTokenValidator()

	if validator == nil {
		t.Error("Expected non-nil validator")
	}

	// 型チェック
	_, ok := validator.(*TokenValidatorImpl)
	if !ok {
		t.Error("Expected *TokenValidatorImpl type")
	}
}

func TestTokenValidatorImpl_Validate(t *testing.T) {
	generator := NewTokenGenerator()
	validator := NewTokenValidator()

	// トークンを生成
	token, err := generator.Generate(1, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// トークンを検証
	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", claims.Email)
	}
}

func TestTokenValidatorImpl_Validate_InvalidToken(t *testing.T) {
	validator := NewTokenValidator()

	_, err := validator.Validate("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
