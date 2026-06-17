# planka-monitor mini-app

Nuxt 4 Telegram Mini App for the Planka monitoring bot. It is based on
`uPack-dev/NuxtBoilerplate` and keeps the boilerplate structure: `app/`,
`server/`, Pinia stores, SCSS tokens, `Ui*` and `C*` components.

## Environment

```bash
HOST=0.0.0.0
PORT=3000
NUXT_PUBLIC_CLIENT_URL=https://monitor.example.com
NUXT_PUBLIC_CMS_URL=
MONITOR_API_URL=http://monitor:8080
```

`MONITOR_API_URL` is server-only. Browser requests go to same-origin
`/api/v1/*`; Nitro proxies them to the Go monitor and forwards
`X-Telegram-Init-Data`.

For local UI/Planka testing without Telegram init-data, set
`MONITOR_DEV_AUTH_TOKEN` on the Go monitor and open the mini-app with
`?monitorDevToken=<token>`. The app forwards the token as
`X-Monitor-Dev-Auth` only for the current page session; it does not persist
dev auth in cookies or local storage.

## Commands

```bash
corepack enable
corepack prepare pnpm@11.1.3 --activate
pnpm install
pnpm dev
pnpm test
pnpm typecheck
pnpm lint
pnpm stylelint
pnpm mock:tasks -- --count=12 --out=.tmp/mock-tasks.json
pnpm build
```

`mock:tasks` generates the same task shape used by the mini-app fallback. The
tasks screen only tops up sparse API responses when `?mockTasks=1` (or
`NUXT_PUBLIC_MOCK_TASKS=1`) is set; pass `mockTaskCount=12` to choose the
minimum number of visible tasks.
