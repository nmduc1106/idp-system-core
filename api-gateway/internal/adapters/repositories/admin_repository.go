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

// GetAllJobsWithUsers fetches paginated, filtered jobs with User and Document associations.
func (r *AdminRepositoryImpl) GetAllJobsWithUsers(ctx context.Context, q domain.PaginationQuery) ([]domain.Job, int64, error) {
	var jobs []domain.Job
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Job{})

	// Apply optional filters
	if q.Status != "" {
		query = query.Where("state = ?", q.Status)
	}
	if q.FileCode != "" {
		query = query.Joins("JOIN documents ON documents.id = jobs.document_id").
			Where("documents.file_code ILIKE ?", "%"+q.FileCode+"%")
	}

	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results with associations
	if err := query.
		Preload("User").
		Preload("Document").
		Order("created_at DESC").
		Offset(q.Offset()).
		Limit(q.Limit).
		Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
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
