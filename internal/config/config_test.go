package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func clearDatastoreEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"CABINET_CONFIG",
		"ALLOWED_IPS",
		"APP_ENV",
		"CABINET_ENV",
		"JWT_SECRET",
		"ENTITLEMENT_SIGNING_KEY",
		"DATASTORE_CONFIG",
		"DATASTORE_DRIVER",
		"DATABASE_URL",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"REDIS_TLS",
		"MONGO_URI",
		"MONGO_DATABASE",
		"KAFKA_ENABLED",
		"KAFKA_BROKERS",
		"KAFKA_CLIENT_ID",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"LOG_SINK",
		"LOKI_URL",
		"AWS_REGION",
		"CLOUDWATCH_REGION",
		"CLOUDWATCH_LOG_GROUP",
		"CLOUDWATCH_LOG_STREAM",
		"BILLING_PROVIDER",
		"BILLING_WEBHOOK_SECRET",
		"BILLING_PRICE_STUDIO_MONTHLY",
		"BILLING_PRICE_TEAM_MONTHLY",
		"MAIL_DRIVER",
		"OAUTH_GOOGLE_CLIENT_ID",
		"OAUTH_GOOGLE_CLIENT_SECRET",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}

func dummyPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		cand := filepath.Join(dir, "configs", "cabinet.local.json")
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("configs/cabinet.local.json not found")
	return ""
}

func TestLoadDummyConfig(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))

	cfg := Load()
	if cfg.Datastore.Driver != "postgres" {
		t.Fatalf("Driver=%q", cfg.Datastore.Driver)
	}
	if cfg.Datastore.SQL.Driver != "postgres" {
		t.Fatalf("SQL.Driver=%q", cfg.Datastore.SQL.Driver)
	}
	if !strings.Contains(cfg.Datastore.SQL.DSN, "localhost:5432") {
		t.Fatalf("SQL.DSN=%q", cfg.Datastore.SQL.DSN)
	}
	if cfg.Datastore.SQL.MaxOpenConns != 25 || cfg.Datastore.SQL.MaxIdleConns != 10 {
		t.Fatalf("SQL pool open=%d idle=%d", cfg.Datastore.SQL.MaxOpenConns, cfg.Datastore.SQL.MaxIdleConns)
	}
	if cfg.Datastore.SQL.ConnMaxLifetime != "1h" {
		t.Fatalf("SQL.ConnMaxLifetime=%q", cfg.Datastore.SQL.ConnMaxLifetime)
	}
	if cfg.Datastore.Redis.Addr != "localhost:6379" {
		t.Fatalf("Redis.Addr=%q", cfg.Datastore.Redis.Addr)
	}
	if cfg.Datastore.NoSQL.URI != "mongodb://localhost:27017" {
		t.Fatalf("NoSQL.URI=%q", cfg.Datastore.NoSQL.URI)
	}
	if cfg.Datastore.NoSQL.Database != "cabinet_studio" {
		t.Fatalf("NoSQL.Database=%q", cfg.Datastore.NoSQL.Database)
	}
	if cfg.DatabaseURL != cfg.Datastore.SQL.DSN {
		t.Fatalf("DatabaseURL not mirrored")
	}
	if cfg.Datastore.NormalizedDriver() != "postgres" {
		t.Fatalf("NormalizedDriver=%q", cfg.Datastore.NormalizedDriver())
	}
}

