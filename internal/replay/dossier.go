package replay

import (
	"cmp"
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// DossierItem kinds.
const (
	DossierBody   = "body"
	DossierHyp    = "hypothesis"
	DossierResult = "result"
	DossierLink   = "link"
)

// DossierItem is one fact of a Dossier. Two items are the same fact exactly
// when they compare equal.
type DossierItem struct {
	Kind string
	Node string // the Hypothesis of a body or hypothesis item, the Result of a result item, the link's from
	To   string // the Hypothesis a result item's verdict is on, the link's to
	Word string // the verdict of a result item, the link's type
	Body string // the text of a body item
}

// Dossier is the evidence the Decision on one DesignQuestion rests on (spec
// §7.4). It is derived, never stored.
type Dossier struct {
	Items []DossierItem // sorted, so that two Dossiers compare item by item
	// Deciding lists the Results whose contradicts alone keeps a rival from
	// being a trade-off candidate. It is read from the statuses, so it stays
	// out of what two Dossiers compare.
	Deciding map[string][]string // rival Hypothesis → Results
}

// dossierItems derives what question q's Dossier holds for a select of hyp,
// from stored state only: hyp's body, the live child Hypotheses, each
// non-retracted Result with a verdict on one of them, found by verdict and not
// by tree parent, and the live links going out from those Hypotheses and
// Results, without external_ref.
func (st *State) dossierItems(q, hyp string) []DossierItem {
	items := []DossierItem{{Kind: DossierBody, Node: hyp, Body: st.Nodes[hyp].Body}}
	links := func(from string) {
		for _, i := range st.EdgesFrom[from] {
			if e := st.Edges[i]; e.Live && e.Type != event.EdgeExternalRef {
				items = append(items, DossierItem{Kind: DossierLink, Node: from, To: e.To, Word: e.Type})
			}
		}
	}
	seen := map[string]bool{}
	for _, h := range st.Nodes[q].Children {
		if !st.isLiveHyp(h) {
			continue
		}
		items = append(items, DossierItem{Kind: DossierHyp, Node: h})
		links(h)
		for _, r := range st.resultsFor(h) {
			items = append(items, DossierItem{Kind: DossierResult, Node: r, To: h, Word: st.Nodes[r].Verdicts[h]})
			if !seen[r] {
				seen[r] = true
				links(r)
			}
		}
	}
	slices.SortFunc(items, func(a, b DossierItem) int {
		return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Node, b.Node), cmp.Compare(a.To, b.To),
			cmp.Compare(a.Word, b.Word), cmp.Compare(a.Body, b.Body))
	})
	return items
}

// diffItems lists what is in now and not in was, then what is in was and not
// in now, each in the order of its list.
func diffItems(was, now []DossierItem) (added, removed []DossierItem) {
	for _, it := range now {
		if !slices.Contains(was, it) {
			added = append(added, it)
		}
	}
	for _, it := range was {
		if !slices.Contains(now, it) {
			removed = append(removed, it)
		}
	}
	return added, removed
}

// Diff lists what o holds that d does not, then what d holds that o does not.
func (d *Dossier) Diff(o *Dossier) (added, removed []DossierItem) { return diffItems(d.Items, o.Items) }

// Equal reports whether two Dossiers hold the same facts.
func (d *Dossier) Equal(o *Dossier) bool { return slices.Equal(d.Items, o.Items) }

// dossier builds question q's Dossier for a select of hyp, with its marks read
// from the last derivation.
func (st *State) dossier(q, hyp string) *Dossier {
	d := &Dossier{Items: st.dossierItems(q, hyp), Deciding: map[string][]string{}}
	for _, it := range d.Items {
		if it.Kind == DossierHyp && it.Node != hyp {
			if rs := st.decidingContradictions(it.Node); len(rs) > 0 {
				d.Deciding[it.Node] = rs
			}
		}
	}
	return d
}

// decidingContradictions lists the Results whose contradicts alone keeps
// Hypothesis h from being a trade-off candidate: without that one verdict, rule
// 3 of spec §7.1 gives supported. A Hypothesis with two contradicts has none,
// since removing either leaves the other.
func (st *State) decidingContradictions(h string) []string {
	var out []string
	for _, r := range st.Contradictions(h) {
		if st.verdictStatus(h, r) == HypSupported {
			out = append(out, r)
		}
	}
	return out
}
