-- =============================================================================
-- Cabinet Studio PostgreSQL Schema v1
-- Canonical DDL for database cabinet_studio, schema public
-- Engine: PostgreSQL 16+
-- Baseline: 2026-09-01
-- Tables: 19 (no extras in v1)
-- API twin: API_CONTRACTS_V1.md (64 unique method+path pairs)
--
-- Apply:
--   CREATE DATABASE cabinet_studio;   -- once, as superuser, not inside this script
--   \c cabinet_studio
--   SET timezone = 'UTC';
--   \i schema-v1.sql
--
-- Conventions:
--   * public_id TEXT PRIMARY KEY with opaque prefixed ids (usr_ org_ …).
--     Never serial integers as public API ids. FKs reference public_id.
--   * TIMESTAMPTZ stored in UTC. DEFAULT now(). Connection timezone UTC.
--   * Enumerations are TEXT + named CHECK constraints (easier migrations than
--     PostgreSQL ENUM types). Choice is consistent across all 19 tables.
--   * Soft-delete tenant rows via status/state columns. Physical DELETE is
--     reserved for TTL-like cleanup jobs (tokens, oauth states, device codes,
--     idempotency_keys 24h, expired data_exports) and never for tenants.
--   * JSONB: plans.capabilities, entitlements.capabilities, audit_events.metadata,
--     outbox.payload, billing_events.summary, idempotency_keys.response_body.
--   * No PAN/CVC. Tokens stored as hashes only (SHA-256(pepper||raw) hex).
--   * Password: argon2id PHC string. Never plaintext.
-- =============================================================================

SET timezone = 'UTC';
SET client_min_messages = WARNING;

-- -----------------------------------------------------------------------------
-- 1. users
-- -----------------------------------------------------------------------------
CREATE TABLE users (
    public_id              TEXT        PRIMARY KEY,
    email                  TEXT        NOT NULL,
    email_normalized       TEXT        NOT NULL,
    email_verified_at      TIMESTAMPTZ,
    password_hash          TEXT,
    password_updated_at    TIMESTAMPTZ,
    display_name           TEXT        NOT NULL,
    terms_version          TEXT        NOT NULL,
    privacy_version        TEXT        NOT NULL,
    terms_accepted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    pending_email          TEXT,
    status                 TEXT        NOT NULL DEFAULT 'active',
    disabled_at            TIMESTAMPTZ,
    deletion_requested_at  TIMESTAMPTZ,
    last_login_at          TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_users_public_id CHECK (public_id ~ '^usr_'),
    CONSTRAINT chk_users_email_len CHECK (char_length(email) BETWEEN 3 AND 254),
    CONSTRAINT chk_users_email_normalized_len CHECK (char_length(email_normalized) BETWEEN 3 AND 254),
    CONSTRAINT chk_users_display_name_len CHECK (char_length(display_name) BETWEEN 1 AND 80),
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'disabled', 'pending_deletion')),
    CONSTRAINT uniq_users_email_normalized UNIQUE (email_normalized)
);

CREATE INDEX idx_users_status_created_at ON users (status, created_at DESC);
CREATE INDEX idx_users_pending_email ON users (pending_email) WHERE pending_email IS NOT NULL;

COMMENT ON TABLE users IS 'Human identity, credentials, legal acceptance, account lifecycle. One row per person. Not tenant-scoped. API user.id = public_id.';
COMMENT ON COLUMN users.public_id IS 'Opaque usr_… primary key; API user.id';
COMMENT ON COLUMN users.email_normalized IS 'lower(trim(email)); unique lookup key';
COMMENT ON COLUMN users.password_hash IS 'argon2id PHC; null if OAuth-only. Never plaintext.';
COMMENT ON COLUMN users.status IS 'active | disabled | pending_deletion. disabled ⇒ HTTP 403 account_suspended.';

-- -----------------------------------------------------------------------------
-- 2. plans (catalog; seeded; no outbound FKs)
-- -----------------------------------------------------------------------------
CREATE TABLE plans (
    public_id              TEXT        PRIMARY KEY,
    code                   TEXT        NOT NULL,
    name                   TEXT        NOT NULL,
    interval               TEXT        NOT NULL,
    currency               TEXT        NOT NULL DEFAULT 'usd',
    unit_amount_minor      INTEGER     NOT NULL DEFAULT 0,
    seats_included         INTEGER     NOT NULL,
    max_devices_per_user   INTEGER     NOT NULL DEFAULT 3,
    trial_days             INTEGER     NOT NULL DEFAULT 14,
    sellable               BOOLEAN     NOT NULL DEFAULT TRUE,
    active                 BOOLEAN     NOT NULL DEFAULT TRUE,
    capabilities           JSONB       NOT NULL,
    provider               TEXT,
    provider_price_id      TEXT,
    sort_order             INTEGER     NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_plans_public_id CHECK (public_id ~ '^plan_'),
    CONSTRAINT chk_plans_code CHECK (code IN ('trial', 'studio', 'team', 'enterprise')),
    CONSTRAINT chk_plans_interval CHECK (interval IN ('none', 'month', 'year')),
    CONSTRAINT chk_plans_currency CHECK (currency = lower(currency) AND char_length(currency) = 3),
    CONSTRAINT chk_plans_unit_amount CHECK (unit_amount_minor >= 0),
    CONSTRAINT chk_plans_seats CHECK (seats_included >= 0),
    CONSTRAINT chk_plans_max_devices CHECK (max_devices_per_user >= 1),
    CONSTRAINT chk_plans_trial_days CHECK (trial_days >= 0),
    CONSTRAINT chk_plans_provider CHECK (provider IS NULL OR provider IN ('stripe', 'fake')),
    CONSTRAINT uniq_plans_code_interval UNIQUE (code, interval)
);

