package handlers

import (
	"net/http"
	"receipt/server/config"
	"receipt/server/internal/models"
	"receipt/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register ユーザー登録
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// パスワードのハッシュ化
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Nickname:     input.Nickname,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

// Login ログイン
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// GetMe 現在のユーザー情報取得
func GetMe(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User record not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type UpdateMeInput struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

// UpdateMe ユーザー情報更新
func UpdateMe(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input UpdateMeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User record not found for update"})
		return
	}

	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Nickname != "" {
		user.Nickname = input.Nickname
	}
	if input.Password != "" {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.PasswordHash = hashedPassword
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
