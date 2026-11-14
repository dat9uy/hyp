// Package recipe reads recipes and hashes them so a Result can be flagged when
// the recipe it followed has changed. A recipe is a bundle directory
// recipes/<name>@<version>/ holding recipe.md plus the scripts it runs. Dirs
// lists the recipes directories searched, from the store up to the repository
// root.
package recipe

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dat9uy/hyp/internal/record"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/herr"
)

// Named is one Standalone command: its label, argv and External: line.
type Named struct {
	Name     string
	Argv     []string
	External External
}

// External is the External: line directly under an Evidence or Standalone
// command in recipe.md, which states what the command reaches outside the
// bundle: its subject arguments and any outside reads or writes, or none
// (ADR 0008). Line is false when there is no such line. hyp only warns
// about it (hyp bundle check); a bundle without one still loads.
type External struct {
	Line bool
	Text string
}

// Stated reports whether the line says something: it exists, and its text
// is neither empty nor the TODO that hyp bundle new writes.
func (e External) Stated() bool { return e.Line && e.Text != "" && e.Text != ExternalTODO }

// ExternalTODO is the text hyp bundle new writes on an External: line it adds.
const ExternalTODO = "TODO"

// Step names a recipe section whose commands hyp runs. Its value is the
// lowercase section heading, as command records and expect.json write it.
type Step string

// The steps, in the order a sandbox runs them, and Standalone, which runs
// outside any sandbox.
const (
	Launch     Step = "launch"
	Doctor     Step = "doctor"
	Drive      Step = "drive"
	Evidence   Step = "evidence"
	Cleanup    Step = "cleanup"
	Standalone Step = "standalone"
)

// Recipe is one resolved recipe. Dir is the bundle directory with its
// symlinks resolved.
type Recipe struct {
	Name, Version, Hash                      string
	From                                     string // the from: line, "<name>@<N> <sha256>", or ""
	Path                                     string // the bundle directory
	Dir                                      string
	Launch, Doctor, Drive, Evidence, Cleanup [][]string // argv; argv[0] is an absolute path inside Dir
	EvidenceExternal                         []External // the External: line of each Evidence command
	Standalone                               []Named    // in file order
	Lib                                      string     // the library Load found it in, Recipes or Rubrics; "" when opened by path
}

// Format is the recipe format a bundle's recipe.md must name to be used
// (ADR 0002).
const Format = "1"

// RepoRoot returns the repository root at or above start (record.FindTop).
func RepoRoot(start string) (string, bool) { return record.FindTop(start) }

// GoFileWarnings returns one warning per .go file in the bundle at dir, by
// its bundle-relative path, when the repository root above dir holds go.mod:
// the go tool reads the bundle's name@N directory as a package outside the
// main module, so go vet ./... fails in that repository. Outside a
// repository, or with no go.mod at its root, it returns nothing.
func GoFileWarnings(dir string) []string {
	top, ok := RepoRoot(dir)
	if !ok || !isFile(filepath.Join(top, "go.mod")) {
		return nil
	}
	var out []string
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() || filepath.Ext(p) != ".go" {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		out = append(out, fmt.Sprintf("%s is a .go file, and go.mod at the repository root makes go vet ./... read this bundle as a package outside the main module; give it a non-.go name, such as %s.txt", rel, filepath.Base(p)))
		return nil
	})
	return out
}

// Dirs lists the recipes directories searched for the store at root, nearest
// first: root/recipes, then recipes/ in each parent up to the repository root,
// the root RepoRoot returns. Outside a repository only root/recipes is
// searched.
func Dirs(root string) []string { return DirsOf(root, Recipes) }

// The two libraries a bundle lives in (ADR 0018): a recipe in recipes/, a
// rubric in rubrics/.
const (
	Recipes = "recipes"
	Rubrics = "rubrics"
)

// DirsOf is Dirs for library lib, Recipes or Rubrics.
func DirsOf(root, lib string) []string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return []string{filepath.Join(root, lib)}
	}
	top, ok := RepoRoot(abs)
	if !ok {
		return []string{filepath.Join(abs, lib)}
	}
	var dirs []string
	for d := abs; ; d = filepath.Dir(d) {
		dirs = append(dirs, filepath.Join(d, lib))
		if d == top {
			return dirs
		}
	}
}

