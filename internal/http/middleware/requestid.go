package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			id = newRequestID()
		}
		c.Set(ctxRequestID, id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}

func newRequestID() string {
	var b [13]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req_00000000000000000000000000"
	}
	return "req_" + hex.EncodeToString(b[:])
}
