package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ExtendDeadlines lifts the server-wide ReadTimeout/WriteTimeout (10s/15s) for
// one slow route group: Phase 5a document calls upload view stills (up to
// 32 MB) and wait for a PDF. Uses http.ResponseController, so it only moves
// this connection's deadlines. Writers that cannot (test recorders) are left as
// they are.
func ExtendDeadlines(read, write time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := http.NewResponseController(c.Writer)
		now := time.Now()
		_ = rc.SetReadDeadline(now.Add(read))
		_ = rc.SetWriteDeadline(now.Add(write))
		c.Next()
	}
}
