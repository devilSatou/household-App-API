package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"household-app/handlers"
	"household-app/middleware"
	"household-app/models"
)

func main() {
	// Ginモード設定
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 共通DB初期化
	if _, err := models.GetAppDB(); err != nil {
		log.Fatalf("Failed to initialize app database: %v", err)
	}
	defer models.CloseAllDBs()

	// ルーター作成
	r := gin.Default()

	// ミドルウェア
	r.Use(middleware.CORS())

	// ヘルスチェック
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")

	// 認証エンドポイント（認証不要）
	auth := v1.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// 認証が必要な認証エンドポイント
	authRequired := v1.Group("/auth")
	authRequired.Use(middleware.AuthRequired())
	{
		authRequired.GET("/me", handlers.GetProfile)
		authRequired.PUT("/me", handlers.UpdateProfile)
		authRequired.PUT("/password", handlers.ChangePassword)
	}

	// 認証が必要なAPIエンドポイント
	// 本番: JWT認証必須、開発: X-User-IDヘッダーも許可
	api := v1.Group("")
	api.Use(middleware.AuthRequired())
	{
		// Transactions
		api.GET("/transactions", handlers.ListTransactions)
		api.POST("/transactions", handlers.CreateTransaction)
		api.GET("/transactions/:id", handlers.GetTransaction)
		api.PUT("/transactions/:id", handlers.UpdateTransaction)
		api.DELETE("/transactions/:id", handlers.DeleteTransaction)

		// Categories
		api.GET("/categories", handlers.ListCategories)
		api.POST("/categories", handlers.CreateCategory)

		// Summary
		api.GET("/summary/monthly", handlers.GetMonthlySummary)
		api.GET("/summary/category", handlers.GetCategorySummary)
		api.GET("/summary/yearly", handlers.GetYearlySummary)

		// Export
		api.GET("/export/csv", handlers.ExportCSV)
	}

	// ポート設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
