package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDatastorePath = "configs/datastore.dummy.json"
	defaultMaxOpenConns  = 25
	defaultConnLifetime  = time.Hour
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	OTLPEndpoint   string
	ServiceName    string
	TermsVersion   string
	PrivacyVersion string
	MinAppVersion  string
	TrialDays      int
	LogLevel       string
	RequestTimeout time.Duration
	BodyLimitBytes int64
	OAuthProviders []string
	BBR            BBR
	Datastore      Datastore
	DatastorePath  string
}

// DurationAsString unmarshals JSON durations like "10s".
type DurationAsString time.Duration

func (d DurationAsString) Duration() time.Duration { return time.Duration(d) }

func (d DurationAsString) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *DurationAsString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = DurationAsString(parsed)
	return nil
}

// BBR is Kratos aegis BBR load-shedding config (JSON tags match the product shape).
type BBR struct {
	Enabled      bool             `json:"Enabled"`
	WindowSize   DurationAsString `json:"WindowSize"`
	Buckets      int              `json:"Buckets"`
	CpuThreshold int64            `json:"CpuThreshold"`
}

// Pool is the shared connection-pool shape. SQL, Redis, and NoSQL repeat these fields.
type Pool struct {
	MaxOpenConns    int    `json:"MaxOpenConns"`
	MaxIdleConns    int    `json:"MaxIdleConns"`
	ConnMaxLifetime string `json:"ConnMaxLifetime"` // Go duration, e.g. "1h"
}

// NormalizedPool is Pool after defaults (never pass 0 as sql MaxOpenConns).
type NormalizedPool struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (p Pool) Normalized() NormalizedPool {
	maxOpen := p.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = defaultMaxOpenConns
	}
	maxIdle := p.MaxIdleConns
	if maxIdle < 0 {
		maxIdle = 0
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	life := defaultConnLifetime
	if s := strings.TrimSpace(p.ConnMaxLifetime); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			life = d
		}
	}
	return NormalizedPool{
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: life,
	}
}

type SQLConfig struct {
	Enabled         bool   `json:"Enabled"`
	Driver          string `json:"Driver"` // "postgres" (required for sql)
	DSN             string `json:"DSN"`    // postgres://... local or remote
	MaxOpenConns    int    `json:"MaxOpenConns"`
	MaxIdleConns    int    `json:"MaxIdleConns"`
	ConnMaxLifetime string `json:"ConnMaxLifetime"`
}

