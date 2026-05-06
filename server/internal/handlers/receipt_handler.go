package handlers

import (
	"net/http"
	"receipt/server/config"
	"receipt/server/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateReceiptInput struct {
	GroupID         uint      `json:"group_id" binding:"required"`
	Date            time.Time `json:"date" binding:"required"`
	SettlementYear  int       `json:"settlement_year"`
	SettlementMonth int       `json:"settlement_month"`
	Shop            string    `json:"shop"`
	Item            string    `json:"item"`
	Amount          int       `json:"amount" binding:"required"`
	PayerID         uint      `json:"payer_id" binding:"required"`
	PaymentMethod   string    `json:"payment_method" binding:"required"`
}

// GetReceipts レシート一覧取得
func GetReceipts(c *gin.Context) {
	groupIDStr := c.Query("group_id")
	if groupIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}
	groupID, _ := strconv.Atoi(groupIDStr)

	var receipts []models.Receipt
	if err := config.DB.Preload("Payer").Where("group_id = ?", groupID).Order("date desc").Find(&receipts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch receipts"})
		return
	}

	c.JSON(http.StatusOK, receipts)
}

// CreateReceipt レシート登録
func CreateReceipt(c *gin.Context) {
	var input CreateReceiptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	// 精算月が指定されていない場合は購入日から設定
	settlementYear := input.SettlementYear
	settlementMonth := input.SettlementMonth
	if settlementYear == 0 || settlementMonth == 0 {
		settlementYear = input.Date.Year()
		settlementMonth = int(input.Date.Month())
	}

	receipt := models.Receipt{
		GroupID:         input.GroupID,
		UserID:          userID.(uint),
		Date:            input.Date,
		SettlementYear:  settlementYear,
		SettlementMonth: settlementMonth,
		Shop:            input.Shop,
		Item:            input.Item,
		Amount:          input.Amount,
		PayerID:         input.PayerID,
		PaymentMethod:   input.PaymentMethod,
	}

	if err := config.DB.Create(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create receipt"})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// GetReceipt レシート詳細取得
func GetReceipt(c *gin.Context) {
	id := c.Param("id")
	var receipt models.Receipt
	if err := config.DB.Preload("Payer").First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// UpdateReceipt レシート更新
func UpdateReceipt(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var receipt models.Receipt
	if err := config.DB.First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	// 登録者本人かチェック
	if receipt.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator can update this receipt"})
		return
	}

	var input CreateReceiptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 精算月が指定されていない場合は購入日から設定
	settlementYear := input.SettlementYear
	settlementMonth := input.SettlementMonth
	if settlementYear == 0 || settlementMonth == 0 {
		settlementYear = input.Date.Year()
		settlementMonth = int(input.Date.Month())
	}

	receipt.Date = input.Date
	receipt.SettlementYear = settlementYear
	receipt.SettlementMonth = settlementMonth
	receipt.Shop = input.Shop
	receipt.Item = input.Item
	receipt.Amount = input.Amount
	receipt.PayerID = input.PayerID
	receipt.PaymentMethod = input.PaymentMethod

	if err := config.DB.Save(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receipt"})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// DeleteReceipt レシート削除
func DeleteReceipt(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var receipt models.Receipt
	if err := config.DB.First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	// 登録者本人かチェック
	if receipt.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator can delete this receipt"})
		return
	}

	if err := config.DB.Delete(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete receipt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receipt deleted successfully"})
}
