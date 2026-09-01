package contracts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, tokens.PathGoMod)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found from working directory")
		}
		dir = parent
	}
}

// TestContract_LRTDocsPolicy fails when unauthorized prose docs appear on disk.
func TestContract_LRTDocsPolicy(t *testing.T) {
	rootDir := moduleRoot(t)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, skip := tokens.IgnoredScanDirs[info.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if _, isProse := tokens.ProseExtensions[ext]; !isProse {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		require.NoError(t, err)
		relPath = filepath.ToSlash(relPath)

		if _, ok := tokens.AllowedProseDocs[relPath]; !ok {
			assert.Failf(
				t,
				"LRT policy violation (Live Runtime Truth)",
				"forbidden documentation file: %s — remove it; move rules into tests/contracts/ or DB DDL",
				relPath,
			)
		}
		return nil
	})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(rootDir, tokens.PathLRTRule))
	require.NoError(t, err, "LRT cursor rule must exist at %s", tokens.PathLRTRule)

	assertFileEqualsRender(t, tokens.PathLRTRule, tokens.RenderLRTRule())
	assertFileEqualsRender(t, tokens.PathVitekPolicyRule, tokens.RenderVitekPolicyMDC())
}