func TestEnvOverrides(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))
	t.Setenv("DATASTORE_DRIVER", "redis")
	t.Setenv("DATABASE_URL", "postgres://cabinet:s3cret@db.example.com:5432/cabinet_studio?sslmode=require")
	t.Setenv("REDIS_ADDR", "redis.example.com:6380")
	t.Setenv("REDIS_PASSWORD", "p@ss")
	t.Setenv("MONGO_URI", "mongodb+srv://u:p@cluster.example.com")
	t.Setenv("MONGO_DATABASE", "prod")

	cfg := Load()
	if cfg.Datastore.Driver != "redis" {
		t.Fatalf("Driver=%q", cfg.Datastore.Driver)
	}
	if cfg.Datastore.SQL.DSN != "postgres://cabinet:s3cret@db.example.com:5432/cabinet_studio?sslmode=require" {
		t.Fatalf("SQL.DSN=%q", cfg.Datastore.SQL.DSN)
	}
	if cfg.Datastore.Redis.Addr != "redis.example.com:6380" {
		t.Fatalf("Redis.Addr=%q", cfg.Datastore.Redis.Addr)
	}
	if cfg.Datastore.Redis.Password != "p@ss" {
		t.Fatalf("Redis.Password not applied")
	}
	if cfg.Datastore.NoSQL.URI != "mongodb+srv://u:p@cluster.example.com" {
		t.Fatalf("NoSQL.URI=%q", cfg.Datastore.NoSQL.URI)
	}
	if cfg.Datastore.NoSQL.Database != "prod" {
		t.Fatalf("NoSQL.Database=%q", cfg.Datastore.NoSQL.Database)
	}
	if cfg.Datastore.NormalizedDriver() != "redis" {
		t.Fatalf("NormalizedDriver=%q", cfg.Datastore.NormalizedDriver())
	}
}

func TestAllowedIPs(t *testing.T) {
	clearDatastoreEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cabinet.json")
	body := `{"http":{"addr":":8080","allowed_ips":["10.0.0.8","192.168.0.0/16"]}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CABINET_CONFIG", path)
	cfg := Load()
	if len(cfg.AllowedIPs) != 2 || cfg.AllowedIPs[0] != "10.0.0.8" {
		t.Fatalf("AllowedIPs=%v", cfg.AllowedIPs)
	}
	t.Setenv("ALLOWED_IPS", `["203.0.113.4","198.51.100.0/24"]`)
	cfg = Load()
	if len(cfg.AllowedIPs) != 2 || cfg.AllowedIPs[0] != "203.0.113.4" || cfg.AllowedIPs[1] != "198.51.100.0/24" {
		t.Fatalf("env AllowedIPs=%v", cfg.AllowedIPs)
	}
}

func TestLoadJSONExactTags(t *testing.T) {
	clearDatastoreEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "ds.json")
	body := `{
		"Driver": "mongodb",
		"SQL": {"Driver": "postgres", "DSN": "postgres://n:p@127.0.0.1/db", "MaxOpenConns": 5, "MaxIdleConns": 2, "ConnMaxLifetime": "30m"},
		"Redis": {"Addr": "127.0.0.1:6379", "Password": "x", "DB": 2, "TLS": true, "MaxOpenConns": 7, "MaxIdleConns": 3, "ConnMaxLifetime": "15m"},
		"NoSQL": {"URI": "mongodb://127.0.0.1:27017", "Database": "app", "MaxOpenConns": 9, "MaxIdleConns": 4, "ConnMaxLifetime": "45m"}
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATASTORE_CONFIG", path)
	cfg := Load()
	if cfg.Datastore.Driver != "mongodb" {
		t.Fatalf("Driver=%q", cfg.Datastore.Driver)
	}
	if cfg.Datastore.Redis.DB != 2 || !cfg.Datastore.Redis.TLS {
		t.Fatalf("Redis=%+v", cfg.Datastore.Redis)
	}
	if cfg.Datastore.SQL.MaxOpenConns != 5 {
		t.Fatalf("SQL.MaxOpenConns=%d", cfg.Datastore.SQL.MaxOpenConns)
	}
	// Round-trip JSON tags.
	b, err := json.Marshal(cfg.Datastore)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{`"Driver"`, `"SQL"`, `"Redis"`, `"NoSQL"`, `"DSN"`, `"Addr"`, `"URI"`, `"MaxOpenConns"`, `"MaxIdleConns"`, `"ConnMaxLifetime"`} {
		if !strings.Contains(string(b), tag) {
			t.Fatalf("missing json tag %s in %s", tag, b)
		}
	}
}

func TestRedactDSN(t *testing.T) {
	in := "postgres://cabinet:super-secret@localhost:5432/cabinet_studio?sslmode=disable"
	out := RedactDSN(in)
	if strings.Contains(out, "super-secret") {
		t.Fatalf("password leaked: %q", out)
	}
	if strings.Contains(out, "cabinet:") && strings.Contains(out, "@") {
		// username may be replaced entirely
	}
	if !strings.Contains(out, "localhost:5432") {
		t.Fatalf("host stripped: %q", out)
	}
	if out == in {
		t.Fatalf("unchanged")
	}

	mongo := "mongodb+srv://alice:hunter2@cluster.example.com/db"
	mout := RedactDSN(mongo)
	if strings.Contains(mout, "hunter2") {
		t.Fatalf("mongo password leaked: %q", mout)
	}

	userinfo := RedactDSN("user:pw@localhost:6379")
	if strings.Contains(userinfo, "pw") {
		t.Fatalf("opaque userinfo leaked: %q", userinfo)
	}

	plain := RedactDSN("localhost:6379")
	if plain != "localhost:6379" {
		t.Fatalf("addr=%q", plain)
	}

	if RedactDSN("") != "" {
		t.Fatal("empty")
	}
}

func TestMaskSecrets(t *testing.T) {
	dsn := "postgres://cabinet:super-secret@localhost:5432/db"
	err := errors.New("cannot parse `" + dsn + "`: failed")
	got := MaskSecrets(err, dsn)
	if strings.Contains(got.Error(), "super-secret") {
		t.Fatalf("leaked: %q", got)
	}
}

func TestPoolDefaulting(t *testing.T) {
	n := (Pool{}).Normalized()
	if n.MaxOpenConns != 25 {
		t.Fatalf("zero MaxOpenConns defaulted to %d want 25", n.MaxOpenConns)
	}
	if n.MaxIdleConns != 0 {
		t.Fatalf("zero MaxIdleConns=%d", n.MaxIdleConns)
	}
	if n.ConnMaxLifetime != time.Hour {
		t.Fatalf("lifetime=%s", n.ConnMaxLifetime)
	}

	n = (Pool{MaxOpenConns: 0, MaxIdleConns: 10, ConnMaxLifetime: ""}).Normalized()
	if n.MaxOpenConns != 25 || n.MaxIdleConns != 10 || n.ConnMaxLifetime != time.Hour {
		t.Fatalf("%+v", n)
	}

	n = (Pool{MaxOpenConns: 50, MaxIdleConns: 5, ConnMaxLifetime: "30m"}).Normalized()
	if n.MaxOpenConns != 50 || n.MaxIdleConns != 5 || n.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("%+v", n)
	}

	n = (Pool{MaxOpenConns: 8, MaxIdleConns: 20, ConnMaxLifetime: "bogus"}).Normalized()
	if n.MaxOpenConns != 8 {
		t.Fatalf("open=%d", n.MaxOpenConns)
	}
	if n.MaxIdleConns != 8 {
		t.Fatalf("idle should cap at open, got %d", n.MaxIdleConns)
	}
	if n.ConnMaxLifetime != time.Hour {
		t.Fatalf("invalid duration should default to 1h, got %s", n.ConnMaxLifetime)
	}

	n = (Pool{MaxOpenConns: -1}).Normalized()
	if n.MaxOpenConns != 25 {
		t.Fatalf("negative open=%d", n.MaxOpenConns)
	}
}

