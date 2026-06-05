package service

import (
	"errors"
	"receipt/server/internal/models"
	"receipt/server/internal/repository"
	"receipt/server/internal/utils"

	"github.com/google/uuid"
)

var (
	// ErrInvalidCredentials 認証情報が無効な場合のエラー
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserNotFound ユーザーが見つからない場合のエラー
	ErrUserNotFound       = errors.New("user not found")
)

// UserService ユーザー認証・情報管理に関するビジネスロジックインターフェース
type UserService interface {
	Register(email, password, nickname string) error
	Login(email, password string) (string, *models.User, error)
	GetMe(userID uuid.UUID) (*models.User, error)
	UpdateMe(userID uuid.UUID, email, nickname, password string) (*models.User, error)
}

type userServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService UserServiceの実装を作成
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userServiceImpl{userRepo: userRepo}
}

func (s *userServiceImpl) Register(email, password, nickname string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Nickname:     nickname,
	}

	return s.userRepo.Create(&user)
}

func (s *userServiceImpl) Login(email, password string) (string, *models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// 存在しない場合も、パスワード不一致と同様のエラーにする（セキュリティ対策）
		return "", nil, ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *userServiceImpl) GetMe(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userServiceImpl) UpdateMe(userID uuid.UUID, email, nickname, password string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if email != "" {
		user.Email = email
	}
	if nickname != "" {
		user.Nickname = nickname
	}
	if password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
