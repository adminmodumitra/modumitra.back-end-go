package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type MemberRepo struct {
	db *gorm.DB
}

func NewMemberRepo(db *gorm.DB) *MemberRepo { return &MemberRepo{db: db} }

type memberOrgRow struct {
	PublicID       string     `gorm:"column:public_id"`
	OrganizationID string     `gorm:"column:organization_id"`
	UserID         string     `gorm:"column:user_id"`
	Role           string     `gorm:"column:role"`
	State          string     `gorm:"column:state"`
	InvitedAt      *time.Time `gorm:"column:invited_at"`
	ActivatedAt    *time.Time `gorm:"column:activated_at"`
	SuspendedAt    *time.Time `gorm:"column:suspended_at"`
	RemovedAt      *time.Time `gorm:"column:removed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`

	OrgPublicID    string     `gorm:"column:org_public_id"`
	OrgName        string     `gorm:"column:org_name"`
	OrgOwnerUserID string     `gorm:"column:org_owner_user_id"`
	OrgStatus      string     `gorm:"column:org_status"`
	OrgTrialEndsAt *time.Time `gorm:"column:org_trial_ends_at"`
	OrgDeletedAt   *time.Time `gorm:"column:org_deleted_at"`
	OrgCreatedAt   time.Time  `gorm:"column:org_created_at"`
	OrgUpdatedAt   time.Time  `gorm:"column:org_updated_at"`
}

func (r *MemberRepo) Insert(ctx context.Context, m *org.Membership) error {
	row := memberToModel(m)
	if err := use(ctx, r.db).Create(&row).Error; err != nil {
		if isUnique(err) {
			return utils.E(utils.CodeAlreadyMember)
		}
		return mapGorm(err)
	}
	return nil
}

func (r *MemberRepo) GetByOrgUser(ctx context.Context, orgID, userID string) (*org.Membership, error) {
	var m MembershipModel
	if err := use(ctx, r.db).Where("organization_id = ? AND user_id = ?", orgID, userID).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return memberFromModel(&m), nil
}

func (r *MemberRepo) ListByUser(ctx context.Context, userID, cursor string, limit int) ([]org.MembershipItem, string, error) {
	ct, cid, err := org.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := use(ctx, r.db).Table("memberships").
		Select("memberships.public_id, memberships.organization_id, memberships.user_id, memberships.role, memberships.state, memberships.invited_at, memberships.activated_at, memberships.suspended_at, memberships.removed_at, memberships.created_at, memberships.updated_at, organizations.public_id AS org_public_id, organizations.name AS org_name, organizations.owner_user_id AS org_owner_user_id, organizations.status AS org_status, organizations.trial_ends_at AS org_trial_ends_at, organizations.deleted_at AS org_deleted_at, organizations.created_at AS org_created_at, organizations.updated_at AS org_updated_at").
		Joins("JOIN organizations ON organizations.public_id = memberships.organization_id").
		Where("memberships.user_id = ?", userID).
		Where("memberships.state IN ?", []string{org.MemberActive, org.MemberSuspended}).
		Where("organizations.status <> ?", org.OrgDeleted)
	if cid != "" {
		q = q.Where("(memberships.created_at, memberships.public_id) < (?, ?)", ct, cid)
	}
	var rows []memberOrgRow
	if err := q.Order("memberships.created_at DESC, memberships.public_id DESC").
		Limit(limit + 1).
		Find(&rows).Error; err != nil {
		return nil, "", mapGorm(err)
	}
	items := make([]org.MembershipItem, 0, len(rows))
	for i := range rows {
		row := rows[i]
		items = append(items, org.MembershipItem{
			Membership: *memberFromModel(&MembershipModel{
				PublicID:       row.PublicID,
				OrganizationID: row.OrganizationID,
				UserID:         row.UserID,
				Role:           row.Role,
				State:          row.State,
				InvitedAt:      row.InvitedAt,
				ActivatedAt:    row.ActivatedAt,
				SuspendedAt:    row.SuspendedAt,
				RemovedAt:      row.RemovedAt,
				CreatedAt:      row.CreatedAt,
				UpdatedAt:      row.UpdatedAt,
			}),
			Organization: *orgFromModel(&OrganizationModel{
				PublicID:    row.OrgPublicID,
				Name:        row.OrgName,
				OwnerUserID: row.OrgOwnerUserID,
				Status:      row.OrgStatus,
				TrialEndsAt: row.OrgTrialEndsAt,
				DeletedAt:   row.OrgDeletedAt,
				CreatedAt:   row.OrgCreatedAt,
				UpdatedAt:   row.OrgUpdatedAt,
			}),
		})
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
	q := use(ctx, r.db).Model(&MembershipModel{}).
		Where("organization_id = ? AND state <> ?", orgID, org.MemberRemoved)
	if status != "" {
		q = q.Where("state = ?", status)
	}
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if cid != "" {
		q = q.Where("(created_at, public_id) < (?, ?)", ct, cid)
	}
	var rows []MembershipModel
	if err := q.Order("created_at DESC, public_id DESC").Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, "", mapGorm(err)
	}
	out := make([]org.Membership, 0, len(rows))
	for i := range rows {
		out = append(out, *memberFromModel(&rows[i]))
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
	res := use(ctx, r.db).Model(&MembershipModel{}).Where("public_id = ?", m.ID).Updates(map[string]any{
		"role":         m.Role,
		"state":        m.Status,
		"invited_at":   m.InvitedAt,
		"activated_at": m.ActivatedAt,
		"suspended_at": m.SuspendedAt,
		"removed_at":   m.RemovedAt,
		"updated_at":   m.UpdatedAt,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return org.ErrNotFound
	}
	return nil
}

func (r *MemberRepo) CountActiveOwners(ctx context.Context, orgID string) (int, error) {
	var n int64
	if err := use(ctx, r.db).Model(&MembershipModel{}).
		Where("organization_id = ? AND state = ? AND role = ?", orgID, org.MemberActive, org.RoleOwner).
		Count(&n).Error; err != nil {
		return 0, mapGorm(err)
	}
	return int(n), nil
}

var _ org.MemberRepo = (*MemberRepo)(nil)
