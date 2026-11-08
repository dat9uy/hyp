package signals

import (
	"fmt"

	"github.com/dat9uy/hyp/internal/replay"
)

// selState is a question's current select.
type selState struct {
	q, sel, hyp string
}

// viewOf copies q's current select from sel. It returns nil when q has no
// current select.
func viewOf(q string, sel *replay.Decision) *selState {
	if sel == nil || sel.Current == "" {
		return nil
	}
	return &selState{q: q, sel: sel.Current, hyp: sel.Selects[sel.Current].Hyp}
}

// describe names the select, for example "Q-KW0PWZ selected H-BDTKAB (select 8VPZAQ)".
func (s *selState) describe(display func(string) string) string {
	return fmt.Sprintf("%s selected %s (select %s)", display(s.q), display(s.hyp), s.sel[len(s.sel)-6:])
}
