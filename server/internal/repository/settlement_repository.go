package repository

import (
	"time"
	"receipt/server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SettlementRepository 精算関連データ操作インターフェース
type SettlementRepository interface {
	Create(settlement *models.Settlement) error
	GetSettlementsByFilter(groupID uuid.UUID, year int, month int) ([]models.Settlement, error)
	CreateSettlementAndSettleReceipts(settlement *models.Settlement) error
}

type gormSettlementRepository struct {
	db *gorm.DB
}

// NewSettlementRepository SettlementRepositoryの実装を作成
func NewSettlementRepository(db *gorm.DB) SettlementRepository {
	return &gormSettlementRepository{db: db}
}

func (r *gormSettlementRepository) Create(settlement *models.Settlement) error {
	return r.db.Create(settlement).Error
}

func (r *gormSettlementRepository) GetSettlementsByFilter(groupID uuid.UUID, year int, month int) ([]models.Settlement, error) {
	var settlements []models.Settlement
	err := r.db.Where("group_id = ? AND year = ? AND month = ?", groupID, year, month).
		Preload("SettledByUser").
		Order("created_at desc").
		Find(&settlements).Error
	return settlements, err
}

func (r *gormSettlementRepository) CreateSettlementAndSettleReceipts(settlement *models.Settlement) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(settlement).Error; err != nil {
			return err
		}

		now := time.Now()
		if err := tx.Model(&models.Receipt{}).
			Where("group_id = ? AND settlement_year = ? AND settlement_month = ? AND settled_at IS NULL", 
				settlement.GroupID, settlement.Year, settlement.Month).
			Update("settled_at", now).Error; err != nil {
			return err
		}

		return nil
	})
}
