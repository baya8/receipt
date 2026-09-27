package utils

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

// InitJWTKey はJWT_SECRET環境変数から署名鍵を読み込む。起動時に一度だけ呼び出すこと。
// 未設定の場合は弱いデフォルト鍵で稼働し続けることを防ぐため、サーバーを起動させない。
func InitJWTKey() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable must be set")
	}
	jwtKey = []byte(secret)
}

// GetJWTKey InitJWTKeyで読み込んだ署名鍵を返す
func GetJWTKey() []byte {
	return jwtKey
}

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
	return token.SignedString(jwtKey)
}
