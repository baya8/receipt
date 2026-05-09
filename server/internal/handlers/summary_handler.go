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
	TotalSpent  int                 `json:"total_spent"`
	Members     []MemberSummary     `json:"members"`
	Settlements []models.Settlement `json:"settlements"`
}

type CreateSettlementInput struct {
	GroupID uint `json:"group_id" binding:"required"`
	Year    int  `json:"year" binding:"required"`
	Month   int  `json:"month" binding:"required"`
	Amount  int  `json:"amount" binding:"required"`
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

	// グループメンバーの取得
	var group models.Group
	if err := config.DB.Preload("Members").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 精算履歴の取得
	var settlements []models.Settlement
	config.DB.Where("group_id = ? AND year = ? AND month = ?", groupID, year, month).
		Preload("SettledByUser").
		Order("created_at desc").
		Find(&settlements)

	// 指定された精算月のレシートを取得
	var receipts []models.Receipt
	config.DB.Where("group_id = ? AND settlement_year = ? AND settlement_month = ?", groupID, year, month).Find(&receipts)

	// 集計用マップ
	paidMap := make(map[uint]int)
	shareMap := make(map[uint]int)
	totalSpent := 0

	for _, r := range receipts {
		totalSpent += r.Amount
		paidMap[r.PayerID] += r.Amount

		// 精算方法に応じた負担額の計算
		switch r.PaymentMethod {
		case "自分が10割負担", "自分が10割":
			// 支払者が全額負担
			shareMap[r.PayerID] += r.Amount

		case "全額相手負担":
			// 支払者以外で均等に負担（2人以上の場合に対応）
			otherCount := len(group.Members) - 1
			if otherCount > 0 {
				sharePerPerson := r.Amount / otherCount
				remainder := r.Amount % otherCount
				
				isRemainderAssigned := false
				for _, m := range group.Members {
					if m.ID != r.PayerID {
						shareMap[m.ID] += sharePerPerson
						if !isRemainderAssigned {
							shareMap[m.ID] += remainder
							isRemainderAssigned = true
						}
					}
				}
			} else {
				// 相手がいない場合は支払者が負担
				shareMap[r.PayerID] += r.Amount
			}

		case "折半":
			fallthrough
		default:
			// メンバー全員で均等割り
			numMembers := len(group.Members)
			if numMembers > 0 {
				sharePerPerson := r.Amount / numMembers
				remainder := r.Amount % numMembers
				
				for _, m := range group.Members {
					shareMap[m.ID] += sharePerPerson
				}
				// 端数は支払者が負担（または誰か一人に寄せる）
				shareMap[r.PayerID] += remainder
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
		TotalSpent:  totalSpent,
		Members:     memberSummaries,
		Settlements: settlements,
	})
}

// CreateSettlement 精算の記録
func CreateSettlement(c *gin.Context) {
	var input CreateSettlementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "精算金額は1円以上にしてください"})
		return
	}

	userID, _ := c.Get("userID")

	// 今回はフロントエンドでの制限を主とし、バックエンドでは単純な記録を行う
	// (実運用ではより厳密な残高チェックが必要)

	settlement := models.Settlement{
		GroupID:   input.GroupID,
		Year:      input.Year,
		Month:     input.Month,
		Amount:    input.Amount,
		SettledBy: userID.(uint),
	}

	// トランザクションで処理
	tx := config.DB.Begin()

	if err := tx.Create(&settlement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settlement"})
		return
	}

	// 該当精算月のレシートをすべて精算済みに更新（部分精算でも一旦「精算アクションがあった」としてマーク）
	now := time.Now()
	if err := tx.Model(&models.Receipt{}).
		Where("group_id = ? AND settlement_year = ? AND settlement_month = ? AND settled_at IS NULL", input.GroupID, input.Year, input.Month).
		Update("settled_at", now).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receipts"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, settlement)
}
