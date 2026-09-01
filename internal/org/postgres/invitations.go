package postgres

import (
	"context"
	"database/sql"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type InvitationRepo struct {
	db *sql.DB
}

func NewInvitationRepo(db *sql.DB) *InvitationRepo { return &InvitationRepo{db: db} }

const inviteCols = `public_id, organization_id, email, role, status, token_hash, invited_by_user_id,
	expires_at, accepted_at, accepted_by_user_id, revoked_at, revoked_by_user_id, resent_at, created_at, updated_at`

func scanInvite(scanner interface{ Scan(dest ...any) error }) (*org.Invitation, error) {
	var inv org.Invitation
	var accepted, revoked, resent sql.NullTime
	var acceptedBy, revokedBy sql.NullString
	err := scanner.Scan(
		&inv.ID, &inv.OrganizationID, &inv.Email, &inv.Role, &inv.Status, &inv.TokenHash, &inv.InvitedByUserID,
		&inv.ExpiresAt, &accepted, &acceptedBy, &revoked, &revokedBy, &resent, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQL(err)
	}
	inv.AcceptedAt = nullTime(accepted)
	inv.AcceptedByUserID = nullString(acceptedBy)
	inv.RevokedAt = nullTime(revoked)
	inv.RevokedByUserID = nullString(revokedBy)
	inv.ResentAt = nullTime(resent)
	inv.ExpiresAt = inv.ExpiresAt.UTC()
	inv.CreatedAt = inv.CreatedAt.UTC()
	inv.UpdatedAt = inv.UpdatedAt.UTC()
	return &inv, nil
}

func (r *InvitationRepo) Insert(ctx context.Context, inv *org.Invitation) error {
	q := `INSERT INTO invitations (
		public_id, organization_id, email, role, status, token_hash, invited_by_user_id,
		expires_at, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		inv.ID, inv.OrganizationID, inv.Email, inv.Role, inv.Status, inv.TokenHash, inv.InvitedByUserID,
		inv.ExpiresAt, inv.CreatedAt, inv.UpdatedAt,
	)
	if err != nil && isUniqueViolation(err) {
		return utils.E(utils.CodeInvitationPending)
	}
	return mapSQL(err)
}

func (r *InvitationRepo) GetByID(ctx context.Context, id string) (*org.Invitation, error) {
	q := `SELECT ` + inviteCols + ` FROM invitations WHERE public_id = $1`
	return scanInvite(conn(ctx, r.db).QueryRowContext(ctx, q, id))
}

func (r *InvitationRepo) GetByTokenHash(ctx context.Context, hash string) (*org.Invitation, error) {
	q := `SELECT ` + inviteCols + ` FROM invitations WHERE token_hash = $1`
	return scanInvite(conn(ctx, r.db).QueryRowContext(ctx, q, hash))
}

func (r *InvitationRepo) GetPendingByOrgEmail(ctx context.Context, orgID, email string) (*org.Invitation, error) {
	q := `SELECT ` + inviteCols + ` FROM invitations WHERE organization_id = $1 AND email = $2 AND status = 'pending'`
	return scanInvite(conn(ctx, r.db).QueryRowContext(ctx, q, orgID, email))
}

func (r *InvitationRepo) ListByOrg(ctx context.Context, orgID, status, cursor string, limit int) ([]org.Invitation, string, error) {
	ct, cid, err := org.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := `SELECT ` + inviteCols + ` FROM invitations WHERE organization_id = $1`
	args := []any{orgID}
	n := 2
	if status != "" {
		q += ` AND status = $` + itoa(n)
		args = append(args, status)
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
	var out []org.Invitation
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, "", err
		}
		out = append(out, *inv)
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

func (r *InvitationRepo) Update(ctx context.Context, inv *org.Invitation) error {
	q := `UPDATE invitations SET status = $1, token_hash = $2, expires_at = $3, accepted_at = $4,
		accepted_by_user_id = $5, revoked_at = $6, revoked_by_user_id = $7, resent_at = $8, updated_at = $9
		WHERE public_id = $10`
	res, err := conn(ctx, r.db).ExecContext(ctx, q,
		inv.Status, inv.TokenHash, inv.ExpiresAt, inv.AcceptedAt, inv.AcceptedByUserID,
		inv.RevokedAt, inv.RevokedByUserID, inv.ResentAt, inv.UpdatedAt, inv.ID,
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

var _ org.InvitationRepo = (*InvitationRepo)(nil)
