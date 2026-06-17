# planka-monitor

Single-repository Planka monitoring service: Go API/bot plus Nuxt Telegram
mini-app in one production container.

The container starts:

- Go monitor API on `:8080`
- Nuxt web app on `:3000`
- Nuxt server-side proxy `/api/v1/* -> http://127.0.0.1:8080/api/v1/*`

## Layout

```text
.
├── Dockerfile
├── main.go
├── internal/
└── web/
```

## Required environment

See `.env.example` for the full list.

Required for the API:

- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`
- `PLANKA_BASE_URL`
- `DATABASE_URL`

Common production values:

- `REDIS_URL=redis://redis:6379/0`
- `MONITOR_WEBAPP_URL=https://monitor.example.com`
- `NUXT_PUBLIC_CLIENT_URL=https://monitor.example.com`
- `MONITOR_API_URL=http://127.0.0.1:8080`
- `PLANKA_INTERNAL_URL=http://planka:1337`

Webhook URL for Planka inside compose:

```yaml
WEBHOOKS=[{"url":"http://monitor:8080/webhook","accessToken":"${WEBHOOK_SECRET}"}]
```

## Development checks

```bash
go test ./...
cd web
pnpm install
pnpm test
```

## Publishing

`.github/workflows/publish.yml` runs tests and publishes the Docker image to
GitHub Container Registry:

```text
ghcr.io/<owner>/<repo>:latest
ghcr.io/<owner>/<repo>:<branch>
ghcr.io/<owner>/<repo>:<tag>
ghcr.io/<owner>/<repo>:sha-<sha>
```

The workflow publishes on pushes to `main`, tags matching `v*.*.*`, and manual
workflow dispatch. Pull requests run tests only.
