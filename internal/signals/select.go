package signals

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/replay"
)

// verdict is one insufficient verifier confirm: its event id and its actor.
type verdict struct{ id, actor string }

// selState is a question's current select and its chain of insufficient
// verifier verdicts.
type selState struct {
	q, sel, hyp string
	tradeoff    bool
	chain       []verdict
	request     string // a request-verify after the chain's last verdict, or ""
}

// viewOf copies q's current select and its verifier chain from sel, the
// chain replay keeps for the verifier-loop signal (replay.Decision). It returns
// nil when q has no current select.
func viewOf(q string, sel *replay.Decision) *selState {
	if sel == nil || sel.Current == "" {
		return nil
	}
	rec := sel.Selects[sel.Current]
	s := &selState{q: q, sel: sel.Current, hyp: rec.Hyp, tradeoff: rec.Tradeoff, request: sel.ChainRequest}
	for _, v := range sel.Chain {
		s.chain = append(s.chain, verdict{v.ID, v.Actor})
	}
	return s
}

// describe names the select and its verdicts, for example
// "Q-KW0PWZ selected H-BDTKAB (trade-off, select 8VPZAQ): verifier insufficient x2 (verifier-s3-1, verifier-s4-1)".
func (s *selState) describe(display func(string) string) string {
	kind := ""
	if s.tradeoff {
		kind = "trade-off, "
	}
	verdicts := "no insufficient verdict"
	if len(s.chain) > 0 {
		var actors []string
		for _, v := range s.chain {
			actors = append(actors, v.actor)
		}
		verdicts = fmt.Sprintf("verifier insufficient x%d (%s)", len(s.chain), strings.Join(actors, ", "))
	}
	return fmt.Sprintf("%s selected %s (%sselect %s): %s", display(s.q), display(s.hyp), kind, s.sel[len(s.sel)-6:], verdicts)
}
