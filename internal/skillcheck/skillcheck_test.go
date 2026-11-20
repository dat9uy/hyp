// Package skillcheck keeps the skills honest: every hyp command they show
// must be one the CLI dispatches, and they use no word the glossary retires.
package skillcheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/channel"
)

// noSkills is why both tests skip when skills/ holds no file: the skills of
// hyp 0.1.0 are kept on the v1 branch.
const noSkills = "no skills under skills/; the hyp 0.1.0 skills are on the v1 branch"

// skillFiles lists every file under skills/, at any depth, whose name ends in
// one of the suffixes. A skill's references and parts live in subfolders, and
// a command they show must exist as much as one SKILL.md shows.
func skillFiles(t *testing.T, suffixes ...string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir("../../skills", func(p string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && p == "../../skills" {
				return filepath.SkipDir
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		for _, s := range suffixes {
			if strings.HasSuffix(p, s) {
				files = append(files, p)
				break
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

var (
	fenceRe   = regexp.MustCompile("(?s)```[a-z]*\n(.*?)```")
	inlineRe  = regexp.MustCompile("`([^`\n]+)`")
	commandRe = regexp.MustCompile(`(?:^|[\s"$(])(?:hyp|"\$HYP"|\$HYP)\s+([a-z][a-z-]*)`)
)

func TestSkillCommandsExist(t *testing.T) {
	known := map[string]bool{}
	for _, c := range channel.Commands() {
		known[c] = true
	}
	files := skillFiles(t, ".md", ".sh")
	if len(files) == 0 {
		t.Skip(noSkills)
	}
	seen := 0
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var code []string
		if strings.HasSuffix(f, ".sh") {
			code = []string{string(data)}
		} else {
			for _, m := range fenceRe.FindAllStringSubmatch(string(data), -1) {
				code = append(code, m[1])
			}
			for _, m := range inlineRe.FindAllStringSubmatch(string(data), -1) {
				code = append(code, m[1])
			}
		}
		for _, c := range code {
			for _, m := range commandRe.FindAllStringSubmatch(c, -1) {
				seen++
				if !known[m[1]] {
					t.Errorf("%s: unknown command `hyp %s`", f, m[1])
				}
			}
		}
	}
	// A skill names only reads (a skill names no write command), so a short
	// skill shows few commands. None at all means the extractor is broken.
	if seen == 0 {
		t.Fatalf("no hyp command found in the skills; the extractor is probably broken")
	}
}

// retiredRe matches the words docs/glossary.md retires for an Experiment's
// rubric. A skill that teaches one makes runners write it into Goal bodies,
// and from there into the spec and the code.
var retiredRe = regexp.MustCompile(`(?i)verdict[ -](rule|check)`)

func TestSkillsUseNoRetiredWord(t *testing.T) {
	files := skillFiles(t, ".md")
	if len(files) == 0 {
		t.Skip(noSkills)
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		// The path from the repository root: every skill file is SKILL.md, so
		// the base name alone does not say which skill holds the word.
		name, err := filepath.Rel("../..", f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if m := retiredRe.FindString(line); m != "" {
				t.Errorf("%s:%d: %q is retired; say the Experiment's rubric, or its body", name, i+1, m)
			}
		}
	}
}
