package domain

import "time"

type User struct {
    ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
    Email        string    `gorm:"uniqueIndex" json:"email"`
    PasswordHash string    `json:"-"`
    FullName     string    `json:"full_name"`
    Role         string    `gorm:"type:varchar(20);default:'EMPLOYEE'" json:"role"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}