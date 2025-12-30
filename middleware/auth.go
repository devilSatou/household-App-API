package middleware

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"household-app/models"
)

// UserContext はユーザーDBをContextに注入するミドルウェア
// 現在は環境変数またはヘッダーからユーザーIDを取得（認証実装後はJWTから取得）
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
