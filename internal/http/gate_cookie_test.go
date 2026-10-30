package httpapi

import (
	"testing"
	"time"
)

// Same vector as edge/app-gate/appGate.test.ts in cabinet-designer-mvp.
func TestGateValueVector(t *testing.T) {
	got := gateValue("gate-test-key", time.Unix(1_800_000_000, 0))
	want := "v1.1800000000.q5TkX7KxPmuCegauD7ecpZRik3FxlnB_mi464gFl7Vo"
	if got != want {
		t.Fatalf("gateValue=%q want %q", got, want)
	}
}
