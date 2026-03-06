package domain

import (
	"time"
	"github.com/google/uuid"
)

// Document đại diện cho file tài liệu
type Document struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid" json:"user_id"` // <--- Đảm bảo có dòng này
	OriginalFilename string    `json:"original_filename"`
	StorageBucket    string    `json:"-"` // Không trả về json
	StoragePath      string    `json:"-"`
	MimeType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	CreatedAt        time.Time `json:"created_at"`
}

type Job struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	DocumentID uuid.UUID `gorm:"type:uuid" json:"document_id"`
	State      string    `json:"state"`
	
	// SỬA DÒNG DƯỚI ĐÂY: Thêm dấu * trước string
	Result     *string   `gorm:"type:jsonb" json:"result,omitempty"` 
	
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
}