CREATE INDEX idx_plans_catalog ON plans (sellable, active, sort_order);

COMMENT ON TABLE plans IS 'Server-side commercial catalog. Desktop never hard-codes prices (BIL-004). Seeded, rarely mutated.';
COMMENT ON COLUMN plans.capabilities IS 'JSONB: cloud_projects, organization_assets, proposal_branding, engineering_handoff (bool); max_seats, max_devices, max_cloud_projects (int).';
COMMENT ON COLUMN plans.seats_included IS 'API included_seats';

-- -----------------------------------------------------------------------------
-- 3. organizations
-- owner_user_id is DEFERRABLE so org+owner-membership can be inserted in one txn.
-- Cycle watch: owner_user_id → users; memberships → organizations. Create
-- organizations first, then memberships. Deferral lets a future "owner must
-- be a member" trigger/constraint check at COMMIT.
-- -----------------------------------------------------------------------------
CREATE TABLE organizations (
    public_id              TEXT        PRIMARY KEY,
    name                   TEXT        NOT NULL,
    owner_user_id          TEXT        NOT NULL,
    status                 TEXT        NOT NULL DEFAULT 'active',
    trial_ends_at          TIMESTAMPTZ,
    provider               TEXT,
    provider_customer_id   TEXT,
    deleted_at             TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_organizations_public_id CHECK (public_id ~ '^org_'),
    CONSTRAINT chk_organizations_name_len CHECK (char_length(name) BETWEEN 1 AND 120),
    CONSTRAINT chk_organizations_status CHECK (status IN ('active', 'pending_deletion', 'deleted')),
    CONSTRAINT chk_organizations_provider CHECK (provider IS NULL OR provider IN ('stripe', 'fake')),
    CONSTRAINT fk_organizations_owner_user_id
        FOREIGN KEY (owner_user_id) REFERENCES users (public_id)
        ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED
);

CREATE UNIQUE INDEX uniq_organizations_provider_customer_id
    ON organizations (provider_customer_id)
    WHERE provider_customer_id IS NOT NULL;
CREATE INDEX idx_organizations_owner_user_id ON organizations (owner_user_id);
CREATE INDEX idx_organizations_status_created_at ON organizations (status, created_at DESC);

COMMENT ON TABLE organizations IS 'Cabinet business, billing boundary, shared-data owner. Trial clock starts at create (created_at + 14d). Soft-delete via status=deleted.';
COMMENT ON COLUMN organizations.public_id IS 'API organization.id; tenant key organization_id on child tables.';
COMMENT ON COLUMN organizations.provider_customer_id IS 'Stripe cus_… / fake id. NOT a card. Unique when present.';
COMMENT ON COLUMN organizations.owner_user_id IS 'Denormalized from membership role=owner. FK users, ON DELETE RESTRICT, DEFERRABLE.';

-- -----------------------------------------------------------------------------
-- 4. oauth_identities
-- -----------------------------------------------------------------------------
CREATE TABLE oauth_identities (
    public_id                    TEXT        PRIMARY KEY,
    user_id                      TEXT        NOT NULL,
    provider                     TEXT        NOT NULL,
    provider_user_id             TEXT        NOT NULL,
    email                        TEXT,
    email_verified_by_provider   BOOLEAN     NOT NULL DEFAULT FALSE,
    linked_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_oauth_identities_public_id CHECK (public_id ~ '^oid_'),
    CONSTRAINT chk_oauth_identities_provider CHECK (provider IN ('google', 'facebook', 'x')),
    CONSTRAINT uniq_oauth_identities_provider_subject UNIQUE (provider, provider_user_id),
    CONSTRAINT uniq_oauth_identities_user_provider UNIQUE (user_id, provider),
    CONSTRAINT fk_oauth_identities_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE CASCADE
);

COMMENT ON TABLE oauth_identities IS 'Links a Cabinet Studio user to one social/SSO subject per provider. Stores no provider tokens. public_id oid_ is never returned by the HTTP API.';
COMMENT ON COLUMN oauth_identities.provider_user_id IS 'Provider subject; not a secret.';

