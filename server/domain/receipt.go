package domain

import "time"

// Receipt はレシート情報を表す構造体です。
type Receipt struct {
	ID            string    `json:"id"`
	Date          string    `json:"date"`
	Store         string    `json:"store"`
	Items         string    `json:"items"`
	TotalAmount   int       `json:"totalAmount"`
	Payer         string    `json:"payer"`
	PaymentMethod string    `json:"paymentMethod"`
	ImageURL      string    `json:"imageUrl"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// NewReceipt は新しいReceiptドメインオブジェクトを生成します。
// 必須の初期値を受け取り、IDやタイムスタンプはリポジトリで設定することを想定。
func NewReceipt(date, store, items string, totalAmount int, payer, paymentMethod, imageURL string) *Receipt {
	return &Receipt{
		Date:          date,
		Store:         store,
		Items:         items,
		TotalAmount:   totalAmount,
		Payer:         payer,
		PaymentMethod: paymentMethod,
		ImageURL:      imageURL,
	}
}
