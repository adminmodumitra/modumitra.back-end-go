package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// patchDevice godoc
// @Summary Rename device
// @Tags devices
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param body body dto.PatchDeviceBody true "Patch"
// @Success 200 {object} dto.DeviceOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/devices/{id} [patch]
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

// deleteDevice godoc
// @Summary Deactivate device
// @Tags devices
// @Security BearerAuth
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} dto.DeviceDeactivatedOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/devices/{id} [delete]
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
