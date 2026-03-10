package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)



type Document struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	OriginalFilename string    `json:"original_filename"`
	StorageBucket    string    `json:"-"`
	StoragePath      string    `json:"-"`
	MimeType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	IsDeleted        bool      `gorm:"default:false" json:"is_deleted"`
	CreatedAt        time.Time `json:"created_at"`
}

type Job struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID       uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	DocumentID   uuid.UUID       `gorm:"type:uuid;not null" json:"document_id"`
	State        string          `gorm:"default:PENDING" json:"state"`
	
	// [ĐÃ SỬA]: Chuyển từ []byte sang json.RawMessage để giữ nguyên định dạng JSON
	Result       json.RawMessage `gorm:"type:jsonb" json:"result,omitempty"` 
	
	RetryCount   int             `gorm:"default:0" json:"retry_count"`
	ErrorMessage string          `json:"error_message,omitempty"`
	TraceID      string          `json:"trace_id,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`

	// Association: GORM Preload for Admin queries (constraint:false to skip FK migration)
	User         *User           `gorm:"foreignKey:UserID;references:ID;constraint:false" json:"user,omitempty"`
}