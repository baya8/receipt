package handlers

import (
	"net/http"

	"github.com/baya8/receipt-server/message"
	"github.com/gin-gonic/gin"
)

type HelloHandler struct {
	// 将来的にデータベース接続などの依存関係をここに追加できます。
}

// NewHelloHandler は HelloHandler の新しいインスタンスを生成します。
func NewHelloHandler() *HelloHandler {
	return &HelloHandler{}
}

// GetHello は GET /hello リクエストを処理し、挨拶メッセージを返します。
func (h *HelloHandler) GetHello(c *gin.Context) {
	response := message.Message{Text: "こんにちは、Ginフレームワークを使ったAPIです"}
	c.JSON(http.StatusOK, response)
}
