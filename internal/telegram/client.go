// Package telegram реалізує бот-клієнт Telegram для Planka monitor.
//
// Файлова структура:
//
//	client.go        — Client, конструктори, Notification.
//	types.go         — JSON-DTO для Telegram Bot API.
//	api.go           — низькорівневі HTTP-обгортки (єдина callJSON-точка).
//	cache.go         — кеш Notification (Redis або in-memory FIFO).
//	notification.go  — публічні Send*-методи + buildNotifKeyboard.
//	polling.go       — getUpdates-цикл, диспетчер callback/message.
//	commands.go      — обробка слеш-команд (handleCommand + cmd*).
//	avatar.go        — /avatar (selfie) + /setavatar (адмін-фото).
//	menu.go          — список botCommands + ensureChatMenu.
//	text.go          — статичні шаблони відповідей (help/start/templates).
package telegram

import (
	"net/http"
	"sync"
	"time"

	"planka-monitor/internal/planka"
	"planka-monitor/internal/store"

	"github.com/redis/go-redis/v9"
)

const (
	telegramAPIBase = "https://api.telegram.org"
	httpTimeout     = 60 * time.Second
	pollTimeout     = 30 // seconds, getUpdates long-poll
	defaultCacheCap = 400
)

type Client struct {
	token     string
	chatID    string
	webAppURL string
	store     *store.Store
	planka    *planka.Client
	http      *http.Client

	mu            sync.Mutex
	chatMenuAdmin map[int64]bool // chatID → останнє надіслане меню було адмінським
	chatMenuKnown map[int64]bool // чи взагалі ставили меню для цього чату
	notifCache    NotifCache     // компакт+деталі для toggle через inline кнопки
}

// Notification — універсальна структура повідомлення зі згорнутими деталями.
type Notification struct {
	Heading    string // основний компактний рядок (HTML)
	Details    string // приховані за кнопкою «Деталі» (HTML)
	CardURL    string // URL для кнопки «Картка»
	BoardURL   string // URL для кнопки «Дошка»
	ProjectURL string // URL для кнопки «Проєкт»
}

func New(token, chatID, webAppURL string, st *store.Store, pk *planka.Client, cache NotifCache) *Client {
	if cache == nil {
		cache = newMemNotifCache(defaultCacheCap)
	}
	return &Client{
		token:         token,
		chatID:        chatID,
		webAppURL:     webAppURL,
		store:         st,
		planka:        pk,
		http:          &http.Client{Timeout: httpTimeout},
		chatMenuAdmin: map[int64]bool{},
		chatMenuKnown: map[int64]bool{},
		notifCache:    cache,
	}
}

// NewRedisCache — конструктор Redis-кешу для DI з main.
// limit — скільки повідомлень тримаємо «розгортуваними». Старіші втратять кнопку «Деталі».
func NewRedisCache(rdb *redis.Client, limit int) NotifCache {
	if rdb == nil || limit <= 0 {
		return nil
	}
	return newRedisNotifCache(rdb, limit)
}

// NewMemoryCache — для випадків коли Redis недоступний / для тестів.
func NewMemoryCache(limit int) NotifCache {
	return newMemNotifCache(limit)
}
