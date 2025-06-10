package presentation

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"receipt/application"
)

// ReceiptHandler はレシート関連のHTTPリクエストを処理します。
type ReceiptHandler struct {
	receiptService *application.ReceiptService
}

// NewReceiptHandler は新しいReceiptHandlerを生成します。
func NewReceiptHandler(service *application.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: service,
	}
}

// CreateReceiptHandler はPOST /receipts リクエストを処理します。
func (h *ReceiptHandler) CreateReceiptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	// multipart/form-data のパース
	// 最大32MBのメモリをファイル処理に使用 (残りはディスクへ)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("プレゼンテーション層: リクエストのパースに失敗しました: %v", err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	// フォームデータの取得
	dateStr := r.FormValue("date") // ユーザーが任意で指定する日付
	payer := r.FormValue("payer")
	paymentMethod := r.FormValue("paymentMethod")

	// 画像ファイルの取得
	file, header, err := r.FormFile("receiptImage")
	if err != nil {
		log.Printf("プレゼンテーション層: 画像ファイルの取得に失敗しました: %v", err)
		http.Error(w, "画像ファイルの取得に失敗しました", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx := r.Context() // リクエストコンテキストを取得

	// imageFile (multipart.File) は io.ReadSeeker を実装しているため、
	// GCSへのアップロードとGemini APIへの送信で複数回読み取れるように先頭に戻す。
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Printf("プレゼンテーション層: ファイルのシークに失敗しました: %v", err)
		http.Error(w, "ファイルの処理に失敗しました", http.StatusInternalServerError)
		return
	}

	// アプリケーションサービスに処理を委譲
	createdReceipt, err := h.receiptService.CreateReceipt(ctx, dateStr, payer, paymentMethod, file, header.Filename)
	if err != nil {
		log.Printf("プレゼンテーション層: アプリケーションサービスの処理中にエラー: %v", err)
		// アプリケーションサービスからのエラーに応じて適切なHTTPステータスを返す
		// TODO: エラーの種類に応じてステータスコードを細かく制御する
		http.Error(w, fmt.Sprintf("レシート作成エラー: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	if err := json.NewEncoder(w).Encode(createdReceipt); err != nil {
		log.Printf("プレゼンテーション層: レスポンスのエンコードに失敗しました: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
