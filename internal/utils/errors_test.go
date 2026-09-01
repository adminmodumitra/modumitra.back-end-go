package utils

import (
	"errors"
	"net/http"
	"testing"
)

func TestECatalog(t *testing.T) {
	e := E(CodeUnauthenticated)
	if e.Code != CodeUnauthenticated || e.HTTP != http.StatusUnauthorized {
		t.Fatalf("%+v", e)
	}
	if e.Name != "unauthenticated" {
		t.Fatalf("name=%q", e.Name)
	}
	if e.Message == "" {
		t.Fatal("empty message")
	}
}

func TestGenericWrapsCause(t *testing.T) {
	cause := errors.New("disk full")
	e := Generic(cause)
	if e.Code != CodeInternal {
		t.Fatalf("code=%s", e.Code)
	}
	if !errors.Is(e, cause) {
		t.Fatal("unwrap")
	}
	if e.Message == "disk full" {
		t.Fatal("generic must not leak cause to client message")
	}
}

func TestFromPassthrough(t *testing.T) {
	orig := E(CodeRateLimited).WithDetails(map[string]any{"limiter": "bbr"})
	got := From(orig)
	if got.Code != CodeRateLimited {
		t.Fatalf("code=%s", got.Code)
	}
	if got.Details["limiter"] != "bbr" {
		t.Fatalf("details=%v", got.Details)
	}
}

func TestFromUnknownIsGeneric(t *testing.T) {
	got := From(errors.New("boom"))
	if got.Code != CodeInternal {
		t.Fatalf("code=%s", got.Code)
	}
}

func TestUnknownCodeBecomesGeneric(t *testing.T) {
	e := E(Code("not_a_real_code"))
	if e.Code != CodeInternal {
		t.Fatalf("code=%s", e.Code)
	}
}
