package gitread

import (
	"bytes"
	"slices"
	"strings"
)

// Parents returns the parents of commit, in order.
func Parents(dir, commit string) ([]string, error) { return parentsOf(dir, commit) }

// Bases returns the base of commit for Goal id whose H0 is h0: the commits
// newer than h0 that commit's own change stands on. It is empty when commit
// stands on no code newer than h0 beside its own change.
//
// The own change is commit itself and each commit newer than h0 it stands
// on that is the runner's: a commit that is not a merge and whose change writes a
// path under goals/<id>/ and none under .tree/, where the owner's acts are;
// a merge one of whose parents is the runner's; and each commit of the
// Goal's rows, which rowBases maps to the base its Result recorded, "" for
// h0. The base is each commit newer than h0 that the walk from commit
// through the own change, and through the bases the rows recorded, reaches
// first, less each one another of them descends from.
func Bases(dir, commit, h0, id string, rowBases map[string]string) ([]string, error) {
	parents, err := revList(dir, commit, "^"+h0)
	if err != nil {
		return nil, err
	}
	if _, newer := parents[commit]; !newer {
		return nil, nil
	}
	own := map[string]bool{}
	var isOwn func(c string) (bool, error)
	isOwn = func(c string) (bool, error) {
		if v, ok := own[c]; ok {
			return v, nil
		}
		ps, newer := parents[c]
		own[c] = false // a walk that comes back to c reads it as not the runner's
		if !newer {
			return false, nil
		}
		if _, row := rowBases[c]; row {
			own[c] = true
			return true, nil
		}
		if len(ps) > 1 {
			for _, p := range ps {
				ok, err := isOwn(p)
				if err != nil {
					return false, err
				}
				if ok {
					own[c] = true
					return true, nil
				}
			}
			return false, nil
		}
		ok, err := writesGoal(dir, c, id)
		own[c] = ok
		return ok, err
	}
	var found []string
	seen := map[string]bool{}
	var walk func(c string, first bool) error
	walk = func(c string, first bool) error {
		if seen[c] {
			return nil
		}
		seen[c] = true
		ps, newer := parents[c]
		if !newer {
			return nil // h0, or a commit h0 holds
		}
		if b, row := rowBases[c]; row && !first {
			if b != "" {
				found = append(found, b)
			}
			return nil
		}
		ok, err := isOwn(c)
		if err != nil {
			return err
		}
		if !ok && !first {
			found = append(found, c)
			return nil
		}
		for _, p := range ps {
			if err := walk(p, false); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(commit, true); err != nil {
		return nil, err
	}
	var top []string
	for _, b := range found {
		if slices.Contains(top, b) {
			continue
		}
		below := false
		for _, other := range found {
			if other == b {
				continue
			}
			ok, err := IsAncestor(dir, b, other)
			if err != nil {
				return nil, err
			}
			if ok {
				below = true
				break
			}
		}
		if !below {
			top = append(top, b)
		}
	}
	return top, nil
}

// writesGoal reports whether the change of commit, not a merge, against its
// parent writes a path under goals/<id>/ and none under .tree/.
func writesGoal(dir, commit, id string) (bool, error) {
	raw, err := Git(dir, "diff-tree", "-r", "--name-only", "-z", commit)
	if err != nil {
		return false, err
	}
	goal := false
	for i, f := range bytes.Split(raw, []byte{0}) {
		p := string(f)
		if p == "" || (i == 0 && p == commit) {
			continue
		}
		if strings.HasPrefix(p, ".tree/") {
			return false, nil
		}
		if strings.HasPrefix(p, "goals/"+id+"/") {
			goal = true
		}
	}
	return goal, nil
}

// revList is the commits the revisions revs name, each with its parents,
// as git rev-list --parents prints them.
func revList(dir string, revs ...string) (map[string][]string, error) {
	out, err := GitStdin(dir, []byte(strings.Join(revs, "\n")+"\n"), append([]string{"rev-list"}, revListForm...)...)
	if err != nil {
		return nil, err
	}
	m := map[string][]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		m[f[0]] = f[1:]
	}
	return m, nil
}
