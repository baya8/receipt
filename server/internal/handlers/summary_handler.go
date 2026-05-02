package handlers

import (
	"net/http"
	"receipt/server/config"
	"receipt/server/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type MemberSummary struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Paid     int    `json:"paid"`  // 実際に支払った合計
	Share    int    `json:"share"` // 負担すべき合計
}

type MonthlySummaryResponse struct {
	TotalSpent int             `json:"total_spent"`
	Members    []MemberSummary `json:"members"`
	IsSettled  bool            `json:"is_settled"`
}

type CreateSettlementInput struct {
	GroupID uint `json:"group_id" binding:"required"`
	Year    int  `json:"year" binding:"required"`
	Month   int  `json:"month" binding:"required"`
}

// GetMonthlySummary 月次サマリーの取得
func GetMonthlySummary(c *gin.Context) {
	groupIDStr := c.Query("group_id")
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	if groupIDStr == "" || yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id, year, and month are required"})
		return
	}

	groupID, _ := strconv.Atoi(groupIDStr)
	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	// 指定された月の開始日と終了日を計算
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	// グループメンバーの取得
	var group models.Group
	if err := config.DB.Preload("Members").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 精算済みかチェック
	var settlement models.Settlement
	isSettled := config.DB.Where("group_id = ? AND year = ? AND month = ?", groupID, year, month).First(&settlement).Error == nil

	// 指定期間のレシートを取得
	var receipts []models.Receipt
	config.DB.Where("group_id = ? AND date >= ? AND date < ?", groupID, startDate, endDate).Find(&receipts)

	// 集計用マップ
	paidMap := make(map[uint]int)
	shareMap := make(map[uint]int)
	totalSpent := 0

	for _, r := range receipts {
		totalSpent += r.Amount
		paidMap[r.PayerID] += r.Amount

		// 精算方法に応じた負担額の計算
		switch r.PaymentMethod {
		case "折半":
			// メンバー全員で均等割り（PoCでは2人を想定）
			sharePerPerson := r.Amount / len(group.Members)
			for _, m := range group.Members {
				shareMap[m.ID] += sharePerPerson
			}
		case "自分が10割負担":
			shareMap[r.PayerID] += r.Amount
		case "全額相手負担":
			// 支払者以外が負担（2人組を想定した簡易実装）
			for _, m := range group.Members {
				if m.ID != r.PayerID {
					shareMap[m.ID] += r.Amount
				}
			}
		}
	}

	// レスポンスの組み立て
	var memberSummaries []MemberSummary
	for _, m := range group.Members {
		memberSummaries = append(memberSummaries, MemberSummary{
			UserID:   m.ID,
			Nickname: m.Nickname,
			Paid:     paidMap[m.ID],
			Share:    shareMap[m.ID],
		})
	}

	c.JSON(http.StatusOK, MonthlySummaryResponse{
		TotalSpent: totalSpent,
		Members:    memberSummaries,
		IsSettled:  isSettled,
	})
}

// CreateSettlement 精算の確定
func CreateSettlement(c *gin.Context) {
	var input CreateSettlementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	// 既に精算済みかチェック
	var existing models.Settlement
	if config.DB.Where("group_id = ? AND year = ? AND month = ?", input.GroupID, input.Year, input.Month).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This month is already settled"})
		return
	}

	settlement := models.Settlement{
		GroupID:   input.GroupID,
		Year:      input.Year,
		Month:     input.Month,
		SettledBy: userID.(uint),
	}

	// トランザクションで処理
	tx := config.DB.Begin()

	if err := tx.Create(&settlement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settlement"})
		return
	}

	// 該当月のレシートをすべて精算済みに更新
	startDate := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)
	now := time.Now()

	if err := tx.Model(&models.Receipt{}).
		Where("group_id = ? AND date >= ? AND date < ? AND settled_at IS NULL", input.GroupID, startDate, endDate).
		Update("settled_at", now).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receipts"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, settlement)
}
