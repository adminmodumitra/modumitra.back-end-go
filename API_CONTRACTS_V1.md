# Cabinet Studio API Contracts v1

**Document role:** Canonical HTTP contract for the Cabinet Studio commercial backend  
**Audience:** Desktop (Tauri) client, backend (Go + Gin + MongoDB), QA, and contract tests  
**API version:** `v1`  
**Baseline date:** 2026-09-01  
**Status:** Approved contract for implementation; not evidence that routes are shipped  

**Documented unique method+path pairs:** 64

This file is the human-readable source of truth for request/response shapes, error codes, auth, idempotency, and rate-limit buckets. The machine-readable twin is [`openapi-v1.yaml`](./openapi-v1.yaml) (OpenAPI 3.1). The two documents describe the same paths and schemas.

---

## 0. Route index

| # | Method | Path | Group | Auth | Bucket |
| --- | --- | --- | --- | --- | --- |
| 1 | GET | `/healthz` | Ops | none | `ops_health` |
| 2 | GET | `/readyz` | Ops | none | `ops_ready` |
| 3 | GET | `/v1/public/config` | Public | public | `public_config` |
| 4 | POST | `/v1/auth/register` | Auth | public | `auth_register` |
| 5 | POST | `/v1/auth/verify-email` | Auth | public | `auth_verify_email` |
| 6 | POST | `/v1/auth/resend-verification` | Auth | public | `auth_resend_verification` |
| 7 | POST | `/v1/auth/login` | Auth | public | `auth_login` |
| 8 | POST | `/v1/auth/refresh` | Auth | public | `auth_refresh` |
| 9 | POST | `/v1/auth/logout` | Auth | public | `auth_logout` |
| 10 | POST | `/v1/auth/logout-all` | Auth | bearer | `auth_logout_all` |
| 11 | POST | `/v1/auth/password-reset` | Auth | public | `auth_password_reset` |
| 12 | POST | `/v1/auth/password-reset/confirm` | Auth | public | `auth_password_reset_confirm` |
| 13 | GET | `/v1/auth/oauth/providers` | OAuth / SSO | public | `oauth_providers` |
| 14 | POST | `/v1/auth/oauth/{provider}/start` | OAuth / SSO | public | `oauth_start` |
| 15 | GET | `/v1/auth/oauth/{provider}/callback` | OAuth / SSO | public | `oauth_callback` |
| 16 | GET | `/v1/me/identities` | OAuth / SSO | bearer | `me_identities_read` |
| 17 | POST | `/v1/me/identities/{provider}/link` | OAuth / SSO | bearer | `me_identities_link` |
| 18 | DELETE | `/v1/me/identities/{provider}` | OAuth / SSO | bearer | `me_identities_unlink` |
| 19 | POST | `/v1/auth/device/code` | Desktop device OAuth | public | `auth_device_code` |
| 20 | POST | `/v1/auth/device/token` | Desktop device OAuth | public | `auth_device_token` |
| 21 | GET | `/v1/me` | Me | bearer | `me_read` |
| 22 | PATCH | `/v1/me` | Me | bearer | `me_write` |
| 23 | POST | `/v1/me/password` | Me | bearer | `me_password` |
| 24 | GET | `/v1/me/sessions` | Me | bearer | `me_sessions_read` |
| 25 | DELETE | `/v1/me/sessions/{id}` | Me | bearer | `me_sessions_revoke` |
| 26 | POST | `/v1/me/export` | Me | bearer | `me_export` |
| 27 | POST | `/v1/me/email/change` | Me | bearer | `me_email_change` |
| 28 | POST | `/v1/me/email/change/confirm` | Me | public | `me_email_change_confirm` |
| 29 | GET | `/v1/me/capabilities` | Me | bearer | `me_capabilities` |
| 30 | GET | `/v1/devices` | Devices | bearer | `devices_read` |
| 31 | POST | `/v1/devices/activate` | Devices | bearer | `devices_activate` |
| 32 | PATCH | `/v1/devices/{id}` | Devices | bearer | `devices_write` |
| 33 | DELETE | `/v1/devices/{id}` | Devices | bearer | `devices_deactivate` |
| 34 | POST | `/v1/organizations` | Organizations | bearer | `organizations_create` |
| 35 | GET | `/v1/organizations` | Organizations | bearer | `organizations_list` |
| 36 | GET | `/v1/organizations/{id}` | Organizations | bearer | `organizations_read` |
| 37 | PATCH | `/v1/organizations/{id}` | Organizations | bearer | `organizations_write` |
| 38 | DELETE | `/v1/organizations/{id}` | Organizations | bearer | `organizations_delete` |
| 39 | POST | `/v1/organizations/{id}/leave` | Organizations | bearer | `organizations_leave` |
| 40 | POST | `/v1/organizations/{id}/transfer-ownership` | Organizations | bearer | `organizations_transfer_ownership` |
| 41 | GET | `/v1/organizations/{id}/roles` | Organizations | bearer | `organizations_roles` |
| 42 | GET | `/v1/organizations/{id}/members` | Members / invitations | bearer | `members_list` |
| 43 | PATCH | `/v1/organizations/{id}/members/{userId}` | Members / invitations | bearer | `members_write` |
| 44 | POST | `/v1/organizations/{id}/members/{userId}/suspend` | Members / invitations | bearer | `members_suspend` |
| 45 | DELETE | `/v1/organizations/{id}/members/{userId}` | Members / invitations | bearer | `members_remove` |
| 46 | POST | `/v1/organizations/{id}/invitations` | Members / invitations | bearer | `invitations_create` |
| 47 | GET | `/v1/organizations/{id}/invitations` | Members / invitations | bearer | `invitations_list` |
| 48 | DELETE | `/v1/organizations/{id}/invitations/{invitationId}` | Members / invitations | bearer | `invitations_revoke` |
| 49 | POST | `/v1/organizations/{id}/invitations/{invitationId}/resend` | Members / invitations | bearer | `invitations_resend` |
| 50 | GET | `/v1/invitations/{token}` | Members / invitations | public | `invitations_preview` |
| 51 | POST | `/v1/invitations/{token}/accept` | Members / invitations | bearer | `invitations_accept` |
| 52 | GET | `/v1/organizations/{id}/entitlements` | Entitlements | bearer | `entitlements_read` |
| 53 | GET | `/v1/billing/plans` | Billing / subscription | bearer | `billing_plans` |
| 54 | GET | `/v1/organizations/{id}/subscription` | Billing / subscription | bearer | `billing_subscription_read` |
| 55 | POST | `/v1/organizations/{id}/checkout-session` | Billing / subscription | bearer | `billing_checkout` |
| 56 | POST | `/v1/organizations/{id}/portal-session` | Billing / subscription | bearer | `billing_portal` |
| 57 | POST | `/v1/organizations/{id}/subscription/cancel` | Billing / subscription | bearer | `billing_cancel` |
| 58 | POST | `/v1/organizations/{id}/subscription/resume` | Billing / subscription | bearer | `billing_resume` |
| 59 | GET | `/v1/organizations/{id}/invoices` | Billing / subscription | bearer | `billing_invoices_list` |
| 60 | GET | `/v1/organizations/{id}/invoices/{invoiceId}` | Billing / subscription | bearer | `billing_invoices_read` |
| 61 | GET | `/v1/organizations/{id}/seats` | Billing / subscription | bearer | `billing_seats` |
| 62 | POST | `/v1/organizations/{id}/seats/preview-downgrade` | Billing / subscription | bearer | `billing_preview_downgrade` |
| 63 | POST | `/v1/webhooks/billing` | Billing / subscription | webhook HMAC | `webhooks_billing` |
| 64 | GET | `/v1/organizations/{id}/audit-events` | Audit | bearer | `audit_events` |

---

## 1. Front matter

### 1.1 Transport and encoding

- HTTPS only outside local development.
- JSON request and response bodies (`Content-Type: application/json`) unless a route says otherwise.
- Public desktop contract lives under the `/v1` prefix. Ops probes (`/healthz`, `/readyz`) are unversioned.
- Timestamps are UTC ISO-8601 with a `Z` suffix, e.g. `2026-09-01T00:00:00Z`.
- Public identifiers are opaque strings (UUID-derived / ULID-style). Sequential Mongo `ObjectId` or integer keys are never exposed.
- ID prefixes: `usr_` users, `org_` organizations, `dev_` devices, `ses_` sessions, `inv_` invitations, `evt_` audit events, `req_` request ids, `sub_` subscriptions, `invc_` invoices. Additional prefixes used in this contract: `st_` OAuth state, `dc_` device codes, `rt_` refresh tokens, `cs_` checkout sessions, `bps_` billing portal sessions, `exp_` data-export jobs, `plan_` catalog plans, `ev_` / `pr_` / `ec_` one-time email tokens (never stored in plaintext).
- Field names are `snake_case`.

### 1.2 Headers (every request)

| Header | Required | Notes |
| --- | --- | --- |
| `X-Request-Id` | no | Client correlation id. If omitted, the server generates `req_…`. Always echoed as a response header **and** present as `request_id` on every JSON body (success: top-level; errors: `error.request_id`). |
| `Authorization` | when Auth = bearer | `Bearer {access_token}` |
| `Idempotency-Key` | retryable mutations as marked | Client-generated UUID. Replay window **24 hours**, scoped to the authenticated principal (or client IP + route for public auth). Same key + same body returns the original result. Same key + different body → `409 idempotency_conflict`. |
| `Content-Type` | JSON bodies | `application/json` |
| `Accept` | no | `application/json` |
| `Stripe-Signature` | billing webhook only | HMAC over the **raw** body |

Response headers of note:

| Header | When | Notes |
| --- | --- | --- |
| `X-Request-Id` | always | Echo / generated id |
| `Retry-After` | `429` and some `503` | Seconds to wait (integer) |
| `X-RateLimit-Limit` | when a bucket applies | Integer |
| `X-RateLimit-Remaining` | when a bucket applies | Integer |
| `X-RateLimit-Reset` | when a bucket applies | Unix epoch seconds (UTC) |
| `WWW-Authenticate` | `401` on bearer routes | `Bearer` |

### 1.3 Tokens

| Token | Lifetime | Storage | Rotation |
| --- | --- | --- | --- |
| Access | **15 minutes** (`expires_in: 900`) | Memory; Authorization: Bearer | Reissued on refresh |
| Refresh | **30 days** | OS credential store on desktop; hashed at rest on the server | **Rotating.** Each refresh returns a new refresh token and invalidates the presented one. Reuse of a spent token → `401 token_reuse_detected` and the **entire family is revoked**. |

Access tokens carry `sub` (`usr_`), `aud`, `exp`, and `jti`. They must **not** embed mutable plan/capability logic. Organization membership and entitlements are re-checked server-side on every request.

High-risk actions (password change, email change, ownership transfer, org delete) require recent authentication (current password in body and/or a step-up performed within the last 5 minutes).

### 1.4 Device, trial, and grace policy

| Policy | v1 value |
| --- | --- |
| Active devices per user | **3** |
| Organization trial | **14 days** from `POST /v1/organizations` |
| Paid-action / offline grace | **7 days** (`grace_until` on the entitlement snapshot and on past-due subscriptions) |

Device limit exhaustion returns `409 device_limit_reached`. The user deactivates a device and retries. Reinstallation must not permanently consume a slot.

