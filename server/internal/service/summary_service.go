package service

import (
	"errors"
	"receipt/server/internal/models"
	"receipt/server/internal/repository"

	"github.com/google/uuid"
)

var (
	// ErrInvalidSettlementAmount 精算金額が不正な場合のエラー
	ErrInvalidSettlementAmount = errors.New("settlement amount must be at least 1")
)

// MemberSummary メンバーごとの出費集計
type MemberSummary struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
	Paid     int       `json:"paid"`  // 実際に支払った合計
	Share    int       `json:"share"` // 負担すべき合計
}

// MonthlySummaryResult 月次サマリー集計結果
type MonthlySummaryResult struct {
	TotalSpent  int                 `json:"total_spent"`
	Members     []MemberSummary     `json:"members"`
	Settlements []models.Settlement `json:"settlements"`
}

// SummaryService 精算計算・月次集計に関するビジネスロジックインターフェース
type SummaryService interface {
	GetMonthlySummary(groupID uuid.UUID, year int, month int) (*MonthlySummaryResult, error)
	CreateSettlement(groupID uuid.UUID, year int, month int, amount int, settledBy uuid.UUID) (*models.Settlement, error)
}

type summaryServiceImpl struct {
	groupRepo      repository.GroupRepository
	receiptRepo    repository.ReceiptRepository
	settlementRepo repository.SettlementRepository
}

// NewSummaryService SummaryServiceの実装を作成
func NewSummaryService(
	groupRepo repository.GroupRepository,
	receiptRepo repository.ReceiptRepository,
	settlementRepo repository.SettlementRepository,
) SummaryService {
	return &summaryServiceImpl{
		groupRepo:      groupRepo,
		receiptRepo:    receiptRepo,
		settlementRepo: settlementRepo,
	}
}

func (s *summaryServiceImpl) GetMonthlySummary(groupID uuid.UUID, year int, month int) (*MonthlySummaryResult, error) {
	group, err := s.groupRepo.GetByIDWithMembers(groupID)
	if err != nil {
		return nil, err
	}

	settlements, err := s.settlementRepo.GetSettlementsByFilter(groupID, year, month)
	if err != nil {
		return nil, err
	}

	receipts, err := s.receiptRepo.GetReceiptsByFilter(groupID, &year, &month)
	if err != nil {
		return nil, err
	}

	paidMap := make(map[uuid.UUID]int)
	shareMap := make(map[uuid.UUID]int)
	totalSpent := 0

	for _, r := range receipts {
		totalSpent += r.Amount
		paidMap[r.PayerID] += r.Amount

		switch r.PaymentMethod {
		case models.PaymentMethodSelf:
			// 支払者が全額負担
			shareMap[r.PayerID] += r.Amount

		case models.PaymentMethodOther:
			// 支払者以外で均等に負担（2人以上の場合に対応）
			otherCount := len(group.Members) - 1
			if otherCount > 0 {
				sharePerPerson := r.Amount / otherCount
				remainder := r.Amount % otherCount

				isRemainderAssigned := false
				for _, m := range group.Members {
					if m.ID != r.PayerID {
						shareMap[m.ID] += sharePerPerson
						if !isRemainderAssigned {
							shareMap[m.ID] += remainder
							isRemainderAssigned = true
						}
					}
				}
			} else {
				// 相手がいない場合は支払者が負担
				shareMap[r.PayerID] += r.Amount
			}

		case models.PaymentMethodHalf:
			fallthrough
		default:
			// メンバー全員で均等割り
			numMembers := len(group.Members)
			if numMembers > 0 {
				sharePerPerson := r.Amount / numMembers
				remainder := r.Amount % numMembers

				for _, m := range group.Members {
					shareMap[m.ID] += sharePerPerson
				}
				// 端数は支払者が負担
				shareMap[r.PayerID] += remainder
			}
		}
	}

	var memberSummaries []MemberSummary
	for _, m := range group.Members {
		memberSummaries = append(memberSummaries, MemberSummary{
			UserID:   m.ID,
			Nickname: m.Nickname,
			Paid:     paidMap[m.ID],
			Share:    shareMap[m.ID],
		})
	}

	return &MonthlySummaryResult{
		TotalSpent:  totalSpent,
		Members:     memberSummaries,
		Settlements: settlements,
	}, nil
}

func (s *summaryServiceImpl) CreateSettlement(groupID uuid.UUID, year int, month int, amount int, settledBy uuid.UUID) (*models.Settlement, error) {
	if amount <= 0 {
		return nil, ErrInvalidSettlementAmount
	}

	settlement := models.Settlement{
		GroupID:   groupID,
		Year:      year,
		Month:     month,
		Amount:    amount,
		SettledBy: settledBy,
	}

	if err := s.settlementRepo.CreateSettlementAndSettleReceipts(&settlement); err != nil {
		return nil, err
	}

	return &settlement, nil
}
