package replay

import "github.com/dat9uy/hyp/internal/event"

// Res is the derived record of one Result. Replay builds it once per
// projection, as it builds Exp, and every reader takes a Result's state from
// it. Uncontained is the one field replay does not fill: contain.Mark sets
// it from files, because replay reads none.
type Res struct {
	Status       string   // ResRecorded or ResRetracted; equal to Projection.Status[id].Status
	Questionable bool     // a pin's ref is missing from the current recipe hashes or hashes differently
	Pins         []Pin    // its Experiment's recipe, then each evidence pin, in the order cited
	Uncontained  []string // why hyp cannot vouch for its verdict script (spec §7.6); set by contain.Mark
}

// results builds the Res record of every Result from the stored state.
func (st *State) results() map[string]*Res {
	out := map[string]*Res{}
	for _, id := range st.Order {
		if st.Nodes[id].Kind != event.KindResult {
			continue
		}
		r := &Res{Status: ResRecorded, Pins: pins(st.Nodes, id)}
		if st.Retracted[id] {
			r.Status = ResRetracted
		}
		for _, pin := range r.Pins {
			if h, ok := st.recipeHashes[pin.Ref.String()]; !ok || h != pin.Hash {
				r.Questionable = true
			}
		}
		out[id] = r
	}
	return out
}
