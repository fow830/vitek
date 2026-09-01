package contracts_test

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// CONTRACT-ARCH-001: transport must not import repository; domain stays innermost.
func TestContract_Architecture_LayerIsolation(t *testing.T) {
	root := moduleRoot(t)
	fset := token.NewFileSet()

	checkImports(t, fset, filepath.Join(root, "internal/transport"), func(imp string) bool {
		if strings.Contains(imp, "internal/repository") {
			t.Errorf("Violation: transport imports repository directly (%s)", imp)
			return false
		}
		return true
	})

	checkImports(t, fset, filepath.Join(root, "internal/domain"), func(imp string) bool {
		forbidden := []string{"internal/transport", "internal/repository", "internal/service"}
		for _, f := range forbidden {
			if strings.Contains(imp, f) {
				t.Errorf("Violation: domain imports forbidden layer %s (%s)", f, imp)
				return false
			}
		}
		return true
	})
}

func checkImports(t *testing.T, fset *token.FileSet, dir string, validator func(string) bool) {
	t.Helper()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ImportsOnly)
	if err != nil {
		return
	}

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
