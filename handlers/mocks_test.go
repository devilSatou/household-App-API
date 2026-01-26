package handlers

import (
	"errors"
	"time"

	"household-app/middleware"
	"household-app/models"
)

// MockUserRepository はテスト用のユーザーリポジトリモック
type MockUserRepository struct {
	Users          map[uint]*models.User
	UsersByEmail   map[string]*models.User
	NextID         uint
	CreateErr      error
	GetByIDErr     error
	GetByEmailErr  error
	UpdateErr      error
	DeleteErr      error
}

// NewMockUserRepository は新しいモックリポジトリを作成する
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:        make(map[uint]*models.User),
		UsersByEmail: make(map[string]*models.User),
		NextID:       1,
	}
}

func (m *MockUserRepository) Create(name, email string) (*models.User, error) {
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	user := &models.User{
		ID:        m.NextID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.Users[user.ID] = user
	m.UsersByEmail[email] = user
	m.NextID++
	return user, nil
}

func (m *MockUserRepository) CreateWithPassword(name, email, password string) (*models.User, error) {
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	user := &models.User{
		ID:        m.NextID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	m.Users[user.ID] = user
	m.UsersByEmail[email] = user
	m.NextID++
	return user, nil
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	user, ok := m.Users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if m.GetByEmailErr != nil {
		return nil, m.GetByEmailErr
	}
	user, ok := m.UsersByEmail[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetAll() ([]models.User, error) {
	users := make([]models.User, 0, len(m.Users))
	for _, u := range m.Users {
		users = append(users, *u)
	}
	return users, nil
}

func (m *MockUserRepository) Update(user *models.User) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.Users[user.ID] = user
	m.UsersByEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) Delete(user *models.User) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.Users, user.ID)
	delete(m.UsersByEmail, user.Email)
	return nil
}

// AddUser はテスト用にユーザーを追加する
func (m *MockUserRepository) AddUser(user *models.User) {
	if user.ID == 0 {
		user.ID = m.NextID
		m.NextID++
	}
	m.Users[user.ID] = user
	if user.Email != "" {
		m.UsersByEmail[user.Email] = user
	}
}

// MockTokenGenerator はテスト用のトークンジェネレーターモック
type MockTokenGenerator struct {
	Token      string
	GenerateErr error
}

func NewMockTokenGenerator() *MockTokenGenerator {
	return &MockTokenGenerator{
		Token: "mock-token-12345",
	}
}

func (m *MockTokenGenerator) Generate(userID uint, email string) (string, error) {
	if m.GenerateErr != nil {
		return "", m.GenerateErr
	}
	return m.Token, nil
}

// MockTransactionRepository はテスト用のトランザクションリポジトリモック
type MockTransactionRepository struct {
	Transactions map[uint]*models.Transaction
	NextID       uint
	CreateErr    error
	GetByIDErr   error
	ListErr      error
	UpdateErr    error
	DeleteErr    error
}

// NewMockTransactionRepository は新しいモックリポジトリを作成する
func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		Transactions: make(map[uint]*models.Transaction),
		NextID:       1,
	}
}

func (m *MockTransactionRepository) Create(t *models.Transaction) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	t.ID = m.NextID
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	m.Transactions[t.ID] = t
	m.NextID++
	return nil
}

func (m *MockTransactionRepository) GetByID(id uint) (*models.Transaction, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	t, ok := m.Transactions[id]
	if !ok {
		return nil, errors.New("transaction not found")
	}
	return t, nil
}

func (m *MockTransactionRepository) List(params models.ListParams) ([]models.Transaction, int64, error) {
	if m.ListErr != nil {
		return nil, 0, m.ListErr
	}

	var result []models.Transaction
	for _, t := range m.Transactions {
		// フィルタリング
		if params.Type != "" && t.Type != params.Type {
			continue
		}
		if !params.StartDate.IsZero() && t.Date.Before(params.StartDate) {
			continue
		}
		if !params.EndDate.IsZero() && t.Date.After(params.EndDate) {
			continue
		}
		result = append(result, *t)
	}

	total := int64(len(result))

	// ページネーション
	if params.Offset > 0 && params.Offset < len(result) {
		result = result[params.Offset:]
	}
	if params.Limit > 0 && params.Limit < len(result) {
		result = result[:params.Limit]
	}

	return result, total, nil
}

