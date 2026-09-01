# Deploy

CI/CD and environment parameters only.

## Build

```bash
task test:all
task tokens:gen
go build -o bin/api ./cmd/app
go build -o bin/worker ./cmd/worker
```

### Container images

```bash
docker build -t vitek-api:local --target api .
docker build -t vitek-worker:local --target worker .
```

Images and stages are defined by `internal/tokens` → `task tokens:gen` → `Dockerfile`.

## Run containers

```bash
docker run --rm -p 8080:8080 \
  -e APP_ENV=staging \
  -e DATABASE_URL="$DATABASE_URL" \
  -e REDIS_URL="$REDIS_URL" \
  vitek-api:local

docker run --rm \
  -e APP_ENV=staging \
  -e DATABASE_URL="$DATABASE_URL" \
  -e REDIS_URL="$REDIS_URL" \
  -e WORKER_TICK=5s \
  vitek-worker:local
```

## Environment (staging / production)

Keys are defined in `internal/tokens` (`Env*`). Required in non-local:

| Variable       | Required | Notes                     |
|----------------|----------|---------------------------|
| `APP_ENV`      | yes      | `staging` \| `production` |
| `DATABASE_URL` | yes      | PostgreSQL DSN            |
| `REDIS_URL`    | yes      | Redis DSN                 |
| `HTTP_ADDR`    | no       | default from tokens       |
| `LOG_LEVEL`    | no       | default from tokens       |
| `WORKER_TICK`  | no       | default from tokens       |

Secrets must not be committed.

## Migrations

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

## HTTP surface (Phase B)

| Method | Path                 | Notes                |
|--------|----------------------|----------------------|
| GET    | `/healthz`           | DB ping              |
| POST   | `/v1/users`          | body: `{email, plan_type}`; 400 bad email, 409 duplicate |
| POST   | `/v1/tasks`          | create task (limits) |
| GET    | `/v1/proxies/active` | ACTIVE proxies only  |

## CI gates

1. `task check:lrt`
2. `task test:all` (race + coverage)
3. `docker build` for targets `api` and `worker`

## Release

Annotated tags. Invariant commits include `LRT-VERIFY:` (see `.cursor/rules/lrt.mdc`).
