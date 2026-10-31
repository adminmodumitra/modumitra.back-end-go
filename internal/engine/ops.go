package engine

// Phase 4 manufacturing ops (cut list, CNC machine files, production package).
const (
	OpCutlist           = "cutlist"
	OpCNC               = "cnc"
	OpProductionPackage = "production-package"
)

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
}

var specs = map[string]Spec{
	OpQuote:             {Op: OpQuote, Field: "document", NeedsBook: true},
	OpEstimate:          {Op: OpEstimate, Field: "document", NeedsBook: true},
	OpCutlist:           {Op: OpCutlist, Field: "project", Manufacturing: true},
	OpCNC:               {Op: OpCNC, Field: "project", Manufacturing: true},
	OpProductionPackage: {Op: OpProductionPackage, Field: "project", NeedsBook: true, Manufacturing: true},
}

// SpecFor returns the spec of a known op; anything else is refused before a
// network call.
func SpecFor(op string) (Spec, bool) {
	s, ok := specs[op]
	return s, ok
}
