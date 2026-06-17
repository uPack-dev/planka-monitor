package store

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

var errEmptyTGUsername = errors.New("empty telegram username")

var (
	ErrPlankaUserNotFound = errors.New("planka user not found")
	ErrWorkspaceTaken     = errors.New("workspace username already bound")
)

const userBindingColumns = `
	telegram_username, chat_id, COALESCE(telegram_user_id, 0),
		COALESCE(planka_username,''), notifications_enabled, is_blocked, onboarding_completed,
		notify_assignments, notify_comments, notify_changes, notify_done`

type monitorUserIDColumn string

const (
	monitorChatIDColumn         monitorUserIDColumn = "chat_id"
	monitorTelegramIDColumn     monitorUserIDColumn = "telegram_user_id"
	monitorUserFallbackUsername                     = `
		AND (planka_username IS NULL OR planka_username = '')`
)

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *Store) SetUser(ctx context.Context, tgUsername string, chatID, tgUserID int64) error {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO monitor_users (telegram_username, chat_id, telegram_user_id) VALUES ($1, $2, $3)
		ON CONFLICT (telegram_username) DO UPDATE
		  SET chat_id = EXCLUDED.chat_id,
		      telegram_user_id = EXCLUDED.telegram_user_id,
		      updated_at = now()`,
		tgUsername, chatID, tgUserID)
	return err
}

func (s *Store) SetNotifications(ctx context.Context, tgUsername string, enabled bool) error {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE monitor_users SET notifications_enabled = $1, updated_at = now() WHERE telegram_username = $2`,
		enabled, tgUsername)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("користувач не зареєстрований у боті")
	}
	return nil
}

func (s *Store) DeleteUser(ctx context.Context, tgUsername string) (bool, error) {
	tgUsername = normalize(tgUsername)
	tag, err := s.pool.Exec(ctx, `DELETE FROM monitor_users WHERE telegram_username = $1`, tgUsername)
	return tag.RowsAffected() > 0, err
}

func (s *Store) GetByTelegram(ctx context.Context, tgUsername string) (*UserBinding, error) {
	tgUsername = normalize(tgUsername)
	row := s.pool.QueryRow(ctx,
		`SELECT `+userBindingColumns+`
		   FROM monitor_users
		  WHERE telegram_username = $1`,
		tgUsername)
	b, err := scanUserBinding(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

// IsBlocked — швидка перевірка для гейту вхідних повідомлень.
func (s *Store) IsBlocked(ctx context.Context, tgUsername string) bool {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return false
	}
	var blocked bool
	if err := s.pool.QueryRow(ctx,
		`SELECT is_blocked FROM monitor_users WHERE telegram_username = $1`, tgUsername).Scan(&blocked); err != nil {
		return false
	}
	return blocked
}

func (s *Store) SetBlocked(ctx context.Context, tgUsername string, blocked bool) error {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO monitor_users (telegram_username, chat_id, is_blocked, notifications_enabled)
		VALUES ($1, 0, $2, false)
		ON CONFLICT (telegram_username) DO UPDATE
		  SET is_blocked = EXCLUDED.is_blocked,
		      notifications_enabled = CASE WHEN EXCLUDED.is_blocked THEN false ELSE monitor_users.notifications_enabled END,
		      updated_at = now()`,
		tgUsername, blocked)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("not affected")
	}
	return nil
}

// GetChatIDForPlanka — спершу шукаємо явну привʼязку planka_username,
// далі fallback на telegram_username = planka_username.
func (s *Store) GetChatIDForPlanka(ctx context.Context, plankaUsername string) (int64, bool) {
	return s.getActiveMonitorIDForPlanka(ctx, plankaUsername, monitorChatIDColumn, false)
}

// GetTelegramIDForPlanka — повертає telegram user_id, привʼязаний до planka-юзера.
func (s *Store) GetTelegramIDForPlanka(ctx context.Context, plankaUsername string) (int64, bool) {
	return s.getActiveMonitorIDForPlanka(ctx, plankaUsername, monitorTelegramIDColumn, true)
}

func (s *Store) getActiveMonitorIDForPlanka(ctx context.Context, plankaUsername string, column monitorUserIDColumn, requireNonZero bool) (int64, bool) {
	plankaUsername = normalize(plankaUsername)
	if plankaUsername == "" {
		return 0, false
	}

	requiredValueCondition := monitorColumnRequiredCondition(column, requireNonZero)

	var id int64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(`+string(column)+`, 0)
		  FROM monitor_users
		 WHERE planka_username = $1`+requiredValueCondition+`
		   AND notifications_enabled
		   AND NOT is_blocked
		 ORDER BY updated_at DESC
		 LIMIT 1`, plankaUsername).Scan(&id)
	if err == nil && (!requireNonZero || id != 0) {
		return id, true
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(`+string(column)+`, 0)
		  FROM monitor_users
		 WHERE telegram_username = $1`+monitorUserFallbackUsername+requiredValueCondition+`
		   AND notifications_enabled
		   AND NOT is_blocked
		 LIMIT 1`, plankaUsername).Scan(&id)
	if err != nil || (requireNonZero && id == 0) {
		return 0, false
	}

	return id, true
}

func monitorColumnRequiredCondition(column monitorUserIDColumn, required bool) string {
	if !required {
		return ""
	}
	return `
		   AND ` + string(column) + ` IS NOT NULL`
}

func (s *Store) BindPlanka(ctx context.Context, tgUsername, plankaUsername string) error {
	tgUsername = normalize(tgUsername)
	plankaUsername = normalize(plankaUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	var p any
	if plankaUsername != "" {
		p = plankaUsername
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE monitor_users SET planka_username = $1, updated_at = now() WHERE telegram_username = $2`,
		p, tgUsername)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("користувач не зареєстрований у боті — нехай він спочатку напише /start")
	}
	return nil
}

