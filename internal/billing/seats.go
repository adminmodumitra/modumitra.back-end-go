package billing

import (
	"context"
	"fmt"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Seats(ctx context.Context, actorID, orgID string) (*Seats, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireRead(m); err != nil {
		return nil, err
	}
	sub, err := s.getOrCreateTrial(ctx, o)
	if err != nil {
		return nil, err
	}
	return s.seatSnapshot(ctx, orgID, sub.SeatsIncluded)
}

func (s *Service) PreviewDowngrade(ctx context.Context, actorID, orgID, planID string, seats *int) (*Preview, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	sub, err := s.getOrCreateTrial(ctx, o)
	if err != nil {
		return nil, err
	}
	targetSeats := 0
	planCode := ""
	var targetCaps Capabilities
	if strings.TrimSpace(planID) != "" {
		plan, err := s.plans.GetByID(ctx, planID)
		if err != nil {
			return nil, utils.E(utils.CodePlanNotFound)
		}
		targetSeats = plan.SeatsIncluded
		planCode = plan.Code
		targetCaps = plan.Capabilities
	} else if seats != nil {
		if *seats < 0 {
			return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "seats"})
		}
		targetSeats = *seats
		planCode = sub.PlanCode
		planID = sub.PlanID
	} else {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "plan_id"})
	}
	snap, err := s.seatSnapshot(ctx, orgID, sub.SeatsIncluded)
	if err != nil {
		return nil, err
	}
	shortfall := snap.Used - targetSeats
	if shortfall < 0 {
		shortfall = 0
	}
	over := []SeatMember{}
	active := []SeatMember{}
	for _, mem := range snap.Members {
		if mem.ConsumesSeat {
			active = append(active, mem)
		}
	}
	if targetSeats < len(active) {
		over = append([]SeatMember{}, active[targetSeats:]...)
	}
	loss := []string{}
	cur, _ := s.plans.GetByID(ctx, sub.PlanID)
	if cur != nil {
		if cur.Capabilities.EngineeringHandoff && !targetCaps.EngineeringHandoff {
			loss = append(loss, "engineering_handoff")
		}
		if cur.Capabilities.ProposalBranding && !targetCaps.ProposalBranding {
			loss = append(loss, "proposal_branding")
		}
		if cur.Capabilities.OrganizationAssets && !targetCaps.OrganizationAssets {
			loss = append(loss, "organization_assets")
		}
		if cur.Capabilities.CloudProjects && !targetCaps.CloudProjects {
			loss = append(loss, "cloud_projects")
		}
	}
	if loss == nil {
		loss = []string{}
	}
	blocked := shortfall > 0
	msg := fmt.Sprintf("Switching to %d seats is allowed.", targetSeats)
	if blocked {
		msg = fmt.Sprintf("Reduce active members to %d before switching.", targetSeats)
	}
	preview := &Preview{
		PlanID: planID, Plan: planCode, PurchasedSeats: sub.SeatsIncluded,
		TargetSeats: targetSeats, SeatShortfall: shortfall,
		MembersOverTarget: over, CapabilityLoss: loss, Blocked: blocked, Message: msg,
	}
	if blocked {
		return nil, utils.E(utils.CodeSeatLimit).WithDetails(map[string]any{
			"purchased_seats":     preview.PurchasedSeats,
			"target_seats":        preview.TargetSeats,
			"used_seats":          snap.Used,
			"seat_shortfall":      preview.SeatShortfall,
			"members_over_target": preview.MembersOverTarget,
			"plan_id":             preview.PlanID,
			"plan":                preview.Plan,
			"capability_loss":     preview.CapabilityLoss,
			"message":             preview.Message,
		})
	}
	return preview, nil
}
