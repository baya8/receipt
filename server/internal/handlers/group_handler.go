package handlers

import (
	"net/http"
	"receipt/server/config"
	"receipt/server/internal/models"

	"github.com/gin-gonic/gin"
)

type CreateGroupInput struct {
	Name string `json:"name" binding:"required"`
}

// CreateGroup グループ作成
func CreateGroup(c *gin.Context) {
	var input CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	// グループ作成
	group := models.Group{
		Name: input.Name,
	}

	if err := config.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// 作成者をメンバーとして追加
	var user models.User
	config.DB.First(&user, userID)
	config.DB.Model(&group).Association("Members").Append(&user)

	c.JSON(http.StatusOK, group)
}

// GetMyGroups 自分が所属するグループ一覧取得
func GetMyGroups(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var user models.User
	if err := config.DB.Preload("Members").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var groups []models.Group
	// 自分がメンバーに含まれるグループを取得
	config.DB.Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Preload("Members").
		Find(&groups)

	c.JSON(http.StatusOK, groups)
}
