package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

// First-party desktop sign-in: no OAuth provider; the signed-in web user
// approves the code the desktop app shows.
func TestDeviceApproveFirstParty(t *testing.T) {
	hr := newSSOHarness(t)
	access, userID := oauthLoginAccess(t, hr, "google", "code-approve", sso.Profile{ProviderUserID: "gid_appr", Email: "appr@example.com", EmailVerified: true, DisplayName: "Appr"})

	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/code", map[string]any{"client_id": "cabinet-studio-desktop"}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
	var issued map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &issued)
	deviceCode, _ := issued["device_code"].(string)
	userCode, _ := issued["user_code"].(string)

	if w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/approve", map[string]any{"user_code": userCode}, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("approve without login=%d", w.Code)
	}
	bearer := map[string]string{"Authorization": "Bearer " + access}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/approve", map[string]any{"user_code": ""}, bearer); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty code=%d %s", w.Code, w.Body.String())
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/approve", map[string]any{"user_code": "NOPE-NOPE"}, bearer); w.Code != http.StatusNotFound {
		t.Fatalf("unknown code=%d", w.Code)
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/approve", map[string]any{"user_code": userCode}, bearer); w.Code != http.StatusOK {
		t.Fatalf("approve=%d %s", w.Code, w.Body.String())
	}

	w = doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/token", map[string]any{"device_code": deviceCode}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("token=%d %s", w.Code, w.Body.String())
	}
	var login map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &login)
	user, _ := login["user"].(map[string]any)
	if login["refresh_token"] == "" || user["id"] != userID {
		t.Fatalf("login=%v", login)
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/approve", map[string]any{"user_code": userCode}, bearer); w.Code < 400 {
		t.Fatalf("re-approve a used code=%d", w.Code)
	}
}
