package middleware

import "github.com/gin-gonic/gin"

// RBAC is a no-op in slice 0.
func RBAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
