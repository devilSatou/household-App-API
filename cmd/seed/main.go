package main

import (
	"fmt"
	"log"
	"os"

	"household-app/models"
)

func main() {
	// 環境変数がなければデフォルト値を設定
	if os.Getenv("DB_PATH") == "" {
		os.Setenv("DB_PATH", "data/app.db")
	}
	if os.Getenv("USER_DB_BASE_PATH") == "" {
		os.Setenv("USER_DB_BASE_PATH", "data/users")
	}

	fmt.Println("Starting seed process...")

	// 共通DB初期化
	if _, err := models.GetAppDB(); err != nil {
		log.Fatalf("Failed to initialize app database: %v", err)
	}
	defer models.CloseAllDBs()

	// シード実行
	if err := models.SeedAll(); err != nil {
		log.Fatalf("Seed failed: %v", err)
	}

	fmt.Println("Seed process completed!")
}
