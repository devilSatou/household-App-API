package models

import (
	"time"
)

// 収支タイプの定数
const (
	TypeIncome  = "income"
	TypeExpense = "expense"
)

// Transaction は収支記録を表す（ユーザー専用DBに保存）
type Transaction struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CategoryID uint      `gorm:"index" json:"category_id"`
	Amount     int       `gorm:"not null" json:"amount"`
	Type       string    `gorm:"not null" json:"type"` // income or expense
	Date       time.Time `gorm:"not null;index" json:"date"`
	Memo       string    `gorm:"type:text" json:"memo"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// リレーション（読み取り専用）
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName はテーブル名を指定
func (Transaction) TableName() string {
	return "transactions"
}

// TransactionRepository はユーザー専用DBでの操作を提供
type TransactionRepository struct {
	userID uint
}

// NewTransactionRepository はリポジトリを作成する
func NewTransactionRepository(userID uint) *TransactionRepository {
	return &TransactionRepository{userID: userID}
}

// Create は新規収支記録を作成する
func (r *TransactionRepository) Create(t *Transaction) error {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return err
	}

	return db.Create(t).Error
}

// GetByID はIDで収支記録を取得する
func (r *TransactionRepository) GetByID(id uint) (*Transaction, error) {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return nil, err
	}

	var t Transaction
	if err := db.Preload("Category").First(&t, id).Error; err != nil {
		return nil, err
	}

	return &t, nil
}

// ListParams は一覧取得のパラメータ
type ListParams struct {
	Limit     int
	Offset    int
	Type      string    // income, expense, または空（全て）
	StartDate time.Time // 開始日（ゼロ値の場合は無視）
	EndDate   time.Time // 終了日（ゼロ値の場合は無視）
}

// List は収支記録の一覧を取得する
func (r *TransactionRepository) List(params ListParams) ([]Transaction, int64, error) {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return nil, 0, err
	}

	query := db.Model(&Transaction{})

	// フィルタ適用
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if !params.StartDate.IsZero() {
		query = query.Where("date >= ?", params.StartDate)
	}
	if !params.EndDate.IsZero() {
		query = query.Where("date <= ?", params.EndDate)
	}

	// 総件数取得
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// データ取得
	var transactions []Transaction
	q := query.Preload("Category").Order("date DESC, id DESC")

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	if err := q.Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// Update は収支記録を更新する
func (r *TransactionRepository) Update(t *Transaction) error {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return err
	}

	return db.Save(t).Error
}

// Delete は収支記録を削除する
func (r *TransactionRepository) Delete(id uint) error {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return err
	}

	return db.Delete(&Transaction{}, id).Error
}

// MonthlySummary は月次集計結果
type MonthlySummary struct {
	Year    int `json:"year"`
	Month   int `json:"month"`
	Income  int `json:"income"`
	Expense int `json:"expense"`
	Balance int `json:"balance"`
}

// GetMonthlySummary は月次集計を取得する
func (r *TransactionRepository) GetMonthlySummary(year, month int) (*MonthlySummary, error) {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return nil, err
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	var income, expense int

	// 収入合計
	db.Model(&Transaction{}).
		Where("type = ? AND date >= ? AND date <= ?", TypeIncome, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&income)

	// 支出合計
	db.Model(&Transaction{}).
		Where("type = ? AND date >= ? AND date <= ?", TypeExpense, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&expense)

	return &MonthlySummary{
		Year:    year,
		Month:   month,
		Income:  income,
		Expense: expense,
		Balance: income - expense,
	}, nil
}

// CategorySummary はカテゴリ別集計結果
type CategorySummary struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	Type         string `json:"type"`
	Total        int    `json:"total"`
}

// GetCategorySummary はカテゴリ別集計を取得する
func (r *TransactionRepository) GetCategorySummary(startDate, endDate time.Time) ([]CategorySummary, error) {
	db, err := GetUserDB(r.userID)
	if err != nil {
		return nil, err
	}

	var results []CategorySummary

	query := db.Model(&Transaction{}).
		Select("transactions.category_id, categories.name as category_name, transactions.type, SUM(transactions.amount) as total").
		Joins("LEFT JOIN categories ON transactions.category_id = categories.id").
		Group("transactions.category_id, transactions.type")

	if !startDate.IsZero() {
		query = query.Where("transactions.date >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("transactions.date <= ?", endDate)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// YearlySummary は年次集計結果
type YearlySummary struct {
	Year    int              `json:"year"`
	Income  int              `json:"income"`
	Expense int              `json:"expense"`
	Balance int              `json:"balance"`
	Monthly []MonthlySummary `json:"monthly"`
}

// GetYearlySummary は年次集計を取得する
func (r *TransactionRepository) GetYearlySummary(year int) (*YearlySummary, error) {
	summary := &YearlySummary{
		Year:    year,
		Monthly: make([]MonthlySummary, 12),
	}

	for month := 1; month <= 12; month++ {
		monthly, err := r.GetMonthlySummary(year, month)
		if err != nil {
			return nil, err
		}
		summary.Monthly[month-1] = *monthly
		summary.Income += monthly.Income
		summary.Expense += monthly.Expense
	}

	summary.Balance = summary.Income - summary.Expense
	return summary, nil
}
