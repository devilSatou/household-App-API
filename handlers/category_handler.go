package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

// CategoryRepositoryFactory は userID から CategoryRepository を生成するファクトリー
type CategoryRepositoryFactory func(userID uint) models.CategoryRepositoryInterface

// CategoryHandler はカテゴリハンドラーの構造体
type CategoryHandler struct {
	repoFactory CategoryRepositoryFactory
}

// NewCategoryHandler は CategoryHandler を作成する
func NewCategoryHandler(repoFactory CategoryRepositoryFactory) *CategoryHandler {
	return &CategoryHandler{
		repoFactory: repoFactory,
	}
}

// DefaultCategoryHandler はデフォルトの依存関係で CategoryHandler を作成する
func DefaultCategoryHandler() *CategoryHandler {
	return NewCategoryHandler(func(userID uint) models.CategoryRepositoryInterface {
		return models.NewCategoryRepository(userID)
	})
}

// ListCategories はカテゴリ一覧を取得する
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := h.repoFactory(userID)

	categoryType := c.Query("type") // income, expense, または空

	categories, err := repo.List(categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// CreateCategory はカテゴリを作成する
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := h.repoFactory(userID)

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &models.Category{
		Name: req.Name,
		Type: req.Type,
	}

	if err := repo.Create(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}
