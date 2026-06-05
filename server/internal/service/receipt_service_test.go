package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"receipt/server/internal/models"
	"receipt/server/internal/service"

	"github.com/google/uuid"
)

type mockReceiptRepository struct {
	receipts map[uuid.UUID]*models.Receipt
}

func newMockReceiptRepository() *mockReceiptRepository {
	return &mockReceiptRepository{
		receipts: make(map[uuid.UUID]*models.Receipt),
	}
}

func (m *mockReceiptRepository) Create(receipt *models.Receipt) error {
	if receipt.ID == uuid.Nil {
		receipt.ID = uuid.New()
	}
	receipt.CreatedAt = time.Now()
	receipt.UpdatedAt = time.Now()
	m.receipts[receipt.ID] = receipt
	return nil
}

func (m *mockReceiptRepository) GetByID(id uuid.UUID) (*models.Receipt, error) {
	receipt, exists := m.receipts[id]
	if !exists {
		return nil, errors.New("record not found")
	}
	copied := *receipt
	return &copied, nil
}

func (m *mockReceiptRepository) GetByIDWithPayer(id uuid.UUID) (*models.Receipt, error) {
	return m.GetByID(id)
}

func (m *mockReceiptRepository) Update(receipt *models.Receipt) error {
	if _, exists := m.receipts[receipt.ID]; !exists {
		return errors.New("record not found")
	}
	receipt.UpdatedAt = time.Now()
	m.receipts[receipt.ID] = receipt
	return nil
}

func (m *mockReceiptRepository) Delete(receipt *models.Receipt) error {
	if _, exists := m.receipts[receipt.ID]; !exists {
		return errors.New("record not found")
	}
	delete(m.receipts, receipt.ID)
	return nil
}

func (m *mockReceiptRepository) GetReceiptsByFilter(groupID uuid.UUID, year *int, month *int) ([]models.Receipt, error) {
	var result []models.Receipt
	for _, receipt := range m.receipts {
		if receipt.GroupID == groupID {
			if year != nil && month != nil {
				if receipt.SettlementYear == *year && receipt.SettlementMonth == *month {
					result = append(result, *receipt)
				}
			} else {
				result = append(result, *receipt)
			}
		}
	}
	return result, nil
}

type mockAIAnalyzer struct {
	analyzeFunc func(ctx context.Context, imgData []byte) (*service.AnalyzeReceiptResult, error)
}

func (m *mockAIAnalyzer) AnalyzeReceipt(ctx context.Context, imgData []byte) (*service.AnalyzeReceiptResult, error) {
	return m.analyzeFunc(ctx, imgData)
}

func TestReceiptService_CreateReceipt(t *testing.T) {
	repo := newMockReceiptRepository()
	svc := service.NewReceiptService(repo)

	userID := uuid.New()
	groupID := uuid.New()
	payerID := userID
	date := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)

	t.Run("Success with automatic settlement date", func(t *testing.T) {
		params := &service.CreateReceiptParams{
			GroupID:       groupID,
			Date:          date,
			Shop:          "Supermarket",
			Item:          "Groceries",
			Amount:        1500,
			PayerID:       payerID,
			PaymentMethod: "half",
		}

		receipt, err := svc.CreateReceipt(params, userID)
		if err != nil {
			t.Fatalf("CreateReceipt failed: %v", err)
		}

		if receipt.Shop != "Supermarket" {
			t.Errorf("Expected shop 'Supermarket', got '%s'", receipt.Shop)
		}
		if receipt.SettlementYear != 2026 || receipt.SettlementMonth != 6 {
			t.Errorf("Expected settlement date 2026/6, got %d/%d", receipt.SettlementYear, receipt.SettlementMonth)
		}
	})

	t.Run("Invalid Amount", func(t *testing.T) {
		params := &service.CreateReceiptParams{
			GroupID:       groupID,
			Date:          date,
			Amount:        0,
			PayerID:       payerID,
			PaymentMethod: "half",
		}

		_, err := svc.CreateReceipt(params, userID)
		if !errors.Is(err, service.ErrInvalidAmount) {
			t.Errorf("Expected error %v, got %v", service.ErrInvalidAmount, err)
		}
	})
}

func TestReceiptService_GetReceipt(t *testing.T) {
	repo := newMockReceiptRepository()
	svc := service.NewReceiptService(repo)

	userID := uuid.New()
	groupID := uuid.New()
	date := time.Now()

	params := &service.CreateReceiptParams{
		GroupID:       groupID,
		Date:          date,
		Shop:          "Supermarket",
		Amount:        1500,
		PayerID:       userID,
		PaymentMethod: "half",
	}

	created, _ := svc.CreateReceipt(params, userID)

	t.Run("Success", func(t *testing.T) {
		receipt, err := svc.GetReceipt(created.ID)
		if err != nil {
			t.Fatalf("GetReceipt failed: %v", err)
		}
		if receipt.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, receipt.ID)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := svc.GetReceipt(uuid.New())
		if !errors.Is(err, service.ErrReceiptNotFound) {
			t.Errorf("Expected error %v, got %v", service.ErrReceiptNotFound, err)
		}
	})
}

