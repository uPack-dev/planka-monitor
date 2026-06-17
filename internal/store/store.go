// Package store розділяється по доменах:
//   - store.go     — pool, конструктор, нормалізація.
//   - migrate.go   — DDL-міграції.
//   - users.go     — підписки/привʼязки в monitor_users.
//   - admins.go    — IsAdmin (env + БД Planka).
//   - planka.go    — read-only запити проти Planka-схеми (user_account, label, card_membership).
//   - templates.go — кастомні шаблони monitor_templates.
package store

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool      *pgxpool.Pool
	envAdmins map[string]struct{} // telegram usernames (lowercase)
}

type UserBinding struct {
	TelegramUsername     string
	ChatID               int64
	TelegramUserID       int64
	PlankaUsername       string // optional manual override; if empty — TelegramUsername is used
	NotificationsEnabled bool
	IsBlocked            bool
	OnboardingCompleted  bool
	Preferences          NotificationPreferences
}

type NotificationPreferences struct {
	Assignments bool `json:"assignments"`
	Comments    bool `json:"comments"`
	Changes     bool `json:"changes"`
	Done        bool `json:"done"`
}

func DefaultNotificationPreferences() NotificationPreferences {
	return NotificationPreferences{
		Assignments: true,
		Comments:    true,
		Changes:     false,
		Done:        true,
	}
}

type Template struct {
	Event    string
	Kind     string // "channel" | "personal"
	Template string
}

func New(ctx context.Context, dsn string, envAdmins []string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 30; i++ {
		if err = pool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		pool.Close()
		return nil, err
	}
	if err := migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	envSet := make(map[string]struct{}, len(envAdmins))
	for _, a := range envAdmins {
		if a = normalize(a); a != "" {
			envSet[a] = struct{}{}
		}
	}
	return &Store{pool: pool, envAdmins: envSet}, nil
}

func (s *Store) Close() { s.pool.Close() }

func normalize(u string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(u), "@"))
}
