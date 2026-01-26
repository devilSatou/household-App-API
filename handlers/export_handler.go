package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"household-app/middleware"
	"household-app/models"
)

// ExportHandler はエクスポートハンドラーの構造体
type ExportHandler struct {
	repoFactory TransactionRepositoryFactory
}

// NewExportHandler は ExportHandler を作成する
func NewExportHandler(repoFactory TransactionRepositoryFactory) *ExportHandler {
	return &ExportHandler{
		repoFactory: repoFactory,
	}
}

// DefaultExportHandler はデフォルトの依存関係で ExportHandler を作成する
func DefaultExportHandler() *ExportHandler {
	return NewExportHandler(func(userID uint) models.TransactionRepositoryInterface {
		return models.NewTransactionRepository(userID)
	})
}

// ExportCSV は収支記録をCSVでエクスポートする
func (h *ExportHandler) ExportCSV(c *gin.Context) {
	userID := middleware.GetUserID(c)
	repo := h.repoFactory(userID)

	// パラメータ
	params := models.ListParams{
		Limit: 0, // 無制限
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
	if t := c.Query("type"); t != "" {
		params.Type = t
	}

	transactions, _, err := repo.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ファイル名
	filename := fmt.Sprintf("transactions_%s.csv", time.Now().Format("20060102"))

	// CSVヘッダー
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// BOM (UTF-8 with BOM for Excel)
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// ヘッダー行
	writer.Write([]string{"ID", "日付", "種別", "カテゴリ", "金額", "メモ", "作成日時"})

	// データ行
	for _, t := range transactions {
		typeLabel := "支出"
		if t.Type == "income" {
			typeLabel = "収入"
		}

		categoryName := ""
		if t.Category != nil {
			categoryName = t.Category.Name
		}

		writer.Write([]string{
			fmt.Sprintf("%d", t.ID),
			t.Date.Format("2006-01-02"),
			typeLabel,
			categoryName,
			fmt.Sprintf("%d", t.Amount),
			t.Memo,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}
