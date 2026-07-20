package service_test

import (
	"errors"
	"testing"
	"time"

	"receipt/server/internal/models"
	"receipt/server/internal/service"

	"github.com/google/uuid"
)

type mockSettlementRepository struct {
	settlements map[uuid.UUID]*models.Settlement
}

func newMockSettlementRepository() *mockSettlementRepository {
	return &mockSettlementRepository{
		settlements: make(map[uuid.UUID]*models.Settlement),
	}
}

func (m *mockSettlementRepository) Create(settlement *models.Settlement) error {
	if settlement.ID == uuid.Nil {
		settlement.ID = uuid.New()
	}
	settlement.CreatedAt = time.Now()
	m.settlements[settlement.ID] = settlement
	return nil
}

func (m *mockSettlementRepository) GetSettlementsByFilter(groupID uuid.UUID, year int, month int) ([]models.Settlement, error) {
	var result []models.Settlement
	for _, settlement := range m.settlements {
		if settlement.GroupID == groupID && settlement.Year == year && settlement.Month == month {
			result = append(result, *settlement)
		}
	}
	return result, nil
}

func (m *mockSettlementRepository) CreateSettlementAndSettleReceipts(settlement *models.Settlement) error {
	return m.Create(settlement)
}

func TestSummaryService_GetMonthlySummary(t *testing.T) {
	userRepo := newMockUserRepository()
	groupRepo := newMockGroupRepository()
	receiptRepo := newMockReceiptRepository()
	settlementRepo := newMockSettlementRepository()

	svc := service.NewSummaryService(groupRepo, receiptRepo, settlementRepo)

	// テストデータ準備
	userA := models.User{Email: "usera@example.com", Nickname: "UserA"}
	_ = userRepo.Create(&userA)
	userB := models.User{Email: "userb@example.com", Nickname: "UserB"}
	_ = userRepo.Create(&userB)

	// グループ作成とメンバー追加
	group := models.Group{Name: "Family", OwnerID: userA.ID}
	_ = groupRepo.Create(&group)
	_ = groupRepo.AddMember(&group, &userA)
	_ = groupRepo.AddMember(&group, &userB)

	year := 2026
	month := 6
	date := time.Date(year, time.Month(month), 15, 12, 0, 0, 0, time.UTC)

	// レシート登録
	// 1. 折半 (1000円) -> 各自500円負担
	r1 := models.Receipt{
		GroupID:         group.ID,
		UserID:          userA.ID,
		Date:            date,
		SettlementYear:  year,
		SettlementMonth: month,
		Amount:          1000,
		PayerID:         userA.ID,
		PaymentMethod:   "half",
	}
	_ = receiptRepo.Create(&r1)

	// 2. 自分が10割 (1500円) -> 支払ったAが1500円負担
	r2 := models.Receipt{
		GroupID:         group.ID,
		UserID:          userA.ID,
		Date:            date,
		SettlementYear:  year,
		SettlementMonth: month,
		Amount:          1500,
		PayerID:         userA.ID,
		PaymentMethod:   "self",
	}
	_ = receiptRepo.Create(&r2)

	// 3. 相手負担 (1200円) -> 支払っていないBが1200円負担
	r3 := models.Receipt{
		GroupID:         group.ID,
		UserID:          userA.ID,
		Date:            date,
		SettlementYear:  year,
		SettlementMonth: month,
		Amount:          1200,
		PayerID:         userA.ID,
		PaymentMethod:   "other",
	}
	_ = receiptRepo.Create(&r3)

	// 4. 折半で端数あり (101円) -> 均等50円、端数1円は支払者Aが負担 (A=51, B=50)
	r4 := models.Receipt{
		GroupID:         group.ID,
		UserID:          userA.ID,
		Date:            date,
		SettlementYear:  year,
		SettlementMonth: month,
		Amount:          101,
		PayerID:         userA.ID,
		PaymentMethod:   "half",
	}
	_ = receiptRepo.Create(&r4)

	result, err := svc.GetMonthlySummary(group.ID, year, month)
	if err != nil {
		t.Fatalf("GetMonthlySummary failed: %v", err)
	}

	expectedTotalSpent := 1000 + 1500 + 1200 + 101 // 3801
	if result.TotalSpent != expectedTotalSpent {
		t.Errorf("Expected total spent %d, got %d", expectedTotalSpent, result.TotalSpent)
	}

	// メンバーごとのアサーション
	// UserA の期待値:
	// - Paid: 1000 + 1500 + 1200 + 101 = 3801
	// - Share: 500 (r1) + 1500 (r2) + 0 (r3) + 51 (r4) = 2051
	// UserB の期待値:
	// - Paid: 0
	// - Share: 500 (r1) + 0 (r2) + 1200 (r3) + 50 (r4) = 1750
	var summaryA, summaryB service.MemberSummary
	for _, m := range result.Members {
		if m.UserID == userA.ID {
			summaryA = m
		} else if m.UserID == userB.ID {
			summaryB = m
		}
	}

	if summaryA.Paid != 3801 || summaryA.Share != 2051 {
		t.Errorf("UserA: Expected Paid=3801, Share=2051. Got Paid=%d, Share=%d", summaryA.Paid, summaryA.Share)
	}

	if summaryB.Paid != 0 || summaryB.Share != 1750 {
		t.Errorf("UserB: Expected Paid=0, Share=1750. Got Paid=%d, Share=%d", summaryB.Paid, summaryB.Share)
	}
}

func TestSummaryService_CreateSettlement(t *testing.T) {
	groupRepo := newMockGroupRepository()
	receiptRepo := newMockReceiptRepository()
	settlementRepo := newMockSettlementRepository()

	svc := service.NewSummaryService(groupRepo, receiptRepo, settlementRepo)

	groupID := uuid.New()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		settlement, err := svc.CreateSettlement(groupID, 2026, 6, 5000, userID)
		if err != nil {
			t.Fatalf("CreateSettlement failed: %v", err)
		}

		if settlement.Amount != 5000 {
			t.Errorf("Expected amount 5000, got %d", settlement.Amount)
		}
		if settlement.GroupID != groupID {
			t.Errorf("Expected group ID %s, got %s", groupID, settlement.GroupID)
		}
	})

	t.Run("Invalid Amount", func(t *testing.T) {
		_, err := svc.CreateSettlement(groupID, 2026, 6, 0, userID)
		if !errors.Is(err, service.ErrInvalidSettlementAmount) {
			t.Errorf("Expected error %v, got %v", service.ErrInvalidSettlementAmount, err)
		}
	})
}
