package domain

import (
	"time"

	"github.com/google/uuid"
)

type MangaStatus string

const (
	StatusOngoing   MangaStatus = "ongoing"
	StatusCompleted MangaStatus = "completed"
	StatusHiatus    MangaStatus = "hiatus"
	StatusCancelled MangaStatus = "cancelled"
)

type Manga struct {
	ID            uuid.UUID
	Title         string
	Description   string
	Author        string
	Status        MangaStatus
	CoverImageURL *string // Use a pointer to handle NULL values
	Genres        []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
