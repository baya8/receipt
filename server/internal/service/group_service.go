package service

import (
	"errors"
	"receipt/server/internal/models"
	"receipt/server/internal/repository"

	"github.com/google/uuid"
)

var (
	// ErrGroupNotFound グループが見つからない場合のエラー
	ErrGroupNotFound       = errors.New("group not found")
	// ErrNotOwner グループオーナー以外の操作に対するエラー
	ErrNotOwner            = errors.New("only group owner can perform this action")
	// ErrInviteUserNotFound 招待相手のユーザーが見つからない場合のエラー
	ErrInviteUserNotFound  = errors.New("招待相手のユーザーが見つかりません。先に相手の方がアカウント登録を完了しているか確認してください。")
	// ErrAlreadyMember 既にメンバーの場合のエラー
	ErrAlreadyMember       = errors.New("user is already a member of this group")
	// ErrOwnerCannotBeRemoved オーナー自身を削除しようとした場合のエラー
	ErrOwnerCannotBeRemoved = errors.New("owner cannot be removed from the group")
	// ErrMemberNotFound 削除対象メンバーが見つからない場合のエラー
	ErrMemberNotFound      = errors.New("user to remove not found")
)

// GroupService グループの管理に関するビジネスロジックインターフェース
type GroupService interface {
	CreateGroup(name string, ownerID uuid.UUID) (*models.Group, error)
	InviteMember(groupID uuid.UUID, ownerID uuid.UUID, email string) error
	RemoveMember(groupID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID) error
	GetMyGroups(userID uuid.UUID) ([]models.Group, error)
	UpdateGroup(groupID uuid.UUID, ownerID uuid.UUID, name string) (*models.Group, error)
	DeleteGroup(groupID uuid.UUID, ownerID uuid.UUID) error
}

type groupServiceImpl struct {
	groupRepo repository.GroupRepository
	userRepo  repository.UserRepository
}

// NewGroupService GroupServiceの実装を作成
func NewGroupService(groupRepo repository.GroupRepository, userRepo repository.UserRepository) GroupService {
	return &groupServiceImpl{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (s *groupServiceImpl) CreateGroup(name string, ownerID uuid.UUID) (*models.Group, error) {
	group := models.Group{
		Name:    name,
		OwnerID: ownerID,
	}

	if err := s.groupRepo.Create(&group); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return nil, err
	}

	if err := s.groupRepo.AddMember(&group, user); err != nil {
		return nil, err
	}

	return s.groupRepo.GetByIDWithMembers(group.ID)
}

func (s *groupServiceImpl) InviteMember(groupID uuid.UUID, ownerID uuid.UUID, email string) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	if group.OwnerID != ownerID {
		return ErrNotOwner
	}

	userToInvite, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return ErrInviteUserNotFound
	}

	isMember, err := s.groupRepo.IsMember(groupID, userToInvite.ID)
	if err != nil {
		return err
	}
	if isMember {
		return ErrAlreadyMember
	}

	return s.groupRepo.AddMember(group, userToInvite)
}

func (s *groupServiceImpl) RemoveMember(groupID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	if group.OwnerID != ownerID {
		return ErrNotOwner
	}

	if memberID == group.OwnerID {
		return ErrOwnerCannotBeRemoved
	}

	userToRemove, err := s.userRepo.GetByID(memberID)
	if err != nil {
		return ErrMemberNotFound
	}

	return s.groupRepo.RemoveMember(group, userToRemove)
}

func (s *groupServiceImpl) GetMyGroups(userID uuid.UUID) ([]models.Group, error) {
	return s.groupRepo.GetGroupsByUserID(userID)
}

func (s *groupServiceImpl) UpdateGroup(groupID uuid.UUID, ownerID uuid.UUID, name string) (*models.Group, error) {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	if group.OwnerID != ownerID {
		return nil, ErrNotOwner
	}

	group.Name = name
	if err := s.groupRepo.Update(group); err != nil {
		return nil, err
	}

	return s.groupRepo.GetByIDWithMembers(group.ID)
}

func (s *groupServiceImpl) DeleteGroup(groupID uuid.UUID, ownerID uuid.UUID) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	if group.OwnerID != ownerID {
		return ErrNotOwner
	}

	return s.groupRepo.Delete(group)
}
