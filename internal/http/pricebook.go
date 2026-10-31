package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/pricebook"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

var jsonNull = json.RawMessage("null")

// orgPriceBook returns the org's book JSON, or nil for "engine defaults".
// Read-through cached; putPriceBook invalidates.
func (h *handlers) orgPriceBook(ctx context.Context, orgID string) (json.RawMessage, error) {
	if orgID == "" || h.priceBooks == nil {
		return nil, nil
	}
	key := cache.OrgPriceBook(orgID)
	if raw, ok := h.cache.Get(ctx, key); ok {
		if string(raw) == "null" {
			return nil, nil
		}
		return raw, nil
	}
	b, err := h.priceBooks.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}
	var raw json.RawMessage
	if b != nil {
		raw = b.Book
	}
	stored := raw
	if stored == nil {
		stored = jsonNull
	}
	h.cache.Set(ctx, key, stored, cache.TTLPriceBook)
	return raw, nil
}

func bookOrNull(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return jsonNull
	}
	return raw
}

// getPriceBook godoc
// @Summary Get the organization price book
// @Description Rates the pricing engine uses for this org. book is null when none is set (engine defaults). Owner, admin or billing_admin.
// @Tags pricing
// @Security BearerAuth
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Success 200 {object} dto.PriceBookOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /v1/org/price-book [get]
func (h *handlers) getPriceBook(c *gin.Context) {
	logEnter(c, "get_price_book")
	orgID, role, err := h.callerOrg(c)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	if orgID == "" {
		middleware.WriteJSON(c, utils.StatusOK, dto.PriceBookOK{Meta: metaFrom(c), Book: jsonNull})
		return
	}
	if !org.RolePermits(role, org.ActionPriceBookRead) {
		middleware.Abort(c, utils.E(utils.CodeForbidden))
		return
	}
	if h.priceBooks == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return
	}
	raw, err := h.orgPriceBook(c.Request.Context(), orgID)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.PriceBookOK{Meta: metaFrom(c), Book: bookOrNull(raw)})
}

// putPriceBook godoc
// @Summary Replace the organization price book
// @Description book must be a JSON object, at most 1 MB. Owner or admin.
// @Tags pricing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.PriceBookBody true "Price book"
// @Success 200 {object} dto.PriceBookOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/org/price-book [put]
func (h *handlers) putPriceBook(c *gin.Context) {
	logEnter(c, "put_price_book")
	orgID, role, err := h.callerOrg(c)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	if orgID == "" {
		middleware.Abort(c, utils.E(utils.CodeNotFound)) // personal accounts have no org book
		return
	}
	if !org.RolePermits(role, org.ActionPriceBookWrite) {
		middleware.Abort(c, utils.E(utils.CodeForbidden))
		return
	}
	if h.priceBooks == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		middleware.Abort(c, bodyReadError(err))
		return
	}
	var body dto.PriceBookBody
	if err := json.Unmarshal(raw, &body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := pricebook.ValidateBook(body.Book); err != nil {
		middleware.Abort(c, err)
		return
	}
	ctx := c.Request.Context()
	book := pricebook.Book{OrgID: orgID, Book: body.Book, UpdatedAt: time.Now().UTC(), UpdatedBy: middleware.UserIDFrom(c)}
	if err := h.priceBooks.Put(ctx, book); err != nil {
		middleware.Abort(c, err)
		return
	}
	h.cache.Delete(ctx, cache.OrgPriceBook(orgID))
	middleware.WriteJSON(c, utils.StatusOK, dto.PriceBookOK{Meta: metaFrom(c), Book: body.Book})
}

// bodyReadError maps http.MaxBytesReader overflow to 413.
func bodyReadError(err error) error {
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		return utils.E(utils.CodePayloadTooLarge)
	}
	return utils.E(utils.CodeValidation)
}
