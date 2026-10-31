-- =============================================================================
-- Migrate existing cabinet_studio installs: add org_price_books (engine
-- protection Phase 3). Idempotent; schema-v1.sql has the same table for new
-- installs. Apply: psql "$DATABASE_URL" -f schema-v1.migrate-org-price-books.sql
-- =============================================================================

CREATE TABLE IF NOT EXISTS org_price_books (
    org_id       TEXT        PRIMARY KEY,
    book         JSONB       NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by   TEXT        NOT NULL,
    CONSTRAINT chk_org_price_books_book_object CHECK (jsonb_typeof(book) = 'object'),
    CONSTRAINT fk_org_price_books_org_id
        FOREIGN KEY (org_id) REFERENCES organizations (public_id)
        ON DELETE CASCADE
);

COMMENT ON TABLE org_price_books IS 'One price book (rates JSON object) per organization. The api. engine proxy injects it into /v1/ops/quote and /v1/ops/estimate; no row = engine defaults. Written by owner/admin via PUT /v1/org/price-book.';
COMMENT ON COLUMN org_price_books.updated_by IS 'usr_… of the last writer. No FK: the row outlives a removed user.';
