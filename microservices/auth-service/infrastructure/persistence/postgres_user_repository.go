package persistence

import (
	"context"
	"database/sql"
	"microservices/auth-service/domain/entities"
)

type PostgresUserRepository struct {
	db *sql.DB
}

// FindByID implements repositories.UserRepository.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	panic("unimplemented")
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *entities.User) error {
	query := `INSERT INTO users (id, email, password_hash, role, created_at) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		string(user.Role),
		user.CreatedAt,
	)
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `SELECT id, email, password_hash, role, created_at 
              FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)
	return scanUser(row)
}

// Helper untuk scan row SQL ke struct User
func scanUser(row *sql.Row) (*entities.User, error) {
	var user entities.User
	var roleStr string
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&roleStr,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	user.Role = entities.Role(roleStr)
	return &user, nil
}
