package store

import (
	"context"
	"strings"
)

func (s *Store) ListAdminSubscribers(ctx context.Context, q, status string) ([]AdminSubscriber, error) {
	q = strings.TrimSpace(q)
	status = strings.ToLower(strings.TrimSpace(status))
	rows, err := s.pool.Query(ctx, `
		SELECT mu.telegram_username,
		       COALESCE(mu.telegram_user_id, 0),
		       COALESCE(NULLIF(mu.planka_username, ''), mu.telegram_username) AS workspace_username,
		       ua.id IS NOT NULL AS workspace_exists,
		       COALESCE(ua.name, '') AS display_name,
		       COALESCE(ua.avatar->>'uploadedFileId', ''),
		       COALESCE(ua.avatar->>'extension', ''),
		       COALESCE(ua.role, '') AS role,
		       g.telegram_username IS NOT NULL AS monitor_admin,
		       mu.notifications_enabled,
		       mu.is_blocked,
		       mu.notify_assignments,
		       mu.notify_comments,
		       mu.notify_changes,
		       mu.notify_done,
		       mu.updated_at
		  FROM monitor_users mu
		  LEFT JOIN user_account ua
		    ON ua.username = COALESCE(NULLIF(mu.planka_username, ''), mu.telegram_username)
		  LEFT JOIN monitor_admin_grants g
		    ON g.telegram_username = mu.telegram_username
		 WHERE ($1 = ''
		        OR lower(mu.telegram_username) LIKE '%' || lower($1) || '%'
		        OR lower(COALESCE(NULLIF(mu.planka_username, ''), mu.telegram_username)) LIKE '%' || lower($1) || '%'
		        OR lower(COALESCE(ua.name, '')) LIKE '%' || lower($1) || '%')
		   AND ($2 = ''
		        OR ($2 = 'active' AND mu.notifications_enabled AND NOT mu.is_blocked)
		        OR ($2 = 'muted' AND NOT mu.notifications_enabled AND NOT mu.is_blocked)
		        OR ($2 = 'blocked' AND mu.is_blocked)
		        OR ($2 = 'admin' AND (g.telegram_username IS NOT NULL OR ua.role = 'admin')))
		 ORDER BY mu.updated_at DESC, mu.telegram_username`, q, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AdminSubscriber
	for rows.Next() {
		var item AdminSubscriber
		var monitorAdmin bool
		if err := rows.Scan(
			&item.TelegramUsername,
			&item.TelegramUserID,
			&item.WorkspaceUsername,
			&item.WorkspaceExists,
			&item.DisplayName,
			&item.AvatarUploadedFileID,
			&item.AvatarExtension,
			&item.Role,
			&monitorAdmin,
			&item.NotificationsEnabled,
			&item.IsBlocked,
			&item.Preferences.Assignments,
			&item.Preferences.Comments,
			&item.Preferences.Changes,
			&item.Preferences.Done,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_, envAdmin := s.envAdmins[item.TelegramUsername]
		item.IsAdmin = envAdmin || monitorAdmin || item.Role == "admin"
		out = append(out, item)
	}
	return out, rows.Err()
}
