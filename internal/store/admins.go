package store

import "context"

// IsAdmin — true, якщо username вказано в ADMIN_TELEGRAM_USERNAMES або
// користувач має роль "admin" у Planka (через привʼязку monitor_users → user_account).
func (s *Store) IsAdmin(ctx context.Context, tgUsername string) bool {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return false
	}
	if _, ok := s.envAdmins[tgUsername]; ok {
		return true
	}
	var granted int
	if err := s.pool.QueryRow(ctx,
		`SELECT 1 FROM monitor_admin_grants WHERE telegram_username = $1 LIMIT 1`,
		tgUsername).Scan(&granted); err == nil {
		return true
	}
	var role string
	err := s.pool.QueryRow(ctx, `
		SELECT ua.role FROM monitor_users mu
		JOIN user_account ua
		  ON ua.username = COALESCE(NULLIF(mu.planka_username,''), mu.telegram_username)
		WHERE mu.telegram_username = $1
		LIMIT 1`, tgUsername).Scan(&role)
	if err != nil {
		return false
	}
	return role == "admin"
}

func (s *Store) SetMonitorAdmin(ctx context.Context, tgUsername string, enabled bool) error {
	tgUsername = normalize(tgUsername)
	if tgUsername == "" {
		return errEmptyTGUsername
	}
	if enabled {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO monitor_admin_grants (telegram_username) VALUES ($1)
			ON CONFLICT (telegram_username) DO NOTHING`, tgUsername)
		return err
	}
	_, err := s.pool.Exec(ctx,
		`DELETE FROM monitor_admin_grants WHERE telegram_username = $1`,
		tgUsername)
	return err
}
