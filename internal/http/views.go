package httpapi

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func metaFrom(c *gin.Context) utils.Meta {
	return utils.Meta{
		RequestID: middleware.RequestIDFrom(c),
		TraceID:   middleware.TraceIDFrom(c),
	}
}

func rfcTime(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func rfcTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func cursorPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func emptyPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func userView(u *identity.User) dto.UserView {
	if u == nil {
		return dto.UserView{}
	}
	return dto.UserView{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified(),
		DisplayName:   u.DisplayName,
		CreatedAt:     u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func capabilitiesView(c device.Capabilities) dto.Capabilities {
	return dto.Capabilities{
		CloudProjects:      c.CloudProjects,
		OrganizationAssets: c.OrganizationAssets,
		ProposalBranding:   c.ProposalBranding,
		EngineeringHandoff: c.EngineeringHandoff,
		MaxSeats:           c.MaxSeats,
		MaxDevices:         c.MaxDevices,
		MaxCloudProjects:   c.MaxCloudProjects,
	}
}
