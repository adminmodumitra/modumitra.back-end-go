package postgres

import (
	"context"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

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
	var m tables.Membership
	if err := use(ctx, r.db).Where("organization_id = ? AND user_id = ?", orgID, userID).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return memberFromModel(&m), nil
}

var _ org.MemberRepo = (*MemberRepo)(nil)