func TestNormalizedDriver(t *testing.T) {
	cases := map[string]string{
		"":          "postgres",
		"postgres":  "postgres",
		"SQL":       "postgres",
		"redis":     "redis",
		"mongodb":   "mongodb",
		"nosql":     "mongodb",
		"cassandra": "cassandra",
	}
	for in, want := range cases {
		got := (Datastore{Driver: in}).NormalizedDriver()
		if got != want {
			t.Fatalf("%q -> %q want %q", in, got, want)
		}
	}
}

func TestKafkaDefaults(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))
	cfg := Load()
	if cfg.KafkaEnabled {
		t.Fatal("KAFKA_ENABLED should default false")
	}
	if cfg.KafkaClientID != "cabinet-studio-api" {
		t.Fatalf("client_id=%q", cfg.KafkaClientID)
	}
	if len(cfg.KafkaBrokers) != 0 {
		t.Fatalf("brokers=%v", cfg.KafkaBrokers)
	}
	t.Setenv("KAFKA_ENABLED", "true")
	t.Setenv("KAFKA_BROKERS", "kafka-1:9092, kafka-2:9092")
	t.Setenv("KAFKA_CLIENT_ID", "cabinet-test")
	cfg = Load()
	if !cfg.KafkaEnabled {
		t.Fatal("expected enabled")
	}
	if cfg.KafkaClientID != "cabinet-test" {
		t.Fatalf("client_id=%q", cfg.KafkaClientID)
	}
	if len(cfg.KafkaBrokers) != 2 || cfg.KafkaBrokers[0] != "kafka-1:9092" {
		t.Fatalf("brokers=%v", cfg.KafkaBrokers)
	}
}

