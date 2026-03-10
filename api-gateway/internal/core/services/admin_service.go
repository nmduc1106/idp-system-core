package services

import (
	"context"
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
)

type AdminServiceImpl struct {
	repo ports.AdminRepository
}

func NewAdminService(repo ports.AdminRepository) *AdminServiceImpl {
	return &AdminServiceImpl{repo: repo}
}

func (s *AdminServiceImpl) GetSystemStats(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetStats(ctx)
}

func (s *AdminServiceImpl) GetAllJobs(ctx context.Context) ([]domain.Job, error) {
	return s.repo.GetAllJobsWithUsers(ctx)
}

func (s *AdminServiceImpl) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAllUsers(ctx)
}
