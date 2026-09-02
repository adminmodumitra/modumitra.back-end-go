# Cabinet Studio PostgreSQL Schema v1

**Document role:** Canonical PostgreSQL relational schema for the Cabinet Studio commercial backend  
**Audience:** Go + Gin implementers, QA, Neon/RDS operators  
**API twin:** [`API_CONTRACTS_V1.md`](./API_CONTRACTS_V1.md) (64 unique method+path pairs) and [`openapi-v1.yaml`](./openapi-v1.yaml)  
**Database:** `cabinet_studio`  
**Schema:** `public`  
**Tables:** 19 (no extras in v1)  
**Baseline date:** 2026-09-01  
**Status:** Approved schema for implementation; not evidence that tables are deployed  

Runnable DDL: [`schema-v1.sql`](./schema-v1.sql).  
Machine-readable companion: [`schema-v1.json`](./schema-v1.json).

---

## 0. Front matter

| Item | Value |
| --- | --- |
| Engine | PostgreSQL **16+** (managed: Neon or AWS RDS). MongoDB is **not** used. |
| Database name | `cabinet_studio` |
| Schema | `public` |
| Tables | **19** — same names as the former document-store collections |
| API routes | **64** unique method+path pairs (unchanged) |
| Baseline | **2026-09-01** |
| Public / row primary key | `public_id TEXT PRIMARY KEY` with opaque prefixed strings (`usr_`, `org_`, …). **Do not** use `SERIAL` / `BIGSERIAL` as public API ids. Internal bigserial is not used in v1. |
| Timestamps | `TIMESTAMPTZ NOT NULL DEFAULT now()`, stored UTC. JSON wire format is ISO-8601 with `Z`. Set `timezone = 'UTC'` on connections. |
| Enumerations | **TEXT + named CHECK** constraints (not PostgreSQL `ENUM` types). CHECK is easier to migrate (`ALTER TABLE … DROP/ADD CONSTRAINT`) than `ALTER TYPE`. Choice is consistent on all 19 tables. |
| Field names | `snake_case` (SQL and API) |
| Tenant isolation field | `organization_id` = `organizations.public_id` (`org_…`) |
| Referential integrity | Real `FOREIGN KEY` constraints (see §1.6). Application transactions still required for multi-row invariants (last owner, seat counts, refresh-family revoke). |
| Soft delete | Status/state columns. **No physical DELETE** of tenant rows in app code. Physical DELETE is only for TTL-like cleanup jobs (§6). |
| JSONB | `plans.capabilities`, `entitlements.capabilities`, `audit_events.metadata`, `billing_events.summary`, `idempotency_keys.response_body` |
| BYTEA | `outbox.payload` (marshaled `cabinet.v1.Envelope` protobuf; never JSON on the wire) |
| Password hash | argon2id PHC string. Never plaintext. |
| Token hash | SHA-256 of `pepper || raw_token` (hex). Never raw tokens. No PAN/CVC. |
| Access tokens | Signed JWT; **not stored** in PostgreSQL |
| Kafka | Downstream of `outbox`. Kafka is **not** the source of truth. |
| Scale | Managed Postgres (Neon/RDS) + a read replica later. **No Citus** required for v1. |

This schema is the persistence counterpart of the 64-route HTTP contract. Handlers commit application rows and an outbox row in the **same PostgreSQL transaction**; a worker publishes to Kafka after commit (`API_CONTRACTS_V1.md` §1.5).

---

## 1. Global conventions

### 1.1 Public IDs

Generate 128 bits of CSPRNG (or ULID) and encode Crockford-base32 / lowercase hex, then prefix. Sequential integers must never appear in API responses, logs that leave the security boundary, or Kafka payloads that clients can see.

Each table CHECKs `public_id ~ '^<prefix>_'`.

| Prefix | Table | Typical API exposure |
| --- | --- | --- |
| `usr_` | `users` | yes (`id`) |
| `oid_` | `oauth_identities` | no (addressed by `provider`) |
| `org_` | `organizations` | yes |
| `mbr_` | `memberships` | no (paths use `userId`) |
| `inv_` | `invitations` | yes |
| `dev_` | `devices` | yes |
| `ses_` | `sessions` | yes |
| `tok_` | `email_tokens` | no (raw token uses `ev_` / `pr_` / `ec_`) |
| `st_` | `oauth_states` | yes (`state`) |
| `dc_` | `device_codes` | yes (`device_code`) |
| `plan_` | `plans` | yes |
| `sub_` | `subscriptions` | yes |
| `invc_` | `invoices` | yes |
| `bev_` | `billing_events` | no |
| `ent_` | `entitlements` | no (snapshot is returned without this id) |
| `evt_` | `audit_events` | yes |
| `obx_` | `outbox` | no |
| `idk_` | `idempotency_keys` | no |
| `exp_` | `data_exports` | yes (`export_id`) |

`oid_` is the only prefix not listed in the original HTTP ID-prefix set; it exists so every table still has a unique `public_id`. The HTTP contract never returns it.

Raw one-time secrets (never persisted):

| Prefix | Secret | Stored as |
| --- | --- | --- |
| `ev_` | email verification token | `email_tokens.token_hash` |
| `pr_` | password-reset token | `email_tokens.token_hash` |
| `ec_` | email-change token | `email_tokens.token_hash` |
| `rt_` | refresh token | `sessions.refresh_token_hash` |
| (email link) | invitation token | `invitations.token_hash` |

### 1.2 Tenant vs user-owned

Every row that belongs to a cabinet business **must** include `organization_id` (`org_…` string) except where noted nullable.

**Tenant-scoped (always have `organization_id` except documented nullables):**  
`organizations` (self id), `memberships`, `invitations`, `subscriptions`, `invoices`, `billing_events` (nullable if provider event arrives before org map), `entitlements` (nullable only when the user has no org yet), `audit_events` (nullable FK), `outbox` (partition key is usually org, no FK).

**User-owned / global (no required `organization_id`):**  
`users`, `oauth_identities`, `devices`, `sessions`, `email_tokens`, `oauth_states`, `device_codes`, `data_exports`, `idempotency_keys`, `plans`.

`idempotency_keys.principal` is `usr_…` when authenticated, else `ip:<sha256>` plus route for public auth.

### 1.3 SQL / JSON mapping

| Logical type | PostgreSQL | JSON in samples below |
| --- | --- | --- |
| Public id | `TEXT` PK | `"usr_ada01f3e8c2b4a90"` |
| Instant | `TIMESTAMPTZ` | `"2026-09-01T00:00:00Z"` (UTC) |
| Enum | `TEXT` + `CHECK` | string |
| Money | `INTEGER` minor units | integer cents |
| Hash | `TEXT` | `"sha256:…"` or PHC |
| Map | `JSONB` | object |
| Null | `NULL` | `null` |
| Bool | `BOOLEAN` | boolean |
| Int | `INTEGER` / `BIGINT` | integer |

`now()` returns `timestamptz`; Postgres stores it as UTC. Always serialize API timestamps with a `Z` suffix.

### 1.4 Soft delete and status

Do not physically delete tenant rows in v1 except cleanup-job tables. Use status/state fields (`removed`, `deleted`, `revoked`, `deactivated`). `DELETE` HTTP routes are logical.

### 1.5 Hashing

| Secret | Algorithm | Notes |
| --- | --- | --- |
| Password | argon2id (`m=65536,t=3,p=4`, PHC) | Nullable if OAuth-only |
| Refresh / invite / email / device secrets | SHA-256(`pepper || raw`) hex | Pepper in KMS / env, not in Postgres |
| Idempotency body | SHA-256 of canonical JSON | Detect key reuse with different body |

Never log raw secrets. Never store provider OAuth access/refresh tokens.

### 1.6 Foreign keys (real CONSTRAINT)

All FKs reference `public_id`. Count: **28**.

| Constraint | From | To | ON DELETE |
| --- | --- | --- | --- |
| `fk_oauth_identities_user_id` | `oauth_identities.user_id` | `users(public_id)` | CASCADE |
| `fk_organizations_owner_user_id` | `organizations.owner_user_id` | `users(public_id)` | RESTRICT, **DEFERRABLE INITIALLY DEFERRED** |
| `fk_memberships_organization_id` | `memberships.organization_id` | `organizations(public_id)` | CASCADE |
| `fk_memberships_user_id` | `memberships.user_id` | `users(public_id)` | RESTRICT |
| `fk_invitations_organization_id` | `invitations.organization_id` | `organizations(public_id)` | CASCADE |
| `fk_invitations_invited_by_user_id` | `invitations.invited_by_user_id` | `users(public_id)` | RESTRICT |
| `fk_invitations_accepted_by_user_id` | `invitations.accepted_by_user_id` | `users(public_id)` | RESTRICT |
| `fk_invitations_revoked_by_user_id` | `invitations.revoked_by_user_id` | `users(public_id)` | RESTRICT |
| `fk_devices_user_id` | `devices.user_id` | `users(public_id)` | CASCADE |
| `fk_sessions_user_id` | `sessions.user_id` | `users(public_id)` | CASCADE |
| `fk_sessions_device_id` | `sessions.device_id` | `devices(public_id)` | SET NULL |
| `fk_sessions_replaced_by_session_id` | `sessions.replaced_by_session_id` | `sessions(public_id)` | SET NULL |
| `fk_email_tokens_user_id` | `email_tokens.user_id` | `users(public_id)` | CASCADE |
| `fk_oauth_states_user_id` | `oauth_states.user_id` | `users(public_id)` | CASCADE |
| `fk_device_codes_user_id` | `device_codes.user_id` | `users(public_id)` | SET NULL |
| `fk_subscriptions_organization_id` | `subscriptions.organization_id` | `organizations(public_id)` | RESTRICT (**UNIQUE**) |
| `fk_subscriptions_plan_id` | `subscriptions.plan_id` | `plans(public_id)` | RESTRICT |
| `fk_invoices_organization_id` | `invoices.organization_id` | `organizations(public_id)` | RESTRICT |
| `fk_invoices_subscription_id` | `invoices.subscription_id` | `subscriptions(public_id)` | SET NULL |
| `fk_billing_events_organization_id` | `billing_events.organization_id` | `organizations(public_id)` | SET NULL (nullable) |
| `fk_billing_events_subscription_id` | `billing_events.subscription_id` | `subscriptions(public_id)` | SET NULL |
| `fk_entitlements_user_id` | `entitlements.user_id` | `users(public_id)` | CASCADE |
| `fk_entitlements_organization_id` | `entitlements.organization_id` | `organizations(public_id)` | SET NULL |
| `fk_entitlements_device_id` | `entitlements.device_id` | `devices(public_id)` | SET NULL |
| `fk_entitlements_plan_id` | `entitlements.plan_id` | `plans(public_id)` | SET NULL |
| `fk_audit_events_organization_id` | `audit_events.organization_id` | `organizations(public_id)` | SET NULL (nullable) |
| `fk_audit_events_actor_user_id` | `audit_events.actor_user_id` | `users(public_id)` | SET NULL (nullable) |
| `fk_data_exports_user_id` | `data_exports.user_id` | `users(public_id)` | CASCADE |

`outbox` has **no FK**. `partition_key` is an org (or user) id **string**; the org may be gone after enqueue.

