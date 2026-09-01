# Vitek

Local run only. Invariants: `tests/contracts/` + `db/migrations/`.

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

Env / UI / compose / sqlc: `task tokens:gen` (SoT = `internal/tokens`).

## Tasks

```bash
task dev            # docker compose up -d
task tokens:gen     # regenerate .env.example + web/tokens.css
task sqlc           # generate DB types (needs migrations + queries)
task test:contracts # LRT contracts
task test:all       # full suite + coverage
task check:lrt      # docs allowlist only
```
