package domain

import (
	"time"

	"github.com/google/uuid"
)

type CommentParentType string

const (
	ParentTypeManga   CommentParentType = "manga"
	ParentTypeChapter CommentParentType = "chapter"
)

type Comment struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	MangaID   *uuid.UUID // Nullable
	ChapterID *uuid.UUID // Nullable
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CommentWithUser includes basic user info for display purposes.
type CommentWithUser struct {
	Comment
	Username string
}
