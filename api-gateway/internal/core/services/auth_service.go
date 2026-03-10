package services

import (
	"context"
	"errors"
	"fmt"
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// RedisTokenStore defines the interface required by AuthService for token storage.
type RedisTokenStore interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type AuthServiceImpl struct {
	repo      ports.UserRepository
	jwtSecret string
	redis     RedisTokenStore
}

func NewAuthService(repo ports.UserRepository, secret string, redisClient RedisTokenStore) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:      repo,
		jwtSecret: secret,
		redis:     redisClient,
	}
}

// 1. Register: Hash password and persist to DB
func (s *AuthServiceImpl) Register(ctx context.Context, email, password, fullName string) error {
	// Check if user exists
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
		Role:         "EMPLOYEE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.repo.CreateUser(ctx, user)
}

// 2. Login: Verify credentials, generate Access Token (15m) + Refresh Token (7d), store refresh in Redis
func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// Generate Access Token (15 minutes)
	accessToken, err := s.generateToken(user, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token (7 days)
	refreshToken, err := s.generateToken(user, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	// Store Refresh Token in Redis with 7-day TTL
	redisKey := fmt.Sprintf("auth:refresh:%s", user.ID)
	if err := s.redis.Set(ctx, redisKey, refreshToken, 7*24*time.Hour).Err(); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// 3. RefreshToken: Validate refresh token, check Redis, return new access token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshTokenStr string) (string, error) {
	// Parse and validate the refresh token
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid or expired refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("invalid user_id in token")
	}

	// Check if refresh token matches the one stored in Redis
	redisKey := fmt.Sprintf("auth:refresh:%s", userID)
	stored, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil || stored != refreshTokenStr {
		return "", errors.New("refresh token revoked or expired")
	}

	// Retrieve user from DB
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Generate new Access Token (15m)
	newAccessToken, err := s.generateToken(user, 15*time.Minute)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

// 4. Logout: Delete refresh token from Redis
func (s *AuthServiceImpl) Logout(ctx context.Context, userID string) error {
	redisKey := fmt.Sprintf("auth:refresh:%s", userID)
	return s.redis.Del(ctx, redisKey).Err()
}

// 5. GetMe: Get current user profile
func (s *AuthServiceImpl) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// --- Helper: Generate JWT with custom expiration ---
func (s *AuthServiceImpl) generateToken(user *domain.User, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(duration).Unix(),
	})
	return token.SignedString([]byte(s.jwtSecret))
}