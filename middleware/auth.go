package middleware

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"household-app/models"
)

// Claims はJWTのクレームを表す構造体
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// getJWTSecret は環境変数からJWTシークレットを取得する
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key-change-in-production"
	}
	return []byte(secret)
}

// getJWTExpiry は環境変数からJWTの有効期限を取得する（時間単位）
func getJWTExpiry() time.Duration {
	expiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	if expiryStr == "" {
		return 24 * time.Hour // デフォルト: 24時間
	}
	hours, err := strconv.Atoi(expiryStr)
	if err != nil {
		return 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

// GenerateToken はユーザー情報からJWTトークンを生成する
func GenerateToken(userID uint, email string) (string, error) {
	expiryDuration := getJWTExpiry()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "household-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ValidateToken はJWTトークンを検証してクレームを返す
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AuthRequired はJWT認証を要求するミドルウェア
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// "Bearer <token>" 形式からトークンを抽出
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// ユーザーDBを取得
		db, err := models.GetUserDB(claims.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user database"})
			c.Abort()
			return
		}

		// コンテキストにユーザー情報を設定
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userDB", db)

		c.Next()
	}
}

// UserContext はユーザーDBをContextに注入するミドルウェア
// 開発用: 環境変数またはヘッダーからユーザーIDを取得
// 本番ではAuthRequired()を使用してください
func UserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID uint

		// 1. ヘッダーから取得（開発用）
		if idStr := c.GetHeader("X-User-ID"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				userID = uint(id)
			}
		}

		// 2. 環境変数から取得（フォールバック）
		if userID == 0 {
			if idStr := os.Getenv("USER_ID"); idStr != "" {
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					userID = uint(id)
				}
			}
		}

		// 3. デフォルトユーザー（開発用）
		if userID == 0 {
			userID = 1
			// デフォルトユーザーが存在しない場合は作成
			if _, err := models.GetUserByID(userID); err != nil {
				if _, err := models.CreateUser("Default User", ""); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default user"})
					c.Abort()
					return
				}
			}
		}

		// ユーザーDBを取得
		db, err := models.GetUserDB(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user database"})
			c.Abort()
			return
		}

		// Contextに保存
		c.Set("userID", userID)
		c.Set("userDB", db)

		c.Next()
	}
}

// GetUserID はContextからユーザーIDを取得する
func GetUserID(c *gin.Context) uint {
	if id, exists := c.Get("userID"); exists {
		return id.(uint)
	}
	return 0
}

// GetUserEmail はコンテキストからメールアドレスを取得するヘルパー関数
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}
