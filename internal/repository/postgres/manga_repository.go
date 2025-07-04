package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMangaRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresMangaRepository(db *pgxpool.Pool) *PostgresMangaRepository {
	return &PostgresMangaRepository{DB: db}
}

// Create inserts a new manga and its genre associations into the database.
func (r *PostgresMangaRepository) Create(ctx context.Context, manga *domain.Manga) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback is a no-op if the transaction is committed

	// 1. Insert into manga table
	mangaQuery := `
        INSERT INTO manga (title, description, author, status, cover_image_url)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at`
	err = tx.QueryRow(ctx, mangaQuery, manga.Title, manga.Description, manga.Author, manga.Status, manga.CoverImageURL).Scan(
		&manga.ID,
		&manga.CreatedAt,
		&manga.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert manga: %w", err)
	}

	// 2. Associate genres
	if len(manga.Genres) > 0 {
		genreQuery := `
            INSERT INTO manga_genres (manga_id, genre_id)
            SELECT $1, id FROM genres WHERE name = ANY($2)`
		_, err = tx.Exec(ctx, genreQuery, manga.ID, manga.Genres)
		if err != nil {
			return fmt.Errorf("failed to associate genres: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// FindByID retrieves a manga and its genres by ID.
func (r *PostgresMangaRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Manga, error) {
	query := `
        SELECT
            m.id, m.title, m.description, m.author, m.status, m.cover_image_url,
            m.created_at, m.updated_at,
            COALESCE(array_agg(g.name) FILTER (WHERE g.name IS NOT NULL), '{}') AS genres
        FROM manga m
        LEFT JOIN manga_genres mg ON m.id = mg.manga_id
        LEFT JOIN genres g ON mg.genre_id = g.id
        WHERE m.id = $1
        GROUP BY m.id`

	var manga domain.Manga
	err := r.DB.QueryRow(ctx, query, id).Scan(
		&manga.ID, &manga.Title, &manga.Description, &manga.Author, &manga.Status, &manga.CoverImageURL,
		&manga.CreatedAt, &manga.UpdatedAt, &manga.Genres,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrMangaNotFound
		}
		return nil, fmt.Errorf("failed to find manga by id: %w", err)
	}

	return &manga, nil
}

// List retrieves a paginated and filtered list of manga.
func (r *PostgresMangaRepository) List(ctx context.Context, params repository.ListMangaParams) ([]*domain.Manga, error) {
	// Base query
	query := `
        SELECT
            m.id, m.title, m.description, m.author, m.status, m.cover_image_url,
            m.created_at, m.updated_at,
            COALESCE(array_agg(g.name) FILTER (WHERE g.name IS NOT NULL), '{}') AS genres
        FROM manga m
        LEFT JOIN manga_genres mg ON m.id = mg.manga_id
        LEFT JOIN genres g ON mg.genre_id = g.id
    `
	var conditions []string
	var args []interface{}
	argID := 1

	// Filtering by Title (simple ILIKE search)
	if params.Title != "" {
		conditions = append(conditions, fmt.Sprintf("m.title ILIKE $%d", argID))
		args = append(args, "%"+params.Title+"%")
		argID++
	}

	// Filtering by Status
	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("m.status = $%d", argID))
		args = append(args, params.Status)
		argID++
	}

	// Filtering by Genres (manga must have ALL specified genres)
	if len(params.Genres) > 0 {
		conditions = append(conditions, fmt.Sprintf(`(
            SELECT count(g.id) FROM genres g
            JOIN manga_genres mg_sub ON g.id = mg_sub.genre_id
            WHERE mg_sub.manga_id = m.id AND g.name = ANY($%d)
        ) = %d`, argID, len(params.Genres)))
		args = append(args, params.Genres)
		argID++
	}

	// Construct WHERE clause
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Grouping is always needed for the genre aggregation
	query += " GROUP BY m.id"

	// Sorting
	if params.SortBy != "" {
		// Whitelist sortable columns to prevent SQL injection
		validSortBy := map[string]bool{"title": true, "created_at": true, "updated_at": true}
		if validSortBy[params.SortBy] {
			order := "ASC"
			if strings.ToLower(params.SortOrder) == "desc" {
				order = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY m.%s %s", params.SortBy, order)
		}
	} else {
		query += " ORDER BY m.created_at DESC" // Default sort
	}

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list manga: %w", err)
	}
	defer rows.Close()

	var mangas []*domain.Manga
	for rows.Next() {
		var manga domain.Manga
		if err := rows.Scan(
			&manga.ID, &manga.Title, &manga.Description, &manga.Author, &manga.Status, &manga.CoverImageURL,
			&manga.CreatedAt, &manga.UpdatedAt, &manga.Genres,
		); err != nil {
			return nil, fmt.Errorf("failed to scan manga row: %w", err)
		}
		mangas = append(mangas, &manga)
	}

	return mangas, nil
}

// Update modifies an existing manga's details and genre associations.
func (r *PostgresMangaRepository) Update(ctx context.Context, manga *domain.Manga) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Update the manga table
	mangaQuery := `
        UPDATE manga
        SET title = $1, description = $2, author = $3, status = $4, cover_image_url = $5, updated_at = now()
        WHERE id = $6
        RETURNING updated_at`
	err = tx.QueryRow(ctx, mangaQuery, manga.Title, manga.Description, manga.Author, manga.Status, manga.CoverImageURL, manga.ID).Scan(&manga.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrMangaNotFound
		}
		return fmt.Errorf("failed to update manga: %w", err)
	}

	// 2. Remove all existing genre associations for this manga
	_, err = tx.Exec(ctx, "DELETE FROM manga_genres WHERE manga_id = $1", manga.ID)
	if err != nil {
		return fmt.Errorf("failed to clear existing genres: %w", err)
	}

	// 3. Add the new genre associations
	if len(manga.Genres) > 0 {
		genreQuery := `
            INSERT INTO manga_genres (manga_id, genre_id)
            SELECT $1, id FROM genres WHERE name = ANY($2)`
		_, err = tx.Exec(ctx, genreQuery, manga.ID, manga.Genres)
		if err != nil {
			return fmt.Errorf("failed to associate new genres: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Delete removes a manga from the database.
func (r *PostgresMangaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM manga WHERE id = $1"
	cmdTag, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete manga: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrMangaNotFound
	}
	return nil
}
