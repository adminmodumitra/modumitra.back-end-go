package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Log is process logging: JSON on configured sinks (default stderr).
type Log struct {
	Level      string
	Format     string
	Sinks      []string
	LokiURL    string
	CloudWatch CloudWatchLog
}

// CloudWatchLog is reserved until a CloudWatch writer ships.
type CloudWatchLog struct {
	Region    string
	LogGroup  string
	LogStream string
}

func loadLog() Log {
	sinks := normalizeLogSinks(splitCSV(os.Getenv("LOG_SINK")))
	if len(sinks) == 0 {
		sinks = []string{utils.SinkStderr}
	}
	region := getenv("AWS_REGION", "")
	if region == "" {
		region = getenv("CLOUDWATCH_REGION", "")
	}
	return Log{
		Level:   getenv("LOG_LEVEL", "info"),
		Format:  getenv("LOG_FORMAT", "json"),
		Sinks:   sinks,
		LokiURL: getenv("LOKI_URL", ""),
		CloudWatch: CloudWatchLog{
			Region:    region,
			LogGroup:  getenv("CLOUDWATCH_LOG_GROUP", ""),
			LogStream: getenv("CLOUDWATCH_LOG_STREAM", ""),
		},
	}
}

func normalizeLogSinks(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// SinksOrDefault is stderr when no sinks are configured.
func (l Log) SinksOrDefault() []string {
	if len(l.Sinks) == 0 {
		return []string{utils.SinkStderr}
	}
	return l.Sinks
}

// Validate rejects unsupported log formats. Sinks are checked by obs.OpenSinks.
func (l Log) Validate() error {
	format := strings.ToLower(strings.TrimSpace(l.Format))
	if format == "" {
		format = "json"
	}
	if format != "json" {
		return fmt.Errorf("log format %q not supported (want json)", l.Format)
	}
	return nil
}
