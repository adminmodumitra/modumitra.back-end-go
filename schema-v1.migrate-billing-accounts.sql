-- =============================================================================
-- Migrate existing cabinet_studio installs: introduce billing_accounts.
-- schema-v1.sql remains the greenfield CREATE (new installs).
-- =============================================================================
-- For each existing subscriptions row with organization_id, insert a
-- billing_account (owner_type=organization, owner_id=that org), backfill
-- subscriptions.billing_account_id, then drop uniq_subscriptions_organization_id.
-- =============================================================================

CREATE TABLE IF NOT EXISTS billing_accounts (
    public_id              TEXT        PRIMARY KEY,
    owner_type             TEXT        NOT NULL,
    owner_id               TEXT        NOT NULL,
    provider               TEXT        NOT NULL DEFAULT 'fake',
    provider_customer_id   TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_billing_accounts_public_id CHECK (public_id ~ '^ba_'),
    CONSTRAINT chk_billing_accounts_owner_type CHECK (owner_type IN ('user', 'organization')),
    CONSTRAINT chk_billing_accounts_provider CHECK (provider IN ('stripe', 'fake')),
    CONSTRAINT uniq_billing_accounts_owner UNIQUE (owner_type, owner_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_billing_accounts_provider_customer_id
    ON billing_accounts (provider_customer_id)
    WHERE provider_customer_id IS NOT NULL;

ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS billing_account_id TEXT;

-- Backfill one organization-owned billing account per existing subscription.
INSERT INTO billing_accounts (
    public_id, owner_type, owner_id, provider, provider_customer_id, created_at, updated_at
)
SELECT
    'ba_' || substr(md5(s.organization_id || s.public_id), 1, 32),
    'organization',
    s.organization_id,
    COALESCE(o.provider, s.provider, 'fake'),
    o.provider_customer_id,
    s.created_at,
    s.updated_at
FROM subscriptions s
LEFT JOIN organizations o ON o.public_id = s.organization_id
WHERE s.organization_id IS NOT NULL
ON CONFLICT (owner_type, owner_id) DO NOTHING;

UPDATE subscriptions s
SET billing_account_id = ba.public_id
FROM billing_accounts ba
WHERE ba.owner_type = 'organization'
  AND ba.owner_id = s.organization_id
  AND s.billing_account_id IS NULL;

ALTER TABLE subscriptions
    ALTER COLUMN billing_account_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uniq_subscriptions_billing_account_id'
    ) THEN
        ALTER TABLE subscriptions
            ADD CONSTRAINT uniq_subscriptions_billing_account_id UNIQUE (billing_account_id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_subscriptions_billing_account_id'
    ) THEN
        ALTER TABLE subscriptions
            ADD CONSTRAINT fk_subscriptions_billing_account_id
            FOREIGN KEY (billing_account_id) REFERENCES billing_accounts (public_id)
            ON DELETE RESTRICT;
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uniq_subscriptions_organization_id'
    ) THEN
        ALTER TABLE subscriptions DROP CONSTRAINT uniq_subscriptions_organization_id;
    END IF;
END $$;

ALTER TABLE subscriptions ALTER COLUMN organization_id DROP NOT NULL;

ALTER TABLE invoices ADD COLUMN IF NOT EXISTS billing_account_id TEXT;
UPDATE invoices i
SET billing_account_id = s.billing_account_id
FROM subscriptions s
WHERE i.subscription_id = s.public_id
  AND i.billing_account_id IS NULL;
UPDATE invoices i
SET billing_account_id = ba.public_id
FROM billing_accounts ba
WHERE ba.owner_type = 'organization'
  AND ba.owner_id = i.organization_id
  AND i.billing_account_id IS NULL;
ALTER TABLE invoices ALTER COLUMN billing_account_id SET NOT NULL;
ALTER TABLE invoices ALTER COLUMN organization_id DROP NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_invoices_billing_account_id'
    ) THEN
        ALTER TABLE invoices
            ADD CONSTRAINT fk_invoices_billing_account_id
            FOREIGN KEY (billing_account_id) REFERENCES billing_accounts (public_id)
            ON DELETE RESTRICT;
    END IF;
END $$;

ALTER TABLE billing_events ADD COLUMN IF NOT EXISTS billing_account_id TEXT;
UPDATE billing_events e
SET billing_account_id = s.billing_account_id
FROM subscriptions s
WHERE e.subscription_id = s.public_id
  AND e.billing_account_id IS NULL;
UPDATE billing_events e
SET billing_account_id = ba.public_id
FROM billing_accounts ba
WHERE ba.owner_type = 'organization'
  AND ba.owner_id = e.organization_id
  AND e.billing_account_id IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_billing_events_billing_account_id'
    ) THEN
        ALTER TABLE billing_events
            ADD CONSTRAINT fk_billing_events_billing_account_id
            FOREIGN KEY (billing_account_id) REFERENCES billing_accounts (public_id)
            ON DELETE SET NULL;
    END IF;
END $$;
