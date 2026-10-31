package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPriceBookOwnerPutGet(t *testing.T) {
	hr := engineRouter(t, nil)
	access, userID := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-pb-1")
	orgID := createOrgAs(t, hr.h, access, "org-pb-1")
	w := doJSON(t, hr.h, http.MethodGet, "/v1/org/price-book", nil, bearer(access))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"book":null`) {
		t.Fatalf("get empty=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{"carcass": 450}}, bearer(access))
	if w.Code != http.StatusOK {
		t.Fatalf("put=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, hr.h, http.MethodGet, "/v1/org/price-book", nil, map[string]string{
		"Authorization": "Bearer " + access, "X-Organization-Id": orgID,
	})
	var got struct {
		Book map[string]int `json:"book"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Book["carcass"] != 450 {
		t.Fatalf("get=%d %s", w.Code, w.Body.String())
	}
	stored, _ := hr.books.Get(t.Context(), orgID)
	if stored == nil || stored.UpdatedBy != userID {
		t.Fatalf("stored=%+v", stored)
	}
}

func TestPriceBookRoles(t *testing.T) {
	hr := engineRouter(t, nil)
	owner, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-pb-2")
	orgID := createOrgAs(t, hr.h, owner, "org-pb-2")
	viewer := joinAs(t, hr, owner, orgID, "viewer@example.com", "viewer")
	admin := joinAs(t, hr, owner, orgID, "admin@example.com", "admin")
	scoped := func(access string) map[string]string {
		return map[string]string{"Authorization": "Bearer " + access, "X-Organization-Id": orgID}
	}
	book := map[string]any{"book": map[string]any{"rate": 1}}
	if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", book, scoped(viewer)); w.Code != http.StatusForbidden {
		t.Fatalf("viewer put=%d %s", w.Code, w.Body.String())
	}
	if w := doJSON(t, hr.h, http.MethodGet, "/v1/org/price-book", nil, scoped(viewer)); w.Code != http.StatusForbidden {
		t.Fatalf("viewer get=%d %s", w.Code, w.Body.String())
	}
	if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", book, scoped(admin)); w.Code != http.StatusOK {
		t.Fatalf("admin put=%d %s", w.Code, w.Body.String())
	}
	outsider, _ := registerVerifyLogin(t, hr.h, hr.mail, "eve@example.com", "reg-pb-eve")
	if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", book, scoped(outsider)); w.Code != http.StatusNotFound {
		t.Fatalf("outsider put=%d %s", w.Code, w.Body.String())
	}
	if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", book, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("signed out put=%d", w.Code)
	}
}

func TestPriceBookValidation(t *testing.T) {
	hr := engineRouter(t, nil)
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-pb-3")
	createOrgAs(t, hr.h, access, "org-pb-3")
	for _, body := range []any{map[string]any{"book": []int{1}}, map[string]any{"book": nil}, map[string]any{}} {
		if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", body, bearer(access)); w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%v=%d %s", body, w.Code, w.Body.String())
		}
	}
	big := map[string]any{"book": map[string]any{"blob": strings.Repeat("a", 1<<20)}}
	if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", big, bearer(access)); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("1MB+=%d %s", w.Code, w.Body.String())
	}
	solo, _ := registerVerifyLogin(t, hr.h, hr.mail, "solo@example.com", "reg-pb-solo")
	if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{}}, bearer(solo)); w.Code != http.StatusNotFound {
		t.Fatalf("personal put=%d %s", w.Code, w.Body.String())
	}
}