// Resolve checks that the recipe declares the name and version in ref
// ("name@version") and returns the name and the hex sha256 of the bundle
// manifest.
func Resolve(root, ref string) (name, hash string, err error) {
	r, err := Load(root, ref)
	if err != nil {
		return "", "", err
	}
	return r.Name, r.Hash, nil
}

// Load resolves ref ("name@version") in the directories Dirs(root) lists,
// each of which may hold recipes/<name>@<version>/. Every copy found must hash
// the same; the nearest is returned.
func Load(root, ref string) (*Recipe, error) { return LoadFrom(root, Recipes, ref) }

// LoadRubric is Load for a rubric, a bundle in rubrics/ (ADR 0018).
func LoadRubric(root, ref string) (*Recipe, error) { return LoadFrom(root, Rubrics, ref) }

// LoadAny finds ref in either library, recipes/ first. A name is used in
// one library only (ADR 0018), so a ref found in both is refused.
func LoadAny(root, ref string) (*Recipe, error) {
	if _, err := parseRef(ref); err != nil {
		return nil, err
	}
	r, err := Load(root, ref)
	if err != nil && herr.KindOf(err) != herr.BadArguments {
		return nil, err
	}
	rb, rerr := LoadRubric(root, ref)
	switch {
	case r != nil && rb != nil:
		return nil, herr.Refused("bundle %s is in both %s and %s; a name is used in one library only", ref, r.Path, rb.Path)
	case r != nil:
		return r, nil
	case rerr != nil && herr.KindOf(rerr) != herr.BadArguments:
		return nil, rerr
	case rb != nil:
		return rb, nil
	}
	return nil, herr.BadArgs("bundle %s not found; searched %s and %s", ref, strings.Join(Dirs(root), ", "), strings.Join(DirsOf(root, Rubrics), ", "))
}

// LoadFrom is Load in library lib, Recipes or Rubrics.
func LoadFrom(root, lib, ref string) (*Recipe, error) {
	r, err := parseRef(ref)
	if err != nil {
		return nil, err
	}
	dirs := DirsOf(root, lib)
	var found []*Recipe
	for _, d := range dirs {
		r, err := loadIn(d, r)
		if err != nil {
			return nil, err
		}
		if r != nil {
			found = append(found, r)
		}
	}
	if len(found) == 0 {
		return nil, herr.BadArgs("%s %s not found; searched %s", kindWord(lib), ref, strings.Join(dirs, ", "))
	}
	for _, r := range found[1:] {
		if r.Hash != found[0].Hash {
			return nil, herr.Refused("%s", differ(ref, found[0].Path, r.Path))
		}
	}
	found[0].Lib = lib
	return found[0], nil
}

// IsInstrument reports whether r has an Evidence or Standalone command: its
// scripts decide or filter what a Result counts, so it must have known
// answers.
func (r *Recipe) IsInstrument() bool { return len(r.Evidence) > 0 || len(r.Standalone) > 0 }

// HasSteps reports whether r has a command under Launch, Doctor, Drive,
// Evidence or Cleanup; a rubric has none (ADR 0018).
func (r *Recipe) HasSteps() bool {
	return len(r.Launch)+len(r.Doctor)+len(r.Drive)+len(r.Evidence)+len(r.Cleanup) > 0
}

// kindWord names a bundle of library lib: recipe or rubric.
func kindWord(lib string) string {
	if lib == Rubrics {
		return "rubric"
	}
	return "recipe"
}

// HasKnownAnswers reports whether the bundle in dir has known-answers/ with
// at least one case directory.
func HasKnownAnswers(dir string) bool {
	ents, err := os.ReadDir(filepath.Join(dir, "known-answers"))
	if err != nil {
		return false
	}
	for _, e := range ents {
		if e.IsDir() {
			return true
		}
	}
	return false
}

