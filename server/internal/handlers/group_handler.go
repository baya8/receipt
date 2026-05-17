package handlers

import (
	"net/http"
	"receipt/server/config"
	"receipt/server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateGroupInput struct {
	Name string `json:"name" binding:"required"`
}

type UpdateGroupInput struct {
	Name string `json:"name" binding:"required"`
}

// CreateGroup グループ作成
func CreateGroup(c *gin.Context) {
	var input CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	// グループ作成
	group := models.Group{
		Name:    input.Name,
		OwnerID: userID,
	}

	if err := config.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// 作成者をメンバーとして追加
	var user models.User
	config.DB.First(&user, "id = ?", userID)
	config.DB.Model(&group).Association("Members").Append(&user)

	c.JSON(http.StatusOK, group)
}

type InviteMemberInput struct {
	Email string `json:"email" binding:"required,email"`
}

// InviteMember メンバーを招待
func InviteMember(c *gin.Context) {
	groupID := c.Param("id")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input InviteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 権限チェック (管理者のみ招待可能)
	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group owner can invite members"})
		return
	}

	var userToInvite models.User
	if err := config.DB.Where("email = ?", input.Email).First(&userToInvite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "招待相手のユーザーが見つかりません。先に相手の方がアカウント登録を完了しているか確認してください。"})
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
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 権限チェック (管理者のみ削除可能。ただし自分自身が脱退する場合は許可しても良いが、一旦要件に従う)
	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group owner can remove members"})
		return
	}

	// オーナー自身は削除できない（別のオーナーを立てる機能が必要になるため）
	if memberID == group.OwnerID.String() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner cannot be removed from the group"})
		return
	}

	var userToRemove models.User
	if err := config.DB.First(&userToRemove, "id = ?", memberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User to remove not found"})
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
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)
	
	var groups []models.Group
	// 自分がメンバーに含まれるグループを取得
	err := config.DB.Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Preload("Members").
		Find(&groups).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// UpdateGroup グループ情報更新 (名前変更)
func UpdateGroup(c *gin.Context) {
	groupID := c.Param("id")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input UpdateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 権限チェック (管理者のみ変更可能)
	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group owner can update group info"})
		return
	}

	group.Name = input.Name
	if err := config.DB.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup グループ削除
func DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// 権限チェック (管理者のみ削除可能)
	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group owner can delete group"})
		return
	}

	// トランザクションで関連データを含めて削除
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. レシートを削除 (物理削除または論理削除はモデルの定義に従う)
		if err := tx.Where("group_id = ?", group.ID).Delete(&models.Receipt{}).Error; err != nil {
			return err
		}

		// 2. 精算履歴を削除
		if err := tx.Where("group_id = ?", group.ID).Delete(&models.Settlement{}).Error; err != nil {
			return err
		}

		// 3. メンバーとの紐付けを解除
		if err := tx.Model(&group).Association("Members").Clear(); err != nil {
			return err
		}

		// 4. グループ自体を削除
		if err := tx.Delete(&group).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}
