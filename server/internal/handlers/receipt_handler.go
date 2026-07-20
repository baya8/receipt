package handlers

import (
	"fmt"
	"io"
	"net/http"
	"receipt/server/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateReceiptInput レシート作成・更新用入力
type CreateReceiptInput struct {
	GroupID         uuid.UUID `json:"group_id" binding:"required"`
	Date            time.Time `json:"date" binding:"required"`
	SettlementYear  int       `json:"settlement_year"`
	SettlementMonth int       `json:"settlement_month"`
	Shop            string    `json:"shop"`
	Item            string    `json:"item"`
	Amount          int       `json:"amount" binding:"required"`
	PayerID         uuid.UUID `json:"payer_id" binding:"required"`
	PaymentMethod   string    `json:"payment_method" binding:"required"`
}

// ReceiptHandler レシート関連ハンドラー
type ReceiptHandler struct {
	receiptService service.ReceiptService
	aiAnalyzer     service.AIAnalyzer
}

// NewReceiptHandler ReceiptHandlerを作成
func NewReceiptHandler(rs service.ReceiptService, ai service.AIAnalyzer) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: rs,
		aiAnalyzer:     ai,
	}
}

// GetReceipts レシート一覧取得
func (h *ReceiptHandler) GetReceipts(c *gin.Context) {
	groupIDStr := c.Query("group_id")
	if groupIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id format"})
		return
	}

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	var yearPtr *int
	var monthPtr *int
	if yearStr != "" && monthStr != "" {
		year, _ := strconv.Atoi(yearStr)
		month, _ := strconv.Atoi(monthStr)
		yearPtr = &year
		monthPtr = &month
	}

	receipts, err := h.receiptService.GetReceipts(groupID, yearPtr, monthPtr)
	if err != nil {
		respondInternalError(c, "Failed to fetch receipts")
		return
	}

	c.JSON(http.StatusOK, receipts)
}

// CreateReceipt レシート登録
func (h *ReceiptHandler) CreateReceipt(c *gin.Context) {
	var input CreateReceiptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	params := &service.CreateReceiptParams{
		GroupID:         input.GroupID,
		Date:            input.Date,
		SettlementYear:  input.SettlementYear,
		SettlementMonth: input.SettlementMonth,
		Shop:            input.Shop,
		Item:            input.Item,
		Amount:          input.Amount,
		PayerID:         input.PayerID,
		PaymentMethod:   input.PaymentMethod,
	}

	receipt, err := h.receiptService.CreateReceipt(params, userID)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to create receipt")
		}
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// GetReceipt レシート詳細取得
func (h *ReceiptHandler) GetReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt id"})
		return
	}

	receipt, err := h.receiptService.GetReceipt(id)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to get receipt")
		}
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// UpdateReceipt レシート更新
func (h *ReceiptHandler) UpdateReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt id"})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input CreateReceiptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := &service.CreateReceiptParams{
		GroupID:         input.GroupID,
		Date:            input.Date,
		SettlementYear:  input.SettlementYear,
		SettlementMonth: input.SettlementMonth,
		Shop:            input.Shop,
		Item:            input.Item,
		Amount:          input.Amount,
		PayerID:         input.PayerID,
		PaymentMethod:   input.PaymentMethod,
	}

	receipt, err := h.receiptService.UpdateReceipt(id, params, userID)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to update receipt")
		}
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// DeleteReceipt レシート削除
func (h *ReceiptHandler) DeleteReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt id"})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	if err := h.receiptService.DeleteReceipt(id, userID); err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to delete receipt")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receipt deleted successfully"})
}

// AnalyzeReceipt レシートAI解析
func (h *ReceiptHandler) AnalyzeReceipt(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		respondInternalError(c, "Failed to open image")
		return
	}
	defer src.Close()

	imgData, err := io.ReadAll(src)
	if err != nil {
		respondInternalError(c, "Failed to read image")
		return
	}

	result, err := h.aiAnalyzer.AnalyzeReceipt(c.Request.Context(), imgData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to analyze receipt: %v", err)})
		return
	}

	c.JSON(http.StatusOK, result)
}
