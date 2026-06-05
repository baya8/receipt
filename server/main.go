package main

import (
	"os"
	"strings"

	"receipt/server/config"
	"receipt/server/internal/handlers"
	"receipt/server/internal/middleware"
	"receipt/server/internal/repository"
	"receipt/server/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルがある場合は読み込む（ローカル開発用）
	godotenv.Load()

	// データベース初期化
	config.InitDB()

	// 依存関係の初期化 (DI)
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	groupRepo := repository.NewGroupRepository(config.DB)
	groupService := service.NewGroupService(groupRepo, userRepo)
	groupHandler := handlers.NewGroupHandler(groupService)

	receiptRepo := repository.NewReceiptRepository(config.DB)
	receiptService := service.NewReceiptService(receiptRepo)
	aiAnalyzer := service.NewAIAnalyzer(os.Getenv("GOOGLE_API_KEY"))
	receiptHandler := handlers.NewReceiptHandler(receiptService, aiAnalyzer)

	settlementRepo := repository.NewSettlementRepository(config.DB)
	summaryService := service.NewSummaryService(groupRepo, receiptRepo, settlementRepo)
	summaryHandler := handlers.NewSummaryHandler(summaryService)

	r := gin.Default()

	// CORS設定
	allowedOrigins := []string{"http://localhost:3000"}
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 認証関連
	auth := r.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(), userHandler.GetMe)
		auth.PUT("/me", middleware.AuthMiddleware(), userHandler.UpdateMe)
	}

	// レシート関連（認証必須）
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/receipts", receiptHandler.GetReceipts)
		api.POST("/receipts", receiptHandler.CreateReceipt)
		api.GET("/receipts/:id", receiptHandler.GetReceipt)
		api.PUT("/receipts/:id", receiptHandler.UpdateReceipt)
		api.DELETE("/receipts/:id", receiptHandler.DeleteReceipt)
		api.POST("/receipts/analyze", receiptHandler.AnalyzeReceipt)

		api.GET("/groups", groupHandler.GetMyGroups)
		api.POST("/groups", groupHandler.CreateGroup)
		api.PUT("/groups/:id", groupHandler.UpdateGroup)
		api.DELETE("/groups/:id", groupHandler.DeleteGroup)
		api.POST("/groups/:id/invite", groupHandler.InviteMember)
		api.DELETE("/groups/:id/members/:userId", groupHandler.RemoveMember)

		api.GET("/summary", summaryHandler.GetMonthlySummary)
		api.POST("/settle", summaryHandler.CreateSettlement)
	}

	// ヘルスチェック
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8080")
}
