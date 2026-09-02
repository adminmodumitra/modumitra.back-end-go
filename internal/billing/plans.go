package billing

import (
	"context"
)

func (s *Service) ListPlans(ctx context.Context) ([]Plan, error) {
	items, err := s.plans.ListSellable(ctx)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if items == nil {
		items = []Plan{}
	}
	return items, nil
}
