package repositories

import (
	"context"
	"idp-api-gateway/internal/core/domain"
	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateDocument(ctx context.Context, doc *domain.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// CreateJob persists the Job including its UserID field (ownership is set by the service layer).
func (r *PostgresRepository) CreateJob(ctx context.Context, job *domain.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

// GetJobByID returns a job only if it belongs to the requesting user (data isolation).
func (r *PostgresRepository) GetJobByID(ctx context.Context, id string, userID string) (*domain.Job, error) {
	var job domain.Job
	if err := r.db.WithContext(ctx).
		Preload("Document").
		First(&job, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// GetJobByIDInternal returns a job by ID without user filter (for internal use only, e.g., webhooks).
func (r *PostgresRepository) GetJobByIDInternal(ctx context.Context, id string) (*domain.Job, error) {
	var job domain.Job
	if err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// GetJobsByUserID returns paginated, filtered jobs for a specific user.
func (r *PostgresRepository) GetJobsByUserID(ctx context.Context, userID string, q domain.PaginationQuery) ([]domain.Job, int64, error) {
	var jobs []domain.Job
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Job{}).Where("user_id = ?", userID)

	// Apply optional filters
	if q.Status != "" {
		query = query.Where("state = ?", q.Status)
	}
	if q.FileCode != "" {
		// Join with documents to filter by file_code (ILIKE for partial match)
		query = query.Joins("JOIN documents ON documents.id = jobs.document_id").
			Where("documents.file_code ILIKE ?", "%"+q.FileCode+"%")
	}

	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	if err := query.
		Preload("Document").
		Order("created_at DESC").
		Offset(q.Offset()).
		Limit(q.Limit).
		Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

// GetCompletedJobsForExport returns all completed jobs without pagination for a specific user to be exported.
func (r *PostgresRepository) GetCompletedJobsForExport(ctx context.Context, userID string, searchCode string) ([]domain.Job, error) {
	var jobs []domain.Job
	query := r.db.WithContext(ctx).Model(&domain.Job{}).
		Where("user_id = ? AND state = ?", userID, "COMPLETED")

	if searchCode != "" {
		query = query.Joins("JOIN documents ON documents.id = jobs.document_id").
			Where("documents.file_code ILIKE ?", "%"+searchCode+"%")
	}

	if err := query.Preload("Document").Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}