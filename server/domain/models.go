package domain

import (
	"context"
)

// ResponseMessage はAPIレスポンスの構造体です。
type ResponseMessage struct {
	Message string `json:"message"`
}

// ReceiptRepository はレシートの永続化を扱うインターフェースです。
type ReceiptRepository interface {
	Save(ctx context.Context, receipt *Receipt) error
	// FindByID(ctx context.Context, id string) (*Receipt, error) // 必要に応じて追加
	// ... 他のメソッド
}
