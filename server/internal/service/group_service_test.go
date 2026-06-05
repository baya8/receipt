package service_test

import (
	"errors"
	"testing"
	"time"

	"receipt/server/internal/models"
	"receipt/server/internal/service"

	"github.com/google/uuid"
)

type mockGroupRepository struct {
	groups       map[uuid.UUID]*models.Group
	groupMembers map[uuid.UUID][]uuid.UUID // groupID -> []userID
}

func newMockGroupRepository() *mockGroupRepository {
	return &mockGroupRepository{
		groups:       make(map[uuid.UUID]*models.Group),
		groupMembers: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockGroupRepository) Create(group *models.Group) error {
	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupRepository) GetByID(id uuid.UUID) (*models.Group, error) {
	group, exists := m.groups[id]
	if !exists {
		return nil, errors.New("record not found")
	}
	copied := *group
	return &copied, nil
}

func (m *mockGroupRepository) GetByIDWithMembers(id uuid.UUID) (*models.Group, error) {
	group, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (m *mockGroupRepository) Update(group *models.Group) error {
	if _, exists := m.groups[group.ID]; !exists {
		return errors.New("record not found")
	}
	group.UpdatedAt = time.Now()
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupRepository) Delete(group *models.Group) error {
	if _, exists := m.groups[group.ID]; !exists {
		return errors.New("record not found")
	}
	delete(m.groups, group.ID)
	delete(m.groupMembers, group.ID)
	return nil
}

func (m *mockGroupRepository) GetGroupsByUserID(userID uuid.UUID) ([]models.Group, error) {
	var result []models.Group
	for groupID, memberIDs := range m.groupMembers {
		for _, mID := range memberIDs {
			if mID == userID {
				if g, exists := m.groups[groupID]; exists {
					result = append(result, *g)
				}
			}
		}
	}
	return result, nil
}

func (m *mockGroupRepository) AddMember(group *models.Group, user *models.User) error {
	m.groupMembers[group.ID] = append(m.groupMembers[group.ID], user.ID)
	if g, exists := m.groups[group.ID]; exists {
		g.Members = append(g.Members, *user)
	}
	return nil
}

func (m *mockGroupRepository) RemoveMember(group *models.Group, user *models.User) error {
	memberIDs := m.groupMembers[group.ID]
	for i, mID := range memberIDs {
		if mID == user.ID {
			m.groupMembers[group.ID] = append(memberIDs[:i], memberIDs[i+1:]...)
			break
		}
	}
	if g, exists := m.groups[group.ID]; exists {
		for i, u := range g.Members {
			if u.ID == user.ID {
				g.Members = append(g.Members[:i], g.Members[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (m *mockGroupRepository) IsMember(groupID uuid.UUID, userID uuid.UUID) (bool, error) {
	memberIDs := m.groupMembers[groupID]
	for _, mID := range memberIDs {
		if mID == userID {
			return true, nil
		}
	}
	return false, nil
}

func TestGroupService_CreateGroup(t *testing.T) {
	userRepo := newMockUserRepository()
	groupRepo := newMockGroupRepository()
	svc := service.NewGroupService(groupRepo, userRepo)

	owner := models.User{Email: "owner@example.com", Nickname: "Owner"}
	_ = userRepo.Create(&owner)

	group, err := svc.CreateGroup("Family", owner.ID)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if group.Name != "Family" {
		t.Errorf("Expected name 'Family', got '%s'", group.Name)
	}
	if group.OwnerID != owner.ID {
		t.Errorf("Expected OwnerID %s, got %s", owner.ID, group.OwnerID)
	}

	// オーナーがメンバーに含まれるかチェック
	isMember, _ := groupRepo.IsMember(group.ID, owner.ID)
	if !isMember {
		t.Errorf("Expected owner to be a member of the group")
	}
}

func TestGroupService_InviteMember(t *testing.T) {
	userRepo := newMockUserRepository()
	groupRepo := newMockGroupRepository()
	svc := service.NewGroupService(groupRepo, userRepo)

	owner := models.User{Email: "owner@example.com", Nickname: "Owner"}
	_ = userRepo.Create(&owner)

	guest := models.User{Email: "guest@example.com", Nickname: "Guest"}
	_ = userRepo.Create(&guest)

	group, _ := svc.CreateGroup("Family", owner.ID)

	t.Run("Success", func(t *testing.T) {
		err := svc.InviteMember(group.ID, owner.ID, guest.Email)
		if err != nil {
			t.Fatalf("InviteMember failed: %v", err)
		}

		isMember, _ := groupRepo.IsMember(group.ID, guest.ID)
		if !isMember {
			t.Errorf("Expected guest to be a member after invitation")
		}
	})

	t.Run("Not Owner", func(t *testing.T) {
		err := svc.InviteMember(group.ID, guest.ID, "other@example.com")
		if !errors.Is(err, service.ErrNotOwner) {
			t.Errorf("Expected error %v, got %v", service.ErrNotOwner, err)
		}
	})

	t.Run("Already Member", func(t *testing.T) {
		err := svc.InviteMember(group.ID, owner.ID, guest.Email)
		if !errors.Is(err, service.ErrAlreadyMember) {
			t.Errorf("Expected error %v, got %v", service.ErrAlreadyMember, err)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		err := svc.InviteMember(group.ID, owner.ID, "nonexistent@example.com")
		if !errors.Is(err, service.ErrInviteUserNotFound) {
			t.Errorf("Expected error %v, got %v", service.ErrInviteUserNotFound, err)
		}
	})
}

func TestGroupService_RemoveMember(t *testing.T) {
	userRepo := newMockUserRepository()
	groupRepo := newMockGroupRepository()
	svc := service.NewGroupService(groupRepo, userRepo)

	owner := models.User{Email: "owner@example.com", Nickname: "Owner"}
	_ = userRepo.Create(&owner)

	guest := models.User{Email: "guest@example.com", Nickname: "Guest"}
	_ = userRepo.Create(&guest)

	group, _ := svc.CreateGroup("Family", owner.ID)
	_ = svc.InviteMember(group.ID, owner.ID, guest.Email)

	t.Run("Not Owner", func(t *testing.T) {
		err := svc.RemoveMember(group.ID, guest.ID, guest.ID)
		if !errors.Is(err, service.ErrNotOwner) {
			t.Errorf("Expected error %v, got %v", service.ErrNotOwner, err)
		}
	})

	t.Run("Remove Owner", func(t *testing.T) {
		err := svc.RemoveMember(group.ID, owner.ID, owner.ID)
		if !errors.Is(err, service.ErrOwnerCannotBeRemoved) {
			t.Errorf("Expected error %v, got %v", service.ErrOwnerCannotBeRemoved, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		err := svc.RemoveMember(group.ID, owner.ID, guest.ID)
		if err != nil {
			t.Fatalf("RemoveMember failed: %v", err)
		}

		isMember, _ := groupRepo.IsMember(group.ID, guest.ID)
		if isMember {
			t.Errorf("Expected guest to be removed from the group")
		}
	})
}

func TestGroupService_UpdateGroup(t *testing.T) {
	userRepo := newMockUserRepository()
	groupRepo := newMockGroupRepository()
	svc := service.NewGroupService(groupRepo, userRepo)

	owner := models.User{Email: "owner@example.com", Nickname: "Owner"}
	_ = userRepo.Create(&owner)

	group, _ := svc.CreateGroup("Family", owner.ID)

	t.Run("Success", func(t *testing.T) {
		updated, err := svc.UpdateGroup(group.ID, owner.ID, "New Family")
		if err != nil {
			t.Fatalf("UpdateGroup failed: %v", err)
		}
		if updated.Name != "New Family" {
			t.Errorf("Expected name 'New Family', got '%s'", updated.Name)
		}
	})

	t.Run("Forbidden", func(t *testing.T) {
		guest := uuid.New()
		_, err := svc.UpdateGroup(group.ID, guest, "Hack")
		if !errors.Is(err, service.ErrNotOwner) {
			t.Errorf("Expected error %v, got %v", service.ErrNotOwner, err)
		}
	})
}

func TestGroupService_DeleteGroup(t *testing.T) {
	userRepo := newMockUserRepository()
	groupRepo := newMockGroupRepository()
	svc := service.NewGroupService(groupRepo, userRepo)

	owner := models.User{Email: "owner@example.com", Nickname: "Owner"}
	_ = userRepo.Create(&owner)

	group, _ := svc.CreateGroup("Family", owner.ID)

	t.Run("Forbidden", func(t *testing.T) {
		guest := uuid.New()
		err := svc.DeleteGroup(group.ID, guest)
		if !errors.Is(err, service.ErrNotOwner) {
			t.Errorf("Expected error %v, got %v", service.ErrNotOwner, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		err := svc.DeleteGroup(group.ID, owner.ID)
		if err != nil {
			t.Fatalf("DeleteGroup failed: %v", err)
		}

		_, err = groupRepo.GetByID(group.ID)
		if err == nil {
			t.Errorf("Expected group to be deleted, but it was found")
		}
	})
}
