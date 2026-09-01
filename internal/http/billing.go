package httpapi

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) requireBilling(c *gin.Context) bool {
	if h.billing == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

type checkoutBody struct {
	PlanID     string `json:"plan_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type portalBody struct {
	ReturnURL string `json:"return_url"`
}

type cancelBody struct {
	AtPeriodEnd *bool `json:"at_period_end"`
}

type previewBody struct {
	PlanID string `json:"plan_id"`
	Seats  *int   `json:"seats"`
}

func (h *handlers) listPlans(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	items, err := h.billing.ListPlans(c.Request.Context())
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]PlanView, 0, len(items))
	for i := range items {
		views = append(views, planView(&items[i]))
	}
	middleware.WriteJSON(c, utils.StatusOK, PlanListOK{
		Meta:       metaFrom(c),
		Items:      views,
		NextCursor: nil,
	})
}

func (h *handlers) getSubscription(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	sub, err := h.billing.GetSubscription(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}

func (h *handlers) checkoutSession(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body checkoutBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	sess, err := h.billing.Checkout(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), body.PlanID, body.SuccessURL, body.CancelURL)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, CheckoutOK{
		Meta:              metaFrom(c),
		CheckoutSessionID: sess.ID,
		URL:               sess.URL,
		ExpiresAt:         rfcTime(sess.ExpiresAt),
	})
}

func (h *handlers) portalSession(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body portalBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	sess, err := h.billing.Portal(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), body.ReturnURL)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, PortalOK{
		Meta:            metaFrom(c),
		PortalSessionID: sess.ID,
		URL:             sess.URL,
		ExpiresAt:       rfcTime(sess.ExpiresAt),
	})
}

func (h *handlers) cancelSubscription(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body cancelBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	atEnd := body.AtPeriodEnd != nil && *body.AtPeriodEnd
	sub, err := h.billing.Cancel(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), atEnd)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}

func (h *handlers) resumeSubscription(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	sub, err := h.billing.Resume(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}

func (h *handlers) listInvoices(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	limit, err := billing.ParseLimit(c.Query("limit"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	items, next, err := h.billing.ListInvoices(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Query("status"), c.Query("cursor"), limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]InvoiceView, 0, len(items))
	for i := range items {
		views = append(views, invoiceView(&items[i]))
	}
	middleware.WriteJSON(c, utils.StatusOK, InvoiceListOK{
		Meta:       metaFrom(c),
		Items:      views,
		NextCursor: cursorPtr(next),
	})
}

func (h *handlers) getInvoice(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	inv, err := h.billing.GetInvoice(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Param("invoiceId"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, InvoiceOK{
		Meta:    metaFrom(c),
		Invoice: invoiceView(inv),
	})
}

func (h *handlers) getSeats(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	seats, err := h.billing.Seats(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	members := make([]SeatMemberView, 0, len(seats.Members))
	for _, m := range seats.Members {
		members = append(members, SeatMemberView{
			UserID: m.UserID, Role: m.Role, Status: m.Status, ConsumesSeat: m.ConsumesSeat,
		})
	}
	middleware.WriteJSON(c, utils.StatusOK, SeatsOK{
		Meta:               metaFrom(c),
		PurchasedSeats:     seats.Purchased,
		UsedSeats:          seats.Used,
		PendingInvitations: seats.PendingInvitations,
		AvailableSeats:     seats.Available,
		Members:            members,
	})
}

func (h *handlers) previewDowngrade(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	var body previewBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	preview, err := h.billing.PreviewDowngrade(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), body.PlanID, body.Seats)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	over := make([]SeatMemberView, 0, len(preview.MembersOverTarget))
	for _, m := range preview.MembersOverTarget {
		over = append(over, SeatMemberView{UserID: m.UserID, Role: m.Role, Status: m.Status, ConsumesSeat: m.ConsumesSeat})
	}
	middleware.WriteJSON(c, utils.StatusOK, PreviewDowngradeOK{
		Meta:              metaFrom(c),
		PlanID:            preview.PlanID,
		Plan:              preview.Plan,
		PurchasedSeats:    preview.PurchasedSeats,
		TargetSeats:       preview.TargetSeats,
		SeatShortfall:     preview.SeatShortfall,
		MembersOverTarget: over,
		CapabilityLoss:    preview.CapabilityLoss,
		Blocked:           preview.Blocked,
		Message:           preview.Message,
	})
}

func (h *handlers) billingWebhook(c *gin.Context) {
	if !h.requireBilling(c) {
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		middleware.Abort(c, utils.E(utils.CodeInvalidPayload))
		return
	}
	secret := h.cfg.BillingWebhookSecret
	if err := billing.VerifyWebhookSignature(secret, raw, c.GetHeader(utils.HeaderStripeSignature), c.GetHeader(utils.HeaderBillingSignature)); err != nil {
		middleware.Abort(c, err)
		return
	}
	if err := h.billing.HandleWebhook(c.Request.Context(), raw); err != nil {
		middleware.Abort(c, err)
		return
	}
	rid := middleware.RequestIDFrom(c)
	if rid == "" {
		rid = "req_wh_billing"
	}
	middleware.WriteJSON(c, utils.StatusOK, WebhookOK{
		Meta:     utils.Meta{RequestID: rid, TraceID: middleware.TraceIDFrom(c)},
		Received: true,
	})
}

func planView(p *billing.Plan) PlanView {
	if p == nil {
		return PlanView{}
	}
	return PlanView{
		ID: p.ID, Plan: p.Code, Interval: p.Interval, Currency: p.Currency,
		UnitAmountMinor: p.UnitAmountMinor, IncludedSeats: p.SeatsIncluded,
		MaxDevicesPerUser: p.MaxDevicesPerUser,
		Capabilities: PlanCapabilitiesView{
			CloudProjects:      p.Capabilities.CloudProjects,
			OrganizationAssets: p.Capabilities.OrganizationAssets,
			ProposalBranding:   p.Capabilities.ProposalBranding,
			EngineeringHandoff: p.Capabilities.EngineeringHandoff,
		},
		TrialDays: p.TrialDays,
	}
}

func subscriptionView(s *billing.Subscription) SubscriptionView {
	if s == nil {
		return SubscriptionView{}
	}
	return SubscriptionView{
		ID: s.ID, OrganizationID: s.OrganizationID, PlanID: s.PlanID, Plan: s.PlanCode,
		Status: s.Status, Seats: s.SeatsIncluded,
		CurrentPeriodStart: rfcTime(s.CurrentPeriodStart), CurrentPeriodEnd: rfcTime(s.CurrentPeriodEnd),
		TrialEndsAt: rfcTimePtr(s.TrialEndsAt), CancelAtPeriodEnd: s.CancelAtPeriodEnd,
		CanceledAt: rfcTimePtr(s.CanceledAt), GraceUntil: rfcTimePtr(s.GraceUntil),
		Provider: s.Provider, ProviderSubscriptionID: emptyPtr(s.ProviderSubscriptionID),
	}
}

func invoiceView(inv *billing.Invoice) InvoiceView {
	if inv == nil {
		return InvoiceView{}
	}
	return InvoiceView{
		ID: inv.ID, OrganizationID: inv.OrganizationID, Number: inv.Number, Status: inv.Status,
		Currency: inv.Currency, TotalMinor: inv.TotalMinor, AmountDueMinor: inv.AmountDueMinor,
		HostedInvoiceURL: inv.HostedInvoiceURL, InvoicePDFURL: inv.InvoicePDFURL,
		PeriodStart: rfcTimePtr(inv.PeriodStart), PeriodEnd: rfcTimePtr(inv.PeriodEnd),
		CreatedAt: rfcTime(inv.CreatedAt), PaidAt: rfcTimePtr(inv.PaidAt),
	}
}

func emptyPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}
