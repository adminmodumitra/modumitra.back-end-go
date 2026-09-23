# App config

The API loads **one** nested JSON, then environment overrides. Empty /
`placeholder` / local-dev values keep fake adapters (FakeProvider, LogMailer,
NoopPublisher). Real credentials select the live adapter — same functions, no
Go change.

| File | Use |
| --- | --- |
| [`cabinet.local.json`](./cabinet.local.json) | Local defaults (`APP_ENV=local`, also the default). Localhost Postgres, Kafka off, billing placeholders, empty OAuth, mail=log, local-dev JWT. Checked in. |
| [`cabinet.dev.json`](./cabinet.dev.json) | Shared/dev compose (`APP_ENV=dev` or `development`). Docker hostnames (`postgres`, `kafka:9092`, `otel:4317`). Kafka on, debug logs, billing still placeholders (`Live()` false). JWT stays the local-dev secret. |
| [`cabinet.production.json`](./cabinet.production.json) | Production (`APP_ENV=production` or `prod`). Same keys, **no real secrets in git**: empty JWT, entitlement key, billing webhook/prices, OAuth clients, Kafka brokers. Fill via env (`JWT_SECRET`, `DATABASE_URL`, `KAFKA_BROKERS`, `BILLING_*`, `OAUTH_*`). Until then `Live()` is false. |
| [`cabinet.dummy.json`](./cabinet.dummy.json) | Alias of `cabinet.local.json` so old `CABINET_CONFIG` paths keep working. |

Load order:

1. `CABINET_CONFIG` if set (explicit path wins)
2. Else `APP_ENV` (or `CABINET_ENV`): `local` (default) → `cabinet.local.json`; `dev`/`development` → `cabinet.dev.json`; `production`/`prod` → `cabinet.production.json`; unknown → local
3. Env wins for secrets (same names as `.env.example`)
4. If the cabinet file is missing, in-memory defaults (localhost Postgres)

Flip local → live by editing the JSON (or env):

- Billing: set `billing.webhook_secret` and replace `price_*_placeholder` with Stripe price ids (`price_1…`). `provider` `auto` then selects StripeProvider (stub until live HTTP ships). Or set `provider` to `stripe` / `fake`. Production JSON leaves these empty so env must fill them.
- OAuth: set `oauth.google` / `facebook` / `x` `client_id` and `client_secret` (values containing `placeholder` stay disabled).
- Kafka: `kafka.enabled` true and a non-empty `brokers` list (production: `KAFKA_BROKERS`).
- Mail: `mail.driver` `log` (default). `smtp` is reserved and fails boot.

`Driver` is the datastore switch: `postgres` (v1 default), `sql` (same Postgres adapter), `redis`, `mongodb`, or `nosql`.

Point DSN / URI / Addr at localhost or a remote host — same binary. Overrides (no code change):

- `http.allowed_ips` — JSON array of client IPs or CIDRs allowed to call the API. Empty allows every IP. A non-empty list returns 401 for any other IP. `/healthz` and `/readyz` stay open. `ALLOWED_IPS` is the same JSON array and overrides the file, for example `["10.0.0.8","192.168.1.0/24"]`.
- `APP_ENV` / `CABINET_ENV` — selects which cabinet JSON to load when `CABINET_CONFIG` is unset
- `CABINET_CONFIG` — explicit path to the nested JSON (wins over `APP_ENV`)
- `DATASTORE_DRIVER` — overrides `Driver`
- `DATABASE_URL` — overrides `SQL.DSN`
- `REDIS_ADDR`, `REDIS_PASSWORD` — Redis
- `MONGO_URI`, `MONGO_DATABASE` — MongoDB
- `JWT_SECRET` — required in production (file secret is empty)

Pool fields (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`) apply to every driver. `MaxOpenConns` of `0` is treated as 25 (`database/sql` would take 0 as unlimited).
