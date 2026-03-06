package services

import (
	"context"
	"errors"
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	repo      ports.UserRepository
	jwtSecret string
}

func NewAuthService(repo ports.UserRepository, secret string) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:      repo,
		jwtSecret: secret,
	}
}

// 1. Register: Mã hóa pass và lưu vào DB
func (s *AuthServiceImpl) Register(ctx context.Context, email, password, fullName string) error {
	// Kiểm tra user tồn tại chưa
	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return errors.New("email already exists")
	}

	// Hash password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPass),
		FullName:     fullName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.repo.CreateUser(ctx, user)
}

// 2. Login: Kiểm tra pass và tạo JWT Token
func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (string, error) {
	// Tìm user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// So sánh password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Tạo JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token sống 24h
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// 3. GetMe: Lấy thông tin user hiện tại
func (s *AuthServiceImpl) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}