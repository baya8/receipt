package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"receipt/server/internal/models"
	"receipt/server/internal/repository"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

var (
	// ErrReceiptNotFound レシートが見つからない場合のエラー
	ErrReceiptNotFound = errors.New("receipt not found")
	// ErrNotCreator 作成者以外による編集・削除時のエラー
	ErrNotCreator      = errors.New("only the creator can modify this receipt")
	// ErrAlreadySettled 精算済みレシートに対する操作エラー
	ErrAlreadySettled  = errors.New("cannot modify settled receipt")
	// ErrInvalidAmount 金額が不正な場合のエラー
	ErrInvalidAmount   = errors.New("amount must be at least 1")
)

// CreateReceiptParams レシート作成・更新用パラメータ
type CreateReceiptParams struct {
	GroupID         uuid.UUID
	Date            time.Time
	SettlementYear  int
	SettlementMonth int
	Shop            string
	Item            string
	Amount          int
	PayerID         uuid.UUID
	PaymentMethod   string
}

// ReceiptService レシートのCRUD管理に関するビジネスロジックインターフェース
type ReceiptService interface {
	GetReceipts(groupID uuid.UUID, year *int, month *int) ([]models.Receipt, error)
	CreateReceipt(params *CreateReceiptParams, userID uuid.UUID) (*models.Receipt, error)
	GetReceipt(id uuid.UUID) (*models.Receipt, error)
	UpdateReceipt(id uuid.UUID, params *CreateReceiptParams, userID uuid.UUID) (*models.Receipt, error)
	DeleteReceipt(id uuid.UUID, userID uuid.UUID) error
}

type receiptServiceImpl struct {
	receiptRepo repository.ReceiptRepository
}

// NewReceiptService ReceiptServiceの実装を作成
func NewReceiptService(receiptRepo repository.ReceiptRepository) ReceiptService {
	return &receiptServiceImpl{receiptRepo: receiptRepo}
}

func (s *receiptServiceImpl) GetReceipts(groupID uuid.UUID, year *int, month *int) ([]models.Receipt, error) {
	return s.receiptRepo.GetReceiptsByFilter(groupID, year, month)
}

func (s *receiptServiceImpl) CreateReceipt(params *CreateReceiptParams, userID uuid.UUID) (*models.Receipt, error) {
	if params.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	settlementYear := params.SettlementYear
	settlementMonth := params.SettlementMonth
	if settlementYear == 0 || settlementMonth == 0 {
		settlementYear = params.Date.Year()
		settlementMonth = int(params.Date.Month())
	}

	receipt := models.Receipt{
		GroupID:         params.GroupID,
		UserID:          userID,
		Date:            params.Date,
		SettlementYear:  settlementYear,
		SettlementMonth: settlementMonth,
		Shop:            params.Shop,
		Item:            params.Item,
		Amount:          params.Amount,
		PayerID:         params.PayerID,
		PaymentMethod:   params.PaymentMethod,
	}

	if err := s.receiptRepo.Create(&receipt); err != nil {
		return nil, err
	}

	return s.receiptRepo.GetByIDWithPayer(receipt.ID)
}

func (s *receiptServiceImpl) GetReceipt(id uuid.UUID) (*models.Receipt, error) {
	receipt, err := s.receiptRepo.GetByIDWithPayer(id)
	if err != nil {
		return nil, ErrReceiptNotFound
	}
	return receipt, nil
}

func (s *receiptServiceImpl) UpdateReceipt(id uuid.UUID, params *CreateReceiptParams, userID uuid.UUID) (*models.Receipt, error) {
	receipt, err := s.receiptRepo.GetByID(id)
	if err != nil {
		return nil, ErrReceiptNotFound
	}

	if receipt.UserID != userID {
		return nil, ErrNotCreator
	}

	if receipt.SettledAt != nil {
		return nil, ErrAlreadySettled
	}

	if params.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	settlementYear := params.SettlementYear
	settlementMonth := params.SettlementMonth
	if settlementYear == 0 || settlementMonth == 0 {
		settlementYear = params.Date.Year()
		settlementMonth = int(params.Date.Month())
	}

	receipt.Date = params.Date
	receipt.SettlementYear = settlementYear
	receipt.SettlementMonth = settlementMonth
	receipt.Shop = params.Shop
	receipt.Item = params.Item
	receipt.Amount = params.Amount
	receipt.PayerID = params.PayerID
	receipt.PaymentMethod = params.PaymentMethod

	if err := s.receiptRepo.Update(receipt); err != nil {
		return nil, err
	}

	return s.receiptRepo.GetByIDWithPayer(receipt.ID)
}

func (s *receiptServiceImpl) DeleteReceipt(id uuid.UUID, userID uuid.UUID) error {
	receipt, err := s.receiptRepo.GetByID(id)
	if err != nil {
		return ErrReceiptNotFound
	}

	if receipt.UserID != userID {
		return ErrNotCreator
	}

	if receipt.SettledAt != nil {
		return ErrAlreadySettled
	}

	return s.receiptRepo.Delete(receipt)
}

// AIAnalyzer AIによるレシート解析インターフェース
type AIAnalyzer interface {
	AnalyzeReceipt(ctx context.Context, imgData []byte) (*AnalyzeReceiptResult, error)
}

// AnalyzeReceiptResult 解析結果
type AnalyzeReceiptResult struct {
	Date   string `json:"date"`
	Shop   string `json:"shop"`
	Item   string `json:"item"`
	Amount int    `json:"amount"`
}

type geminiAIAnalyzer struct {
	apiKey string
}

// NewAIAnalyzer AIAnalyzerの実装を作成
func NewAIAnalyzer(apiKey string) AIAnalyzer {
	return &geminiAIAnalyzer{apiKey: apiKey}
}

func (a *geminiAIAnalyzer) AnalyzeReceipt(ctx context.Context, imgData []byte) (*AnalyzeReceiptResult, error) {
	if a.apiKey == "" {
		return nil, errors.New("GOOGLE_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(a.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-flash-latest")

	prompt := []genai.Part{
		genai.ImageData("jpeg", imgData),
		genai.Text("Analyze this receipt and return JSON only. Use YYYY-MM-DD for date, name for shop, summary for item, and integer for amount. JSON:\n{\"date\": \"YYYY-MM-DD\", \"shop\": \"name\", \"item\": \"summary\", \"amount\": 1234}"),
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no results from Gemini")
	}

	var resultText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			resultText += string(text)
		}
	}

	resultText = strings.TrimSpace(resultText)
	resultText = strings.TrimPrefix(resultText, "```json")
	resultText = strings.TrimPrefix(resultText, "```")
	resultText = strings.TrimSuffix(resultText, "```")
	resultText = strings.TrimSpace(resultText)

	var analyzeResult AnalyzeReceiptResult
	if err := json.Unmarshal([]byte(resultText), &analyzeResult); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	return &analyzeResult, nil
}
