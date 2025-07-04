package domain

import (
	"time"

	"github.com/google/uuid"
)

type Chapter struct {
	ID            uuid.UUID
	MangaID       uuid.UUID
	ChapterNumber string
	Title         *string // Optional
	Pages         []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
