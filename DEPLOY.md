# Deploy

CI/CD and environment parameters only. Image tags, ports, env keys, and HTTP paths are defined in `internal/tokens` and verified by `tests/contracts`.

## Production (vdserv)

- Host: `vdserv` (`83.217.192.46`, SSH alias `vdserv`, key `~/.ssh/id_vdback_agent`)
- Domain: `https://vitek.tech` (+ `www`)
- App root: `/opt/vitek` (compose + `.env` + `src/` + `nginx/`)
- Shares TLS edge with VapeDetector via `vd-nginx` (`/opt/vitek/nginx/vitek.conf` mounted into conf.d)
- DB hostname in DSN must be `vitek-postgres` (not `postgres` — name clash on `vapedetector_vd-network`)

```bash
ssh vdserv
cd /opt/vitek/src && git pull --ff-only
docker build -t vitek-api:prod --target api .
docker build -t vitek-worker:prod --target worker .
cd /opt/vitek && docker compose up -d
docker run --rm --network vitek_net -v /opt/vitek/src/db/migrations:/migrations \
  migrate/migrate -path=/migrations -database "$DATABASE_URL" up
```

Smoke:

```bash
curl -s https://vitek.tech/healthz
curl -sI https://vitek.tech/admin
curl -sI https://vitek.tech/tokens.css
```

## Build (local)

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