Trial starts when the organization is created; it does not require Stripe. Checkout completion does **not** grant access by itself — verified billing state does (`BIL-002`).

After grace, the desktop enters limited mode (open/inspect/export local files). Entitlement loss must not delete or encrypt local projects.

### 1.5 Kafka is not in the HTTP contract

HTTP handlers commit application state and an **outbox row in the same Mongo transaction**. A worker publishes to Kafka (or equivalent) **after commit**. Clients never send Kafka metadata. Webhook handlers follow the same rule: verify, persist, outbox, `200`. Duplicate provider event ids are replay-safe.

### 1.6 Password hashing

Passwords are hashed with **argon2id** server-side. They are never logged, never returned, and never stored reversibly. Compromised-password screening may reject with `weak_password` without logging the secret.

### 1.7 Existence oracles

- **Login** and **password-reset start** and **resend-verification** must not reveal whether an email exists.
- **Register** **may** return `409 email_taken` (desktop needs to tell the user). Timing should still run the password hasher.
- Cross-tenant reads return `404 not_found` (not `403`) when the id exists in another organization.

### 1.8 Rate-limit buckets

Each route names a bucket. Limits are per principal when authenticated, else per IP. Exceeding a bucket returns `429 rate_limited` with `Retry-After`. Auth buckets are tighter than read buckets.

---

## 2. Shared error envelope (all errors)

Every non-success JSON body uses this shape. Never leak stack traces, SQL, Mongo, tokens, or provider payloads.

