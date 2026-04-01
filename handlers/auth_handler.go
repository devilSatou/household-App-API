package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

// AuthHandler は認証ハンドラーの構造体（依存性注入対応）
type AuthHandler struct {
	userRepo       models.UserRepositoryInterface
	tokenGenerator middleware.TokenGeneratorInterface
}

// NewAuthHandler は AuthHandler を作成する
func NewAuthHandler(userRepo models.UserRepositoryInterface, tokenGenerator middleware.TokenGeneratorInterface) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		tokenGenerator: tokenGenerator,
	}
}

// DefaultAuthHandler はデフォルトの依存関係で AuthHandler を作成する
func DefaultAuthHandler() *AuthHandler {
	return NewAuthHandler(
		models.NewUserRepository(),
		middleware.NewTokenGenerator(),
	)
}

// Register はユーザー登録を処理する
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// メールアドレスを小文字に正規化
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// メールアドレスの重複チェック
	if _, err := h.userRepo.GetByEmail(req.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already registered",
		})
		return
	}

	// ユーザーを作成
	user, err := h.userRepo.CreateWithPassword(strings.TrimSpace(req.Name), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	// JWTトークンを生成
	token, err := h.tokenGenerator.Generate(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login はユーザーログインを処理する
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// メールアドレスを小文字に正規化
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// ユーザーを検索
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// パスワードを検証
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// JWTトークンを生成
	token, err := h.tokenGenerator.Generate(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile は現在のユーザー情報を取得する
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateProfile は現在のユーザー情報を更新する
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// 名前を更新
	if req.Name != "" {
		user.Name = strings.TrimSpace(req.Name)
	}

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// RequestPasswordReset はパスワードリセットトークンを発行する
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email is registered, a password reset token has been issued",
		})
		return
	}

	models.InvalidateTokensForUser(user.ID)

	resetToken, err := models.CreatePasswordResetToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create reset token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "If the email is registered, a password reset token has been issued",
		"reset_token": resetToken.Token,
		"expires_at":  resetToken.ExpiresAt,
	})
}

// ConfirmPasswordReset はトークンを使ってパスワードをリセットする
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var req PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	resetToken, err := models.GetPasswordResetTokenByToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired token",
		})
		return
	}

	if resetToken.IsUsed() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token has already been used",
		})
		return
	}

	if resetToken.IsExpired() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token has expired",
		})
		return
	}

	user, err := h.userRepo.GetByID(resetToken.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "User not found",
		})
		return
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update password",
		})
		return
	}

	if err := resetToken.MarkAsUsed(); err != nil {
		_ = err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

// ChangePassword は現在のユーザーのパスワードを変更する
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// 現在のパスワードを検証
	if !user.CheckPassword(req.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Current password is incorrect",
		})
		return
	}

	// 新しいパスワードを設定
	if err := user.SetPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}
