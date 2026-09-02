package httpapi

import (
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) requireDevices(c *gin.Context) bool {
	if h.devices == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

func parseBoolQuery(raw string, fallback bool) (bool, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, utils.E(utils.CodeValidation)
	}
}

func (h *handlers) listDevices(c *gin.Context) {
	logEnter(c, "list_devices")
	if !h.requireDevices(c) {
		return
	}
	include, err := parseBoolQuery(c.Query("include_deactivated"), false)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	limit, err := device.ParseLimit(c.Query("limit"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	items, next, active, err := h.devices.List(c.Request.Context(), middleware.UserIDFrom(c), include, c.Query("cursor"), limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]dto.DeviceView, 0, len(items))
	for i := range items {
		views = append(views, deviceView(&items[i]))
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.DeviceListOK{
		Meta:        metaFrom(c),
		Items:       views,
		NextCursor:  cursorPtr(next),
		ActiveCount: active,
		DeviceLimit: device.DeviceLimit,
	})
}

func (h *handlers) activateDevice(c *gin.Context) {
	logEnter(c, "activate_device")
	if !h.requireDevices(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.ActivateDeviceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	d, snap, err := h.devices.Activate(c.Request.Context(), middleware.UserIDFrom(c), body.Name, body.Platform)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusCreated, dto.ActivateOK{
		Meta:                metaFrom(c),
		Device:              deviceView(d),
		EntitlementSnapshot: snapshotView(snap),
	})
}

func (h *handlers) patchDevice(c *gin.Context) {
	logEnter(c, "patch_device")
	if !h.requireDevices(c) {
		return
	}
	var body dto.PatchDeviceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	name := ""
	if body.Name != nil {
		name = *body.Name
	}
	d, err := h.devices.Rename(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), name)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.DeviceOK{
		Meta:   metaFrom(c),
		Device: deviceView(d),
	})
}

func (h *handlers) deleteDevice(c *gin.Context) {
	logEnter(c, "delete_device")
	if !h.requireDevices(c) {
		return
	}
	d, err := h.devices.Deactivate(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.DeviceDeactivatedOK{
		Meta:   metaFrom(c),
		Status: "deactivated",
		ID:     d.ID,
	})
}
