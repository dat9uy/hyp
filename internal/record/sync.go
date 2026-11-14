package record

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
)

// Sync joins the record of the repository whose work tree holds dir with
// the record of remote, and pushes the join. The two logs join as a set by
// event id: an events file both sides changed keeps every line of both,
// once per id, in id order; any other file both sides hold with other bytes
// keeps the local copy at its path and the remote's under its sha256 in
// the kept folder of its store, so each event still finds the bytes it
// names, and the conflict is printed to warn; a bundle both hold with
// other files keeps the local copy at its path and the remote's whole in
// the kept folder of its library (KeptDir). The join
// commit has both records as parents, so every subject commit either side
// keeps stays reachable. A remote with no record gets the local one; a
// remote that refuses the ref is an environment problem.
func Sync(dir, remote string, stdout, warn io.Writer) error {
	repo, ok, err := OpenRepo(dir)
	if err != nil {
		return err
	}
	if !ok || repo.Common == "" {
		return herr.Env("hyp sync: no git repository holds %s", dir)
	}
	if _, err := os.Stat(filepath.Join(repo.Common, sandboxMark)); err == nil {
		return herr.Refused("hyp sync: %s is a sandbox clone, which keeps no record", repo.Top)
	}
	if err := os.MkdirAll(repo.hypDir(), 0o755); err != nil {
		return err
	}
	s := &Store{Repo: &Repo{Top: repo.Top, Common: repo.Common, Record: true}}
	release, err := s.Lock()
	if err != nil {
		return err
	}
	defer release()
	local := s.tip
	remoteTip, err := gitOut(repo.Top, nil, nil, "ls-remote", remote, Ref)
	if err != nil {
		return err
	}
	if f := strings.Fields(remoteTip); len(f) > 0 {
		remoteTip = f[0]
	} else {
		remoteTip = ""
	}
	if remoteTip == "" && local == "" {
		fmt.Fprintf(stdout, "no record here or on %s\n", remote)
		return nil
	}
	joined := local
	if remoteTip != "" {
		if _, err := gitOut(repo.Top, nil, nil, "fetch", "--quiet", "--no-tags", remote, Ref); err != nil {
			return err
		}
		fetched, err := gitOut(repo.Top, nil, nil, "rev-parse", "--verify", "FETCH_HEAD^{commit}")
		if err != nil {
			return err
		}
		switch {
		case local == "" || local == fetched || s.Repo.isAncestor(local, fetched):
			joined = fetched
		case s.Repo.isAncestor(fetched, local):
			joined = local
		default:
			if joined, err = s.Repo.join(local, fetched, warn); err != nil {
				return err
			}
		}
		if joined != local {
			if _, err := gitOut(repo.Top, nil, nil, "update-ref", Ref, joined, local); err != nil {
				return err
			}
		}
	}
	if joined != remoteTip {
		if _, err := gitOut(repo.Top, nil, nil, "push", "--quiet", remote, joined+":"+Ref); err != nil {
			return herr.Env("hyp sync: %s refused the record: %s", remote, herr.Text(err))
		}
	}
	fmt.Fprintf(stdout, "record %s synced with %s\n", short(joined), remote)
	return nil
}

func short(c string) string {
	if len(c) > 12 {
		return c[:12]
	}
	return c
}

