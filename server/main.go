package main

import (
	"receipt/server/config"
	"receipt/server/internal/handlers"
	"receipt/server/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルがある場合は読み込む（ローカル開発用）
	godotenv.Load()

	// データベース初期化
	config.InitDB()

	r := gin.Default()

	// CORS設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://192.168.50.18:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 認証関連
	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.GET("/me", middleware.AuthMiddleware(), handlers.GetMe)
		auth.PUT("/me", middleware.AuthMiddleware(), handlers.UpdateMe)
	}

	// レシート関連（認証必須）
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/receipts", handlers.GetReceipts)
		api.POST("/receipts", handlers.CreateReceipt)
		api.GET("/receipts/:id", handlers.GetReceipt)
		api.PUT("/receipts/:id", handlers.UpdateReceipt)
		api.DELETE("/receipts/:id", handlers.DeleteReceipt)
		api.POST("/receipts/analyze", handlers.AnalyzeReceipt)

		api.GET("/groups", handlers.GetMyGroups)
		api.POST("/groups", handlers.CreateGroup)
		api.PUT("/groups/:id", handlers.UpdateGroup)
		api.DELETE("/groups/:id", handlers.DeleteGroup)
		api.POST("/groups/:id/invite", handlers.InviteMember)
		api.DELETE("/groups/:id/members/:userId", handlers.RemoveMember)

		api.GET("/summary", handlers.GetMonthlySummary)
		api.POST("/settle", handlers.CreateSettlement)
	}

	// ヘルスチェック
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8080")
}
