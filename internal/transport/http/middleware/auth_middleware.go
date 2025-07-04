package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/0xpanadol/manga/pkg/jwtauth"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "Authorization"
	AuthorizationTypeBearer = "Bearer"
	UserIDKey               = "UserID"
	UserRoleKey             = "UserRole"
	UserPermissionsKey      = "UserPermissions"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 || !strings.EqualFold(fields[0], AuthorizationTypeBearer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		accessToken := fields[1]
		claims, err := jwtauth.ValidateToken(accessToken, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Set user info in context for downstream handlers
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(UserPermissionsKey, claims.Permissions)

		c.Next()
	}
}

func PermissionRequired(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get(UserPermissionsKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permissions not found in token"})
			return
		}

		permissionSlice, ok := permissions.([]string)
		if !ok {
			// This might happen if the token was malformed.
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid permissions format in token"})
			return
		}

		if !slices.Contains(permissionSlice, requiredPermission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}
