# Deploy

CI/CD and environment parameters only.

## Build (Day 0)

```bash
task test:all
go build -o bin/api ./cmd/app
go build -o bin/worker ./cmd/worker
```

Container images are added when a `Dockerfile` lands; until then CI ships binaries.

## Environment (staging / production)

Keys are defined in `internal/tokens` (`Env*`). Required in non-local:

| Variable       | Required | Notes                     |
|----------------|----------|---------------------------|
| `APP_ENV`      | yes      | `staging` \| `production` |
| `DATABASE_URL` | yes      | PostgreSQL DSN            |
| `REDIS_URL`    | yes      | Redis DSN                 |
| `HTTP_ADDR`    | no       | default from tokens       |
| `LOG_LEVEL`    | no       | default from tokens       |

Secrets must not be committed.

## Migrations

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

## CI gates

1. `task check:lrt`
2. `task test:all` (race + coverage)
3. `go build` for `cmd/app` and `cmd/worker`

## Release

Annotated tags. Invariant commits include `LRT-VERIFY:` (see `.cursor/rules/lrt.mdc`).
