package sso

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func testAccounts(t *testing.T) *identity.Service {
	t.Helper()
	svc, _, _ := identity.NewMemoryService(identity.Settings{
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
		JWTSecret:      "test-secret-for-cabinet-studio-hs256",
	})
	return svc
}

func TestStartRejectsBadRedirect(t *testing.T) {
	ident := testAccounts(t)
	svc, _, _ := NewMemoryService(ident, Settings{TermsVersion: "2026-06-01", PrivacyVersion: "2026-06-01", PKCESecret: "s"})
	_, err := svc.Start(context.Background(), StartInput{Provider: ProviderGoogle, RedirectURI: "http://evil.example/cb", Intent: IntentLogin})
	var ae *utils.Error
	if !errors.As(err, &ae) || ae.Code != utils.CodeRedirectURINotAllowed {
		t.Fatalf("err=%v", err)
	}
}

func TestStartUnknownProviderValidation(t *testing.T) {
	ident := testAccounts(t)
	svc, _, _ := NewMemoryService(ident, Settings{PKCESecret: "s"})
	_, err := svc.Start(context.Background(), StartInput{Provider: "yahoo", RedirectURI: "cabinetstudio://oauth/callback", Intent: IntentLogin})
	var ae *utils.Error
	if !errors.As(err, &ae) || ae.Code != utils.CodeValidation {
		t.Fatalf("err=%v", err)
	}
}

func TestPKCERoundTrip(t *testing.T) {
	v := DeriveVerifier("secret", "st_abc")
	if v == "" || strings.Contains(v, "+") || strings.Contains(v, "/") {
		t.Fatalf("verifier=%q", v)
	}
	if HashVerifier(v) == HashVerifier(v+"x") {
		t.Fatal("hash collision")
	}
	if S256Challenge(v) == "" {
		t.Fatal("challenge")
	}
}
