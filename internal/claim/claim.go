// Package claim reports where a Result's text disagrees with the files it
// cites: a typed verdict that a cited `verdict:` line contradicts, and a
// commit or hash in the observation or a limit that no cited file holds. These are
// warnings: hyp never reads a rubric's prose, and replay never reads files
// (ADR 0004, spec §7.6).
package claim

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// maxRead caps the bytes read from one cited file.
const maxRead = 16 << 20

var (
	verdictLine = regexp.MustCompile(`(?m)^\s*verdict:\s*(\S+?)=(supports|contradicts|inconclusive)\b`)
	hexToken    = regexp.MustCompile(`\b[0-9a-f]{7,64}\b`)
)

// Warnings lists the warn lines for Result id: one per Hypothesis whose
// typed verdict no cited `verdict:` line for it agrees with, while some
// cited line names it, and one per hex token of 7 or more characters (a
// commit or hash) in the observation or a limit that no cited file holds. Only local
// cited files are read; a URL is skipped, and a Result citing no readable
// file gives no hex warning.
func Warnings(root string, p replay.Projection, id string) []string {
	r := p.Nodes[id]
	if r == nil || r.Kind != event.KindResult {
		return nil
	}
	var texts []string
	for _, e := range r.Evidence {
		if strings.HasPrefix(e, "http://") || strings.HasPrefix(e, "https://") {
			continue
		}
		if b, ok := read(filepath.Join(root, e)); ok {
			texts = append(texts, b)
		}
	}
	if len(texts) == 0 {
		return nil
	}
	disp := event.DisplayID(event.KindResult, id)
	var out []string

	// said maps each Hypothesis to the words cited verdict lines give it.
	said := map[string][]string{}
	for _, t := range texts {
		for _, m := range verdictLine.FindAllStringSubmatch(t, -1) {
			if h := hypothesis(p, r.Store, m[1]); h != "" && !slices.Contains(said[h], m[2]) {
				said[h] = append(said[h], m[2])
			}
		}
	}
	for _, h := range r.VerdictOrder {
		words := said[h]
		if len(words) == 0 || slices.Contains(words, r.Verdicts[h]) {
			continue
		}
		sort.Strings(words)
		out = append(out, fmt.Sprintf("warn: verdict mismatch %s: %s typed %s, its cited files say %s",
			disp, label(p, h), r.Verdicts[h], strings.Join(words, ", ")))
	}

	// fields are the texts that state facts about the run: the observation
	// and each limit. A reading quotes the owner, so it is not read.
	fields := []struct{ name, text string }{{"the observation", r.Observation}}
	for i, l := range r.Limits {
		fields = append(fields, struct{ name, text string }{"limit " + disp + "#" + strconv.Itoa(i+1), l})
	}
	for _, f := range fields {
		seen := map[string]bool{}
		for _, tok := range hexToken.FindAllString(f.text, -1) {
			if seen[tok] || !strings.ContainsAny(tok, "0123456789") || !strings.ContainsAny(tok, "abcdef") {
				continue
			}
			seen[tok] = true
			if !slices.ContainsFunc(texts, func(t string) bool { return strings.Contains(t, tok) }) {
				out = append(out, fmt.Sprintf("warn: uncited %s: %s is in %s but in no cited file", disp, tok, f.name))
			}
		}
	}
	return out
}

// hypothesis resolves a verdict line's label, an alias or an id, to a
// Hypothesis id of store, the store of the Result that cites the line, or
// "". An alias is local to its store, and arms reuse labels such as h-null.
func hypothesis(p replay.Projection, store, lbl string) string {
	for id, n := range p.Nodes {
		if n.Kind != event.KindHypothesis || n.Store != store {
			continue
		}
		if n.Alias == lbl || id == lbl || event.DisplayID(event.KindHypothesis, id) == lbl {
			return id
		}
	}
	return ""
}

func label(p replay.Projection, h string) string {
	s := event.DisplayID(event.KindHypothesis, h)
	if a := p.Nodes[h].Alias; a != "" {
		s += " (" + a + ")"
	}
	return s
}

func read(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	if st, err := f.Stat(); err != nil || !st.Mode().IsRegular() {
		return "", false
	}
	b := make([]byte, maxRead)
	n, _ := f.Read(b)
	return string(b[:n]), true
}
