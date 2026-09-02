package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) listInvoices(c *gin.Context) {
	logEnter(c, "list_invoices")
	if !h.requireBilling(c) {
		return
	}
	limit, err := billing.ParseLimit(c.Query("limit"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	items, next, err := h.billing.ListInvoices(c.Request.Context(), middleware.UserIDFrom(c), h.billingTenant(c), c.Query("status"), c.Query("cursor"), limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]dto.InvoiceView, 0, len(items))
	for i := range items {
		views = append(views, invoiceView(&items[i]))
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.InvoiceListOK{
		Meta:       metaFrom(c),
		Items:      views,
		NextCursor: cursorPtr(next),
	})
}

func (h *handlers) getInvoice(c *gin.Context) {
	logEnter(c, "get_invoice")
	if !h.requireBilling(c) {
		return
	}
	inv, err := h.billing.GetInvoice(c.Request.Context(), middleware.UserIDFrom(c), h.billingTenant(c), c.Param("invoiceId"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.InvoiceOK{
		Meta:    metaFrom(c),
		Invoice: invoiceView(inv),
	})
}
