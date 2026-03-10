package repositories

import (
	"context"
	"idp-api-gateway/internal/core/domain"

	"gorm.io/gorm"
)

type AdminRepositoryImpl struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepositoryImpl {
	return &AdminRepositoryImpl{db: db}
}

// GetStats returns system-wide statistics: total users, total jobs, and job counts by state.
func (r *AdminRepositoryImpl) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var totalUsers int64
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	var totalJobs int64
	if err := r.db.WithContext(ctx).Model(&domain.Job{}).Count(&totalJobs).Error; err != nil {
		return nil, err
	}

	// Group jobs by state
	type StateCount struct {
		State string
		Count int64
	}
	var stateCounts []StateCount
	if err := r.db.WithContext(ctx).
		Model(&domain.Job{}).
		Select("state, count(*) as count").
		Group("state").
		Scan(&stateCounts).Error; err != nil {
		return nil, err
	}

	jobsByState := make(map[string]int64)
	for _, sc := range stateCounts {
		jobsByState[sc.State] = sc.Count
	}

	return map[string]interface{}{
		"total_users":  totalUsers,
		"total_jobs":   totalJobs,
		"jobs_by_state": jobsByState,
	}, nil
}

// GetAllJobsWithUsers fetches all jobs ordered by created_at DESC with associated User info.
func (r *AdminRepositoryImpl) GetAllJobsWithUsers(ctx context.Context) ([]domain.Job, error) {
	var jobs []domain.Job
	if err := r.db.WithContext(ctx).
		Preload("User").
		Order("created_at DESC").
		Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

// GetAllUsers fetches all users ordered by created_at DESC. PasswordHash is hidden by json:"-" tag.
func (r *AdminRepositoryImpl) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
