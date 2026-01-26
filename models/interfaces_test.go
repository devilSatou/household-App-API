package models

import (
	"testing"
)

func TestUserRepositoryImpl_Create(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	repo := NewUserRepository()

	user, err := repo.Create("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", user.Name)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestUserRepositoryImpl_CreateWithPassword(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	repo := NewUserRepository()

	user, err := repo.CreateWithPassword("Test User", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user with password: %v", err)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", user.Name)
	}

	if !user.CheckPassword("password123") {
		t.Error("Password verification failed")
	}
}

func TestUserRepositoryImpl_GetByID(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	repo := NewUserRepository()

	// ユーザーを作成
	created, err := repo.Create("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// IDで取得
	user, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, user.ID)
	}
}

func TestUserRepositoryImpl_GetByEmail(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	repo := NewUserRepository()

	// ユーザーを作成
	_, err := repo.Create("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// メールで取得
	user, err := repo.GetByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestUserRepositoryImpl_GetAll(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	repo := NewUserRepository()

	// 複数ユーザーを作成
	_, err := repo.Create("User 1", "user1@example.com")
	if err != nil {
		t.Fatalf("Failed to create user 1: %v", err)
	}

	_, err = repo.Create("User 2", "user2@example.com")
	if err != nil {
		t.Fatalf("Failed to create user 2: %v", err)
	}

	// 全ユーザー取得
	users, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(users))
	}
}

func TestUserRepositoryImpl_Update(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	repo := NewUserRepository()

	// ユーザーを作成
	user, err := repo.Create("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 更新
	user.Name = "Updated Name"
	err = repo.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// 確認
	updated, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestUserRepositoryImpl_Delete(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	repo := NewUserRepository()

	// ユーザーを作成
	user, err := repo.Create("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	userID := user.ID

	// 削除
	err = repo.Delete(user)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// 確認
	_, err = repo.GetByID(userID)
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestNewUserRepository(t *testing.T) {
	repo := NewUserRepository()

	if repo == nil {
		t.Error("Expected non-nil repository")
	}

	// 型チェック
	_, ok := repo.(*UserRepositoryImpl)
	if !ok {
		t.Error("Expected *UserRepositoryImpl type")
	}
}
