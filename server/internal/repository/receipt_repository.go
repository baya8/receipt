package repository

import (
	"receipt/server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReceiptRepository レシート関連データ操作インターフェース
type ReceiptRepository interface {
	Create(receipt *models.Receipt) error
	GetByID(id uuid.UUID) (*models.Receipt, error)
	GetByIDWithPayer(id uuid.UUID) (*models.Receipt, error)
	Update(receipt *models.Receipt) error
	Delete(receipt *models.Receipt) error
	GetReceiptsByFilter(groupID uuid.UUID, year *int, month *int) ([]models.Receipt, error)
}

type gormReceiptRepository struct {
	db *gorm.DB
}

// NewReceiptRepository ReceiptRepositoryの実装を作成
func NewReceiptRepository(db *gorm.DB) ReceiptRepository {
	return &gormReceiptRepository{db: db}
}

func (r *gormReceiptRepository) Create(receipt *models.Receipt) error {
	return r.db.Create(receipt).Error
}

func (r *gormReceiptRepository) GetByID(id uuid.UUID) (*models.Receipt, error) {
	var receipt models.Receipt
	if err := r.db.First(&receipt, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *gormReceiptRepository) GetByIDWithPayer(id uuid.UUID) (*models.Receipt, error) {
	var receipt models.Receipt
	if err := r.db.Preload("Payer").First(&receipt, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *gormReceiptRepository) Update(receipt *models.Receipt) error {
	return r.db.Save(receipt).Error
}

func (r *gormReceiptRepository) Delete(receipt *models.Receipt) error {
	return r.db.Delete(receipt).Error
}

func (r *gormReceiptRepository) GetReceiptsByFilter(groupID uuid.UUID, year *int, month *int) ([]models.Receipt, error) {
	db := r.db.Preload("Payer").Where("group_id = ?", groupID)
	if year != nil && month != nil {
		db = db.Where("settlement_year = ? AND settlement_month = ?", *year, *month)
	}

	var receipts []models.Receipt
	err := db.Order("date desc").Find(&receipts).Error
	return receipts, err
}
