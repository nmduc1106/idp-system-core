package ports

import (
	"context"
	"idp-api-gateway/internal/core/domain"
	"io"

	"github.com/google/uuid"
)

// --- ĐÃ XÓA UserRepository và AuthService (Vì đã có bên auth_ports.go) ---

// DocumentRepository: Giao tiếp với Database (Bảng Documents, Jobs)
type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *domain.Document) error
	CreateJob(ctx context.Context, job *domain.Job) error
	GetJobByID(ctx context.Context, id string, userID string) (*domain.Job, error)
}

// FileStorage: Giao tiếp với MinIO/S3
type FileStorage interface {
	UploadFile(ctx context.Context, bucket string, objectName string, fileStream io.Reader, fileSize int64, contentType string) error
}

// QueueProducer: Giao tiếp với RabbitMQ
type QueueProducer interface {
	PublishJob(ctx context.Context, jobID string, docID string) error
}

// IDPService: Interface chính cho Handler gọi vào
type IDPService interface {
	// UploadDocument: Nhận userID từ Handler
	UploadDocument(ctx context.Context, userID uuid.UUID, filename string, fileSize int64, fileStream io.Reader, contentType string) (*domain.Job, error)
	GetJobStatus(ctx context.Context, userID uuid.UUID, jobID string) (*domain.Job, error)
}