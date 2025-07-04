package repository

import (
	"context"
	"errors"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrMangaNotFound = errors.New("manga not found")
)

// ListMangaParams defines the parameters for listing manga.
type ListMangaParams struct {
	Limit     int
	Offset    int
	Title     string
	Genres    []string
	Status    string
	SortBy    string // e.g., "title", "created_at"
	SortOrder string // "asc" or "desc"
}

type MangaRepository interface {
	Create(ctx context.Context, manga *domain.Manga) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Manga, error)
	List(ctx context.Context, params ListMangaParams) ([]*domain.Manga, error)
	Update(ctx context.Context, manga *domain.Manga) error
	Delete(ctx context.Context, id uuid.UUID) error
}
