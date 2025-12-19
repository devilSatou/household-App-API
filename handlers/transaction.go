package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

// TransactionRequest は収支記録のリクエストボディ
type TransactionRequest struct {
	CategoryID uint   `json:"category_id" binding:"required"`
	Amount     int    `json:"amount" binding:"required,gt=0"`
	Type       string `json:"type" binding:"required,oneof=income expense"`
	Date       string `json:"date" binding:"required"` // YYYY-MM-DD
	Memo       string `json:"memo"`
}

// ListTransactions は収支記録一覧を取得する
func ListTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	// クエリパラメータ
	params := models.ListParams{
		Limit:  20,
		Offset: 0,
	}

	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}
	if offset, err := strconv.Atoi(c.Query("offset")); err == nil && offset >= 0 {
		params.Offset = offset
	}
	if t := c.Query("type"); t != "" {
		params.Type = t
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			params.StartDate = t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			params.EndDate = t
		}
	}

	transactions, total, err := repo.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   transactions,
		"total":  total,
		"limit":  params.Limit,
		"offset": params.Offset,
	})
}

// CreateTransaction は収支記録を作成する
func CreateTransaction(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	transaction := &models.Transaction{
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Type:       req.Type,
		Date:       date,
		Memo:       req.Memo,
	}

	if err := repo.Create(transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// GetTransaction は収支記録の詳細を取得する
func GetTransaction(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	transaction, err := repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// UpdateTransaction は収支記録を更新する
func UpdateTransaction(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	transaction, err := repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	transaction.CategoryID = req.CategoryID
	transaction.Amount = req.Amount
	transaction.Type = req.Type
	transaction.Date = date
	transaction.Memo = req.Memo

	if err := repo.Update(transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// DeleteTransaction は収支記録を削除する
func DeleteTransaction(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted"})
}
