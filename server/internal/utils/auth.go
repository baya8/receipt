package utils

import (
	"time"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var JwtKey = []byte("your_secret_key") // 本番環境では環境変数から読み込む

// HashPassword パスワードをハッシュ化する
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash パスワードとハッシュを比較する
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken JWTトークンを生成する
func GenerateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24時間有効
	})
	return token.SignedString(JwtKey)
}
