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
		IsSettled:  false, // TODO: 精算テーブルを作成して状態を管理する
	})
}