```json
{
  "error": {
    "code": "stable_snake_case",
    "message": "user-facing",
    "request_id": "req_01j6n8x4k2m9p7q3r5t8",
    "details": {}
  }
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `error.code` | string | Stable machine-readable snake_case |
| `error.message` | string | Safe to show in the desktop |
| `error.request_id` | string | Same value as `X-Request-Id` |
| `error.details` | object | Optional field-level hints (`field`, `reason`); never secrets |

### 2.1 Common codes

| HTTP | Code | Meaning |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing/expired/invalid access token |
| 401 | `invalid_credentials` | Login / current-password mismatch (no existence leak on login) |
| 401 | `token_reuse_detected` | Spent refresh token replayed; family revoked |
| 401 | `invalid_signature` | Billing webhook HMAC failed |
| 403 | `forbidden` | Authenticated but lacks role/capability |
| 403 | `tenant_mismatch` | Organization not in the caller's membership scope (prefer `not_found` for unknown ids) |
| 403 | `email_unverified` | Cloud sharing or billing change requires verified email |
| 403 | `recent_auth_required` | Step-up / current password required |
| 404 | `not_found` | Resource missing or not visible |
| 409 | `conflict` | Generic state conflict |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different body |
| 422 | `validation_error` | Schema or field validation |
| 429 | `rate_limited` | Bucket exhausted; `Retry-After` present |
| 503 | `service_unavailable` | Dependency down; retry with backoff |

`428` is used only by `POST /v1/auth/device/token` for `authorization_pending`.

---

## 3. Shared pagination

List endpoints return:

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [],
  "next_cursor": "cur_opaque_or_null"
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `items` | array | Page of resources |
| `next_cursor` | string \| null | Opaque; pass as `cursor` query param; `null` means end |
| `request_id` | string | Always present |

Default `limit` is 50; maximum 100. Cursors are not user-meaningful and must not be forged (`422 validation_error` if malformed). Do not use offset pagination on mutable collections.

---

## 4. Auth roles and capabilities

### 4.1 Roles

`owner` · `admin` · `billing_admin` · `sales_designer` · `engineer` · `production` · `viewer`

Permission checks are **capability-based**. UI visibility is not authorization. Organization-scoped routes verify membership server-side (`API-001`).

### 4.2 Product capability keys

| Key | Type | Meaning |
| --- | --- | --- |
| `cloud_projects` | boolean | Cloud project sync |
| `organization_assets` | boolean | Org 3D/texture catalog |
| `proposal_branding` | boolean | Branded proposal outputs |
| `engineering_handoff` | boolean | Engineering packet generation |

Numeric limits also appear in capability maps (`max_seats`, `max_devices`, `max_cloud_projects`).

### 4.3 Static role → API authorization (GET /v1/organizations/{id}/roles)

See that route. Product keys are intersected with the organization's plan and subscription status at entitlement evaluation time. `billing_admin` can manage billing without project-content capabilities (`ORG-005`).

---

## 5. Entitlement snapshot object

Returned by `POST /v1/devices/activate` and `GET /v1/organizations/{id}/entitlements`. The desktop verifies `signature` and caches the most recent valid snapshot.

```json
{
  "user_id": "usr_ada01f3e8c2b4a90",
  "organization_id": "org_studio9d2e4c6a80",
  "plan": "trial",
  "subscription_status": "trialing",
  "capabilities": {
    "cloud_projects": true,
    "organization_assets": true,
    "proposal_branding": true,
    "engineering_handoff": true,
    "max_seats": 3,
    "max_devices": 3,
    "max_cloud_projects": 10
  },
  "issued_at": "2026-09-01T00:00:00Z",
  "refresh_after": "2026-09-01T12:00:00Z",
  "expires_at": "2026-09-02T00:00:00Z",
  "grace_until": "2026-09-08T00:00:00Z",
  "device_id": "dev_desk01a9c4e7f2",
  "schema_version": 1,
  "signature": "c2lnbmVkLWVudGl0bGVtZW50LXYyLXRyaWFs"
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `user_id` | string | `usr_` |
| `organization_id` | string \| null | `org_`; null if the user has no org yet |
| `plan` | string | `none` \| `trial` \| `studio` \| `team` \| `enterprise` |
| `subscription_status` | string | `none` \| `trialing` \| `active` \| `past_due` \| `grace` \| `suspended` \| `canceled` \| `unpaid` \| `incomplete` |
| `capabilities` | object | Map of key → boolean or number |
| `issued_at` | string (date-time) | Snapshot issue time |
| `refresh_after` | string (date-time) | Client should refresh while online |
| `expires_at` | string (date-time) | Hard expiry of this snapshot |
| `grace_until` | string (date-time) \| null | 7-day grace end |
| `device_id` | string \| null | Bound `dev_` when issued at activation |
| `schema_version` | integer | Currently `1` |
| `signature` | string | Base64 server signature over canonical snapshot bytes (payload excluding `signature`) |

---

## 6. Shared resource shapes

### 6.1 User

```json
{
  "id": "usr_ada01f3e8c2b4a90",
  "email": "ada@example.com",
  "email_verified": true,
  "display_name": "Ada Lovelace",
  "created_at": "2026-08-01T10:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z"
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string | `usr_` |
| `email` | string | Lowercase |
| `email_verified` | boolean | |
| `display_name` | string | 1–80 |
| `created_at` | string (date-time) | UTC |
| `updated_at` | string (date-time) | UTC |

### 6.2 Organization

```json
{
  "id": "org_studio9d2e4c6a80",
  "name": "Lovelace Cabinets",
  "owner_user_id": "usr_ada01f3e8c2b4a90",
  "created_at": "2026-09-01T00:00:00Z",
  "updated_at": "2026-09-01T00:00:00Z",
  "trial_ends_at": "2026-09-15T00:00:00Z"
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string | `org_` |
| `name` | string | 1–120 |
| `owner_user_id` | string | `usr_` |
| `created_at` | string (date-time) | UTC |
| `updated_at` | string (date-time) | UTC |
| `trial_ends_at` | string (date-time) \| null | Creation + 14 days |

---

## 7. Route conventions (how to read each entry)

Every route below documents:

- Method + path
- Auth: `public` | `bearer` | `webhook HMAC` | `none`
- Required roles
- Request headers (beyond TLS)
- Path / query parameters
- Request JSON body (example + field table) or explicit none
- Success status + response JSON example + field table
- Route-specific error codes with HTTP status
- Idempotency and rate-limit bucket name

---

## 8. Routes

## 8.1 Ops

### 1. `GET /healthz`

**Group:** Ops  
**Summary:** Liveness probe. Returns 200 if the HTTP process is up. Does not check MongoDB.  
**Auth:** none  
**Required roles:** none  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `ops_health`

Must not require auth. Used by load balancers. No PII.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "ok"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | Always ok when the process can serve this route |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |

### 2. `GET /readyz`

**Group:** Ops  
**Summary:** Readiness probe. Returns 200 only if MongoDB ping and critical dependencies succeed.  
**Auth:** none  
**Required roles:** none  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `ops_ready`

Do not leak Mongo hostnames, replica-set names, or exception text.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "ready",
  "checks": {
    "mongodb": "ok"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | ready |
| checks | object | yes | Named dependency results; values are ok only — never connection strings |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 503 | `service_unavailable` | A required check failed; body uses the shared error envelope, no Mongo internals |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |

## 8.2 Public

### 3. `GET /v1/public/config`

**Group:** Public  
**Summary:** Unauthenticated bootstrap config for the desktop client.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `public_config`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "terms_version": "2026-06-01",
  "privacy_version": "2026-06-01",
  "oauth_providers": [
    "google",
    "facebook",
    "x"
  ],
  "min_app_version": "1.0.0",
  "trial_days": 14
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| terms_version | string | yes | Current terms document version the client must accept on register |
| privacy_version | string | yes | Current privacy notice version |
| oauth_providers | string[] | yes | Enabled providers: google, facebook, x |
| min_app_version | string | yes | Semver; desktop below this must force-update before login |
| trial_days | integer | yes | Always 14 for v1 |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

## 8.3 Auth

### 4. `POST /v1/auth/register`

**Group:** Auth  
**Summary:** Create an unverified user. Does not issue tokens. Password hashed with argon2id server-side. Returns 409 email_taken if the email is already registered (desktop needs a distinct UX).  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `auth_register`

Does NOT return access_token or refresh_token. Client must verify email then login.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+body returns stored result. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Password is never returned. Timing of argon2id is applied even on conflict to reduce oracle risk, but existence is revealed via 409 for this route only.

```json
{
  "email": "ada@example.com",
  "password": "correct-horse-battery-staple-19",
  "display_name": "Ada Lovelace",
  "terms_version": "2026-06-01",
  "privacy_version": "2026-06-01"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| email | string | yes | Valid email; stored lowercase; max 254 |
| password | string | yes | Min 12 chars; never logged; hashed argon2id |
| display_name | string | yes | 1–80 characters |
| terms_version | string | yes | Must equal current public config terms_version |
| privacy_version | string | yes | Must equal current public config privacy_version |

#### Success — HTTP 201

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada@example.com",
    "email_verified": false,
    "display_name": "Ada Lovelace",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | Newly created unverified user; no tokens |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 409 | `email_taken` | Email already registered |
| 422 | `terms_outdated` | terms_version or privacy_version is not current |
| 422 | `weak_password` | Password fails strength or compromised-password screen |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 5. `POST /v1/auth/verify-email`

**Group:** Auth  
**Summary:** Consume a single-use email verification token.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes, with Idempotency-Key (replay of same used token after success returns stored 200)  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `auth_verify_email`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Invalid and expired tokens share the same error to avoid token probing oracles where possible.

```json
{
  "token": "ev_single_use_opaque_token"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| token | string | yes | From verification email; hashed at rest; single-use; TTL 24h |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada@example.com",
    "email_verified": true,
    "display_name": "Ada Lovelace",
    "created_at": "2026-08-01T10:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | email_verified=true; still no tokens |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 400 | `invalid_token` | Token missing, malformed, expired, or already used |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 6. `POST /v1/auth/resend-verification`

**Group:** Auth  
**Summary:** Request a new verification email. Always 202. Does not reveal whether the email exists or is already verified.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes (same email within window is coalesced)  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `auth_resend_verification`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Constant-time style response. Email is sent only if an unverified account exists.

```json
{
  "email": "ada@example.com"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| email | string | yes | Looked up case-insensitively; no existence leak |

#### Success — HTTP 202

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "accepted"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | Always accepted |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 7. `POST /v1/auth/login`

**Group:** Auth  
**Summary:** Email/password login. Issues rotating refresh token and 15-minute access token. Does not reveal whether the email exists.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** no (new session each success)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `auth_login`

Creates a session (ses_). Refresh token is stored hashed. Audit event auth.login is written via outbox after commit.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Invalid credentials always 401 invalid_credentials with uniform timing. Unverified users may still receive tokens but cloud/billing routes return 403 email_unverified.

```json
{
  "email": "ada@example.com",
  "password": "correct-horse-battery-staple-19",
  "device_name": "Ada’s Studio Mac"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| email | string | yes | Normalized lowercase |
| password | string | yes | Never logged; compared via argon2id |
| device_name | string | no | Human label for the session; 1–80 chars |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada@example.com",
    "email_verified": true,
    "display_name": "Ada Lovelace",
    "created_at": "2026-08-01T10:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  },
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.access.example",
  "refresh_token": "rt_rotating_01j6n8x4k2m9p7q3r5t8",
  "expires_in": 900,
  "token_type": "Bearer",
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinets",
    "owner_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | Authenticated user |
| access_token | string | yes | JWT; 15 minutes |
| refresh_token | string | yes | Opaque; 30 days; rotating |
| expires_in | integer | yes | 900 |
| token_type | string | yes | Bearer |
| organization | object \| null | yes | Primary/last-used org if the user has memberships; else null |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `invalid_credentials` | Email unknown or password mismatch — identical response |
| 403 | `account_suspended` | User record is suspended |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 8. `POST /v1/auth/refresh`

**Group:** Auth  
**Summary:** Rotate refresh token. Reuse of a spent refresh token revokes the entire family and returns 401 token_reuse_detected.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** no (rotation)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `auth_refresh`

Concurrent refresh of the same token: first wins; loser is treated as reuse → family revoke. Clients must serialize refresh.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Successful refresh invalidates the presented token and returns a new pair. Access token still 15m; refresh TTL is renewed to 30d from now.

```json
{
  "refresh_token": "rt_rotating_01j6n8x4k2m9p7q3r5t8"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| refresh_token | string | yes | Current unexpired refresh token |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.access.example",
  "refresh_token": "rt_rotating_01j6n8x4k2m9p7q3r5t9",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| access_token | string | yes | JWT; 15 minutes; Authorization: Bearer |
| refresh_token | string | yes | Opaque rotating token; 30 days; single-use |
| expires_in | integer | yes | Access token lifetime in seconds (900) |
| token_type | string | yes | Always Bearer |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Refresh token missing, expired, or revoked |
| 401 | `token_reuse_detected` | Spent refresh token replayed; family revoked |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 9. `POST /v1/auth/logout`

**Group:** Auth  
**Summary:** Revoke one refresh-token family member (this session).  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `auth_logout`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Always 204-equivalent success even if the token is already unknown, to avoid oracles. v1 returns 200 with status.

```json
{
  "refresh_token": "rt_rotating_01j6n8x4k2m9p7q3r5t8"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| refresh_token | string | yes | Token to revoke; hashed lookup |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "logged_out"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | logged_out |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 10. `POST /v1/auth/logout-all`

**Group:** Auth  
**Summary:** Revoke every refresh-token family for the authenticated user.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `auth_logout_all`

The presented access token remains valid until expiry (≤15m) unless the implementation also maintains an access jti denylist. Desktop must discard both tokens.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body. Identity comes from the access token.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "logged_out_all",
  "revoked_sessions": 3
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | logged_out_all |
| revoked_sessions | integer | yes | Count of sessions revoked including the caller |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 11. `POST /v1/auth/password-reset`

**Group:** Auth  
**Summary:** Start password reset. Always 202. Does not reveal account existence.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes (coalesced per email)  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `auth_password_reset`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Sends email only if an account exists. Token is single-use, hashed at rest, TTL 1h.

```json
{
  "email": "ada@example.com"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| email | string | yes | No existence leak |

#### Success — HTTP 202

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "accepted"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | Always accepted |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 12. `POST /v1/auth/password-reset/confirm`

**Group:** Auth  
**Summary:** Consume password-reset token and set a new password. Revokes all sessions.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `auth_password_reset_confirm`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+body returns stored result. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Does not auto-login. Client must call login afterwards.

```json
{
  "token": "pr_single_use_opaque_token",
  "new_password": "new-correct-horse-battery-21"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| token | string | yes | From reset email; single-use; hashed at rest |
| new_password | string | yes | Min 12; argon2id; never logged |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "password_reset"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | password_reset |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 400 | `invalid_token` | Token missing, expired, or already used |
| 422 | `weak_password` | Password fails strength policy |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

## 8.4 OAuth / SSO

### 13. `GET /v1/auth/oauth/providers`

**Group:** OAuth / SSO  
**Summary:** List OAuth providers enabled for this environment.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `oauth_providers`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "providers": [
    {
      "id": "google",
      "label": "Google",
      "device_code_supported": true
    },
    {
      "id": "facebook",
      "label": "Facebook",
      "device_code_supported": false
    },
    {
      "id": "x",
      "label": "X",
      "device_code_supported": false
    }
  ]
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| providers | object[] | yes | Enabled providers |
| providers[].id | string | yes | google \| facebook \| x |
| providers[].label | string | yes | Display label |
| providers[].device_code_supported | boolean | yes | Whether /v1/auth/device/code accepts this provider |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 14. `POST /v1/auth/oauth/{provider}/start`

**Group:** OAuth / SSO  
**Summary:** Start OAuth authorization-code flow. Returns the provider authorization URL and a server-stored state.  
**Auth:** public  
**Required roles:** none (intent=login) or bearer (intent=link recommended)  
**Idempotent?** yes with Idempotency-Key; otherwise new state each call  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `oauth_start`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| provider | string | yes | google \| facebook \| x |

#### Query parameters

_None._

#### Request JSON body

state is bound to intent, redirect_uri, PKCE verifier, and optional user_id when linking. TTL 10 minutes.

```json
{
  "redirect_uri": "cabinetstudio://oauth/callback",
  "intent": "login"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| redirect_uri | string | yes | Must be on the server allowlist (desktop custom scheme or https loopback) |
| intent | string | yes | login or link |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "authorization_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&state=st_...&redirect_uri=...",
  "state": "st_01j6n8x4k2m9p7q3r5t8"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| authorization_url | string | yes | Full URL to open in system browser |
| state | string | yes | Opaque CSRF state; prefix st_ |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 400 | `unsupported_provider` | Path provider is not enabled |
| 422 | `redirect_uri_not_allowed` | redirect_uri not on allowlist |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 15. `GET /v1/auth/oauth/{provider}/callback`

**Group:** OAuth / SSO  
**Summary:** OAuth redirect callback. Exchanges code+state. Desktop typically uses 200 JSON; browser may 302 to redirect_uri with a one-time code.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** code is single-use; replay → invalid_state  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `oauth_callback`

Never return raw provider tokens or provider payload. New users from login intent are email_verified when the provider email is verified.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| provider | string | yes | google \| facebook \| x |

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| code | string | yes | Provider authorization code |
| state | string | yes | Must match start-flow state |
| error | string | no | Provider error code if the user denied |
| error_description | string | no | Provider description; not echoed verbatim to clients beyond stable codes |

#### Request JSON body

No JSON body. Query-only.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada@example.com",
    "email_verified": true,
    "display_name": "Ada Lovelace",
    "created_at": "2026-08-01T10:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  },
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.access.example",
  "refresh_token": "rt_rotating_01j6n8x4k2m9p7q3r5t8",
  "expires_in": 900,
  "token_type": "Bearer",
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinets",
    "owner_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  },
  "linked_provider": "google"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | Existing or newly provisioned user (email verified if provider asserts email) |
| access_token | string | yes | Issued when intent=login |
| refresh_token | string | yes | Issued when intent=login |
| expires_in | integer | yes | 900 |
| token_type | string | yes | Bearer |
| organization | object \| null | yes | Primary org if any |
| linked_provider | string | yes | Provider id just completed |

#### Additional status — HTTP 302

When the stored redirect_uri is a non-JSON client; Location never includes tokens

```json
{
  "Location": "cabinetstudio://oauth/callback?status=ok"
}
```

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 400 | `invalid_state` | state unknown, expired, or already used |
| 400 | `oauth_denied` | User denied or provider returned error |
| 409 | `identity_already_linked` | Provider identity belongs to a different user |
| 409 | `email_taken` | Login intent would create a user whose email belongs to another account |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 16. `GET /v1/me/identities`

**Group:** OAuth / SSO  
**Summary:** List social identities linked to the current user.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `me_identities_read`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "provider": "google",
      "provider_user_id": "gid_opaque",
      "email": "ada@example.com",
      "linked_at": "2026-08-12T09:00:00Z"
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Linked identities |
| items[].provider | string | yes | google \| facebook \| x |
| items[].provider_user_id | string | yes | Opaque provider subject; not a secret |
| items[].email | string \| null | yes | Email asserted by provider if any |
| items[].linked_at | string (date-time) | yes | UTC |
| next_cursor | string \| null | yes | Always null in v1 (small set) |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 17. `POST /v1/me/identities/{provider}/link`

**Group:** OAuth / SSO  
**Summary:** Link a provider identity to the current user using the authorization code from a start intent=link flow.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `me_identities_link`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| provider | string | yes | google \| facebook \| x |

#### Query parameters

_None._

#### Request JSON body

Alternatively the desktop may complete via GET callback with a linked session; this POST is the explicit API for the app.

```json
{
  "authorization_code": "4/0Aan...",
  "state": "st_01j6n8x4k2m9p7q3r5t8"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| authorization_code | string | yes | Provider code from the completed browser flow |
| state | string | yes | Must match start-flow state with intent=link and this user |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "identity": {
    "provider": "google",
    "provider_user_id": "gid_opaque",
    "email": "ada@example.com",
    "linked_at": "2026-09-01T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| identity | object | yes | Newly linked identity |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 400 | `invalid_state` | state not a link intent for this user |
| 409 | `identity_already_linked` | This provider identity is linked to another user |
| 409 | `provider_already_linked` | This user already has this provider |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 18. `DELETE /v1/me/identities/{provider}`

**Group:** OAuth / SSO  
**Summary:** Unlink a provider identity. Cannot remove the last remaining login method.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `me_identities_unlink`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| provider | string | yes | google \| facebook \| x |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "unlinked",
  "provider": "google"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | unlinked |
| provider | string | yes | Provider just removed |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `last_login_method` | Would leave the user with no password and no remaining identity |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

## 8.5 Desktop device OAuth

### 19. `POST /v1/auth/device/code`

**Group:** Desktop device OAuth  
**Summary:** RFC 8628-style device authorization for desktop. User types user_code at verification_uri.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes with Idempotency-Key; otherwise new codes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `auth_device_code`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

user_code is shown in the Tauri window. device_code never displayed.

```json
{
  "client_id": "cabinet-studio-desktop",
  "provider": "google"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| client_id | string | yes | Must be cabinet-studio-desktop in v1 |
| provider | string | yes | google (required for v1 device flow); facebook/x may return unsupported_provider |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "device_code": "dc_01j6n8x4k2m9p7q3r5t8",
  "user_code": "WDJB-MJHT",
  "verification_uri": "https://app.cabinetstudio.example/device",
  "verification_uri_complete": "https://app.cabinetstudio.example/device?user_code=WDJB-MJHT",
  "expires_in": 600,
  "interval": 5
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| device_code | string | yes | Opaque; prefix dc_; polled by the desktop |
| user_code | string | yes | Short code the user types |
| verification_uri | string | yes | HTTPS page to open |
| verification_uri_complete | string | no | Optional prefilled URL |
| expires_in | integer | yes | Seconds until device_code expires (600) |
| interval | integer | yes | Minimum seconds between token polls (5) |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 400 | `unsupported_provider` | Provider does not support device flow |
| 401 | `invalid_client` | client_id unknown |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 20. `POST /v1/auth/device/token`

**Group:** Desktop device OAuth  
**Summary:** Poll for device-flow completion. 200 tokens, 428 authorization_pending, 403 slow_down / expired_token / access_denied.  
**Auth:** public  
**Required roles:** none  
**Idempotent?** poll is intentionally not idempotent; 200 is single-issue  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `auth_device_token`

428 is used for authorization_pending so clients can distinguish pending from success without parsing codes only. Error envelope still applies.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Clients MUST wait `interval` seconds. Polling faster returns 403 slow_down.

```json
{
  "device_code": "dc_01j6n8x4k2m9p7q3r5t8"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| device_code | string | yes | From /v1/auth/device/code |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada@example.com",
    "email_verified": true,
    "display_name": "Ada Lovelace",
    "created_at": "2026-08-01T10:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  },
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.access.example",
  "refresh_token": "rt_rotating_01j6n8x4k2m9p7q3r5t8",
  "expires_in": 900,
  "token_type": "Bearer",
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinets",
    "owner_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | Authenticated user after approval |
| access_token | string | yes | JWT 15m |
| refresh_token | string | yes | Rotating 30d |
| expires_in | integer | yes | 900 |
| token_type | string | yes | Bearer |
| organization | object \| null | yes | Primary org if any |

#### Additional status — HTTP 428

User has not completed verification; keep polling

```json
{
  "error": {
    "code": "authorization_pending",
    "message": "User has not yet approved this device.",
    "request_id": "req_01j6n8x4k2m9p7q3r5t8",
    "details": {
      "interval": 5
    }
  }
}
```

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 403 | `slow_down` | Polled faster than interval; increase wait |
| 403 | `expired_token` | device_code expired |
| 403 | `access_denied` | User denied the request |
| 400 | `invalid_grant` | device_code unknown |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

## 8.6 Me

### 21. `GET /v1/me`

**Group:** Me  
**Summary:** Current user profile.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `me_read`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada@example.com",
    "email_verified": true,
    "display_name": "Ada Lovelace",
    "created_at": "2026-08-01T10:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | Current user |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 22. `PATCH /v1/me`

**Group:** Me  
**Summary:** Update profile fields. Email is not changed here (see /v1/me/email/change).  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `me_write`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Unknown fields → validation_error. Empty patch → validation_error.

```json
{
  "display_name": "Ada L."
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| display_name | string | yes | 1–80 characters; the only v1 patchable field |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada@example.com",
    "email_verified": true,
    "display_name": "Ada L.",
    "created_at": "2026-08-01T10:00:00Z",
    "updated_at": "2026-09-01T00:15:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | Updated user |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 23. `POST /v1/me/password`

**Group:** Me  
**Summary:** Change password with current password (recent authentication). Revokes other sessions; keeps current refresh family.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `me_password`

High-risk. Audit auth.password_change via outbox after commit.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |
| X-Cabinet-Reauth | no | Optional proof of recent authentication (step-up). Some high-risk routes also accept current_password in body. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

OAuth-only users without a password receive 409 password_not_set — they must set password via reset after adding email, or keep OAuth.

```json
{
  "current_password": "correct-horse-battery-staple-19",
  "new_password": "new-correct-horse-battery-21"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| current_password | string | yes | Must match; uniform invalid_credentials on mismatch |
| new_password | string | yes | Min 12; must differ from current; argon2id |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "password_changed"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | password_changed |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 401 | `invalid_credentials` | current_password mismatch |
| 403 | `recent_auth_required` | Action requires recent authentication (re-enter password or step-up) |
| 409 | `password_not_set` | Account has no password (OAuth-only) |
| 422 | `weak_password` | Fails strength policy or equals current |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 24. `GET /v1/me/sessions`

**Group:** Me  
**Summary:** List refresh sessions for the current user.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `me_sessions_read`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| cursor | string | no | Opaque pagination cursor |
| limit | integer | no | 1–100; default 50 |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "id": "ses_web01b8d3c6e4",
      "created_at": "2026-08-20T08:00:00Z",
      "last_seen_at": "2026-09-01T00:00:00Z",
      "expires_at": "2026-10-01T00:00:00Z",
      "device_name": "Ada’s Studio Mac",
      "device_id": "dev_desk01a9c4e7f2",
      "ip_country": "IN",
      "user_agent_summary": "Cabinet Studio/1.2.0 macOS",
      "current": true
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Sessions; never includes refresh token values or hashes |
| items[].id | string | yes | ses_ |
| items[].created_at | string (date-time) | yes | UTC |
| items[].last_seen_at | string (date-time) | yes | UTC |
| items[].expires_at | string (date-time) | yes | UTC |
| items[].device_name | string \| null | yes | From login device_name |
| items[].device_id | string \| null | yes | dev_ if bound |
| items[].ip_country | string \| null | yes | ISO country from last seen IP; never raw IP in v1 list |
| items[].user_agent_summary | string \| null | yes | Truncated UA product |
| items[].current | boolean | yes | True for the session that minted this access token |
| next_cursor | string \| null | yes | Opaque; null if no more |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 25. `DELETE /v1/me/sessions/{id}`

**Group:** Me  
**Summary:** Revoke one session. May be the current session.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `me_sessions_revoke`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | ses_ to revoke |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "revoked",
  "id": "ses_web01b8d3c6e4"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | revoked |
| id | string | yes | ses_ revoked |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 26. `POST /v1/me/export`

**Group:** Me  
**Summary:** Request a personal-data export. Asynchronous. 202 with export_id.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `me_export`

Job is written after commit via outbox. Download of the archive is a later route; v1 only starts the job. Contains profile, memberships, sessions metadata — not org-owned projects.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body. Idempotency-Key required so double-clicks do not spawn duplicate jobs.

#### Success — HTTP 202

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "export_id": "exp_me01f4c9d2e6",
  "status": "queued"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| export_id | string | yes | exp_ job id |
| status | string | yes | queued \| processing \| ready \| failed |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 409 | `export_in_progress` | An export is already queued or processing for this user |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 27. `POST /v1/me/email/change`

**Group:** Me  
**Summary:** Start email change. Requires current password. Sends confirmation to the new address. Does not change email until confirm.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `me_email_change`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |
| X-Cabinet-Reauth | no | Optional proof of recent authentication (step-up). Some high-risk routes also accept current_password in body. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

If new_email is taken, v1 returns 409 email_taken (user is authenticated, existence of *their intended* address is acceptable). Confirmation token TTL 24h.

```json
{
  "new_email": "ada.new@example.com",
  "password": "correct-horse-battery-staple-19"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| new_email | string | yes | Must not be the current email; unique |
| password | string | yes | Current password; OAuth-only users: 409 password_not_set |

#### Success — HTTP 202

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "confirmation_sent",
  "new_email": "ada.new@example.com"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | confirmation_sent |
| new_email | string | yes | Normalized target email |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 401 | `invalid_credentials` | password mismatch |
| 403 | `recent_auth_required` | Action requires recent authentication (re-enter password or step-up) |
| 409 | `email_taken` | new_email already registered |
| 409 | `password_not_set` | OAuth-only account |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 28. `POST /v1/me/email/change/confirm`

**Group:** Me  
**Summary:** Consume email-change token. Updates email and marks it verified. Revokes other sessions.  
**Auth:** public  
**Required roles:** none (token is the credential); bearer optional  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `me_email_change_confirm`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | no | Optional Bearer; if present must be the user who requested the change |
| X-Request-Id | no | Client correlation id. Echoed; generated as req_ if omitted. |
| Content-Type | yes | application/json |
| Accept | no | application/json |
| Idempotency-Key | no | Recommended; 24h replay window |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Public because the user may confirm from a mail client. If bearer is present it must match the user that requested the change.

```json
{
  "token": "ec_single_use_opaque_token"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| token | string | yes | From the new-email inbox; single-use; hashed at rest |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user": {
    "id": "usr_ada01f3e8c2b4a90",
    "email": "ada.new@example.com",
    "email_verified": true,
    "display_name": "Ada Lovelace",
    "created_at": "2026-08-01T10:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user | object | yes | Updated user; email_verified=true |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 400 | `invalid_token` | Token missing, expired, or already used |
| 409 | `email_taken` | Race: email claimed after start |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 29. `GET /v1/me/capabilities`

**Group:** Me  
**Summary:** Effective capabilities and role for a given organization.  
**Auth:** bearer  
**Required roles:** member of the organization  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `me_capabilities`

Suspended members receive 403 forbidden. This is unsigned and for UI; desktop licensing uses GET /v1/organizations/{id}/entitlements.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| organization_id | string | yes | org_ to evaluate |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "user_id": "usr_ada01f3e8c2b4a90",
  "organization_id": "org_studio9d2e4c6a80",
  "role": "owner",
  "membership_status": "active",
  "plan": "trial",
  "subscription_status": "trialing",
  "capabilities": {
    "cloud_projects": true,
    "organization_assets": true,
    "proposal_branding": true,
    "engineering_handoff": true,
    "max_seats": 3,
    "max_devices": 3,
    "max_cloud_projects": 10
  },
  "issued_at": "2026-09-01T00:00:00Z"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| user_id | string | yes | usr_ |
| organization_id | string | yes | org_ |
| role | string | yes | owner \| admin \| billing_admin \| sales_designer \| engineer \| production \| viewer |
| membership_status | string | yes | active \| suspended |
| plan | string | yes | Normalized plan key |
| subscription_status | string | yes | Normalized billing status |
| capabilities | object | yes | Effective map (role ∩ plan ∩ subscription). Booleans and numeric limits. |
| issued_at | string (date-time) | yes | Evaluation time; this response is not signed (see entitlements snapshot) |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

## 8.7 Devices

### 30. `GET /v1/devices`

**Group:** Devices  
**Summary:** List devices activated for the current user (limit 3 active).  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `devices_read`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| cursor | string | no | Opaque cursor |
| limit | integer | no | 1–100; default 50 |
| include_deactivated | boolean | no | Default false |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "id": "dev_desk01a9c4e7f2",
      "name": "Studio iMac",
      "platform": "macos",
      "created_at": "2026-08-02T11:00:00Z",
      "last_seen_at": "2026-09-01T00:00:00Z",
      "deactivated_at": null
    }
  ],
  "next_cursor": null,
  "active_count": 1,
  "device_limit": 3
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Devices |
| items[].id | string | yes | dev_ |
| items[].name | string | yes | User-visible name |
| items[].platform | string | yes | macos \| windows \| linux |
| items[].created_at | string (date-time) | yes | UTC |
| items[].last_seen_at | string (date-time) | yes | UTC |
| items[].deactivated_at | string (date-time) \| null | yes | Null if active |
| next_cursor | string \| null | yes | Pagination |
| active_count | integer | yes | Currently active devices |
| device_limit | integer | yes | Always 3 in v1 |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 31. `POST /v1/devices/activate`

**Group:** Devices  
**Summary:** Activate this installation. Returns the device and a signed entitlement snapshot. Fails if 3 devices are already active.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes, with Idempotency-Key (replay returns the same device)  
**Idempotency-Key:** required  
**Rate-limit bucket:** `devices_activate`

If the user has no organization yet, entitlement_snapshot.organization_id is null, plan is none, capabilities are all false / zeros, signature still present.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Reinstallation should not permanently consume a slot: deactivate the old device first. Fingerprints are not accepted in v1 body (privacy).

```json
{
  "name": "Studio iMac",
  "platform": "macos"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| name | string | yes | 1–80 characters |
| platform | string | yes | macos \| windows \| linux |

#### Success — HTTP 201

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "device": {
    "id": "dev_desk01a9c4e7f2",
    "name": "Studio iMac",
    "platform": "macos",
    "created_at": "2026-09-01T00:00:00Z",
    "last_seen_at": "2026-09-01T00:00:00Z",
    "deactivated_at": null
  },
  "entitlement_snapshot": {
    "user_id": "usr_ada01f3e8c2b4a90",
    "organization_id": "org_studio9d2e4c6a80",
    "plan": "trial",
    "subscription_status": "trialing",
    "capabilities": {
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": true,
      "engineering_handoff": true,
      "max_seats": 3,
      "max_devices": 3,
      "max_cloud_projects": 10
    },
    "issued_at": "2026-09-01T00:00:00Z",
    "refresh_after": "2026-09-01T12:00:00Z",
    "expires_at": "2026-09-02T00:00:00Z",
    "grace_until": "2026-09-08T00:00:00Z",
    "device_id": "dev_desk01a9c4e7f2",
    "schema_version": 1,
    "signature": "c2lnbmVkLWVudGl0bGVtZW50LXYyLXRyaWFs"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| device | object | yes | Activated device |
| entitlement_snapshot | object | yes | Signed snapshot bound to this device_id |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 409 | `device_limit_reached` | Already 3 active devices; deactivate one |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 32. `PATCH /v1/devices/{id}`

**Group:** Devices  
**Summary:** Rename a device owned by the current user.  
**Auth:** bearer  
**Required roles:** any authenticated user (owner of the device)  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `devices_write`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | dev_ |

#### Query parameters

_None._

#### Request JSON body

Cannot patch platform. Deactivated devices may still be renamed.

```json
{
  "name": "Shop Floor PC"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| name | string | yes | 1–80 characters |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "device": {
    "id": "dev_desk01a9c4e7f2",
    "name": "Shop Floor PC",
    "platform": "macos",
    "created_at": "2026-08-02T11:00:00Z",
    "last_seen_at": "2026-09-01T00:00:00Z",
    "deactivated_at": null
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| device | object | yes | Updated device |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 33. `DELETE /v1/devices/{id}`

**Group:** Devices  
**Summary:** Deactivate a device. Frees a slot. Does not delete historical audit rows.  
**Auth:** bearer  
**Required roles:** any authenticated user (owner of the device)  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `devices_deactivate`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | dev_ |

#### Query parameters

_None._

#### Request JSON body

No request body. Idempotent: already deactivated → 200.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "deactivated",
  "id": "dev_desk01a9c4e7f2"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | deactivated |
| id | string | yes | dev_ |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

## 8.8 Organizations

### 34. `POST /v1/organizations`

**Group:** Organizations  
**Summary:** Create an organization, owner membership, and start a 14-day trial.  
**Auth:** bearer  
**Required roles:** any authenticated user with verified email  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `organizations_create`

Outbox: org.created, membership.created, billing.trial_started. Kafka is not called from the handler.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Creator becomes owner. Trial 14 days. No Stripe customer is required to start trial.

```json
{
  "name": "Lovelace Cabinets"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| name | string | yes | 1–120 characters |

#### Success — HTTP 201

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinets",
    "owner_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  },
  "membership": {
    "user_id": "usr_ada01f3e8c2b4a90",
    "organization_id": "org_studio9d2e4c6a80",
    "role": "owner",
    "status": "active",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  },
  "trial_ends_at": "2026-09-15T00:00:00Z"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| organization | object | yes | Created org |
| membership | object | yes | Owner membership |
| trial_ends_at | string (date-time) | yes | created_at + 14 days |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `email_unverified` | Email must be verified before creating an org |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 35. `GET /v1/organizations`

**Group:** Organizations  
**Summary:** List organizations the current user belongs to.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `organizations_list`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| cursor | string | no | Opaque cursor |
| limit | integer | no | 1–100; default 50 |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "organization": {
        "id": "org_studio9d2e4c6a80",
        "name": "Lovelace Cabinets",
        "owner_user_id": "usr_ada01f3e8c2b4a90",
        "created_at": "2026-09-01T00:00:00Z",
        "updated_at": "2026-09-01T00:00:00Z",
        "trial_ends_at": "2026-09-15T00:00:00Z"
      },
      "role": "owner",
      "membership_status": "active"
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Memberships with embedded org |
| items[].organization | object | yes | Organization |
| items[].role | string | yes | Role in that org |
| items[].membership_status | string | yes | active \| suspended |
| next_cursor | string \| null | yes | Pagination |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 36. `GET /v1/organizations/{id}`

**Group:** Organizations  
**Summary:** Get one organization. Caller must be a member.  
**Auth:** bearer  
**Required roles:** any org member (including suspended? no — active member)  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `organizations_read`

Unknown ids and ids the caller cannot see both return 404 not_found (no existence leak across tenants).

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinets",
    "owner_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| organization | object | yes | Organization |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 37. `PATCH /v1/organizations/{id}`

**Group:** Organizations  
**Summary:** Rename an organization.  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `organizations_write`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

billing_admin cannot rename.

```json
{
  "name": "Lovelace Cabinet Studio"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| name | string | yes | 1–120 characters |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinet Studio",
    "owner_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:15:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| organization | object | yes | Updated organization |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 38. `DELETE /v1/organizations/{id}`

**Group:** Organizations  
**Summary:** Close/delete organization. Owner only. Last-owner close. Irreversible after retention window; v1 marks deleted immediately for API purposes.  
**Auth:** bearer  
**Required roles:** owner  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `organizations_delete`

Does not delete member personal accounts. Org-owned cloud data enters retention. Audit org.deleted via outbox.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |
| X-Cabinet-Reauth | no | Optional proof of recent authentication (step-up). Some high-risk routes also accept current_password in body. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

No body. High-risk; recent authentication required.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "deleted",
  "id": "org_studio9d2e4c6a80"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | deleted |
| id | string | yes | org_ |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 403 | `recent_auth_required` | Action requires recent authentication (re-enter password or step-up) |
| 409 | `active_subscription` | Cancel subscription first (or it is not in a closeable state) |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 39. `POST /v1/organizations/{id}/leave`

**Group:** Organizations  
**Summary:** Leave an organization. Last owner cannot leave; they must transfer or delete.  
**Auth:** bearer  
**Required roles:** any active member except last owner  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `organizations_leave`

Revokes this user's org-scoped sessions/entitlements. Does not delete org data they authored.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "left",
  "organization_id": "org_studio9d2e4c6a80"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | left |
| organization_id | string | yes | org_ |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `last_owner` | Last owner cannot leave without transfer-ownership or delete |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 40. `POST /v1/organizations/{id}/transfer-ownership`

**Group:** Organizations  
**Summary:** Transfer ownership to another active member. Requires recent authentication.  
**Auth:** bearer  
**Required roles:** owner  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `organizations_transfer_ownership`

Audit org.ownership_transferred. High-risk.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |
| X-Cabinet-Reauth | no | Optional proof of recent authentication (step-up). Some high-risk routes also accept current_password in body. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

Current owner becomes admin. Target must already be an active member.

```json
{
  "new_owner_user_id": "usr_bb02e4d9c1a7f380"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| new_owner_user_id | string | yes | usr_ of an active member; cannot be self |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinets",
    "owner_user_id": "usr_bb02e4d9c1a7f380",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:15:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  },
  "previous_owner_user_id": "usr_ada01f3e8c2b4a90",
  "new_owner_user_id": "usr_bb02e4d9c1a7f380"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| organization | object | yes | Updated org |
| previous_owner_user_id | string | yes | usr_ |
| new_owner_user_id | string | yes | usr_ |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 403 | `recent_auth_required` | Action requires recent authentication (re-enter password or step-up) |
| 409 | `already_owner` | Target is already the owner |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 41. `GET /v1/organizations/{id}/roles`

**Group:** Organizations  
**Summary:** Static role → capabilities map for this API version (not plan-specific). Plan limits still apply at entitlement time.  
**Auth:** bearer  
**Required roles:** any active member  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `organizations_roles`

Product capability keys: cloud_projects, organization_assets, proposal_branding, engineering_handoff. Additional manage_* keys describe API authorization, not licensed product features.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "schema_version": 1,
  "roles": {
    "owner": {
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": true,
      "engineering_handoff": true,
      "manage_members": true,
      "manage_billing": true,
      "transfer_ownership": true,
      "delete_organization": true
    },
    "admin": {
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": true,
      "engineering_handoff": true,
      "manage_members": true,
      "manage_billing": false,
      "transfer_ownership": false,
      "delete_organization": false
    },
    "billing_admin": {
      "cloud_projects": false,
      "organization_assets": false,
      "proposal_branding": false,
      "engineering_handoff": false,
      "manage_members": false,
      "manage_billing": true,
      "transfer_ownership": false,
      "delete_organization": false
    },
    "sales_designer": {
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": true,
      "engineering_handoff": false,
      "manage_members": false,
      "manage_billing": false,
      "transfer_ownership": false,
      "delete_organization": false
    },
    "engineer": {
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": false,
      "engineering_handoff": true,
      "manage_members": false,
      "manage_billing": false,
      "transfer_ownership": false,
      "delete_organization": false
    },
    "production": {
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": false,
      "engineering_handoff": false,
      "manage_members": false,
      "manage_billing": false,
      "transfer_ownership": false,
      "delete_organization": false
    },
    "viewer": {
      "cloud_projects": false,
      "organization_assets": false,
      "proposal_branding": false,
      "engineering_handoff": false,
      "manage_members": false,
      "manage_billing": false,
      "transfer_ownership": false,
      "delete_organization": false
    }
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| schema_version | integer | yes | 1 |
| roles | object | yes | Map of role name → capability/permission booleans |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

## 8.9 Members / invitations

### 42. `GET /v1/organizations/{id}/members`

**Group:** Members / invitations  
**Summary:** List organization members.  
**Auth:** bearer  
**Required roles:** owner, admin (full emails); other members see directory without billing-only filtering — all active members may list  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `members_list`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| cursor | string | no | Opaque cursor |
| limit | integer | no | 1–100; default 50 |
| status | string | no | active \| suspended; default all non-removed |
| role | string | no | Filter by role |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "user_id": "usr_ada01f3e8c2b4a90",
      "organization_id": "org_studio9d2e4c6a80",
      "role": "owner",
      "status": "active",
      "created_at": "2026-08-01T10:05:00Z",
      "updated_at": "2026-09-01T00:00:00Z",
      "user": {
        "id": "usr_ada01f3e8c2b4a90",
        "email": "ada@example.com",
        "display_name": "Ada Lovelace",
        "email_verified": true
      }
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Members |
| items[].user_id | string | yes | usr_ |
| items[].organization_id | string | yes | org_ |
| items[].role | string | yes | Role key |
| items[].status | string | yes | active \| suspended |
| items[].created_at | string (date-time) | yes | UTC |
| items[].updated_at | string (date-time) | yes | UTC |
| items[].user | object | yes | Public user projection: id, email, display_name, email_verified |
| next_cursor | string \| null | yes | Pagination |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 43. `PATCH /v1/organizations/{id}/members/{userId}`

**Group:** Members / invitations  
**Summary:** Change a member's role. Cannot change the last owner's role.  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `members_write`

ORG-003: role changes revoke stale authorization promptly (refresh entitlement; access token claims are not the source of truth).

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |
| userId | string | yes | usr_ |

#### Query parameters

_None._

#### Request JSON body

Admin cannot promote to owner or demote another admin/owner. Owner can assign any non-owner role.

```json
{
  "role": "engineer"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| role | string | yes | admin \| billing_admin \| sales_designer \| engineer \| production \| viewer. owner is not assignable here — use transfer-ownership. |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "membership": {
    "user_id": "usr_ada01f3e8c2b4a90",
    "organization_id": "org_studio9d2e4c6a80",
    "role": "engineer",
    "status": "active",
    "created_at": "2026-08-01T10:05:00Z",
    "updated_at": "2026-09-01T00:15:00Z",
    "user": {
      "id": "usr_ada01f3e8c2b4a90",
      "email": "ada@example.com",
      "display_name": "Ada Lovelace",
      "email_verified": true
    }
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| membership | object | yes | Updated membership |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `last_owner` | Cannot change role of the last owner |
| 422 | `role_not_assignable` | role=owner is not valid on this route |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 44. `POST /v1/organizations/{id}/members/{userId}/suspend`

**Group:** Members / invitations  
**Summary:** Suspend a member. Owner cannot be suspended. Repeat POST is idempotent. v1 has no unsuspend path: owner/admin PATCH role reactivates (status returns to active).  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `members_suspend`

There is no unsuspend path in v1. Owner/admin reactivates a suspended member by PATCH /members/{userId} {role} which sets status back to active. Sessions/entitlements are revoked on suspend.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |
| userId | string | yes | usr_ |

#### Query parameters

_None._

#### Request JSON body

No request body. Idempotent freeze.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "membership": {
    "user_id": "usr_ada01f3e8c2b4a90",
    "organization_id": "org_studio9d2e4c6a80",
    "role": "owner",
    "status": "suspended",
    "created_at": "2026-08-01T10:05:00Z",
    "updated_at": "2026-09-01T00:15:00Z",
    "user": {
      "id": "usr_ada01f3e8c2b4a90",
      "email": "ada@example.com",
      "display_name": "Ada Lovelace",
      "email_verified": true
    }
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| membership | object | yes | status=suspended |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `last_owner` | Cannot suspend the last owner |
| 409 | `cannot_suspend_self` | Caller cannot suspend themselves |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 45. `DELETE /v1/organizations/{id}/members/{userId}`

**Group:** Members / invitations  
**Summary:** Remove a member. Does not delete the user account or organization-owned data they authored.  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `members_remove`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |
| userId | string | yes | usr_ |

#### Query parameters

_None._

#### Request JSON body

No request body. Removing self is not allowed; use leave.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "removed",
  "user_id": "usr_bb02e4d9c1a7f380",
  "organization_id": "org_studio9d2e4c6a80"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | removed |
| user_id | string | yes | usr_ |
| organization_id | string | yes | org_ |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `last_owner` | Cannot remove the last owner |
| 409 | `cannot_remove_self` | Use POST /leave |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 46. `POST /v1/organizations/{id}/invitations`

**Group:** Members / invitations  
**Summary:** Invite a person by email. Full invitation token is sent only in email, never in list or create API bodies (create may return invitation id and email, not the raw token).  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `invitations_create`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

Token hashed at rest, single-use, TTL 7 days. May reserve a seat depending on plan rules (v1: pending invitations count toward seat preview, not toward billed seats until accepted).

```json
{
  "email": "charles@example.com",
  "role": "sales_designer"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| email | string | yes | Invitee email; lowercase |
| role | string | yes | Any role except owner |

#### Success — HTTP 201

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "invitation": {
    "id": "inv_join01c7e2a5b9",
    "organization_id": "org_studio9d2e4c6a80",
    "email": "charles@example.com",
    "role": "sales_designer",
    "status": "pending",
    "invited_by_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "expires_at": "2026-09-08T00:00:00Z",
    "accepted_at": null
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| invitation | object | yes | Invitation metadata; field token is NEVER present |
| invitation.id | string | yes | inv_ |
| invitation.organization_id | string | yes | org_ |
| invitation.email | string | yes | Invitee |
| invitation.role | string | yes | Offered role |
| invitation.status | string | yes | pending |
| invitation.invited_by_user_id | string | yes | usr_ |
| invitation.created_at | string (date-time) | yes | UTC |
| invitation.expires_at | string (date-time) | yes | UTC |
| invitation.accepted_at | string (date-time) \| null | yes | Null until accepted |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `already_member` | Email already has an active membership |
| 409 | `invitation_pending` | A pending invitation for this email already exists |
| 409 | `seat_limit` | Accepting would exceed plan seats; upgrade or remove members |
| 422 | `role_not_assignable` | role=owner is invalid |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 47. `GET /v1/organizations/{id}/invitations`

**Group:** Members / invitations  
**Summary:** List invitations. Tokens never included.  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `invitations_list`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| cursor | string | no | Opaque cursor |
| limit | integer | no | 1–100; default 50 |
| status | string | no | pending \| accepted \| revoked \| expired; default pending |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "id": "inv_join01c7e2a5b9",
      "organization_id": "org_studio9d2e4c6a80",
      "email": "charles@example.com",
      "role": "sales_designer",
      "status": "pending",
      "invited_by_user_id": "usr_ada01f3e8c2b4a90",
      "created_at": "2026-09-01T00:00:00Z",
      "expires_at": "2026-09-08T00:00:00Z",
      "accepted_at": null
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Invitations without tokens |
| next_cursor | string \| null | yes | Pagination |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 48. `DELETE /v1/organizations/{id}/invitations/{invitationId}`

**Group:** Members / invitations  
**Summary:** Revoke a pending invitation.  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `invitations_revoke`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |
| invitationId | string | yes | inv_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "revoked",
  "id": "inv_join01c7e2a5b9"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | revoked |
| id | string | yes | inv_ |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `invitation_not_pending` | Already accepted, expired, or revoked |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 49. `POST /v1/organizations/{id}/invitations/{invitationId}/resend`

**Group:** Members / invitations  
**Summary:** Resend invitation email. Rotates the hashed token. Does not return the new token.  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `invitations_resend`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |
| invitationId | string | yes | inv_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 202

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "status": "resent",
  "id": "inv_join01c7e2a5b9",
  "expires_at": "2026-09-08T00:00:00Z"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| status | string | yes | resent |
| id | string | yes | inv_ |
| expires_at | string (date-time) | yes | New expiry (now+7d) |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `invitation_not_pending` | Cannot resend accepted/revoked/expired |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 50. `GET /v1/invitations/{token}`

**Group:** Members / invitations  
**Summary:** Public invitation preview. Token in the path is the secret from the email link (not inv_ id).  
**Auth:** public  
**Required roles:** none  
**Idempotent?** yes (GET; does not consume the token)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `invitations_preview`

Does not return organization_id (prevents id enumeration). Accept still requires bearer.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| token | string | yes | Opaque single-use secret from email; not the inv_ id |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "organization_name": "Lovelace Cabinets",
  "role": "sales_designer",
  "inviter_display_name": "Ada Lovelace",
  "email": "charles@example.com",
  "expires_at": "2026-09-08T00:00:00Z",
  "status": "pending"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| organization_name | string | yes | Public org name only; not org id |
| role | string | yes | Offered role |
| inviter_display_name | string | yes | Inviter display name |
| email | string | yes | Intended recipient (lets the UI show mismatch) |
| expires_at | string (date-time) | yes | UTC |
| status | string | yes | pending \| expired \| revoked \| accepted |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 404 | `not_found` | Token unknown — do not distinguish expired vs never-existed beyond status when found |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 51. `POST /v1/invitations/{token}/accept`

**Group:** Members / invitations  
**Summary:** Accept an invitation as the authenticated user. Email of the user must match the invitation email.  
**Auth:** bearer  
**Required roles:** any authenticated user with verified email matching the invitation  
**Idempotent?** yes, with Idempotency-Key (replay returns the membership)  
**Idempotency-Key:** required  
**Rate-limit bucket:** `invitations_accept`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| token | string | yes | Opaque email token |

#### Query parameters

_None._

#### Request JSON body

No request body. Identity from bearer token.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "membership": {
    "user_id": "usr_bb02e4d9c1a7f380",
    "organization_id": "org_studio9d2e4c6a80",
    "role": "sales_designer",
    "status": "active",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z"
  },
  "organization": {
    "id": "org_studio9d2e4c6a80",
    "name": "Lovelace Cabinets",
    "owner_user_id": "usr_ada01f3e8c2b4a90",
    "created_at": "2026-09-01T00:00:00Z",
    "updated_at": "2026-09-01T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| membership | object | yes | New active membership |
| organization | object | yes | Joined organization |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `email_unverified` | Verify email before accepting |
| 403 | `invitation_email_mismatch` | Authenticated email does not match invitation |
| 404 | `not_found` | Token unknown |
| 409 | `invitation_expired` | Token expired |
| 409 | `invitation_revoked` | Invitation revoked |
| 409 | `invitation_accepted` | Already accepted |
| 409 | `already_member` | User already in this org |
| 409 | `seat_limit` | No free seat |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

## 8.10 Entitlements

### 52. `GET /v1/organizations/{id}/entitlements`

**Group:** Entitlements  
**Summary:** Signed entitlement snapshot for the caller in this organization. Desktop caches and verifies the signature offline.  
**Auth:** bearer  
**Required roles:** any active member  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `entitlements_read`

Capabilities are the intersection of role, plan, and subscription_status (including grace). Client must treat signature as authoritative, not plan name strings.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| device_id | string | no | dev_ to bind into the snapshot; omitted → device_id null |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "entitlement_snapshot": {
    "user_id": "usr_ada01f3e8c2b4a90",
    "organization_id": "org_studio9d2e4c6a80",
    "plan": "trial",
    "subscription_status": "trialing",
    "capabilities": {
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": true,
      "engineering_handoff": true,
      "max_seats": 3,
      "max_devices": 3,
      "max_cloud_projects": 10
    },
    "issued_at": "2026-09-01T00:00:00Z",
    "refresh_after": "2026-09-01T12:00:00Z",
    "expires_at": "2026-09-02T00:00:00Z",
    "grace_until": "2026-09-08T00:00:00Z",
    "device_id": "dev_desk01a9c4e7f2",
    "schema_version": 1,
    "signature": "c2lnbmVkLWVudGl0bGVtZW50LXYyLXRyaWFs"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| entitlement_snapshot | object | yes | Signed snapshot; see Entitlement snapshot in Front matter |
| user_id | string | yes | usr_ |
| organization_id | string | yes | org_ |
| plan | string | yes | trial \| studio \| team \| enterprise |
| subscription_status | string | yes | trialing \| active \| past_due \| grace \| suspended \| canceled \| unpaid \| incomplete |
| capabilities | object | yes | Map of capability key → boolean or number |
| issued_at | string (date-time) | yes | Snapshot issue time |
| refresh_after | string (date-time) | yes | Client should refresh after this time while online |
| expires_at | string (date-time) | yes | Hard expiry of this snapshot |
| grace_until | string (date-time) \| null | yes | Offline/paid-action grace end; 7d policy |
| device_id | string \| null | yes | Bound device when issued at activation; otherwise null |
| schema_version | integer | yes | Currently 1 |
| signature | string | yes | Base64 server signature over canonical snapshot bytes |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 400 | `device_not_owned` | device_id is not an active device of the caller |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

## 8.11 Billing / subscription

### 53. `GET /v1/billing/plans`

**Group:** Billing / subscription  
**Summary:** Public-ish catalog of sellable plans. Auth optional but recommended. Prices from server, never hardcoded in the desktop.  
**Auth:** bearer  
**Required roles:** any authenticated user  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `billing_plans`

Local/dev adapter may return fake catalog amounts. Never include Stripe secret keys.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "id": "plan_studio_monthly",
      "plan": "studio",
      "interval": "month",
      "currency": "usd",
      "unit_amount_minor": 2900,
      "included_seats": 1,
      "max_devices_per_user": 3,
      "capabilities": {
        "cloud_projects": true,
        "organization_assets": true,
        "proposal_branding": true,
        "engineering_handoff": false
      },
      "trial_days": 14
    },
    {
      "id": "plan_team_monthly",
      "plan": "team",
      "interval": "month",
      "currency": "usd",
      "unit_amount_minor": 4900,
      "included_seats": 5,
      "max_devices_per_user": 3,
      "capabilities": {
        "cloud_projects": true,
        "organization_assets": true,
        "proposal_branding": true,
        "engineering_handoff": true
      },
      "trial_days": 14
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Sellable plans; enterprise may be omitted (sales-led) |
| items[].id | string | yes | plan_ opaque catalog id |
| items[].plan | string | yes | studio \| team \| enterprise |
| items[].interval | string | yes | month \| year |
| items[].currency | string | yes | ISO currency |
| items[].unit_amount_minor | integer | yes | Minor units (cents) |
| items[].included_seats | integer | yes | Seats included in the unit amount |
| items[].max_devices_per_user | integer | yes | Always 3 in v1 |
| items[].capabilities | object | yes | Boolean feature keys for this plan |
| items[].trial_days | integer | yes | 14 |
| next_cursor | string \| null | yes | Usually null |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 54. `GET /v1/organizations/{id}/subscription`

**Group:** Billing / subscription  
**Summary:** Normalized subscription for the organization.  
**Auth:** bearer  
**Required roles:** owner, admin, billing_admin  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `billing_subscription_read`

Checkout completion does not alone grant access; this document is sourced from verified provider state (BIL-002).

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "subscription": {
    "id": "sub_stripe01e5a2d7",
    "organization_id": "org_studio9d2e4c6a80",
    "plan_id": "plan_team_monthly",
    "plan": "team",
    "status": "trialing",
    "seats": 5,
    "current_period_start": "2026-09-01T00:00:00Z",
    "current_period_end": "2026-09-15T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z",
    "cancel_at_period_end": false,
    "canceled_at": null,
    "grace_until": null,
    "provider": "stripe",
    "provider_subscription_id": "sub_stripe_opaque"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| subscription | object | yes | Normalized subscription; never raw Stripe payload |
| subscription.id | string | yes | sub_ |
| subscription.organization_id | string | yes | org_ |
| subscription.plan_id | string | yes | plan_ catalog id |
| subscription.plan | string | yes | trial \| studio \| team \| enterprise |
| subscription.status | string | yes | trialing \| active \| past_due \| grace \| suspended \| canceled \| unpaid \| incomplete |
| subscription.seats | integer | yes | Purchased seats |
| subscription.current_period_start | string (date-time) | yes | UTC |
| subscription.current_period_end | string (date-time) | yes | UTC |
| subscription.trial_ends_at | string (date-time) \| null | yes | UTC |
| subscription.cancel_at_period_end | boolean | yes | True if cancel scheduled |
| subscription.canceled_at | string (date-time) \| null | yes | When cancel was requested |
| subscription.grace_until | string (date-time) \| null | yes | past_due grace end (7d policy) |
| subscription.provider | string | yes | stripe \| fake |
| subscription.provider_subscription_id | string \| null | yes | Opaque provider id; not a secret |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 55. `POST /v1/organizations/{id}/checkout-session`

**Group:** Billing / subscription  
**Summary:** Create a Stripe Checkout session (or fake local URL). Card data never touches this API.  
**Auth:** bearer  
**Required roles:** owner, billing_admin  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `billing_checkout`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

Email must be verified (AUT-001). Returns a URL to open in the system browser.

```json
{
  "plan_id": "plan_team_monthly",
  "success_url": "cabinetstudio://billing/success",
  "cancel_url": "cabinetstudio://billing/cancel"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| plan_id | string | yes | From GET /v1/billing/plans |
| success_url | string | yes | Allowlisted desktop/https URL |
| cancel_url | string | yes | Allowlisted desktop/https URL |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "checkout_session_id": "cs_01j6n8x4k2m9p7q3r5t8",
  "url": "https://checkout.stripe.example/c/pay/cs_01j6n8x4k2m9p7q3r5t8",
  "expires_at": "2026-09-01T01:00:00Z"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| checkout_session_id | string | yes | cs_ opaque id |
| url | string | yes | Hosted checkout URL |
| expires_at | string (date-time) | yes | Session expiry |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 403 | `email_unverified` | Verify email before billing changes |
| 404 | `plan_not_found` | plan_id not in catalog |
| 409 | `subscription_active` | Org already has an active/trialing paid subscription; use portal |
| 422 | `redirect_uri_not_allowed` | success_url or cancel_url not allowlisted |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 56. `POST /v1/organizations/{id}/portal-session`

**Group:** Billing / subscription  
**Summary:** Create a Stripe Customer Portal session for payment method and invoice management.  
**Auth:** bearer  
**Required roles:** owner, billing_admin  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `billing_portal`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

Requires an existing customer/subscription mapping.

```json
{
  "return_url": "cabinetstudio://billing/return"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| return_url | string | yes | Allowlisted URL |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "portal_session_id": "bps_01j6n8x4k2m9p7q3r5t8",
  "url": "https://billing.stripe.example/session/bps_01j6n8x4k2m9p7q3r5t8",
  "expires_at": "2026-09-01T01:00:00Z"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| portal_session_id | string | yes | bps_ opaque id |
| url | string | yes | Hosted portal URL |
| expires_at | string (date-time) | yes | Session expiry |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `billing_customer_missing` | No provider customer yet; use checkout-session |
| 422 | `redirect_uri_not_allowed` | return_url not allowlisted |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 57. `POST /v1/organizations/{id}/subscription/cancel`

**Group:** Billing / subscription  
**Summary:** Cancel subscription. v1 requires at_period_end=true (no immediate destructive cancel through this API).  
**Auth:** bearer  
**Required roles:** owner, billing_admin  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `billing_cancel`

Audit billing.subscription_cancel_requested via outbox.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

Paid capabilities remain until current_period_end, then status canceled.

```json
{
  "at_period_end": true
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| at_period_end | boolean | yes | Must be true in v1 |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "subscription": {
    "id": "sub_stripe01e5a2d7",
    "organization_id": "org_studio9d2e4c6a80",
    "plan_id": "plan_team_monthly",
    "plan": "team",
    "status": "active",
    "seats": 5,
    "current_period_start": "2026-09-01T00:00:00Z",
    "current_period_end": "2026-09-15T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z",
    "cancel_at_period_end": true,
    "canceled_at": "2026-09-01T00:15:00Z",
    "grace_until": null,
    "provider": "stripe",
    "provider_subscription_id": "sub_stripe_opaque"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| subscription | object | yes | Updated subscription; cancel_at_period_end true |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `already_canceled` | Already scheduled or fully canceled |
| 409 | `subscription_not_cancelable` | No cancelable subscription (e.g. already canceled) |
| 422 | `validation_error` | at_period_end must be true |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 58. `POST /v1/organizations/{id}/subscription/resume`

**Group:** Billing / subscription  
**Summary:** Undo a cancel-at-period-end while the period is still open.  
**Auth:** bearer  
**Required roles:** owner, billing_admin  
**Idempotent?** yes, with Idempotency-Key  
**Idempotency-Key:** required  
**Rate-limit bucket:** `billing_resume`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Idempotency-Key | yes | Client UUID. Replay window 24h. Same key+different body → 409 idempotency_conflict. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "subscription": {
    "id": "sub_stripe01e5a2d7",
    "organization_id": "org_studio9d2e4c6a80",
    "plan_id": "plan_team_monthly",
    "plan": "team",
    "status": "trialing",
    "seats": 5,
    "current_period_start": "2026-09-01T00:00:00Z",
    "current_period_end": "2026-09-15T00:00:00Z",
    "trial_ends_at": "2026-09-15T00:00:00Z",
    "cancel_at_period_end": false,
    "canceled_at": null,
    "grace_until": null,
    "provider": "stripe",
    "provider_subscription_id": "sub_stripe_opaque"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| subscription | object | yes | Resumed subscription |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 409 | `not_canceling` | cancel_at_period_end is not true |
| 409 | `subscription_not_resumable` | Period ended or status is canceled |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 59. `GET /v1/organizations/{id}/invoices`

**Group:** Billing / subscription  
**Summary:** List invoices for the organization (provider-referenced, locally stored summaries).  
**Auth:** bearer  
**Required roles:** owner, admin, billing_admin  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `billing_invoices_list`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| cursor | string | no | Opaque cursor |
| limit | integer | no | 1–100; default 50 |
| status | string | no | draft \| open \| paid \| void \| uncollectible |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "id": "invc_20260901a3b8",
      "organization_id": "org_studio9d2e4c6a80",
      "number": "INV-2026-00041",
      "status": "paid",
      "currency": "usd",
      "total_minor": 4900,
      "amount_due_minor": 0,
      "hosted_invoice_url": "https://billing.stripe.example/invoice/invc_20260901a3b8",
      "invoice_pdf_url": "https://billing.stripe.example/invoice/invc_20260901a3b8/pdf",
      "period_start": "2026-08-01T00:00:00Z",
      "period_end": "2026-09-01T00:00:00Z",
      "created_at": "2026-08-01T00:01:00Z",
      "paid_at": "2026-08-01T00:02:00Z"
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Invoice summaries |
| items[].id | string | yes | invc_ |
| items[].organization_id | string | yes | org_ |
| items[].number | string \| null | yes | Human invoice number |
| items[].status | string | yes | draft \| open \| paid \| void \| uncollectible |
| items[].currency | string | yes | ISO currency |
| items[].total_minor | integer | yes | Minor units |
| items[].amount_due_minor | integer | yes | Minor units |
| items[].hosted_invoice_url | string \| null | yes | Provider hosted page |
| items[].invoice_pdf_url | string \| null | yes | Provider PDF |
| items[].period_start | string (date-time) \| null | yes | UTC |
| items[].period_end | string (date-time) \| null | yes | UTC |
| items[].created_at | string (date-time) | yes | UTC |
| items[].paid_at | string (date-time) \| null | yes | UTC |
| next_cursor | string \| null | yes | Pagination |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 60. `GET /v1/organizations/{id}/invoices/{invoiceId}`

**Group:** Billing / subscription  
**Summary:** Get one invoice summary.  
**Auth:** bearer  
**Required roles:** owner, admin, billing_admin  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `billing_invoices_read`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |
| invoiceId | string | yes | invc_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "invoice": {
    "id": "invc_20260901a3b8",
    "organization_id": "org_studio9d2e4c6a80",
    "number": "INV-2026-00041",
    "status": "paid",
    "currency": "usd",
    "total_minor": 4900,
    "amount_due_minor": 0,
    "hosted_invoice_url": "https://billing.stripe.example/invoice/invc_20260901a3b8",
    "invoice_pdf_url": "https://billing.stripe.example/invoice/invc_20260901a3b8/pdf",
    "period_start": "2026-08-01T00:00:00Z",
    "period_end": "2026-09-01T00:00:00Z",
    "created_at": "2026-08-01T00:01:00Z",
    "paid_at": "2026-08-01T00:02:00Z"
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| invoice | object | yes | Invoice summary; never raw provider payload |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 61. `GET /v1/organizations/{id}/seats`

**Group:** Billing / subscription  
**Summary:** Seat usage versus purchased seats.  
**Auth:** bearer  
**Required roles:** owner, admin, billing_admin  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `billing_seats`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "purchased_seats": 5,
  "used_seats": 2,
  "pending_invitations": 1,
  "available_seats": 3,
  "members": [
    {
      "user_id": "usr_ada01f3e8c2b4a90",
      "role": "owner",
      "status": "active",
      "consumes_seat": true
    }
  ]
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| purchased_seats | integer | yes | From subscription; trial uses plan included_seats |
| used_seats | integer | yes | Active members who consume a seat |
| pending_invitations | integer | yes | Pending invites (informational; v1 do not consume billed seats) |
| available_seats | integer | yes | purchased - used |
| members | object[] | yes | Per-member seat consumption |
| members[].user_id | string | yes | usr_ |
| members[].role | string | yes | Role |
| members[].status | string | yes | active \| suspended |
| members[].consumes_seat | boolean | yes | v1: active members consume a seat; suspended do not |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

### 62. `POST /v1/organizations/{id}/seats/preview-downgrade`

**Group:** Billing / subscription  
**Summary:** Preview which members and capabilities would be affected by switching to a plan. Does not mutate.  
**Auth:** bearer  
**Required roles:** owner, billing_admin  
**Idempotent?** yes (preview has no side effects)  
**Idempotency-Key:** optional  
**Rate-limit bucket:** `billing_preview_downgrade`

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |
| Content-Type | yes | application/json |
| Idempotency-Key | no | Client UUID. Recommended on retries. Replay window 24h. |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

_None._

#### Request JSON body

A downgrade must not silently delete projects or assets (BIL-006). This preview only reports over-quota entities.

```json
{
  "plan_id": "plan_studio_monthly"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| plan_id | string | yes | Target catalog plan_id |

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "plan_id": "plan_studio_monthly",
  "plan": "studio",
  "purchased_seats": 5,
  "target_seats": 1,
  "seat_shortfall": 4,
  "members_over_target": [
    {
      "user_id": "usr_bb02e4d9c1a7f380",
      "role": "sales_designer",
      "status": "active"
    }
  ],
  "capability_loss": [
    "engineering_handoff"
  ],
  "blocked": true,
  "message": "Reduce active members to 1 before switching to Studio."
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| plan_id | string | yes | Target plan |
| plan | string | yes | studio \| team \| enterprise |
| purchased_seats | integer | yes | Current |
| target_seats | integer | yes | Target included seats |
| seat_shortfall | integer | yes | max(0, used - target) |
| members_over_target | object[] | yes | Members that would not fit; does not auto-remove them |
| capability_loss | string[] | yes | Capability keys that would become false |
| blocked | boolean | yes | True if the switch would be rejected until seats are reduced |
| message | string | yes | User-facing explanation |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 404 | `plan_not_found` | Unknown plan_id |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |
| 409 | `idempotency_conflict` | Idempotency-Key reused with a different request body |

### 63. `POST /v1/webhooks/billing`

**Group:** Billing / subscription  
**Summary:** Stripe (or fake adapter) billing webhook. Raw body. Signature required. Always 200 {received:true} after accept; 401 invalid_signature otherwise.  
**Auth:** webhook HMAC  
**Required roles:** n/a (provider)  
**Idempotent?** yes (provider event id dedupe; replay-safe)  
**Idempotency-Key:** n/a (dedupe by Stripe event id)  
**Rate-limit bucket:** `webhooks_billing`

Handler verifies signature, writes the event + outbox row in one commit, returns 200. Kafka is not invoked in-request. Duplicate event ids return 200 received=true without re-applying side effects (BIL-003).

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Stripe-Signature | yes | t=timestamp,v1=hmac_sha256; verified against endpoint secret |
| Content-Type | yes | application/json; charset=utf-8 |
| X-Request-Id | no | If omitted, derived from provider event id as req_wh_... |

#### Path parameters

_None._

#### Query parameters

_None._

#### Request JSON body

Treat body as raw bytes for HMAC. Do not parse then re-serialize before verifying. Fake local adapter uses the same header name with a dev secret.

```json
{
  "id": "evt_stripe_opaque",
  "type": "invoice.paid",
  "data": {
    "object": {
      "id": "in_opaque"
    }
  }
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | Provider event id (dedupe key) |
| type | string | yes | Provider type; normalized internally |
| data | object | yes | Provider payload; never echoed back or logged raw |

#### Success — HTTP 200

```json
{
  "received": true,
  "request_id": "req_wh_evt_stripe_opaque"
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| received | boolean | yes | Always true on accept (including duplicate event ids) |
| request_id | string | yes | req_ correlation |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `invalid_signature` | Missing or invalid Stripe-Signature |
| 400 | `invalid_payload` | Body not JSON or missing event id |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

## 8.12 Audit

### 64. `GET /v1/organizations/{id}/audit-events`

**Group:** Audit  
**Summary:** Query organization audit events. Append-only. No credentials, tokens, or provider payloads.  
**Auth:** bearer  
**Required roles:** owner, admin  
**Idempotent?** yes (GET)  
**Idempotency-Key:** n/a  
**Rate-limit bucket:** `audit_events`

Typical actions: auth.login, org.created, org.ownership_transferred, member.suspended, invitation.created, billing.subscription_cancel_requested, device.activated.

#### Request headers

| Header | Required | Notes |
| --- | --- | --- |
| Authorization | yes | Bearer {access_token} |
| X-Request-Id | no | Client correlation id. Echoed on the response. Generated as req_ if omitted. |
| Accept | no | application/json |

#### Path parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| id | string | yes | org_ |

#### Query parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| cursor | string | no | Opaque cursor |
| limit | integer | no | 1–100; default 50 |
| actor_id | string | no | Filter usr_ |
| action | string | no | Stable action key, e.g. member.role_changed |
| from | string (date-time) | no | Inclusive UTC lower bound |
| to | string (date-time) | no | Exclusive UTC upper bound |

#### Request JSON body

No request body.

#### Success — HTTP 200

```json
{
  "request_id": "req_01j6n8x4k2m9p7q3r5t8",
  "items": [
    {
      "id": "evt_audit01d6f1c8",
      "organization_id": "org_studio9d2e4c6a80",
      "actor_id": "usr_ada01f3e8c2b4a90",
      "actor_type": "user",
      "action": "member.role_changed",
      "resource_type": "membership",
      "resource_id": "usr_bb02e4d9c1a7f380",
      "created_at": "2026-09-01T00:00:00Z",
      "ip_country": "IN",
      "request_id": "req_01j6n8x4k2m9p7q3r5t8",
      "metadata": {
        "from_role": "viewer",
        "to_role": "engineer"
      }
    }
  ],
  "next_cursor": null
}
```

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| request_id | string | yes | Echoed/generated req_ correlation id |
| items | object[] | yes | Audit events, newest first |
| items[].id | string | yes | evt_ |
| items[].organization_id | string | yes | org_ |
| items[].actor_id | string \| null | yes | usr_ or null for system/provider |
| items[].actor_type | string | yes | user \| system \| webhook |
| items[].action | string | yes | Stable snake_case key |
| items[].resource_type | string | yes | organization \| membership \| invitation \| subscription \| device \| session \| ... |
| items[].resource_id | string \| null | yes | Opaque id of the resource |
| items[].created_at | string (date-time) | yes | UTC |
| items[].ip_country | string \| null | yes | Never raw IP |
| items[].request_id | string \| null | yes | Originating req_ |
| items[].metadata | object | yes | Redacted small JSON; never tokens, passwords, or Stripe payloads |
| next_cursor | string \| null | yes | Pagination |

#### Error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Missing, expired, or invalid access token |
| 403 | `forbidden` | Authenticated but lacks required role or capability |
| 403 | `tenant_mismatch` | Organization id is not in the caller's membership scope |
| 404 | `not_found` | Resource does not exist or is not visible to this principal |
| 422 | `validation_error` | Body or query failed schema/validation |
| 429 | `rate_limited` | Bucket exhausted; Retry-After present |
| 503 | `service_unavailable` | Dependency or process unavailable; safe to retry with backoff |

---

## 9. Implementation notes for Go + Gin

- Bind JSON with `disallow unknown fields` in production validation.
- Hash secrets (passwords argon2id; refresh tokens, email tokens, invitation tokens SHA-256/HMAC at rest).
- Put `request_id` middleware first; echo `X-Request-Id`.
- Organization scope: load membership by (`user_id`, `org_id`) on every org-scoped route; never trust a client-supplied tenant besides the path id.
- Stripe adapter behind an interface; local `fake` provider honors the same shapes.
- Outbox collection written in the same transaction as domain mutations; a publisher ships to Kafka afterwards.
- Contract tests should freeze these examples and error codes.

_End of contract. Unique method+path pairs documented: **64**._