func (s *Store) BindPlankaIfAvailable(ctx context.Context, tgUsername, plankaUsername string) error {
	tgUsername = normalize(tgUsername)
	plankaUsername = normalize(plankaUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	if plankaUsername == "" {
		return ErrPlankaUserNotFound
	}
	exists, err := s.PlankaUserExists(ctx, plankaUsername)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPlankaUserNotFound
	}

	var owner string
	err = s.pool.QueryRow(ctx, `
		SELECT telegram_username FROM monitor_users
		WHERE telegram_username <> $1
		  AND COALESCE(NULLIF(planka_username, ''), telegram_username) = $2
		LIMIT 1`, tgUsername, plankaUsername).Scan(&owner)
	if err == nil {
		return ErrWorkspaceTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return s.BindPlanka(ctx, tgUsername, plankaUsername)
}

func (s *Store) SetNotificationPreferences(ctx context.Context, tgUsername string, prefs NotificationPreferences) error {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE monitor_users
		   SET notify_assignments = $1,
		       notify_comments = $2,
		       notify_changes = $3,
		       notify_done = $4,
		       updated_at = now()
		 WHERE telegram_username = $5`,
		prefs.Assignments, prefs.Comments, prefs.Changes, prefs.Done, tgUsername)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("користувач не зареєстрований у боті")
	}
	return nil
}

func (s *Store) SetOnboardingCompleted(ctx context.Context, tgUsername string, completed bool) error {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE monitor_users SET onboarding_completed = $1, updated_at = now() WHERE telegram_username = $2`,
		completed, tgUsername)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("користувач не зареєстрований у боті")
	}
	return nil
}

func (s *Store) MuteCard(ctx context.Context, tgUsername, cardID string) error {
	tgUsername = normalize(tgUsername)
	cardID = strings.TrimSpace(cardID)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	if cardID == "" {
		return errors.New("empty card id")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO monitor_muted_cards (telegram_username, card_id)
		VALUES ($1, $2)
		ON CONFLICT (telegram_username, card_id) DO NOTHING`,
		tgUsername, cardID)
	return err
}

func (s *Store) UnmuteCard(ctx context.Context, tgUsername, cardID string) error {
	tgUsername = normalize(tgUsername)
	cardID = strings.TrimSpace(cardID)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	if cardID == "" {
		return errors.New("empty card id")
	}
	_, err := s.pool.Exec(ctx, `
		DELETE FROM monitor_muted_cards
		 WHERE telegram_username = $1
		   AND card_id = $2`,
		tgUsername, cardID)
	return err
}

func (s *Store) CardMutedForPlankaUser(ctx context.Context, plankaUsername, cardID string) bool {
	plankaUsername = normalize(plankaUsername)
	cardID = strings.TrimSpace(cardID)
	if plankaUsername == "" || cardID == "" {
		return false
	}
	var ok int
	err := s.pool.QueryRow(ctx, `
		SELECT 1
		  FROM monitor_muted_cards mc
		  JOIN monitor_users mu ON mu.telegram_username = mc.telegram_username
		 WHERE mc.card_id = $2
		   AND (
		     mu.planka_username = $1
		     OR (mu.telegram_username = $1 AND (mu.planka_username IS NULL OR mu.planka_username = ''))
		   )
		 LIMIT 1`, plankaUsername, cardID).Scan(&ok)
	return err == nil && ok == 1
}

func (s *Store) PersonalCategoryEnabled(ctx context.Context, plankaUsername, category string) bool {
	plankaUsername = normalize(plankaUsername)
	if plankaUsername == "" || category == "" {
		return false
	}
	row := s.pool.QueryRow(ctx, `
		SELECT notifications_enabled, is_blocked,
		       notify_assignments, notify_comments, notify_changes, notify_done
		  FROM monitor_users
		 WHERE planka_username = $1
		    OR (telegram_username = $1 AND (planka_username IS NULL OR planka_username = ''))
		 ORDER BY CASE WHEN planka_username = $1 THEN 0 ELSE 1 END, updated_at DESC
		 LIMIT 1`, plankaUsername)

	var enabled, blocked bool
	var prefs NotificationPreferences
	if err := row.Scan(
		&enabled,
		&blocked,
		&prefs.Assignments,
		&prefs.Comments,
		&prefs.Changes,
		&prefs.Done,
	); err != nil {
		return false
	}
	if !enabled || blocked {
		return false
	}
	switch category {
	case "assignments":
		return prefs.Assignments
	case "comments":
		return prefs.Comments
	case "changes":
		return prefs.Changes
	case "done":
		return prefs.Done
	default:
		return true
	}
}

func (s *Store) ListUsers(ctx context.Context) ([]UserBinding, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+userBindingColumns+`
		   FROM monitor_users
		  ORDER BY telegram_username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []UserBinding
	for rows.Next() {
		b, err := scanUserBinding(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func scanUserBinding(row rowScanner) (UserBinding, error) {
	var b UserBinding
	err := row.Scan(
		&b.TelegramUsername,
		&b.ChatID,
		&b.TelegramUserID,
		&b.PlankaUsername,
		&b.NotificationsEnabled,
		&b.IsBlocked,
		&b.OnboardingCompleted,
		&b.Preferences.Assignments,
		&b.Preferences.Comments,
		&b.Preferences.Changes,
		&b.Preferences.Done,
	)
	return b, err
}
