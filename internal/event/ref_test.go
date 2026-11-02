package event

import (
	"encoding/json"
	"testing"
)

func TestParseRef(t *testing.T) {
	for _, c := range []struct {
		in            string
		name, version string
	}{
		{"a@1", "a", "1"},
		{"skill-text-arms@2", "skill-text-arms", "2"},
		{"none", "none", ""},
		{"x@y@3", "x@y", "3"}, // the last @ splits
	} {
		r, err := ParseRef(c.in)
		if err != nil {
			t.Fatalf("ParseRef(%q): %v", c.in, err)
		}
		if r.Name != c.name || r.Version != c.version {
			t.Errorf("ParseRef(%q) = %+v, want %s %s", c.in, r, c.name, c.version)
		}
		if r.String() != c.in {
			t.Errorf("ParseRef(%q).String() = %q", c.in, r.String())
		}
	}
	for _, bad := range []string{"", "a", "a@", "@1"} {
		if r, err := ParseRef(bad); err == nil {
			t.Errorf("ParseRef(%q) = %+v, want an error", bad, r)
		}
	}
}

func TestRefPredicates(t *testing.T) {
	none, _ := ParseRef("none")
	if !none.IsNone() || none.Versioned() || none.IsZero() {
		t.Errorf("none: IsNone %v Versioned %v IsZero %v", none.IsNone(), none.Versioned(), none.IsZero())
	}
	a, _ := ParseRef("a@1")
	if a.IsNone() || !a.Versioned() || a.IsZero() {
		t.Errorf("a@1: IsNone %v Versioned %v IsZero %v", a.IsNone(), a.Versioned(), a.IsZero())
	}
	if !(Ref{}).IsZero() || (Ref{}).String() != "" {
		t.Errorf("zero Ref: IsZero %v String %q", (Ref{}).IsZero(), (Ref{}).String())
	}
	// A versionless bundle keeps the name@ form hyp printed before.
	if s := (Ref{Name: "x"}).String(); s != "x@" {
		t.Errorf("Ref{Name: x}.String() = %q, want x@", s)
	}
}

func TestRefJSON(t *testing.T) {
	type rec struct {
		Ref Ref `json:"ref"`
	}
	for _, s := range []string{"a@1", "none", "x@y@3", ""} {
		in := `{"ref":"` + s + `"}`
		var r rec
		if err := json.Unmarshal([]byte(in), &r); err != nil {
			t.Fatalf("unmarshal %s: %v", in, err)
		}
		out, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != in {
			t.Errorf("round trip %s = %s", in, out)
		}
	}
	var r rec
	if err := json.Unmarshal([]byte(`{"ref":"a@"}`), &r); err == nil {
		t.Errorf(`unmarshal "a@" = %+v, want an error`, r)
	}
}

func TestEvidenceRef(t *testing.T) {
	for _, c := range []struct {
		in  string
		ref string
		ok  bool
	}{
		{"recipes/a@1/x.txt", "a@1", true},
		{"recipes/./a@1/../a@2/x", "a@2", true},
		// a known-answer input name may hold @; only the bundle
		// directory is the ref
		{"recipes/a@1/known-answers/c/input/n@2.txt", "a@1", true},
		{"recipes/a/x", "", false},
		{"recipes/none/x", "", false},
		{".tree/evidence/a@1", "", false},
	} {
		r, ok := EvidenceRef(c.in)
		if ok != c.ok || (ok && r.String() != c.ref) {
			t.Errorf("EvidenceRef(%q) = %v %v, want %s %v", c.in, r, ok, c.ref, c.ok)
		}
	}
}

func TestBadEvidenceRef(t *testing.T) {
	for _, c := range []struct {
		in  string
		bad bool
	}{
		{"recipes/a@1/x", false},
		{"recipes/a/x", false},
		{".tree/evidence/a@/x", false},
		{"recipes/a@/x", true},
		{"recipes/@1/x", true},
	} {
		if err := BadEvidenceRef(c.in); (err != nil) != c.bad {
			t.Errorf("BadEvidenceRef(%q) = %v, want bad %v", c.in, err, c.bad)
		}
	}
}

func TestMeantAsRef(t *testing.T) {
	for s, want := range map[string]bool{"a@1": true, "a@": true, "@1": true, "a": false, "none": false} {
		if got := MeantAsRef(s); got != want {
			t.Errorf("MeantAsRef(%q) = %v, want %v", s, got, want)
		}
	}
}
