package main

import (
	"fmt"
	"os"
	"path/filepath"

	"vitek/internal/tokens"
)

func main() {
	root, err := findModuleRoot()
	if err != nil {
		fatal(err)
	}

	writes := map[string]string{
		tokens.PathEnvExample:  tokens.RenderEnvExample(),
		tokens.PathDesignCSS:   tokens.RenderDesignCSS(),
		tokens.PathCompose:     tokens.RenderComposeYAML(),
		tokens.PathSQLC:        tokens.RenderSQLCYAML(),
		tokens.PathDockerfile:  tokens.RenderDockerfile(),
	}

	for rel, body := range writes {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			fatal(err)
		}
	}
	fmt.Println("tokens: regenerated .env.example, web/tokens.css, docker-compose.yml, sqlc.yaml, Dockerfile")
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, tokens.PathGoMod)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found", tokens.PathGoMod)
		}
		dir = parent
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "tokensgen: %v\n", err)
	os.Exit(1)
}
