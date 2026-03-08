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
	if err := r.db.WithContext(ctx).First(&job, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}