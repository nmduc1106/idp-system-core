# Database Design & GORM Guidelines (IDP System)

This document defines the mandatory conventions for adding or modifying Database Schemas (PostgreSQL) and Go Models (GORM) in the IDP System. These rules prevent migration conflicts and ensure system consistency.

## 1. PostgreSQL Rules (`database/init.sql`)

When creating a new table, strictly adhere to the following:

- **Primary Key (ID):** Always use the `UUID` data type with the default value `uuid_generate_v4()`.
- **Explicit Constraint Names (Mandatory):** Never use the standalone `UNIQUE` keyword. You must explicitly name constraints using the syntax: `CONSTRAINT <constraint_name> UNIQUE (<column_name>)`.
- **Naming Conventions:**
  - Unique Constraints: `uni_<table_name>_<column_name>` (e.g., `uni_users_email`).
  - Indexes: `idx_<table_name>_<column_name>` (e.g., `idx_jobs_user_id`).
- **Default System Columns:** Every table (where applicable) should include:
  - `created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP`
  - `updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP`
  - `is_deleted BOOLEAN DEFAULT FALSE` (For Soft Delete support).

## 2. Go Models & GORM Rules (`internal/core/domain/`)

Go structs must map exactly 1-to-1 with the `init.sql` schema:

- **ID Data Type:** Use `uuid.UUID` from the `github.com/google/uuid` package.
- **Standard GORM Tag for ID:** `` `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"` ``
- **Constraint Synchronization (Mandatory):** If a column has a Unique Constraint in SQL, you MUST declare that exact name in the GORM tag using `uniqueIndex:<constraint_name>`.
  - *Example:* `` `gorm:"uniqueIndex:uni_users_email"` ``
- **Foreign Keys:** Declare as `uuid.UUID` and explicitly specify if it cannot be null: `` `gorm:"type:uuid;not null"` ``.

## 3. Standard Template Example

### Step 1: SQL Declaration (`init.sql`)
```sql
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) NOT NULL,
    CONSTRAINT uni_organizations_code UNIQUE (code), -- Explicit name
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Step 2: Model Declaration (domain/models.go)
```go
import (
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Organization struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    Code      string    `gorm:"uniqueIndex:uni_organizations_code;not null" json:"code"`
    Name      string    `gorm:"not null" json:"name"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
```