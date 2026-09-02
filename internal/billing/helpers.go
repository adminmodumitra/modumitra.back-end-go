package billing

import (
	"context"
	"strings"
	"time"
)

func allowedRedirect(u string) bool {
	u = strings.TrimSpace(u)
	return strings.HasPrefix(u, "cabinetstudio://") || strings.HasPrefix(u, "https://")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	u := t.UTC()
	return &u
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *Service) accountIDFor(ctx context.Context, orgID, subID string) *string {
	if subID != "" {
		if sub, err := s.subs.GetByID(ctx, subID); err == nil && sub.BillingAccountID != "" {
			return strPtr(sub.BillingAccountID)
		}
	}
	if orgID != "" {
		if acc, err := s.accounts.GetByOwner(ctx, OwnerOrganization, orgID); err == nil {
			return strPtr(acc.ID)
		}
	}
	return nil
}
