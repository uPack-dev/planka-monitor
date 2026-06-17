package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Read-only запити проти Planka-схеми (user_account, label, card_membership).
// Використовуються для збагачення подій (delete-events часто прилітають без
// потрібних даних в included).

// GetPlankaAvatar — повертає (uploadedFileId, extension) для зчитування картинки
// аватара через Planka API (path: /user-avatars/{id}/cover-180.{ext}).
// ok=false означає «у юзера немає аватара або юзера не існує».
func (s *Store) GetPlankaAvatar(ctx context.Context, username string) (uploadedFileID, extension string, ok bool) {
	username = normalize(username)
	if username == "" {
		return "", "", false
	}
	var fileID, ext string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(avatar->>'uploadedFileId',''),
		       COALESCE(avatar->>'extension','')
		  FROM user_account
		 WHERE username = $1`, username).Scan(&fileID, &ext)
	if err != nil || fileID == "" || ext == "" {
		return "", "", false
	}
	return fileID, ext, true
}

// GetPlankaDisplayName — повертає імʼя Planka-користувача, fallback на username.
func (s *Store) GetPlankaDisplayName(ctx context.Context, username string) string {
	username = normalize(username)
	if username == "" {
		return ""
	}
	var name string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(name, ''), username, '')
		  FROM user_account
		 WHERE username = $1`, username).Scan(&name)
	if err != nil {
		return ""
	}
	return name
}

// GetUsernameByID — повертає planka username за id користувача (порожньо, якщо не знайдено).
func (s *Store) GetUsernameByID(ctx context.Context, userID string) string {
	if userID == "" {
		return ""
	}
	var u string
	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(username,'') FROM user_account WHERE id::text = $1`, userID).Scan(&u); err != nil {
		return ""
	}
	return u
}

// GetLabelName — повертає назву мітки за її id; падає до color, якщо name порожня.
func (s *Store) GetLabelName(ctx context.Context, labelID string) string {
	if labelID == "" {
		return ""
	}
	var name, color string
	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(name,''), COALESCE(color,'') FROM label WHERE id::text = $1`, labelID).Scan(&name, &color); err != nil {
		return ""
	}
	if name != "" {
		return name
	}
	return color
}

// ListCardMemberUsernames — повертає planka usernames усіх учасників картки.
func (s *Store) ListCardMemberUsernames(ctx context.Context, cardID string) ([]string, error) {
	if cardID == "" {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ua.username FROM card_membership cm
		JOIN user_account ua ON ua.id = cm.user_id
		WHERE cm.card_id::text = $1`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		if u != "" {
			out = append(out, u)
		}
	}
	return out, rows.Err()
}

func (s *Store) GetPlankaUserID(ctx context.Context, username string) (string, error) {
	username = normalize(username)
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id::text FROM user_account WHERE username = $1`, username).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("planka-юзера не знайдено")
	}
	return id, err
}

func (s *Store) PlankaUserExists(ctx context.Context, username string) (bool, error) {
	username = normalize(username)
	var n int
	err := s.pool.QueryRow(ctx, `SELECT 1 FROM user_account WHERE username = $1 LIMIT 1`, username).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) UpdatePlankaUsername(ctx context.Context, oldUsername, newUsername string) error {
	oldUsername = normalize(oldUsername)
	newUsername = normalize(newUsername)
	if oldUsername == "" || newUsername == "" {
		return errors.New("обидва username обовʼязкові")
	}
	tag, err := s.pool.Exec(ctx, `UPDATE user_account SET username = $1 WHERE username = $2`, newUsername, oldUsername)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("такого planka-юзера не знайдено")
	}
	_, _ = s.pool.Exec(ctx, `UPDATE monitor_users SET planka_username = $1 WHERE planka_username = $2`,
		newUsername, oldUsername)
	return nil
}