// parseRef parses "name@version" with event.ParseRef and requires each part
// to be one path component, so the ref cannot reach outside recipes/.
func parseRef(ref string) (event.Ref, error) {
	r, err := event.ParseRef(ref)
	if err == nil && componentRe.MatchString(r.Name) && componentRe.MatchString(r.Version) && r.Name != "." && r.Name != ".." {
		return r, nil
	}
	return event.Ref{}, herr.BadArgs("invalid recipe ref %q (want name@version)", ref)
}

// Ref returns the recipe's ref, name@version.
func (r *Recipe) Ref() event.Ref { return event.Ref{Name: r.Name, Version: r.Version} }

// loadIn finds the bundle ref in the recipes directory d, or returns nil when
// d holds none.
func loadIn(d string, ref event.Ref) (*Recipe, error) {
	vdir := filepath.Join(d, ref.String())
	if !isFile(filepath.Join(vdir, "recipe.md")) {
		return nil, nil
	}
	if err := hasFormat(vdir, ref.Name, ref.Version); err != nil {
		return nil, err
	}
	return Open(vdir)
}

// hasFormat refuses the bundle in vdir when its recipe.md does not name the
// recipe format, a required field. Only the frontmatter is read, so the
// refusal comes before any command is parsed (ADR 0002).
func hasFormat(vdir, name, version string) error {
	data, err := os.ReadFile(filepath.Join(vdir, "recipe.md"))
	if err != nil {
		return herr.Env("recipe %s: %v", name, err)
	}
	if frontmatter(data)["format"] != Format {
		return herr.Refused("recipe %s@%s: recipe.md has no format: %s line; copy it with hyp bundle new %s@<next> --from %s@%s, then add its known answers",
			name, version, Format, name, name, version)
	}
	return nil
}

// differ is the refusal for one ref whose copies hash differently.
func differ(ref, a, b string) string {
	return fmt.Sprintf("recipe %s: %s and %s differ; one ref must hash the same everywhere", ref, a, b)
}

// Open reads the library bundle in dir, recipes/<name>@<version>/, and checks
// it as OpenAs(dir, "<name>@<version>"). A directory whose name has no
// @version is refused: a copy hyp makes is opened by its ref instead
// (docs/adr/0005-a-bundle-is-its-ref-and-hash.md).
func Open(dir string) (*Recipe, error) {
	base := filepath.Base(dir)
	if _, err := parseRef(base); err != nil {
		return nil, herr.Refused("recipe bundle %s: the directory name has no @version; a library bundle is recipes/<name>@<N>/", dir)
	}
	return OpenAs(dir, base)
}

// OpenAs reads the bundle in dir, whose recipe.md must have ref's name and
// version whatever the directory is called, and validates every command in
// it. hyp opens every copy it makes this way, so a copy's identity comes from
// the ref the caller holds. Every file other than a command's script is hashed
// content.
func OpenAs(dir, ref string) (*Recipe, error) {
	r, err := parseRef(ref)
	if err != nil {
		return nil, err
	}
	return open(dir, r)
}

// OpenToHash opens dir for hyp bundle hash DIR: a library bundle name@N/ as
// Open does, and any other directory by the name and version its recipe.md
// gives, so a copy hashes as the bundle it came from. The name must match the
// directory and the version is not checked, as HashOnly does
// (docs/adr/0005-a-bundle-is-its-ref-and-hash.md).
func OpenToHash(dir string) (*Recipe, error) {
	if event.MeantAsRef(filepath.Base(dir)) {
		return Open(dir)
	}
	name, version, _, err := HashOnly(dir)
	if err != nil {
		return nil, err
	}
	return open(dir, event.Ref{Name: name, Version: version})
}

// open is OpenAs for a ref already parsed.
func open(dir string, ref event.Ref) (*Recipe, error) {
	name, version := ref.Name, ref.Version
	fm, data, err := readBundle(dir, name)
	if err != nil {
		return nil, err
	}
	if fm["name"] != name {
		return nil, herr.Refused("recipe %s: recipe.md has name %q", ref, fm["name"])
	}
	if fm["version"] != version {
		return nil, herr.Refused("recipe %s: recipe.md has version %s", ref, fm["version"])
	}
	hash, err := bundleHash(dir)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, herr.Env("recipe %s: %v", name, err)
	}
	c, err := parse(data, resolved)
	if err != nil {
		return nil, herr.Refused("recipe %s: %v", name, err)
	}
	return &Recipe{Name: name, Version: version, From: fm["from"], Hash: hash, Path: dir, Dir: resolved,
		Launch: c.launch, Doctor: c.doctor, Drive: c.drive, Evidence: c.evidence, Cleanup: c.cleanup,
		EvidenceExternal: c.evidenceExternal, Standalone: c.standalone}, nil
}