func (m *MockTransactionRepository) Update(t *models.Transaction) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	t.UpdatedAt = time.Now()
	m.Transactions[t.ID] = t
	return nil
}

func (m *MockTransactionRepository) Delete(id uint) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.Transactions, id)
	return nil
}

func (m *MockTransactionRepository) GetMonthlySummary(year, month int) (*models.MonthlySummary, error) {
	return &models.MonthlySummary{
		Year:    year,
		Month:   month,
		Income:  10000,
		Expense: 5000,
		Balance: 5000,
	}, nil
}

func (m *MockTransactionRepository) GetCategorySummary(startDate, endDate time.Time) ([]models.CategorySummary, error) {
	return []models.CategorySummary{}, nil
}

func (m *MockTransactionRepository) GetYearlySummary(year int) (*models.YearlySummary, error) {
	return &models.YearlySummary{
		Year:    year,
		Income:  120000,
		Expense: 60000,
		Balance: 60000,
	}, nil
}

// AddTransaction はテスト用にトランザクションを追加する
func (m *MockTransactionRepository) AddTransaction(t *models.Transaction) {
	if t.ID == 0 {
		t.ID = m.NextID
		m.NextID++
	}
	m.Transactions[t.ID] = t
}

// MockTokenValidator はテスト用のトークンバリデーターモック
type MockTokenValidator struct {
	Claims      *middleware.Claims
	ValidateErr error
}

func NewMockTokenValidator() *MockTokenValidator {
	return &MockTokenValidator{}
}

func (m *MockTokenValidator) Validate(tokenString string) (*middleware.Claims, error) {
	if m.ValidateErr != nil {
		return nil, m.ValidateErr
	}
	return m.Claims, nil
}

// MockCategoryRepository はテスト用のカテゴリリポジトリモック
type MockCategoryRepository struct {
	Categories   map[uint]*models.Category
	NextID       uint
	CreateErr    error
	GetByIDErr   error
	ListErr      error
	UpdateErr    error
	DeleteErr    error
	GetByNameErr error
}

// NewMockCategoryRepository は新しいモックリポジトリを作成する
func NewMockCategoryRepository() *MockCategoryRepository {
	return &MockCategoryRepository{
		Categories: make(map[uint]*models.Category),
		NextID:     1,
	}
}

func (m *MockCategoryRepository) Create(c *models.Category) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	c.ID = m.NextID
	c.CreatedAt = time.Now()
	m.Categories[c.ID] = c
	m.NextID++
	return nil
}

func (m *MockCategoryRepository) GetByID(id uint) (*models.Category, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	c, ok := m.Categories[id]
	if !ok {
		return nil, errors.New("category not found")
	}
	return c, nil
}

func (m *MockCategoryRepository) List(categoryType string) ([]models.Category, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}

	var result []models.Category
	for _, c := range m.Categories {
		if categoryType == "" || c.Type == categoryType {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *MockCategoryRepository) GetByName(name string) (*models.Category, error) {
	if m.GetByNameErr != nil {
		return nil, m.GetByNameErr
	}
	for _, c := range m.Categories {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.New("category not found")
}

func (m *MockCategoryRepository) Update(c *models.Category) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.Categories[c.ID] = c
	return nil
}

func (m *MockCategoryRepository) Delete(id uint) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.Categories, id)
	return nil
}

// AddCategory はテスト用にカテゴリを追加する
func (m *MockCategoryRepository) AddCategory(c *models.Category) {
	if c.ID == 0 {
		c.ID = m.NextID
		m.NextID++
	}
	m.Categories[c.ID] = c
}
