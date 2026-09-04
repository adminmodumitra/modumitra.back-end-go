package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// createOrg godoc
// @Summary Create organization
// @Tags organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param body body dto.CreateOrgBody true "Org"
// @Success 201 {object} dto.OrgCreatedOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations [post]
func (h *handlers) createOrg(c *gin.Context) {
	logEnter(c, "create_org")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.CreateOrgBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	o, m, err := h.orgs.Create(c.Request.Context(), middleware.UserIDFrom(c), body.Name)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	trial := ""
	if o.TrialEndsAt != nil {
		trial = rfcTime(*o.TrialEndsAt)
	}
	middleware.WriteJSON(c, utils.StatusCreated, dto.OrgCreatedOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
		Membership:   membershipView(m, nil),
		TrialEndsAt:  trial,
	})
}

// listOrgs godoc
// @Summary List my organizations
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param cursor query string false "Cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.OrgListOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations [get]
func (h *handlers) listOrgs(c *gin.Context) {
	logEnter(c, "list_orgs")
	if !h.requireOrgs(c) {
		return
	}
	limit, err := org.ParseLimit(c.Query("limit"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	items, next, err := h.orgs.List(c.Request.Context(), middleware.UserIDFrom(c), c.Query("cursor"), limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	out := make([]dto.OrgListItem, 0, len(items))
	for _, it := range items {
		o := it.Organization
		out = append(out, dto.OrgListItem{
			Organization:     organizationView(&o),
			Role:             it.Membership.Role,
			MembershipStatus: it.Membership.Status,
		})
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OrgListOK{
		Meta:       metaFrom(c),
		Items:      out,
		NextCursor: cursorPtr(next),
	})
}

// getOrg godoc
// @Summary Get organization
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} dto.OrgOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /v1/organizations/{id} [get]
func (h *handlers) getOrg(c *gin.Context) {
	logEnter(c, "get_org")
	if !h.requireOrgs(c) {
		return
	}
	o, err := h.orgs.Get(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OrgOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
	})
}
