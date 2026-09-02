package org

import (
	"context"
	"errors"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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
		if err := s.invites.Update(ctx, inv); err != nil {
			return err
		}
		if err := outbox.Enqueue(ctx, s.outbox, outbox.InvitationAccepted(ctx, inv.ID, o.ID, actorID, membership.Role, now)); err != nil {
			return err
		}
		return outbox.Enqueue(ctx, s.outbox, outbox.MembershipCreated(ctx, membership.ID, o.ID, actorID, membership.Role, now))
	})
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("organization_id", o.ID).Str("user_id", actorID).Msg("org.invitation_accepted")
	return o, membership, nil
}
