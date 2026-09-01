package org

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Service struct {
	orgs      OrgRepo
	members   MemberRepo
	invites   InvitationRepo
	users     identity.UserRepo
	tx        TxRunner
	mailer    Mailer
	clock     Clock
	ids       IDGen
	tokens    Tokens
	trialDays int
}

func NewService(
	orgs OrgRepo,
	members MemberRepo,
	invites InvitationRepo,
	users identity.UserRepo,
	tx TxRunner,
	mailer Mailer,
	clock Clock,
	ids IDGen,
	tokens Tokens,
	cfg Settings,
) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	if ids == nil {
		ids = NewIDGen()
	}
	if tokens == nil {
		tokens = NewTokens()
	}
	if cfg.TrialDays <= 0 {
		cfg.TrialDays = 14
	}
	return &Service{
		orgs:      orgs,
		members:   members,
		invites:   invites,
		users:     users,
		tx:        tx,
		mailer:    mailer,
		clock:     clock,
		ids:       ids,
		tokens:    tokens,
		trialDays: cfg.TrialDays,
	}
}

func NewMemoryService(users identity.UserRepo, trialDays int) (*Service, *RecordingMailer, *Memory) {
	mem := NewMemory()
	mail := &RecordingMailer{}
	svc := NewService(mem.Orgs(), mem.Members(), mem.Invites(), users, mem, mail, SystemClock{}, NewIDGen(), NewTokens(), Settings{TrialDays: trialDays})
	return svc, mail, mem
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validEmail(raw string) bool {
	s := strings.TrimSpace(raw)
	if len(s) < 3 || len(s) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address != ""
}

func validOrgName(s string) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(s))
	return n >= 1 && n <= 120
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) || errors.Is(err, identity.ErrNotFound) {
		return utils.E(utils.CodeNotFound)
	}
	return utils.From(err)
}

func (s *Service) loadOrg(ctx context.Context, id string) (*Organization, error) {
	o, err := s.orgs.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if o.Deleted() {
		return nil, utils.E(utils.CodeNotFound)
	}
	return o, nil
}

func (s *Service) loadActiveMember(ctx context.Context, orgID, userID string) (*Organization, *Membership, error) {
	o, err := s.loadOrg(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}
	m, err := s.members.GetByOrgUser(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeNotFound)
		}
		return nil, nil, mapRepoErr(err)
	}
	if !m.Active() {
		return nil, nil, utils.E(utils.CodeNotFound)
	}
	return o, m, nil
}

func requireRoles(m *Membership, roles ...string) error {
	if m == nil {
		return utils.E(utils.CodeNotFound)
	}
	for _, r := range roles {
		if m.Role == r {
			return nil
		}
	}
	return utils.E(utils.CodeForbidden)
}

func (s *Service) userByID(ctx context.Context, id string) (*identity.User, error) {
	if s.users == nil {
		return nil, utils.E(utils.CodeServiceUnavailable)
	}
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, identity.ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	return u, nil
}