// join writes the commit whose tree joins the trees of a and b, with a and
// b as its parents.
func (r *Repo) join(a, b string, warn io.Writer) (string, error) {
	ea, err := r.listTree(a)
	if err != nil {
		return "", err
	}
	eb, err := r.listTree(b)
	if err != nil {
		return "", err
	}
	var info bytes.Buffer
	paths := make([]string, 0, len(ea)+len(eb))
	for p := range ea {
		paths = append(paths, p)
	}
	for p := range eb {
		if _, ok := ea[p]; !ok {
			paths = append(paths, p)
		}
	}
	slices.Sort(paths)
	union := make(map[string]entry, len(paths))
	for p, e := range eb {
		union[p] = e
	}
	for p, e := range ea {
		union[p] = e
	}
	kept := map[string]entry{}
	// A bundle both hold at one path with other files is kept whole on
	// each side: the local copy at its path, the remote's in the kept
	// folder of its library, so no file of one joins the other's copy.
	other := map[string]bool{}
	ba := bundles(ea)
	for dir, lb := range bundles(eb) {
		la, ok := ba[dir]
		if !ok || sameListing(la, lb) {
			continue
		}
		kd := keptBundleDir(dir, listingKey(lb))
		for rel, e := range lb {
			other[dir+"/"+rel] = true
			if _, held := union[kd+"/"+rel]; !held {
				kept[kd+"/"+rel] = e
			}
		}
		fmt.Fprintf(warn, "hyp sync: bundle %s differs between the two records; kept the local copy there and the other at %s\n", dir, kd)
	}
	for _, p := range paths {
		x, inA := ea[p]
		y, inB := eb[p]
		if other[p] {
			if !inA {
				continue
			}
			inB = false
		}
		e := x
		switch {
		case !inA:
			e = y
		case inB && x.typ != y.typ:
			// One record names a file at p, the other files below a folder
			// p: the folder stays, and the file is kept by its sha256.
			f := x
			if f.typ != "blob" {
				f = y
			}
			sum, err := r.blobSum(f.sha)
			if err != nil {
				return "", err
			}
			k := keptPath(union, p, sum)
			if _, held := union[k]; !held {
				kept[k] = f
			}
			fmt.Fprintf(warn, "hyp sync: %s is a file in one record and a folder in the other; kept the folder there and the file at %s\n", p, k)
			continue
		case inB && x.typ == "blob" && y.typ == "blob" && x.sha != y.sha:
			if isEventsFile(p) {
				data, err := r.joinLines(x.sha, y.sha)
				if err != nil {
					return "", err
				}
				sha, err := gitOut(r.Top, nil, data, "hash-object", "-w", "--stdin")
				if err != nil {
					return "", err
				}
				e = entry{mode: "100644", typ: "blob", sha: sha}
			} else {
				sum, err := r.blobSum(y.sha)
				if err != nil {
					return "", err
				}
				k := keptPath(union, p, sum)
				if _, held := union[k]; !held {
					kept[k] = y
				}
				fmt.Fprintf(warn, "hyp sync: %s differs between the two records; kept the local copy there and the other at %s\n", p, k)
			}
		}
		if e.typ != "blob" {
			continue
		}
		fmt.Fprintf(&info, "%s %s\t%s\n", e.mode, e.sha, p)
	}
	for _, k := range slices.Sorted(maps.Keys(kept)) {
		fmt.Fprintf(&info, "%s %s\t%s\n", kept[k].mode, kept[k].sha, k)
	}
	idx, err := os.CreateTemp(r.hypDir(), "index-*")
	if err != nil {
		return "", err
	}
	idx.Close()
	defer os.Remove(idx.Name())
	env := []string{"GIT_INDEX_FILE=" + idx.Name()}
	if _, err := gitOut(r.Top, env, nil, "read-tree", "--empty"); err != nil {
		return "", err
	}
	if _, err := gitOut(r.Top, env, info.Bytes(), "update-index", "--add", "--index-info"); err != nil {
		return "", err
	}
	tree, err := gitOut(r.Top, env, nil, "write-tree")
	if err != nil {
		return "", err
	}
	return r.commitTree(tree, []string{a, b}, "hyp: sync")
}

// isEventsFile reports whether a record path is a store's events file:
// <store>/.tree/events/<actor>.jsonl. A file a store keeps deeper, such as
// evidence whose path holds a store folder of its own, or a file of a
// bundle, is not one, so sync keeps both copies of it instead of joining
// its lines.
func isEventsFile(p string) bool {
	_, rest, ok := storeOf(p)
	file, found := strings.CutPrefix(rest, "events/")
	return ok && found && strings.HasSuffix(file, ".jsonl") && !strings.Contains(file, "/")
}

// joinLines joins two events files as a set by id: each id once, the first
// side's line kept when both hold it, every line in id order and, for one
// id, in byte order. A line with no id is kept once by its bytes.
func (r *Repo) joinLines(a, b string) ([]byte, error) {
	type line struct{ id, raw string }
	var lines []line
	byID := map[string]bool{}
	byRaw := map[string]bool{}
	err := r.readBlobs([]string{a, b}, func(_ string, data []byte) error {
		for _, l := range bytes.Split(data, []byte{'\n'}) {
			l = bytes.TrimSpace(l)
			if len(l) == 0 {
				continue
			}
			var head struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(l, &head)
			if head.ID != "" {
				if byID[head.ID] {
					continue
				}
				byID[head.ID] = true
			} else {
				if byRaw[string(l)] {
					continue
				}
				byRaw[string(l)] = true
			}
			lines = append(lines, line{head.ID, string(l)})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(lines, func(x, y line) int {
		if c := strings.Compare(x.id, y.id); c != 0 {
			return c
		}
		return strings.Compare(x.raw, y.raw)
	})
	var out bytes.Buffer
	for _, l := range lines {
		out.WriteString(l.raw)
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

// blobSum is the hex sha256 of the bytes of the blob sha.
func (r *Repo) blobSum(sha string) (string, error) {
	var sum string
	err := r.readBlobs([]string{sha}, func(_ string, data []byte) error {
		h := sha256.Sum256(data)
		sum = hex.EncodeToString(h[:])
		return nil
	})
	return sum, err
}
