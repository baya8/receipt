package handlers

import (
	"fmt"
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
		Name:    input.Name,
		OwnerID: userID.(uint),
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

type InviteMemberInput struct {
	Email string `json:"email" binding:"required,email"`
}

// InviteMember メンバーを招待
func InviteMember(c *gin.Context) {
	groupID := c.Param("id")
	userID, _ := c.Get("userID")

	var input InviteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var group models.Group
	if err := config.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 権限チェック (管理者のみ招待可能)
	if group.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group owner can invite members"})
		return
	}

	var userToInvite models.User
	if err := config.DB.Where("email = ?", input.Email).First(&userToInvite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
		return
	}

	// すでにメンバーかチェック
	var count int64
	config.DB.Table("group_members").Where("group_id = ? AND user_id = ?", group.ID, userToInvite.ID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a member of this group"})
		return
	}

	if err := config.DB.Model(&group).Association("Members").Append(&userToInvite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invite member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member invited successfully"})
}

// RemoveMember メンバーを削除
func RemoveMember(c *gin.Context) {
	groupID := c.Param("id")
	memberID := c.Param("userId")
	userID, _ := c.Get("userID")

	var group models.Group
	if err := config.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 権限チェック (管理者のみ削除可能。ただし自分自身が脱退する場合は許可しても良いが、一旦要件に従う)
	if group.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group owner can remove members"})
		return
	}

	// オーナー自身は削除できない（別のオーナーを立てる機能が必要になるため）
	if memberID == fmt.Sprintf("%d", group.OwnerID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner cannot be removed from the group"})
		return
	}

	var userToRemove models.User
	if err := config.DB.First(&userToRemove, memberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := config.DB.Model(&group).Association("Members").Delete(&userToRemove); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
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
