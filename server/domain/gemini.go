package domain

import (
	"context"
	"io"
)

// ExtractedReceiptInfo はGemini APIによって画像から抽出されたレシート情報を表します。
type ExtractedReceiptInfo struct {
	Date        string // 購入日付 (YYYY-MM-DD)
	Store       string // 店名
	Items       string // 品目 (カンマ区切りなど)
	TotalAmount int    // 金額
}

// GeminiService はレシート画像から情報を抽出するサービスです。
type GeminiService interface {
	ExtractReceiptInfo(ctx context.Context, imageFile io.Reader) (*ExtractedReceiptInfo, error)
}
