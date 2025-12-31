package models

import (
	"time"
)

// Category はカテゴリマスタを表す（ユーザー専用DBに保存）
type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"unique;not null" json:"name"`
	Type      string    `gorm:"not null" json:"type"` // income or expense
	CreatedAt time.Time `json:"created_at"`
}

// TableName はテーブル名を指定
func (Category) TableName() string {
	return "categories"
}

// CategoryRepository はユーザー専用DBでの操作を提供
type CategoryRepository struct {
	userID uint
}

// NewCategoryRepository はリポジトリを作成する
func NewCategoryRepository(userID uint) *CategoryRepository {
	return &CategoryRepository{userID: userID}
}

// Create は新規カテゴリを作成する
func (r *CategoryRepository) Create(c *Category) error {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return err
	}

	return db.Create(c).Error
}

// GetByID はIDでカテゴリを取得する
func (r *CategoryRepository) GetByID(id uint) (*Category, error) {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return nil, err
	}

	var c Category
	if err := db.First(&c, id).Error; err != nil {
		return nil, err
	}

	return &c, nil
}

// List はカテゴリ一覧を取得する
func (r *CategoryRepository) List(categoryType string) ([]Category, error) {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return nil, err
	}

	query := db.Model(&Category{})

	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	var categories []Category
	if err := query.Order("type, name").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// Update はカテゴリを更新する
func (r *CategoryRepository) Update(c *Category) error {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return err
	}

	return db.Save(c).Error
}

// Delete はカテゴリを削除する
// 注意: このカテゴリを使用しているTransactionがある場合はエラーになる可能性あり
func (r *CategoryRepository) Delete(id uint) error {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return err
	}

	return db.Delete(&Category{}, id).Error
}

// GetByName は名前でカテゴリを取得する
func (r *CategoryRepository) GetByName(name string) (*Category, error) {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return nil, err
	}

	var c Category
	if err := db.Where("name = ?", name).First(&c).Error; err != nil {
		return nil, err
	}

	return &c, nil
}