**Circular-FK watch.** `organizations.owner_user_id` → `users` and `memberships.organization_id` → `organizations` are not a true cycle, but org-create writes both an organization and an owner membership in one transaction. `fk_organizations_owner_user_id` is `DEFERRABLE INITIALLY DEFERRED` so a future “owner must be a member” check can run at `COMMIT`. **Create organizations first, then memberships.**

`idempotency_keys.principal` is **not** an FK (value may be `ip:<sha256>` or `wh:stripe`).

---

## 2. ER diagram

```mermaid
erDiagram
  users ||--o{ oauth_identities : has
  users ||--o{ sessions : has
  users ||--o{ devices : has
  users ||--o{ email_tokens : has
  users ||--o{ data_exports : has
  users ||--o{ memberships : joins
  users ||--o{ entitlements : issued
  users ||--o{ organizations : owns
  users ||--o{ invitations : invites
  organizations ||--o{ memberships : has
  organizations ||--o{ invitations : issues
  organizations ||--|{ subscriptions : billed_as
  organizations ||--o{ invoices : has
  organizations ||--o{ billing_events : receives
  organizations ||--o{ entitlements : grants
  organizations ||--o{ audit_events : records
  plans ||--o{ subscriptions : catalogs
  plans ||--o{ entitlements : snapshots
  subscriptions ||--o{ invoices : generates
  devices ||--o{ sessions : may_bind
  devices ||--o{ entitlements : may_bind
  sessions }o--o| sessions : replaced_by
  oauth_states }o--o| users : link_intent
  device_codes }o--o| users : authorized

  users {
    text public_id PK
    text email_normalized UK
  }
  oauth_identities {
    text public_id PK
    text user_id FK
    text provider
    text provider_user_id
  }
  organizations {
    text public_id PK
    text owner_user_id FK
  }
  memberships {
    text public_id PK
    text organization_id FK
    text user_id FK
  }
  invitations {
    text public_id PK
    text organization_id FK
    text token_hash UK
    text invited_by_user_id FK
  }
  devices {
    text public_id PK
    text user_id FK
  }
  sessions {
    text public_id PK
    text user_id FK
    text refresh_token_hash UK
    text device_id FK
    text replaced_by_session_id FK
    text family_id
  }
  email_tokens {
    text public_id PK
    text token_hash UK
    text user_id FK
  }
  oauth_states {
    text public_id PK
    text state UK
    text user_id FK
  }
  device_codes {
    text public_id PK
    text device_code UK
    text user_code UK
    text user_id FK
  }
  plans {
    text public_id PK
  }
  subscriptions {
    text public_id PK
    text organization_id UK_FK
    text plan_id FK
  }
  invoices {
    text public_id PK
    text organization_id FK
    text subscription_id FK
    text provider_invoice_id UK
  }
  billing_events {
    text public_id PK
    text provider
    text provider_event_id
    text organization_id FK
  }
  entitlements {
    text public_id PK
    text organization_id FK
    text user_id FK
    text device_id FK
  }
  audit_events {
    text public_id PK
    text organization_id FK
    text actor_user_id FK
  }
  outbox {
    text public_id PK
    text event_id UK
    text partition_key
  }
  idempotency_keys {
    text public_id PK
    text principal
    text key
  }
  data_exports {
    text public_id PK
    text user_id FK
  }
```

---

## 3. Tables

Null column: `NO` = `NOT NULL`. Default `now()` means `DEFAULT now()` UTC timestamptz.

### 3.1 `users`

**Purpose.** Human identity, credentials, legal acceptance, and account lifecycle. One row per person. Not tenant-scoped.

**PK.** `public_id TEXT PRIMARY KEY` (`usr_`).

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | generated `usr_` | PK; API `user.id` |
| `email` | `TEXT` | NO | — | Lowercase display email, max 254 |
| `email_normalized` | `TEXT` | NO | `lower(trim(email))` | Unique lookup key |
| `email_verified_at` | `TIMESTAMPTZ` | YES | `NULL` | Null ⇒ API `email_verified=false` |
| `password_hash` | `TEXT` | YES | `NULL` | argon2id PHC; null if OAuth-only |
| `password_updated_at` | `TIMESTAMPTZ` | YES | `NULL` | Set on set/change/reset |
| `display_name` | `TEXT` | NO | — | 1–80 chars |
| `terms_version` | `TEXT` | NO | — | Must match public config at accept time |
| `privacy_version` | `TEXT` | NO | — | Must match public config at accept time |
| `terms_accepted_at` | `TIMESTAMPTZ` | NO | `now()` | Legal acceptance instant |
| `pending_email` | `TEXT` | YES | `NULL` | Target of in-flight email change (normalized) |
| `status` | `TEXT` | NO | `'active'` | CHECK `active` \| `disabled` \| `pending_deletion` |
| `disabled_at` | `TIMESTAMPTZ` | YES | `NULL` | When `status=disabled` |
| `deletion_requested_at` | `TIMESTAMPTZ` | YES | `NULL` | Account deletion request; org ownership must be clear |
| `last_login_at` | `TIMESTAMPTZ` | YES | `NULL` | Last successful login/oauth/device-token |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

`disabled` maps to HTTP `403 account_suspended`. Cloud sharing and billing still require `email_verified_at IS NOT NULL` (`AUT-001`).

#### Foreign keys

None outbound. Referenced by user-owned tables via `user_id` → `users.public_id`.

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `users_pkey` | PK unique | `public_id` |
| `uniq_users_email_normalized` | UNIQUE | `email_normalized` |
| `idx_users_status_created_at` | index | `(status, created_at DESC)` |
| `idx_users_pending_email` | partial index | `pending_email` WHERE NOT NULL |

#### Sample row

```json
{
  "public_id": "usr_ada01f3e8c2b4a90",
  "email": "ada@example.com",
  "email_normalized": "ada@example.com",
  "email_verified_at": "2026-08-01T10:12:00Z",
  "password_hash": "$argon2id$v=19$m=65536,t=3,p=4$REDACTED_SALT$REDACTED_HASH",
  "password_updated_at": "2026-08-01T10:00:00Z",
  "display_name": "Ada Lovelace",
  "terms_version": "2026-06-01",
  "privacy_version": "2026-06-01",
  "terms_accepted_at": "2026-08-01T10:00:00Z",
  "pending_email": null,
  "status": "active",
  "disabled_at": null,
  "deletion_requested_at": null,
  "last_login_at": "2026-09-01T00:00:00Z",
  "created_at": "2026-08-01T10:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/auth/register`, `POST /v1/auth/verify-email`, `POST /v1/auth/password-reset/confirm`, `PATCH /v1/me`, `POST /v1/me/password`, `POST /v1/me/email/change`, `POST /v1/me/email/change/confirm`, `GET /v1/auth/oauth/{provider}/callback` (create on first login), `POST /v1/auth/device/token` (touch `last_login_at`) |
| R | `GET /v1/me`, login/refresh/oauth success bodies, member directory projections, invitation preview (`inviter_display_name`) |
| Auth lookup | `POST /v1/auth/login`, `POST /v1/auth/resend-verification`, `POST /v1/auth/password-reset` (existence-oracle safe) |

---

### 3.2 `oauth_identities`

**Purpose.** Links a Cabinet Studio user to one social/SSO subject per provider (`google` \| `facebook` \| `x`). Stores no provider tokens.

**PK.** `public_id TEXT PRIMARY KEY` (`oid_`). Not returned by API.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `oid_` | PK; not returned |
| `user_id` | `TEXT` | NO | — | → `users.public_id` |
| `provider` | `TEXT` | NO | — | CHECK `google` \| `facebook` \| `x` |
| `provider_user_id` | `TEXT` | NO | — | Provider subject; not a secret |
| `email` | `TEXT` | YES | `NULL` | Email asserted by provider (normalized) |
| `email_verified_by_provider` | `BOOLEAN` | NO | `FALSE` | Used when creating a new user on login intent |
| `linked_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Unlink refuses if it would leave the user with neither `password_hash` nor another identity.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_oauth_identities_user_id` | `user_id` | `users(public_id)` | CASCADE |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `oauth_identities_pkey` | PK unique | `public_id` |
| `uniq_oauth_identities_provider_subject` | UNIQUE | `(provider, provider_user_id)` |
| `uniq_oauth_identities_user_provider` | UNIQUE | `(user_id, provider)` |

The two compound uniques cover list-by-user and login-by-subject.

#### Sample row

```json
{
  "public_id": "oid_gada01c9e4b7a210",
  "user_id": "usr_ada01f3e8c2b4a90",
  "provider": "google",
  "provider_user_id": "gid_opaque",
  "email": "ada@example.com",
  "email_verified_by_provider": true,
  "linked_at": "2026-08-12T09:00:00Z",
  "created_at": "2026-08-12T09:00:00Z",
  "updated_at": "2026-08-12T09:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| R | `GET /v1/me/identities` |
| W | `GET /v1/auth/oauth/{provider}/callback` (login upsert), `POST /v1/me/identities/{provider}/link`, `DELETE /v1/me/identities/{provider}`, `POST /v1/auth/device/token` |

---

### 3.3 `organizations`

**Purpose.** Cabinet business, billing boundary, and shared-data owner. Trial clock starts at create (`trial_ends_at` = `created_at + 14d`).

**PK.** `public_id TEXT PRIMARY KEY` (`org_`). `organization_id` on other tables equals this `public_id`.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `org_` | API `organization.id`; tenant key |
| `name` | `TEXT` | NO | — | 1–120 chars |
| `owner_user_id` | `TEXT` | NO | — | → `users.public_id`; denormalized from membership role=owner |
| `status` | `TEXT` | NO | `'active'` | CHECK `active` \| `pending_deletion` \| `deleted` |
| `trial_ends_at` | `TIMESTAMPTZ` | YES | create+14d | Kept after conversion for history |
| `provider` | `TEXT` | YES | `NULL` | CHECK `stripe` \| `fake` once a customer exists |
| `provider_customer_id` | `TEXT` | YES | `NULL` | Stripe `cus_…` / fake id. **Not** a card. Partial unique |
| `deleted_at` | `TIMESTAMPTZ` | YES | `NULL` | API treats as gone |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

v1 `DELETE /v1/organizations/{id}` sets `status=deleted` immediately for API purposes. Member personal accounts are not deleted. Close is blocked while a cancelable paid subscription is still `active`/`trialing`/`past_due`/`grace` (`409 active_subscription`).

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_organizations_owner_user_id` | `owner_user_id` | `users(public_id)` | RESTRICT, DEFERRABLE INITIALLY DEFERRED |

Owner must also be an `active` owner membership (application invariant).

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `organizations_pkey` | PK unique | `public_id` |
| `uniq_organizations_provider_customer_id` | partial unique | `provider_customer_id` WHERE NOT NULL |
| `idx_organizations_owner_user_id` | index | `owner_user_id` |
| `idx_organizations_status_created_at` | index | `(status, created_at DESC)` |

#### Sample row

```json
{
  "public_id": "org_studio9d2e4c6a80",
  "name": "Lovelace Cabinets",
  "owner_user_id": "usr_ada01f3e8c2b4a90",
  "status": "active",
  "trial_ends_at": "2026-09-15T00:00:00Z",
  "provider": "fake",
  "provider_customer_id": null,
  "deleted_at": null,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/organizations`, `PATCH /v1/organizations/{id}`, `DELETE /v1/organizations/{id}`, `POST /v1/organizations/{id}/transfer-ownership`, checkout/webhook (customer id) |
