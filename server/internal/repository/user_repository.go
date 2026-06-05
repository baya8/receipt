package repository

import (
	"receipt/server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository ユーザー情報に関するデータ操作インターフェース
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
}

type gormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository UserRepositoryの実装を作成
func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *gormUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}
