package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"household-app/handlers"
	"household-app/middleware"
	"household-app/models"
)

func main() {
	// データベースパスを取得
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/household.db"
	}

	// データベース接続
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// マイグレーション実行
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Ginモード設定
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	}

	// ルーター初期化
	r := gin.Default()

	// ヘルスチェック
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// ハンドラー初期化
	authHandler := handlers.NewAuthHandler(db)

	// APIルート
	api := r.Group("/api/v1")
	{
		// 認証不要のエンドポイント
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 認証が必要なエンドポイント
		authRequired := api.Group("/auth")
		authRequired.Use(middleware.AuthRequired())
		{
			authRequired.GET("/me", authHandler.GetProfile)
			authRequired.PUT("/me", authHandler.UpdateProfile)
			authRequired.PUT("/password", authHandler.ChangePassword)
		}
	}

	// ポート設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
