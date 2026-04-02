package models

import (
	"testing"
	"time"
)

func TestCreatePasswordResetToken(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")

	resetToken, err := CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatalf("Failed to create password reset token: %v", err)
	}

	if resetToken.ID == 0 {
		t.Error("Token ID should not be 0")
	}

	if resetToken.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, resetToken.UserID)
	}

	if resetToken.Token == "" {
		t.Error("Token string should not be empty")
	}

	if len(resetToken.Token) != 64 {
		t.Errorf("Expected token length 64, got %d", len(resetToken.Token))
	}

	if resetToken.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not already be expired")
	}

	if resetToken.UsedAt != nil {
		t.Error("Token should not be used at creation")
	}
}

func TestGetPasswordResetTokenByToken(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")

	created, err := CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	retrieved, err := GetPasswordResetTokenByToken(created.Token)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}

	if retrieved.UserID != created.UserID {
		t.Errorf("Expected user ID %d, got %d", created.UserID, retrieved.UserID)
	}
}

func TestGetPasswordResetTokenByToken_NotFound(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	_, err := GetPasswordResetTokenByToken("nonexistent-token")
	if err == nil {
		t.Error("Expected error for non-existent token, got nil")
	}
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		token := &PasswordResetToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if token.IsExpired() {
			t.Error("Token should not be expired")
		}
	})

	t.Run("expired", func(t *testing.T) {
		token := &PasswordResetToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		if !token.IsExpired() {
			t.Error("Token should be expired")
		}
	})
}

func TestPasswordResetToken_IsUsed(t *testing.T) {
	t.Run("not used", func(t *testing.T) {
		token := &PasswordResetToken{}
		if token.IsUsed() {
			t.Error("Token should not be used")
		}
	})

	t.Run("used", func(t *testing.T) {
		now := time.Now()
		token := &PasswordResetToken{
			UsedAt: &now,
		}
		if !token.IsUsed() {
			t.Error("Token should be used")
		}
	})
}

func TestPasswordResetToken_MarkAsUsed(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")

	resetToken, err := CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	if err := resetToken.MarkAsUsed(); err != nil {
		t.Fatalf("Failed to mark token as used: %v", err)
	}

	if resetToken.UsedAt == nil {
		t.Error("UsedAt should be set after marking as used")
	}

	// DBから再取得して確認
	retrieved, err := GetPasswordResetTokenByToken(resetToken.Token)
	if err != nil {
		t.Fatalf("Failed to retrieve token: %v", err)
	}

	if !retrieved.IsUsed() {
		t.Error("Token should be marked as used in DB")
	}
}

func TestInvalidateTokensForUser(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")

	// 複数のトークンを作成
	token1, _ := CreatePasswordResetToken(user.ID)
	token2, _ := CreatePasswordResetToken(user.ID)

	// 全トークンを無効化
	if err := InvalidateTokensForUser(user.ID); err != nil {
		t.Fatalf("Failed to invalidate tokens: %v", err)
	}

	// 両方のトークンが使用済みになっていることを確認
	retrieved1, _ := GetPasswordResetTokenByToken(token1.Token)
	retrieved2, _ := GetPasswordResetTokenByToken(token2.Token)

	if !retrieved1.IsUsed() {
		t.Error("Token 1 should be invalidated")
	}

	if !retrieved2.IsUsed() {
		t.Error("Token 2 should be invalidated")
	}
}

func TestPasswordResetToken_TableName(t *testing.T) {
	token := PasswordResetToken{}
	if token.TableName() != "password_reset_tokens" {
		t.Errorf("Expected table name 'password_reset_tokens', got '%s'", token.TableName())
	}
}

func TestPasswordResetToken_UniqueTokens(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user := CreateTestUser(t, "Test User", "test@example.com", "password123")

	token1, _ := CreatePasswordResetToken(user.ID)
	token2, _ := CreatePasswordResetToken(user.ID)

	if token1.Token == token2.Token {
		t.Error("Each token should be unique")
	}
}
