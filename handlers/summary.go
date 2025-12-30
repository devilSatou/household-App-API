package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

// GetMonthlySummary は月次集計を取得する
func GetMonthlySummary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	// デフォルトは今月
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y, err := strconv.Atoi(c.Query("year")); err == nil {
		year = y
	}
	if m, err := strconv.Atoi(c.Query("month")); err == nil && m >= 1 && m <= 12 {
		month = m
	}

	summary, err := repo.GetMonthlySummary(year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetCategorySummary はカテゴリ別集計を取得する
func GetCategorySummary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	var startDate, endDate time.Time

	// デフォルトは今月
	now := time.Now()
	startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)

	if sd := c.Query("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = t
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = t
		}
	}

	summary, err := repo.GetCategorySummary(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       summary,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})
}

// GetYearlySummary は年次集計を取得する
func GetYearlySummary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := models.NewTransactionRepository(userID)

	// デフォルトは今年
	year := time.Now().Year()
	if y, err := strconv.Atoi(c.Query("year")); err == nil {
		year = y
	}

	summary, err := repo.GetYearlySummary(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
