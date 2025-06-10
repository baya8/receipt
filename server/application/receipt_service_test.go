package application

import (
	"context"
	"errors"
	"io"
	"receipt/domain"
	"strings"
	"testing"
)

func TestDetermineFinalDate(t *testing.T) {
	tests := []struct {
		name                 string
		dateStrFromRequest   string
		extractedDate        string
		expectedDate         string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:               "ユーザー入力の日付が有効",
			dateStrFromRequest: "2023-01-15",
			extractedDate:      "2023-01-20", // ユーザー入力が優先される
			expectedDate:       "2023-01-15",
			expectError:        false,
		},
		{
			name:                 "ユーザー入力の日付が無効な形式",
			dateStrFromRequest:   "2023/01/15",
			extractedDate:        "2023-01-20",
			expectError:          true,
			expectedErrorMessage: "ユーザー入力の日付形式が無効です (YYYY-MM-DD): 2023/01/15",
		},
		{
			name:               "ユーザー入力なし、抽出された日付が有効",
			dateStrFromRequest: "",
			extractedDate:      "2023-02-10",
			expectedDate:       "2023-02-10",
			expectError:        false,
		},
		{
			name:                 "ユーザー入力なし、抽出された日付が無効な形式",
			dateStrFromRequest:   "",
			extractedDate:        "invalid-date",
			expectError:          true,
			expectedErrorMessage: "Gemini APIから抽出された日付の形式が無効です: invalid-date",
		},
		{
			name:                 "どちらの日付もない",
			dateStrFromRequest:   "",
			extractedDate:        "",
			expectError:          true,
			expectedErrorMessage: "有効な日付情報が取得できませんでした (ユーザー入力もGemini抽出もなし)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualDate, err := determineFinalDate(tt.dateStrFromRequest, tt.extractedDate)

			if tt.expectError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、nilでした")
				} else if err.Error() != tt.expectedErrorMessage {
					t.Errorf("期待されるエラーメッセージ '%s' でしたが、'%s' でした", tt.expectedErrorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("エラーは期待されていませんでしたが、発生しました: %v", err)
				}
				if actualDate != tt.expectedDate {
					t.Errorf("期待される日付 '%s' でしたが、'%s' でした", tt.expectedDate, actualDate)
				}
			}
		})
	}
}

// --- Mocks ---

type mockStorageService struct {
	UploadImageFunc func(ctx context.Context, file io.Reader, objectName string) (string, error)
}

func (m *mockStorageService) UploadImage(ctx context.Context, file io.Reader, objectName string) (string, error) {
	if m.UploadImageFunc != nil {
		return m.UploadImageFunc(ctx, file, objectName)
	}
	return "http://fake.gcs.url/" + objectName, nil // デフォルトの成功レスポンス
}

type mockGeminiService struct {
	ExtractReceiptInfoFunc func(ctx context.Context, imageFile io.Reader) (*domain.ExtractedReceiptInfo, error)
}

func (m *mockGeminiService) ExtractReceiptInfo(ctx context.Context, imageFile io.Reader) (*domain.ExtractedReceiptInfo, error) {
	if m.ExtractReceiptInfoFunc != nil {
		return m.ExtractReceiptInfoFunc(ctx, imageFile)
	}
	// デフォルトの成功レスポンス
	return &domain.ExtractedReceiptInfo{
		Date:        "2023-12-25",
		Store:       "Mock Store",
		Items:       "Mock Item 1, Mock Item 2",
		TotalAmount: 1234,
	}, nil
}

type mockReceiptRepository struct {
	SaveFunc func(ctx context.Context, receipt *domain.Receipt) error
}

func (m *mockReceiptRepository) Save(ctx context.Context, receipt *domain.Receipt) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, receipt)
	}
	receipt.ID = "mock-receipt-id" // 保存成功時にIDが設定されることを模倣
	return nil                     // デフォルトで成功
}

// --- CreateReceipt Tests ---

func TestCreateReceipt_Success(t *testing.T) {
	mockStorage := &mockStorageService{}
	mockGemini := &mockGeminiService{}
	mockRepo := &mockReceiptRepository{}

	service := NewReceiptService(mockRepo, mockGemini, mockStorage)

	// テスト用の入力データ
	ctx := context.Background()
	dateStrFromRequest := "2023-12-24" // ユーザー入力の日付
	payer := "Test Payer"
	paymentMethod := "Test Card"
	imageFile := strings.NewReader("dummy image data")
	imageFilename := "receipt.jpg"

	createdReceipt, err := service.CreateReceipt(ctx, dateStrFromRequest, payer, paymentMethod, imageFile, imageFilename)

	if err != nil {
		t.Fatalf("CreateReceipt failed: %v", err)
	}
	if createdReceipt == nil {
		t.Fatalf("Expected a receipt, got nil")
	}
	if createdReceipt.Date != dateStrFromRequest {
		t.Errorf("Expected date %s, got %s", dateStrFromRequest, createdReceipt.Date)
	}
	if createdReceipt.Payer != payer {
		t.Errorf("Expected payer %s, got %s", payer, createdReceipt.Payer)
	}
	if !strings.HasPrefix(createdReceipt.ImageURL, "http://fake.gcs.url/") {
		t.Errorf("Expected ImageURL to have prefix http://fake.gcs.url/, got %s", createdReceipt.ImageURL)
	}
	if createdReceipt.ID == "" { // モックでIDが設定されることを期待
		t.Errorf("Expected receipt ID to be set, but it was empty")
	}
}

func TestCreateReceipt_StorageError(t *testing.T) {
	expectedError := errors.New("GCS upload failed")
	mockStorage := &mockStorageService{
		UploadImageFunc: func(ctx context.Context, file io.Reader, objectName string) (string, error) {
			return "", expectedError
		},
	}
	mockGemini := &mockGeminiService{}   // エラーケースでは呼び出されないはず
	mockRepo := &mockReceiptRepository{} // エラーケースでは呼び出されないはず

	service := NewReceiptService(mockRepo, mockGemini, mockStorage)

	_, err := service.CreateReceipt(context.Background(), "", "Payer", "Method", strings.NewReader(""), "file.jpg")

	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	if !strings.Contains(err.Error(), expectedError.Error()) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError.Error(), err.Error())
	}
}

func TestCreateReceipt_GeminiError(t *testing.T) {
	expectedError := errors.New("Gemini extraction failed")
	mockStorage := &mockStorageService{}
	mockGemini := &mockGeminiService{
		ExtractReceiptInfoFunc: func(ctx context.Context, imageFile io.Reader) (*domain.ExtractedReceiptInfo, error) {
			return nil, expectedError
		},
	}
	mockRepo := &mockReceiptRepository{}

	service := NewReceiptService(mockRepo, mockGemini, mockStorage)
	_, err := service.CreateReceipt(context.Background(), "", "Payer", "Method", strings.NewReader(""), "file.jpg")

	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	if !strings.Contains(err.Error(), expectedError.Error()) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError.Error(), err.Error())
	}
}

// TODO: Repository.Save でエラーが発生するケースのテストも追加
