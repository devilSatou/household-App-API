package models

import "time"

// UserRepositoryInterface はユーザーリポジトリのインターフェース
type UserRepositoryInterface interface {
	Create(name, email string) (*User, error)
	CreateWithPassword(name, email, password string) (*User, error)
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]User, error)
	Update(user *User) error
	Delete(user *User) error
}

// TransactionRepositoryInterface はトランザクションリポジトリのインターフェース
type TransactionRepositoryInterface interface {
	Create(t *Transaction) error
	GetByID(id uint) (*Transaction, error)
	List(params ListParams) ([]Transaction, int64, error)
	Update(t *Transaction) error
	Delete(id uint) error
	GetMonthlySummary(year, month int) (*MonthlySummary, error)
	GetCategorySummary(startDate, endDate time.Time) ([]CategorySummary, error)
	GetYearlySummary(year int) (*YearlySummary, error)
}

// CategoryRepositoryInterface はカテゴリリポジトリのインターフェース
type CategoryRepositoryInterface interface {
	Create(c *Category) error
	GetByID(id uint) (*Category, error)
	List(categoryType string) ([]Category, error)
	GetByName(name string) (*Category, error)
	Update(c *Category) error
	Delete(id uint) error
}

// TransactionRepository がインターフェースを満たすことを確認
var _ TransactionRepositoryInterface = (*TransactionRepository)(nil)

// UserRepositoryImpl は UserRepositoryInterface の実装
type UserRepositoryImpl struct{}

// NewUserRepository は UserRepositoryInterface を返す
func NewUserRepository() UserRepositoryInterface {
	return &UserRepositoryImpl{}
}

func (r *UserRepositoryImpl) Create(name, email string) (*User, error) {
	return CreateUser(name, email)
}

func (r *UserRepositoryImpl) CreateWithPassword(name, email, password string) (*User, error) {
	return CreateUserWithPassword(name, email, password)
}

func (r *UserRepositoryImpl) GetByID(id uint) (*User, error) {
	return GetUserByID(id)
}

func (r *UserRepositoryImpl) GetByEmail(email string) (*User, error) {
	return GetUserByEmail(email)
}

func (r *UserRepositoryImpl) GetAll() ([]User, error) {
	return GetAllUsers()
}

func (r *UserRepositoryImpl) Update(user *User) error {
	return user.Update()
}

func (r *UserRepositoryImpl) Delete(user *User) error {
	return user.Delete()
}
