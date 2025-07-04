package service

import (
	"context"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/google/uuid"
)

type SocialService struct {
	socialRepo repository.SocialRepository
}

func NewSocialService(socialRepo repository.SocialRepository) *SocialService {
	return &SocialService{socialRepo: socialRepo}
}

func (s *SocialService) ToggleFavorite(ctx context.Context, userID, mangaID uuid.UUID) (*repository.ToggleFavoriteResult, error) {
	return s.socialRepo.ToggleFavorite(ctx, userID, mangaID)
}

func (s *SocialService) ListFavorites(ctx context.Context, userID uuid.UUID, params repository.ListMangaParams) ([]*domain.Manga, error) {
	return s.socialRepo.ListFavorites(ctx, userID, params)
}

func (s *SocialService) MarkChapterAsRead(ctx context.Context, userID, chapterID uuid.UUID) error {
	return s.socialRepo.MarkChapterAsRead(ctx, userID, chapterID)
}

func (s *SocialService) ListReadChapters(ctx context.Context, userID uuid.UUID) ([]*domain.Chapter, error) {
	return s.socialRepo.ListReadChapters(ctx, userID)
}

func (s *SocialService) CreateComment(ctx context.Context, comment *domain.Comment) error {
	// Business logic could go here, e.g., checking if the parent manga/chapter exists.
	// For now, we rely on the database's foreign key constraints.
	return s.socialRepo.CreateComment(ctx, comment)
}

func (s *SocialService) ListComments(ctx context.Context, params repository.ListCommentsParams) ([]*domain.CommentWithUser, error) {
	return s.socialRepo.ListComments(ctx, params)
}
