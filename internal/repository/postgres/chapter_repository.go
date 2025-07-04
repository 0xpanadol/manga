package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresChapterRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresChapterRepository(db *pgxpool.Pool) *PostgresChapterRepository {
	return &PostgresChapterRepository{DB: db}
}

func (r *PostgresChapterRepository) Create(ctx context.Context, chapter *domain.Chapter) error {
	query := `
        INSERT INTO chapters (manga_id, chapter_number, title, pages)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at`

	err := r.DB.QueryRow(ctx, query, chapter.MangaID, chapter.ChapterNumber, chapter.Title, chapter.Pages).Scan(
		&chapter.ID,
		&chapter.CreatedAt,
		&chapter.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrChapterAlreadyExists
		}
		return fmt.Errorf("failed to create chapter: %w", err)
	}
	return nil
}

func (r *PostgresChapterRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Chapter, error) {
	var chapter domain.Chapter
	query := `SELECT id, manga_id, chapter_number, title, pages, created_at, updated_at FROM chapters WHERE id = $1`

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&chapter.ID, &chapter.MangaID, &chapter.ChapterNumber, &chapter.Title, &chapter.Pages, &chapter.CreatedAt, &chapter.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrChapterNotFound
		}
		return nil, fmt.Errorf("failed to find chapter by id: %w", err)
	}
	return &chapter, nil
}

func (r *PostgresChapterRepository) ListByMangaID(ctx context.Context, params repository.ListChaptersParams) ([]*domain.Chapter, error) {
	query := `
        SELECT id, manga_id, chapter_number, title, pages, created_at, updated_at
        FROM chapters
        WHERE manga_id = $1
        ORDER BY chapter_number DESC -- Or however you want to sort
        LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, query, params.MangaID, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chapters: %w", err)
	}
	defer rows.Close()

	var chapters []*domain.Chapter
	for rows.Next() {
		var chapter domain.Chapter
		if err := rows.Scan(
			&chapter.ID, &chapter.MangaID, &chapter.ChapterNumber, &chapter.Title, &chapter.Pages, &chapter.CreatedAt, &chapter.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chapter row: %w", err)
		}
		chapters = append(chapters, &chapter)
	}
	return chapters, nil
}

func (r *PostgresChapterRepository) Update(ctx context.Context, chapter *domain.Chapter) error {
	query := `
        UPDATE chapters
        SET chapter_number = $1, title = $2, pages = $3, updated_at = now()
        WHERE id = $4
        RETURNING updated_at`

	err := r.DB.QueryRow(ctx, query, chapter.ChapterNumber, chapter.Title, chapter.Pages, chapter.ID).Scan(&chapter.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrChapterNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrChapterAlreadyExists
		}
		return fmt.Errorf("failed to update chapter: %w", err)
	}
	return nil
}

func (r *PostgresChapterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cmdTag, err := r.DB.Exec(ctx, "DELETE FROM chapters WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrChapterNotFound
	}
	return nil
}

func (r *PostgresChapterRepository) UpdatePages(ctx context.Context, id uuid.UUID, pages []string) error {
	query := `UPDATE chapters SET pages = $1, updated_at = now() WHERE id = $2`
	cmdTag, err := r.DB.Exec(ctx, query, pages, id)
	if err != nil {
		return fmt.Errorf("failed to update chapter pages: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrChapterNotFound
	}
	return nil
}
