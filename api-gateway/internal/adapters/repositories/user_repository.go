package repositories
import (
	"context"
	"database/sql"
	"fmt"
	"idp-api-gateway/internal/core/domain"
)

type UserRepositoryPostgres struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{db: db}
}

func (r *UserRepositoryPostgres) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *UserRepositoryPostgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, full_name, role FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, full_name, role, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	// PasswordHash rỗng để bảo mật
	return &user, nil
}