func TestReceiptService_UpdateReceipt(t *testing.T) {
	repo := newMockReceiptRepository()
	svc := service.NewReceiptService(repo)

	userID := uuid.New()
	otherUserID := uuid.New()
	groupID := uuid.New()
	date := time.Now()

	params := &service.CreateReceiptParams{
		GroupID:       groupID,
		Date:          date,
		Shop:          "Supermarket",
		Amount:        1500,
		PayerID:       userID,
		PaymentMethod: "half",
	}

	created, _ := svc.CreateReceipt(params, userID)

	t.Run("Success", func(t *testing.T) {
		updateParams := &service.CreateReceiptParams{
			GroupID:       groupID,
			Date:          date,
			Shop:          "Convenience Store",
			Amount:        500,
			PayerID:       userID,
			PaymentMethod: "self",
		}

		updated, err := svc.UpdateReceipt(created.ID, updateParams, userID)
		if err != nil {
			t.Fatalf("UpdateReceipt failed: %v", err)
		}
		if updated.Shop != "Convenience Store" {
			t.Errorf("Expected shop 'Convenience Store', got '%s'", updated.Shop)
		}
		if updated.Amount != 500 {
			t.Errorf("Expected amount 500, got %d", updated.Amount)
		}
	})

	t.Run("Forbidden - Not Creator", func(t *testing.T) {
		updateParams := &service.CreateReceiptParams{
			GroupID: groupID,
			Date:    date,
			Amount:  1000,
		}

		_, err := svc.UpdateReceipt(created.ID, updateParams, otherUserID)
		if !errors.Is(err, service.ErrNotCreator) {
			t.Errorf("Expected error %v, got %v", service.ErrNotCreator, err)
		}
	})

	t.Run("Forbidden - Already Settled", func(t *testing.T) {
		// モック上で直接精算済みフラグを立てる
		rawReceipt, _ := repo.GetByID(created.ID)
		now := time.Now()
		rawReceipt.SettledAt = &now
		_ = repo.Update(rawReceipt)

		updateParams := &service.CreateReceiptParams{
			GroupID: groupID,
			Date:    date,
			Amount:  1000,
		}

		_, err := svc.UpdateReceipt(created.ID, updateParams, userID)
		if !errors.Is(err, service.ErrAlreadySettled) {
			t.Errorf("Expected error %v, got %v", service.ErrAlreadySettled, err)
		}
	})
}

func TestReceiptService_DeleteReceipt(t *testing.T) {
	repo := newMockReceiptRepository()
	svc := service.NewReceiptService(repo)

	userID := uuid.New()
	otherUserID := uuid.New()
	groupID := uuid.New()
	date := time.Now()

	params := &service.CreateReceiptParams{
		GroupID:       groupID,
		Date:          date,
		Shop:          "Supermarket",
		Amount:        1500,
		PayerID:       userID,
		PaymentMethod: "half",
	}

	created, _ := svc.CreateReceipt(params, userID)

	t.Run("Forbidden - Not Creator", func(t *testing.T) {
		err := svc.DeleteReceipt(created.ID, otherUserID)
		if !errors.Is(err, service.ErrNotCreator) {
			t.Errorf("Expected error %v, got %v", service.ErrNotCreator, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		err := svc.DeleteReceipt(created.ID, userID)
		if err != nil {
			t.Fatalf("DeleteReceipt failed: %v", err)
		}

		_, err = repo.GetByID(created.ID)
		if err == nil {
			t.Errorf("Expected receipt to be deleted, but it was found")
		}
	})
}

func TestAIAnalyzer_Mock(t *testing.T) {
	mockResult := &service.AnalyzeReceiptResult{
		Date:   "2026-06-05",
		Shop:   "Mock Shop",
		Item:   "Mock Item",
		Amount: 2000,
	}

	analyzer := &mockAIAnalyzer{
		analyzeFunc: func(ctx context.Context, imgData []byte) (*service.AnalyzeReceiptResult, error) {
			return mockResult, nil
		},
	}

	result, err := analyzer.AnalyzeReceipt(context.Background(), []byte("fake image data"))
	if err != nil {
		t.Fatalf("AnalyzeReceipt failed: %v", err)
	}

	if result.Shop != "Mock Shop" {
		t.Errorf("Expected shop 'Mock Shop', got '%s'", result.Shop)
	}
	if result.Amount != 2000 {
		t.Errorf("Expected amount 2000, got %d", result.Amount)
	}
}