func TestLogDefaults(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))
	cfg := Load()
	if cfg.Log.Level != "info" {
		t.Fatalf("level=%q", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Fatalf("format=%q", cfg.Log.Format)
	}
	if err := cfg.Log.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(cfg.Log.Sinks) != 1 || cfg.Log.Sinks[0] != utils.SinkStderr {
		t.Fatalf("sinks=%v", cfg.Log.Sinks)
	}
	def := cfg.Log.SinksOrDefault()
	if len(def) != 1 || def[0] != utils.SinkStderr {
		t.Fatalf("SinksOrDefault=%v", def)
	}
}

func TestLogSinkCSV(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))
	t.Setenv("LOG_SINK", "stderr,stdout")
	cfg := Load()
	if len(cfg.Log.Sinks) != 2 || cfg.Log.Sinks[0] != utils.SinkStderr || cfg.Log.Sinks[1] != utils.SinkStdout {
		t.Fatalf("sinks=%v", cfg.Log.Sinks)
	}
	t.Setenv("LOG_SINK", " STDERR , stdout,stderr ")
	cfg = Load()
	if len(cfg.Log.Sinks) != 2 || cfg.Log.Sinks[0] != "stderr" || cfg.Log.Sinks[1] != "stdout" {
		t.Fatalf("lowercase/dedupe sinks=%v", cfg.Log.Sinks)
	}
}

func TestLogSinkLokiLoadsConfig(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))
	t.Setenv("LOG_SINK", "loki")
	t.Setenv("LOKI_URL", "http://loki:3100")
	cfg := Load()
	if len(cfg.Log.Sinks) != 1 || cfg.Log.Sinks[0] != utils.SinkLoki {
		t.Fatalf("sinks=%v", cfg.Log.Sinks)
	}
	if cfg.Log.LokiURL != "http://loki:3100" {
		t.Fatalf("LokiURL=%q", cfg.Log.LokiURL)
	}
	if err := cfg.Log.Validate(); err != nil {
		t.Fatalf("loki must not fail config: %v", err)
	}
}

func TestLogFormatUnsupported(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))
	t.Setenv("LOG_FORMAT", "text")
	cfg := Load()
	if cfg.Log.Format != "text" {
		t.Fatalf("format=%q", cfg.Log.Format)
	}
	err := cfg.Log.Validate()
	if err == nil || !strings.Contains(err.Error(), `log format "text" not supported (want json)`) {
		t.Fatalf("validate=%v", err)
	}
}

func TestLogCloudWatchRegionPrecedence(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", dummyPath(t))
	t.Setenv("CLOUDWATCH_REGION", "us-east-1")
	t.Setenv("CLOUDWATCH_LOG_GROUP", "/cabinet/studio-api")
	t.Setenv("CLOUDWATCH_LOG_STREAM", "api")
	cfg := Load()
	if cfg.Log.CloudWatch.Region != "us-east-1" {
		t.Fatalf("region=%q", cfg.Log.CloudWatch.Region)
	}
	if cfg.Log.CloudWatch.LogGroup != "/cabinet/studio-api" || cfg.Log.CloudWatch.LogStream != "api" {
		t.Fatalf("cloudwatch=%+v", cfg.Log.CloudWatch)
	}
	t.Setenv("AWS_REGION", "ap-south-1")
	cfg = Load()
	if cfg.Log.CloudWatch.Region != "ap-south-1" {
		t.Fatalf("AWS_REGION should win, got %q", cfg.Log.CloudWatch.Region)
	}
}

func TestSinksOrDefaultEmpty(t *testing.T) {
	got := (Log{}).SinksOrDefault()
	if len(got) != 1 || got[0] != utils.SinkStderr {
		t.Fatalf("%v", got)
	}
}
