package postgres

import (
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

// OrganizationModel maps schema-v1 organizations.
type OrganizationModel struct {
	PublicID    string     `gorm:"column:public_id;primaryKey"`
	Name        string     `gorm:"column:name"`
	OwnerUserID string     `gorm:"column:owner_user_id"`
	Status      string     `gorm:"column:status"`
	TrialEndsAt *time.Time `gorm:"column:trial_ends_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (OrganizationModel) TableName() string { return "organizations" }

func orgToModel(o *org.Organization) OrganizationModel {
	return OrganizationModel{
		PublicID:    o.ID,
		Name:        o.Name,
		OwnerUserID: o.OwnerUserID,
		Status:      o.Status,
		TrialEndsAt: utcPtr(o.TrialEndsAt),
		DeletedAt:   utcPtr(o.DeletedAt),
		CreatedAt:   o.CreatedAt.UTC(),
		UpdatedAt:   o.UpdatedAt.UTC(),
	}
}

func orgFromModel(m *OrganizationModel) *org.Organization {
	return &org.Organization{
		ID:          m.PublicID,
		Name:        m.Name,
		OwnerUserID: m.OwnerUserID,
		Status:      m.Status,
		TrialEndsAt: utcPtr(m.TrialEndsAt),
		DeletedAt:   utcPtr(m.DeletedAt),
		CreatedAt:   m.CreatedAt.UTC(),
		UpdatedAt:   m.UpdatedAt.UTC(),
	}
}

// MembershipModel maps schema-v1 memberships. Column "state" is API "status".
type MembershipModel struct {
	PublicID       string     `gorm:"column:public_id;primaryKey"`
	OrganizationID string     `gorm:"column:organization_id"`
	UserID         string     `gorm:"column:user_id"`
	Role           string     `gorm:"column:role"`
	State          string     `gorm:"column:state"`
	InvitedAt      *time.Time `gorm:"column:invited_at"`
	ActivatedAt    *time.Time `gorm:"column:activated_at"`
	SuspendedAt    *time.Time `gorm:"column:suspended_at"`
	RemovedAt      *time.Time `gorm:"column:removed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (MembershipModel) TableName() string { return "memberships" }

func memberToModel(m *org.Membership) MembershipModel {
	return MembershipModel{
		PublicID:       m.ID,
		OrganizationID: m.OrganizationID,
		UserID:         m.UserID,
		Role:           m.Role,
		State:          m.Status,
		InvitedAt:      utcPtr(m.InvitedAt),
		ActivatedAt:    utcPtr(m.ActivatedAt),
		SuspendedAt:    utcPtr(m.SuspendedAt),
		RemovedAt:      utcPtr(m.RemovedAt),
		CreatedAt:      m.CreatedAt.UTC(),
		UpdatedAt:      m.UpdatedAt.UTC(),
	}
}

func memberFromModel(m *MembershipModel) *org.Membership {
	return &org.Membership{
		ID:             m.PublicID,
		OrganizationID: m.OrganizationID,
		UserID:         m.UserID,
		Role:           m.Role,
		Status:         m.State,
		InvitedAt:      utcPtr(m.InvitedAt),
		ActivatedAt:    utcPtr(m.ActivatedAt),
		SuspendedAt:    utcPtr(m.SuspendedAt),
		RemovedAt:      utcPtr(m.RemovedAt),
		CreatedAt:      m.CreatedAt.UTC(),
		UpdatedAt:      m.UpdatedAt.UTC(),
	}
}

// InvitationModel maps schema-v1 invitations.
type InvitationModel struct {
	PublicID         string     `gorm:"column:public_id;primaryKey"`
	OrganizationID   string     `gorm:"column:organization_id"`
	Email            string     `gorm:"column:email"`
	Role             string     `gorm:"column:role"`
	Status           string     `gorm:"column:status"`
	TokenHash        string     `gorm:"column:token_hash"`
	InvitedByUserID  string     `gorm:"column:invited_by_user_id"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	AcceptedAt       *time.Time `gorm:"column:accepted_at"`
	AcceptedByUserID *string    `gorm:"column:accepted_by_user_id"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	RevokedByUserID  *string    `gorm:"column:revoked_by_user_id"`
	ResentAt         *time.Time `gorm:"column:resent_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (InvitationModel) TableName() string { return "invitations" }

func inviteToModel(inv *org.Invitation) InvitationModel {
	return InvitationModel{
		PublicID:         inv.ID,
		OrganizationID:   inv.OrganizationID,
		Email:            inv.Email,
		Role:             inv.Role,
		Status:           inv.Status,
		TokenHash:        inv.TokenHash,
		InvitedByUserID:  inv.InvitedByUserID,
		ExpiresAt:        inv.ExpiresAt.UTC(),
		AcceptedAt:       utcPtr(inv.AcceptedAt),
		AcceptedByUserID: inv.AcceptedByUserID,
		RevokedAt:        utcPtr(inv.RevokedAt),
		RevokedByUserID:  inv.RevokedByUserID,
		ResentAt:         utcPtr(inv.ResentAt),
		CreatedAt:        inv.CreatedAt.UTC(),
		UpdatedAt:        inv.UpdatedAt.UTC(),
	}
}

func inviteFromModel(m *InvitationModel) *org.Invitation {
	return &org.Invitation{
		ID:               m.PublicID,
		OrganizationID:   m.OrganizationID,
		Email:            m.Email,
		Role:             m.Role,
		Status:           m.Status,
		TokenHash:        m.TokenHash,
		InvitedByUserID:  m.InvitedByUserID,
		ExpiresAt:        m.ExpiresAt.UTC(),
		AcceptedAt:       utcPtr(m.AcceptedAt),
		AcceptedByUserID: m.AcceptedByUserID,
		RevokedAt:        utcPtr(m.RevokedAt),
		RevokedByUserID:  m.RevokedByUserID,
		ResentAt:         utcPtr(m.ResentAt),
		CreatedAt:        m.CreatedAt.UTC(),
		UpdatedAt:        m.UpdatedAt.UTC(),
	}
}
