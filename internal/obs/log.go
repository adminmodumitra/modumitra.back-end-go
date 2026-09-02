package obs

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Log is the process-wide zerolog logger. Prefer Logger() / FromCtx / FromGin.
var Log zerolog.Logger

// InitLogger configures JSON zerolog on stderr (tests).
func InitLogger(level string, serviceName ...string) {
	svc := ""
	if len(serviceName) > 0 {
		svc = serviceName[0]
	}
	InitLoggerWithWriter(os.Stderr, level, svc)
}

// InitLoggerWithWriter configures JSON zerolog on w.
func InitLoggerWithWriter(w io.Writer, level, serviceName string) {
	if w == nil {
		w = os.Stderr
	}
	initLogger(w, level, serviceName)
}

// OpenSinks returns the combined log writer for named sinks.
// Empty defaults to stderr. loki and cloudwatch are reserved until writers ship.
func OpenSinks(sinks []string) (io.Writer, error) {
	if len(sinks) == 0 {
		return os.Stderr, nil
	}
	seen := make(map[string]struct{}, len(sinks))
	var writers []io.Writer
	for _, raw := range sinks {
		s := strings.ToLower(strings.TrimSpace(raw))
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		switch s {
		case utils.SinkStderr:
			writers = append(writers, os.Stderr)
		case utils.SinkStdout:
			writers = append(writers, os.Stdout)
		case utils.SinkLoki:
			return nil, fmt.Errorf("log sink %q is reserved; scrape stderr with Promtail/Alloy (LOKI_URL is accepted in config but unused until the writer ships)", utils.SinkLoki)
		case utils.SinkCloudWatch:
			return nil, fmt.Errorf("log sink %q is reserved; use ECS awslogs / CloudWatch agent on stderr", utils.SinkCloudWatch)
		default:
			return nil, fmt.Errorf("unknown log sink %q (want stderr, stdout, loki, cloudwatch)", s)
		}
	}
	if len(writers) == 0 {
		return os.Stderr, nil
	}
	if len(writers) == 1 {
		return writers[0], nil
	}
	return io.MultiWriter(writers...), nil
}

func initLogger(w io.Writer, level, serviceName string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	s := strings.ToLower(strings.TrimSpace(level))
	if s == "warning" {
		s = "warn"
	}
	lvl, err := zerolog.ParseLevel(s)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	ctx := zerolog.New(w).Level(lvl).With().Timestamp()
	if serviceName != "" {
		ctx = ctx.Str("service", serviceName)
	}
	SetLogger(ctx.Logger())
}

// SetLogger replaces the process logger (tests).
func SetLogger(l zerolog.Logger) {
	Log = l
	log.Logger = l
	zerolog.DefaultContextLogger = &Log
}

// Logger returns the process-wide zerolog logger.
func Logger() zerolog.Logger {
	return Log
}

// FromCtx returns a logger annotated with request_id / trace_id when present.
func FromCtx(ctx context.Context) zerolog.Logger {
	l := Logger()
	if ctx == nil {
		return l
	}
	ev := l.With()
	if rid := utils.RequestIDFromContext(ctx); rid != "" {
		ev = ev.Str("request_id", rid)
	}
	if tid := TraceIDFromContext(ctx); tid != "" {
		ev = ev.Str("trace_id", tid)
	}
	return ev.Logger()
}

// FromGin returns FromCtx of the gin request context.
func FromGin(c *gin.Context) zerolog.Logger {
	if c == nil || c.Request == nil {
		return Logger()
	}
	return FromCtx(c.Request.Context())
}
