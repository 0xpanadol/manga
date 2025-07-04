package repository

import (
	"context"
	"errors"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrChapterNotFound      = errors.New("chapter not found")
	ErrChapterAlreadyExists = errors.New("chapter with this number already exists for this manga")
)

type ListChaptersParams struct {
	MangaID uuid.UUID
	Limit   int
	Offset  int
}

type ChapterRepository interface {
	Create(ctx context.Context, chapter *domain.Chapter) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Chapter, error)
	ListByMangaID(ctx context.Context, params ListChaptersParams) ([]*domain.Chapter, error)
	Update(ctx context.Context, chapter *domain.Chapter) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePages(ctx context.Context, id uuid.UUID, pages []string) error
}
