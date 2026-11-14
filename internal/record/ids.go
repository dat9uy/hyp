package record

import (
	"bytes"
	"encoding/json"
	"path/filepath"
)

// StoreIDs lists the event id of each line of every store's events files
// the record holds at the tip Lock or discovery read, keyed by store name.
// A line with no id lists nothing. Outside a record repository it is nil.
func (s *Store) StoreIDs() (map[string][]string, error) {
	if !s.RecordMode() || s.tip == "" {
		return nil, nil
	}
	owner := map[string]string{}
	var shas []string
	for p, e := range s.entries {
		if e.typ != "blob" || !isEventsFile(p) {
			continue
		}
		name, _, _ := storeOf(p)
		if _, seen := owner[e.sha]; !seen {
			shas = append(shas, e.sha)
		}
		// One blob can be the events file of two stores; each lists its ids.
		owner[e.sha] += "\x00" + name
	}
	out := map[string][]string{}
	err := s.Repo.readBlobs(shas, func(sha string, data []byte) error {
		ids := lineIDs(data)
		for _, name := range bytes.Split([]byte(owner[sha])[1:], []byte{0}) {
			out[string(name)] = append(out[string(name)], ids...)
		}
		return nil
	})
	return out, err
}

// lineIDs is the id of each line of an events file that holds one.
func lineIDs(data []byte) []string {
	var ids []string
	for _, l := range bytes.Split(data, []byte{'\n'}) {
		l = bytes.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		var head struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(l, &head) == nil && head.ID != "" {
			ids = append(ids, head.ID)
		}
	}
	return ids
}

// Sibling is the store the record names name, read at the same tip as s,
// so a write that holds s's Lock reads every store it adds to at one tip.
func (s *Store) Sibling(name string) *Store {
	t := &Store{Repo: s.Repo, Name: name, tip: s.tip, entries: s.entries}
	if s.Repo != nil {
		t.Dir = filepath.Join(s.Repo.Top, filepath.FromSlash(name))
	}
	t.InRecord = t.entries != nil && t.hasStore(name)
	return t
}
