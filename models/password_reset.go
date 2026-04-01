package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// PasswordResetToken はパスワードリセットトークンを表す（共通DBに保存）
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName はテーブル名を指定
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsExpired はトークンが期限切れかどうかを判定する
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed はトークンが使用済みかどうかを判定する
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// generateToken はセキュアなランダムトークンを生成する
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreatePasswordResetToken はパスワードリセットトークンを作成する
func CreatePasswordResetToken(userID uint) (*PasswordResetToken, error) {
	db, err := GetAppDB()
	if err != nil {
		return nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	resetToken := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := db.Create(resetToken).Error; err != nil {
		return nil, err
	}

	return resetToken, nil
}

// GetPasswordResetTokenByToken はトークン文字列でリセットトークンを取得する
func GetPasswordResetTokenByToken(token string) (*PasswordResetToken, error) {
	db, err := GetAppDB()
	if err != nil {
		return nil, err
	}

	var resetToken PasswordResetToken
	if err := db.Where("token = ?", token).First(&resetToken).Error; err != nil {
		return nil, err
	}

	return &resetToken, nil
}

// MarkAsUsed はトークンを使用済みにする
func (t *PasswordResetToken) MarkAsUsed() error {
	db, err := GetAppDB()
	if err != nil {
		return err
	}

	now := time.Now()
	t.UsedAt = &now
	return db.Save(t).Error
}

// InvalidateTokensForUser はユーザーの未使用トークンを無効化する
func InvalidateTokensForUser(userID uint) error {
	db, err := GetAppDB()
	if err != nil {
		return err
	}

	now := time.Now()
	return db.Model(&PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}
