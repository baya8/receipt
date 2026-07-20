package handlers

import (
	"net/http"
	"receipt/server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterInput ユーザー登録用入力
type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"required"`
}

// LoginInput ログイン用入力
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateMeInput ユーザー情報更新用入力
type UpdateMeInput struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

// UserHandler ユーザー関連ハンドラー
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler UserHandlerを作成
func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

// Register ユーザー登録
func (h *UserHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Register(input.Email, input.Password, input.Nickname); err != nil {
		respondInternalError(c, "Could not create user")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

// Login ログイン
func (h *UserHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.userService.Login(input.Email, input.Password)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to login")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// GetMe 現在のユーザー情報取得
func (h *UserHandler) GetMe(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	user, err := h.userService.GetMe(userID)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to get user info")
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateMe ユーザー情報更新
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input UpdateMeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateMe(userID, input.Email, input.Nickname, input.Password)
	if err != nil {
		if !respondWithServiceError(c, err) {
			respondInternalError(c, "Failed to update user")
		}
		return
	}

	c.JSON(http.StatusOK, user)
}
