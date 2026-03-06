package ports

import (
	"context"
	"idp-api-gateway/internal/core/domain"
)

// UserRepository: Giao tiếp với Database
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
}

// AuthService: Giao tiếp với Logic nghiệp vụ
type AuthService interface {
	Register(ctx context.Context, email, password, fullName string) error
	Login(ctx context.Context, email, password string) (string, error) // Trả về Token string
	GetMe(ctx context.Context, userID string) (*domain.User, error)
}