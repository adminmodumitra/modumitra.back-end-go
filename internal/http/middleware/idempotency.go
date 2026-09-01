package middleware

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type captured struct {
	status int
	body   []byte
	ct     string
}

type pending struct {
	done chan struct{}
	cap  captured
}

// Idempotency remembers Idempotency-Key responses in memory for process lifetime.
// Missing key → pass through.
func Idempotency() gin.HandlerFunc {
	var mu sync.Mutex
	keys := map[string]*pending{}

	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}
		mapKey := c.Request.Method + " " + c.Request.URL.Path + " " + key

		mu.Lock()
		if p, ok := keys[mapKey]; ok {
			mu.Unlock()
			<-p.done
			InjectHeaders(c)
			if p.cap.ct != "" {
				c.Header("Content-Type", p.cap.ct)
			}
			c.Abort()
			c.Data(p.cap.status, p.cap.ct, p.cap.body)
			return
		}
		p := &pending{done: make(chan struct{})}
		keys[mapKey] = p
		mu.Unlock()

		buf := &bodyWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
		c.Writer = buf
		c.Next()

		p.cap = captured{
			status: buf.Status(),
			body:   buf.buf.Bytes(),
			ct:     buf.Header().Get("Content-Type"),
		}
		close(p.done)
	}
}

type bodyWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyWriter) WriteString(s string) (int, error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
}

var _ http.ResponseWriter = (*bodyWriter)(nil)