func (c SQLConfig) Pool() Pool {
	return Pool{
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}

type RedisConfig struct {
	Enabled         bool   `json:"Enabled"`
	Addr            string `json:"Addr"` // host:port
	Password        string `json:"Password"`
	DB              int    `json:"DB"`
	TLS             bool   `json:"TLS"`
	MaxOpenConns    int    `json:"MaxOpenConns"`    // maps to go-redis PoolSize
	MaxIdleConns    int    `json:"MaxIdleConns"`    // MinIdleConns
	ConnMaxLifetime string `json:"ConnMaxLifetime"` // ConnMaxLifetime
}

func (c RedisConfig) Pool() Pool {
	return Pool{
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}

type NoSQLConfig struct {
	Enabled         bool   `json:"Enabled"`
	URI             string `json:"URI"` // mongodb:// or mongodb+srv://
	Database        string `json:"Database"`
	MaxOpenConns    int    `json:"MaxOpenConns"`    // MaxPoolSize
	MaxIdleConns    int    `json:"MaxIdleConns"`    // MinPoolSize
	ConnMaxLifetime string `json:"ConnMaxLifetime"` // MaxConnIdleTime
}

func (c NoSQLConfig) Pool() Pool {
	return Pool{
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}

type Datastore struct {
	Driver  string      `json:"Driver"`  // single-driver fallback: "postgres" | "sql" | "redis" | "mongodb" | "nosql"
	Drivers []string    `json:"Drivers"` // concurrent adapters, e.g. ["postgres","redis","mongodb"]
	SQL     SQLConfig   `json:"SQL"`
	Redis   RedisConfig `json:"Redis"`
	NoSQL   NoSQLConfig `json:"NoSQL"`
}


// EnabledDrivers is the set of adapters to open at once.
// Priority: Drivers list (or DATASTORE_DRIVERS) → per-block Enabled → single Driver fallback.
func (d Datastore) EnabledDrivers() []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		n := normalizeDriverName(raw)
		if n == "" {
			return
		}
		if n != "postgres" && n != "redis" && n != "mongodb" {
			n = strings.ToLower(strings.TrimSpace(raw))
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	for _, raw := range d.Drivers {
		add(raw)
	}
	if len(out) > 0 {
		return out
	}
	if d.SQL.Enabled {
		add("postgres")
	}
	if d.Redis.Enabled {
		add("redis")
	}
	if d.NoSQL.Enabled {
		add("mongodb")
	}
	if len(out) > 0 {
		return out
	}
	add(d.NormalizedDriver())
	return out
}

func normalizeDriverName(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "postgres", "postgresql", "sql", "pgx":
		if strings.TrimSpace(s) == "" {
			return ""
		}
		return "postgres"
	case "redis":
		return "redis"
	case "mongodb", "mongo", "nosql":
		return "mongodb"
	default:
		return strings.ToLower(strings.TrimSpace(s))
	}
}

// NormalizedDriver returns the adapter key: postgres, redis, or mongodb.
// Empty driver is Postgres (v1 product default). Unknown values are returned lowercased
// so Open can reject them.
func (d Datastore) NormalizedDriver() string {
	switch strings.ToLower(strings.TrimSpace(d.Driver)) {
	case "", "postgres", "postgresql", "sql", "pgx":
		return "postgres"
	case "redis":
		return "redis"
	case "mongodb", "mongo", "nosql":
		return "mongodb"
	default:
		return strings.ToLower(strings.TrimSpace(d.Driver))
	}
}

// ReadyCheckName is the /readyz checks key for the active driver.
func (d Datastore) ReadyCheckName() string {
	switch d.NormalizedDriver() {
	case "redis":
		return "redis"
	case "mongodb":
		return "mongodb"
	default:
		return "postgres"
	}
}

func Load() Config {
	path := resolveDatastorePath()
	ds := defaultDatastore()
	if raw, err := os.ReadFile(path); err == nil {
		var loaded Datastore
		if err := json.Unmarshal(raw, &loaded); err == nil {
			ds = loaded
		}
	}
	applyDatastoreEnv(&ds)

	return Config{
		HTTPAddr:       getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:    ds.SQL.DSN,
		OTLPEndpoint:   getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		ServiceName:    getenv("OTEL_SERVICE_NAME", "cabinet-studio-api"),
		TermsVersion:   getenv("TERMS_VERSION", "2026-06-01"),
		PrivacyVersion: getenv("PRIVACY_VERSION", "2026-06-01"),
		MinAppVersion:  getenv("MIN_APP_VERSION", "1.0.0"),
		TrialDays:      getenvInt("TRIAL_DAYS", 14),
		LogLevel:       getenv("LOG_LEVEL", "info"),
		RequestTimeout: 3 * time.Second,
		BodyLimitBytes: 1 << 20,
		OAuthProviders: []string{"google", "facebook", "x"},
		BBR:            loadBBR(),
		Datastore:      ds,
		DatastorePath:  path,
	}
}

func defaultDatastore() Datastore {
	const poolLife = "1h"
	return Datastore{
		Driver: "postgres",
		SQL: SQLConfig{
			Enabled:         true,
			Driver:          "postgres",
			DSN:             "postgres://cabinet:cabinet@localhost:5432/cabinet_studio?sslmode=disable",
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    10,
			ConnMaxLifetime: poolLife,
		},
		Redis: RedisConfig{
			Addr:            "localhost:6379",
			Password:        "",
			DB:              0,
			TLS:             false,
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    10,
			ConnMaxLifetime: poolLife,
		},
		NoSQL: NoSQLConfig{
			URI:             "mongodb://localhost:27017",
			Database:        "cabinet_studio",
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    10,
			ConnMaxLifetime: poolLife,
		},
	}
}

func resolveDatastorePath() string {
	if p := strings.TrimSpace(os.Getenv("DATASTORE_CONFIG")); p != "" {
		return p
	}
	dir, err := os.Getwd()
	if err != nil {
		return defaultDatastorePath
	}
	for i := 0; i < 8; i++ {
		cand := filepath.Join(dir, defaultDatastorePath)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return defaultDatastorePath
}

func applyDatastoreEnv(ds *Datastore) {
	if v := strings.TrimSpace(os.Getenv("DATASTORE_DRIVER")); v != "" {
		ds.Driver = v
	}
	if v := strings.TrimSpace(os.Getenv("DATASTORE_DRIVERS")); v != "" {
		var parts []string
		for _, p := range strings.Split(v, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) > 0 {
			ds.Drivers = parts
		}
	}
	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		ds.SQL.DSN = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_ADDR")); v != "" {
		ds.Redis.Addr = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_PASSWORD")); v != "" {
		ds.Redis.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_DB")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ds.Redis.DB = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_TLS")); v != "" {
		ds.Redis.TLS = getenvBool("REDIS_TLS", ds.Redis.TLS)
	}
	if v := strings.TrimSpace(os.Getenv("MONGO_URI")); v != "" {
		ds.NoSQL.URI = v
	}
	if v := strings.TrimSpace(os.Getenv("MONGO_DATABASE")); v != "" {
		ds.NoSQL.Database = v
	}
}

func loadBBR() BBR {
	window := 10 * time.Second
	if v := strings.TrimSpace(os.Getenv("BBR_WINDOW_SIZE")); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			window = parsed
		}
	}
	return BBR{
		Enabled:      getenvBool("BBR_ENABLED", true),
		WindowSize:   DurationAsString(window),
		Buckets:      getenvInt("BBR_BUCKETS", 100),
		CpuThreshold: getenvInt64("BBR_CPU_THRESHOLD", 800),
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvInt64(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
