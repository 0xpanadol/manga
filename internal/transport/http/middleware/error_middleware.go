package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/0xpanadol/manga/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse defines the standard JSON error response structure.
type ErrorResponse struct {
	Error struct {
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Defer the error handling to run after the handler
		c.Next()

		// Check for any errors that occurred in the handlers
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Log the raw error for debugging
			log.Printf("An error occurred: %v", err)

			// Handle validation errors specifically
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				out := make([]map[string]string, len(ve))
				for i, fe := range ve {
					out[i] = map[string]string{"field": fe.Field(), "reason": fe.Tag()}
				}
				resp := ErrorResponse{}
				resp.Error.Message = "Validation failed"
				resp.Error.Details = out
				c.JSON(http.StatusBadRequest, resp)
				return
			}

			// Map our domain errors to HTTP responses
			appErr := apperrors.MapDomainErrors(err)
			resp := ErrorResponse{}
			resp.Error.Message = appErr.Error()
			c.JSON(appErr.Code, resp)
			return
		}
	}
}
