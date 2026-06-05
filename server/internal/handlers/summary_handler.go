package handlers

import (
	"net/http"
	"receipt/server/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateSettlementInput 精算登録用入力
type CreateSettlementInput struct {
	GroupID uuid.UUID `json:"group_id" binding:"required"`
	Year    int       `json:"year" binding:"required"`
	Month   int       `json:"month" binding:"required"`
	Amount  int       `json:"amount" binding:"required"`
}

// SummaryHandler 集計・精算関連ハンドラー
type SummaryHandler struct {
	summaryService service.SummaryService
}

// NewSummaryHandler SummaryHandlerを作成
func NewSummaryHandler(ss service.SummaryService) *SummaryHandler {
	return &SummaryHandler{summaryService: ss}
}

// GetMonthlySummary 月次サマリーの取得
func (h *SummaryHandler) GetMonthlySummary(c *gin.Context) {
	groupIDStr := c.Query("group_id")
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	if groupIDStr == "" || yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id, year, and month are required"})
		return
	}

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id format"})
		return
	}
	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	result, err := h.summaryService.GetMonthlySummary(groupID, year, month)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to get monthly summary")
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateSettlement 精算の記録
func (h *SummaryHandler) CreateSettlement(c *gin.Context) {
	var input CreateSettlementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	settlement, err := h.summaryService.CreateSettlement(input.GroupID, input.Year, input.Month, input.Amount, userID)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to create settlement")
		}
		return
	}

	c.JSON(http.StatusOK, settlement)
}
