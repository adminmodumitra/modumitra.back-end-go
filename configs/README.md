# Datastore config

The API picks **one** live store from JSON, then environment overrides.

| File | Use |
| --- | --- |
| [`datastore.dummy.json`](./datastore.dummy.json) | Local Docker defaults (Postgres on `localhost:5432`). Checked in. |
| [`datastore.remote.example.json`](./datastore.remote.example.json) | Placeholder host/user/password for a remote server. Copy, fill in, point `DATASTORE_CONFIG` at the copy. Do not commit secrets. |

`Driver` is the switch: `postgres` (v1 default), `sql` (same Postgres adapter), `redis`, `mongodb`, or `nosql`.

Point DSN / URI / Addr at localhost or a remote host — same binary. Overrides (no code change):

- `DATASTORE_CONFIG` — path to this JSON (default: `configs/datastore.dummy.json`)
- `DATASTORE_DRIVER` — overrides `Driver`
- `DATABASE_URL` — overrides `SQL.DSN`
- `REDIS_ADDR`, `REDIS_PASSWORD` — Redis
- `MONGO_URI`, `MONGO_DATABASE` — MongoDB

Pool fields (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`) apply to every driver. `MaxOpenConns` of `0` is treated as 25 (`database/sql` would take 0 as unlimited).
