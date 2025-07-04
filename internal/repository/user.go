package repository

import (
	"context"
	"errors"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email or username already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindDefaultUserRoleID(ctx context.Context) (uuid.UUID, error)
	GetRoleAndPermissions(ctx context.Context, userID uuid.UUID) (*domain.Role, error)
}
