package services

import (
	"context"
	"encoding/json"
	"fmt"
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
	"io"
	"log"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const jobCacheTTL = 5 * time.Minute

type IDPServiceImpl struct {
	repo    ports.DocumentRepository
	storage ports.FileStorage
	queue   ports.QueueProducer
	pubsub  ports.PubSubClient
	cache   ports.CacheClient
}

func NewIDPService(repo ports.DocumentRepository, storage ports.FileStorage, queue ports.QueueProducer, pubsub ports.PubSubClient, cache ports.CacheClient) *IDPServiceImpl {
	return &IDPServiceImpl{
		repo:    repo,
		storage: storage,
		queue:   queue,
		pubsub:  pubsub,
		cache:   cache,
	}
}

// UploadDocument handles file upload, DB persistence, queue publishing, and cache invalidation.
func (s *IDPServiceImpl) UploadDocument(ctx context.Context, userID uuid.UUID, filename string, fileSize int64, fileStream io.Reader, contentType string, fileCode string, notes string) (*domain.Job, error) {
	docID := uuid.New()
	jobID := uuid.New()

	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("%s%s", docID.String(), ext)

	// 1. Upload to MinIO
	err := s.storage.UploadFile(ctx, "documents", objectName, fileStream, fileSize, contentType)
	if err != nil {
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	// 2. Save Document metadata to DB
	doc := &domain.Document{
		ID:               docID,
		UserID:           userID,
		OriginalFilename: filename,
		FileName:         filename,
		FileCode:         fileCode,
		Notes:            notes,
		StorageBucket:    "documents",
		StoragePath:      objectName,
		MimeType:         contentType,
		FileSize:         fileSize,
		CreatedAt:        time.Now(),
	}
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("db create document failed: %w", err)
	}

	// 3. Create Job
	job := &domain.Job{
		ID:         jobID,
		UserID:     userID,
		DocumentID: docID,
		State:      "PENDING",
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("db create job failed: %w", err)
	}

	// 4. Publish to RabbitMQ
	if err := s.queue.PublishJob(ctx, jobID.String(), docID.String()); err != nil {
		return nil, fmt.Errorf("queue publish failed: %w", err)
	}

	// 5. Invalidate caches (new job added)
	s.invalidateJobCachesInternal(ctx, userID.String())

	return job, nil
}

func (s *IDPServiceImpl) GetJobStatus(ctx context.Context, userID uuid.UUID, jobID string) (*domain.Job, error) {
	return s.repo.GetJobByID(ctx, jobID, userID.String())
}

// GetUserJobs returns paginated jobs for a specific user with Redis cache-aside.
func (s *IDPServiceImpl) GetUserJobs(ctx context.Context, userID uuid.UUID, q domain.PaginationQuery) (*domain.PaginatedResponse, error) {
	// 1. Build cache key
	cacheKey := fmt.Sprintf("idp:jobs:user:%s:page:%d:limit:%d:status:%s:code:%s",
		userID.String(), q.Page, q.Limit, q.Status, q.FileCode)

	// 2. Try cache first
	var cached domain.PaginatedResponse
	if err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
		return &cached, nil // Cache hit
	} else if err != redis.Nil {
		log.Printf("⚠️ Cache read warning: %v", err)
	}

	// 3. Cache miss — query DB
	jobs, total, err := s.repo.GetJobsByUserID(ctx, userID.String(), q)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	result := domain.NewPaginatedResponse(jobs, total, q)

	// 4. Write to cache (fire-and-forget, don't block on cache errors)
	if err := s.cache.SetJSON(ctx, cacheKey, result, jobCacheTTL); err != nil {
		log.Printf("⚠️ Cache write warning: %v", err)
	}

	return result, nil
}

// StreamJobStatus verifies ownership, subscribes to real-time Redis updates,
// and invalidates paginated caches when a terminal state (COMPLETED/FAILED) is detected.
func (s *IDPServiceImpl) StreamJobStatus(ctx context.Context, userID uuid.UUID, jobID string) (<-chan string, error) {
	_, err := s.repo.GetJobByID(ctx, jobID, userID.String())
	if err != nil {
		return nil, fmt.Errorf("job not found or unauthorized: %w", err)
	}

	rawChan, err := s.pubsub.SubscribeJobStatus(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Intercept messages to trigger cache invalidation on terminal states
	outChan := make(chan string, 1)
	go func() {
		defer close(outChan)
		for msg := range rawChan {
			// Forward the message immediately
			select {
			case outChan <- msg:
			case <-ctx.Done():
				return
			}

			// Check for terminal state in the payload and invalidate caches
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(msg), &payload); err == nil {
				status, _ := payload["status"].(string)
				if status == "COMPLETED" || status == "FAILED" {
					log.Printf("🔄 Terminal state [%s] detected for job %s — invalidating caches", status, jobID)
					s.invalidateJobCachesInternal(context.Background(), userID.String())
				}
			}
		}
	}()

	return outChan, nil
}

// InvalidateJobCaches is the public method called by the internal webhook.
// It resolves job_id → user_id via DB lookup, then clears all related caches.
func (s *IDPServiceImpl) InvalidateJobCaches(ctx context.Context, jobID string) error {
	log.Printf("[CACHE INVALIDATION] Looking up JobID: %s to resolve UserID...", jobID)
	
	job, err := s.repo.GetJobByIDInternal(ctx, jobID)
	if err != nil {
		log.Printf("[CACHE INVALIDATION] ❌ DB Lookup failed for JobID %s: %v", jobID, err)
		return fmt.Errorf("job not found: %w", err)
	}
	
	log.Printf("[CACHE INVALIDATION] ✅ Fetched UserID: %s for JobID: %s", job.UserID.String(), jobID)
	s.invalidateJobCachesInternal(ctx, job.UserID.String())
	return nil
}

// invalidateJobCachesInternal clears all cached job list pages for a user and admin views.
func (s *IDPServiceImpl) invalidateJobCachesInternal(ctx context.Context, userID string) {
	patterns := []string{
		fmt.Sprintf("idp:jobs:user:%s:*", userID),
		"idp:jobs:admin:*",
	}
	for _, p := range patterns {
		if err := s.cache.DeleteByPattern(ctx, p); err != nil {
			log.Printf("⚠️ Cache invalidation warning for pattern %s: %v", p, err)
		}
	}
}