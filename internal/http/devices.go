package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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
	uid := middleware.UserIDFrom(c)
	cursor := c.Query("cursor")
	key := cache.Devices(uid, include, cursor, limit)
	var cached dto.DeviceListOK
	if cache.GetJSON(c.Request.Context(), h.cache, key, &cached) {
		cached.Meta = metaFrom(c)
		middleware.WriteJSON(c, utils.StatusOK, cached)
		return
	}
	items, next, active, err := h.devices.List(c.Request.Context(), uid, include, cursor, limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]dto.DeviceView, 0, len(items))
	for i := range items {
		views = append(views, deviceView(&items[i]))
	}
	body := dto.DeviceListOK{
		Meta:        metaFrom(c),
		Items:       views,
		NextCursor:  cursorPtr(next),
		ActiveCount: active,
		DeviceLimit: device.DeviceLimit,
	}
	cache.SetJSON(c.Request.Context(), h.cache, key, body, cache.TTLDevices)
	middleware.WriteJSON(c, utils.StatusOK, body)
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
	uid := middleware.UserIDFrom(c)
	d, snap, err := h.devices.Activate(c.Request.Context(), uid, body.Name, body.Platform)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateDevices(c, uid)
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
	uid := middleware.UserIDFrom(c)
	d, err := h.devices.Rename(c.Request.Context(), uid, c.Param("id"), name)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateDevices(c, uid)
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
	uid := middleware.UserIDFrom(c)
	d, err := h.devices.Deactivate(c.Request.Context(), uid, c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateDevices(c, uid)
	middleware.WriteJSON(c, utils.StatusOK, dto.DeviceDeactivatedOK{
		Meta:   metaFrom(c),
		Status: "deactivated",
		ID:     d.ID,
	})
}
