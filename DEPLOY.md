# Deploy

CI/CD and environment parameters only. Image tags, ports, env keys, and HTTP paths are defined in `internal/tokens` and verified by `tests/contracts`.

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

## Run containers

```bash
docker run --rm -p 8080:8080 \
  -e APP_ENV=staging \
  -e DATABASE_URL="$DATABASE_URL" \
  vitek-api:local
```

## Environment

| Variable       | Required | Notes                     |
|----------------|----------|---------------------------|
| `APP_ENV`      | yes      | `staging` \| `production` |
| `DATABASE_URL` | yes      | PostgreSQL DSN            |
| `HTTP_ADDR`    | no       | default from tokens       |
| `WORKER_TICK`  | no       | default from tokens       |

## Migrations

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

## HTTP surface

Paths from `tokens.HTTPPathAllowlist` (contracted): healthz, users, tasks, proxies/active, Magic Link request/consume, admin proxies/avito, `/admin` + `/admin/sse`, `/tokens.css`.

## CI gates

1. `task check:lrt`
2. `task test:all`
3. `docker build` for `api` and `worker`
