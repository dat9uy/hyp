package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Line 10: package record alone names the store folder and git's, and with
// package sandbox alone runs git. Every non-test Go file of the module is
// read; a file elsewhere that holds either name, or the string literal
// "git", fails.
func TestOnlyRecordNamesTheFoldersAndRunsGit(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found from %s: %v", root, err)
	}
	names := []string{"." + "tree", "." + "git"}
	var bad []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || (strings.HasPrefix(d.Name(), ".") && p != root) || d.Name() == "investigations" || d.Name() == "recipes" || d.Name() == "rubrics" {
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
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && strings.Trim(lit.Value, "\"`") == "git" {
				bad = append(bad, rel+" runs git")
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(bad) > 0 {
		t.Fatalf("outside package record:\n%s", strings.Join(bad, "\n"))
	}
}
