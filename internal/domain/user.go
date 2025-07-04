package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	RoleID       uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
