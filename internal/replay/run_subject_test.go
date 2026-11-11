package replay

import (
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// A Result keeps one subject per cited run, and a Hypothesis's subject may
// name a base (ADR 0020). Replay rejects a log that holds a subject outside
// that shape.
func TestRunSubjectReplayRejections(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	e := b.Exp(h, event.RecipeNone)
	base := strings.Repeat("b", 40)
	file := strings.Repeat("f", 64)
	res := func(p event.ResultRecord) string {
		p.Parent, p.Verdicts, p.Observation = e, map[string]string{h: sup}, "o"
		return b.Emit(event.TypeResultRecord, "", p)
	}
	both := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base}}, Fingerprints: []string{file}})
	empty := res(event.ResultRecord{Subjects: []event.RunSubject{{}}})
	short := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: "bbbbbbb"}}})
	twice := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base, Files: []string{file}}, {Base: base, Files: []string{file}}}})
	ok := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base, Files: []string{file}}, {Base: base}}})
	noBase := b.Emit(event.TypeHypothesisSubject, h, event.HypothesisSubject{})
	badBase := b.Emit(event.TypeHypothesisSubject, h, event.HypothesisSubject{Base: "main"})
	p := replayB(b)
	wantRejected(t, p, both, "a Result keeps subjects or fingerprints, not both")
	wantRejected(t, p, empty, "a subject with no base and no files")
	wantRejected(t, p, short, `base "bbbbbbb" is not a full commit id`)
	wantRejected(t, p, twice, "the same subject twice")
	wantRejected(t, p, noBase, "bad subject: no base and no files")
	wantRejected(t, p, badBase, `bad subject: base "main" is not a full commit id`)
	for _, r := range p.Rejections {
		if r.EventID == ok {
			t.Fatalf("one subject per cited run was rejected: %s", r.Reason)
		}
	}
}

// HoldsSubject: a subject that names a base is held only by a run at that
// base given exactly its files; one with no base by a run given each of
// its files, at any base.
func TestHoldsSubject(t *testing.T) {
	B, C := strings.Repeat("b", 40), strings.Repeat("c", 40)
	P, I := strings.Repeat("1", 64), strings.Repeat("2", 64)
	file := func(h string) []event.SubjectFile { return []event.SubjectFile{{Path: "x", SHA256: h}} }
	cases := []struct {
		name string
		runs []event.RunSubject
		sub  event.HypothesisSubject
		want bool
	}{
		{"bare B, run at B", []event.RunSubject{{Base: B}}, event.HypothesisSubject{Base: B}, true},
		{"bare B, run at B plus a patch", []event.RunSubject{{Base: B, Files: []string{P}}}, event.HypothesisSubject{Base: B}, false},
		{"bare B, merged across two runs", []event.RunSubject{{Base: B, Files: []string{P}}, {Base: C}}, event.HypothesisSubject{Base: B}, false},
		{"bare B, run at C", []event.RunSubject{{Base: C}}, event.HypothesisSubject{Base: B}, false},
		{"B plus P, run at B plus P", []event.RunSubject{{Base: B, Files: []string{P}}}, event.HypothesisSubject{Base: B, Files: file(P)}, true},
		{"B plus P, run at C plus P", []event.RunSubject{{Base: C, Files: []string{P}}}, event.HypothesisSubject{Base: B, Files: file(P)}, false},
		{"files only, old Result", []event.RunSubject{{Files: []string{P, I}}}, event.HypothesisSubject{Files: file(P)}, true},
		{"files only, run at C", []event.RunSubject{{Base: C, Files: []string{P}}}, event.HypothesisSubject{Files: file(P)}, true},
		{"no subject", []event.RunSubject{{Base: B}}, event.HypothesisSubject{}, false},
	}
	for _, c := range cases {
		if got := HoldsSubject(c.runs, c.sub); got != c.want {
			t.Errorf("%s: HoldsSubject = %v, want %v", c.name, got, c.want)
		}
	}
}
