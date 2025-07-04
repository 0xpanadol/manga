package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool" // Use pgxpool
)

// Corrected struct definition
type PostgresSocialRepository struct {
	DB *pgxpool.Pool
}

// Corrected constructor
func NewPostgresSocialRepository(db *pgxpool.Pool) *PostgresSocialRepository {
	return &PostgresSocialRepository{DB: db}
}

func (r *PostgresSocialRepository) ToggleFavorite(ctx context.Context, userID, mangaID uuid.UUID) (*repository.ToggleFavoriteResult, error) {
	// This logic is transactional and should be wrapped.
	// For simplicity and since it's idempotent, we can run as separate queries.
	// A more robust solution would use a transaction.

	// First, try to delete. If rows are affected, it was a favorite and is now removed.
	deleteQuery := "DELETE FROM user_favorites WHERE user_id = $1 AND manga_id = $2"
	cmdTag, err := r.DB.Exec(ctx, deleteQuery, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute delete from favorites: %w", err)
	}

	if cmdTag.RowsAffected() > 0 {
		return &repository.ToggleFavoriteResult{IsFavorited: false}, nil
	}

	// If no rows were deleted, it means it wasn't a favorite. So, we insert.
	insertQuery := "INSERT INTO user_favorites (user_id, manga_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	_, err = r.DB.Exec(ctx, insertQuery, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert into favorites: %w", err)
	}

	return &repository.ToggleFavoriteResult{IsFavorited: true}, nil
}

func (r *PostgresSocialRepository) ListFavorites(ctx context.Context, userID uuid.UUID, params repository.ListMangaParams) ([]*domain.Manga, error) {
	query := `
        SELECT
            m.id, m.title, m.description, m.author, m.status, m.cover_image_url,
            m.created_at, m.updated_at,
            COALESCE(array_agg(g.name) FILTER (WHERE g.name IS NOT NULL), '{}') AS genres
        FROM manga m
        JOIN user_favorites uf ON m.id = uf.manga_id
        LEFT JOIN manga_genres mg ON m.id = mg.manga_id
        LEFT JOIN genres g ON mg.genre_id = g.id
        WHERE uf.user_id = $1
    `
	args := []interface{}{userID}
	argID := 2

	query += " GROUP BY m.id, uf.created_at ORDER BY uf.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list favorite manga: %w", err)
	}
	defer rows.Close()

	var mangas []*domain.Manga
	for rows.Next() {
		var manga domain.Manga
		if err := rows.Scan(
			&manga.ID, &manga.Title, &manga.Description, &manga.Author, &manga.Status, &manga.CoverImageURL,
			&manga.CreatedAt, &manga.UpdatedAt, &manga.Genres,
		); err != nil {
			return nil, fmt.Errorf("failed to scan favorite manga row: %w", err)
		}
		mangas = append(mangas, &manga)
	}
	return mangas, nil
}

func (r *PostgresSocialRepository) MarkChapterAsRead(ctx context.Context, userID, chapterID uuid.UUID) error {
	query := "INSERT INTO user_reading_progress (user_id, chapter_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	_, err := r.DB.Exec(ctx, query, userID, chapterID)
	if err != nil {
		return fmt.Errorf("failed to mark chapter as read: %w", err)
	}
	return nil
}

func (r *PostgresSocialRepository) ListReadChapters(ctx context.Context, userID uuid.UUID) ([]*domain.Chapter, error) {
	query := `
        SELECT c.id, c.manga_id, c.chapter_number, c.title, c.pages, c.created_at, c.updated_at
        FROM chapters c
        JOIN user_reading_progress urp ON c.id = urp.chapter_id
        WHERE urp.user_id = $1
        ORDER BY urp.created_at DESC`

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list read chapters: %w", err)
	}
	defer rows.Close()

	var chapters []*domain.Chapter
	for rows.Next() {
		var chapter domain.Chapter
		if err := rows.Scan(
			&chapter.ID, &chapter.MangaID, &chapter.ChapterNumber, &chapter.Title, &chapter.Pages, &chapter.CreatedAt, &chapter.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan read chapter row: %w", err)
		}
		chapters = append(chapters, &chapter)
	}
	return chapters, nil
}

func (r *PostgresSocialRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	query := `
        INSERT INTO comments (user_id, manga_id, chapter_id, content)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at`

	err := r.DB.QueryRow(ctx, query, comment.UserID, comment.MangaID, comment.ChapterID, comment.Content).Scan(
		&comment.ID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

func (r *PostgresSocialRepository) ListComments(ctx context.Context, params repository.ListCommentsParams) ([]*domain.CommentWithUser, error) {
	// Base query
	query := `
        SELECT
            c.id, c.user_id, c.manga_id, c.chapter_id, c.content, c.created_at, c.updated_at,
            u.username
        FROM comments c
        JOIN users u ON c.user_id = u.id
    `
	var args []interface{}
	var conditions []string

	// Polymorphic condition
	if params.ParentType == domain.ParentTypeManga {
		conditions = append(conditions, "c.manga_id = $1")
		args = append(args, params.ParentID)
	} else if params.ParentType == domain.ParentTypeChapter {
		conditions = append(conditions, "c.chapter_id = $1")
		args = append(args, params.ParentID)
	} else {
		return nil, errors.New("invalid comment parent type")
	}

	query += " WHERE " + strings.Join(conditions, " AND ")
	query += " ORDER BY c.created_at DESC"

	// Add pagination
	argID := len(args) + 1
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []*domain.CommentWithUser
	for rows.Next() {
		var comment domain.CommentWithUser
		if err := rows.Scan(
			&comment.ID, &comment.UserID, &comment.MangaID, &comment.ChapterID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
			&comment.Username,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}
