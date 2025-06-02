package main

import (
	"github.com/baya8/receipt-server/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Ginルーターのデフォルトインスタンスを作成
	router := gin.Default()

	// HelloHandlerのインスタンスを作成
	helloHdlr := handlers.NewHelloHandler()

	// ルートとハンドラーメソッドを登録
	router.GET("/hello", helloHdlr.GetHello)

	// サーバーをポート8080で起動
	router.Run(":8080")
}
