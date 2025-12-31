package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

// CategoryRequest はカテゴリのリクエストボディ
type CategoryRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=income expense"`
}

// ListCategories はカテゴリ一覧を取得する
func ListCategories(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewCategoryRepository(userID)

	categoryType := c.Query("type") // income, expense, または空

	categories, err := repo.List(categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// CreateCategory はカテゴリを作成する
func CreateCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewCategoryRepository(userID)

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
