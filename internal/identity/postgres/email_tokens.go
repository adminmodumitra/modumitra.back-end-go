package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

type EmailTokenRepo struct {
	db *sql.DB
}

func NewEmailTokenRepo(db *sql.DB) *EmailTokenRepo { return &EmailTokenRepo{db: db} }

func (r *EmailTokenRepo) Insert(ctx context.Context, t *identity.EmailToken) error {
	q := `INSERT INTO email_tokens (public_id, user_id, purpose, token_hash, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		t.ID, t.UserID, t.Purpose, t.TokenHash, t.ExpiresAt, t.CreatedAt,
	)
	return mapSQL(err)
}

func (r *EmailTokenRepo) GetByHash(ctx context.Context, hash string) (*identity.EmailToken, error) {
	q := `SELECT public_id, user_id, purpose, token_hash, expires_at, consumed_at, created_at
		FROM email_tokens WHERE token_hash = $1`
	var t identity.EmailToken
	var consumed sql.NullTime
	err := conn(ctx, r.db).QueryRowContext(ctx, q, hash).Scan(
		&t.ID, &t.UserID, &t.Purpose, &t.TokenHash, &t.ExpiresAt, &consumed, &t.CreatedAt,
	)
	if err != nil {
		return nil, mapSQL(err)
	}
	t.ConsumedAt = nullTime(consumed)
	t.ExpiresAt = t.ExpiresAt.UTC()
	t.CreatedAt = t.CreatedAt.UTC()
	return &t, nil
}

func (r *EmailTokenRepo) Consume(ctx context.Context, id string, at time.Time) error {
	q := `UPDATE email_tokens SET consumed_at = $1 WHERE public_id = $2 AND consumed_at IS NULL`
	res, err := conn(ctx, r.db).ExecContext(ctx, q, at, id)
	if err != nil {
		return mapSQL(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func (r *EmailTokenRepo) InvalidateOpen(ctx context.Context, userID, purpose string, at time.Time) error {
	q := `UPDATE email_tokens SET consumed_at = $1 WHERE user_id = $2 AND purpose = $3 AND consumed_at IS NULL`
	_, err := conn(ctx, r.db).ExecContext(ctx, q, at, userID, purpose)
	return mapSQL(err)
}

var _ identity.EmailTokenRepo = (*EmailTokenRepo)(nil)
