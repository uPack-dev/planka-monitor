package store

import (
	"context"
	"strings"
	"time"
)

func (s *Store) ListTimelineItems(ctx context.Context, workspaceUsername, q string, from, to time.Time, limit int, includeUndated bool) ([]TimelineItem, error) {
	workspaceUsername = normalize(workspaceUsername)
	if limit <= 0 || limit > 200 {
		limit = 80
	}
	rows, err := s.pool.Query(ctx, `
		WITH me AS (
			SELECT id FROM user_account WHERE username = $1 LIMIT 1
		), personal_items AS (
			SELECT 'card'::text AS kind,
			       c.id::text AS card_id,
			       ''::text AS task_id,
			       c.name AS title,
			       COALESCE(c.due_date, c.end_date, c.start_date) AS due_at,
			       c.start_date AS start_at,
			       c.end_date AS end_at,
			       COALESCE(c.description, '') AS description,
			       COALESCE(c.is_due_completed, false) OR done.is_completed_by_label AS is_completed,
			       b.id::text AS board_id,
			       b.name AS board_name,
			       p.id::text AS project_id,
			       p.name AS project_name,
			       COALESCE(l.name, '') AS list_name
			  FROM card c
			  JOIN board b ON b.id = c.board_id
			  JOIN project p ON p.id = b.project_id
			  JOIN list l ON l.id = c.list_id
			  JOIN me ON true
			  JOIN board_membership bm ON bm.board_id = b.id AND bm.user_id = me.id
			  CROSS JOIN LATERAL (
			    SELECT c.due_date IS NULL
			       AND c.end_date IS NULL
			       AND c.start_date IS NULL
			       AND EXISTS (
			         SELECT 1
			           FROM card_label cl
			           JOIN label lb ON lb.id = cl.label_id
			          WHERE cl.card_id = c.id
			            AND lower(trim(lb.name)) IN ('выполнено', 'виконано')
			       ) AS is_completed_by_label
			  ) done
			 WHERE NOT c.is_closed
			   AND NOT b.is_archived
			   AND NOT p.is_archived
			   AND NOT (COALESCE(c.is_due_completed, false) OR done.is_completed_by_label)
			   AND (
			     EXISTS (
			       SELECT 1 FROM card_membership cm
			        WHERE cm.card_id = c.id AND cm.user_id = me.id
			     )
			     OR EXISTS (
			       SELECT 1
			         FROM task_list tl
			         JOIN task t ON t.task_list_id = tl.id
			         JOIN me ON me.id = t.assignee_user_id
			        WHERE tl.card_id = c.id
			          AND NOT t.is_completed
			     )
			   )
		)
		SELECT kind, card_id, task_id, title, due_at, start_at, end_at, description,
		       is_completed, board_id, board_name, project_id, project_name, list_name
		  FROM personal_items
		 WHERE ($6 OR due_at IS NOT NULL)
		   AND (due_at IS NULL OR (due_at >= $2 AND due_at <= $3))
		   AND ($4 = ''
		        OR lower(title) LIKE '%' || lower($4) || '%'
		        OR lower(board_name) LIKE '%' || lower($4) || '%'
		        OR lower(project_name) LIKE '%' || lower($4) || '%')
		 ORDER BY due_at IS NULL, due_at, title
		 LIMIT $5`, workspaceUsername, from, to, strings.TrimSpace(q), limit, includeUndated)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TimelineItem
	cardSet := map[string]struct{}{}
	for rows.Next() {
		var item TimelineItem
		if err := rows.Scan(
			&item.Kind,
			&item.CardID,
			&item.TaskID,
			&item.Title,
			&item.DueAt,
			&item.StartAt,
			&item.EndAt,
			&item.Description,
			&item.IsCompleted,
			&item.BoardID,
			&item.BoardName,
			&item.ProjectID,
			&item.ProjectName,
			&item.ListName,
		); err != nil {
			return nil, err
		}
		item.ID = item.Kind + ":" + item.CardID
		if item.TaskID != "" {
			item.ID = item.Kind + ":" + item.TaskID
		}
		out = append(out, item)
		cardSet[item.CardID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	members, err := s.listCardMembers(ctx, keys(cardSet))
	if err != nil {
		return nil, err
	}
	mutedCards, err := s.listMutedCardIDsForPlankaUser(ctx, workspaceUsername, keys(cardSet))
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Members = members[out[i].CardID]
		out[i].IsMuted = mutedCards[out[i].CardID]
	}
	return out, nil
}

func (s *Store) GetTimelineStats(ctx context.Context, workspaceUsername string, from, to time.Time) (TimelineStats, error) {
	workspaceUsername = normalize(workspaceUsername)
	var stats TimelineStats
	err := s.pool.QueryRow(ctx, `
		WITH me AS (
			SELECT id FROM user_account WHERE username = $1 LIMIT 1
		), completed_task_actions AS (
			SELECT a.id::text AS id
			  FROM action a
			  JOIN card c ON c.id = a.card_id
			  JOIN board b ON b.id = COALESCE(a.board_id, c.board_id)
			  JOIN project p ON p.id = b.project_id
			  JOIN me ON true
			  JOIN board_membership bm ON bm.board_id = b.id AND bm.user_id = me.id
			 WHERE COALESCE(a.created_at, now()) >= $2
			   AND COALESCE(a.created_at, now()) <= $3
			   AND NOT c.is_closed
			   AND NOT b.is_archived
			   AND NOT p.is_archived
			   AND a.type = 'completeTask'
			   AND a.user_id = me.id
		)
		SELECT count(*) FROM completed_task_actions`, workspaceUsername, from, to).Scan(&stats.Completed)
	return stats, err
}