// HashOnly checks the bundle in dir (not a symlink, recipe.md naming the
// directory and, for a versioned directory, its version) and hashes it,
// without parsing a command. A bundle whose recipe.md breaks a command rule
// still hashes, so its pins still verify. A directory without @version hashes
// too, with no version check, so hyp bundle hash DIR accepts any directory.
func HashOnly(dir string) (name, version, hash string, err error) {
	name = filepath.Base(dir)
	dirRef, perr := event.ParseRef(name)
	versioned := perr == nil && dirRef.Versioned()
	dirVersion := ""
	if versioned {
		name, dirVersion = dirRef.Name, dirRef.Version
	}
	fm, _, err := readBundle(dir, name)
	if err != nil {
		return "", "", "", err
	}
	if fm["name"] != name {
		return "", "", "", herr.Refused("recipe name mismatch: file has %q", fm["name"])
	}
	if versioned && fm["version"] != dirVersion {
		return "", "", "", herr.Refused("recipe %s@%s: recipe.md has version %s, the directory says %s", name, dirVersion, fm["version"], dirVersion)
	}
	if hash, err = bundleHash(dir); err != nil {
		return "", "", "", err
	}
	return name, fm["version"], hash, nil
}

// readBundle refuses a symlinked bundle and reads its recipe.md; name only
// labels an error. The caller checks the frontmatter, then hashes the bundle.
func readBundle(dir, name string) (fm map[string]string, data []byte, err error) {
	// WalkDir does not descend a symlinked root, so it would hash as empty.
	if fi, err := os.Lstat(dir); err == nil && fi.Mode()&fs.ModeSymlink != 0 {
		return nil, nil, herr.Refused("recipe bundle %s: is a symlink, not a regular directory", dir)
	}
	data, err = os.ReadFile(filepath.Join(dir, "recipe.md"))
	if err != nil {
		return nil, nil, herr.Env("recipe %s: %v", name, err)
	}
	return frontmatter(data), data, nil
}

// IsScript reports whether file is a regular file that ends in .py or .sh,
// has an execute bit, or starts with #! (docs/replay-spec.md §7.6 reason 1). A missing or unreadable file is not a script.
func IsScript(file string) bool {
	fi, err := os.Stat(file)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}
	if strings.HasSuffix(file, ".py") || strings.HasSuffix(file, ".sh") || fi.Mode()&0o111 != 0 {
		return true
	}
	f, err := os.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, 2)
	n, _ := io.ReadFull(f, head)
	return n == 2 && string(head) == "#!"
}

// Hashes is Scan without the problems.
func Hashes(root string) map[string]string {
	h, _ := Scan(root)
	return h
}

