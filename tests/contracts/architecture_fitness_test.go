package contracts_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
)

// CONTRACT-ARCH-001: transport must not import repository; domain stays innermost.
func TestContract_Architecture_LayerIsolation(t *testing.T) {
	root := moduleRoot(t)
	fset := token.NewFileSet()

	checkImports(t, fset, filepath.Join(root, tokens.PathTransport), func(imp string) bool {
		if imp == tokens.ModuleImport(tokens.PathRepository) {
			t.Errorf("Violation: transport imports repository directly (%s)", imp)
			return false
		}
		return true
	})

	checkImports(t, fset, filepath.Join(root, tokens.PathDomain), func(imp string) bool {
		for _, rel := range tokens.ArchitectureDomainForbiddenImports {
			if imp == tokens.ModuleImport(rel) {
				t.Errorf("Violation: domain imports forbidden layer %s (%s)", rel, imp)
				return false
			}
		}
		return true
	})
}

func checkImports(t *testing.T, fset *token.FileSet, dir string, validator func(string) bool) {
	t.Helper()
	require.DirExists(t, dir)

	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ImportsOnly)
	require.NoError(t, err, "parse imports in %s", dir)

	for _, pkg := range pkgs {
		for filePath, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if !validator(path) {
					t.Errorf("File: %s", filepath.Base(filePath))
				}
			}
		}
	}
}

// CONTRACT-ARCH-002: architecture layer path tokens stay aligned with repo layout.
func TestContract_Architecture_LayerPathsExist(t *testing.T) {
	root := moduleRoot(t)
	for _, rel := range []string{tokens.PathTransport, tokens.PathDomain, tokens.PathService, tokens.PathRepository} {
		_, err := os.Stat(filepath.Join(root, rel))
		require.NoError(t, err, "layer path %s must exist", rel)
	}
}
