package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

type OrgRepo struct {
	db *sql.DB
}

func NewOrgRepo(db *sql.DB) *OrgRepo { return &OrgRepo{db: db} }

const orgCols = `public_id, name, owner_user_id, status, trial_ends_at, deleted_at, created_at, updated_at`

func scanOrg(row *sql.Row) (*org.Organization, error) {
	var o org.Organization
	var trial, deleted sql.NullTime
	err := row.Scan(&o.ID, &o.Name, &o.OwnerUserID, &o.Status, &trial, &deleted, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, mapSQL(err)
	}
	o.TrialEndsAt = nullTime(trial)
	o.DeletedAt = nullTime(deleted)
	o.CreatedAt = o.CreatedAt.UTC()
	o.UpdatedAt = o.UpdatedAt.UTC()
	return &o, nil
}

func (r *OrgRepo) Insert(ctx context.Context, o *org.Organization) error {
	q := `INSERT INTO organizations (
		public_id, name, owner_user_id, status, trial_ends_at, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		o.ID, o.Name, o.OwnerUserID, o.Status, o.TrialEndsAt, o.CreatedAt, o.UpdatedAt,
	)
	return mapSQL(err)
}

func (r *OrgRepo) GetByID(ctx context.Context, id string) (*org.Organization, error) {
	q := `SELECT ` + orgCols + ` FROM organizations WHERE public_id = $1`
	return scanOrg(conn(ctx, r.db).QueryRowContext(ctx, q, id))
}

func (r *OrgRepo) UpdateName(ctx context.Context, id, name string, at time.Time) error {
	q := `UPDATE organizations SET name = $1, updated_at = $2 WHERE public_id = $3 AND status <> 'deleted'`
	res, err := conn(ctx, r.db).ExecContext(ctx, q, name, at, id)
	if err != nil {
		return mapSQL(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return org.ErrNotFound
	}
	return nil
}

func (r *OrgRepo) SoftDelete(ctx context.Context, id string, at time.Time) error {
	q := `UPDATE organizations SET status = 'deleted', deleted_at = $1, updated_at = $1 WHERE public_id = $2 AND status <> 'deleted'`
	res, err := conn(ctx, r.db).ExecContext(ctx, q, at, id)
	if err != nil {
		return mapSQL(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return org.ErrNotFound
	}
	return nil
}

func (r *OrgRepo) SetOwner(ctx context.Context, id, ownerUserID string, at time.Time) error {
	q := `UPDATE organizations SET owner_user_id = $1, updated_at = $2 WHERE public_id = $3 AND status <> 'deleted'`
	res, err := conn(ctx, r.db).ExecContext(ctx, q, ownerUserID, at, id)
	if err != nil {
		return mapSQL(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return org.ErrNotFound
	}
	return nil
}

var _ org.OrgRepo = (*OrgRepo)(nil)
