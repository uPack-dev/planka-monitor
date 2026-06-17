// Package server — HTTP-фронт monitor: /healthz та /webhook (Planka → Telegram).
//
// Файли:
//
//	server.go   — конструктор/Run + хендлери /healthz, /webhook.
//	dispatch.go — перетворення events.Render у Telegram-надсилання.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"planka-monitor/internal/events"
	"planka-monitor/internal/planka"
	"planka-monitor/internal/store"
	"planka-monitor/internal/telegram"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 5 * time.Second
	dispatchTimeout   = 10 * time.Second
)

type Server struct {
	tg                  *telegram.Client
	store               *store.Store
	planka              *planka.Client
	baseURL             string
	secret              string
	botToken            string
	plankaSessionSecret string
	devAuth             DevAuthConfig
}

type DevAuthConfig struct {
	Token             string
	TelegramUsername  string
	WorkspaceUsername string
	IsAdmin           bool
}

func New(tg *telegram.Client, st *store.Store, pk *planka.Client, baseURL, secret, botToken, plankaSessionSecret string, devAuth DevAuthConfig) *Server {
	devAuth.TelegramUsername = normalizeAPIUsername(devAuth.TelegramUsername)
	devAuth.WorkspaceUsername = normalizeAPIUsername(devAuth.WorkspaceUsername)
	return &Server{
		tg:                  tg,
		store:               st,
		planka:              pk,
		baseURL:             baseURL,
		secret:              secret,
		botToken:            botToken,
		plankaSessionSecret: plankaSessionSecret,
		devAuth:             devAuth,
	}
}

func (s *Server) Run(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/me", s.handleAPI)
	mux.HandleFunc("/api/v1/me/", s.handleAPI)
	mux.HandleFunc("/api/v1/tasks", s.handleAPI)
	mux.HandleFunc("/api/v1/tasks/", s.handleAPI)
	mux.HandleFunc("/api/v1/cards/", s.handleAPI)
	mux.HandleFunc("/api/v1/avatars/", s.handleAPI)
	mux.HandleFunc("/api/v1/feed", s.handleAPI)
	mux.HandleFunc("/api/v1/feed/", s.handleAPI)
	mux.HandleFunc("/api/v1/admin/", s.handleAPI)
	mux.HandleFunc("/webhook", s.handleWebhook)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.checkSecret(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var p events.Payload
	if err := json.Unmarshal(body, &p); err != nil {
		log.Printf("bad payload: %v; body=%s", err, string(body))
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	// Діагностика: бачимо, які саме події надсилає Planka. Допомагає відладжувати
	// відсутні сповіщення (taskMembership, описи задач тощо).
	log.Printf("webhook event=%s prevKeys=%s", p.Event, prevKeys(p.PrevData))

	w.WriteHeader(http.StatusNoContent)
	go s.dispatch(&p)
}

func (s *Server) checkSecret(r *http.Request) bool {
	if s.secret == "" {
		return true
	}
	got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if got == "" {
		got = r.Header.Get("X-Webhook-Secret")
	}
	return got == s.secret
}

// prevKeys — для діагностичного логу: перелік ключів попереднього item (PrevData).
func prevKeys(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "-"
	}
	// Цей форк Planka шле prevData як {item:{...}, included:{...}}.
	// Спочатку дістаємо item, інакше fallback на плоский формат.
	var wrapper struct {
		Item json.RawMessage `json:"item"`
	}
	target := raw
	if err := json.Unmarshal(raw, &wrapper); err == nil && len(wrapper.Item) > 0 {
		target = wrapper.Item
	}
	var m map[string]any
	if err := json.Unmarshal(target, &m); err != nil {
		return "?"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ",")
}
