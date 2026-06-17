package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL SQLSTATE codes which we treat as benign for idempotent ALTER
// statements (column/table уже існує/не існує тощо).
//
// Раніше ми порівнювали текст помилки рядково, що ламалося на локалізованих
// білдах postgres. SQLSTATE-коди стабільні.
var benignMigrateCodes = map[string]struct{}{
	"42701": {}, // duplicate_column
	"42P07": {}, // duplicate_table
	"42703": {}, // undefined_column
	"42P01": {}, // undefined_table (RENAME COLUMN of legacy table)
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS monitor_users (
			telegram_username     text PRIMARY KEY,
			chat_id               bigint NOT NULL,
			telegram_user_id      bigint,
			planka_username       text,
			notifications_enabled boolean NOT NULL DEFAULT false,
			notify_assignments    boolean NOT NULL DEFAULT true,
			notify_comments       boolean NOT NULL DEFAULT true,
			notify_changes        boolean NOT NULL DEFAULT false,
			notify_done           boolean NOT NULL DEFAULT true,
			onboarding_completed  boolean NOT NULL DEFAULT false,
			updated_at            timestamptz NOT NULL DEFAULT now()
		)`,
		// Upgrade-кроки від старих схем — ідемпотентні.
		`ALTER TABLE monitor_users RENAME COLUMN username TO telegram_username`,
		`ALTER TABLE monitor_users ADD COLUMN planka_username text`,
		`ALTER TABLE monitor_users ADD COLUMN telegram_user_id bigint`,
		`ALTER TABLE monitor_users ADD COLUMN notifications_enabled boolean NOT NULL DEFAULT false`,
		`ALTER TABLE monitor_users ADD COLUMN notify_assignments boolean NOT NULL DEFAULT true`,
		`ALTER TABLE monitor_users ADD COLUMN notify_comments boolean NOT NULL DEFAULT true`,
		`ALTER TABLE monitor_users ADD COLUMN notify_changes boolean NOT NULL DEFAULT true`,
		`ALTER TABLE monitor_users ADD COLUMN notify_done boolean NOT NULL DEFAULT true`,
		`ALTER TABLE monitor_users ALTER COLUMN notify_changes SET DEFAULT false`,
		`ALTER TABLE monitor_users ADD COLUMN onboarding_completed boolean NOT NULL DEFAULT true`,
		`ALTER TABLE monitor_users ALTER COLUMN onboarding_completed SET DEFAULT false`,
		`ALTER TABLE monitor_users ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now()`,
		`ALTER TABLE monitor_users ADD COLUMN is_blocked boolean NOT NULL DEFAULT false`,
		`CREATE INDEX IF NOT EXISTS monitor_users_planka_idx ON monitor_users (planka_username)`,
		`CREATE TABLE IF NOT EXISTS monitor_templates (
			event      text NOT NULL,
			kind       text NOT NULL,
			template   text NOT NULL,
			updated_at timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (event, kind)
		)`,
		`CREATE TABLE IF NOT EXISTS monitor_admin_grants (
			telegram_username text PRIMARY KEY,
			created_at        timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS monitor_muted_cards (
			telegram_username text NOT NULL,
			card_id           text NOT NULL,
			created_at        timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (telegram_username, card_id)
		)`,
		`CREATE INDEX IF NOT EXISTS monitor_muted_cards_card_idx ON monitor_muted_cards (card_id)`,
	}
	for _, sql := range stmts {
		if _, err := pool.Exec(ctx, sql); err != nil && !isBenignMigrateError(err) {
			return err
		}
	}
	return nil
}

func isBenignMigrateError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	_, ok := benignMigrateCodes[pgErr.Code]
	return ok
}
