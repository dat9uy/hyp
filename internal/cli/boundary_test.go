package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Line 10: package record alone names the store folder and git's, and with
// package sandbox alone runs git. Every non-test Go file of the module is
// read; a file elsewhere that holds either name fails. A file elsewhere
// that can start a program fails too, since git can be run by any name,
// through a shell or by an absolute path: it imports os/exec or calls
// os.StartProcess, syscall.Exec, syscall.ForkExec or syscall.StartProcess.
// A string literal "git", or one ending in "/git", fails as well.
func TestOnlyRecordNamesTheFoldersAndRunsGit(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found from %s: %v", root, err)
	}
	if bad := boundaryFindings(t, root); len(bad) > 0 {
		t.Fatalf("outside package record:\n%s", strings.Join(bad, "\n"))
	}
}

// Each way of running git outside packages record and sandbox is found:
// a bare "git" literal, a shell, an absolute path and a renamed import.
func TestBoundaryFindsEveryWayToRunGit(t *testing.T) {
	cases := map[string]string{
		"literal": "package x\n\nimport \"os/exec\"\n\nfunc f() { exec.Command(\"git\", \"status\").Run() }\n",
		"shell":   "package x\n\nimport \"os/exec\"\n\nfunc f() { exec.Command(\"sh\", \"-c\", \"git rev-parse HEAD\").Run() }\n",
		"path":    "package x\n\nimport \"os/exec\"\n\nfunc f() { exec.Command(\"/usr/bin/git\", \"status\").Run() }\n",
		"renamed": "package x\n\nimport run \"os/exec\"\n\nfunc f() { run.Command(\"sh\").Run() }\n",
		"process": "package x\n\nimport \"os\"\n\nfunc f() { os.StartProcess(\"/bin/sh\", nil, nil) }\n",
		"syscall": "package x\n\nimport \"syscall\"\n\nfunc f() { syscall.Exec(\"/bin/sh\", nil, nil) }\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "go.mod"), "module m\n")
			writeFile(t, filepath.Join(root, "internal", "render", "probe.go"), src)
			writeFile(t, filepath.Join(root, "internal", "sandbox", "ok.go"), src)
			bad := boundaryFindings(t, root)
			if len(bad) == 0 {
				t.Fatalf("internal/render/probe.go runs a program and the boundary test found nothing")
			}
			for _, b := range bad {
				if strings.HasPrefix(b, "internal/sandbox/") {
					t.Errorf("package sandbox may run git, but the boundary test reports %q", b)
				}
			}
		})
	}
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module m\n")
	writeFile(t, filepath.Join(root, "internal", "check", "msg.go"), "package check\n\nvar e = \"not inside a git repository\"\n")
	if bad := boundaryFindings(t, root); len(bad) > 0 {
		t.Fatalf("a message that names git runs nothing, but the boundary test reports %q", bad)
	}
}

// boundaryFindings lists each line-10 breach in the module at root.
func boundaryFindings(t *testing.T, root string) []string {
	t.Helper()
	names := []string{"." + "tree", "." + "git"}
	starts := map[string][]string{"os": {"StartProcess"}, "syscall": {"Exec", "ForkExec", "StartProcess"}}
	var bad []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || (strings.HasPrefix(d.Name(), ".") && p != root) || (filepath.Dir(p) == root && (d.Name() == "investigations" || d.Name() == "recipes" || d.Name() == "rubrics")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		pkg := filepath.ToSlash(filepath.Dir(rel))
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if pkg != "internal/record" {
			for _, n := range names {
				if strings.Contains(string(data), n) {
					bad = append(bad, rel+" holds "+n)
				}
			}
		}
		if pkg == "internal/record" || pkg == "internal/sandbox" {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), p, data, 0)
		if err != nil {
			return err
		}
		local := map[string]string{}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if p == "os/exec" {
				bad = append(bad, rel+" imports os/exec")
			}
			name := path.Base(p)
			if imp.Name != nil {
				name = imp.Name.Name
			}
			local[name] = p
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.BasicLit:
				if v, err := strconv.Unquote(n.Value); err == nil && n.Kind == token.STRING && (v == "git" || strings.HasSuffix(v, "/git")) {
					bad = append(bad, rel+" runs git")
				}
			case *ast.SelectorExpr:
				if id, ok := n.X.(*ast.Ident); ok && slices.Contains(starts[local[id.Name]], n.Sel.Name) {
					bad = append(bad, rel+" starts a program with "+local[id.Name]+"."+n.Sel.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return bad
}

// The boundary scan skips investigations, recipes and rubrics only at the
// module root: a package folder of that name below it is Go source.
func TestBoundaryScansPackageFoldersNamedLikeStoreFolders(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module m\n")
	src := "package p\n\nimport \"os/exec\"\n\nvar s = \"x/." + "tree/events\"\n\nfunc f() { exec.Command(\"git\", \"status\").Run() }\n"
	for _, d := range []string{"recipes", "rubrics", "investigations"} {
		writeFile(t, filepath.Join(root, d, "skip.go"), src)
		writeFile(t, filepath.Join(root, "internal", "check", d, "probe.go"), src)
	}
	bad := strings.Join(boundaryFindings(t, root), "\n")
	for _, d := range []string{"recipes", "rubrics", "investigations"} {
		if !strings.Contains(bad, "internal/check/"+d+"/probe.go") {
			t.Errorf("the boundary scan reports nothing for internal/check/%s/probe.go:\n%s", d, bad)
		}
		if strings.Contains(bad, d+"/skip.go") && !strings.Contains(bad, "internal/check/"+d+"/skip.go") {
			t.Errorf("the boundary scan reads %s/ at the module root:\n%s", d, bad)
		}
	}
}
