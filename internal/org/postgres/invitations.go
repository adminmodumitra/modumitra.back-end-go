package postgres

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type InvitationRepo struct {
	db *gorm.DB
}

func NewInvitationRepo(db *gorm.DB) *InvitationRepo { return &InvitationRepo{db: db} }

func (r *InvitationRepo) Insert(ctx context.Context, inv *org.Invitation) error {
	m := inviteToModel(inv)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		if isUnique(err) {
			return utils.E(utils.CodeInvitationPending)
		}
		return mapGorm(err)
	}
	return nil
}

func (r *InvitationRepo) GetByID(ctx context.Context, id string) (*org.Invitation, error) {
	var m tables.Invitation
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return inviteFromModel(&m), nil
}

func (r *InvitationRepo) GetByTokenHash(ctx context.Context, hash string) (*org.Invitation, error) {
	var m tables.Invitation
	if err := use(ctx, r.db).Where("token_hash = ?", hash).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return inviteFromModel(&m), nil
}

func (r *InvitationRepo) GetPendingByOrgEmail(ctx context.Context, orgID, email string) (*org.Invitation, error) {
	var m tables.Invitation
	if err := use(ctx, r.db).
		Where("organization_id = ? AND email = ? AND status = ?", orgID, email, org.InvitePending).
		Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return inviteFromModel(&m), nil
}

func (r *InvitationRepo) ListByOrg(ctx context.Context, orgID, status, cursor string, limit int) ([]org.Invitation, string, error) {
	ct, cid, err := org.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := use(ctx, r.db).Model(&tables.Invitation{}).Where("organization_id = ?", orgID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if cid != "" {
		q = q.Where("(created_at, public_id) < (?, ?)", ct, cid)
	}
	var rows []tables.Invitation
	if err := q.Order("created_at DESC, public_id DESC").Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, "", mapGorm(err)
	}
	out := make([]org.Invitation, 0, len(rows))
	for i := range rows {
		out = append(out, *inviteFromModel(&rows[i]))
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
	res := use(ctx, r.db).Model(&tables.Invitation{}).Where("public_id = ?", inv.ID).Updates(map[string]any{
		"status":              inv.Status,
		"token_hash":          inv.TokenHash,
		"expires_at":          inv.ExpiresAt,
		"accepted_at":         inv.AcceptedAt,
		"accepted_by_user_id": inv.AcceptedByUserID,
		"revoked_at":          inv.RevokedAt,
		"revoked_by_user_id":  inv.RevokedByUserID,
		"resent_at":           inv.ResentAt,
		"updated_at":          inv.UpdatedAt,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return org.ErrNotFound
	}
	return nil
}

var _ org.InvitationRepo = (*InvitationRepo)(nil)
