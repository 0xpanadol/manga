package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/0xpanadol/manga/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MangaHandler struct {
	mangaService *service.MangaService
}

func NewMangaHandler(mangaService *service.MangaService) *MangaHandler {
	return &MangaHandler{mangaService: mangaService}
}

type createMangaRequest struct {
	Title       string   `json:"title" binding:"required,min=2,max=255"`
	Description string   `json:"description" binding:"required"`
	Author      string   `json:"author" binding:"required,min=2,max=100"`
	Status      string   `json:"status" binding:"required,oneof=ongoing completed hiatus cancelled"`
	Genres      []string `json:"genres" binding:"required,min=1"`
}

// @Summary      Create a new manga
// @Description  Adds a new manga to the catalog. Requires 'manga:manage' permission.
// @Tags         Manga
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body handler.createMangaRequest true "Manga Creation Info"
// @Success      201  {object}  domain.Manga
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manga [post]
func (h *MangaHandler) CreateManga(c *gin.Context) {
	var req createMangaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	manga := &domain.Manga{
		Title:       req.Title,
		Description: req.Description,
		Author:      req.Author,
		Status:      domain.MangaStatus(req.Status),
		Genres:      req.Genres,
	}

	if err := h.mangaService.Create(c.Request.Context(), manga); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create manga"})
		return
	}

	c.JSON(http.StatusCreated, manga)
}

// @Summary      Get a single manga by ID
// @Description  Retrieves details for a single manga, including its genres.
// @Tags         Manga
// @Produce      json
// @Param        id   path      string  true  "Manga ID"
// @Success      200  {object}  domain.Manga
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manga/{id} [get]
func (h *MangaHandler) GetManga(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	manga, err := h.mangaService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve manga"})
		return
	}

	c.JSON(http.StatusOK, manga)
}

// listMangaRequest defines the query parameters for listing manga.
type listMangaRequest struct {
	Page    int    `form:"page,default=1"`
	PerPage int    `form:"per_page,default=20"`
	Query   string `form:"q"`
	Genres  string `form:"genres"` // Comma-separated
	Status  string `form:"status"`
	Sort    string `form:"sort"` // e.g., "title", "-created_at"
}

// @Summary      List manga
// @Description  Retrieves a paginated and filtered list of manga.
// @Tags         Manga
// @Produce      json
// @Param        page      query     int     false  "Page number" default(1)
// @Param        per_page  query     int     false  "Items per page" default(20)
// @Param        q         query     string  false  "Full-text search query for title and description"
// @Param        genres    query     string  false  "Filter by comma-separated genre names (e.g., Action,Fantasy)"
// @Param        status    query     string  false  "Filter by status" Enums(ongoing, completed, hiatus, cancelled)
// @Param        sort      query     string  false  "Sort order (e.g., title, -created_at)"
// @Success      200       {array}   domain.Manga
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /manga [get]
func (h *MangaHandler) ListManga(c *gin.Context) {
	var req listMangaRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "details": err.Error()})
		return
	}

	params := repository.ListMangaParams{
		Limit:       req.PerPage,
		Offset:      (req.Page - 1) * req.PerPage,
		SearchQuery: req.Query,
		Status:      req.Status,
	}

	if req.Genres != "" {
		params.Genres = strings.Split(req.Genres, ",")
	}

	if req.Sort != "" {
		if strings.HasPrefix(req.Sort, "-") {
			params.SortOrder = "desc"
			params.SortBy = strings.TrimPrefix(req.Sort, "-")
		} else {
			params.SortOrder = "asc"
			params.SortBy = req.Sort
		}
	}

	mangas, err := h.mangaService.List(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list manga"})
		return
	}

	c.JSON(http.StatusOK, mangas)
}

// updateMangaRequest uses the same fields as createMangaRequest.
type updateMangaRequest createMangaRequest

func (h *MangaHandler) UpdateManga(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	var req updateMangaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	manga := &domain.Manga{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Author:      req.Author,
		Status:      domain.MangaStatus(req.Status),
		Genres:      req.Genres,
	}

	err = h.mangaService.Update(c.Request.Context(), manga)
	if err != nil {
		if errors.Is(err, repository.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update manga"})
		return
	}

	c.JSON(http.StatusOK, manga)
}

func (h *MangaHandler) DeleteManga(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	err = h.mangaService.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete manga"})
		return
	}

	c.Status(http.StatusNoContent)
}
