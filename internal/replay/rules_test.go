package replay

import (
	"reflect"
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

// Every event type has one definition with a replay rule and an apply.
func TestEveryEventTypeHasARule(t *testing.T) {
	for typ := range event.EventTypes {
		r, ok := rules[typ]
		if !ok {
			t.Errorf("%s has no rule", typ)
			continue
		}
		if r.validate == nil || r.apply == nil {
			t.Errorf("%s: validate or apply is nil", typ)
		}
	}
	for typ := range rules {
		if !event.EventTypes[typ] {
			t.Errorf("rule for unknown event type %s", typ)
		}
	}
}

// factsFields is what a replay rule may read: stored facts and lookup indexes
// over them. A field added to Facts needs adding here, which is the moment to
// ask whether it is stored.
var factsFields = []string{
	"Nodes", "Order", "SeenIDs", "Retracted", "Abandoned", "Withdrawn", "Edges", "Hold", "Decisions", "Aliases", "Goals",
	"ResultsByHyp", "EdgesFrom", "SupersededN", "Answers", "LimitHand", "Lines", "LinesBy", "Subjects",
}

// derivedFields record derived status, so a replay rule must not reach them
// (spec §5.1): Snaps, EverAnswered and RivalsAtAdd store it, and the others
// are it or depend on the recipe hashes a Step was given.
var derivedFields = []string{"Hyps", "Questions", "goals", "Snaps", "EverAnswered", "RivalsAtAdd", "recipeHashes", "Applied"}

// A replay rule takes Facts, so what Facts holds is what a rule can read.
func TestFactsHoldsStoredFactsOnly(t *testing.T) {
	ft := reflect.TypeFor[Facts]()
	for f := range ft.Fields() {
		if !slices.Contains(factsFields, f.Name) {
			t.Errorf("Facts has the field %s, which is not in the allowed set; a replay rule can read it", f.Name)
		}
		if slices.Contains(derivedFields, f.Name) {
			t.Errorf("Facts has the derived field %s", f.Name)
		}
	}
	st := reflect.TypeFor[State]()
	for _, name := range derivedFields {
		f, ok := st.FieldByName(name)
		if !ok {
			t.Errorf("State has no field %s", name)
		} else if len(f.Index) != 1 {
			t.Errorf("State's %s comes through an embedded struct", name)
		}
	}
}