// Scan maps every bundle ref (name@version) in the recipes and rubrics
// directories DirsOf(root, lib) lists to the hex sha256 of its bundle manifest. Hashing a bundle
// never parses it, so a bundle that no longer opens still hashes and its pins
// verify; problems names it. A bundle that cannot be hashed and a ref whose
// copies hash differently are left out, so Results recorded under them read
// as questionable; problems says why, one line each. No recipes gives an
// empty map.
func Scan(root string) (hashes map[string]string, problems []string) {
	type copyAt struct{ hash, path string }
	seen := map[string][]copyAt{}
	var refs []string
	add := func(ref, hash, path string) {
		if _, ok := seen[ref]; !ok {
			refs = append(refs, ref)
		}
		seen[ref] = append(seen[ref], copyAt{hash, path})
	}
	// libOf maps each bundle name to the first library it was found in,
	// so a name in both is named once.
	libOf := map[string]string{}
	for _, lib := range []string{Recipes, Rubrics} {
		for _, d := range DirsOf(root, lib) {
			bundles, _ := filepath.Glob(filepath.Join(d, "*@*", "recipe.md"))
			for _, b := range bundles {
				dir := filepath.Dir(b)
				name, version, hash, err := HashOnly(dir)
				if err != nil {
					problems = append(problems, fmt.Sprintf("recipe %s does not open: %v", dir, err))
					continue
				}
				if l, ok := libOf[name]; !ok {
					libOf[name] = lib
				} else if l != lib && l != "" {
					problems = append(problems, fmt.Sprintf("bundle name %s is used in both %s/ and %s/; a name lives in one library only", name, l, lib))
					libOf[name] = ""
				}
				add(event.Ref{Name: name, Version: version}.String(), hash, dir)
				if _, err := Open(dir); err != nil {
					problems = append(problems, fmt.Sprintf("recipe %s does not open: %v; its pins still verify; copy it with hyp bundle new to use it", dir, err))
				}
			}
		}
	}
	hashes = map[string]string{}
	for _, ref := range refs {
		cs, ok := seen[ref], true
		for _, c := range cs[1:] {
			if c.hash != cs[0].hash {
				problems = append(problems, differ(ref, cs[0].path, c.path))
				ok = false
				break
			}
		}
		if ok {
			hashes[ref] = cs[0].hash
		}
	}
	for _, c := range keptCopies(root) {
		hashes[HeldKey(c.ref, c.hash)] = c.hash
	}
	return hashes, problems
}

// A kept copy is another copy of a bundle ref that the record keeps beside
// the copy at the ref's own path, in <lib>/<record.KeptDir>/<key>/<ref>/:
// two clones each pinned that ref with their own bytes. Only a record's
// snapshot holds one.
type keptCopy struct{ ref, hash, dir string }

// keptCopies lists the kept copies in the libraries DirsOf(root, lib)
// lists, for both libraries. A copy that does not hash is left out.
func keptCopies(root string) []keptCopy {
	var out []keptCopy
	for _, lib := range []string{Recipes, Rubrics} {
		for _, d := range DirsOf(root, lib) {
			mds, _ := filepath.Glob(filepath.Join(d, record.KeptDir, "*", "*@*", "recipe.md"))
			for _, md := range mds {
				dir := filepath.Dir(md)
				name, version, hash, err := HashOnly(dir)
				if err != nil {
					continue
				}
				out = append(out, keptCopy{event.Ref{Name: name, Version: version}.String(), hash, dir})
			}
		}
	}
	return out
}

// HeldKey is the key under which Scan maps a kept copy of bundle ref with
// bundle hash hash to that hash.
func HeldKey(ref, hash string) string { return ref + " " + hash }

// Holds reports whether hashes, as Scan returns them, hold bundle ref with
// bundle hash hash: the copy at the ref's path or a kept copy.
func Holds(hashes map[string]string, ref, hash string) bool {
	if h, ok := hashes[ref]; ok && h == hash {
		return true
	}
	h, ok := hashes[HeldKey(ref, hash)]
	return ok && h == hash
}

// LoadPinned is LoadAny for the copy of ref with bundle hash hash: the copy
// at the ref's path when it hashes so, a kept copy that does otherwise.
func LoadPinned(root, ref, hash string) (*Recipe, error) {
	r, err := LoadAny(root, ref)
	if err == nil && (hash == "" || r.Hash == hash) {
		return r, nil
	}
	for _, c := range keptCopies(root) {
		if c.ref == ref && c.hash == hash {
			if k, kerr := Open(c.dir); kerr == nil {
				return k, nil
			}
		}
	}
	return r, err
}

var (
	sectionRe   = regexp.MustCompile(`^## (Launch|Doctor|Drive|Evidence|Cleanup|Standalone)\s*$`)
	labelRe     = regexp.MustCompile(`^Run ([^\s:]+):(.*)$`)
	nameRe      = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	componentRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	// inputRe is the name rule inside a known answer's input/, which also
	// admits @, so an input can hold a store with its pinned recipes/<name>@<N>/.
	inputRe = regexp.MustCompile(`^[A-Za-z0-9@._-]+$`)
)

