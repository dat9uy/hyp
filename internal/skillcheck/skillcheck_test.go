// Package skillcheck keeps the skills honest: every hyp command they show
// must be one the CLI dispatches.
package skillcheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/cli"
)

var (
	fenceRe   = regexp.MustCompile("(?s)```[a-z]*\n(.*?)```")
	inlineRe  = regexp.MustCompile("`([^`\n]+)`")
	commandRe = regexp.MustCompile(`(?:^|[\s"$(])(?:hyp|"\$HYP"|\$HYP)\s+([a-z][a-z-]*)`)
)

func TestSkillCommandsExist(t *testing.T) {
	known := map[string]bool{}
	for _, c := range cli.Commands() {
		known[c] = true
	}
	files, err := filepath.Glob("../../skills/*/*")
	if err != nil || len(files) == 0 {
		t.Fatalf("no skill files found: %v", err)
	}
	seen := 0
	for _, f := range files {
		if !strings.HasSuffix(f, ".md") && !strings.HasSuffix(f, ".sh") {
			continue
		}
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
					t.Errorf("%s: unknown command `hyp %s`", filepath.Base(f), m[1])
				}
			}
		}
	}
	if seen < 20 {
		t.Fatalf("only %d hyp commands found in the skills; the extractor is probably broken", seen)
	}
}
