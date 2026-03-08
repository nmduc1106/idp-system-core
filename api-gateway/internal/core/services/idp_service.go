package services

import (
	"context"
	"fmt"
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type IDPServiceImpl struct {
	repo    ports.DocumentRepository
	storage ports.FileStorage
	queue   ports.QueueProducer
}

func NewIDPService(repo ports.DocumentRepository, storage ports.FileStorage, queue ports.QueueProducer) *IDPServiceImpl {
	return &IDPServiceImpl{
		repo:    repo,
		storage: storage,
		queue:   queue,
	}
}

// UploadDocument: Đã được sửa để nhận userID thật
func (s *IDPServiceImpl) UploadDocument(ctx context.Context, userID uuid.UUID, filename string, fileSize int64, fileStream io.Reader, contentType string) (*domain.Job, error) {
	// 1. Business Logic: Tạo ID
	docID := uuid.New()
	jobID := uuid.New()

	// [QUAN TRỌNG] Đã xóa dummyUserID (0000...)
	// Bây giờ hệ thống dùng userID được truyền vào từ tham số

	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("%s%s", docID.String(), ext)

	// 2. Upload lên MinIO (Stream trực tiếp)
	err := s.storage.UploadFile(ctx, "documents", objectName, fileStream, fileSize, contentType)
	if err != nil {
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	// 3. Lưu Metadata vào DB
	doc := &domain.Document{
		ID:               docID,
		UserID:           userID, // [FIX] Dùng UserID thật
		OriginalFilename: filename,
		StorageBucket:    "documents",
		StoragePath:      objectName,
		MimeType:         contentType,
		FileSize:         fileSize,
		CreatedAt:        time.Now(),
	}
	// Lưu Document trước
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("db create document failed: %w", err)
	}

	// Tạo Job
	job := &domain.Job{
		ID:         jobID,
		UserID:     userID,
		DocumentID: docID,
		State:      "PENDING",
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}
	// Lưu Job sau
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("db create job failed: %w", err)
	}

	// 4. Bắn sự kiện sang RabbitMQ
	if err := s.queue.PublishJob(ctx, jobID.String(), docID.String()); err != nil {
		// Note: Có thể implement retry logic ở đây
		return nil, fmt.Errorf("queue publish failed: %w", err)
	}

	return job, nil
}

func (s *IDPServiceImpl) GetJobStatus(ctx context.Context, userID uuid.UUID, jobID string) (*domain.Job, error) {
	return s.repo.GetJobByID(ctx, jobID, userID.String())
}