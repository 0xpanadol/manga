package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CorsMiddleware configures the Cross-Origin Resource Sharing settings.
func CorsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// SecurityHeadersMiddleware adds security-related headers to every response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protects against MIME-type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		// Protects against clickjacking
		c.Header("X-Frame-Options", "DENY")
		// Enables the XSS filter in browsers
		c.Header("X-XSS-Protection", "1; mode=block")
		// If you are serving over HTTPS, uncomment the line below
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}
