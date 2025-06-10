package main

// ResponseMessage はAPIレスポンスの構造体です。
type ResponseMessage struct {
	Message string `json:"message"`
}

// Receipt はレシート情報を表す構造体です。
type Receipt struct {
	Date          string `json:"date"`          // 日付 (YYYY-MM-DD)
	Payer         string `json:"payer"`         // 支払者
	PaymentMethod string `json:"paymentMethod"` // 精算方法
}
