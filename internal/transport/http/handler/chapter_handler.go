package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/0xpanadol/manga/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChapterHandler struct {
	chapterService *service.ChapterService
}

func NewChapterHandler(chapterService *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{chapterService: chapterService}
}

type createChapterRequest struct {
	ChapterNumber string   `json:"chapter_number" binding:"required,max=20"`
	Title         *string  `json:"title,omitempty" binding:"max=255"`
	Pages         []string `json:"pages,omitempty"` // Initially, pages might be empty before upload
}

// @Summary      Create a new chapter
// @Description  Adds a new chapter to a specific manga. Requires 'chapters:manage' permission.
// @Tags         Chapters
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        manga_id path      string  true  "Manga ID"
// @Param        request  body      handler.createChapterRequest true "Chapter Creation Info"
// @Success      201      {object}  domain.Chapter
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /manga/{manga_id}/chapters [post]
func (h *ChapterHandler) CreateChapter(c *gin.Context) {
	mangaIDStr := c.Param("manga_id")
	mangaID, err := uuid.Parse(mangaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	var req createChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	pages := req.Pages
	if pages == nil {
		pages = []string{} // Initialize as an empty slice instead of nil
	}

	chapter := &domain.Chapter{
		MangaID:       mangaID,
		ChapterNumber: req.ChapterNumber,
		Title:         req.Title,
		Pages:         pages, // Use the non-nil slice
	}

	if err := h.chapterService.Create(c.Request.Context(), chapter); err != nil {
		if errors.Is(err, repository.ErrChapterAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create chapter"})
		return
	}

	c.JSON(http.StatusCreated, chapter)
}

// @Summary      Get a single chapter by ID
// @Description  Retrieves details for a single chapter.
// @Tags         Chapters
// @Produce      json
// @Param        id   path      string  true  "Chapter ID"
// @Success      200  {object}  domain.Chapter
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /chapters/{id} [get]
func (h *ChapterHandler) GetChapter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter ID format"})
		return
	}

	chapter, err := h.chapterService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrChapterNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve chapter"})
		return
	}
	c.JSON(http.StatusOK, chapter)
}

// @Summary      List chapters for a manga
// @Description  Retrieves a paginated list of chapters for a specific manga.
// @Tags         Chapters
// @Produce      json
// @Param        manga_id  path      string  true  "Manga ID"
// @Param        page      query     int     false "Page number" default(1)
// @Param        per_page  query     int     false "Items per page" default(20)
// @Success      200       {array}   domain.Chapter
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /manga/{manga_id}/chapters [get]
func (h *ChapterHandler) ListChapters(c *gin.Context) {
	mangaIDStr := c.Param("id")
	mangaID, err := uuid.Parse(mangaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	var req listMangaRequest // Re-using the struct from manga handler
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "details": err.Error()})
		return
	}

	params := repository.ListChaptersParams{
		MangaID: mangaID,
		Limit:   req.PerPage,
		Offset:  (req.Page - 1) * req.PerPage,
	}

	chapters, err := h.chapterService.ListByMangaID(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list chapters"})
		return
	}
	c.JSON(http.StatusOK, chapters)
}

// @Summary      Update a chapter
// @Description  Updates the details of a specific chapter. Requires 'chapters:manage' permission.
// @Tags         Chapters
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Chapter ID"
// @Param        request body      handler.createChapterRequest true "Chapter Update Info"
// @Success      200     {object}  domain.Chapter
// @Failure      400     {object}  map[string]string
// @Failure      401     {object}  map[string]string
// @Failure      403     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /chapters/{id} [put]
func (h *ChapterHandler) UpdateChapter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter ID format"})
		return
	}

	var req createChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	pages := req.Pages
	if pages == nil {
		pages = []string{}
	}

	// Fetch existing chapter to get manga_id (or pass it in the request)
	// For simplicity, we just update the fields provided
	chapter := &domain.Chapter{
		ID:            id,
		ChapterNumber: req.ChapterNumber,
		Title:         req.Title,
		Pages:         pages, // Use the non-nil slice
	}

	if err := h.chapterService.Update(c.Request.Context(), chapter); err != nil {
		if errors.Is(err, repository.ErrChapterNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
			return
		}
		if errors.Is(err, repository.ErrChapterAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update chapter"})
		return
	}

	c.JSON(http.StatusOK, chapter)
}

// @Summary      Delete a chapter
// @Description  Deletes a specific chapter. Requires 'chapters:manage' permission.
// @Tags         Chapters
// @Security     BearerAuth
// @Param        id   path      string  true  "Chapter ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /chapters/{id} [delete]
func (h *ChapterHandler) DeleteChapter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter ID format"})
		return
	}

	if err := h.chapterService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrChapterNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete chapter"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      Upload chapter pages
// @Description  Uploads one or more image files for a chapter. Requires 'chapters:manage' permission.
// @Tags         Chapters
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string  true  "Chapter ID"
// @Param        pages  formData  file    true  "Image files for the chapter pages. Can be sent multiple times."
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /chapters/{id}/pages [post]
func (h *ChapterHandler) UploadPages(c *gin.Context) {
	chapterIDStr := c.Param("id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter ID format"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form", "details": err.Error()})
		return
	}

	// "pages" is the field name in the multipart form
	files := form.File["pages"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded"})
		return
	}

	err = h.chapterService.UploadPages(c.Request.Context(), chapterID, files)
	if err != nil {
		if errors.Is(err, repository.ErrChapterNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload pages", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d pages uploaded successfully", len(files))})
}