-- -----------------------------------------------------------------------------
-- 5. memberships
-- -----------------------------------------------------------------------------
CREATE TABLE memberships (
    public_id          TEXT        PRIMARY KEY,
    organization_id    TEXT        NOT NULL,
    user_id            TEXT        NOT NULL,
    role               TEXT        NOT NULL,
    state              TEXT        NOT NULL DEFAULT 'active',
    invited_at         TIMESTAMPTZ,
    activated_at       TIMESTAMPTZ,
    suspended_at       TIMESTAMPTZ,
    removed_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_memberships_public_id CHECK (public_id ~ '^mbr_'),
    CONSTRAINT chk_memberships_role CHECK (role IN (
        'owner', 'admin', 'billing_admin', 'sales_designer',
        'engineer', 'production', 'viewer'
    )),
    CONSTRAINT chk_memberships_state CHECK (state IN ('invited', 'active', 'suspended', 'removed')),
    CONSTRAINT uniq_memberships_org_user UNIQUE (organization_id, user_id),
    CONSTRAINT fk_memberships_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (public_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_memberships_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_memberships_user_state ON memberships (user_id, state);
CREATE INDEX idx_memberships_org_state_role ON memberships (organization_id, state, role);

COMMENT ON TABLE memberships IS 'User ↔ organization relationship. Unique (organization_id, user_id) including removed rows (reuse the same row). API projects state as status.';
COMMENT ON COLUMN memberships.state IS 'invited | active | suspended | removed. API field name is status.';

-- -----------------------------------------------------------------------------
-- 6. invitations
-- -----------------------------------------------------------------------------
CREATE TABLE invitations (
    public_id             TEXT        PRIMARY KEY,
    organization_id       TEXT        NOT NULL,
    email                 TEXT        NOT NULL,
    role                  TEXT        NOT NULL,
    status                TEXT        NOT NULL DEFAULT 'pending',
    token_hash            TEXT        NOT NULL,
    invited_by_user_id    TEXT        NOT NULL,
    expires_at            TIMESTAMPTZ NOT NULL,
    accepted_at           TIMESTAMPTZ,
    accepted_by_user_id   TEXT,
    revoked_at            TIMESTAMPTZ,
    revoked_by_user_id    TEXT,
    resent_at             TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_invitations_public_id CHECK (public_id ~ '^inv_'),
    CONSTRAINT chk_invitations_email_len CHECK (char_length(email) BETWEEN 3 AND 254),
    CONSTRAINT chk_invitations_role CHECK (role IN (
        'admin', 'billing_admin', 'sales_designer',
        'engineer', 'production', 'viewer'
    )),
    CONSTRAINT chk_invitations_status CHECK (status IN ('pending', 'accepted', 'revoked', 'expired')),
    CONSTRAINT uniq_invitations_token_hash UNIQUE (token_hash),
    CONSTRAINT fk_invitations_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (public_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_invitations_invited_by_user_id
        FOREIGN KEY (invited_by_user_id) REFERENCES users (public_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_invitations_accepted_by_user_id
        FOREIGN KEY (accepted_by_user_id) REFERENCES users (public_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_invitations_revoked_by_user_id
        FOREIGN KEY (revoked_by_user_id) REFERENCES users (public_id)
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uniq_invitations_org_email_pending
    ON invitations (organization_id, email)
    WHERE status = 'pending';
CREATE INDEX idx_invitations_org_email_status ON invitations (organization_id, email, status);
CREATE INDEX idx_invitations_org_status_created_at ON invitations (organization_id, status, created_at DESC);
CREATE INDEX idx_invitations_invited_by_user_id ON invitations (invited_by_user_id);

COMMENT ON TABLE invitations IS 'Time-limited single-use offer to join an organization at a non-owner role. Stores token_hash only. Expired rows stay for audit (status flipped lazily or by sweeper).';
COMMENT ON COLUMN invitations.token_hash IS 'SHA-256(pepper||raw). Unique. Never raw token.';

-- -----------------------------------------------------------------------------
-- 7. devices
-- -----------------------------------------------------------------------------
CREATE TABLE devices (
    public_id            TEXT        PRIMARY KEY,
    user_id              TEXT        NOT NULL,
    name                 TEXT        NOT NULL,
    platform             TEXT        NOT NULL,
    status               TEXT        NOT NULL DEFAULT 'active',
    app_version          TEXT,
    last_seen_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deactivated_at       TIMESTAMPTZ,
    deactivated_reason   TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_devices_public_id CHECK (public_id ~ '^dev_'),
    CONSTRAINT chk_devices_name_len CHECK (char_length(name) BETWEEN 1 AND 80),
    CONSTRAINT chk_devices_platform CHECK (platform IN ('macos', 'windows', 'linux')),
    CONSTRAINT chk_devices_status CHECK (status IN ('active', 'deactivated')),
    CONSTRAINT chk_devices_deactivated_reason CHECK (
        deactivated_reason IS NULL
        OR deactivated_reason IN ('user', 'replaced', 'admin', 'account_disabled')
    ),
    CONSTRAINT fk_devices_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_devices_user_status ON devices (user_id, status);
CREATE INDEX idx_devices_user_last_seen ON devices (user_id, last_seen_at DESC);

COMMENT ON TABLE devices IS 'Activated Cabinet Studio installation. v1 hard limit: 3 active devices per user. No hardware fingerprints.';

-- -----------------------------------------------------------------------------
-- 8. sessions
-- -----------------------------------------------------------------------------
CREATE TABLE sessions (
    public_id                TEXT        PRIMARY KEY,
    user_id                  TEXT        NOT NULL,
    device_id                TEXT,
    refresh_token_hash       TEXT        NOT NULL,
    family_id                TEXT        NOT NULL,
    expires_at               TIMESTAMPTZ NOT NULL,
    revoked_at               TIMESTAMPTZ,
    reuse_detected_at        TIMESTAMPTZ,
    replaced_by_session_id   TEXT,
    device_name              TEXT,
    user_agent_summary       TEXT,
    ip_hash                  TEXT,
    ip_country               TEXT,
    last_seen_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_sessions_public_id CHECK (public_id ~ '^ses_'),
    CONSTRAINT chk_sessions_ip_country CHECK (ip_country IS NULL OR char_length(ip_country) = 2),
    CONSTRAINT uniq_sessions_refresh_token_hash UNIQUE (refresh_token_hash),
    CONSTRAINT fk_sessions_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_sessions_device_id
        FOREIGN KEY (device_id) REFERENCES devices (public_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_sessions_replaced_by_session_id
        FOREIGN KEY (replaced_by_session_id) REFERENCES sessions (public_id)
        ON DELETE SET NULL
);

CREATE INDEX idx_sessions_family_id ON sessions (family_id);
CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_user_active ON sessions (user_id, revoked_at, expires_at);
CREATE INDEX idx_sessions_device_id ON sessions (device_id) WHERE device_id IS NOT NULL;

COMMENT ON TABLE sessions IS 'Rotating refresh-token family. Access JWTs are NOT stored. Reuse of a spent refresh token sets reuse_detected_at and revokes the entire family. Do not TTL-drop at expires_at.';
COMMENT ON COLUMN sessions.refresh_token_hash IS 'Unique. Never raw rt_… token.';
COMMENT ON COLUMN sessions.ip_hash IS 'SHA-256(pepper||ip); never raw IP.';

-- -----------------------------------------------------------------------------
-- 9. email_tokens
-- Cleanup job (not Mongo TTL): DELETE WHERE expires_at < now().
-- -----------------------------------------------------------------------------
CREATE TABLE email_tokens (
    public_id            TEXT        PRIMARY KEY,
    user_id              TEXT        NOT NULL,
    purpose              TEXT        NOT NULL,
    token_hash           TEXT        NOT NULL,
    new_email            TEXT,
    expires_at           TIMESTAMPTZ NOT NULL,
    consumed_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    request_ip_country   TEXT,
    CONSTRAINT chk_email_tokens_public_id CHECK (public_id ~ '^tok_'),
    CONSTRAINT chk_email_tokens_purpose CHECK (purpose IN ('email_verify', 'password_reset', 'email_change', 'otp_login')),
    CONSTRAINT chk_email_tokens_new_email CHECK (purpose <> 'email_change' OR new_email IS NOT NULL),
    CONSTRAINT chk_email_tokens_ip_country CHECK (request_ip_country IS NULL OR char_length(request_ip_country) = 2),
    CONSTRAINT uniq_email_tokens_token_hash UNIQUE (token_hash),
    CONSTRAINT fk_email_tokens_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_email_tokens_expires_at ON email_tokens (expires_at);
CREATE INDEX idx_email_tokens_user_purpose_created_at ON email_tokens (user_id, purpose, created_at DESC);

COMMENT ON TABLE email_tokens IS 'Single-use hashed tokens for email verify, password reset, email change, otp login. Raw prefixes ev_/pr_/ec_ never stored; otp_login hashes include user id. Cleanup job deletes expired rows.';

-- -----------------------------------------------------------------------------
-- 10. oauth_states
-- Cleanup job: DELETE WHERE expires_at < now() (typical TTL 10 minutes).
-- -----------------------------------------------------------------------------
CREATE TABLE oauth_states (
    public_id            TEXT        PRIMARY KEY,
    state                TEXT        NOT NULL,
    provider             TEXT        NOT NULL,
    intent               TEXT        NOT NULL,
    redirect_uri         TEXT        NOT NULL,
    code_challenge       TEXT,
    code_verifier_hash   TEXT,
    user_id              TEXT,
    consumed_at          TIMESTAMPTZ,
    expires_at           TIMESTAMPTZ NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_oauth_states_public_id CHECK (public_id ~ '^st_'),
    CONSTRAINT chk_oauth_states_provider CHECK (provider IN ('google', 'facebook', 'x')),
    CONSTRAINT chk_oauth_states_intent CHECK (intent IN ('login', 'link')),
    CONSTRAINT uniq_oauth_states_state UNIQUE (state),
    CONSTRAINT fk_oauth_states_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_oauth_states_expires_at ON oauth_states (expires_at);

COMMENT ON TABLE oauth_states IS 'Short-lived CSRF/PKCE state for authorization-code start → callback. v1 stores state = public_id. Cleanup after 10 minutes.';
COMMENT ON COLUMN oauth_states.code_verifier_hash IS 'Hash of PKCE verifier; never raw verifier.';

-- -----------------------------------------------------------------------------
-- 11. device_codes
-- Cleanup job: DELETE WHERE expires_at < now() (typical TTL 600s).
-- -----------------------------------------------------------------------------
CREATE TABLE device_codes (
    public_id           TEXT        PRIMARY KEY,
    device_code         TEXT        NOT NULL,
    user_code           TEXT        NOT NULL,
    client_id           TEXT        NOT NULL DEFAULT 'cabinet-studio-desktop',
    provider            TEXT        NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'pending',
    user_id             TEXT,
    interval_seconds    INTEGER     NOT NULL DEFAULT 5,
    last_poll_at        TIMESTAMPTZ,
    authorized_at       TIMESTAMPTZ,
    consumed_at         TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_device_codes_public_id CHECK (public_id ~ '^dc_'),
    CONSTRAINT chk_device_codes_provider CHECK (provider IN ('google', 'facebook', 'x')),
    CONSTRAINT chk_device_codes_status CHECK (status IN ('pending', 'authorized', 'denied', 'consumed', 'expired')),
    CONSTRAINT chk_device_codes_interval CHECK (interval_seconds >= 1),
    CONSTRAINT uniq_device_codes_device_code UNIQUE (device_code),
    CONSTRAINT uniq_device_codes_user_code UNIQUE (user_code),
    CONSTRAINT fk_device_codes_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE SET NULL
);

CREATE INDEX idx_device_codes_expires_at ON device_codes (expires_at);

COMMENT ON TABLE device_codes IS 'RFC 8628-style device authorization. device_code equals public_id in v1. user_code is uppercase XXXX-XXXX. Cleanup after 600s.';

-- -----------------------------------------------------------------------------
-- 12. subscriptions  (one per organization — UNIQUE FK)
-- -----------------------------------------------------------------------------
CREATE TABLE subscriptions (
    public_id                   TEXT        PRIMARY KEY,
    organization_id             TEXT        NOT NULL,
    plan_id                     TEXT        NOT NULL,
    plan_code                   TEXT        NOT NULL,
    status                      TEXT        NOT NULL DEFAULT 'trialing',
    provider                    TEXT        NOT NULL DEFAULT 'fake',
    provider_subscription_id    TEXT,
    seats_included              INTEGER     NOT NULL,
    trial_ends_at               TIMESTAMPTZ,
    current_period_start        TIMESTAMPTZ NOT NULL,
    current_period_end          TIMESTAMPTZ NOT NULL,
    cancel_at_period_end        BOOLEAN     NOT NULL DEFAULT FALSE,
    canceled_at                 TIMESTAMPTZ,
    grace_until                 TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_subscriptions_public_id CHECK (public_id ~ '^sub_'),
    CONSTRAINT chk_subscriptions_plan_code CHECK (plan_code IN ('trial', 'studio', 'team', 'enterprise')),
    CONSTRAINT chk_subscriptions_status CHECK (status IN (
        'trialing', 'active', 'past_due', 'grace', 'suspended', 'canceled'
    )),
    CONSTRAINT chk_subscriptions_provider CHECK (provider IN ('stripe', 'fake')),
    CONSTRAINT chk_subscriptions_seats CHECK (seats_included >= 0),
    CONSTRAINT uniq_subscriptions_organization_id UNIQUE (organization_id),
    CONSTRAINT fk_subscriptions_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (public_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_subscriptions_plan_id
        FOREIGN KEY (plan_id) REFERENCES plans (public_id)
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uniq_subscriptions_provider_subscription_id
    ON subscriptions (provider_subscription_id)
    WHERE provider_subscription_id IS NOT NULL;
CREATE INDEX idx_subscriptions_status_period_end ON subscriptions (status, current_period_end);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions (plan_id);

COMMENT ON TABLE subscriptions IS 'Normalized commercial agreement: one organization, one subscription. No card data. Webhook-verified state is source of truth (BIL-002).';
COMMENT ON COLUMN subscriptions.status IS 'Normalized six-value set. Never persist raw Stripe unpaid/incomplete enums.';

-- -----------------------------------------------------------------------------
-- 13. invoices
-- -----------------------------------------------------------------------------
CREATE TABLE invoices (
    public_id              TEXT        PRIMARY KEY,
    organization_id        TEXT        NOT NULL,
    subscription_id        TEXT,
    provider               TEXT        NOT NULL,
    provider_invoice_id    TEXT        NOT NULL,
    number                 TEXT,
    status                 TEXT        NOT NULL,
    currency               TEXT        NOT NULL DEFAULT 'usd',
    total_minor            INTEGER     NOT NULL DEFAULT 0,
    amount_due_minor       INTEGER     NOT NULL DEFAULT 0,
    hosted_invoice_url     TEXT,
    invoice_pdf_url        TEXT,
    period_start           TIMESTAMPTZ,
    period_end             TIMESTAMPTZ,
    paid_at                TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_invoices_public_id CHECK (public_id ~ '^invc_'),
    CONSTRAINT chk_invoices_provider CHECK (provider IN ('stripe', 'fake')),
    CONSTRAINT chk_invoices_status CHECK (status IN ('draft', 'open', 'paid', 'void', 'uncollectible')),
    CONSTRAINT chk_invoices_currency CHECK (currency = lower(currency) AND char_length(currency) = 3),
    CONSTRAINT uniq_invoices_provider_invoice_id UNIQUE (provider_invoice_id),
    CONSTRAINT fk_invoices_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (public_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_invoices_subscription_id
        FOREIGN KEY (subscription_id) REFERENCES subscriptions (public_id)
        ON DELETE SET NULL
);

CREATE INDEX idx_invoices_org_created_at ON invoices (organization_id, created_at DESC);
CREATE INDEX idx_invoices_subscription_id ON invoices (subscription_id) WHERE subscription_id IS NOT NULL;

COMMENT ON TABLE invoices IS 'Local projection of provider invoices for the org billing UI. Provider remains system of record for tax/PDF. No PAN/CVC. Never store the raw Stripe invoice object.';

-- -----------------------------------------------------------------------------
-- 14. billing_events
-- organization_id is nullable: provider event may arrive before org mapping.
-- -----------------------------------------------------------------------------
CREATE TABLE billing_events (
    public_id            TEXT        PRIMARY KEY,
    provider             TEXT        NOT NULL,
    provider_event_id    TEXT        NOT NULL,
    type                 TEXT        NOT NULL,
    organization_id      TEXT,
    subscription_id      TEXT,
    status               TEXT        NOT NULL DEFAULT 'received',
    payload_digest       TEXT        NOT NULL,
    summary              JSONB       NOT NULL DEFAULT '{}'::jsonb,
    error                TEXT,
    processed_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_billing_events_public_id CHECK (public_id ~ '^bev_'),
    CONSTRAINT chk_billing_events_provider CHECK (provider IN ('stripe', 'fake')),
    CONSTRAINT chk_billing_events_status CHECK (status IN ('received', 'processed', 'ignored', 'failed')),
    CONSTRAINT uniq_billing_events_provider_event UNIQUE (provider, provider_event_id),
    CONSTRAINT fk_billing_events_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (public_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_billing_events_subscription_id
        FOREIGN KEY (subscription_id) REFERENCES subscriptions (public_id)
        ON DELETE SET NULL
);

CREATE INDEX idx_billing_events_org_created_at ON billing_events (organization_id, created_at DESC);
CREATE INDEX idx_billing_events_status_created_at ON billing_events (status, created_at);

COMMENT ON TABLE billing_events IS 'Provider webhook inbox and dedupe log (BIL-003). Unique (provider, provider_event_id) makes replays 200 no-ops. Payload is redacted/digested, never a raw card-bearing object.';
COMMENT ON COLUMN billing_events.payload_digest IS 'SHA-256 of raw body bytes. HMAC verified before insert.';
COMMENT ON COLUMN billing_events.summary IS 'JSONB redacted ids/amounts only.';

-- -----------------------------------------------------------------------------
-- 15. entitlements
-- -----------------------------------------------------------------------------
CREATE TABLE entitlements (
    public_id              TEXT        PRIMARY KEY,
    user_id                TEXT        NOT NULL,
    organization_id        TEXT,
    device_id              TEXT,
    plan_id                TEXT,
    plan                   TEXT        NOT NULL DEFAULT 'none',
    subscription_status    TEXT        NOT NULL DEFAULT 'none',
    capabilities           JSONB       NOT NULL,
    issued_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    refresh_after          TIMESTAMPTZ NOT NULL,
    expires_at             TIMESTAMPTZ NOT NULL,
    grace_until            TIMESTAMPTZ,
    schema_version         INTEGER     NOT NULL DEFAULT 1,
    key_id                 TEXT        NOT NULL,
    signature              TEXT        NOT NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_entitlements_public_id CHECK (public_id ~ '^ent_'),
    CONSTRAINT chk_entitlements_plan CHECK (plan IN ('none', 'trial', 'studio', 'team', 'enterprise')),
    CONSTRAINT chk_entitlements_subscription_status CHECK (subscription_status IN (
        'none', 'trialing', 'active', 'past_due', 'grace', 'suspended', 'canceled'
    )),
    CONSTRAINT chk_entitlements_schema_version CHECK (schema_version >= 1),
    CONSTRAINT fk_entitlements_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_entitlements_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (public_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_entitlements_device_id
        FOREIGN KEY (device_id) REFERENCES devices (public_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_entitlements_plan_id
        FOREIGN KEY (plan_id) REFERENCES plans (public_id)
        ON DELETE SET NULL
);

CREATE INDEX idx_entitlements_org_user_issued_at ON entitlements (organization_id, user_id, issued_at DESC);
CREATE INDEX idx_entitlements_user_issued_at ON entitlements (user_id, issued_at DESC);
CREATE INDEX idx_entitlements_device_issued_at ON entitlements (device_id, issued_at DESC) WHERE device_id IS NOT NULL;

COMMENT ON TABLE entitlements IS 'Issued, signed capability snapshots. Rows are immutable after insert (issue a new row to refresh). organization_id nullable when the user has no org (plan=none). public_id not in HTTP body.';
COMMENT ON COLUMN entitlements.capabilities IS 'JSONB role ∩ plan ∩ subscription. Mixed bool/int.';
COMMENT ON COLUMN entitlements.signature IS 'Base64 over canonical snapshot excluding signature.';

-- -----------------------------------------------------------------------------
-- 16. audit_events  (append-only)
-- organization_id and actor_user_id are nullable FKs.
-- API field actor_id maps to actor_user_id.
-- -----------------------------------------------------------------------------
CREATE TABLE audit_events (
    public_id          TEXT        PRIMARY KEY,
    organization_id    TEXT,
    actor_user_id      TEXT,
    actor_type         TEXT        NOT NULL DEFAULT 'user',
    action             TEXT        NOT NULL,
    resource_type      TEXT        NOT NULL,
    resource_id        TEXT,
    ip_country         TEXT,
    request_id         TEXT,
    metadata           JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_audit_events_public_id CHECK (public_id ~ '^evt_'),
    CONSTRAINT chk_audit_events_actor_type CHECK (actor_type IN ('user', 'system', 'webhook')),
    CONSTRAINT chk_audit_events_resource_type CHECK (resource_type IN (
        'organization', 'membership', 'invitation', 'subscription',
        'device', 'session', 'user', 'invoice', 'entitlement'
    )),
    CONSTRAINT chk_audit_events_ip_country CHECK (ip_country IS NULL OR char_length(ip_country) = 2),
    CONSTRAINT fk_audit_events_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (public_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_audit_events_actor_user_id
        FOREIGN KEY (actor_user_id) REFERENCES users (public_id)
        ON DELETE SET NULL
);

CREATE INDEX idx_audit_events_org_created_at ON audit_events (organization_id, created_at DESC);
CREATE INDEX idx_audit_events_action ON audit_events (action);
CREATE INDEX idx_audit_events_org_action_created_at ON audit_events (organization_id, action, created_at DESC);
CREATE INDEX idx_audit_events_org_actor_created_at ON audit_events (organization_id, actor_user_id, created_at DESC);

COMMENT ON TABLE audit_events IS 'Append-only security and commercial audit trail. No updates. No deletes from app code (legal hold / retention jobs only). Never credentials, tokens, raw IPs, or provider payloads.';
COMMENT ON COLUMN audit_events.actor_user_id IS 'API field actor_id (usr_… or null).';
COMMENT ON COLUMN audit_events.metadata IS 'JSONB small redacted object.';

REVOKE UPDATE, DELETE ON audit_events FROM PUBLIC;

-- -----------------------------------------------------------------------------
-- 17. outbox
-- No FK: partition_key is typically an org or user public_id string; the org
-- may be deleted after enqueue. Kafka is downstream, not source of truth.
-- -----------------------------------------------------------------------------
CREATE TABLE outbox (
    public_id        TEXT        PRIMARY KEY,
    event_id         TEXT        NOT NULL,
    topic            TEXT        NOT NULL,
    partition_key    TEXT        NOT NULL,
    payload          JSONB       NOT NULL,
    attempts         INTEGER     NOT NULL DEFAULT 0,
    last_error       TEXT,
    available_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_outbox_public_id CHECK (public_id ~ '^obx_'),
    CONSTRAINT chk_outbox_attempts CHECK (attempts >= 0),
    CONSTRAINT uniq_outbox_event_id UNIQUE (event_id)
);

CREATE INDEX idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;
CREATE INDEX idx_outbox_publisher_due ON outbox (available_at, published_at);

COMMENT ON TABLE outbox IS 'Transactional outbox. HTTP/webhook handlers insert a row in the SAME transaction as domain writes. A publisher worker sends to Kafka after commit and sets published_at. Kafka is not the source of truth.';
COMMENT ON COLUMN outbox.event_id IS 'Publish idempotency key; often equal to public_id.';
COMMENT ON COLUMN outbox.partition_key IS 'Usually organization public_id string. Not existence-checked.';
COMMENT ON COLUMN outbox.payload IS 'JSONB redacted domain event (public ids, no secrets).';

-- -----------------------------------------------------------------------------
-- 18. idempotency_keys
-- Cleanup job: DELETE WHERE created_at < now() - interval '24 hours'.
-- principal is usr_… when authenticated, else ip:<sha256> or wh:stripe.
-- No FK: principal is not always a user id.
-- -----------------------------------------------------------------------------
CREATE TABLE idempotency_keys (
    public_id          TEXT        PRIMARY KEY,
    principal          TEXT        NOT NULL,
    key                TEXT        NOT NULL,
    method             TEXT        NOT NULL,
    path               TEXT        NOT NULL,
    request_hash       TEXT        NOT NULL,
    response_status    INTEGER     NOT NULL,
    response_body      JSONB,
    in_progress        BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at       TIMESTAMPTZ,
    CONSTRAINT chk_idempotency_keys_public_id CHECK (public_id ~ '^idk_'),
    CONSTRAINT chk_idempotency_keys_method CHECK (method IN ('GET', 'POST', 'PATCH', 'PUT', 'DELETE')),
    CONSTRAINT uniq_idempotency_keys_principal_key UNIQUE (principal, key)
);

CREATE INDEX idx_idempotency_keys_created_at ON idempotency_keys (created_at);

COMMENT ON TABLE idempotency_keys IS '24-hour replay cache for retryable mutations. Same principal+key+body hash returns stored response. Same key + different body → 409 idempotency_conflict. Never persist raw refresh tokens in response_body.';

-- -----------------------------------------------------------------------------
-- 19. data_exports
-- Cleanup job: DELETE WHERE expires_at IS NOT NULL AND expires_at < now().
-- -----------------------------------------------------------------------------
CREATE TABLE data_exports (
    public_id      TEXT        PRIMARY KEY,
    user_id        TEXT        NOT NULL,
    status         TEXT        NOT NULL DEFAULT 'queued',
    scope          TEXT        NOT NULL DEFAULT 'personal',
    object_key     TEXT,
    bytes          BIGINT,
    error_code     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at   TIMESTAMPTZ,
    expires_at     TIMESTAMPTZ,
    CONSTRAINT chk_data_exports_public_id CHECK (public_id ~ '^exp_'),
    CONSTRAINT chk_data_exports_status CHECK (status IN ('queued', 'processing', 'ready', 'failed')),
    CONSTRAINT chk_data_exports_scope CHECK (scope = 'personal'),
    CONSTRAINT chk_data_exports_bytes CHECK (bytes IS NULL OR bytes >= 0),
    CONSTRAINT fk_data_exports_user_id
        FOREIGN KEY (user_id) REFERENCES users (public_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_data_exports_user_status_created_at ON data_exports (user_id, status, created_at DESC);
CREATE INDEX idx_data_exports_expires_at ON data_exports (expires_at) WHERE expires_at IS NOT NULL;
CREATE UNIQUE INDEX uniq_data_exports_user_in_progress
    ON data_exports (user_id)
    WHERE status IN ('queued', 'processing');

COMMENT ON TABLE data_exports IS 'Asynchronous personal-data export jobs (POST /v1/me/export). Archive bytes live in object storage; this table stores job state. Not org-owned projects.';

-- =============================================================================
-- Seed plans (insert-once; public_id values are stable)
-- GET /v1/billing/plans returns sellable=true AND active=true (studio + team).
-- Trial is used internally when creating an organization (not in public catalog).
-- Enterprise is sales-led and is NOT seeded.
-- =============================================================================
INSERT INTO plans (
    public_id, code, name, interval, currency, unit_amount_minor,
    seats_included, max_devices_per_user, trial_days, sellable, active,
    capabilities, provider, provider_price_id, sort_order,
    created_at, updated_at
) VALUES
(
    'plan_trial',
    'trial',
    'Trial',
    'none',
    'usd',
    0,
    3,
    3,
    14,
    FALSE,
    TRUE,
    '{
      "cloud_projects": true,
      "organization_assets": false,
      "proposal_branding": false,
      "engineering_handoff": false,
      "max_seats": 3,
      "max_devices": 3,
      "max_cloud_projects": 10
    }'::jsonb,
    NULL,
    NULL,
    0,
    TIMESTAMPTZ '2026-09-01 00:00:00+00',
    TIMESTAMPTZ '2026-09-01 00:00:00+00'
),
(
    'plan_studio_monthly',
    'studio',
    'Studio',
    'month',
    'usd',
    2900,
    1,
    3,
    14,
    TRUE,
    TRUE,
    '{
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": true,
      "engineering_handoff": false,
      "max_seats": 1,
      "max_devices": 3,
      "max_cloud_projects": 50
    }'::jsonb,
    'stripe',
    'price_studio_monthly_placeholder',
    10,
    TIMESTAMPTZ '2026-09-01 00:00:00+00',
    TIMESTAMPTZ '2026-09-01 00:00:00+00'
),
(
    'plan_team_monthly',
    'team',
    'Team',
    'month',
    'usd',
    4900,
    5,
    3,
    14,
    TRUE,
    TRUE,
    '{
      "cloud_projects": true,
      "organization_assets": true,
      "proposal_branding": true,
      "engineering_handoff": true,
      "max_seats": 5,
      "max_devices": 3,
      "max_cloud_projects": 500
    }'::jsonb,
    'stripe',
    'price_team_monthly_placeholder',
    20,
    TIMESTAMPTZ '2026-09-01 00:00:00+00',
    TIMESTAMPTZ '2026-09-01 00:00:00+00'
)
ON CONFLICT (public_id) DO NOTHING;

-- =============================================================================
-- Cleanup jobs (application / cron). Postgres has no Mongo-style TTL indexes.
-- Always check expires_at / consumed_at in application code first; jobs are
-- storage hygiene, not security invalidation.
--
--   DELETE FROM email_tokens     WHERE expires_at < now();
--   DELETE FROM oauth_states     WHERE expires_at < now();
--   DELETE FROM device_codes     WHERE expires_at < now();
--   DELETE FROM idempotency_keys WHERE created_at < now() - interval '24 hours';
--   DELETE FROM data_exports     WHERE expires_at IS NOT NULL AND expires_at < now();
--   -- sessions: reuse detection needs spent hashes. Sweeper may drop:
--   DELETE FROM sessions
--     WHERE expires_at < now() - interval '7 days' AND revoked_at IS NOT NULL;
-- =============================================================================
