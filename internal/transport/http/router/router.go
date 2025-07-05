package router

import (
	"github.com/0xpanadol/manga/internal/transport/http/handler"
	"github.com/0xpanadol/manga/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func Setup(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	mangaHandler *handler.MangaHandler,
	chapterHandler *handler.ChapterHandler,
	socialHandler *handler.SocialHandler,
	jwtSecret string,
) {
	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/request-password-reset", authHandler.RequestPasswordReset)

		}

		// Protected routes
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtSecret))
		{
			users.GET("/me", userHandler.GetMe)
		}

		// Manga ROUTES
		manga := api.Group("/manga")
		{
			// Public routes
			manga.GET("/", mangaHandler.ListManga)
			manga.GET("/:id", mangaHandler.GetManga)
			// Chapter routes nested under manga
			manga.GET("/:id/chapters", chapterHandler.ListChapters)

			// Admin-only routes
			adminManga := manga.Group("/")
			adminManga.Use(
				middleware.AuthMiddleware(jwtSecret),
				middleware.PermissionRequired("manga:manage"),
			)
			{
				adminManga.POST("/", mangaHandler.CreateManga)
				adminManga.PUT("/:id", mangaHandler.UpdateManga)
				adminManga.DELETE("/:id", mangaHandler.DeleteManga)
			}
		}

		// Chapters ROUTES
		chapters := api.Group("/chapters")
		{
			chapters.GET("/:id", chapterHandler.GetChapter)
		}
		// Admin-only routes
		adminPermission := middleware.PermissionRequired("chapters:manage")
		authMiddleware := middleware.AuthMiddleware(jwtSecret)

		// Create chapter is nested under manga for context
		api.POST("/manga/:manga_id/chapters", authMiddleware, adminPermission, chapterHandler.CreateChapter)

		// Update/Delete chapter can be at the top level
		api.PUT("/chapters/:id", authMiddleware, adminPermission, chapterHandler.UpdateChapter)
		api.DELETE("/chapters/:id", authMiddleware, adminPermission, chapterHandler.DeleteChapter)
		api.POST("/chapters/:id/pages", authMiddleware, adminPermission, chapterHandler.UploadPages) // New

		// Public Comment Routes
		api.GET("/manga/:id/comments", socialHandler.ListMangaComments)
		api.GET("/chapters/:id/comments", socialHandler.ListChapterComments)

		// Authenticated Routes
		authenticated := api.Group("/")
		authenticated.Use(middleware.AuthMiddleware(jwtSecret))
		{
			// Favorites & Progress
			authenticated.POST("/manga/:manga_id/favorite", socialHandler.ToggleFavorite)
			authenticated.GET("/users/me/favorites", socialHandler.ListFavorites)
			authenticated.POST("/chapters/:id/progress", socialHandler.MarkChapterAsRead)
			authenticated.GET("/users/me/progress", socialHandler.ListReadChapters)

			// Comment Creation
			authenticated.POST("/manga/:manga_id/comments", socialHandler.CreateMangaComment)
			authenticated.POST("/chapters/:id/comments", socialHandler.CreateChapterComment)
		}
	}
}
