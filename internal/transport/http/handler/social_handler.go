package handler

import (
	"net/http"

	"github.com/0xpanadol/manga/internal/domain"
	"github.com/0xpanadol/manga/internal/repository"
	"github.com/0xpanadol/manga/internal/service"
	"github.com/0xpanadol/manga/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SocialHandler struct {
	socialService *service.SocialService
}

func NewSocialHandler(socialService *service.SocialService) *SocialHandler {
	return &SocialHandler{socialService: socialService}
}

// @Summary      Toggle manga favorite status
// @Description  Adds or removes a manga from the current user's favorites list.
// @Tags         Social
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Manga ID"
// @Success      200  {object}  repository.ToggleFavoriteResult
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manga/{id}/favorite [post]
func (h *SocialHandler) ToggleFavorite(c *gin.Context) {
	mangaIDStr := c.Param("manga_id")
	mangaID, err := uuid.Parse(mangaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	result, err := h.socialService.ToggleFavorite(c.Request.Context(), userID, mangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update favorite status"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary      List user's favorite manga
// @Description  Retrieves a paginated list of the current user's favorite manga.
// @Tags         Social
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false "Page number" default(1)
// @Param        per_page  query     int     false "Items per page" default(20)
// @Success      200       {array}   domain.Manga
// @Failure      401       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /users/me/favorites [get]
func (h *SocialHandler) ListFavorites(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	var req listMangaRequest // Re-using from manga_handler
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	params := repository.ListMangaParams{
		Limit:  req.PerPage,
		Offset: (req.Page - 1) * req.PerPage,
	}

	mangas, err := h.socialService.ListFavorites(c.Request.Context(), userID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list favorites"})
		return
	}

	c.JSON(http.StatusOK, mangas)
}

// @Summary      Mark chapter as read
// @Description  Marks a chapter as read for the current user.
// @Tags         Social
// @Security     BearerAuth
// @Param        id   path      string  true  "Chapter ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /chapters/{id}/progress [post]
func (h *SocialHandler) MarkChapterAsRead(c *gin.Context) {
	chapterIDStr := c.Param("id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter ID format"})
		return
	}

	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	if err := h.socialService.MarkChapterAsRead(c.Request.Context(), userID, chapterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark chapter as read"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      List user's read chapters
// @Description  Retrieves a list of all chapters marked as read by the current user.
// @Tags         Social
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.Chapter
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/me/progress [get]
func (h *SocialHandler) ListReadChapters(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	chapters, err := h.socialService.ListReadChapters(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list read chapters"})
		return
	}

	c.JSON(http.StatusOK, chapters)
}

type createCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// @Summary      Post a comment on a manga
// @Description  Adds a new comment to a specific manga.
// @Tags         Social
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Manga ID"
// @Param        request body      handler.createCommentRequest true "Comment Content"
// @Success      201     {object}  domain.Comment
// @Failure      400     {object}  map[string]string
// @Failure      401     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /manga/{id}/comments [post]
func (h *SocialHandler) CreateMangaComment(c *gin.Context) {
	mangaIDStr := c.Param("manga_id")
	mangaID, err := uuid.Parse(mangaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	comment := &domain.Comment{
		UserID:  userID,
		MangaID: &mangaID, // Set the manga ID
		Content: req.Content,
	}

	if err := h.socialService.CreateComment(c.Request.Context(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// @Summary      Post a comment on a chapter
// @Description  Adds a new comment to a specific chapter.
// @Tags         Social
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Chapter ID"
// @Param        request body      handler.createCommentRequest true "Comment Content"
// @Success      201     {object}  domain.Comment
// @Failure      400     {object}  map[string]string
// @Failure      401     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /chapters/{id}/comments [post]
func (h *SocialHandler) CreateChapterComment(c *gin.Context) {
	chapterIDStr := c.Param("id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter ID format"})
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	comment := &domain.Comment{
		UserID:    userID,
		ChapterID: &chapterID, // Set the chapter ID
		Content:   req.Content,
	}

	if err := h.socialService.CreateComment(c.Request.Context(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// @Summary      List manga comments
// @Description  Retrieves a paginated list of comments for a specific manga.
// @Tags         Social
// @Produce      json
// @Param        id        path      string  true  "Manga ID"
// @Param        page      query     int     false "Page number" default(1)
// @Param        per_page  query     int     false "Items per page" default(20)
// @Success      200       {array}   domain.CommentWithUser
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /manga/{id}/comments [get]
func (h *SocialHandler) ListMangaComments(c *gin.Context) {
	mangaIDStr := c.Param("id")
	mangaID, err := uuid.Parse(mangaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	var req listMangaRequest // Re-using for pagination
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	params := repository.ListCommentsParams{
		ParentID:   mangaID,
		ParentType: domain.ParentTypeManga,
		Limit:      req.PerPage,
		Offset:     (req.Page - 1) * req.PerPage,
	}

	comments, err := h.socialService.ListComments(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// @Summary      List chapter comments
// @Description  Retrieves a paginated list of comments for a specific chapter.
// @Tags         Social
// @Produce      json
// @Param        id        path      string  true  "Chapter ID"
// @Param        page      query     int     false "Page number" default(1)
// @Param        per_page  query     int     false "Items per page" default(20)
// @Success      200       {array}   domain.CommentWithUser
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /chapters/{id}/comments [get]
func (h *SocialHandler) ListChapterComments(c *gin.Context) {
	chapterIDStr := c.Param("id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter ID format"})
		return
	}

	var req listMangaRequest // Re-using for pagination
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	params := repository.ListCommentsParams{
		ParentID:   chapterID,
		ParentType: domain.ParentTypeChapter,
		Limit:      req.PerPage,
		Offset:     (req.Page - 1) * req.PerPage,
	}

	comments, err := h.socialService.ListComments(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}
