package ports

import (
	"bytes"
	"context"
	"idp-api-gateway/internal/core/domain"
	"io"
	"time"

	"github.com/google/uuid"
)

// DocumentRepository: Giao tiếp với Database (Bảng Documents, Jobs)
type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *domain.Document) error
	CreateJob(ctx context.Context, job *domain.Job) error
	GetJobByID(ctx context.Context, id string, userID string) (*domain.Job, error)
	GetJobByIDInternal(ctx context.Context, id string) (*domain.Job, error) // No user filter (for internal webhook)
	GetJobsByUserID(ctx context.Context, userID string, q domain.PaginationQuery) ([]domain.Job, int64, error)
	GetCompletedJobsForExport(ctx context.Context, userID string, searchCode string) ([]domain.Job, error)
}

// FileStorage: Giao tiếp với MinIO/S3
type FileStorage interface {
	UploadFile(ctx context.Context, bucket string, objectName string, fileStream io.Reader, fileSize int64, contentType string) error
}

// QueueProducer: Giao tiếp với RabbitMQ
type QueueProducer interface {
	PublishJob(ctx context.Context, jobID string, docID string) error
}

// PubSubClient: Giao tiếp với Redis Pub/Sub (Real-time SSE)
type PubSubClient interface {
	SubscribeJobStatus(ctx context.Context, jobID string) (<-chan string, error)
}

// CacheClient: Giao tiếp với Redis Cache layer
type CacheClient interface {
	SetJSON(ctx context.Context, key string, v interface{}, ttl time.Duration) error
	GetJSON(ctx context.Context, key string, dest interface{}) error
	DeleteByPattern(ctx context.Context, pattern string) error
}

// IDPService: Interface chính cho Handler gọi vào
type IDPService interface {
	UploadDocument(ctx context.Context, userID uuid.UUID, filename string, fileSize int64, fileStream io.Reader, contentType string, fileCode string, notes string) (*domain.Job, error)
	GetJobStatus(ctx context.Context, userID uuid.UUID, jobID string) (*domain.Job, error)
	GetUserJobs(ctx context.Context, userID uuid.UUID, q domain.PaginationQuery) (*domain.PaginatedResponse, error)
	StreamJobStatus(ctx context.Context, userID uuid.UUID, jobID string) (<-chan string, error)
	InvalidateJobCaches(ctx context.Context, jobID string) error // Called by internal webhook
	ExportJobsToExcel(ctx context.Context, userID string, searchCode string) (*bytes.Buffer, error)
}

// AdminRepository: Giao tiếp với Database cho Admin queries
type AdminRepository interface {
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetAllJobsWithUsers(ctx context.Context, q domain.PaginationQuery) ([]domain.Job, int64, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
}

// AdminService: Interface cho Admin Handler gọi vào
type AdminService interface {
	GetSystemStats(ctx context.Context) (map[string]interface{}, error)
	GetAllJobs(ctx context.Context, q domain.PaginationQuery) (*domain.PaginatedResponse, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
}