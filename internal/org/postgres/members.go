package postgres

import (
	"context"
	"database/sql"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type MemberRepo struct {
	db *sql.DB
}

func NewMemberRepo(db *sql.DB) *MemberRepo { return &MemberRepo{db: db} }

const memberCols = `public_id, organization_id, user_id, role, state, invited_at, activated_at, suspended_at, removed_at, created_at, updated_at`

func scanMember(scanner interface{ Scan(dest ...any) error }) (*org.Membership, error) {
	var m org.Membership
	var invited, activated, suspended, removed sql.NullTime
	err := scanner.Scan(
		&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.Status,
		&invited, &activated, &suspended, &removed, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQL(err)
	}
	m.InvitedAt = nullTime(invited)
	m.ActivatedAt = nullTime(activated)
	m.SuspendedAt = nullTime(suspended)
	m.RemovedAt = nullTime(removed)
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return &m, nil
}

func (r *MemberRepo) Insert(ctx context.Context, m *org.Membership) error {
	q := `INSERT INTO memberships (
		public_id, organization_id, user_id, role, state, invited_at, activated_at,
		suspended_at, removed_at, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		m.ID, m.OrganizationID, m.UserID, m.Role, m.Status,
		m.InvitedAt, m.ActivatedAt, m.SuspendedAt, m.RemovedAt, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil && isUniqueViolation(err) {
		return utils.E(utils.CodeAlreadyMember)
	}
	return mapSQL(err)
}

func (r *MemberRepo) GetByOrgUser(ctx context.Context, orgID, userID string) (*org.Membership, error) {
	q := `SELECT ` + memberCols + ` FROM memberships WHERE organization_id = $1 AND user_id = $2`
	return scanMember(conn(ctx, r.db).QueryRowContext(ctx, q, orgID, userID))
}

func (r *MemberRepo) ListByUser(ctx context.Context, userID, cursor string, limit int) ([]org.MembershipItem, string, error) {
	ct, cid, err := org.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := `SELECT m.public_id, m.organization_id, m.user_id, m.role, m.state, m.invited_at, m.activated_at,
		m.suspended_at, m.removed_at, m.created_at, m.updated_at,
		o.public_id, o.name, o.owner_user_id, o.status, o.trial_ends_at, o.deleted_at, o.created_at, o.updated_at
		FROM memberships m
		JOIN organizations o ON o.public_id = m.organization_id
		WHERE m.user_id = $1
		  AND m.state IN ('active','suspended')
		  AND o.status <> 'deleted'`
	args := []any{userID}
	if cid != "" {
		q += ` AND (m.created_at, m.public_id) < ($2, $3)`
		args = append(args, ct, cid)
	}
	q += ` ORDER BY m.created_at DESC, m.public_id DESC LIMIT $` + itoa(len(args)+1)
	args = append(args, limit+1)
	rows, err := conn(ctx, r.db).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, "", mapSQL(err)
	}
	defer rows.Close()
	var items []org.MembershipItem
	for rows.Next() {
		var m org.Membership
		var o org.Organization
		var invited, activated, suspended, removed, trial, deleted sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.Status,
			&invited, &activated, &suspended, &removed, &m.CreatedAt, &m.UpdatedAt,
			&o.ID, &o.Name, &o.OwnerUserID, &o.Status, &trial, &deleted, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, "", mapSQL(err)
		}
		m.InvitedAt = nullTime(invited)
		m.ActivatedAt = nullTime(activated)
		m.SuspendedAt = nullTime(suspended)
		m.RemovedAt = nullTime(removed)
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		o.TrialEndsAt = nullTime(trial)
		o.DeletedAt = nullTime(deleted)
		o.CreatedAt = o.CreatedAt.UTC()
		o.UpdatedAt = o.UpdatedAt.UTC()
		items = append(items, org.MembershipItem{Organization: o, Membership: m})
	}
	if err := rows.Err(); err != nil {
		return nil, "", mapSQL(err)
	}
	next := ""
	if limit > 0 && len(items) > limit {
		last := items[limit-1]
		next = org.EncodeCursor(last.Membership.CreatedAt, last.Membership.ID)
		items = items[:limit]
	}
	return items, next, nil
}

func (r *MemberRepo) ListByOrg(ctx context.Context, orgID, status, role, cursor string, limit int) ([]org.Membership, string, error) {
	ct, cid, err := org.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := `SELECT ` + memberCols + ` FROM memberships WHERE organization_id = $1 AND state <> 'removed'`
	args := []any{orgID}
	n := 2
	if status != "" {
		q += ` AND state = $` + itoa(n)
		args = append(args, status)
		n++
	}
	if role != "" {
		q += ` AND role = $` + itoa(n)
		args = append(args, role)
		n++
	}
	if cid != "" {
		q += ` AND (created_at, public_id) < ($` + itoa(n) + `, $` + itoa(n+1) + `)`
		args = append(args, ct, cid)
		n += 2
	}
	q += ` ORDER BY created_at DESC, public_id DESC LIMIT $` + itoa(n)
	args = append(args, limit+1)
	rows, err := conn(ctx, r.db).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, "", mapSQL(err)
	}
	defer rows.Close()
	var out []org.Membership
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, "", err
		}
		out = append(out, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, "", mapSQL(err)
	}
	next := ""
	if limit > 0 && len(out) > limit {
		last := out[limit-1]
		next = org.EncodeCursor(last.CreatedAt, last.ID)
		out = out[:limit]
	}
	return out, next, nil
}

func (r *MemberRepo) Update(ctx context.Context, m *org.Membership) error {
	q := `UPDATE memberships SET role = $1, state = $2, invited_at = $3, activated_at = $4,
		suspended_at = $5, removed_at = $6, updated_at = $7 WHERE public_id = $8`
	res, err := conn(ctx, r.db).ExecContext(ctx, q,
		m.Role, m.Status, m.InvitedAt, m.ActivatedAt, m.SuspendedAt, m.RemovedAt, m.UpdatedAt, m.ID,
	)
	if err != nil {
		return mapSQL(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return org.ErrNotFound
	}
	return nil
}

func (r *MemberRepo) CountActiveOwners(ctx context.Context, orgID string) (int, error) {
	q := `SELECT COUNT(*) FROM memberships WHERE organization_id = $1 AND state = 'active' AND role = 'owner'`
	var n int
	err := conn(ctx, r.db).QueryRowContext(ctx, q, orgID).Scan(&n)
	if err != nil {
		return 0, mapSQL(err)
	}
	return n, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

var _ org.MemberRepo = (*MemberRepo)(nil)
