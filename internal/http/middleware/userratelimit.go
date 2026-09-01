package middleware

import "github.com/gin-gonic/gin"

// UserRateLimit is a no-op when there is no authenticated user (slice 0).
func UserRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
