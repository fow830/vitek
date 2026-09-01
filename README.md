# Vitek

Local run only. **SoT = `tests/contracts/` + `db/migrations/` + `internal/tokens`.** No side docs, no отсебятина.

Product: **Витёк** — multi-service Avito SaaS. Domain: vitek.tech. First shipped service: similar listings search. Auth: Magic Link (HTTP + sessions + admin face).

## Requirements

- Go 1.26+
- Docker / Docker Compose
- [Task](https://taskfile.dev)
- [sqlc](https://docs.sqlc.dev) (after first migration)
- [golang-migrate](https://github.com/golang-migrate/migrate)

## Quick start

```bash
cp .env.example .env
task dev
task test:contracts
```

Env / CSS / compose / sqlc / Dockerfile / admin face: `task tokens:gen` (SoT = `internal/tokens`).

## Tasks

```bash
task dev            # docker compose up -d (Postgres)
task tokens:gen     # regenerate derived files from tokens
task sqlc           # generate DB types
task test:contracts # LRT contracts (only truth that matters)
task test:all       # full suite + coverage
task check:lrt      # docs allowlist only
```
