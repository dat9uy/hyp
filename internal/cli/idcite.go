package cli

import (
	"fmt"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/replay"
)

// textFlags are the flags whose text a write command checks for cites.
var textFlags = []string{"why", "body", "observation", "summary", "limit", "line", "refers", "answer", "accept", "returns"}

// placeholders are the marks that make "X-" plus what follows a stand-in
// for an id that was never filled in.
var placeholders = []string{"<", "{", "…", "...", "XXX"}

// unresolvedCites names each id-shaped token in text that is no node of p
// (spec §5.1). A token starts at a word boundary with G-, Q-, H-, E- or R-
// and goes on with a placeholder mark, or with six Crockford-base32
// upper-case characters that end the word. A token right after ':' cites
// another store ("containment:R-ABC123") and is skipped.
func unresolvedCites(p replay.Projection, text string) []string {
	var out []string
	for i := 0; i+2 <= len(text); i++ {
		if !strings.ContainsRune("GQHER", rune(text[i])) || text[i+1] != '-' {
			continue
		}
		if i > 0 && (isWordByte(text[i-1]) || text[i-1] == ':' || text[i-1] == '-') {
			continue
		}
		rest := text[i+2:]
		if hasPlaceholder(rest) {
			word := text[i:]
			if n := strings.IndexAny(word, " \t\n"); n >= 0 {
				word = word[:n]
			}
			out = append(out, strings.TrimRight(word, ",;:)"))
			continue
		}
		if len(rest) < 6 || !isCrockford(rest[:6]) || len(rest) > 6 && isWordByte(rest[6]) {
			continue
		}
		token := text[i : i+8]
		if _, err := resolveRef(p, "", token); err != nil {
			out = append(out, token)
		}
	}
	return out
}

func hasPlaceholder(s string) bool {
	for _, m := range placeholders {
		if strings.HasPrefix(s, m) {
			return true
		}
	}
	return false
}

// isCrockford reports whether s is upper-case Crockford base32, the
// alphabet of a ULID tail and so of a display id.
func isCrockford(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z') || strings.IndexByte("ILOU", c) >= 0 {
			return false
		}
	}
	return true
}

func isWordByte(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_'
}

// checkCites refuses a write whose text flags cite an id that is no node
// of the locked store. It runs only before an append, never in replay, so
// an old log replays unchanged.
func (w *writer) checkCites() error {
	var bad []string
	for _, name := range textFlags {
		f := w.fs.Lookup(name)
		if f == nil {
			continue
		}
		texts := []string{f.Value.String()}
		if m, ok := f.Value.(*multi); ok {
			texts = *m
		}
		for _, t := range texts {
			for _, tok := range unresolvedCites(w.p, t) {
				bad = append(bad, fmt.Sprintf("--%s names %s, which is no node here", name, tok))
			}
		}
	}
	if len(bad) == 0 {
		return nil
	}
	return herr.BadArgs("%s; cite an id the history holds, the ids of every store of this repository; write <name>:<id> only for a tree outside the repository", strings.Join(bad, "; "))
}
