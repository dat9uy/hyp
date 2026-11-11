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
		if p.Evidence == nil {
			p.Evidence = []string{".tree/evidence/E/sandbox.json"}
		}
		return b.Emit(event.TypeResultRecord, "", p)
	}
	both := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base}}, Fingerprints: []string{file}})
	empty := res(event.ResultRecord{Subjects: []event.RunSubject{{}}})
	short := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: "bbbbbbb"}}})
	twice := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base, Files: []string{file}}, {Base: base, Files: []string{file}}}})
	noEvidence := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base}}, Evidence: []string{}})
	ok := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base, Files: []string{file}}, {Base: base}, {Base: base, Patch: file, Files: []string{file}}}})
	badPatch := res(event.ResultRecord{Subjects: []event.RunSubject{{Base: base, Patch: "p"}}})
	noBase := b.Emit(event.TypeHypothesisSubject, h, event.HypothesisSubject{})
	patchNoBase := b.Emit(event.TypeHypothesisSubject, h, event.HypothesisSubject{Patch: &event.SubjectFile{Path: "p", SHA256: file}, Files: []event.SubjectFile{{Path: "x", SHA256: file}}})
	patchNotSHA := b.Emit(event.TypeHypothesisSubject, h, event.HypothesisSubject{Base: base, Patch: &event.SubjectFile{Path: "p", SHA256: "p"}})
	badBase := b.Emit(event.TypeHypothesisSubject, h, event.HypothesisSubject{Base: "main"})
	p := replayB(b)
	wantRejected(t, p, both, "a Result keeps subjects or fingerprints, not both")
	wantRejected(t, p, noEvidence, "bad subjects: a Result with no evidence has no run to keep a subject of")
	wantRejected(t, p, empty, "a subject with no base, no patch and no files")
	wantRejected(t, p, short, `base "bbbbbbb" is not a full commit id`)
	wantRejected(t, p, badPatch, `bad subjects: patch "p" is not a sha256`)
	wantRejected(t, p, twice, "the same subject twice")
	wantRejected(t, p, noBase, "bad subject: no base and no files")
	wantRejected(t, p, patchNoBase, "bad subject: a patch with no base")
	wantRejected(t, p, patchNotSHA, `bad subject: patch "p" is not a sha256`)
	wantRejected(t, p, badBase, `bad subject: base "main" is not a full commit id`)
	for _, r := range p.Rejections {
		if r.EventID == ok {
			t.Fatalf("one subject per cited run was rejected: %s", r.Reason)
		}
	}
}

// HoldsSubject: a subject that names a base is held only by a run at that
// base with the same patch, none when it names none, given each of its
// files as input; input files are data, so a run at B given inputs holds
// bare B. One with no base is held by a run given each of its files, as
// patch or input, at any base.
func TestHoldsSubject(t *testing.T) {
	B, C := strings.Repeat("b", 40), strings.Repeat("c", 40)
	P, I, Q := strings.Repeat("1", 64), strings.Repeat("2", 64), strings.Repeat("3", 64)
	file := func(h string) []event.SubjectFile { return []event.SubjectFile{{Path: "x", SHA256: h}} }
	patch := func(h string) *event.SubjectFile { return &event.SubjectFile{Path: "p", SHA256: h} }
	cases := []struct {
		name string
		runs []event.RunSubject
		sub  event.HypothesisSubject
		want bool
	}{
		{"bare B, run at B", []event.RunSubject{{Base: B}}, event.HypothesisSubject{Base: B}, true},
		{"bare B, run at B plus a patch", []event.RunSubject{{Base: B, Patch: P}}, event.HypothesisSubject{Base: B}, false},
		{"bare B, run at B plus a patch and an input", []event.RunSubject{{Base: B, Patch: P, Files: []string{I}}}, event.HypothesisSubject{Base: B}, false},
		{"bare B, merged across two runs", []event.RunSubject{{Base: B, Patch: P}, {Base: C}}, event.HypothesisSubject{Base: B}, false},
		{"bare B, run at C", []event.RunSubject{{Base: C}}, event.HypothesisSubject{Base: B}, false},
		{"bare B, run at B given an input", []event.RunSubject{{Base: B, Files: []string{I}}}, event.HypothesisSubject{Base: B}, true},
		{"B plus patch P, run at B plus P", []event.RunSubject{{Base: B, Patch: P}}, event.HypothesisSubject{Base: B, Patch: patch(P)}, true},
		{"B plus patch P, run at B plus P given an input", []event.RunSubject{{Base: B, Patch: P, Files: []string{I}}}, event.HypothesisSubject{Base: B, Patch: patch(P)}, true},
		{"B plus patch P, run at C plus P", []event.RunSubject{{Base: C, Patch: P}}, event.HypothesisSubject{Base: B, Patch: patch(P)}, false},
		{"B plus patch P, run at B plus another patch", []event.RunSubject{{Base: B, Patch: Q}}, event.HypothesisSubject{Base: B, Patch: patch(P)}, false},
		{"B plus patch P, run at bare B given P's bytes as input", []event.RunSubject{{Base: B, Files: []string{P}}}, event.HypothesisSubject{Base: B, Patch: patch(P)}, false},
		{"B plus input I, run at B given I", []event.RunSubject{{Base: B, Files: []string{I}}}, event.HypothesisSubject{Base: B, Files: file(I)}, true},
		{"B plus input I, run at B whose patch is I", []event.RunSubject{{Base: B, Patch: I}}, event.HypothesisSubject{Base: B, Files: file(I)}, false},
		{"B plus input I, run at B plus P given I", []event.RunSubject{{Base: B, Patch: P, Files: []string{I}}}, event.HypothesisSubject{Base: B, Files: file(I)}, false},
		{"B plus patch P and input I, run at B plus P given I", []event.RunSubject{{Base: B, Patch: P, Files: []string{I}}}, event.HypothesisSubject{Base: B, Patch: patch(P), Files: file(I)}, true},
		{"B plus patch P and input P, run at B plus P given no inputs", []event.RunSubject{{Base: B, Patch: P}}, event.HypothesisSubject{Base: B, Patch: patch(P), Files: file(P)}, false},
		{"B plus patch P and input P, run at B plus P given P", []event.RunSubject{{Base: B, Patch: P, Files: []string{P}}}, event.HypothesisSubject{Base: B, Patch: patch(P), Files: file(P)}, true},
		{"B plus patch P and input I, run at B plus P", []event.RunSubject{{Base: B, Patch: P}}, event.HypothesisSubject{Base: B, Patch: patch(P), Files: file(I)}, false},
		{"files only, old Result", []event.RunSubject{{Files: []string{P, I}}}, event.HypothesisSubject{Files: file(P)}, true},
		{"files only, run at C plus P", []event.RunSubject{{Base: C, Patch: P}}, event.HypothesisSubject{Files: file(P)}, true},
		{"B plus two paths of one input's bytes, run at B given I", []event.RunSubject{{Base: B, Files: []string{I}}}, event.HypothesisSubject{Base: B, Files: append(file(I), event.SubjectFile{Path: "y", SHA256: I})}, true},
		{"no subject", []event.RunSubject{{Base: B}}, event.HypothesisSubject{}, false},
	}
	for _, c := range cases {
		if got := HoldsSubject(c.runs, c.sub); got != c.want {
			t.Errorf("%s: HoldsSubject = %v, want %v", c.name, got, c.want)
		}
	}
}
