package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User はユーザー情報を表す（共通DBに保存）
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"unique" json:"email,omitempty"`
	Password  string    `gorm:"column:password" json:"-"` // JSONレスポンスには含めない
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName はテーブル名を指定
func (User) TableName() string {
	return "users"
}

// SetPassword はパスワードをハッシュ化して設定する
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword はパスワードが正しいかを検証する
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// CreateUser は新規ユーザーを作成する
func CreateUser(name, email string) (*User, error) {
	db, err := GetAppDB()
	if err != nil {
		return nil, err
	}

	user := &User{
		Name:  name,
		Email: email,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	// ユーザー専用DBを初期化（ディレクトリ作成とマイグレーション）
	if _, err := GetUserDB(user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateUserWithPassword はパスワード付きで新規ユーザーを作成する
func CreateUserWithPassword(name, email, password string) (*User, error) {
	db, err := GetAppDB()
	if err != nil {
		return nil, err
	}

	user := &User{
		Name:  name,
		Email: email,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	// ユーザー専用DBを初期化（ディレクトリ作成とマイグレーション）
	if _, err := GetUserDB(user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID はIDでユーザーを取得する
func GetUserByID(id uint) (*User, error) {
	db, err := GetAppDB()
	if err != nil {
		return nil, err
	}

	var user User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail はメールアドレスでユーザーを取得する
func GetUserByEmail(email string) (*User, error) {
	db, err := GetAppDB()
	if err != nil {
		return nil, err
	}

	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAllUsers は全ユーザーを取得する
func GetAllUsers() ([]User, error) {
	db, err := GetAppDB()
	if err != nil {
		return nil, err
	}

	var users []User
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// Update はユーザー情報を更新する
func (u *User) Update() error {
	db, err := GetAppDB()
	if err != nil {
		return err
	}

	return db.Save(u).Error
}

// Delete はユーザーを削除する
// 注意: ユーザー専用DBファイルは別途削除が必要
func (u *User) Delete() error {
	db, err := GetAppDB()
	if err != nil {
		return err
	}

	return db.Delete(u).Error
}
