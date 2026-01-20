package models

import (
	"testing"

)

func TestCreateUser(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should not be 0")
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", user.Name)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestCreateUserWithPassword(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	password := "securepass123"
	user, err := CreateUserWithPassword("Test User", "test@example.com", password)
	if err != nil {
		t.Fatalf("Failed to create user with password: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should not be 0")
	}

	// パスワードが正しくハッシュ化されているか確認
	if !user.CheckPassword(password) {
		t.Error("Password verification failed")
	}

	// 間違ったパスワードでは失敗することを確認
	if user.CheckPassword("wrongpassword") {
		t.Error("Wrong password should not verify")
	}
}

func TestGetUserByID(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	created, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}

	if retrieved.Email != created.Email {
		t.Errorf("Expected email %s, got %s", created.Email, retrieved.Email)
	}
}

func TestGetUserByEmail(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	email := "test@example.com"
	created, err := CreateUser("Test User", email)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := GetUserByEmail(email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	_, err := GetUserByEmail("nonexistent@example.com")
	if err == nil {
		t.Error("Expected error for non-existent email, got nil")
	}
}

func TestUserUpdate(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user.Name = "Updated Name"
	if err := user.Update(); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	retrieved, err := GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated user: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
	}
}

func TestUserSetPassword(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	password := "newpassword123"
	if err := user.SetPassword(password); err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	if err := user.Update(); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// パスワードが正しく設定されているか確認
	retrieved, err := GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}

	if !retrieved.CheckPassword(password) {
		t.Error("Password verification failed after update")
	}
}

func TestGetAllUsers(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	// 複数のユーザーを作成
	_, err := CreateUser("User 1", "user1@example.com")
	if err != nil {
		t.Fatalf("Failed to create user 1: %v", err)
	}

	_, err = CreateUser("User 2", "user2@example.com")
	if err != nil {
		t.Fatalf("Failed to create user 2: %v", err)
	}

	users, err := GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestUserDelete(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user, err := CreateUser("Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// ユーザーを削除
	if err := user.Delete(); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// 削除されたユーザーを取得しようとするとエラーになることを確認
	_, err = GetUserByID(user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user, got nil")
	}
}

func TestUserDelete_WithPassword(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	user, err := CreateUserWithPassword("Test User", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// ユーザーを削除
	if err := user.Delete(); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// 削除されたユーザーをメールで取得しようとするとエラーになることを確認
	_, err = GetUserByEmail("test@example.com")
	if err == nil {
		t.Error("Expected error when getting deleted user by email, got nil")
	}
}

func TestUser_TableName(t *testing.T) {
	user := User{}
	if user.TableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", user.TableName())
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	SetupTestDB(t)
	defer TeardownTestDB(t)

	_, err := GetUserByID(99999)
	if err == nil {
		t.Error("Expected error for non-existent user ID, got nil")
	}
}
