package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestSizeLimit limits the size of incoming request bodies
func RequestSizeLimit(maxSizeMB int) gin.HandlerFunc {
	maxBytes := int64(maxSizeMB) << 20 // Convert MB to bytes

	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

		c.Next()

		// Check if request was too large
		if c.Writer.Status() == http.StatusRequestEntityTooLarge {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request body too large",
				"max_size_mb": maxSizeMB,
			})
			c.Abort()
		}
	}
}
