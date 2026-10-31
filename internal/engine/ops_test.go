package engine

import (
	"encoding/json"
	"testing"
)

func TestSpecForKnownOps(t *testing.T) {
	for op, want := range map[string]Spec{
		OpQuote:             {Op: OpQuote, Field: "document", NeedsBook: true},
		OpCutlist:           {Op: OpCutlist, Field: "project", Manufacturing: true},
		OpCNC:               {Op: OpCNC, Field: "project", Manufacturing: true},
		OpProductionPackage: {Op: OpProductionPackage, Field: "project", NeedsBook: true, Manufacturing: true},
	} {
		if got, ok := SpecFor(op); !ok || got != want {
			t.Errorf("%s: got %+v ok=%v", op, got, ok)
		}
	}
	if _, ok := SpecFor("../admin"); ok {
		t.Fatal("unknown op accepted")
	}
}

func TestBuildBodyForManufacturing(t *testing.T) {
	cut, _ := SpecFor(OpCutlist)
	out, err := BuildBodyFor([]byte(`{"project":{"cabinets":[]},"scope":"room-1","priceBook":{"rate":1}}`), cut, json.RawMessage(`{"rate":450}`))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	_ = json.Unmarshal(out, &got)
	if _, leaked := got["priceBook"]; leaked || string(got["scope"]) != `"room-1"` || string(got["project"]) != `{"cabinets":[]}` {
		t.Fatalf("%s", out)
	}
	pkg, _ := SpecFor(OpProductionPackage)
	out, err = BuildBodyFor([]byte(`{"project":{},"room":{"dimensions":{}}}`), pkg, json.RawMessage(`{"rate":450}`))
	if err != nil {
		t.Fatal(err)
	}
	got = map[string]json.RawMessage{}
	_ = json.Unmarshal(out, &got)
	if string(got["priceBook"]) != `{"rate":450}` || string(got["room"]) != `{"dimensions":{}}` {
		t.Fatalf("%s", out)
	}
	for _, raw := range []string{`{"document":{}}`, `{"project":null}`, `[]`} {
		if _, err := BuildBodyFor([]byte(raw), cut, nil); err == nil {
			t.Errorf("%q: want error", raw)
		}
	}
}

func TestSpecForDocumentOps(t *testing.T) {
	for op, want := range map[string]Spec{
		OpProposalPDF:        {Op: OpProposalPDF, Field: "document", NeedsBook: true, Document: true},
		OpRevisionSummaryPDF: {Op: OpRevisionSummaryPDF, Field: "project", Document: true},
		OpQuoteExport:        {Op: OpQuoteExport, Field: "project", NeedsBook: true, Document: true},
	} {
		if got, ok := SpecFor(op); !ok || got != want {
			t.Errorf("%s: got %+v ok=%v", op, got, ok)
		}
	}
	if _, ok := SpecFor("production-packet-pdf"); ok {
		t.Fatal("the packet PDF is not served yet (Node cannot rasterise its sheets)")
	}
	if MaxResponseFor(OpProposalPDF) != MaxDocumentResponseBytes || MaxResponseFor(OpCNC) != MaxResponseBytes {
		t.Fatal("response caps")
	}
}
