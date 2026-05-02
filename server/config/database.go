package config

import (
	"fmt"
	"os"
	"receipt/server/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := "3306"
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// オートマイグレーション
	err = db.AutoMigrate(&models.User{}, &models.Group{}, &models.Receipt{}, &models.Settlement{})
	if err != nil {
		panic("failed to migrate database")
	}

	DB = db
}
