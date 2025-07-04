package domain

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID
	Name        string
	Permissions []string // Slice of permission codes
}

type Permission struct {
	ID   uuid.UUID
	Code string
}
