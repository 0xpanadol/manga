package service

import (
	"context"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/google/uuid"
)

type MangaService struct {
	mangaRepo repository.MangaRepository
}

func NewMangaService(mangaRepo repository.MangaRepository) *MangaService {
	return &MangaService{mangaRepo: mangaRepo}
}

func (s *MangaService) Create(ctx context.Context, manga *domain.Manga) error {
	return s.mangaRepo.Create(ctx, manga)
}

func (s *MangaService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Manga, error) {
	return s.mangaRepo.FindByID(ctx, id)
}

func (s *MangaService) List(ctx context.Context, params repository.ListMangaParams) ([]*domain.Manga, error) {
	return s.mangaRepo.List(ctx, params)
}

func (s *MangaService) Update(ctx context.Context, manga *domain.Manga) error {
	// You could add business logic here, e.g., checking if the manga exists first.
	// For simplicity, we'll let the repository handle the ErrMangaNotFound.
	return s.mangaRepo.Update(ctx, manga)
}

func (s *MangaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.mangaRepo.Delete(ctx, id)
}