// inInput reports whether a bundle-relative slash path lies inside a known
// answer's input directory, known-answers/NAME/input/, below input itself.
func inInput(rel string) bool {
	parts := strings.SplitN(rel, "/", 4)
	return len(parts) == 4 && parts[0] == "known-answers" && parts[2] == "input"
}

// commands is what parse reads from recipe.md.
type commands struct {
	launch, doctor, drive, evidence, cleanup [][]string
	evidenceExternal                         []External
	standalone                               []Named
}

// splitLines reads data as lines without their line ends.
func splitLines(data []byte) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// eachLine calls fn with the index of every line of recipe.md that is
// outside a fenced block and is not a heading, and the command section it
// is in: Launch, Doctor, Drive, Evidence, Cleanup, Standalone, or "".
func eachLine(lines []string, fn func(i int, section Step) error) error {
	var section Step
	fenced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		if m := sectionRe.FindStringSubmatch(line); m != nil {
			section = Step(strings.ToLower(m[1]))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			section = ""
			continue
		}
		if err := fn(i, section); err != nil {
			return err
		}
	}
	return nil
}

// isRun reports whether a trimmed line is a Run: or Run NAME: line.
func isRun(trimmed string) bool {
	return strings.HasPrefix(trimmed, "Run:") || labelRe.MatchString(trimmed)
}

// externalAt is the External: line directly under line i: the next line,
// trimmed and with an optional list marker "- ", starts with External:.
func externalAt(lines []string, i int) External {
	if i+1 >= len(lines) {
		return External{}
	}
	t := strings.TrimPrefix(strings.TrimSpace(lines[i+1]), "- ")
	rest, ok := strings.CutPrefix(t, "External:")
	if !ok {
		return External{}
	}
	return External{Line: true, Text: strings.TrimSpace(rest)}
}

// parse reads the commands of the six sections: Run: lines under Launch,
// Doctor, Drive, Evidence and Cleanup, and labelled Run NAME: lines under
// Standalone. Fenced blocks are skipped. It refuses a Run: line outside those
// sections, a plain Run: under Standalone, a labelled line elsewhere, and a
// label that breaks [a-z][a-z0-9-]* or is used twice. No shell runs a
// command, so shell syntax is refused rather than interpreted, and argv[0]
// must be an executable script inside dir, which is already free of symlinks.
func parse(data []byte, dir string) (commands, error) {
	var c commands
	seen := map[string]bool{}
	lines, err := splitLines(data)
	if err != nil {
		return c, err
	}
	err = eachLine(lines, func(i int, section Step) error {
		n := i + 1
		trimmed := strings.TrimSpace(lines[i])
		var label, rest string
		labelled := false
		if m := labelRe.FindStringSubmatch(trimmed); m != nil {
			labelled, label, rest = true, m[1], m[2]
		} else if r, ok := strings.CutPrefix(trimmed, "Run:"); ok {
			rest = r
		} else {
			return nil
		}
		switch {
		case labelled && section != Standalone:
			return fmt.Errorf("line %d: Run %s: is allowed only under Standalone", n, label)
		case !labelled && section == Standalone:
			return fmt.Errorf("line %d: a Standalone command is named: Run NAME: ./script WORDS", n)
		case !labelled && section == "":
			return fmt.Errorf("line %d: Run: is allowed only under Launch, Doctor, Drive, Evidence, Cleanup and Standalone", n)
		case labelled && !nameRe.MatchString(label):
			return fmt.Errorf("line %d: command name %q must match [a-z][a-z0-9-]*", n, label)
		case labelled && seen[label]:
			return fmt.Errorf("line %d: command name %q is already used", n, label)
		}
		if labelled {
			seen[label] = true
		}
		rest = strings.TrimSpace(rest)
		if strings.ContainsAny(rest, "'\"$\\`|;&<>()*?{}~") {
			return fmt.Errorf("line %d: Run: %s is not a plain command (no quotes, variables or shell operators)", n, rest)
		}
		argv := strings.Fields(rest)
		if len(argv) == 0 {
			return fmt.Errorf("line %d: empty Run:", n)
		}
		script, err := scriptInside(dir, argv[0])
		if err != nil {
			return fmt.Errorf("line %d: Run: %s: %w", n, rest, err)
		}
		argv[0] = script
		switch section {
		case Launch:
			c.launch = append(c.launch, argv)
		case Doctor:
			c.doctor = append(c.doctor, argv)
		case Drive:
			c.drive = append(c.drive, argv)
		case Evidence:
			c.evidence = append(c.evidence, argv)
			c.evidenceExternal = append(c.evidenceExternal, externalAt(lines, i))
		case Cleanup:
			c.cleanup = append(c.cleanup, argv)
		case Standalone:
			c.standalone = append(c.standalone, Named{label, argv, externalAt(lines, i)})
		}
		return nil
	})
	return c, err
}

