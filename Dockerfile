# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS api-build
WORKDIR /src/api
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY main.go ./
COPY internal ./internal
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/monitor ./

FROM node:22-alpine AS web-base
WORKDIR /src/web
ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable && corepack prepare pnpm@11.1.3 --activate

FROM web-base AS web-deps
COPY web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN --mount=type=cache,id=planka-monitor-pnpm,target=/pnpm/store \
    pnpm fetch --frozen-lockfile

FROM web-deps AS web-build
COPY web/package.json ./
RUN --mount=type=cache,id=planka-monitor-pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile --offline --prod=false
COPY web/ ./
ENV NODE_ENV=production
RUN pnpm build

FROM node:22-alpine
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 app
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
ENV LISTEN_ADDR=:8080
ENV MONITOR_API_URL=http://127.0.0.1:8080
COPY --from=api-build /out/monitor /usr/local/bin/monitor
COPY --from=web-build /src/web/.output ./.output
COPY docker/entrypoint.sh /usr/local/bin/monitor-entrypoint
RUN chmod +x /usr/local/bin/monitor-entrypoint \
    && chown -R app:app /app
USER app
EXPOSE 3000 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- --tries=1 http://127.0.0.1:8080/healthz >/dev/null \
    && wget -qO- --tries=1 http://127.0.0.1:3000/ >/dev/null \
    || exit 1
ENTRYPOINT ["/usr/local/bin/monitor-entrypoint"]
