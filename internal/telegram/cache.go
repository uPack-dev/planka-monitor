package telegram

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// NotifCache — інтерфейс кешу для Notification.
// Реалізації: Redis (за замовчуванням) та in-memory (fallback).
type NotifCache interface {
	put(ctx context.Context, key string, n Notification)
	get(ctx context.Context, key string) (Notification, bool)
}

func notifKey(chatID, msgID int64) string {
	return fmt.Sprintf("%d:%d", chatID, msgID)
}

// ---------- in-memory FIFO ----------

type memNotifCache struct {
	mu    sync.Mutex
	max   int
	order []string
	data  map[string]Notification
}

func newMemNotifCache(max int) *memNotifCache {
	return &memNotifCache{max: max, data: make(map[string]Notification)}
}

func (c *memNotifCache) put(_ context.Context, key string, n Notification) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.data[key]; !exists {
		c.order = append(c.order, key)
	}
	c.data[key] = n
	for len(c.order) > c.max {
		evict := c.order[0]
		c.order = c.order[1:]
		delete(c.data, evict)
	}
}

func (c *memNotifCache) get(_ context.Context, key string) (Notification, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	n, ok := c.data[key]
	return n, ok
}

// ---------- Redis ----------

const (
	redisListKey  = "planka:notif:order"
	redisDataPfx  = "planka:notif:"
	redisCacheTTL = 7 * 24 * time.Hour
)

// redisNotifCache — кеш на Redis: HASH-поля для деталей + LIST для FIFO-евікшну.
type redisNotifCache struct {
	rdb   *redis.Client
	limit int
}

func newRedisNotifCache(rdb *redis.Client, limit int) *redisNotifCache {
	return &redisNotifCache{rdb: rdb, limit: limit}
}

func (c *redisNotifCache) dataKey(k string) string { return redisDataPfx + k }

// trimScript — атомарно: LPUSH ключ, LTRIM до limit, видалити «вибиті» дата-ключі.
// KEYS[1] = list, ARGV[1] = key, ARGV[2] = limit, ARGV[3] = data-key prefix.
var redisTrimScript = redis.NewScript(`
redis.call("LPUSH", KEYS[1], ARGV[1])
local extra = redis.call("LRANGE", KEYS[1], tonumber(ARGV[2]), -1)
if #extra > 0 then
  redis.call("LTRIM", KEYS[1], 0, tonumber(ARGV[2]) - 1)
  for _, k in ipairs(extra) do
    redis.call("DEL", ARGV[3] .. k)
  end
end
return 1
`)

func (c *redisNotifCache) put(ctx context.Context, key string, n Notification) {
	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, c.dataKey(key),
		"h", n.Heading,
		"d", n.Details,
		"c", n.CardURL,
		"b", n.BoardURL,
		"p", n.ProjectURL,
	)
	// TTL — на випадок, якщо trim-скрипт промахнеться, ключі самі почистяться.
	pipe.Expire(ctx, c.dataKey(key), redisCacheTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("redis cache put: %v", err)
		return
	}
	if _, err := redisTrimScript.Run(ctx, c.rdb,
		[]string{redisListKey},
		key, c.limit, redisDataPfx,
	).Result(); err != nil && err != redis.Nil {
		log.Printf("redis cache trim: %v", err)
	}
}

func (c *redisNotifCache) get(ctx context.Context, key string) (Notification, bool) {
	m, err := c.rdb.HGetAll(ctx, c.dataKey(key)).Result()
	if err != nil || len(m) == 0 {
		return Notification{}, false
	}
	return Notification{
		Heading:    m["h"],
		Details:    m["d"],
		CardURL:    m["c"],
		BoardURL:   m["b"],
		ProjectURL: m["p"],
	}, true
}
