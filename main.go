package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"planka-monitor/internal/planka"
	"planka-monitor/internal/server"
	"planka-monitor/internal/store"
	"planka-monitor/internal/telegram"
)

const (
	defaultListenAddr     = ":8080"
	defaultPlankaInternal = "http://planka:1337"
	defaultCacheLimit     = 400
	redisPingTimeout      = 3 * time.Second
)

func main() {
	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	st, err := store.New(ctx, cfg.dsn, cfg.envAdmins)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	pk := planka.New(cfg.plankaInternalURL, cfg.plankaAPIEmail, cfg.plankaAPIPassword)

	cache := setupNotifCache(ctx, cfg.redisURL, cfg.cacheLimit)

	tg := telegram.New(cfg.token, cfg.chatID, cfg.monitorWebAppURL, st, pk, cache)
	go tg.RunPolling(ctx)

	srv := server.New(tg, st, pk, cfg.baseURL, cfg.webhookSecret, cfg.token, cfg.plankaSecretKey, server.DevAuthConfig{
		Token:             cfg.devAuthToken,
		TelegramUsername:  cfg.devAuthTelegramUsername,
		WorkspaceUsername: cfg.devAuthWorkspaceUsername,
		IsAdmin:           cfg.devAuthAdmin,
	})
	log.Printf("listening on %s", cfg.listenAddr)
	if err := srv.Run(ctx, cfg.listenAddr); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	token, chatID, baseURL, dsn       string
	monitorWebAppURL                  string
	webhookSecret, listenAddr         string
	redisURL                          string
	cacheLimit                        int
	envAdmins                         []string
	plankaInternalURL                 string
	plankaAPIEmail, plankaAPIPassword string
	plankaSecretKey                   string
	devAuthToken                      string
	devAuthTelegramUsername           string
	devAuthWorkspaceUsername          string
	devAuthAdmin                      bool
}

func loadConfig() config {
	return config{
		token:                    mustEnv("TELEGRAM_BOT_TOKEN"),
		chatID:                   mustEnv("TELEGRAM_CHAT_ID"),
		baseURL:                  mustEnv("PLANKA_BASE_URL"),
		monitorWebAppURL:         os.Getenv("MONITOR_WEBAPP_URL"),
		dsn:                      mustEnv("DATABASE_URL"),
		webhookSecret:            os.Getenv("WEBHOOK_SECRET"),
		listenAddr:               envDefault("LISTEN_ADDR", defaultListenAddr),
		redisURL:                 os.Getenv("REDIS_URL"),
		cacheLimit:               envInt("NOTIF_CACHE_LIMIT", defaultCacheLimit),
		envAdmins:                splitCSV(os.Getenv("ADMIN_TELEGRAM_USERNAMES")),
		plankaInternalURL:        envDefault("PLANKA_INTERNAL_URL", defaultPlankaInternal),
		plankaAPIEmail:           os.Getenv("PLANKA_API_EMAIL"),
		plankaAPIPassword:        os.Getenv("PLANKA_API_PASSWORD"),
		plankaSecretKey:          envDefault("PLANKA_SECRET_KEY", os.Getenv("SECRET_KEY")),
		devAuthToken:             os.Getenv("MONITOR_DEV_AUTH_TOKEN"),
		devAuthTelegramUsername:  os.Getenv("MONITOR_DEV_TELEGRAM_USERNAME"),
		devAuthWorkspaceUsername: os.Getenv("MONITOR_DEV_WORKSPACE_USERNAME"),
		devAuthAdmin:             envBool("MONITOR_DEV_IS_ADMIN", false),
	}
}

// setupNotifCache повертає Redis-кеш, якщо REDIS_URL задано та сервер відповідає,
// інакше — fallback in-memory (буде зкинуто при рестарті).
func setupNotifCache(ctx context.Context, redisURL string, limit int) telegram.NotifCache {
	if redisURL == "" {
		log.Printf("notif cache: REDIS_URL не задано, використовую in-memory (limit=%d)", limit)
		return telegram.NewMemoryCache(limit)
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("notif cache: невалідний REDIS_URL (%v) — fallback на in-memory", err)
		return telegram.NewMemoryCache(limit)
	}
	rdb := redis.NewClient(opt)
	pingCtx, cancel := context.WithTimeout(ctx, redisPingTimeout)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Printf("notif cache: Redis недоступний (%v) — fallback на in-memory", err)
		return telegram.NewMemoryCache(limit)
	}
	log.Printf("notif cache: Redis %s, limit=%d", redisURL, limit)
	return telegram.NewRedisCache(rdb, limit)
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("env %s is required", k)
	}
	return v
}

func envDefault(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func envBool(k string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
