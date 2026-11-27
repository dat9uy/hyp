package acts

import (
	"slices"
	"testing"
)

func TestRunCommands(t *testing.T) {
	cases := []struct {
		run  string
		want []string
	}{
		{"toy list\n", []string{"toy list"}},
		{"hyp finding add --fact 'a; b' 2>&1; echo \"exit: $?\"; hyp finding list",
			[]string{"hyp finding add --fact 'a; b' 2>&1", "echo \"exit: $?\"", "hyp finding list"}},
		{"# a comment\ntoy add && toy list || toy fold | cat &\n\n",
			[]string{"toy add", "toy list", "toy fold", "cat"}},
		{"", nil},
	}
	for _, c := range cases {
		if got := RunCommands(c.run); !slices.Equal(got, c.want) {
			t.Errorf("RunCommands(%q) = %q, want %q", c.run, got, c.want)
		}
	}
}

func TestCommandWords(t *testing.T) {
	got := commandWords(`HYP_ACTOR=a hyp finding add --fact 'two words' "x"`)
	want := []string{"hyp", "finding", "add", "--fact", "two words", "x"}
	if !slices.Equal(got, want) {
		t.Errorf("commandWords = %q, want %q", got, want)
	}
}