// Resolve is used by Tenant middleware: active membership role, or not_found.
func (s *Service) Resolve(ctx context.Context, orgID, userID string) (string, error) {
	_, m, err := s.loadActiveMember(ctx, orgID, userID)
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (s *Service) PrimaryForUser(ctx context.Context, userID string) (*Organization, error) {
	if userID == "" {
		return nil, nil
	}
	items, _, err := s.members.ListByUser(ctx, userID, "", 1)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	for _, it := range items {
		if it.Membership.Active() && !it.Organization.Deleted() {
			o := it.Organization
			return &o, nil
		}
	}
	return nil, nil
}

func (s *Service) Create(ctx context.Context, actorID, name string) (*Organization, *Membership, error) {
	name = strings.TrimSpace(name)
	if actorID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	if !validOrgName(name) {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	u, err := s.userByID(ctx, actorID)
	if err != nil {
		return nil, nil, err
	}
	if !u.EmailVerified() {
		return nil, nil, utils.E(utils.CodeEmailUnverified)
	}
	now := s.clock.Now()
	trial := now.Add(time.Duration(s.trialDays) * 24 * time.Hour)
	o := &Organization{
		ID:          s.ids.Org(),
		Name:        name,
		OwnerUserID: actorID,
		Status:      OrgActive,
		TrialEndsAt: &trial,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m := &Membership{
		ID:             s.ids.Member(),
		OrganizationID: o.ID,
		UserID:         actorID,
		Role:           RoleOwner,
		Status:         MemberActive,
		ActivatedAt:    &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.orgs.Insert(ctx, o); err != nil {
			return err
		}
		return s.members.Insert(ctx, m)
	})
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.created", "organization_id", o.ID, "user_id", actorID)
	return o, m, nil
}

func (s *Service) List(ctx context.Context, actorID, cursor string, limit int) ([]MembershipItem, string, error) {
	if actorID == "" {
		return nil, "", utils.E(utils.CodeUnauthenticated)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		return nil, "", utils.E(utils.CodeValidation)
	}
	items, next, err := s.members.ListByUser(ctx, actorID, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	return items, next, nil
}

func (s *Service) Get(ctx context.Context, actorID, orgID string) (*Organization, error) {
	o, _, err := s.loadActiveMember(ctx, orgID, actorID)
	return o, err
}

func (s *Service) UpdateName(ctx context.Context, actorID, orgID, name string) (*Organization, error) {
	name = strings.TrimSpace(name)
	if !validOrgName(name) {
		return nil, utils.E(utils.CodeValidation)
	}
	o, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(m, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := s.orgs.UpdateName(ctx, orgID, name, now); err != nil {
		return nil, mapRepoErr(err)
	}
	o.Name = name
	o.UpdatedAt = now
	slog.InfoContext(ctx, "org.updated", "organization_id", orgID, "user_id", actorID)
	return o, nil
}

func (s *Service) Delete(ctx context.Context, actorID, orgID string) error {
	_, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if err := requireRoles(m, RoleOwner); err != nil {
		return err
	}
	now := s.clock.Now()
	if err := s.orgs.SoftDelete(ctx, orgID, now); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.deleted", "organization_id", orgID, "user_id", actorID)
	return nil
}

func (s *Service) Leave(ctx context.Context, actorID, orgID string) error {
	_, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if m.Role == RoleOwner {
		n, err := s.members.CountActiveOwners(ctx, orgID)
		if err != nil {
			return mapRepoErr(err)
		}
		if n <= 1 {
			return utils.E(utils.CodeLastOwner)
		}
	}
	now := s.clock.Now()
	m.Status = MemberRemoved
	m.RemovedAt = &now
	m.UpdatedAt = now
	if err := s.members.Update(ctx, m); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.left", "organization_id", orgID, "user_id", actorID)
	return nil
}

func (s *Service) TransferOwnership(ctx context.Context, actorID, orgID, targetUserID string) (*Organization, error) {
	if targetUserID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	o, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(m, RoleOwner); err != nil {
		return nil, err
	}
	if targetUserID == actorID {
		return nil, utils.E(utils.CodeAlreadyOwner)
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	if !target.Active() {
		return nil, utils.E(utils.CodeNotFound)
	}
	if target.Role == RoleOwner {
		return nil, utils.E(utils.CodeAlreadyOwner)
	}
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		m.Role = RoleAdmin
		m.UpdatedAt = now
		if err := s.members.Update(ctx, m); err != nil {
			return err
		}
		target.Role = RoleOwner
		target.UpdatedAt = now
		if err := s.members.Update(ctx, target); err != nil {
			return err
		}
		return s.orgs.SetOwner(ctx, orgID, targetUserID, now)
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	o.OwnerUserID = targetUserID
	o.UpdatedAt = now
	slog.InfoContext(ctx, "org.ownership_transferred", "organization_id", orgID, "from", actorID, "to", targetUserID)
	return o, nil
}

func (s *Service) Roles(ctx context.Context, actorID, orgID string) (map[string]map[string]bool, error) {
	if _, _, err := s.loadActiveMember(ctx, orgID, actorID); err != nil {
		return nil, err
	}
	return RolesCatalog(), nil
}

func (s *Service) ListMembers(ctx context.Context, actorID, orgID, status, role, cursor string, limit int) ([]Membership, string, error) {
	if _, _, err := s.loadActiveMember(ctx, orgID, actorID); err != nil {
		return nil, "", err
	}
	if status != "" && status != MemberActive && status != MemberSuspended {
		return nil, "", utils.E(utils.CodeValidation)
	}
	if role != "" && !ValidRole(role) {
		return nil, "", utils.E(utils.CodeValidation)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	items, next, err := s.members.ListByOrg(ctx, orgID, status, role, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	return items, next, nil
}

func (s *Service) PatchMember(ctx context.Context, actorID, orgID, targetUserID, role, status string) (*Membership, error) {
	if targetUserID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	if role == "" && status == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if target.Status == MemberRemoved {
		return nil, utils.E(utils.CodeNotFound)
	}
	if role != "" {
		if role == RoleOwner {
			return nil, utils.E(utils.CodeRoleNotAssignable)
		}
		if !AssignableRole(role) {
			return nil, utils.E(utils.CodeValidation)
		}
	}
	if actor.Role == RoleAdmin {
		if target.Role == RoleOwner || target.Role == RoleAdmin {
			return nil, utils.E(utils.CodeForbidden)
		}
		if role == RoleAdmin {
			return nil, utils.E(utils.CodeForbidden)
		}
	}
	if target.Role == RoleOwner && role != "" && role != RoleOwner {
		n, err := s.members.CountActiveOwners(ctx, orgID)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		if n <= 1 {
			return nil, utils.E(utils.CodeLastOwner)
		}
	}
	now := s.clock.Now()
	if role != "" {
		target.Role = role
	}
	if status == MemberActive || (status == "" && role != "" && target.Status == MemberSuspended) {
		target.Status = MemberActive
		target.SuspendedAt = nil
		if target.ActivatedAt == nil {
			target.ActivatedAt = &now
		}
	} else if status != "" && status != MemberActive && status != MemberSuspended {
		return nil, utils.E(utils.CodeValidation)
	}
	target.UpdatedAt = now
	if err := s.members.Update(ctx, target); err != nil {
		return nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.member_patched", "organization_id", orgID, "user_id", targetUserID, "actor", actorID)
	return target, nil
}

func (s *Service) SuspendMember(ctx context.Context, actorID, orgID, targetUserID string) (*Membership, error) {
	if targetUserID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	if targetUserID == actorID {
		return nil, utils.E(utils.CodeCannotSuspendSelf)
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if target.Status == MemberRemoved {
		return nil, utils.E(utils.CodeNotFound)
	}
	if target.Role == RoleOwner {
		return nil, utils.E(utils.CodeLastOwner)
	}
	if actor.Role == RoleAdmin && (target.Role == RoleAdmin || target.Role == RoleOwner) {
		return nil, utils.E(utils.CodeForbidden)
	}
	if target.Status == MemberSuspended {
		return target, nil
	}
	now := s.clock.Now()
	target.Status = MemberSuspended
	target.SuspendedAt = &now
	target.UpdatedAt = now
	if err := s.members.Update(ctx, target); err != nil {
		return nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.member_suspended", "organization_id", orgID, "user_id", targetUserID)
	return target, nil
}

func (s *Service) RemoveMember(ctx context.Context, actorID, orgID, targetUserID string) error {
	if targetUserID == "" {
		return utils.E(utils.CodeValidation)
	}
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return err
	}
	if targetUserID == actorID {
		return utils.E(utils.CodeCannotRemoveSelf)
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		return mapRepoErr(err)
	}
	if target.Status == MemberRemoved {
		return nil
	}
	if target.Role == RoleOwner {
		n, err := s.members.CountActiveOwners(ctx, orgID)
		if err != nil {
			return mapRepoErr(err)
		}
		if n <= 1 {
			return utils.E(utils.CodeLastOwner)
		}
	}
	if actor.Role == RoleAdmin && (target.Role == RoleAdmin || target.Role == RoleOwner) {
		return utils.E(utils.CodeForbidden)
	}
	now := s.clock.Now()
	target.Status = MemberRemoved
	target.RemovedAt = &now
	target.UpdatedAt = now
	if err := s.members.Update(ctx, target); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.member_removed", "organization_id", orgID, "user_id", targetUserID)
	return nil
}

func (s *Service) CreateInvitation(ctx context.Context, actorID, orgID, email, role string) (*Invitation, error) {
	email = normalizeEmail(email)
	if !validEmail(email) {
		return nil, utils.E(utils.CodeValidation)
	}
	if role == RoleOwner {
		return nil, utils.E(utils.CodeRoleNotAssignable)
	}
	if !AssignableRole(role) {
		return nil, utils.E(utils.CodeValidation)
	}
	o, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	if _, err := s.invites.GetPendingByOrgEmail(ctx, orgID, email); err == nil {
		return nil, utils.E(utils.CodeInvitationPending)
	} else if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	if s.users != nil {
		if u, err := s.users.GetByEmailNormalized(ctx, email); err == nil && u != nil {
			if mem, merr := s.members.GetByOrgUser(ctx, orgID, u.ID); merr == nil && mem.Active() {
				return nil, utils.E(utils.CodeAlreadyMember)
			}
		} else if err != nil && !errors.Is(err, identity.ErrNotFound) {
			return nil, mapRepoErr(err)
		}
	}
	now := s.clock.Now()
	raw, hash := s.tokens.NewInvite()
	inv := &Invitation{
		ID:              s.ids.Invitation(),
		OrganizationID:  orgID,
		Email:           email,
		Role:            role,
		Status:          InvitePending,
		TokenHash:       hash,
		InvitedByUserID: actorID,
		ExpiresAt:       now.Add(InviteTTL),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.invites.Insert(ctx, inv); err != nil {
		return nil, mapRepoErr(err)
	}
	if s.mailer != nil {
		if err := s.mailer.SendInvite(ctx, email, raw, o.Name); err != nil {
			slog.ErrorContext(ctx, "send invite mail", "organization_id", orgID, "error", err)
		}
	}
	slog.InfoContext(ctx, "org.invitation_created", "organization_id", orgID, "invitation_id", inv.ID)
	return inv, nil
}

func (s *Service) ListInvitations(ctx context.Context, actorID, orgID, status, cursor string, limit int) ([]Invitation, string, error) {
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, "", err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, "", err
	}
	if status == "" {
		status = InvitePending
	}
	switch status {
	case InvitePending, InviteAccepted, InviteRevoked, InviteExpired:
	default:
		return nil, "", utils.E(utils.CodeValidation)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	items, next, err := s.invites.ListByOrg(ctx, orgID, status, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	now := s.clock.Now()
	for i := range items {
		if items[i].Status == InvitePending && !items[i].ExpiresAt.After(now) {
			items[i].Status = InviteExpired
		}
	}
	return items, next, nil
}

func (s *Service) RevokeInvitation(ctx context.Context, actorID, orgID, invitationID string) error {
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return err
	}
	inv, err := s.invites.GetByID(ctx, invitationID)
	if err != nil {
		return mapRepoErr(err)
	}
	if inv.OrganizationID != orgID {
		return utils.E(utils.CodeNotFound)
	}
	now := s.clock.Now()
	if inv.Status != InvitePending || !inv.ExpiresAt.After(now) {
		return utils.E(utils.CodeInvitationNotPending)
	}
	inv.Status = InviteRevoked
	inv.RevokedAt = &now
	inv.RevokedByUserID = &actorID
	inv.UpdatedAt = now
	if err := s.invites.Update(ctx, inv); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.invitation_revoked", "invitation_id", invitationID)
	return nil
}

func (s *Service) ResendInvitation(ctx context.Context, actorID, orgID, invitationID string) (*Invitation, error) {
	o, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	inv, err := s.invites.GetByID(ctx, invitationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if inv.OrganizationID != orgID {
		return nil, utils.E(utils.CodeNotFound)
	}
	now := s.clock.Now()
	if inv.Status != InvitePending || !inv.ExpiresAt.After(now) {
		return nil, utils.E(utils.CodeInvitationNotPending)
	}
	raw, hash := s.tokens.NewInvite()
	inv.TokenHash = hash
	inv.ExpiresAt = now.Add(InviteTTL)
	inv.ResentAt = &now
	inv.UpdatedAt = now
	if err := s.invites.Update(ctx, inv); err != nil {
		return nil, mapRepoErr(err)
	}
	if s.mailer != nil {
		if err := s.mailer.SendInvite(ctx, inv.Email, raw, o.Name); err != nil {
			slog.ErrorContext(ctx, "resend invite mail", "invitation_id", invitationID, "error", err)
		}
	}
	slog.InfoContext(ctx, "org.invitation_resent", "invitation_id", invitationID)
	return inv, nil
}

func (s *Service) PreviewInvitation(ctx context.Context, rawToken string) (*InvitePreview, error) {
	inv, err := s.lookupPendingInvite(ctx, rawToken, false)
	if err != nil {
		return nil, err
	}
	o, err := s.orgs.GetByID(ctx, inv.OrganizationID)
	if err != nil || o.Deleted() {
		return nil, utils.E(utils.CodeNotFound)
	}
	inviterName := ""
	if u, uerr := s.userByID(ctx, inv.InvitedByUserID); uerr == nil && u != nil {
		inviterName = u.DisplayName
	}
	return &InvitePreview{
		OrganizationName:   o.Name,
		Role:               inv.Role,
		InviterDisplayName: inviterName,
		Email:              inv.Email,
		ExpiresAt:          inv.ExpiresAt,
		Status:             InvitePending,
	}, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, actorID, rawToken string) (*Organization, *Membership, error) {
	if actorID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	u, err := s.userByID(ctx, actorID)
	if err != nil {
		return nil, nil, err
	}
	if !u.EmailVerified() {
		return nil, nil, utils.E(utils.CodeEmailUnverified)
	}
	inv, err := s.lookupInviteForAccept(ctx, rawToken)
	if err != nil {
		return nil, nil, err
	}
	if normalizeEmail(u.Email) != normalizeEmail(inv.Email) {
		return nil, nil, utils.E(utils.CodeInvitationEmailMismatch)
	}
	o, err := s.loadOrg(ctx, inv.OrganizationID)
	if err != nil {
		return nil, nil, err
	}
	now := s.clock.Now()
	var membership *Membership
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		existing, gerr := s.members.GetByOrgUser(ctx, o.ID, actorID)
		if gerr != nil && !errors.Is(gerr, ErrNotFound) {
			return gerr
		}
		if existing != nil && existing.Active() {
			return utils.E(utils.CodeAlreadyMember)
		}
		if existing != nil {
			existing.Role = inv.Role
			existing.Status = MemberActive
			existing.ActivatedAt = &now
			existing.SuspendedAt = nil
			existing.RemovedAt = nil
			existing.UpdatedAt = now
			if err := s.members.Update(ctx, existing); err != nil {
				return err
			}
			membership = existing
		} else {
			membership = &Membership{
				ID:             s.ids.Member(),
				OrganizationID: o.ID,
				UserID:         actorID,
				Role:           inv.Role,
				Status:         MemberActive,
				InvitedAt:      &inv.CreatedAt,
				ActivatedAt:    &now,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := s.members.Insert(ctx, membership); err != nil {
				return err
			}
		}
		inv.Status = InviteAccepted
		inv.AcceptedAt = &now
		inv.AcceptedByUserID = &actorID
		inv.UpdatedAt = now
		return s.invites.Update(ctx, inv)
	})
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.invitation_accepted", "organization_id", o.ID, "user_id", actorID)
	return o, membership, nil
}

func (s *Service) lookupPendingInvite(ctx context.Context, rawToken string, forAccept bool) (*Invitation, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || !strings.HasPrefix(rawToken, PrefixInviteTok) {
		return nil, utils.E(utils.CodeNotFound)
	}
	inv, err := s.invites.GetByTokenHash(ctx, s.tokens.Hash(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	expired := !inv.ExpiresAt.After(now)
	if !forAccept {
		if inv.Status != InvitePending || expired {
			return nil, utils.E(utils.CodeNotFound)
		}
		return inv, nil
	}
	return inv, nil
}

func (s *Service) lookupInviteForAccept(ctx context.Context, rawToken string) (*Invitation, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || !strings.HasPrefix(rawToken, PrefixInviteTok) {
		return nil, utils.E(utils.CodeNotFound)
	}
	inv, err := s.invites.GetByTokenHash(ctx, s.tokens.Hash(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	switch {
	case inv.Status == InviteRevoked:
		return nil, utils.E(utils.CodeInvitationRevoked)
	case inv.Status == InviteAccepted:
		return nil, utils.E(utils.CodeInvitationAccepted)
	case inv.Status == InviteExpired || !inv.ExpiresAt.After(now):
		return nil, utils.E(utils.CodeInvitationExpired)
	case inv.Status != InvitePending:
		return nil, utils.E(utils.CodeNotFound)
	}
	return inv, nil
}

func (s *Service) MemberUser(ctx context.Context, userID string) (*identity.User, error) {
	return s.userByID(ctx, userID)
}
