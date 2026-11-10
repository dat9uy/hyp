package check

import (
	"errors"
	"strings"
	"testing"
)

// TestRepositoryRecipePinsMatch runs hyp bundle check over this
// repository, so an in-place edit of a cited recipe bundle fails go test.
func TestRepositoryRecipePinsMatch(t *testing.T) {
	report, failures, _, err := Recipes(".")
	if errors.Is(err, ErrNoRepo) {
		t.Skip("not inside a git repository")
	}
	if err != nil {
		t.Fatal(err)
	}
	if failures != 0 {
		t.Fatalf("recipe check found %d failures:\n%s", failures, strings.Join(report, "\n"))
	}
	if last := report[len(report)-1]; strings.HasPrefix(last, "recipe check: 0 trees") {
		t.Fatalf("found no stores: %s", last)
	}
}
