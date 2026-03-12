package services

import (
	"context"
	"fmt"
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const adminJobCacheTTL = 5 * time.Minute

type AdminServiceImpl struct {
	repo  ports.AdminRepository
	cache ports.CacheClient
}

func NewAdminService(repo ports.AdminRepository, cache ports.CacheClient) *AdminServiceImpl {
	return &AdminServiceImpl{repo: repo, cache: cache}
}

func (s *AdminServiceImpl) GetSystemStats(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetStats(ctx)
}

// GetAllJobs returns paginated, filtered admin job listing with Redis cache-aside.
func (s *AdminServiceImpl) GetAllJobs(ctx context.Context, q domain.PaginationQuery) (*domain.PaginatedResponse, error) {
	// 1. Build cache key
	cacheKey := fmt.Sprintf("idp:jobs:admin:page:%d:limit:%d:status:%s:code:%s",
		q.Page, q.Limit, q.Status, q.FileCode)

	// 2. Try cache first
	var cached domain.PaginatedResponse
	if err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	} else if err != redis.Nil {
		log.Printf("⚠️ Admin cache read warning: %v", err)
	}

	// 3. Cache miss — query DB
	jobs, total, err := s.repo.GetAllJobsWithUsers(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	result := domain.NewPaginatedResponse(jobs, total, q)

	// 4. Write to cache
	if err := s.cache.SetJSON(ctx, cacheKey, result, adminJobCacheTTL); err != nil {
		log.Printf("⚠️ Admin cache write warning: %v", err)
	}

	return result, nil
}

func (s *AdminServiceImpl) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAllUsers(ctx)
}
