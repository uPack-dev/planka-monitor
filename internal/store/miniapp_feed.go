package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *Store) ListFeedEvents(ctx context.Context, workspaceUsername, filter string, limit int, before *time.Time) ([]FeedEvent, error) {
	workspaceUsername = normalize(workspaceUsername)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	beforeValue := time.Now().Add(24 * time.Hour)
	if before != nil {
		beforeValue = *before
	}
	rows, err := s.pool.Query(ctx, `
		WITH me AS (
			SELECT id FROM user_account WHERE username = $1 LIMIT 1
		), feed AS (
			SELECT 'n:' || n.id::text AS id,
			       n.type AS type,
			       COALESCE(n.created_at, now()) AS created_at,
			       n.is_read AS is_read,
			       COALESCE(n.card_id::text, '') AS card_id,
			       COALESCE(c.name, n.data->'card'->>'name', '') AS card_title,
			       COALESCE(b.id::text, '') AS board_id,
			       COALESCE(b.name, '') AS board_name,
			       COALESCE(p.id::text, '') AS project_id,
			       COALESCE(p.name, '') AS project_name,
			       COALESCE(n.data->>'text', cm.text, '') AS text,
			       COALESCE(u.id::text, '') AS actor_id,
			       COALESCE(u.username, '') AS actor_username,
			       COALESCE(u.name, '') AS actor_name,
			       COALESCE(u.avatar->>'uploadedFileId', '') AS actor_avatar_file_id,
			       COALESCE(u.avatar->>'extension', '') AS actor_avatar_ext
			  FROM notification n
			  JOIN me ON me.id = n.user_id
			  LEFT JOIN card c ON c.id = n.card_id
			  LEFT JOIN board b ON b.id = COALESCE(n.board_id, c.board_id)
			  LEFT JOIN project p ON p.id = b.project_id
			  LEFT JOIN comment cm ON cm.id = n.comment_id
			  LEFT JOIN user_account u ON u.id = n.creator_user_id
			 WHERE COALESCE(n.created_at, now()) < $3
			   AND NOT COALESCE(b.is_archived, false)
			   AND NOT COALESCE(p.is_archived, false)
			   AND (
			     n.card_id IS NULL
			     OR EXISTS (
			       SELECT 1 FROM board_membership bm
			        WHERE bm.board_id = COALESCE(n.board_id, c.board_id)
			          AND bm.user_id = me.id
			     )
			   )
			UNION ALL
			SELECT 'a:' || a.id::text AS id,
			       a.type AS type,
			       COALESCE(a.created_at, now()) AS created_at,
			       true AS is_read,
			       a.card_id::text AS card_id,
			       COALESCE(c.name, a.data->'card'->>'name', '') AS card_title,
			       COALESCE(b.id::text, '') AS board_id,
			       COALESCE(b.name, '') AS board_name,
			       COALESCE(p.id::text, '') AS project_id,
			       COALESCE(p.name, '') AS project_name,
			       '' AS text,
			       COALESCE(u.id::text, '') AS actor_id,
			       COALESCE(u.username, '') AS actor_username,
			       COALESCE(u.name, '') AS actor_name,
			       COALESCE(u.avatar->>'uploadedFileId', '') AS actor_avatar_file_id,
			       COALESCE(u.avatar->>'extension', '') AS actor_avatar_ext
			  FROM action a
			  JOIN card c ON c.id = a.card_id
			  JOIN board b ON b.id = COALESCE(a.board_id, c.board_id)
			  JOIN project p ON p.id = b.project_id
			  JOIN me ON true
			  LEFT JOIN user_account u ON u.id = a.user_id
			  JOIN board_membership bm ON bm.board_id = b.id AND bm.user_id = me.id
			 WHERE COALESCE(a.created_at, now()) < $3
			   AND NOT b.is_archived
			   AND NOT p.is_archived
			   AND a.type IN ('completeTask', 'uncompleteTask', 'moveCard', 'addMemberToCard')
			   AND (
			     EXISTS (SELECT 1 FROM card_membership cm WHERE cm.card_id = c.id AND cm.user_id = me.id)
			     OR EXISTS (
			       SELECT 1 FROM task_list tl JOIN task t ON t.task_list_id = tl.id
			        WHERE tl.card_id = c.id AND t.assignee_user_id = me.id
			     )
			   )
		)
		SELECT id, type, created_at, is_read, card_id, card_title,
		       board_id, board_name, project_id, project_name, text,
		       actor_id, actor_username, actor_name, actor_avatar_file_id, actor_avatar_ext
		  FROM feed
		 WHERE ($2 = 'all'
		        OR ($2 = 'comments' AND type IN ('commentCard', 'mentionInComment'))
		        OR ($2 = 'assignments' AND type = 'addMemberToCard')
		        OR ($2 = 'done' AND type IN ('completeTask', 'uncompleteTask')))
		 ORDER BY created_at DESC
		 LIMIT $4`, workspaceUsername, normalizeFeedFilter(filter), beforeValue, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FeedEvent
	cardSet := map[string]struct{}{}
	for rows.Next() {
		var item FeedEvent
		var actorID, actorUsername, actorName, actorAvatarFileID, actorAvatarExt string
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.CreatedAt,
			&item.IsRead,
			&item.CardID,
			&item.CardTitle,
			&item.BoardID,
			&item.BoardName,
			&item.ProjectID,
			&item.ProjectName,
			&item.Text,
			&actorID,
			&actorUsername,
			&actorName,
			&actorAvatarFileID,
			&actorAvatarExt,
		); err != nil {
			return nil, err
		}
		item.Actor = miniUserWithAvatar(actorID, actorUsername, actorName, actorAvatarFileID, actorAvatarExt)
		out = append(out, item)
		if item.CardID != "" {
			cardSet[item.CardID] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	mutedCards, err := s.listMutedCardIDsForPlankaUser(ctx, workspaceUsername, keys(cardSet))
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].IsMuted = mutedCards[out[i].CardID]
	}
	return out, nil
}

func (s *Store) MarkFeedRead(ctx context.Context, workspaceUsername, feedID string) error {
	if !strings.HasPrefix(feedID, "n:") {
		return nil
	}
	workspaceUsername = normalize(workspaceUsername)
	id := strings.TrimPrefix(feedID, "n:")
	tag, err := s.pool.Exec(ctx, `
		WITH me AS (SELECT id FROM user_account WHERE username = $1 LIMIT 1)
		UPDATE notification n
		   SET is_read = true, updated_at = now()
		  FROM me
		 WHERE n.id::text = $2 AND n.user_id = me.id`, workspaceUsername, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("feed item not found")
	}
	return nil
}

func normalizeFeedFilter(filter string) string {
	switch strings.ToLower(strings.TrimSpace(filter)) {
	case "comments", "assignments", "done":
		return strings.ToLower(strings.TrimSpace(filter))
	default:
		return "all"
	}
}
