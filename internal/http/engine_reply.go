package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/engine"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// maxEngineMessage trims engine validation messages relayed to the client.
const maxEngineMessage = 300

type engineErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// relayEngine copies a 2xx engine answer through unchanged (streamed, capped).
// 400/422 mean the design itself is invalid and are relayed in our error
// shape. Anything else (401 wrong key, 5xx, ...) is the engine's fault: 502,
// or 504 when the engine itself timed out. Engine internals never reach the
// client.
func relayEngine(c *gin.Context, resp *http.Response) {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		middleware.InjectHeaders(c)
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Status(utils.StatusOK)
		_, _ = io.Copy(c.Writer, io.LimitReader(resp.Body, engine.MaxResponseBytes))
		middleware.Success(c, middleware.OpFrom(c), utils.StatusOK)
	case resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity:
		relayInvalid(c, resp)
	case resp.StatusCode == http.StatusRequestEntityTooLarge:
		middleware.Abort(c, utils.E(utils.CodePayloadTooLarge))
	case resp.StatusCode == http.StatusGatewayTimeout:
		middleware.Abort(c, utils.E(utils.CodeEngineTimeout))
	default:
		middleware.Abort(c, utils.E(utils.CodeEngineError))
	}
}

func relayInvalid(c *gin.Context, resp *http.Response) {
	var eb engineErrorBody
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&eb)
	e := utils.E(utils.CodeValidation)
	code, msg := eb.Error.Code, eb.Error.Message
	if code == "" {
		code = e.Code.String()
	}
	if msg == "" {
		msg = e.Message
	}
	if len(msg) > maxEngineMessage {
		msg = msg[:maxEngineMessage]
	}
	middleware.WriteError(c, utils.StatusUnprocessableEntity, code, msg, map[string]any{"source": "engine"})
	middleware.Exit(c, middleware.OpFrom(c), utils.StatusUnprocessableEntity, code)
}
