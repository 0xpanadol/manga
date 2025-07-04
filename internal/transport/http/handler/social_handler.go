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
