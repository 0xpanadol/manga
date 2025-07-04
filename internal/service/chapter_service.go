package service

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/0xpanadol/manga/pkg/uploader"
	"github.com/google/uuid"
)

type ChapterService struct {
	chapterRepo repository.ChapterRepository
	uploader    *uploader.MinioUploader // Add uploader

}

func NewChapterService(chapterRepo repository.ChapterRepository, uploader *uploader.MinioUploader) *ChapterService {
	return &ChapterService{
		chapterRepo: chapterRepo,
		uploader:    uploader,
	}
}

func (s *ChapterService) Create(ctx context.Context, chapter *domain.Chapter) error {
	return s.chapterRepo.Create(ctx, chapter)
}

func (s *ChapterService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Chapter, error) {
	return s.chapterRepo.FindByID(ctx, id)
}

func (s *ChapterService) ListByMangaID(ctx context.Context, params repository.ListChaptersParams) ([]*domain.Chapter, error) {
	return s.chapterRepo.ListByMangaID(ctx, params)
}

func (s *ChapterService) Update(ctx context.Context, chapter *domain.Chapter) error {
	return s.chapterRepo.Update(ctx, chapter)
}

func (s *ChapterService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.chapterRepo.Delete(ctx, id)
}

func (s *ChapterService) UploadPages(ctx context.Context, chapterID uuid.UUID, files []*multipart.FileHeader) error {
	// 1. Fetch the chapter to get the existing pages
	chapter, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return err // e.g., ErrChapterNotFound
	}

	var newPageURLs []string
	// 2. Iterate over the files and upload them
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
		}
		defer file.Close()

		url, err := s.uploader.UploadFile(ctx, file, fileHeader.Size, fileHeader.Filename)
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %w", fileHeader.Filename, err)
		}
		newPageURLs = append(newPageURLs, url)
	}

	// 3. Combine old and new pages and update the database
	allPages := append(chapter.Pages, newPageURLs...)
	return s.chapterRepo.UpdatePages(ctx, chapterID, allPages)
}
