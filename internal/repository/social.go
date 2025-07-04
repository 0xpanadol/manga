package repository

import (
	"context"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/google/uuid"
)

type ToggleFavoriteResult struct {
	IsFavorited bool // True if the manga is now a favorite, false if it was removed.
}

type ListCommentsParams struct {
	ParentID   uuid.UUID
	ParentType domain.CommentParentType
	Limit      int
	Offset     int
}

type SocialRepository interface {
	// Favorites
	ToggleFavorite(ctx context.Context, userID, mangaID uuid.UUID) (*ToggleFavoriteResult, error)
	ListFavorites(ctx context.Context, userID uuid.UUID, params ListMangaParams) ([]*domain.Manga, error)

	// Reading Progress
	MarkChapterAsRead(ctx context.Context, userID, chapterID uuid.UUID) error
	ListReadChapters(ctx context.Context, userID uuid.UUID) ([]*domain.Chapter, error)

	// Comments
	CreateComment(ctx context.Context, comment *domain.Comment) error
	ListComments(ctx context.Context, params ListCommentsParams) ([]*domain.CommentWithUser, error)
}
