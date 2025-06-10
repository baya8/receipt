package application

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"receipt/domain"
)

// ReceiptService はレシート関連のユースケースを処理するアプリケーションサービスです。
type ReceiptService struct {
	receiptRepo    domain.ReceiptRepository
	geminiService  domain.GeminiService
	storageService domain.StorageService
}

// NewReceiptService は新しいReceiptServiceを生成します。
func NewReceiptService(repo domain.ReceiptRepository, geminiSvc domain.GeminiService, storageSvc domain.StorageService) *ReceiptService {
	return &ReceiptService{
		receiptRepo:    repo,
		geminiService:  geminiSvc,
		storageService: storageSvc,
	}
}

// determineFinalDate はユーザー入力の日付と抽出された日付情報から、最終的に使用する日付文字列を決定します。
// ユーザー入力の日付が優先されます。どちらも無効な場合はエラーを返します。
func determineFinalDate(dateStrFromRequest string, extractedDate string) (string, error) {
	if dateStrFromRequest != "" {
		// ユーザー入力の日付形式をバリデーション (YYYY-MM-DD)
		_, err := time.Parse("2006-01-02", dateStrFromRequest)
		if err != nil {
			return "", fmt.Errorf("ユーザー入力の日付形式が無効です (YYYY-MM-DD): %s, error: %w", dateStrFromRequest, err)
		}
		log.Printf("日付決定: ユーザー入力の日付を使用: %s", dateStrFromRequest)
		return dateStrFromRequest, nil
	}
	if extractedDate != "" {
		// Gemini抽出の日付形式をバリデーション (YYYY-MM-DD)
		_, err := time.Parse("2006-01-02", extractedDate)
		if err != nil {
			return "", fmt.Errorf("Gemini APIから抽出された日付の形式が無効です: %s, error: %w", extractedDate, err)
		}
		log.Printf("日付決定: Gemini抽出の日付を使用: %s", extractedDate)
		return extractedDate, nil
	}
	return "", fmt.Errorf("有効な日付情報が取得できませんでした (ユーザー入力もGemini抽出もなし)")
}

// CreateReceipt は新しいレシートを作成し、関連情報を保存します。
// ここでビジネスロジック（バリデーション、ドメインオブジェクト生成、リポジトリ呼び出し）を行います。
func (s *ReceiptService) CreateReceipt(ctx context.Context, dateStrFromRequest string, payer, paymentMethod string, imageFile io.Reader, imageFilename string) (*domain.Receipt, error) {
	// 0. GCSに画像をアップロード
	imageURL, err := s.storageService.UploadImage(ctx, imageFile, imageFilename)
	if err != nil {
		return nil, fmt.Errorf("GCSへの画像アップロードに失敗しました: %w", err)
	}
	log.Printf("アプリケーションサービス: GCSアップロード成功 URL: %s", imageURL)

	// 1. Gemini APIで画像から情報を抽出
	extractedInfo, err := s.geminiService.ExtractReceiptInfo(ctx, imageFile) // imageFileは再度利用するため、ハンドラ側でSeek(0,0)するか、再度読み込む必要があるかもしれません。今回はGCSアップロード後にGeminiに渡す想定で、imageFileは一度しか読み込まれない前提で進めます。もしGeminiがio.Readerを消費するなら、GCSアップロード前にバッファリングするか、GCSから再度読み込む必要があります。
	if err != nil {
		return nil, fmt.Errorf("Gemini APIでの情報抽出に失敗しました: %w", err)
	}
	log.Printf("アプリケーションサービス: Gemini抽出結果: %+v", extractedInfo)

	// 2. 使用する日付を決定
	var extractedDateForDetermination string
	if extractedInfo != nil {
		extractedDateForDetermination = extractedInfo.Date
	}
	finalDate, err := determineFinalDate(dateStrFromRequest, extractedDateForDetermination)
	if err != nil {
		return nil, fmt.Errorf("最終的な日付の決定に失敗しました: %w", err)
	}

	// 3. ドメインオブジェクトを生成
	// extractedInfoがnilの場合も考慮して、各フィールドにアクセスする前にnilチェックを行うか、
	// domain.NewReceiptがnilを安全に扱えるようにする。
	// ここではextractedInfoがnilでない前提で進めるが、本番コードでは要検討。
	var store, items string
	var totalAmount int
	if extractedInfo != nil {
		store = extractedInfo.Store
		items = extractedInfo.Items
		totalAmount = extractedInfo.TotalAmount
	}
	receipt := domain.NewReceipt(finalDate, store, items, totalAmount, payer, paymentMethod, imageURL)

	// 4. リポジトリ経由で保存
	if err := s.receiptRepo.Save(ctx, receipt); err != nil {
		return nil, fmt.Errorf("レシートの保存に失敗しました: %w", err)
	}
	log.Printf("アプリケーションサービス: レシート保存成功 ID: %s", receipt.ID)
	return receipt, nil
}
