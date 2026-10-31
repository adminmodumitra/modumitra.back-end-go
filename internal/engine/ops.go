package engine

// Phase 4 manufacturing ops (cut list, CNC machine files, production package).
const (
	OpCutlist           = "cutlist"
	OpCNC               = "cnc"
	OpProductionPackage = "production-package"
)

// Phase 5a document ops: built on the server, watermarked with the org id the
// proxy sends (X-Org-Id). Phase 5b adds the production packet PDF (the engine
// inlines the print styles its Node rasteriser needs).
const (
	OpProposalPDF         = "proposal-pdf"
	OpRevisionSummaryPDF  = "revision-summary-pdf"
	OpQuoteExport         = "quote-export"
	OpProductionPacketPDF = "production-packet-pdf"
)

// ViewPrefix: Phase 5b display ops, POST /v1/engine/view/<op> → /v1/ops/view/<op>,
// body {"args": [...]}. Results the app shows (report, cut list, review gates)
// or applies in a review action; computed ahead after edits.
const ViewPrefix = "view/"

// viewOps lists the display ops (src/domain/engineDisplay/displayOps.ts) and
// whether the engine prices them (the proxy then injects the org book).
var viewOps = map[string]bool{
	"createProjectReport":            true,
	"createWholeProjectReport":       true,
	"createExportableProjectCutlist": false,
	"createExportableCabinetCutlist": false,
	"createMachineJobDocument":       false,
	"scopedProductionOutputs":        false,
	"createCabinetProductionCutlist": false,
	"createCabinetMachineParts":      false,
	"csvFromMachineParts":            false,
	"resolveCommercialInputs":        true,
	"measureInteriorEstimate":        true,
	"canReleaseForProduction":        false,
	"buildProductionReadinessGate":   false,
	"createRevisionFingerprint":      false,
	"createRevisionSnapshot":         false,
	"releaseForProduction":           false,
	"gateFreezeQuotes":               false,
	"freezeCabinetProjectQuote":      true,
	"tryFreezeProposal":              true,
	"planSeamlessCreate":             true,
	"buildProposalGate":              true,
}

// Spec is how the proxy treats one op.
type Spec struct {
	Op string
	// Field is the body field that must be a JSON value: "document" (an
	// InteriorProject, pricing) or "project" (a CabinetProject, manufacturing).
	Field string
	// NeedsBook: the proxy loads the org price book and injects it. Cut list
	// and CNC never price, so they skip the database read.
	NeedsBook bool
	// Manufacturing ops draw from their own quota bucket (ENGINE_MFG_QUOTA_*),
	// so computed-ahead exports never starve live quotes.
	Manufacturing bool
	// Document ops (5a) draw from the document bucket (ENGINE_DOC_QUOTA_*), take
	// a larger body (proposal stills) and get the longer ENGINE_DOC_TIMEOUT.
	Document bool
	// View ops (5b) draw from the display bucket (ENGINE_VIEW_QUOTA_*): several
	// per committed edit, computed ahead, so they never starve quotes or exports.
	View bool
}

var specs = map[string]Spec{
	OpQuote:             {Op: OpQuote, Field: "document", NeedsBook: true},
	OpEstimate:          {Op: OpEstimate, Field: "document", NeedsBook: true},
	OpCutlist:           {Op: OpCutlist, Field: "project", Manufacturing: true},
	OpCNC:               {Op: OpCNC, Field: "project", Manufacturing: true},
	OpProductionPackage: {Op: OpProductionPackage, Field: "project", NeedsBook: true, Manufacturing: true},
	// The proposal and the quote export price with the org book; the revision
	// summary prices from the project's own preferences, as the app does.
	OpProposalPDF:        {Op: OpProposalPDF, Field: "document", NeedsBook: true, Document: true},
	OpRevisionSummaryPDF: {Op: OpRevisionSummaryPDF, Field: "project", Document: true},
	OpQuoteExport:        {Op: OpQuoteExport, Field: "project", NeedsBook: true, Document: true},
	// 5b: prices from the project's own preferences, as the app does (no book).
	OpProductionPacketPDF: {Op: OpProductionPacketPDF, Field: "project", Document: true},
}

func init() {
	for name, priced := range viewOps {
		op := ViewPrefix + name
		specs[op] = Spec{Op: op, Field: "args", NeedsBook: priced, View: true}
	}
}

// SpecFor returns the spec of a known op; anything else is refused before a
// network call.
func SpecFor(op string) (Spec, bool) {
	s, ok := specs[op]
	return s, ok
}
