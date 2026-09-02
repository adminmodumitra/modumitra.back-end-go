# Cabinet Studio API — slice 0

Go + Gin skeleton for the Cabinet Studio commercial backend. The live datastore is selected from JSON + env (Postgres, Redis, or Mongo) with real connection pools.

This is **slice 0**: process health, datastore readiness, public bootstrap config, twelve middlewares, and one W3C `trace_id` through Gin and `store.ping`. v1 product default remains Postgres. **Identity (register/login/JWT issuance) is slice 1.** Kafka, Stripe, and SSO are not in this slice.

Module: `github.com/Danish-13-Khan/backend-cabinet`

## What is running

| Piece | Notes |
| --- | --- |
| HTTP | Gin on `:8080` |
| Datastore | Config-driven. v1 default **Postgres 16** (`cabinet_studio` / `cabinet`). Redis and Mongo adapters are live (not stubs) when `Driver` is switched. |
| Schema | **19 tables** applied from [`schema-v1.sql`](./schema-v1.sql) on first Postgres boot |
| Tracing | W3C `traceparent`. OTLP → collector → Tempo. Grafana at `:3000` |

## Routes (only these)

| Method | Path | Auth |
| --- | --- | --- |
| `GET` | `/healthz` | none — liveness |
| `GET` | `/readyz` | none — ping of the active store (`postgres` / `redis` / `mongodb`) |
| `GET` | `/v1/public/config` | public |
| any | other `/v1/*` | **401** `unauthenticated` (JWT middleware; no tokens in slice 0) |

Every JSON body includes `request_id` and `trace_id`. Errors use:

```json
{
  "error": {
    "code": "stable_snake_case",
    "message": "user-facing",
    "request_id": "req_…",
    "trace_id": "32 lowercase hex",
    "details": {}
  }
}
```

Response headers always: `X-Request-Id`, `X-Trace-Id` (32 hex), `traceparent`. If the client sends `traceparent`, that `trace_id` is continued. Child spans (for example the readyz `store.ping`) stay on the same trace (`db.system` set on the span).

## Run with Docker Compose

Requires Docker. From the repo root:

```bash
make up
# or: docker compose up --build -d
```

First Postgres start mounts `schema-v1.sql` into `docker-entrypoint-initdb.d` and creates all **19** v1 tables.

```bash
curl -sS http://localhost:8080/healthz
curl -sS http://localhost:8080/readyz
curl -sS http://localhost:8080/v1/public/config
```

Continue a trace:

```bash
curl -sS -D - \
  -H 'traceparent: 00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01' \
  -H 'X-Request-Id: req_client_example' \
  http://localhost:8080/healthz
```

`X-Trace-Id` and the JSON `trace_id` must equal `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`.

Stop:

```bash
make down
# or: docker compose down
```

Grafana (anonymous): http://localhost:3000 — Tempo / Loki / Prometheus datasources are provisioned. Tempo search for the API service name `cabinet-studio-api`.

## Run the API locally (Go 1.22+)

Postgres must already be up (`docker compose up -d postgres` is enough).

```bash
cp .env.example .env   # optional; the process reads the environment
export CABINET_CONFIG=configs/cabinet.dummy.json
export DATABASE_URL=postgres://cabinet:cabinet@localhost:5432/cabinet_studio?sslmode=disable
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
go run ./cmd/api
```

Local dummy JSON uses billing `price_*_placeholder`, empty OAuth, and Kafka off — FakeProvider + LogMailer + NoopPublisher. Replace those values in [`configs/cabinet.dummy.json`](./configs/cabinet.dummy.json) or the matching env (`BILLING_PRICE_*`, `OAUTH_*`, `KAFKA_BROKERS`) to select the live adapter. No Go change. StripeProvider is still a stub (no live Stripe HTTP).

