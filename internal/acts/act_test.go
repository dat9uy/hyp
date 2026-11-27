package acts

import (
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
)

// TestBeginChecksActorThenWhyThenTopLevel shows the order of the argument
// checks every write act runs before it reads the store: a bad actor is named
// before a why that is too long, and that why before a folder outside git.
func TestBeginChecksActorThenWhyThenTopLevel(t *testing.T) {
	long := strings.Repeat("w", logstore.MaxWhy+1)
	outside := t.TempDir()
	cases := []struct {
		name, actor, role, why string
		want                   fault.Code
	}{
		{"no actor", "", "owner", long, fault.NoActor},
		{"bad role", "ann", "boss", long, fault.BadRole},
		{"why too long", "ann", "owner", long, fault.WhyTooLong},
	}
	for _, c := range cases {
		if _, err := begin(outside, c.actor, c.role, c.why); fault.CodeOf(err) != c.want {
			t.Errorf("%s: got %v, want %s", c.name, err, c.want)
		}
	}
	if _, err := begin(outside, "ann", "owner", strings.Repeat("w", logstore.MaxWhy)); err == nil || fault.KindOf(err) == fault.BadArguments {
		t.Errorf("a why of the longest length outside git: got %v, want the top level's error", err)
	}
}
