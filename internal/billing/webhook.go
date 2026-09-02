package billing

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) CompleteCheckout(ctx context.Context, sessionID string) error {
	fp, ok := s.provider.(*FakeProvider)
	if !ok {
		return utils.E(utils.CodeServiceUnavailable)
	}
	raw, err := fp.CompletedPayload(sessionID)
	if err != nil {
		return mapRepoErr(err)
	}
	return s.HandleWebhook(ctx, raw)
}

func (s *Service) HandleWebhook(ctx context.Context, raw []byte) error {
	var env struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object map[string]any `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil || strings.TrimSpace(env.ID) == "" || strings.TrimSpace(env.Type) == "" {
		return utils.E(utils.CodeInvalidPayload)
	}
	provider := s.provider.Name()
	if existing, err := s.events.GetByProviderEvent(ctx, provider, env.ID); err == nil && existing != nil {
		return nil
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return mapRepoErr(err)
	}
	now := s.clock.Now()
	summary := redactSummary(env.Type, env.Data.Object)
	evt := &Event{
		ID: s.ids.Event(), Provider: provider, ProviderEventID: env.ID,
		Type: env.Type, Status: EventReceived, PayloadDigest: PayloadDigest(raw),
		Summary: summary, CreatedAt: now,
	}
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.events.Insert(ctx, evt); err != nil {
			return err
		}
		orgID, subID, applyErr := s.applyEvent(ctx, env.Type, env.Data.Object, now)
		if applyErr != nil {
			_ = s.events.MarkFailed(ctx, evt.ID, applyErr.Error(), now)
			return applyErr
		}
		if env.Type == EventCheckoutCompleted && orgID != "" {
			planID := asString(env.Data.Object, "plan_id")
			status := StatusActive
			if subID != "" {
				if sub, gerr := s.subs.GetByID(ctx, subID); gerr == nil && sub != nil {
					planID = sub.PlanID
					status = sub.Status
				}
			}
			if err := outbox.Enqueue(ctx, s.outbox, outbox.BillingSubscriptionUpdated(ctx, orgID, subID, planID, status, now)); err != nil {
				return err
			}
		}
		return s.events.MarkProcessed(ctx, evt.ID, strPtr(orgID), strPtr(subID), s.accountIDFor(ctx, orgID, subID), now)
	})
	if errors.Is(err, ErrDuplicate) {
		return nil
	}
	return mapRepoErr(err)
}

func redactSummary(typ string, obj map[string]any) map[string]any {
	out := map[string]any{"type": typ}
	for _, k := range []string{"id", "organization_id", "billing_account_id", "plan_id", "customer", "subscription", "amount_total", "currency", "status"} {
		if v, ok := obj[k]; ok {
			out[k] = v
		}
	}
	return out
}

func asString(obj map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func asInt(obj map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case json.Number:
				i, _ := n.Int64()
				return int(i)
			}
		}
	}
	return 0
}

func (s *Service) applyEvent(ctx context.Context, typ string, obj map[string]any, now time.Time) (orgID, subID string, err error) {
	switch typ {
	case EventCheckoutCompleted:
		return s.applyCheckoutCompleted(ctx, obj, now)
	case EventInvoicePaid:
		return s.applyInvoicePaid(ctx, obj, now)
	default:
		return "", "", nil
	}
}