To point at a remote server, keep the cabinet JSON and override `DATABASE_URL` / `REDIS_ADDR` / `MONGO_URI`. Switch engines with `DATASTORE_DRIVER=redis` or `mongodb`. Logs never print raw DSN, password, or URI.

Pool settings in the dummy file (`MaxOpenConns` 25, `MaxIdleConns` 10, `ConnMaxLifetime` `1h`) map onto `sql.DB`, go-redis, and the Mongo driver. `MaxOpenConns` of 0 is treated as 25.

```bash
make test    # go test ./...
make build   # bin/api
```

Unit tests use a fake pinger (no Docker, no live Redis/Mongo) plus httptest for `traceparent` echo. Config load, DSN redaction, pool defaulting, and unknown-driver errors are covered in-process.

## Environment

See [`.env.example`](./.env.example). Do not commit `.env`.

| Variable | Default | Purpose |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Listen address |
| `CABINET_CONFIG` | `configs/cabinet.local.json` | Nested app JSON (http, log, datastore, jwt, billing, oauth, kafka, mail). See [`configs/`](./configs/) |
| `DATASTORE_DRIVER` | (file) | `postgres` / `sql` / `redis` / `mongodb` / `nosql`. Unknown driver refuses boot |
| `DATABASE_URL` | (file SQL.DSN) | Overrides `SQL.DSN` (local or remote Postgres) |
| `REDIS_ADDR` | (file) | Overrides `Redis.Addr` (`host:port`) |
| `REDIS_PASSWORD` | (file) | Overrides `Redis.Password` |
| `MONGO_URI` | (file) | Overrides `NoSQL.URI` |
| `MONGO_DATABASE` | (file) | Overrides `NoSQL.Database` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | Collector gRPC. If unreachable, traces are still generated locally |
| `OTEL_SERVICE_NAME` | `cabinet-studio-api` | Resource service.name |
| `TERMS_VERSION` | `2026-06-01` | `GET /v1/public/config` |
| `PRIVACY_VERSION` | `2026-06-01` | `GET /v1/public/config` |
| `MIN_APP_VERSION` | `1.0.0` | `GET /v1/public/config` |
| `TRIAL_DAYS` | `14` | Always 14 for v1 |
| `LOG_LEVEL` | `info` | slog level |
| `BBR_ENABLED` | `true` | Kratos BBR load-shedding; `false` is pass-through |
| `BBR_WINDOW_SIZE` | `10s` | BBR window (`WithWindow`) |
| `BBR_BUCKETS` | `100` | BBR buckets (`WithBucket`) |
| `BBR_CPU_THRESHOLD` | `800` | millicpu 0–1000 (`WithCPUThreshold`) |

## Middleware order (12)

1. RequestID — `req_…` if `X-Request-Id` is missing  
2. Recover — panic → 500 error envelope, no stack leak  
3. OTEL — `otelgin` + W3C; echo `X-Trace-Id` / `traceparent`  
4. BBR — Kratos aegis; uses `Enabled` / `WindowSize` / `Buckets` / `CpuThreshold`; process-wide, not per-IP; skips `/healthz` `/readyz`  
5. Timeout — 3s request context  
6. BodyLimit — 1 MiB  
7. IPRateLimit — in-memory token bucket; `/healthz` and `/readyz` skipped  
8. UserRateLimit — no-op when there is no user (always in slice 0)  
9. JWTAuth — skip `/healthz`, `/readyz`, `/v1/public/config`; else 401  
10. Tenant — no-op (no org on these routes)  
11. RBAC — no-op  
12. Idempotency — if `Idempotency-Key` is present, remember the response in memory for process lifetime; else pass  

## Layout

```
cmd/api/main.go
configs/cabinet.local.json
configs/cabinet.dev.json
configs/cabinet.production.json
internal/config
internal/http
internal/http/middleware
internal/obs
internal/store          # Postgres (database/sql + pgx stdlib), Redis, Mongo adapters
schema-v1.sql           # 19 tables; applied on first compose Postgres boot
```
