package services

import (
	"bytes"
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
	"github.com/xuri/excelize/v2"
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

// ExportJobsToExcel generates an Excel file from completed jobs without leaking DB IDs.
func (s *IDPServiceImpl) ExportJobsToExcel(ctx context.Context, userID string, searchCode string) (*bytes.Buffer, error) {
	jobs, err := s.repo.GetCompletedJobsForExport(ctx, userID, searchCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs for export: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Sheet1"

	// Create a Header Row
	headers := []string{"STT", "Mã Hồ Sơ", "Tên File Gốc", "Ngày Tải Lên", "Tên Nhà Cung Cấp", "Mã Số Thuế", "Số Hóa Đơn", "Ngày Hóa Đơn", "Tổng Tiền"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	// Set Column Widths for readability
	f.SetColWidth(sheet, "A", "A", 5)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 25)
	f.SetColWidth(sheet, "D", "D", 22)
	f.SetColWidth(sheet, "E", "E", 35)
	f.SetColWidth(sheet, "F", "F", 20)
	f.SetColWidth(sheet, "G", "G", 20)
	f.SetColWidth(sheet, "H", "H", 15)
	f.SetColWidth(sheet, "I", "I", 15)

	// Bold style and Center alignment for header
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetRowStyle(sheet, 1, 1, style)

	for i, job := range jobs {
		row := i + 2
		stt := i + 1

		fileCode := ""
		fileName := ""
		if job.Document != nil {
			fileCode = job.Document.FileCode
			fileName = job.Document.OriginalFilename
			if fileName == "" {
				fileName = job.Document.FileName
			}
		}
		
		uploadDate := ""
		if !job.CreatedAt.IsZero() {
			uploadDate = job.CreatedAt.Format("2006-01-02 15:04:05")
		}

		// Parse the JSON Result
		var extractedData map[string]interface{}
		if len(job.Result) > 0 {
			var resultPayload map[string]interface{}
			if err := json.Unmarshal(job.Result, &resultPayload); err == nil {
				// The Python worker wraps it in "extracted_data", gracefully handle both structures
				if data, ok := resultPayload["extracted_data"].(map[string]interface{}); ok {
					extractedData = data
				} else {
					extractedData = resultPayload
				}
			}
		}

		// Safe JSON nested field extractor
		getStr := func(key string) string {
			if extractedData == nil {
				return ""
			}
			if val, ok := extractedData[key]; ok && val != nil {
				return fmt.Sprintf("%v", val)
			}
			return ""
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), stt)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fileCode)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fileName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), uploadDate)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), getStr("vendor_name"))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), getStr("tax_id"))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), getStr("invoice_number"))
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), getStr("date"))
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), getStr("total_amount"))
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel to buffer: %w", err)
	}

	return &buf, nil
}