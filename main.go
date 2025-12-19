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
	v1.Use(middleware.UserContext())
	{
		// Transactions
		v1.GET("/transactions", handlers.ListTransactions)
		v1.POST("/transactions", handlers.CreateTransaction)
		v1.GET("/transactions/:id", handlers.GetTransaction)
		v1.PUT("/transactions/:id", handlers.UpdateTransaction)
		v1.DELETE("/transactions/:id", handlers.DeleteTransaction)

		// Categories
		v1.GET("/categories", handlers.ListCategories)
		v1.POST("/categories", handlers.CreateCategory)

		// Summary
		v1.GET("/summary/monthly", handlers.GetMonthlySummary)
		v1.GET("/summary/category", handlers.GetCategorySummary)
		v1.GET("/summary/yearly", handlers.GetYearlySummary)

		// Export
		v1.GET("/export/csv", handlers.ExportCSV)
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
