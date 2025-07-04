package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository is the PostgreSQL implementation of the UserRepository interface.
type PostgresUserRepository struct {
	DB *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgresUserRepository.
func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{DB: db}
}

// Create inserts a new user into the database.
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (username, email, password_hash, role_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at`

	err := r.DB.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash, user.RoleID).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: user with this email or username already exists", apperrors.ErrConflict) // Use new error
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByEmail retrieves a user by their email address.
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `
        SELECT id, username, email, password_hash, role_id, created_at, updated_at
        FROM users
        WHERE email = $1`

	err := r.DB.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: user not found", apperrors.ErrNotFound) // Use new error
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, nil
}

// FindDefaultUserRoleID retrieves the UUID for the "User" role.
func (r *PostgresUserRepository) FindDefaultUserRoleID(ctx context.Context) (uuid.UUID, error) {
	var roleID uuid.UUID
	query := `SELECT id FROM roles WHERE name = 'User' LIMIT 1`

	err := r.DB.QueryRow(ctx, query).Scan(&roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This is a critical failure, the system can't register users without this role.
			return uuid.Nil, errors.New("default 'User' role not found in database")
		}
		return uuid.Nil, fmt.Errorf("failed to find default user role ID: %w", err)
	}

	return roleID, nil
}

// FindByID retrieves a user by their ID.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `
        SELECT id, username, email, password_hash, role_id, created_at, updated_at
        FROM users
        WHERE id = $1`

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: user not found", apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return &user, nil
}

// GetRoleAndPermissions retrieves a user's role and all associated permission codes.
func (r *PostgresUserRepository) GetRoleAndPermissions(ctx context.Context, userID uuid.UUID) (*domain.Role, error) {
	query := `
        SELECT r.id, r.name, array_agg(p.code) AS permissions
        FROM users u
        JOIN roles r ON u.role_id = r.id
        LEFT JOIN roles_permissions rp ON r.id = rp.role_id
        LEFT JOIN permissions p ON rp.permission_id = p.id
        WHERE u.id = $1
        GROUP BY r.id, r.name`

	var role domain.Role
	// array_agg can return a NULL value if there are no permissions, which pgx needs to handle.
	var permissions []string
	err := r.DB.QueryRow(ctx, query, userID).Scan(&role.ID, &role.Name, &permissions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: failed to get role and permissions", apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get role and permissions: %w", err)
	}
	role.Permissions = permissions

	return &role, nil
}

func (r *PostgresUserRepository) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	query := `INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}
	return nil
}
