package contracts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/tokens"
)

// CONTRACT-DAY0-001: deferred product surface must not appear early (no отсебятина).
func TestContract_Day0ForbiddenPackagesAbsent(t *testing.T) {
	root := moduleRoot(t)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, skip := tokens.IgnoredScanDirs[info.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		require.NoError(t, err)
		rel = filepath.ToSlash(rel)
		lower := strings.ToLower(rel)
		for _, frag := range tokens.ForbiddenPackagePathFragments {
			require.NotContainsf(t, lower, frag, "forbidden Day-0 path fragment %q in %s", frag, rel)
		}
		return nil
	})
	require.NoError(t, err)
}

// CONTRACT-DAY0-002: go.mod must not pull Telegram / Redis clients until contracted.
func TestContract_Day0NoPrematureDeps(t *testing.T) {
	root := moduleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, tokens.PathGoMod))
	require.NoError(t, err)
	body := string(raw)
	for _, bad := range []string{
		"github.com/go-telegram",
		"gopkg.in/telebot",
		"github.com/redis/go-redis",
		"github.com/go-redis/redis",
	} {
		require.NotContains(t, body, bad)
	}
}
