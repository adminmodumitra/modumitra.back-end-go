package middleware

import "github.com/gin-gonic/gin"

// Tenant is a no-op in slice 0 (no organization on health/ready/public config).
func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
