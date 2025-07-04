package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	cacheDuration = 5 * time.Minute
)

type MangaService struct {
	mangaRepo repository.MangaRepository
	redis     *redis.Client
}

func NewMangaService(
	mangaRepo repository.MangaRepository,
	redisClient *redis.Client,
) *MangaService {
	return &MangaService{
		mangaRepo: mangaRepo,
		redis:     redisClient,
	}
}

// getMangaCacheKey generates a consistent key for a manga.
func getMangaCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("manga:%s", id.String())
}

func (s *MangaService) Create(ctx context.Context, manga *domain.Manga) error {
	return s.mangaRepo.Create(ctx, manga)
}

// GetByID now implements the cache-aside pattern.
func (s *MangaService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Manga, error) {
	key := getMangaCacheKey(id)

	// 1. Try to get the manga from the cache
	cachedManga, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		// Cache hit!
		log.Println("CACHE HIT for key:", key)
		var manga domain.Manga
		if err := json.Unmarshal([]byte(cachedManga), &manga); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached manga: %w", err)
		}
		return &manga, nil
	}

	// If the error is anything other than "not found", it's a real error.
	if err != redis.Nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	// 2. Cache miss. Get the manga from the database.
	log.Println("CACHE MISS for key:", key)
	manga, err := s.mangaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err // e.g., repository.ErrMangaNotFound
	}

	// 3. Store the result in the cache for next time.
	mangaJSON, err := json.Marshal(manga)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manga for caching: %w", err)
	}

	if err := s.redis.Set(ctx, key, mangaJSON, cacheDuration).Err(); err != nil {
		// If caching fails, we still return the data but log the error.
		log.Printf("Failed to cache manga %s: %v\n", id, err)
	}

	return manga, nil
}

func (s *MangaService) List(ctx context.Context, params repository.ListMangaParams) ([]*domain.Manga, error) {
	return s.mangaRepo.List(ctx, params)
}

// Update now includes cache invalidation.
func (s *MangaService) Update(ctx context.Context, manga *domain.Manga) error {
	if err := s.mangaRepo.Update(ctx, manga); err != nil {
		return err
	}

	// Invalidate the cache for this manga
	key := getMangaCacheKey(manga.ID)
	log.Println("CACHE INVALIDATED for key:", key)
	s.redis.Del(ctx, key) // We can ignore the error here for simplicity

	return nil
}

// Delete now includes cache invalidation.
func (s *MangaService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.mangaRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate the cache for this manga
	key := getMangaCacheKey(id)
	log.Println("CACHE INVALIDATED for key:", key)
	s.redis.Del(ctx, key) // We can ignore the error here for simplicity

	return nil
}
