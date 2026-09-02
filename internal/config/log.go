package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Log is process logging: JSON on configured sinks (default stderr).
type Log struct {
	Level      string        `json:"level"`
	Format     string        `json:"format"`
	Sinks      []string      `json:"sinks"`
	LokiURL    string        `json:"loki_url"`
	CloudWatch CloudWatchLog `json:"cloudwatch"`
}

// CloudWatchLog is reserved until a CloudWatch writer ships.
type CloudWatchLog struct {
	Region    string `json:"region"`
	LogGroup  string `json:"log_group"`
	LogStream string `json:"log_stream"`
}

func applyFileLog(cfg *Config, l Log) {
	if l.Level != "" {
		cfg.Log.Level = l.Level
	}
	if l.Format != "" {
		cfg.Log.Format = l.Format
	}
	if len(l.Sinks) > 0 {
		cfg.Log.Sinks = normalizeLogSinks(l.Sinks)
	}
	if l.LokiURL != "" {
		cfg.Log.LokiURL = l.LokiURL
	}
	if l.CloudWatch.Region != "" {
		cfg.Log.CloudWatch.Region = l.CloudWatch.Region
	}
	if l.CloudWatch.LogGroup != "" {
		cfg.Log.CloudWatch.LogGroup = l.CloudWatch.LogGroup
	}
	if l.CloudWatch.LogStream != "" {
		cfg.Log.CloudWatch.LogStream = l.CloudWatch.LogStream
	}
}

func applyLogEnv(cfg *Config) {
	cfg.Log.Level = getenv("LOG_LEVEL", cfg.Log.Level)
	cfg.Log.Format = getenv("LOG_FORMAT", cfg.Log.Format)
	if v := strings.TrimSpace(os.Getenv("LOG_SINK")); v != "" {
		cfg.Log.Sinks = normalizeLogSinks(splitCSV(v))
	}
	cfg.Log.LokiURL = getenv("LOKI_URL", cfg.Log.LokiURL)
	region := getenv("AWS_REGION", "")
	if region == "" {
		region = getenv("CLOUDWATCH_REGION", cfg.Log.CloudWatch.Region)
	}
	if region != "" {
		cfg.Log.CloudWatch.Region = region
	}
	cfg.Log.CloudWatch.LogGroup = getenv("CLOUDWATCH_LOG_GROUP", cfg.Log.CloudWatch.LogGroup)
	cfg.Log.CloudWatch.LogStream = getenv("CLOUDWATCH_LOG_STREAM", cfg.Log.CloudWatch.LogStream)
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
