package httpapi

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"

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
		c.Header("Vary", "Accept-Encoding")
		body := io.Reader(resp.Body)
		if engineGzipped(resp) {
			if acceptsGzip(c) {
				// Phase 4: CNC answers are MBs of JSON; pass the engine's gzip through.
				c.Header("Content-Encoding", "gzip")
			} else if body = gunzipped(resp.Body); body == nil {
				middleware.Abort(c, utils.E(utils.CodeEngineError))
				return
			}
		}
		c.Status(utils.StatusOK)
		_, _ = io.Copy(c.Writer, io.LimitReader(body, engine.MaxResponseBytes))
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
	body := io.Reader(resp.Body)
	if engineGzipped(resp) {
		body = gunzipped(resp.Body)
	}
	if body != nil {
		_ = json.NewDecoder(io.LimitReader(body, 64<<10)).Decode(&eb)
	}
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

// engineGzipped: the engine compressed its answer (the client asks for gzip
// explicitly, so Go does not unpack it).
func engineGzipped(resp *http.Response) bool {
	return strings.EqualFold(strings.TrimSpace(resp.Header.Get("Content-Encoding")), "gzip")
}

func acceptsGzip(c *gin.Context) bool {
	return strings.Contains(strings.ToLower(c.GetHeader("Accept-Encoding")), "gzip")
}

// gunzipped unpacks a gzip body, capped like an uncompressed one; nil when the
// stream is not gzip.
func gunzipped(r io.Reader) io.Reader {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil
	}
	return io.LimitReader(zr, engine.MaxResponseBytes)
}
