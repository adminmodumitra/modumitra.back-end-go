package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo { return &SessionRepo{db: db} }

const sessionCols = `public_id, user_id, device_id, refresh_token_hash, family_id,
	expires_at, revoked_at, reuse_detected_at, replaced_by_session_id, device_name,
	created_at, updated_at, last_seen_at`

func scanSession(row *sql.Row) (*identity.Session, error) {
	var s identity.Session
	var deviceID, replacedBy, deviceName sql.NullString
	var revoked, reuse sql.NullTime
	err := row.Scan(
		&s.ID, &s.UserID, &deviceID, &s.RefreshHash, &s.FamilyID,
		&s.ExpiresAt, &revoked, &reuse, &replacedBy, &deviceName,
		&s.CreatedAt, &s.UpdatedAt, &s.LastSeenAt,
	)
	if err != nil {
		return nil, mapSQL(err)
	}
	s.DeviceID = nullString(deviceID)
	s.ReplacedBy = nullString(replacedBy)
	if deviceName.Valid {
		s.DeviceName = deviceName.String
	}
	s.RevokedAt = nullTime(revoked)
	s.ReuseDetectedAt = nullTime(reuse)
	s.ExpiresAt = s.ExpiresAt.UTC()
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	s.LastSeenAt = s.LastSeenAt.UTC()
	return &s, nil
}

func (r *SessionRepo) GetByID(ctx context.Context, id string) (*identity.Session, error) {
	q := `SELECT ` + sessionCols + ` FROM sessions WHERE public_id = $1`
	return scanSession(conn(ctx, r.db).QueryRowContext(ctx, q, id))
}

func (r *SessionRepo) GetByRefreshHash(ctx context.Context, hash string) (*identity.Session, error) {
	q := `SELECT ` + sessionCols + ` FROM sessions WHERE refresh_token_hash = $1`
	return scanSession(conn(ctx, r.db).QueryRowContext(ctx, q, hash))
}

func (r *SessionRepo) GetByRefreshHashForUpdate(ctx context.Context, hash string) (*identity.Session, error) {
	q := `SELECT ` + sessionCols + ` FROM sessions WHERE refresh_token_hash = $1 FOR UPDATE`
	return scanSession(conn(ctx, r.db).QueryRowContext(ctx, q, hash))
}

func (r *SessionRepo) Insert(ctx context.Context, s *identity.Session) error {
	q := `INSERT INTO sessions (
		public_id, user_id, device_id, refresh_token_hash, family_id,
		expires_at, device_name, last_seen_at, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		s.ID, s.UserID, nil, s.RefreshHash, s.FamilyID,
		s.ExpiresAt, nullIfEmpty(s.DeviceName), s.LastSeenAt, s.CreatedAt, s.UpdatedAt,
	)
	return mapSQL(err)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *SessionRepo) RevokeAndReplace(ctx context.Context, oldID, newID string, at time.Time) error {
	q := `UPDATE sessions SET revoked_at = $1, replaced_by_session_id = $2, updated_at = $1
		WHERE public_id = $3`
	res, err := conn(ctx, r.db).ExecContext(ctx, q, at, newID, oldID)
	if err != nil {
		return mapSQL(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func (r *SessionRepo) MarkReuseAndRevokeFamily(ctx context.Context, sessionID, familyID string, at time.Time) error {
	q1 := `UPDATE sessions SET reuse_detected_at = $1, revoked_at = COALESCE(revoked_at, $1), updated_at = $1
		WHERE public_id = $2`
	if _, err := conn(ctx, r.db).ExecContext(ctx, q1, at, sessionID); err != nil {
		return mapSQL(err)
	}
	q2 := `UPDATE sessions SET revoked_at = $1, updated_at = $1
		WHERE family_id = $2 AND revoked_at IS NULL`
	if _, err := conn(ctx, r.db).ExecContext(ctx, q2, at, familyID); err != nil {
		return mapSQL(err)
	}
	return nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id string, at time.Time) error {
	q := `UPDATE sessions SET revoked_at = $1, updated_at = $1
		WHERE public_id = $2 AND revoked_at IS NULL`
	_, err := conn(ctx, r.db).ExecContext(ctx, q, at, id)
	return mapSQL(err)
}

func (r *SessionRepo) RevokeAllForUser(ctx context.Context, userID string, at time.Time) (int, error) {
	q := `UPDATE sessions SET revoked_at = $1, updated_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL`
	res, err := conn(ctx, r.db).ExecContext(ctx, q, at, userID)
	if err != nil {
		return 0, mapSQL(err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

var _ identity.SessionRepo = (*SessionRepo)(nil)
