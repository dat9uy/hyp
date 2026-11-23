package spec

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
)

// faultConsts maps each constant name of package fault to its value.
func faultConsts(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	fs := token.NewFileSet()
	for _, name := range []string{"fault.go", "codes.go"} {
		f, err := parser.ParseFile(fs, filepath.Join("..", "fault", name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.CONST {
				continue
			}
			for _, s := range g.Specs {
				v := s.(*ast.ValueSpec)
				for i, n := range v.Names {
					if i < len(v.Values) {
						if lit, ok := v.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							out[n.Name], _ = strconv.Unquote(lit.Value)
						}
					}
				}
			}
		}
	}
	return out
}

// TestGuardsHoldEveryRefusal: the guard entries name exactly the pairs of a
// guard and a code that hyp's code refuses with, so a refusal hyp gives is
// always the one hyp spec prints, and no entry names a refusal hyp never
// gives.
func TestGuardsHoldEveryRefusal(t *testing.T) {
	consts := faultConsts(t)
	re := regexp.MustCompile(`fault\.Refuse\(fault\.([A-Za-z0-9]+), fault\.([A-Za-z0-9]+)`)
	inCode := map[string]bool{}
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range re.FindAllStringSubmatch(string(data), -1) {
			g, gok := consts[m[1]]
			c, cok := consts[m[2]]
			if !gok || !cok {
				continue // a code held in a variable, such as a decode's
			}
			inCode[g+" "+c] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(inCode) < 50 {
		t.Fatalf("found only %d refusals in hyp's code", len(inCode))
	}
	inSpec := map[string]bool{}
	for _, g := range Guards {
		for _, c := range g.Codes {
			inSpec[string(g.Name)+" "+string(c)] = true
		}
	}
	for k := range inCode {
		if !inSpec[k] {
			t.Errorf("hyp refuses with %q, and no guard entry names it", k)
		}
	}
	for k := range inSpec {
		if !inCode[k] {
			t.Errorf("a guard entry names %q, and hyp never refuses with it", k)
		}
	}
}

// TestEveryGuardHasOneEntry: each guard is one entry, and each guard an act or
// a state entry names has one.
func TestEveryGuardHasOneEntry(t *testing.T) {
	seen := map[fault.Guard]int{}
	for _, g := range Guards {
		seen[g.Name]++
		if seen[g.Name] > 1 {
			t.Errorf("guard %q has two entries", g.Name)
		}
	}
	for _, c := range Commands {
		for _, g := range c.Guards {
			if seen[g] == 0 {
				t.Errorf("%s names guard %q, which has no entry", c.Name, g)
			}
		}
	}
	for _, s := range States {
		for _, st := range s.Stops {
			if seen[st.Guard] == 0 {
				t.Errorf("state %s names guard %q, which has no entry", s.Name, st.Guard)
			}
		}
	}
}

// TestStatesNameActsAndGuardsOfTheirActs: a state stops an act only by a
// guard that act checks, and allows and stops only acts that exist.
func TestStatesNameActsAndGuardsOfTheirActs(t *testing.T) {
	acts := map[string]Command{}
	for _, c := range Commands {
		acts[c.Name] = c
	}
	for _, s := range States {
		for _, a := range s.Allows {
			if _, ok := acts[a]; !ok {
				t.Errorf("state %s allows %q, which is no act", s.Name, a)
			}
		}
		for _, st := range s.Stops {
			c, ok := acts[st.Act]
			if !ok {
				t.Errorf("state %s stops %q, which is no act", s.Name, st.Act)
				continue
			}
			if !slices.Contains(c.Guards, st.Guard) {
				t.Errorf("state %s stops %s by %q, which %s does not check", s.Name, st.Act, st.Guard, st.Act)
			}
			if st.When == "" && slices.Contains(s.Allows, st.Act) {
				t.Errorf("state %s both allows and stops every %s", s.Name, st.Act)
			}
		}
	}
}

// TestEveryEventKindHasOneEntry: each event type has one entry, with a home.
func TestEveryEventKindHasOneEntry(t *testing.T) {
	seen := map[string]bool{}
	for _, k := range events.Kinds {
		if seen[k.Type] {
			t.Errorf("event type %s has two entries", k.Type)
		}
		seen[k.Type] = true
		if want := map[bool]string{true: events.HomeGit, false: events.HomeTrace}[k.ChangesState]; k.Home() != want {
			t.Errorf("event type %s is stored in %q, want %q", k.Type, k.Home(), want)
		}
	}
}

// TestExitCodes: each kind of failure has its own exit code.
func TestExitCodes(t *testing.T) {
	for k, want := range map[fault.Kind]int{fault.Refusal: 1, fault.BadArguments: 2, fault.Environment: 3} {
		if got := ExitCode(k); got != want {
			t.Errorf("ExitCode(%v) = %d, want %d", k, got, want)
		}
	}
}
