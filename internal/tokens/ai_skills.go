package tokens

import (
	"fmt"
	"os"
	"path/filepath"
)

// RenderAISkills writes the Cursor anti-drift policy rule from tokens SoT.
func RenderAISkills(root string) error {
	targetPath := filepath.Join(root, PathVitekPolicyRule)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("failed to create rules dir: %w", err)
	}
	return os.WriteFile(targetPath, []byte(renderVitekPolicyMDC()), 0o644)
}

func renderVitekPolicyMDC() string {
	return `---
description: Vitek Architecture & Anti-Drift Policy
globs: **/*
---
# Vitek Development Rules

1. NEVER edit rendered infrastructure files directly (Dockerfile, .env.example, docker-compose.yml).
   Always edit 'internal/tokens/*.go' and run 'task tokens:gen'.
2. NEVER write DB models or DTOs manually.
   Modify 'db/migrations/*.sql' and 'db/queries/*.sql', then run 'task sqlc'.
3. Transport layer (HTTP/gRPC) must NOT import 'internal/repository' directly — calls must go through 'internal/service'.
4. Do NOT leave draft files (.md, .txt) in root. Only README.md and DEPLOY.md are allowed.
5. Every new endpoint or core logic change requires a contract test in 'tests/contracts/'.
`
}

// RenderVitekPolicyMDC returns the vitek-policy.mdc body (contracts / drift checks).
func RenderVitekPolicyMDC() string {
	return renderVitekPolicyMDC()
}
