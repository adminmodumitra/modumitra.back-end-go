package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

const userCols = `public_id, email, email_normalized, email_verified_at, password_hash,
	display_name, terms_version, privacy_version, status, last_login_at, created_at, updated_at`

func scanUser(row *sql.Row) (*identity.User, error) {
	var u identity.User
	var verified, lastLogin sql.NullTime
	var pw sql.NullString
	err := row.Scan(
		&u.ID, &u.Email, &u.EmailNormalized, &verified, &pw,
		&u.DisplayName, &u.TermsVersion, &u.PrivacyVersion, &u.Status,
		&lastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQL(err)
	}
	u.EmailVerifiedAt = nullTime(verified)
	u.LastLoginAt = nullTime(lastLogin)
	if pw.Valid {
		u.PasswordHash = pw.String
	}
	u.CreatedAt = u.CreatedAt.UTC()
	u.UpdatedAt = u.UpdatedAt.UTC()
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*identity.User, error) {
	q := `SELECT ` + userCols + ` FROM users WHERE public_id = $1`
	return scanUser(conn(ctx, r.db).QueryRowContext(ctx, q, id))
}

func (r *UserRepo) GetByEmailNormalized(ctx context.Context, email string) (*identity.User, error) {
	q := `SELECT ` + userCols + ` FROM users WHERE email_normalized = $1`
	return scanUser(conn(ctx, r.db).QueryRowContext(ctx, q, email))
}

func (r *UserRepo) Insert(ctx context.Context, u *identity.User) error {
	q := `INSERT INTO users (
		public_id, email, email_normalized, password_hash, password_updated_at,
		display_name, terms_version, privacy_version, terms_accepted_at, status,
		created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		u.ID, u.Email, u.EmailNormalized, u.PasswordHash, u.CreatedAt,
		u.DisplayName, u.TermsVersion, u.PrivacyVersion, u.CreatedAt, u.Status,
		u.CreatedAt, u.UpdatedAt,
	)
	if err != nil && isUniqueViolation(err) {
		return utils.E(utils.CodeEmailTaken)
	}
	return mapSQL(err)
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id, hash string, at time.Time) error {
	q := `UPDATE users SET password_hash = $1, password_updated_at = $2, updated_at = $2 WHERE public_id = $3`
	res, err := conn(ctx, r.db).ExecContext(ctx, q, hash, at, id)
	if err != nil {
		return mapSQL(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func (r *UserRepo) SetEmailVerified(ctx context.Context, id string, at time.Time) error {
	q := `UPDATE users SET email_verified_at = $1, updated_at = $1 WHERE public_id = $2`
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

func (r *UserRepo) SetLastLogin(ctx context.Context, id string, at time.Time) error {
	q := `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE public_id = $2`
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

var _ identity.UserRepo = (*UserRepo)(nil)
