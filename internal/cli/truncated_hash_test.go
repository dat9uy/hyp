package cli

import (
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

const fullSHA = "c47a38dee2456cc6172dff28472052c0f7b28d7baf310af6e83b5ca655d10a24"

// A line never changes, so a hash it names must be whole: hyp add goal
// refuses a line that names a run of 8 to 63 hex digits followed by an
// ellipsis, and accepts the same line with all 64 digits.
func TestAddGoalRefusesATruncatedHash(t *testing.T) {
	setup(t)
	for _, short := range []string{fullSHA[:8] + "…", fullSHA[:8] + "...", fullSHA[:63] + "…", strings.ToUpper(fullSHA[:12]) + "..."} {
		line := "[P] graded by eval_skill.py (sha256 " + short + ") with 0 errors"
		r := hyp(t, "add", "goal", "--title", "refused goal", "--actor", "o", "--line", "[N] fine", "--line", line, "--why", "w")
		want(t, r, 1, "line 2 names a hash cut to ")
		want(t, r, 1, "name the file by its whole sha256 (64 hex digits)")
	}
	notIn(t, "refused goal")

	r := hyp(t, "add", "goal", "--title", "G3", "--actor", "o", "--line", "[P] graded by eval_skill.py (sha256 "+fullSHA+") with 0 errors", "--why", "w")
	want(t, r, 0, "line ")
	// A hex run outside 8 to 63 digits, or one with no ellipsis, is not a cut hash.
	for _, ok := range []string{"commit 1aba8da… on main", fullSHA + "…", "sha256 " + fullSHA[:12] + " in a list", "a tag like abcdefg..."} {
		want(t, hyp(t, "add", "goal", "--title", "G4", "--actor", "o", "--line", ok, "--why", "w"), 0, "line ")
	}
}

// Only the write refuses: a log that already holds a Goal whose line names a
// cut hash, written by an older hyp, still replays.
func TestALogWithATruncatedHashStillReplays(t *testing.T) {
	setup(t)
	text := "[P] graded by eval_skill.py (sha256 " + fullSHA[:8] + "…)"
	id := rawEvent(t, "o", event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindGoal, Title: "old Goal", Lines: []event.GoalLine{{Text: text}}}, "written by an older hyp")
	want(t, hyp(t, "check"), 0, "check: 0 errors")
	want(t, hyp(t, "show", event.DisplayID(event.KindGoal, id)+"#1"), 0, text)
}
