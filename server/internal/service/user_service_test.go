package service_test

import (
	"errors"
	"testing"
	"time"

	"receipt/server/internal/models"
	"receipt/server/internal/service"
	"receipt/server/internal/utils"

	"github.com/google/uuid"
)

type mockUserRepository struct {
	users map[uuid.UUID]*models.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *mockUserRepository) Create(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("record not found")
	}
	// コピーを返す（参照渡しによるテスト中の意図しない値変更を防ぐため）
	copiedUser := *user
	return &copiedUser, nil
}

func (m *mockUserRepository) GetByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			copiedUser := *user
			return &copiedUser, nil
		}
	}
	return nil, errors.New("record not found")
}

func (m *mockUserRepository) Update(user *models.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return errors.New("record not found")
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func TestUserService_Register(t *testing.T) {
	repo := newMockUserRepository()
	svc := service.NewUserService(repo)

	email := "test@example.com"
	password := "securepassword"
	nickname := "TestUser"

	err := svc.Register(email, password, nickname)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	user, err := repo.GetByEmail(email)
	if err != nil {
		t.Fatalf("Failed to fetch registered user: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
	if user.Nickname != nickname {
		t.Errorf("Expected nickname %s, got %s", nickname, user.Nickname)
	}
	if user.PasswordHash == password {
		t.Errorf("Password should be hashed, but it is stored in plain text")
	}
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		t.Errorf("Stored password hash does not match the original password")
	}
}

func TestUserService_Login(t *testing.T) {
	repo := newMockUserRepository()
	svc := service.NewUserService(repo)

	email := "test@example.com"
	password := "securepassword"
	nickname := "TestUser"

	// 事前にユーザーを登録
	_ = svc.Register(email, password, nickname)

	t.Run("Success", func(t *testing.T) {
		token, user, err := svc.Login(email, password)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		if token == "" {
			t.Errorf("Expected token to be generated, got empty string")
		}
		if user.Email != email {
			t.Errorf("Expected user email %s, got %s", email, user.Email)
		}
	})

	t.Run("Invalid Password", func(t *testing.T) {
		_, _, err := svc.Login(email, "wrongpassword")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("Expected error %v, got %v", service.ErrInvalidCredentials, err)
		}
	})

	t.Run("Non-existent User", func(t *testing.T) {
		_, _, err := svc.Login("notfound@example.com", password)
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("Expected error %v, got %v", service.ErrInvalidCredentials, err)
		}
	})
}

func TestUserService_GetMe(t *testing.T) {
	repo := newMockUserRepository()
	svc := service.NewUserService(repo)

	email := "test@example.com"
	password := "securepassword"
	nickname := "TestUser"

	_ = svc.Register(email, password, nickname)
	registeredUser, _ := repo.GetByEmail(email)

	t.Run("Success", func(t *testing.T) {
		user, err := svc.GetMe(registeredUser.ID)
		if err != nil {
			t.Fatalf("GetMe failed: %v", err)
		}
		if user.ID != registeredUser.ID {
			t.Errorf("Expected ID %s, got %s", registeredUser.ID, user.ID)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		_, err := svc.GetMe(uuid.New())
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("Expected error %v, got %v", service.ErrUserNotFound, err)
		}
	})
}

func TestUserService_UpdateMe(t *testing.T) {
	repo := newMockUserRepository()
	svc := service.NewUserService(repo)

	email := "test@example.com"
	password := "securepassword"
	nickname := "TestUser"

	_ = svc.Register(email, password, nickname)
	registeredUser, _ := repo.GetByEmail(email)

	t.Run("Update Nickname and Email", func(t *testing.T) {
		newEmail := "updated@example.com"
		newNickname := "UpdatedUser"

		user, err := svc.UpdateMe(registeredUser.ID, newEmail, newNickname, "")
		if err != nil {
			t.Fatalf("UpdateMe failed: %v", err)
		}

		if user.Email != newEmail {
			t.Errorf("Expected email %s, got %s", newEmail, user.Email)
		}
		if user.Nickname != newNickname {
			t.Errorf("Expected nickname %s, got %s", newNickname, user.Nickname)
		}
	})

	t.Run("Update Password", func(t *testing.T) {
		newPassword := "newsecurepassword"

		user, err := svc.UpdateMe(registeredUser.ID, "", "", newPassword)
		if err != nil {
			t.Fatalf("UpdateMe failed: %v", err)
		}

		if !utils.CheckPasswordHash(newPassword, user.PasswordHash) {
			t.Errorf("Updated password hash does not match the new password")
		}

		// 古いパスワードでログインできないことを確認
		_, _, err = svc.Login(user.Email, password)
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("Should not login with old password, but got err: %v", err)
		}

		// 新しいパスワードでログインできることを確認
		_, _, err = svc.Login(user.Email, newPassword)
		if err != nil {
			t.Errorf("Should login with new password, but failed: %v", err)
		}
	})
}
