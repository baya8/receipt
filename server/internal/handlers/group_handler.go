package handlers

import (
	"net/http"
	"receipt/server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateGroupInput グループ作成用入力
type CreateGroupInput struct {
	Name string `json:"name" binding:"required"`
}

// UpdateGroupInput グループ情報更新用入力
type UpdateGroupInput struct {
	Name string `json:"name" binding:"required"`
}

// InviteMemberInput メンバー招待用入力
type InviteMemberInput struct {
	Email string `json:"email" binding:"required,email"`
}

// GroupHandler グループ関連ハンドラー
type GroupHandler struct {
	groupService service.GroupService
}

// NewGroupHandler GroupHandlerを作成
func NewGroupHandler(gs service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: gs}
}

// CreateGroup グループ作成
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var input CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	group, err := h.groupService.CreateGroup(input.Name, userID)
	if err != nil {
		respondInternalError(c, "Failed to create group")
		return
	}

	c.JSON(http.StatusOK, group)
}

// InviteMember メンバーを招待
func (h *GroupHandler) InviteMember(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id format"})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input InviteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupService.InviteMember(groupID, userID, input.Email); err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to invite member")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member invited successfully"})
}

// RemoveMember メンバーを削除
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id format"})
		return
	}

	memberIDStr := c.Param("userId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member user_id format"})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	if err := h.groupService.RemoveMember(groupID, userID, memberID); err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to remove member")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// GetMyGroups 自分が所属するグループ一覧取得
func (h *GroupHandler) GetMyGroups(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	groups, err := h.groupService.GetMyGroups(userID)
	if err != nil {
		respondInternalError(c, "Failed to fetch groups")
		return
	}

	c.JSON(http.StatusOK, groups)
}

// UpdateGroup グループ情報更新 (名前変更)
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id format"})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input UpdateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.groupService.UpdateGroup(groupID, userID, input.Name)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to update group")
		}
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup グループ削除
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id format"})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	if err := h.groupService.DeleteGroup(groupID, userID); err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to delete group")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}
