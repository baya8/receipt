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
	GroupID       uint      `json:"group_id" binding:"required"`
	Date          time.Time `json:"date" binding:"required"`
	Shop          string    `json:"shop"`
	Item          string    `json:"item"`
	Amount        int       `json:"amount" binding:"required"`
	PayerID       uint      `json:"payer_id" binding:"required"`
	PaymentMethod string    `json:"payment_method" binding:"required"`
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
	if err := config.DB.Where("group_id = ?", groupID).Order("date desc").Find(&receipts).Error; err != nil {
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

	receipt := models.Receipt{
		GroupID:       input.GroupID,
		UserID:        userID.(uint),
		Date:          input.Date,
		Shop:          input.Shop,
		Item:          input.Item,
		Amount:        input.Amount,
		PayerID:       input.PayerID,
		PaymentMethod: input.PaymentMethod,
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
	if err := config.DB.First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// UpdateReceipt レシート更新
func UpdateReceipt(c *gin.Context) {
	id := c.Param("id")
	var receipt models.Receipt
	if err := config.DB.First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	var input CreateReceiptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	receipt.Date = input.Date
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
	if err := config.DB.Delete(&models.Receipt{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete receipt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receipt deleted successfully"})
}
