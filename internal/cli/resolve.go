package cli

import (
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/replay"
)

// resolveRef turns a user-typed reference into a node id (spec §2): an exact
// ULID first, then an alias, then a display id.
func resolveRef(p replay.Projection, ref string) (string, error) {
	if _, ok := p.Nodes[ref]; ok {
		return ref, nil
	}
	var byAlias, byDisplay []string
	for id, n := range p.Nodes {
		if n.Alias != "" && n.Alias == ref {
			byAlias = append(byAlias, id)
		}
		if strings.EqualFold(event.DisplayID(n.Kind, id), ref) {
			byDisplay = append(byDisplay, id)
		}
	}
	for _, matches := range [][]string{byAlias, byDisplay} {
		switch len(matches) {
		case 0:
			continue
		case 1:
			return matches[0], nil
		default:
			return "", herr.BadArgs("ambiguous node %s", ref)
		}
	}
	return "", herr.BadArgs("unknown node %s", ref)
}

// display returns the display id of node id, or id itself when unknown.
func display(p replay.Projection, id string) string {
	if n := p.Nodes[id]; n != nil {
		return event.DisplayID(n.Kind, id)
	}
	return id
}