// scriptInside resolves a relative command path against root (already free of
// symlinks) to an executable regular file strictly inside root that the bundle
// hash covers. Its error names the rule the path breaks and the fix.
func scriptInside(root, cmd string) (string, error) {
	if filepath.IsAbs(cmd) {
		return "", errors.New("a command is a relative path inside the bundle, not an absolute path (put the call in a script in the bundle)")
	}
	if !strings.Contains(cmd, "/") {
		return "", fmt.Errorf("a command is a relative path containing / (write ./%s)", cmd)
	}
	p, err := filepath.EvalSymlinks(filepath.Join(root, cmd))
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("%s does not exist in the bundle", cmd)
	}
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("%s resolves outside the bundle (move the script into the bundle)", cmd)
	}
	rel = filepath.ToSlash(rel)
	if rel != "." && Unhashed(rel) {
		return "", fmt.Errorf("%s is under a dot or __pycache__ path the bundle hash skips, so no pin covers it (move it to a hashed path)", cmd)
	}
	info, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", cmd)
	}
	if info.Mode()&0o111 == 0 {
		return "", fmt.Errorf("%s is not executable (chmod +x %s)", cmd, rel)
	}
	return p, nil
}

// Unhashed reports whether a bundle-relative slash path lies under an entry
// the hash skips: a name starting with "." or a __pycache__ directory. A
// known answer's input (known-answers/NAME/input/...) is hashed whole, dot
// entries and __pycache__ included, since a case input is often a store
// holding a store folder; only its first three components are checked.
func Unhashed(rel string) bool {
	parts := strings.Split(rel, "/")
	if len(parts) > 3 && parts[0] == "known-answers" && parts[2] == "input" {
		parts = parts[:3]
	}
	for _, c := range parts {
		if strings.HasPrefix(c, ".") || c == "__pycache__" {
			return true
		}
	}
	return false
}

// bundleHash hashes the manifest of every regular file in dir: one
// "<sha256>  <path>\n" line per file, sorted by path. Dot-entries and
// __pycache__ directories are skipped; symlinks anywhere and odd names are
// refused.
func bundleHash(dir string) (string, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dir {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		// Checked before the skip: the sandbox copies skipped entries too.
		if !d.IsDir() && !d.Type().IsRegular() {
			return herr.Refused("recipe bundle %s: %s is not a regular file", dir, rel)
		}
		if Unhashed(rel) {
			return nil
		}
		if inInput(rel) {
			if !inputRe.MatchString(d.Name()) {
				return herr.Refused("recipe bundle %s: %s has a character outside [A-Za-z0-9@._-]", dir, rel)
			}
		} else if !componentRe.MatchString(d.Name()) {
			return herr.Refused("recipe bundle %s: %s has a character outside [A-Za-z0-9._-]; only a path under known-answers/NAME/input/ may also hold @", dir, rel)
		}
		if !d.IsDir() {
			paths = append(paths, rel)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	var manifest bytes.Buffer
	for _, rel := range paths {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&manifest, "%s  %s\n", sum(data), rel)
	}
	return sum(manifest.Bytes()), nil
}

func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular()
}

// frontmatter returns the key: value lines between the leading --- fences.
func frontmatter(data []byte) map[string]string {
	out := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	if !sc.Scan() || strings.TrimSpace(sc.Text()) != "---" {
		return out
	}
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "---" {
			break
		}
		if k, v, ok := strings.Cut(line, ":"); ok {
			out[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return out
}

func sum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