| R | `GET /v1/organizations`, `GET /v1/organizations/{id}`, login/oauth/device-token org projection, invitation accept |

---

### 3.4 `memberships`

**Purpose.** User ↔ organization relationship with a role and lifecycle state. Unique per `(organization_id, user_id)` including `removed` rows (reuse the same row).

**PK.** `public_id TEXT PRIMARY KEY` (`mbr_`).

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `mbr_` | Not used in HTTP paths |
| `organization_id` | `TEXT` | NO | — | → `organizations.public_id` |
| `user_id` | `TEXT` | NO | — | → `users.public_id` |
| `role` | `TEXT` | NO | — | CHECK `owner` \| `admin` \| `billing_admin` \| `sales_designer` \| `engineer` \| `production` \| `viewer` |
| `state` | `TEXT` | NO | `'active'` | CHECK `invited` \| `active` \| `suspended` \| `removed`. API field name is `status` |
| `invited_at` | `TIMESTAMPTZ` | YES | `NULL` | When state first became `invited` |
| `activated_at` | `TIMESTAMPTZ` | YES | `NULL` | When first became `active` |
| `suspended_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `removed_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

**App rules.** Last `owner` cannot leave, be removed, or be demoted (`ORG-004`, `409 last_owner`). `state=invited` is only used when the invitee already has a user id; pending email-only invites live in `invitations` until accept, which upserts `state=active`. v1: `active` members consume a seat; `suspended`/`removed`/`invited` do not. Removal does not delete org-owned projects. Role/state changes must take effect without waiting for access-token expiry (re-check membership every org-scoped request).

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_memberships_organization_id` | `organization_id` | `organizations(public_id)` | CASCADE |
| `fk_memberships_user_id` | `user_id` | `users(public_id)` | RESTRICT |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `memberships_pkey` | PK unique | `public_id` |
| `uniq_memberships_org_user` | UNIQUE | `(organization_id, user_id)` |
| `idx_memberships_user_state` | index | `(user_id, state)` |
| `idx_memberships_org_state_role` | index | `(organization_id, state, role)` |

`GET /v1/organizations` uses `idx_memberships_user_state`.

#### Sample row

```json
{
  "public_id": "mbr_own01ada9c2e4f180",
  "organization_id": "org_studio9d2e4c6a80",
  "user_id": "usr_ada01f3e8c2b4a90",
  "role": "owner",
  "state": "active",
  "invited_at": null,
  "activated_at": "2026-09-01T00:00:00Z",
  "suspended_at": null,
  "removed_at": null,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/organizations` (owner), `POST /v1/invitations/{token}/accept`, `PATCH /v1/organizations/{id}/members/{userId}`, `POST …/suspend`, `DELETE …/members/{userId}`, `POST …/leave`, `POST …/transfer-ownership` |
| R | `GET /v1/organizations/{id}/members`, `GET /v1/organizations`, every org-scoped authz check, `GET /v1/me/capabilities`, `GET /v1/organizations/{id}/seats` |

---

### 3.5 `invitations`

**Purpose.** Time-limited, single-use offer to join an organization at a non-owner role. Stores `token_hash` only.

**PK.** `public_id TEXT PRIMARY KEY` (`inv_`).

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `inv_` | API `invitation.id` |
| `organization_id` | `TEXT` | NO | — | → `organizations.public_id` |
| `email` | `TEXT` | NO | — | Invitee, lowercase |
| `role` | `TEXT` | NO | — | Any role **except** `owner` |
| `status` | `TEXT` | NO | `'pending'` | CHECK `pending` \| `accepted` \| `revoked` \| `expired` |
| `token_hash` | `TEXT` | NO | — | SHA-256(pepper\|\|raw). Unique. Never raw token |
| `invited_by_user_id` | `TEXT` | NO | — | → `users.public_id` |
| `expires_at` | `TIMESTAMPTZ` | NO | now+7d | Preview/accept after this ⇒ `expired` |
| `accepted_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `accepted_by_user_id` | `TEXT` | YES | `NULL` | → `users.public_id` |
| `revoked_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `revoked_by_user_id` | `TEXT` | YES | `NULL` | |
| `resent_at` | `TIMESTAMPTZ` | YES | `NULL` | Resend rotates `token_hash` and refreshes `expires_at` |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Pending invitations count toward seat **preview**, not billed seats, until accepted. Public preview (`GET /v1/invitations/{token}`) must **not** return `organization_id`. Resend rotates the hash and does not return the new raw token. Expired rows stay for audit (`status` flipped to `expired` lazily on read or by a sweeper).

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_invitations_organization_id` | `organization_id` | `organizations(public_id)` | CASCADE |
| `fk_invitations_invited_by_user_id` | `invited_by_user_id` | `users(public_id)` | RESTRICT |
| `fk_invitations_accepted_by_user_id` | `accepted_by_user_id` | `users(public_id)` | RESTRICT |
| `fk_invitations_revoked_by_user_id` | `revoked_by_user_id` | `users(public_id)` | RESTRICT |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `invitations_pkey` | PK unique | `public_id` |
| `uniq_invitations_token_hash` | UNIQUE | `token_hash` |
| `uniq_invitations_org_email_pending` | partial unique | `(organization_id, email)` WHERE `status = 'pending'` |
| `idx_invitations_org_email_status` | index | `(organization_id, email, status)` |
| `idx_invitations_org_status_created_at` | index | `(organization_id, status, created_at DESC)` |
| `idx_invitations_invited_by_user_id` | index | `invited_by_user_id` |

#### Sample row

```json
{
  "public_id": "inv_join01c7e2a5b9",
  "organization_id": "org_studio9d2e4c6a80",
  "email": "charles@example.com",
  "role": "sales_designer",
  "status": "pending",
  "token_hash": "sha256:6b1c9e0a4d7f2c8b5a3e1d9f0c7b4a2e8d6c1b0a9f7e5d3c2b1a0f9e8d7c6b5a",
  "invited_by_user_id": "usr_ada01f3e8c2b4a90",
  "expires_at": "2026-09-08T00:00:00Z",
  "accepted_at": null,
  "accepted_by_user_id": null,
  "revoked_at": null,
  "revoked_by_user_id": null,
  "resent_at": null,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/organizations/{id}/invitations`, `DELETE …/invitations/{invitationId}`, `POST …/invitations/{invitationId}/resend`, `POST /v1/invitations/{token}/accept` |
| R | `GET /v1/organizations/{id}/invitations`, `GET /v1/invitations/{token}`, seat preview pending count |

---

### 3.6 `devices`

**Purpose.** An activated Cabinet Studio installation. v1 hard limit: **3 active devices per user**. Reinstallation must not permanently consume a slot — deactivate first. No hardware fingerprints in v1.

**PK.** `public_id TEXT PRIMARY KEY` (`dev_`). User-owned.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `dev_` | API `device.id` |
| `user_id` | `TEXT` | NO | — | → `users.public_id` |
| `name` | `TEXT` | NO | — | 1–80, user-visible |
| `platform` | `TEXT` | NO | — | CHECK `macos` \| `windows` \| `linux` |
| `status` | `TEXT` | NO | `'active'` | CHECK `active` \| `deactivated` |
| `app_version` | `TEXT` | YES | `NULL` | Last seen desktop semver |
| `last_seen_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `deactivated_at` | `TIMESTAMPTZ` | YES | `NULL` | Null if active |
| `deactivated_reason` | `TEXT` | YES | `NULL` | CHECK `user` \| `replaced` \| `admin` \| `account_disabled` |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Activation is serialized with a transaction counting `status='active'` for the user. `409 device_limit_reached` when count ≥ 3.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_devices_user_id` | `user_id` | `users(public_id)` | CASCADE |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `devices_pkey` | PK unique | `public_id` |
| `idx_devices_user_status` | index | `(user_id, status)` |
| `idx_devices_user_last_seen` | index | `(user_id, last_seen_at DESC)` |

#### Sample row

```json
{
  "public_id": "dev_desk01a9c4e7f2",
  "user_id": "usr_ada01f3e8c2b4a90",
  "name": "Studio iMac",
  "platform": "macos",
  "status": "active",
  "app_version": "1.2.0",
  "last_seen_at": "2026-09-01T00:00:00Z",
  "deactivated_at": null,
  "deactivated_reason": null,
  "created_at": "2026-08-02T11:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| R | `GET /v1/devices`, session list (`device_name` / `device_id`) |
| W | `POST /v1/devices/activate`, `PATCH /v1/devices/{id}`, `DELETE /v1/devices/{id}` |
| Entitlements bind | `GET /v1/organizations/{id}/entitlements?device_id=` |

---

### 3.7 `sessions`

**Purpose.** Rotating refresh-token family. Access JWTs are **not** stored. Reuse of a spent refresh token sets `reuse_detected_at` and revokes the **entire family**.

**PK.** `public_id TEXT PRIMARY KEY` (`ses_`). User-owned.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `ses_` | API session `id` |
| `user_id` | `TEXT` | NO | — | → `users.public_id` |
| `device_id` | `TEXT` | YES | `NULL` | → `devices.public_id` if bound |
| `refresh_token_hash` | `TEXT` | NO | — | Unique. Never raw `rt_…` |
| `family_id` | `TEXT` | NO | generated ULID | Stable across rotations |
| `expires_at` | `TIMESTAMPTZ` | NO | now+30d | Refresh TTL; renewed on rotation |
| `revoked_at` | `TIMESTAMPTZ` | YES | `NULL` | Logout, rotation of this row, family revoke |
| `reuse_detected_at` | `TIMESTAMPTZ` | YES | `NULL` | Set on the **presented spent** row; then revoke family |
| `replaced_by_session_id` | `TEXT` | YES | `NULL` | → `sessions.public_id` of the successor |
| `device_name` | `TEXT` | YES | `NULL` | From login `device_name` |
| `user_agent_summary` | `TEXT` | YES | `NULL` | Truncated product string |
| `ip_hash` | `TEXT` | YES | `NULL` | SHA-256(pepper\|\|ip); never raw IP |
| `ip_country` | `TEXT` | YES | `NULL` | ISO-3166-1 alpha-2 |
| `last_seen_at` | `TIMESTAMPTZ` | NO | `now()` | Updated on refresh |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Rotation: insert new session (same `family_id`), set this row `revoked_at` + `replaced_by_session_id`. Concurrent refresh of the same hash: first commit wins; loser is reuse → family revoke → `401 token_reuse_detected`. Clients must serialize refresh.

Do **not** TTL-drop sessions at `expires_at`: reuse detection needs the spent hash to remain. A sweeper may delete rows with `expires_at < now() - 7 days` AND `revoked_at IS NOT NULL`.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_sessions_user_id` | `user_id` | `users(public_id)` | CASCADE |
| `fk_sessions_device_id` | `device_id` | `devices(public_id)` | SET NULL |
| `fk_sessions_replaced_by_session_id` | `replaced_by_session_id` | `sessions(public_id)` | SET NULL |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `sessions_pkey` | PK unique | `public_id` |
| `uniq_sessions_refresh_token_hash` | UNIQUE | `refresh_token_hash` |
| `idx_sessions_family_id` | index | `family_id` |
| `idx_sessions_user_id` | index | `user_id` |
| `idx_sessions_user_active` | index | `(user_id, revoked_at, expires_at)` |
| `idx_sessions_device_id` | partial index | `device_id` WHERE NOT NULL |

#### Sample row

```json
{
  "public_id": "ses_web01b8d3c6e4",
  "user_id": "usr_ada01f3e8c2b4a90",
  "device_id": "dev_desk01a9c4e7f2",
  "refresh_token_hash": "sha256:9f8e7d6c5b4a39281706f5e4d3c2b1a09e8d7c6b5a49382716f0e9d8c7b6a594",
  "family_id": "fam_01j6n8x4k2m9p7q3",
  "expires_at": "2026-10-01T00:00:00Z",
  "revoked_at": null,
  "reuse_detected_at": null,
  "replaced_by_session_id": null,
  "device_name": "Ada Studio Mac",
  "user_agent_summary": "Cabinet Studio/1.2.0 macOS",
  "ip_hash": "sha256:c0ffeeaa11bb22cc33dd44ee55ff6677889900aabbccddeeff00112233445566",
  "ip_country": "IN",
  "last_seen_at": "2026-09-01T00:00:00Z",
  "created_at": "2026-08-20T08:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`, `POST /v1/auth/logout-all`, `DELETE /v1/me/sessions/{id}`, OAuth callback / device token, `POST /v1/me/password`, `POST /v1/me/email/change/confirm` (revoke others) |
| R | `GET /v1/me/sessions` (never hashes or raw tokens) |

---

### 3.8 `email_tokens`

**Purpose.** Single-use hashed tokens for email verify, password reset, email change, and otp login. Cleanup job deletes after `expires_at` (typical 24h; otp_login TTL is 10 minutes).

**PK.** `public_id TEXT PRIMARY KEY` (`tok_`). User-owned. Not returned.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `tok_` | Not returned |
| `user_id` | `TEXT` | NO | — | → `users.public_id` |
| `purpose` | `TEXT` | NO | — | CHECK `email_verify` \| `password_reset` \| `email_change` \| `otp_login` |
| `token_hash` | `TEXT` | NO | — | Unique. Raw prefixes `ev_` / `pr_` / `ec_` never stored. `otp_login` is SHA-256(`otp_login:` + user id + `:` + 6-digit code) |
| `new_email` | `TEXT` | YES | `NULL` | Required when `purpose=email_change` (CHECK) |
| `expires_at` | `TIMESTAMPTZ` | NO | now+24h | Cleanup job |
| `consumed_at` | `TIMESTAMPTZ` | YES | `NULL` | Single-use |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `request_ip_country` | `TEXT` | YES | `NULL` | Never raw IP |

Issuing a new token of the same purpose invalidates prior unconsumed rows (`consumed_at=now()` or delete). Invalid and expired tokens share `400 invalid_token`.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_email_tokens_user_id` | `user_id` | `users(public_id)` | CASCADE |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `email_tokens_pkey` | PK unique | `public_id` |
| `uniq_email_tokens_token_hash` | UNIQUE | `token_hash` |
| `idx_email_tokens_expires_at` | index | `expires_at` (cleanup job) |
| `idx_email_tokens_user_purpose_created_at` | index | `(user_id, purpose, created_at DESC)` |

#### Sample row

```json
{
  "public_id": "tok_ev01ada4c9e2",
  "user_id": "usr_ada01f3e8c2b4a90",
  "purpose": "email_verify",
  "token_hash": "sha256:1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f809",
  "new_email": null,
  "expires_at": "2026-09-02T00:00:00Z",
  "consumed_at": null,
  "created_at": "2026-09-01T00:00:00Z",
  "request_ip_country": "IN"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/auth/register`, `POST /v1/auth/resend-verification`, `POST /v1/auth/password-reset`, `POST /v1/auth/password-reset/confirm`, `POST /v1/auth/verify-email`, `POST /v1/auth/otp/request`, `POST /v1/auth/otp/resend`, `POST /v1/auth/otp/verify`, `POST /v1/me/email/change`, `POST /v1/me/email/change/confirm` |
| R | consume paths above (by `token_hash`) |

---

### 3.9 `oauth_states`

**Purpose.** Short-lived CSRF/PKCE state for authorization-code start → callback. Cleanup after 10 minutes.

**PK.** `public_id TEXT PRIMARY KEY` (`st_`). `state` equals `public_id` in v1.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `st_` | Returned as `state` |
| `state` | `TEXT` | NO | = `public_id` | Unique CSRF value |
| `provider` | `TEXT` | NO | — | CHECK `google` \| `facebook` \| `x` |
| `intent` | `TEXT` | NO | — | CHECK `login` \| `link` |
| `redirect_uri` | `TEXT` | NO | — | Allowlisted desktop/https |
| `code_challenge` | `TEXT` | YES | `NULL` | S256 challenge if PKCE |
| `code_verifier_hash` | `TEXT` | YES | `NULL` | Hash of verifier; never raw verifier |
| `user_id` | `TEXT` | YES | `NULL` | Set when `intent=link` |
| `consumed_at` | `TIMESTAMPTZ` | YES | `NULL` | Single-use; replay → `invalid_state` |
| `expires_at` | `TIMESTAMPTZ` | NO | now+10m | Cleanup job |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

v1 stores `state = public_id`. Keep both unique so the fields may diverge later.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_oauth_states_user_id` | `user_id` | `users(public_id)` | CASCADE |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `oauth_states_pkey` | PK unique | `public_id` |
| `uniq_oauth_states_state` | UNIQUE | `state` |
| `idx_oauth_states_expires_at` | index | `expires_at` (cleanup job) |

#### Sample row

```json
{
  "public_id": "st_01j6n8x4k2m9p7q3r5t8",
  "state": "st_01j6n8x4k2m9p7q3r5t8",
  "provider": "google",
  "intent": "login",
  "redirect_uri": "cabinetstudio://oauth/callback",
  "code_challenge": "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
  "code_verifier_hash": "sha256:aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899",
  "user_id": null,
  "consumed_at": null,
  "expires_at": "2026-09-01T00:10:00Z",
  "created_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/auth/oauth/{provider}/start`, `GET /v1/auth/oauth/{provider}/callback` (consume), `POST /v1/me/identities/{provider}/link` (consume) |
| R | callback / link (lookup by `state`) |

---

### 3.10 `device_codes`

**Purpose.** RFC 8628-style device authorization for the desktop client. `device_code` is polled; `user_code` is typed by the human. Cleanup after 600s.

**PK.** `public_id TEXT PRIMARY KEY` (`dc_`). `device_code` equals `public_id` in v1.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `dc_` | Same as `device_code` |
| `device_code` | `TEXT` | NO | = `public_id` | Unique; never displayed in UI |
| `user_code` | `TEXT` | NO | generated | Unique; stored uppercase `XXXX-XXXX` |
| `client_id` | `TEXT` | NO | `'cabinet-studio-desktop'` | v1 desktop client |
| `provider` | `TEXT` | NO | — | v1 device flow: `google` |
| `status` | `TEXT` | NO | `'pending'` | CHECK `pending` \| `authorized` \| `denied` \| `consumed` \| `expired` |
| `user_id` | `TEXT` | YES | `NULL` | Set on authorize |
| `interval_seconds` | `INTEGER` | NO | `5` | Slow-down if polled faster |
| `last_poll_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `authorized_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `consumed_at` | `TIMESTAMPTZ` | YES | `NULL` | Token endpoint issues session once |
| `expires_at` | `TIMESTAMPTZ` | NO | now+600s | Cleanup job |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Poll faster than `interval_seconds` → `403 slow_down`. Pending → HTTP 428 `authorization_pending`. 200 is single-issue.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_device_codes_user_id` | `user_id` | `users(public_id)` | SET NULL |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `device_codes_pkey` | PK unique | `public_id` |
| `uniq_device_codes_device_code` | UNIQUE | `device_code` |
| `uniq_device_codes_user_code` | UNIQUE | `user_code` |
| `idx_device_codes_expires_at` | index | `expires_at` (cleanup job) |

#### Sample row

```json
{
  "public_id": "dc_01j6n8x4k2m9p7q3r5t8",
  "device_code": "dc_01j6n8x4k2m9p7q3r5t8",
  "user_code": "WDJB-MJHT",
  "client_id": "cabinet-studio-desktop",
  "provider": "google",
  "status": "pending",
  "user_id": null,
  "interval_seconds": 5,
  "last_poll_at": null,
  "authorized_at": null,
  "consumed_at": null,
  "expires_at": "2026-09-01T00:10:00Z",
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/auth/device/code`, `POST /v1/auth/device/token` (poll/consume). Browser verification/authorize UI is outside the 64-route desktop contract. |
| R | device token poll |

---

### 3.11 `plans`

**Purpose.** Server-side commercial catalog. Desktop never hard-codes prices (`BIL-004`). Seeded, rarely mutated.

**PK.** `public_id TEXT PRIMARY KEY` (`plan_`). Global, not tenant-scoped.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `plan_` | API `plan.id` |
| `code` | `TEXT` | NO | — | CHECK `trial` \| `studio` \| `team` \| `enterprise` |
| `name` | `TEXT` | NO | — | Display name |
| `interval` | `TEXT` | NO | — | CHECK `none` \| `month` \| `year` (`none` = trial) |
| `currency` | `TEXT` | NO | `'usd'` | ISO 4217 lowercase |
| `unit_amount_minor` | `INTEGER` | NO | `0` | Cents; 0 for trial |
| `seats_included` | `INTEGER` | NO | — | API `included_seats` |
| `max_devices_per_user` | `INTEGER` | NO | `3` | Always 3 in v1 |
| `trial_days` | `INTEGER` | NO | `14` | |
| `sellable` | `BOOLEAN` | NO | `TRUE` | `false` for trial (omitted from catalog GET) |
| `active` | `BOOLEAN` | NO | `TRUE` | Hide retired SKUs |
| `capabilities` | `JSONB` | NO | — | Product keys + numeric limits |
| `provider` | `TEXT` | YES | `NULL` | CHECK `stripe` \| `fake` |
| `provider_price_id` | `TEXT` | YES | `NULL` | Stripe price id |
| `sort_order` | `INTEGER` | NO | `0` | Catalog order |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

**Capability object keys (all required on every plan):**

| Key | Type | Meaning |
| --- | --- | --- |
| `cloud_projects` | bool | Cloud project sync |
| `organization_assets` | bool | Org 3D/texture catalog |
| `proposal_branding` | bool | Branded proposal outputs |
| `engineering_handoff` | bool | Engineering packet generation |
| `max_seats` | int | Same as `seats_included` (duplicated for snapshot maps) |
| `max_devices` | int | Always 3 |
| `max_cloud_projects` | int | Soft quota; never destructive (`BIL-006`) |

See §5 for full seed rows.

**Trial capabilities:** product booleans are **false/0 except** the documented evaluation set: `cloud_projects=true`, `max_cloud_projects=10`, `max_seats=3`, `max_devices=3`. `organization_assets`, `proposal_branding`, and `engineering_handoff` stay **false** on the production seed. HTTP samples in `API_CONTRACTS_V1.md` that show all four product keys true on a trial snapshot are a **full-evaluation fixture**; contract tests may override the seed, production must not.

#### Foreign keys

None outbound. Referenced by `subscriptions.plan_id` and `entitlements.plan_id`.

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `plans_pkey` | PK unique | `public_id` |
| `uniq_plans_code_interval` | UNIQUE | `(code, interval)` |
| `idx_plans_catalog` | index | `(sellable, active, sort_order)` |

Catalog is tiny; no partitioning.

#### Sample row

See seed `plan_studio_monthly` in §5.

#### API routes

| Op | Routes |
| --- | --- |
| R | `GET /v1/billing/plans` (`sellable=true AND active=true`), `GET /v1/organizations/{id}/subscription` (join), checkout, preview-downgrade, entitlement evaluation |
| W | Seed / admin only (no public write route in the 64) |

---

### 3.12 `subscriptions`

**Purpose.** Normalized commercial agreement for **one organization** (unique `organization_id`). Provider is Stripe or the local `fake` adapter. **No card data.** Checkout completion does not grant access — this row, updated from verified webhooks, does (`BIL-002`).

**PK.** `public_id TEXT PRIMARY KEY` (`sub_`).

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `sub_` | API `subscription.id` |
| `organization_id` | `TEXT` | NO | — | Unique FK → `organizations.public_id` |
| `plan_id` | `TEXT` | NO | — | → `plans.public_id` |
| `plan_code` | `TEXT` | NO | — | Denormalized `trial` \| `studio` \| `team` \| `enterprise` |
| `status` | `TEXT` | NO | `'trialing'` | CHECK `trialing` \| `active` \| `past_due` \| `grace` \| `suspended` \| `canceled` |
| `provider` | `TEXT` | NO | `'fake'` | CHECK `stripe` \| `fake` |
| `provider_subscription_id` | `TEXT` | YES | `NULL` | Partial unique. Null during local-only trial |
| `seats_included` | `INTEGER` | NO | from plan | Purchased/included seats |
| `trial_ends_at` | `TIMESTAMPTZ` | YES | org trial end | |
| `current_period_start` | `TIMESTAMPTZ` | NO | now | UTC |
| `current_period_end` | `TIMESTAMPTZ` | NO | trial end | UTC |
| `cancel_at_period_end` | `BOOLEAN` | NO | `FALSE` | v1 cancel API requires true |
| `canceled_at` | `TIMESTAMPTZ` | YES | `NULL` | When cancel was requested |
| `grace_until` | `TIMESTAMPTZ` | YES | `NULL` | `past_due` + 7 days policy |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Stripe `unpaid` / `incomplete` / `incomplete_expired` are **normalized** into this six-value set (`past_due` / `suspended` / `canceled`) before write. Never persist raw provider enums as `status`.

Org create writes a `trialing` row with `provider='fake'` (or `stripe` later) and **no** `provider_subscription_id`. Paid conversion fills the provider id from a verified webhook, not from checkout return (`BIL-002`).

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_subscriptions_organization_id` | `organization_id` | `organizations(public_id)` | RESTRICT (UNIQUE) |
| `fk_subscriptions_plan_id` | `plan_id` | `plans(public_id)` | RESTRICT |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `subscriptions_pkey` | PK unique | `public_id` |
| `uniq_subscriptions_organization_id` | UNIQUE | `organization_id` |
| `uniq_subscriptions_provider_subscription_id` | partial unique | `provider_subscription_id` WHERE NOT NULL |
| `idx_subscriptions_status_period_end` | index | `(status, current_period_end)` |
| `idx_subscriptions_plan_id` | index | `plan_id` |

#### Sample row

```json
{
  "public_id": "sub_trial01e5a2d7",
  "organization_id": "org_studio9d2e4c6a80",
  "plan_id": "plan_trial",
  "plan_code": "trial",
  "status": "trialing",
  "provider": "fake",
  "provider_subscription_id": null,
  "seats_included": 3,
  "trial_ends_at": "2026-09-15T00:00:00Z",
  "current_period_start": "2026-09-01T00:00:00Z",
  "current_period_end": "2026-09-15T00:00:00Z",
  "cancel_at_period_end": false,
  "canceled_at": null,
  "grace_until": null,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| R | `GET /v1/organizations/{id}/subscription`, entitlements, seats, capabilities, org create response |
| W | `POST /v1/organizations` (trial), `POST …/checkout-session` (does **not** flip status), `POST …/subscription/cancel`, `POST …/subscription/resume`, `POST /v1/webhooks/billing` |

---

### 3.13 `invoices`

**Purpose.** Local projection of provider invoices for the org billing UI. Provider remains system of record for tax/PDF; we store the summary the desktop needs. No PAN/CVC.

**PK.** `public_id TEXT PRIMARY KEY` (`invc_`).

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `invc_` | API `invoice.id` |
| `organization_id` | `TEXT` | NO | — | → `organizations.public_id` |
| `subscription_id` | `TEXT` | YES | `NULL` | → `subscriptions.public_id` |
| `provider` | `TEXT` | NO | — | CHECK `stripe` \| `fake` |
| `provider_invoice_id` | `TEXT` | NO | — | Unique. Stripe `in_…` |
| `number` | `TEXT` | YES | `NULL` | Human invoice number |
| `status` | `TEXT` | NO | — | CHECK `draft` \| `open` \| `paid` \| `void` \| `uncollectible` |
| `currency` | `TEXT` | NO | `'usd'` | |
| `total_minor` | `INTEGER` | NO | `0` | |
| `amount_due_minor` | `INTEGER` | NO | `0` | |
| `hosted_invoice_url` | `TEXT` | YES | `NULL` | Provider hosted page |
| `invoice_pdf_url` | `TEXT` | YES | `NULL` | Provider PDF |
| `period_start` | `TIMESTAMPTZ` | YES | `NULL` | |
| `period_end` | `TIMESTAMPTZ` | YES | `NULL` | |
| `paid_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | Provider created, stored UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Never store the raw Stripe invoice object.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_invoices_organization_id` | `organization_id` | `organizations(public_id)` | RESTRICT |
| `fk_invoices_subscription_id` | `subscription_id` | `subscriptions(public_id)` | SET NULL |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `invoices_pkey` | PK unique | `public_id` |
| `uniq_invoices_provider_invoice_id` | UNIQUE | `provider_invoice_id` |
| `idx_invoices_org_created_at` | index | `(organization_id, created_at DESC)` |
| `idx_invoices_subscription_id` | partial index | `subscription_id` WHERE NOT NULL |

#### Sample row

```json
{
  "public_id": "invc_20260901a3b8",
  "organization_id": "org_studio9d2e4c6a80",
  "subscription_id": "sub_stripe01e5a2d7",
  "provider": "stripe",
  "provider_invoice_id": "in_opaque_stripe",
  "number": "INV-2026-00041",
  "status": "paid",
  "currency": "usd",
  "total_minor": 4900,
  "amount_due_minor": 0,
  "hosted_invoice_url": "https://billing.stripe.example/invoice/invc_20260901a3b8",
  "invoice_pdf_url": "https://billing.stripe.example/invoice/invc_20260901a3b8/pdf",
  "period_start": "2026-08-01T00:00:00Z",
  "period_end": "2026-09-01T00:00:00Z",
  "paid_at": "2026-08-01T00:02:00Z",
  "created_at": "2026-08-01T00:01:00Z",
  "updated_at": "2026-08-01T00:02:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| R | `GET /v1/organizations/{id}/invoices`, `GET /v1/organizations/{id}/invoices/{invoiceId}` |
| W | `POST /v1/webhooks/billing` (`invoice.paid`, `invoice.finalized`, …) |

---

### 3.14 `billing_events`

**Purpose.** Provider webhook inbox and dedupe log (`BIL-003`). Unique `(provider, provider_event_id)` makes replays 200 no-ops. Payload is **redacted / digested**, never a raw card-bearing object.

**PK.** `public_id TEXT PRIMARY KEY` (`bev_`). Internal.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `bev_` | Internal |
| `provider` | `TEXT` | NO | — | CHECK `stripe` \| `fake` |
| `provider_event_id` | `TEXT` | NO | — | Stripe `evt_…` |
| `type` | `TEXT` | NO | — | Provider type, e.g. `invoice.paid` |
| `organization_id` | `TEXT` | YES | `NULL` | Filled when mapping is known |
| `subscription_id` | `TEXT` | YES | `NULL` | |
| `status` | `TEXT` | NO | `'received'` | CHECK `received` \| `processed` \| `ignored` \| `failed` |
| `payload_digest` | `TEXT` | NO | — | SHA-256 of **raw** body bytes |
| `summary` | `JSONB` | NO | `'{}'` | Redacted ids/amounts only |
| `error` | `TEXT` | YES | `NULL` | Safe error code if `failed` |
| `processed_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

HMAC is verified on the raw body **before** insert. Duplicate key on `(provider, provider_event_id)` → return `{received:true}` without re-applying side effects.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_billing_events_organization_id` | `organization_id` | `organizations(public_id)` | SET NULL (nullable) |
| `fk_billing_events_subscription_id` | `subscription_id` | `subscriptions(public_id)` | SET NULL |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `billing_events_pkey` | PK unique | `public_id` |
| `uniq_billing_events_provider_event` | UNIQUE | `(provider, provider_event_id)` |
| `idx_billing_events_org_created_at` | index | `(organization_id, created_at DESC)` |
| `idx_billing_events_status_created_at` | index | `(status, created_at)` |

#### Sample row

```json
{
  "public_id": "bev_01j6n8x4k2m9p7q3",
  "provider": "stripe",
  "provider_event_id": "evt_stripe_opaque",
  "type": "invoice.paid",
  "organization_id": "org_studio9d2e4c6a80",
  "subscription_id": "sub_stripe01e5a2d7",
  "status": "processed",
  "payload_digest": "sha256:aa11bb22cc33dd44ee55ff6677889900aa11bb22cc33dd44ee55ff6677889900",
  "summary": {
    "provider_invoice_id": "in_opaque_stripe",
    "amount_paid_minor": 4900,
    "currency": "usd"
  },
  "error": null,
  "processed_at": "2026-09-01T00:02:01Z",
  "created_at": "2026-09-01T00:02:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W/R | `POST /v1/webhooks/billing` only |

---

### 3.15 `entitlements`

**Purpose.** Issued, signed capability snapshots for a user (+ org + optional device). Desktop verifies `signature` and caches the latest valid snapshot (`LIC-001`). Rows are immutable after insert (issue a new row to refresh).

**PK.** `public_id TEXT PRIMARY KEY` (`ent_`). `organization_id` nullable when the user has no org (`plan=none`). `public_id` is not in the HTTP body.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `ent_` | Not in HTTP body |
| `user_id` | `TEXT` | NO | — | → `users.public_id` |
| `organization_id` | `TEXT` | YES | `NULL` | → `organizations.public_id` |
| `device_id` | `TEXT` | YES | `NULL` | Bound when issued at activation |
| `plan_id` | `TEXT` | YES | `NULL` | → `plans.public_id` |
| `plan` | `TEXT` | NO | `'none'` | CHECK `none` \| `trial` \| `studio` \| `team` \| `enterprise` |
| `subscription_status` | `TEXT` | NO | `'none'` | `none` plus subscription statuses |
| `capabilities` | `JSONB` | NO | all false/0 | Role ∩ plan ∩ subscription. Mixed bool/int |
| `issued_at` | `TIMESTAMPTZ` | NO | `now()` | |
| `refresh_after` | `TIMESTAMPTZ` | NO | now+12h | Client should refresh online |
| `expires_at` | `TIMESTAMPTZ` | NO | now+24h | Hard snapshot expiry |
| `grace_until` | `TIMESTAMPTZ` | YES | `NULL` | 7-day paid-action / offline grace |
| `schema_version` | `INTEGER` | NO | `1` | |
| `key_id` | `TEXT` | NO | — | Signing key identifier |
| `signature` | `TEXT` | NO | — | Base64 over canonical snapshot **excluding** `signature` |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | = `issued_at` |

Canonical signed payload fields: `user_id`, `organization_id`, `plan`, `subscription_status`, `capabilities`, `issued_at`, `refresh_after`, `expires_at`, `grace_until`, `device_id`, `schema_version`. Do not embed mutable plan logic in JWTs — this table is the license document.

`GET /v1/me/capabilities` is **unsigned** UI and may be computed live without writing a row. Device activation and `GET …/entitlements` persist a row.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_entitlements_user_id` | `user_id` | `users(public_id)` | CASCADE |
| `fk_entitlements_organization_id` | `organization_id` | `organizations(public_id)` | SET NULL |
| `fk_entitlements_device_id` | `device_id` | `devices(public_id)` | SET NULL |
| `fk_entitlements_plan_id` | `plan_id` | `plans(public_id)` | SET NULL |

`device_id`, if set, must be owned by the user and `active` (application check).

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `entitlements_pkey` | PK unique | `public_id` |
| `idx_entitlements_org_user_issued_at` | index | `(organization_id, user_id, issued_at DESC)` |
| `idx_entitlements_user_issued_at` | index | `(user_id, issued_at DESC)` |
| `idx_entitlements_device_issued_at` | partial index | `(device_id, issued_at DESC)` WHERE `device_id IS NOT NULL` |

#### Sample row

```json
{
  "public_id": "ent_01j6n8x4k2m9p7q3",
  "user_id": "usr_ada01f3e8c2b4a90",
  "organization_id": "org_studio9d2e4c6a80",
  "device_id": "dev_desk01a9c4e7f2",
  "plan_id": "plan_trial",
  "plan": "trial",
  "subscription_status": "trialing",
  "capabilities": {
    "cloud_projects": true,
    "organization_assets": false,
    "proposal_branding": false,
    "engineering_handoff": false,
    "max_seats": 3,
    "max_devices": 3,
    "max_cloud_projects": 10
  },
  "issued_at": "2026-09-01T00:00:00Z",
  "refresh_after": "2026-09-01T12:00:00Z",
  "expires_at": "2026-09-02T00:00:00Z",
  "grace_until": "2026-09-08T00:00:00Z",
  "schema_version": 1,
  "key_id": "ent-ed25519-2026-09",
  "signature": "c2lnbmVkLWVudGl0bGVtZW50LXYyLXRyaWFs",
  "created_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/devices/activate`, `GET /v1/organizations/{id}/entitlements` (issue-on-read is allowed) |
| R | those two (return snapshot). `GET /v1/me/capabilities` is computed, may skip persist |

---

### 3.16 `audit_events`

**Purpose.** Append-only security and commercial audit trail. **No updates. No deletes** (except legal hold / retention jobs outside v1 app code). Metadata is redacted: never credentials, tokens, raw IPs, or provider payloads (`AUT-006`, `BIL-005`).

**PK.** `public_id TEXT PRIMARY KEY` (`evt_`). v1 **requires** `organization_id` on rows returned by `GET /v1/organizations/{id}/audit-events`. Login events are written per org the user belongs to, or omitted from org audit if the user has none. Column `actor_user_id` is the API field `actor_id`.

DDL revokes `UPDATE`/`DELETE` on this table from `PUBLIC`. App role should be insert-only.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `evt_` | API `id` |
| `organization_id` | `TEXT` | YES | `NULL` | → `organizations.public_id`; required on org audit API rows |
| `actor_user_id` | `TEXT` | YES | `NULL` | API `actor_id`; `usr_` or null |
| `actor_type` | `TEXT` | NO | `'user'` | CHECK `user` \| `system` \| `webhook` |
| `action` | `TEXT` | NO | — | Stable key, e.g. `member.role_changed` |
| `resource_type` | `TEXT` | NO | — | CHECK organization \| membership \| invitation \| subscription \| device \| session \| user \| invoice \| entitlement |
| `resource_id` | `TEXT` | YES | `NULL` | Opaque public id |
| `ip_country` | `TEXT` | YES | `NULL` | Never raw IP |
| `request_id` | `TEXT` | YES | `NULL` | `req_…` |
| `metadata` | `JSONB` | NO | `'{}'` | Small redacted JSON |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC. No `updated_at` |

Typical `action` values: `auth.login`, `org.created`, `org.deleted`, `org.ownership_transferred`, `member.role_changed`, `member.suspended`, `member.removed`, `invitation.created`, `invitation.accepted`, `billing.trial_started`, `billing.subscription_cancel_requested`, `billing.subscription_updated`, `device.activated`, `device.deactivated`.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_audit_events_organization_id` | `organization_id` | `organizations(public_id)` | SET NULL (nullable) |
| `fk_audit_events_actor_user_id` | `actor_user_id` | `users(public_id)` | SET NULL (nullable) |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `audit_events_pkey` | PK unique | `public_id` |
| `idx_audit_events_org_created_at` | index | `(organization_id, created_at DESC)` |
| `idx_audit_events_action` | index | `action` |
| `idx_audit_events_org_action_created_at` | index | `(organization_id, action, created_at DESC)` |
| `idx_audit_events_org_actor_created_at` | index | `(organization_id, actor_user_id, created_at DESC)` |

#### Sample row

```json
{
  "public_id": "evt_audit01d6f1c8",
  "organization_id": "org_studio9d2e4c6a80",
  "actor_user_id": "usr_ada01f3e8c2b4a90",
  "actor_type": "user",
  "action": "member.role_changed",
  "resource_type": "membership",
  "resource_id": "usr_bb02e4d9c1a7f380",
  "ip_country": "IN",
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "metadata": {
    "from_role": "viewer",
    "to_role": "engineer"
  },
  "created_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

| Op | Routes |
| --- | --- |
| R | `GET /v1/organizations/{id}/audit-events` |
| W | Most mutating routes via the same transaction as the domain write (insert + outbox). Not a client-writable resource |

---

### 3.17 `outbox`

**Purpose.** Transactional outbox. HTTP/webhook handlers insert a row in the **same PostgreSQL transaction** as domain writes. A publisher worker sends to Kafka (or equivalent) and sets `published_at` after ack. Kafka is not the source of truth.

**PK.** `public_id TEXT PRIMARY KEY` (`obx_`). `event_id` is the publish idempotency key (often equal to `public_id`).

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `obx_` | Internal |
| `event_id` | `TEXT` | NO | = `public_id` | Unique. Kafka key / idempotent publish |
| `topic` | `TEXT` | NO | — | e.g. `cabinet.org.v1` |
| `partition_key` | `TEXT` | NO | org or user id | Usually `organization_id` string; **no FK** |
| `payload` | `BYTEA` | NO | — | Marshaled `cabinet.v1.Envelope` protobuf (public ids, no secrets). Kafka value is these raw bytes — never JSON |
| `attempts` | `INTEGER` | NO | `0` | Publish attempts |
| `last_error` | `TEXT` | YES | `NULL` | Truncated safe error |
| `available_at` | `TIMESTAMPTZ` | NO | `now()` | Backoff; publisher queries `available_at <= now()` |
| `published_at` | `TIMESTAMPTZ` | YES | `NULL` | Null until Kafka ack |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |

Unpublished query: `WHERE published_at IS NULL` served by partial index `idx_outbox_unpublished`.

#### Foreign keys

None. `partition_key` is typically `organizations.public_id` or `users.public_id`. Not existence-checked (the org may be deleted after enqueue).

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `outbox_pkey` | PK unique | `public_id` |
| `uniq_outbox_event_id` | UNIQUE | `event_id` |
| `idx_outbox_unpublished` | partial index | `created_at` WHERE `published_at IS NULL` |
| `idx_outbox_publisher_due` | index | `(available_at, published_at)` |

#### Sample row

```json
{
  "public_id": "obx_01j6n8x4k2m9p7q3",
  "event_id": "obx_01j6n8x4k2m9p7q3",
  "topic": "cabinet.org.events",
  "partition_key": "org_studio9d2e4c6a80",
  "payload": "<BYTEA marshaled cabinet.v1.Envelope protobuf; Kafka value is these raw bytes, never JSON>",
  "attempts": 0,
  "last_error": null,
  "available_at": "2026-09-01T00:00:00Z",
  "published_at": null,
  "created_at": "2026-09-01T00:00:00Z"
}
```

#### API routes

No direct HTTP. Written by mutating handlers (org create, invite accept, webhook, login, export, …). Read by the outbox publisher worker. `GET /readyz` may check publisher lag, not this table's contents.

---

### 3.18 `idempotency_keys`

**Purpose.** 24-hour replay cache for retryable mutations. Same principal + key + body hash returns the stored response. Same key + different body → `409 idempotency_conflict`.

**PK.** `public_id TEXT PRIMARY KEY` (`idk_`). User-scoped when authenticated; else `ip:<hash>:<route>`. Cleanup job deletes rows older than 24h.

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `idk_` | Internal |
| `principal` | `TEXT` | NO | — | `usr_…` or `ip:<sha256>` or `wh:stripe` |
| `key` | `TEXT` | NO | — | Client `Idempotency-Key` UUID |
| `method` | `TEXT` | NO | — | HTTP method |
| `path` | `TEXT` | NO | — | Normalized route path |
| `request_hash` | `TEXT` | NO | — | SHA-256 of canonical body |
| `response_status` | `INTEGER` | NO | — | Stored HTTP status |
| `response_body` | `JSONB` | YES | `NULL` | Stored JSON; never persist raw refresh tokens |
| `in_progress` | `BOOLEAN` | NO | `TRUE` | Lock while the original request runs |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | Cleanup after 24h |
| `completed_at` | `TIMESTAMPTZ` | YES | `NULL` | |

For login/refresh, **do not** persist raw `refresh_token` inside `response_body`. Those routes are not idempotent in the contract.

Public auth routes that require Idempotency-Key (`POST /v1/auth/register`) use `principal = 'ip:' || sha256(ip)`.

#### Foreign keys

None. `principal` is `users.public_id` only when prefixed `usr_`.

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `idempotency_keys_pkey` | PK unique | `public_id` |
| `uniq_idempotency_keys_principal_key` | UNIQUE | `(principal, key)` |
| `idx_idempotency_keys_created_at` | index | `created_at` (cleanup job) |

#### Sample row

```json
{
  "public_id": "idk_01j6n8x4k2m9p7q3",
  "principal": "usr_ada01f3e8c2b4a90",
  "key": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/v1/organizations",
  "request_hash": "sha256:bb11cc22dd33ee44ff5500112233445566778899aabbccddeeff001122334455",
  "response_status": 201,
  "response_body": {
    "request_id": "req_01j6n8x4k2m9p7q3r5t8",
    "organization": { "id": "org_studio9d2e4c6a80" }
  },
  "in_progress": false,
  "created_at": "2026-09-01T00:00:00Z",
  "completed_at": "2026-09-01T00:00:01Z"
}
```

#### API routes

Read/write middleware on every route whose contract marks `Idempotency-Key` required or optional. Not a resource of its own.

---

### 3.19 `data_exports`

**Purpose.** Asynchronous personal-data export jobs (`POST /v1/me/export`). Contains profile, memberships, session **metadata** — not org-owned projects. Archive bytes live in object storage; Postgres stores job state. User-owned.

**PK.** `public_id TEXT PRIMARY KEY` (`exp_`).

#### Columns

| Name | Type | Null | Default | Notes |
| --- | --- | --- | --- | --- |
| `public_id` | `TEXT` | NO | `exp_` | API `export_id` |
| `user_id` | `TEXT` | NO | — | → `users.public_id` |
| `status` | `TEXT` | NO | `'queued'` | CHECK `queued` \| `processing` \| `ready` \| `failed` |
| `scope` | `TEXT` | NO | `'personal'` | v1 only personal |
| `object_key` | `TEXT` | YES | `NULL` | Object-storage key; not a durable public URL |
| `bytes` | `BIGINT` | YES | `NULL` | Archive size |
| `error_code` | `TEXT` | YES | `NULL` | Safe code if failed |
| `created_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `updated_at` | `TIMESTAMPTZ` | NO | `now()` | UTC |
| `completed_at` | `TIMESTAMPTZ` | YES | `NULL` | |
| `expires_at` | `TIMESTAMPTZ` | YES | completed+7d | Cleanup after ready; download route is LATER |

`409 export_in_progress` if another row for this user is `queued` or `processing` (enforced by partial unique `uniq_data_exports_user_in_progress`). Job is inserted with outbox `user.export_requested` after commit.

#### Foreign keys

| Constraint | Field | References | ON DELETE |
| --- | --- | --- | --- |
| `fk_data_exports_user_id` | `user_id` | `users(public_id)` | CASCADE |

#### Unique / indexes

| Name | Kind | Keys |
| --- | --- | --- |
| `data_exports_pkey` | PK unique | `public_id` |
| `uniq_data_exports_user_in_progress` | partial unique | `user_id` WHERE `status IN ('queued','processing')` |
| `idx_data_exports_user_status_created_at` | index | `(user_id, status, created_at DESC)` |
| `idx_data_exports_expires_at` | partial index | `expires_at` WHERE NOT NULL (cleanup job) |

#### Sample row

```json
{
  "public_id": "exp_me01f4c9d2e6",
  "user_id": "usr_ada01f3e8c2b4a90",
  "status": "queued",
  "scope": "personal",
  "object_key": null,
  "bytes": null,
  "error_code": null,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z",
  "completed_at": null,
  "expires_at": null
}
```

#### API routes

| Op | Routes |
| --- | --- |
| W | `POST /v1/me/export` (insert `queued` + outbox) |
| R | v1 has no GET-export route; worker updates status. Future download is out of the 64 |

---

## 4. Transaction boundaries

PostgreSQL. Use a single `BEGIN … COMMIT` (pgx `Tx`) for the following. Single-row updates (e.g. `PATCH /v1/me` display_name) do **not** need a multi-statement transaction beyond autocommit.

Isolation: `REPEATABLE READ` or `SERIALIZABLE` for refresh-family rotation and last-owner guards. Default `READ COMMITTED` is acceptable for most other writes if seat counts use `SELECT … FOR UPDATE`. Outbox publisher runs **after** commit; at-least-once Kafka delivery is expected (consumers must be idempotent on `event_id`).

| Name | Writes in one txn | Trigger |
| --- | --- | --- |
| **Register** | `users` insert + `email_tokens` insert + `idempotency_keys` upsert + optional `outbox` `user.created` | `POST /v1/auth/register` |
| **Register + org + trial** | Today this is **org-create**, not register. `organizations` insert + `memberships` owner insert + `subscriptions` trialing insert + `audit_events` (`org.created`, `membership.created`, `billing.trial_started`) + `outbox` (same three topics) + `idempotency_keys`. Do **not** call Kafka in-handler. Creator must already exist and be email-verified. Insert **organization first, then membership** (deferred `owner_user_id` FK). If a future combined register+org handler is added, add `users` + `email_tokens` to this same txn. | `POST /v1/organizations` |
| **Accept invite** | `invitations` → `accepted` + `memberships` upsert `active` + `audit_events` `invitation.accepted` / `membership.created` + `outbox` + `idempotency_keys`. Seat check is a transactional read of memberships + subscription (`SELECT … FOR UPDATE`). | `POST /v1/invitations/{token}/accept` |
| **Transfer ownership** | `organizations.owner_user_id` + old membership `role=admin` + new membership `role=owner` + `audit_events` `org.ownership_transferred` + `outbox`. Recent-auth check is outside the write set. | `POST /v1/organizations/{id}/transfer-ownership` |
| **Refresh rotation** | Insert new `sessions` row (same `family_id`) + update presented row (`revoked_at`, `replaced_by_session_id`). **Reuse path:** set `reuse_detected_at` on presented row + `revoked_at=now()` on **all** family members + `audit_events` + `outbox` `auth.refresh_reuse`. Must be serializable vs concurrent refresh. | `POST /v1/auth/refresh` |
| **Webhook process + outbox** | Insert `billing_events` (unique provider+event_id; duplicate ⇒ abort empty / return 200) + update `subscriptions` and/or upsert `invoices` + `organizations.provider_customer_id` if first customer + `audit_events` + `outbox`. HMAC verify **before** the txn. | `POST /v1/webhooks/billing` |
| **Login** | `sessions` insert + `users.last_login_at` + `audit_events` `auth.login` + `outbox` | `POST /v1/auth/login` (and oauth/device success) |
| **Device activate** | Count active devices + `devices` insert + `entitlements` insert + `audit_events` + `outbox` | `POST /v1/devices/activate` |
| **Email change confirm** | consume `email_tokens` + update `users.email*` + revoke other `sessions` + `audit_events` + `outbox` | `POST /v1/me/email/change/confirm` |
| **Member remove / leave / suspend** | membership state + `audit_events` + `outbox` (+ last-owner guard reads) | matching org member routes |
| **Export request** | `data_exports` insert + `outbox` `user.export_requested` + idempotency | `POST /v1/me/export` |

---

## 5. Plans seed (full rows)

Insert-once (`ON CONFLICT (public_id) DO NOTHING` in `schema-v1.sql`). `public_id` values are stable. `GET /v1/billing/plans` returns rows with `sellable=true` AND `active=true` (studio + team). Trial is used internally when creating an organization.

Timestamps below are seed-time; implementations may set `created_at` to deploy time.

### 5.1 Trial — `plan_trial` (not sellable)

14-day evaluation. Product keys are **false/0 except** `cloud_projects=true` and the numeric caps below.

```json
{
  "public_id": "plan_trial",
  "code": "trial",
  "name": "Trial",
  "interval": "none",
  "currency": "usd",
  "unit_amount_minor": 0,
  "seats_included": 3,
  "max_devices_per_user": 3,
  "trial_days": 14,
  "sellable": false,
  "active": true,
  "capabilities": {
    "cloud_projects": true,
    "organization_assets": false,
    "proposal_branding": false,
    "engineering_handoff": false,
    "max_seats": 3,
    "max_devices": 3,
    "max_cloud_projects": 10
  },
  "provider": null,
  "provider_price_id": null,
  "sort_order": 0,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

### 5.2 Studio monthly — `plan_studio_monthly`

```json
{
  "public_id": "plan_studio_monthly",
  "code": "studio",
  "name": "Studio",
  "interval": "month",
  "currency": "usd",
  "unit_amount_minor": 2900,
  "seats_included": 1,
  "max_devices_per_user": 3,
  "trial_days": 14,
  "sellable": true,
  "active": true,
  "capabilities": {
    "cloud_projects": true,
    "organization_assets": true,
    "proposal_branding": true,
    "engineering_handoff": false,
    "max_seats": 1,
    "max_devices": 3,
    "max_cloud_projects": 50
  },
  "provider": "stripe",
  "provider_price_id": "price_studio_monthly_placeholder",
  "sort_order": 10,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

### 5.3 Team monthly — `plan_team_monthly`

```json
{
  "public_id": "plan_team_monthly",
  "code": "team",
  "name": "Team",
  "interval": "month",
  "currency": "usd",
  "unit_amount_minor": 4900,
  "seats_included": 5,
  "max_devices_per_user": 3,
  "trial_days": 14,
  "sellable": true,
  "active": true,
  "capabilities": {
    "cloud_projects": true,
    "organization_assets": true,
    "proposal_branding": true,
    "engineering_handoff": true,
    "max_seats": 5,
    "max_devices": 3,
    "max_cloud_projects": 500
  },
  "provider": "stripe",
  "provider_price_id": "price_team_monthly_placeholder",
  "sort_order": 20,
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

Enterprise is sales-led and **not** seeded. Local `fake` adapter may copy these amounts without Stripe price ids (`provider_price_id` null, `provider='fake'`).

---

## 6. Cleanup jobs (replaces Mongo-style TTL)

PostgreSQL has no TTL indexes. A cron / worker (every 60s is plenty) runs the deletes below. **Do not rely on these jobs for security invalidation** — check `expires_at` / `consumed_at` in application code first.

| Table | Predicate | Effect |
| --- | --- | --- |
| `email_tokens` | `expires_at < now()` | Drop after token expiry (24h) |
| `oauth_states` | `expires_at < now()` | Drop after 10 minutes |
| `device_codes` | `expires_at < now()` | Drop after 600s |
| `idempotency_keys` | `created_at < now() - interval '24 hours'` | 24h replay window |
| `data_exports` | `expires_at IS NOT NULL AND expires_at < now()` | Drop ready archives' job rows after retention |
| `sessions` (optional sweeper) | `expires_at < now() - interval '7 days' AND revoked_at IS NOT NULL` | Spent-hash retention for reuse detection; then drop |

Sessions are **not** expired-deleted at `expires_at` so refresh-reuse hashes survive.

Suggested SQL (also in `schema-v1.sql` comments):

```sql
DELETE FROM email_tokens     WHERE expires_at < now();
DELETE FROM oauth_states     WHERE expires_at < now();
DELETE FROM device_codes     WHERE expires_at < now();
DELETE FROM idempotency_keys WHERE created_at < now() - interval '24 hours';
DELETE FROM data_exports     WHERE expires_at IS NOT NULL AND expires_at < now();
DELETE FROM sessions
  WHERE expires_at < now() - interval '7 days' AND revoked_at IS NOT NULL;
```

---

## 7. Index inventory

Counts **include** primary-key unique indexes (Postgres creates one per `PRIMARY KEY`). There is no document-store `_id`. Cleanup-job supporting indexes replace former TTL indexes.

| Table | Indexes | Unique (incl. PK) | Cleanup-supporting |
| --- | --- | --- | --- |
| users | 4 | 2 | 0 |
| oauth_identities | 3 | 3 | 0 |
| organizations | 4 | 2 | 0 |
| memberships | 4 | 2 | 0 |
| invitations | 6 | 3 | 0 |
| devices | 3 | 1 | 0 |
| sessions | 6 | 2 | 0 |
| email_tokens | 4 | 2 | 1 |
| oauth_states | 3 | 2 | 1 |
| device_codes | 4 | 3 | 1 |
| plans | 3 | 2 | 0 |
| subscriptions | 5 | 3 | 0 |
| invoices | 4 | 2 | 0 |
| billing_events | 4 | 2 | 0 |
| entitlements | 4 | 1 | 0 |
| audit_events | 5 | 1 | 0 |
| outbox | 4 | 2 | 0 |
| idempotency_keys | 3 | 2 | 1 |
| data_exports | 4 | 2 | 1 |
| **Total** | **77** | **39** | **5** |

Foreign keys: **28**. Partial uniques added vs the former unique set: `uniq_invitations_org_email_pending`, `uniq_data_exports_user_in_progress`. FK-covering indexes: `idx_invitations_invited_by_user_id`, `idx_sessions_device_id`, `idx_subscriptions_plan_id`, `idx_invoices_subscription_id`.

Required-by-contract indexes (must exist; named above):

- `users`: `uniq_users_email_normalized`, PK `public_id`
- `oauth_identities`: `uniq_oauth_identities_provider_subject` `(provider, provider_user_id)`, `uniq_oauth_identities_user_provider` `(user_id, provider)`
- `memberships`: `uniq_memberships_org_user` `(organization_id, user_id)`
- `invitations`: `uniq_invitations_token_hash`; `uniq_invitations_org_email_pending`; `idx_invitations_org_email_status`
- `sessions`: `uniq_sessions_refresh_token_hash`; `idx_sessions_family_id`; `idx_sessions_user_id`
- `email_tokens`: `uniq_email_tokens_token_hash`; `idx_email_tokens_expires_at`
- `oauth_states`: `uniq_oauth_states_state`; `idx_oauth_states_expires_at`
- `device_codes`: `uniq_device_codes_device_code`, `uniq_device_codes_user_code`; `idx_device_codes_expires_at`
- `subscriptions`: `uniq_subscriptions_organization_id`; `uniq_subscriptions_provider_subscription_id` (partial)
- `billing_events`: `uniq_billing_events_provider_event` `(provider, provider_event_id)`
- `invoices`: `uniq_invoices_provider_invoice_id`; `idx_invoices_org_created_at`
- `entitlements`: `idx_entitlements_org_user_issued_at`
- `audit_events`: `idx_audit_events_org_created_at`; `idx_audit_events_action`
- `outbox`: `uniq_outbox_event_id`; `idx_outbox_unpublished`
- `idempotency_keys`: `uniq_idempotency_keys_principal_key` `(principal, key)`; `idx_idempotency_keys_created_at`
- `devices`: `idx_devices_user_status` `(user_id, status)`

---

## 8. What is NOT stored

| Forbidden | Where it lives instead |
| --- | --- |
| PAN / CVC / raw card data | Stripe (or fake checkout). `BIL-001` |
| Bank account numbers | Billing provider |
| Raw JWT access tokens | Client memory only (15m) |
| Raw JWT / opaque refresh tokens (`rt_…`) | Client OS credential store; server has `refresh_token_hash` only |
| Raw invite tokens | Email + client; server `token_hash` |
| Raw `ev_` / `pr_` / `ec_` email tokens | Email + client; server `token_hash` |
| Provider OAuth access/refresh tokens | Not persisted after the code exchange |
| Kafka offsets / topics as source of truth | `outbox` + domain tables. Kafka is downstream |
| Device hardware fingerprints | Not in v1 request or schema |
| Raw client IP | `ip_hash` + `ip_country` only |
| Password plaintext or reversible encryption | argon2id only |
| Stripe webhook raw bodies | `payload_digest` + redacted `summary` |
| Checkout / portal session secrets | Provider-hosted; API returns URL only, not stored as a table |
| Cloud project documents / GLB bytes | Local-first files + LATER object storage. Not these 19 tables |
| Document-store `_id` ObjectId | Gone. `public_id` is the only row identity |

---

## 9. Rejected (not v1 tables)

| Rejected | Why |
| --- | --- |
| `cloud_projects`, `revisions`, `assets` | Product-book LATER. Local-first; commercial v1 is identity/billing/licensing |
| `payment_methods`, `cards` | Provider-owned; PAN never touches us |
| `checkout_sessions`, `portal_sessions` | Ephemeral provider URLs; not SoT |
| `rate_limits` | Redis / in-process buckets, not Postgres |
| `kafka_offsets` | Worker consumer group; Postgres outbox is SoT |
| `access_tokens` / `jti_blocklist` | Short-lived JWT; family revoke covers refresh. JTI denylist is LATER if needed |
| `notifications`, `email_log` | Outbox + email provider. Full notification service is LATER |
| `roles` table | Static map in API (`GET …/roles`); not data |
| Document store / MongoDB | Replaced by this PostgreSQL schema |
| Citus / distributed SQL | Not required for v1; single primary + later replica |

---

## 10. Route → table map (64)

Ops/public with no durable write: `GET /healthz` (none), `GET /readyz` (connectivity ping, no table contract), `GET /v1/public/config` (config/env, not Postgres — terms versions are code/config, not a 20th table).

| # | Route | Primary tables |
| --- | --- | --- |
| 4 | `POST /v1/auth/register` | users W, email_tokens W, idempotency_keys W, outbox W |
| 5 | `POST /v1/auth/verify-email` | email_tokens W, users W |
| 6 | `POST /v1/auth/resend-verification` | users R, email_tokens W (if unverified) |
| 7 | `POST /v1/auth/login` | users R, sessions W, audit_events W, outbox W |
| 8 | `POST /v1/auth/refresh` | sessions RW |
| 9 | `POST /v1/auth/logout` | sessions W |
| 10 | `POST /v1/auth/logout-all` | sessions W |
| 11 | `POST /v1/auth/password-reset` | users R, email_tokens W |
| 12 | `POST /v1/auth/password-reset/confirm` | email_tokens W, users W, sessions W |
| 13 | `GET /v1/auth/oauth/providers` | none (config) |
| 14 | `POST /v1/auth/oauth/{provider}/start` | oauth_states W |
| 15 | `GET /v1/auth/oauth/{provider}/callback` | oauth_states W, oauth_identities W, users RW, sessions W |
| 16 | `GET /v1/me/identities` | oauth_identities R |
| 17 | `POST /v1/me/identities/{provider}/link` | oauth_states W, oauth_identities W |
| 18 | `DELETE /v1/me/identities/{provider}` | oauth_identities W, users R |
| 19 | `POST /v1/auth/device/code` | device_codes W |
| 20 | `POST /v1/auth/device/token` | device_codes W, users R, oauth_identities R, sessions W |
| 21 | `GET /v1/me` | users R |
| 22 | `PATCH /v1/me` | users W |
| 23 | `POST /v1/me/password` | users W, sessions W |
| 24 | `GET /v1/me/sessions` | sessions R, devices R |
| 25 | `DELETE /v1/me/sessions/{id}` | sessions W |
| 26 | `POST /v1/me/export` | data_exports W, outbox W, idempotency_keys W |
| 27 | `POST /v1/me/email/change` | users RW, email_tokens W |
| 28 | `POST /v1/me/email/change/confirm` | email_tokens W, users W, sessions W |
| 29 | `GET /v1/me/capabilities` | memberships R, subscriptions R, plans R (compute; entitlements optional) |
| 30 | `GET /v1/devices` | devices R |
| 31 | `POST /v1/devices/activate` | devices W, entitlements W, memberships R, subscriptions R, plans R |
| 32 | `PATCH /v1/devices/{id}` | devices W |
| 33 | `DELETE /v1/devices/{id}` | devices W |
| 34 | `POST /v1/organizations` | organizations W, memberships W, subscriptions W, plans R, audit_events W, outbox W |
| 35 | `GET /v1/organizations` | memberships R, organizations R |
| 36 | `GET /v1/organizations/{id}` | organizations R, memberships R |
| 37 | `PATCH /v1/organizations/{id}` | organizations W |
| 38 | `DELETE /v1/organizations/{id}` | organizations W, subscriptions R, audit_events W, outbox W |
| 39 | `POST /v1/organizations/{id}/leave` | memberships W, audit_events W, outbox W |
| 40 | `POST /v1/organizations/{id}/transfer-ownership` | organizations W, memberships W, audit_events W, outbox W |
| 41 | `GET /v1/organizations/{id}/roles` | memberships R (authz only; map is static) |
| 42 | `GET /v1/organizations/{id}/members` | memberships R, users R |
| 43 | `PATCH …/members/{userId}` | memberships W, audit_events W, outbox W |
| 44 | `POST …/members/{userId}/suspend` | memberships W, audit_events W, outbox W |
| 45 | `DELETE …/members/{userId}` | memberships W, audit_events W, outbox W |
| 46 | `POST …/invitations` | invitations W, memberships R, subscriptions R, audit_events W, outbox W |
| 47 | `GET …/invitations` | invitations R |
| 48 | `DELETE …/invitations/{invitationId}` | invitations W |
| 49 | `POST …/invitations/{invitationId}/resend` | invitations W |
| 50 | `GET /v1/invitations/{token}` | invitations R, organizations R, users R |
| 51 | `POST /v1/invitations/{token}/accept` | invitations W, memberships W, subscriptions R, audit_events W, outbox W |
| 52 | `GET …/entitlements` | entitlements RW, memberships R, subscriptions R, plans R, devices R |
| 53 | `GET /v1/billing/plans` | plans R |
| 54 | `GET …/subscription` | subscriptions R, plans R |
| 55 | `POST …/checkout-session` | organizations R, subscriptions R, plans R, users R (no card write) |
| 56 | `POST …/portal-session` | organizations R, subscriptions R |
| 57 | `POST …/subscription/cancel` | subscriptions W, audit_events W, outbox W |
| 58 | `POST …/subscription/resume` | subscriptions W, audit_events W, outbox W |
| 59 | `GET …/invoices` | invoices R |
| 60 | `GET …/invoices/{invoiceId}` | invoices R |
| 61 | `GET …/seats` | subscriptions R, memberships R, invitations R |
| 62 | `POST …/seats/preview-downgrade` | plans R, memberships R, subscriptions R |
| 63 | `POST /v1/webhooks/billing` | billing_events W, subscriptions W, invoices W, organizations W, audit_events W, outbox W |
| 64 | `GET …/audit-events` | audit_events R |

Every org-scoped route also **reads** `memberships` for `API-001`.

---

## 11. Implementation notes (Go + Gin + pgx)

- Apply [`schema-v1.sql`](./schema-v1.sql) at deploy (idempotent seeds via `ON CONFLICT DO NOTHING`). Fail `GET /readyz` if required unique indexes / FKs are missing (`pg_indexes` / `information_schema.table_constraints`).
- Use typed structs with `db:"snake_case"` / sqlc matching this document. Driver: **pgx** (or `database/sql` + pgx stdlib).
- Transactions: `pool.Begin` + `tx.Commit`. Set `SET timezone = 'UTC'` (or `PGTZ=UTC`) on every connection.
- JSONB: scan into `json.RawMessage` / `map[string]any`. Do not store PAN or raw tokens in JSONB.
- `SELECT … FOR UPDATE` on memberships/subscriptions/sessions rows that participate in seat, last-owner, or refresh-family invariants.
- Clock: store UTC `timestamptz` only. Entitlement tests inject a clock (`LIC-004`).
- Backup: Neon PITR / RDS PITR on `cabinet_studio`. `audit_events` and `outbox` are high-value.
- **Scale:** one primary on managed Postgres (Neon or RDS) is enough for v1. Add a **read replica** later for `GET` heavy paths (plans, invoices list, audit). **Citus is not required** for v1.

_End of schema. Tables documented: **19**. Foreign keys: **28**. Indexes: **77** (unique **39**). Baseline: **2026-09-01**. Matches API_CONTRACTS_V1.md **64** routes._
