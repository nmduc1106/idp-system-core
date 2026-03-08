package domain

import "time"

type User struct {
    ID           string    `json:"id"`
    Email        string    `gorm:"uniqueIndex" json:"email"` // <--- Thêm gorm:"uniqueIndex" vào đây
    PasswordHash string    `json:"-"`
    FullName     string    `json:"full_name"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}