package repository

import (
	"receipt/server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupRepository グループ関連データ操作インターフェース
type GroupRepository interface {
	Create(group *models.Group) error
	GetByID(id uuid.UUID) (*models.Group, error)
	GetByIDWithMembers(id uuid.UUID) (*models.Group, error)
	Update(group *models.Group) error
	Delete(group *models.Group) error
	GetGroupsByUserID(userID uuid.UUID) ([]models.Group, error)
	AddMember(group *models.Group, user *models.User) error
	RemoveMember(group *models.Group, user *models.User) error
	IsMember(groupID uuid.UUID, userID uuid.UUID) (bool, error)
}

type gormGroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository GroupRepositoryの実装を作成
func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &gormGroupRepository{db: db}
}

func (r *gormGroupRepository) Create(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *gormGroupRepository) GetByID(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	if err := r.db.First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *gormGroupRepository) GetByIDWithMembers(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	if err := r.db.Preload("Members").First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *gormGroupRepository) Update(group *models.Group) error {
	return r.db.Save(group).Error
}

func (r *gormGroupRepository) Delete(group *models.Group) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. レシートを削除
		if err := tx.Where("group_id = ?", group.ID).Delete(&models.Receipt{}).Error; err != nil {
			return err
		}

		// 2. 精算履歴を削除
		if err := tx.Where("group_id = ?", group.ID).Delete(&models.Settlement{}).Error; err != nil {
			return err
		}

		// 3. メンバーとの紐付けを解除
		if err := tx.Model(group).Association("Members").Clear(); err != nil {
			return err
		}

		// 4. グループ自体を削除
		if err := tx.Delete(group).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *gormGroupRepository) GetGroupsByUserID(userID uuid.UUID) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Preload("Members").
		Find(&groups).Error
	return groups, err
}

func (r *gormGroupRepository) AddMember(group *models.Group, user *models.User) error {
	return r.db.Model(group).Association("Members").Append(user)
}

func (r *gormGroupRepository) RemoveMember(group *models.Group, user *models.User) error {
	return r.db.Model(group).Association("Members").Delete(user)
}

func (r *gormGroupRepository) IsMember(groupID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("group_members").Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error
	return count > 0, err
}
