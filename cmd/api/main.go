package main

import (
	"context"
	"log"
	"net/http"

	"github.com/0xpanadol/manga/internal/config"
	"github.com/0xpanadol/manga/pkg/broker"
	"github.com/0xpanadol/manga/pkg/logger"
	"github.com/0xpanadol/manga/pkg/uploader"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	// Import the new postgres repository package with an alias
	postgresrepo "github.com/0xpanadol/manga/internal/repository/postgres"
	"github.com/0xpanadol/manga/internal/service"
	"github.com/0xpanadol/manga/internal/transport/http/handler"
	"github.com/0xpanadol/manga/internal/transport/http/middleware"
	"github.com/0xpanadol/manga/internal/transport/http/router"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/0xpanadol/manga/docs"
)

// @title           Manga-Dex-Style API
// @version         1.0
// @description     This is a production-ready HTTP API in Go for a manga reading platform.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a JWT token.
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// === INITIALIZE LOGGER ===
	appLogger := logger.New(cfg.AppEnv)
	defer appLogger.Sync() // Flushes buffer, if any
	appLogger.Info("Application configuration loaded")

	dbpool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Database connection established")

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	// Ping the redis server to check the connection
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("could not connect to redis: %v", err)
	}
	log.Println("Redis connection established")

	// === INITIALIZE BROKER ===
	messageBroker, err := broker.NewRabbitMQBroker(cfg.RabbitMQUrl)
	if err != nil {
		appLogger.Fatal("Could not initialize message broker", zap.Error(err))
	}
	defer messageBroker.Close()
	appLogger.Info("Message broker connected")

	// WIRING
	userRepo := postgresrepo.NewPostgresUserRepository(dbpool)
	mangaRepo := postgresrepo.NewPostgresMangaRepository(dbpool)
	chapterRepo := postgresrepo.NewPostgresChapterRepository(dbpool)
	socialRepo := postgresrepo.NewPostgresSocialRepository(dbpool)

	// New: Initialize MinIO Uploader
	minioUploader, err := uploader.NewMinioUploader(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucketName,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("could not initialize minio uploader: %v", err)
	}
	log.Println("MinIO uploader initialized")

	authService := service.NewAuthService(
		userRepo,
		messageBroker,
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessExpiresIn,
		cfg.JWTRefreshExpiresIn,
	)
	userService := service.NewUserService(userRepo)
	mangaService := service.NewMangaService(mangaRepo, redisClient)
	chapterService := service.NewChapterService(chapterRepo, minioUploader)
	socialService := service.NewSocialService(socialRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	mangaHandler := handler.NewMangaHandler(mangaService)
	chapterHandler := handler.NewChapterHandler(chapterService)
	socialHandler := handler.NewSocialHandler(socialService)

	// ROUTER
	ginRouter := gin.Default()
	ginRouter.Use(middleware.MetricsMiddleware()) // Metrics should be early
	ginRouter.Use(middleware.LoggerMiddleware(appLogger))
	ginRouter.Use(middleware.CorsMiddleware(cfg.CorsAllowedOrigins))
	ginRouter.Use(middleware.SecurityHeadersMiddleware())
	ginRouter.Use(middleware.ErrorHandler())

	// === ADD NEW SYSTEM ENDPOINTS ===
	ginRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Add Swagger route
	ginRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	ginRouter.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.Setup(
		ginRouter,
		authHandler,
		userHandler,
		mangaHandler,
		chapterHandler,
		socialHandler,
		cfg.JWTAccessSecret,
	)

	// SERVER
	appLogger.Info("Server starting", zap.String("port", cfg.AppPort))
	if err := ginRouter.Run(":" + cfg.AppPort); err != nil {
		appLogger.Fatal("failed to run server", zap.Error(err))
	}
}
