package event

import (
	"fmt"
	"strings"
)

// Ref is a Recipe ref, name@version, or the ref none that an Experiment
// following no recipe records. The zero Ref names nothing. Its JSON and text
// form is the string name@N, or "none", so no file format changes.
type Ref struct {
	Name, Version string
}

// RefNone is the ref of an Experiment that follows no recipe.
var RefNone = Ref{Name: RecipeNone}

// ParseRef parses name@version, splitting on the last @, or none. Name and
// version must both be non-empty.
func ParseRef(s string) (Ref, error) {
	if s == RecipeNone {
		return RefNone, nil
	}
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		return Ref{}, fmt.Errorf("invalid recipe ref %q (want name@version)", s)
	}
	return Ref{Name: s[:at], Version: s[at+1:]}, nil
}

// String returns name@version, "none" for RefNone and "" for the zero Ref.
// A bundle whose recipe.md gives no version, opened from an unversioned
// folder, has the ref name@ with an empty version.
func (r Ref) String() string {
	if r.IsNone() || r.IsZero() {
		return r.Name
	}
	return r.Name + "@" + r.Version
}

// MeantAsRef reports whether s holds an @, so it is meant as a ref and must
// parse as one.
func MeantAsRef(s string) bool { return strings.Contains(s, "@") }

// IsNone reports whether r is the ref none.
func (r Ref) IsNone() bool { return r == RefNone }

// IsZero reports whether r names nothing.
func (r Ref) IsZero() bool { return r == Ref{} }

// Versioned reports whether r names a versioned recipe, not none or nothing.
func (r Ref) Versioned() bool { return r.Name != "" && r.Version != "" }

// MarshalText encodes r as its String form.
func (r Ref) MarshalText() ([]byte, error) { return []byte(r.String()), nil }

// UnmarshalText decodes name@version or none; an empty text is the zero Ref.
func (r *Ref) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*r = Ref{}
		return nil
	}
	p, err := ParseRef(string(b))
	if err != nil {
		return err
	}
	*r = p
	return nil
}
