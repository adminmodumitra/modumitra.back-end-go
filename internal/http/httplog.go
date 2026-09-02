package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
)

// logEnter starts handler access logging. WriteJSON logs success; Abort logs exit.
func logEnter(c *gin.Context, op string) {
	middleware.Enter(c, op)
}
