package billing

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Cancel(ctx context.Context, actorID, orgID string, atPeriodEnd bool) (*Subscription, error) {
	_, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	if !atPeriodEnd {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "at_period_end", "reason": "must be true"})
	}
	sub, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if sub.Status == StatusCanceled || sub.CancelAtPeriodEnd {
		return nil, utils.E(utils.CodeAlreadyCanceled)
	}
	if sub.Status != StatusActive && sub.Status != StatusTrialing {
		return nil, utils.E(utils.CodeSubscriptionNotCancelable)
	}
	if sub.ProviderSubscriptionID != "" {
		if err := s.provider.Cancel(ctx, sub.ProviderSubscriptionID, true); err != nil {
			return nil, mapRepoErr(err)
		}
	}
	now := s.clock.Now()
	sub.CancelAtPeriodEnd = true
	sub.CanceledAt = timePtr(now)
	sub.UpdatedAt = now
	if err := s.subs.Update(ctx, sub); err != nil {
		return nil, mapRepoErr(err)
	}
	return sub, nil
}

func (s *Service) Resume(ctx context.Context, actorID, orgID string) (*Subscription, error) {
	_, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	sub, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	if sub.Status == StatusCanceled || !sub.CurrentPeriodEnd.After(now) {
		return nil, utils.E(utils.CodeSubscriptionNotResumable)
	}
	if !sub.CancelAtPeriodEnd {
		return nil, utils.E(utils.CodeNotCanceling)
	}
	if sub.ProviderSubscriptionID != "" {
		if err := s.provider.Resume(ctx, sub.ProviderSubscriptionID); err != nil {
			return nil, mapRepoErr(err)
		}
	}
	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.UpdatedAt = now
	if err := s.subs.Update(ctx, sub); err != nil {
		return nil, mapRepoErr(err)
	}
	return sub, nil
}
