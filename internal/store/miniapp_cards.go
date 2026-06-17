package store

import "context"

func (s *Store) GetCardDetail(ctx context.Context, cardID, workspaceUsername string) (*CardDetail, error) {
	var c CardDetail
	err := s.pool.QueryRow(ctx, `
		SELECT c.id::text, c.name, COALESCE(c.description, ''), c.due_date,
		       c.start_date, c.end_date,
		       COALESCE(c.is_due_completed, false) OR done.is_completed_by_label,
		       b.id::text, b.name, p.id::text, p.name, COALESCE(l.name, '')
		  FROM card c
		  JOIN board b ON b.id = c.board_id
		  JOIN project p ON p.id = b.project_id
		  JOIN list l ON l.id = c.list_id
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
		 WHERE c.id::text = $1
		   AND NOT b.is_archived
		   AND NOT p.is_archived
		 LIMIT 1`, cardID).Scan(
		&c.ID,
		&c.Title,
		&c.Description,
		&c.DueAt,
		&c.StartAt,
		&c.EndAt,
		&c.IsDueCompleted,
		&c.BoardID,
		&c.BoardName,
		&c.ProjectID,
		&c.ProjectName,
		&c.ListName,
	)
	if err != nil {
		return nil, err
	}
	members, err := s.listCardMembers(ctx, []string{c.ID})
	if err != nil {
		return nil, err
	}
	c.Members = members[c.ID]
	tasks, err := s.listCardTasks(ctx, c.ID, workspaceUsername)
	if err != nil {
		return nil, err
	}
	c.Tasks = tasks
	comments, err := s.listCardComments(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.Comments = comments
	c.IsMuted = s.CardMutedForPlankaUser(ctx, workspaceUsername, c.ID)
	return &c, nil
}

func (s *Store) UserCanAccessCard(ctx context.Context, workspaceUsername, cardID string, isAdmin bool) bool {
	return s.hasCardAccess(ctx, workspaceUsername, cardID, "")
}

func (s *Store) UserCanCompleteCard(ctx context.Context, workspaceUsername, cardID string, isAdmin bool) bool {
	_, ok := s.GetCardCompletionTarget(ctx, workspaceUsername, cardID, isAdmin)
	return ok
}

func (s *Store) GetCardCompletionTarget(ctx context.Context, workspaceUsername, cardID string, isAdmin bool) (CardCompletionTarget, bool) {
	workspaceUsername = normalize(workspaceUsername)
	var target CardCompletionTarget
	err := s.pool.QueryRow(ctx, `
		WITH me AS (
			SELECT id FROM user_account WHERE username = $1 LIMIT 1
		), card_context AS (
			SELECT c.id,
			       c.board_id,
			       c.due_date IS NULL
			          AND c.start_date IS NULL
			          AND c.end_date IS NULL AS is_undated,
			       bm.role
			  FROM card c
			  JOIN board b ON b.id = c.board_id
			  JOIN project p ON p.id = b.project_id
			  JOIN me ON true
			  JOIN board_membership bm ON bm.board_id = c.board_id AND bm.user_id = me.id
			 WHERE c.id::text = $2
			   AND NOT b.is_archived
			   AND NOT p.is_archived
			 LIMIT 1
		), done_label AS (
			SELECT lb.id::text AS label_id
			  FROM card_context cc
			  JOIN label lb ON lb.board_id = cc.board_id
			 WHERE lower(trim(lb.name)) IN ('выполнено', 'виконано')
			 ORDER BY CASE lower(trim(lb.name))
			            WHEN 'виконано' THEN 0
			            ELSE 1
			          END,
			          lb.position,
			          lb.id::text
			 LIMIT 1
		)
		SELECT cc.is_undated,
		       COALESCE(dl.label_id, ''),
		       EXISTS (
		         SELECT 1
		           FROM card_label cl
		          WHERE cl.card_id = cc.id
		            AND cl.label_id::text = dl.label_id
		       )
		  FROM card_context cc
		  LEFT JOIN done_label dl ON true
		 WHERE (NOT cc.is_undated AND cc.role = 'editor')
		    OR (cc.is_undated AND cc.role IN ('editor', 'employee') AND dl.label_id IS NOT NULL)
		 LIMIT 1`, workspaceUsername, cardID).Scan(
		&target.UseDoneLabel,
		&target.DoneLabelID,
		&target.DoneLabelAttached,
	)
	return target, err == nil
}

func (s *Store) UserCanCompleteTask(ctx context.Context, workspaceUsername, taskID string, isAdmin bool) bool {
	workspaceUsername = normalize(workspaceUsername)
	var ok int
	err := s.pool.QueryRow(ctx, `
		WITH me AS (SELECT id FROM user_account WHERE username = $1 LIMIT 1)
		SELECT 1
		  FROM task t
		  JOIN task_list tl ON tl.id = t.task_list_id
		  JOIN card c ON c.id = tl.card_id
		  JOIN board b ON b.id = c.board_id
		  JOIN project p ON p.id = b.project_id
		  JOIN me ON true
		  JOIN board_membership bm ON bm.board_id = b.id AND bm.user_id = me.id
		 WHERE t.id::text = $2
		   AND t.assignee_user_id = me.id
		   AND NOT b.is_archived
		   AND NOT p.is_archived
		 LIMIT 1`, workspaceUsername, taskID).Scan(&ok)
	return err == nil && ok == 1
}

func (s *Store) UserCanCommentCard(ctx context.Context, workspaceUsername, cardID string, isAdmin bool) bool {
	workspaceUsername = normalize(workspaceUsername)
	var ok int
	err := s.pool.QueryRow(ctx, `
		WITH me AS (SELECT id FROM user_account WHERE username = $1 LIMIT 1)
		SELECT 1
		  FROM card c
		  JOIN board b ON b.id = c.board_id
		  JOIN project p ON p.id = b.project_id
		  JOIN me ON true
		  JOIN board_membership bm ON bm.board_id = b.id AND bm.user_id = me.id
		 WHERE c.id::text = $2
		   AND (bm.role IN ('editor', 'employee') OR bm.can_comment)
		   AND NOT b.is_archived
		   AND NOT p.is_archived
		 LIMIT 1`, workspaceUsername, cardID).Scan(&ok)
	return err == nil && ok == 1
}

func (s *Store) hasCardAccess(ctx context.Context, workspaceUsername, cardID, minRole string) bool {
	workspaceUsername = normalize(workspaceUsername)
	var ok int
	roleClause := ""
	if minRole == "editor" {
		roleClause = "AND bm.role = 'editor'"
	}
	err := s.pool.QueryRow(ctx, `
		WITH me AS (SELECT id FROM user_account WHERE username = $1 LIMIT 1)
		SELECT 1
		  FROM card c
		  JOIN board b ON b.id = c.board_id
		  JOIN project p ON p.id = b.project_id
		  JOIN me ON true
		  JOIN board_membership bm ON bm.board_id = b.id AND bm.user_id = me.id
		  LEFT JOIN card_membership cm ON cm.card_id = c.id AND cm.user_id = me.id
		 WHERE c.id::text = $2
		   AND NOT b.is_archived
		   AND NOT p.is_archived
		   AND (
		     (bm.user_id IS NOT NULL `+roleClause+`)
		     OR ($3 = '' AND cm.id IS NOT NULL)
		     OR ($3 = '' AND c.creator_user_id = me.id)
		     OR ($3 = '' AND EXISTS (
		       SELECT 1
		         FROM task_list tl
		         JOIN task t ON t.task_list_id = tl.id
		        WHERE tl.card_id = c.id AND t.assignee_user_id = me.id
		     ))
		   )
		 LIMIT 1`, workspaceUsername, cardID, minRole).Scan(&ok)
	return err == nil && ok == 1
}

func (s *Store) listCardMembers(ctx context.Context, cardIDs []string) (map[string][]MiniUser, error) {
	out := map[string][]MiniUser{}
	if len(cardIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT cm.card_id::text, ua.id::text, COALESCE(ua.username, ''), COALESCE(ua.name, ''),
		       COALESCE(ua.avatar->>'uploadedFileId', ''), COALESCE(ua.avatar->>'extension', '')
		  FROM card_membership cm
		  JOIN user_account ua ON ua.id = cm.user_id
		 WHERE cm.card_id::text = ANY($1)
		 ORDER BY cm.created_at`, cardIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cardID, id, username, name, avatarFileID, avatarExt string
		if err := rows.Scan(&cardID, &id, &username, &name, &avatarFileID, &avatarExt); err != nil {
			return nil, err
		}
		out[cardID] = append(out[cardID], miniUserWithAvatar(id, username, name, avatarFileID, avatarExt))
	}
	return out, rows.Err()
}

func (s *Store) listMutedCardIDsForPlankaUser(ctx context.Context, plankaUsername string, cardIDs []string) (map[string]bool, error) {
	out := map[string]bool{}
	plankaUsername = normalize(plankaUsername)
	if plankaUsername == "" || len(cardIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT mc.card_id
		  FROM monitor_muted_cards mc
		  JOIN monitor_users mu ON mu.telegram_username = mc.telegram_username
		 WHERE mc.card_id = ANY($2)
		   AND (
		     mu.planka_username = $1
		     OR (mu.telegram_username = $1 AND (mu.planka_username IS NULL OR mu.planka_username = ''))
		   )`, plankaUsername, cardIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			return nil, err
		}
		out[cardID] = true
	}
	return out, rows.Err()
}

func (s *Store) listCardTasks(ctx context.Context, cardID, workspaceUsername string) ([]CardTask, error) {
	workspaceUsername = normalize(workspaceUsername)
	rows, err := s.pool.Query(ctx, `
		WITH me AS (SELECT id FROM user_account WHERE username = $2 LIMIT 1)
		SELECT t.id::text, t.name, t.is_completed,
		       COALESCE(ua.id::text, ''), COALESCE(ua.username, ''), COALESCE(ua.name, ''),
		       COALESCE(ua.avatar->>'uploadedFileId', ''), COALESCE(ua.avatar->>'extension', '')
		  FROM task_list tl
		  JOIN task t ON t.task_list_id = tl.id
		  LEFT JOIN user_account ua ON ua.id = t.assignee_user_id
		 WHERE tl.card_id::text = $1
		   AND t.assignee_user_id = (SELECT id FROM me)
		 ORDER BY tl.position, t.position`, cardID, workspaceUsername)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CardTask
	for rows.Next() {
		var task CardTask
		var id, username, name, avatarFileID, avatarExt string
		if err := rows.Scan(&task.ID, &task.Name, &task.IsCompleted, &id, &username, &name, &avatarFileID, &avatarExt); err != nil {
			return nil, err
		}
		if id != "" {
			assignee := miniUserWithAvatar(id, username, name, avatarFileID, avatarExt)
			task.Assignee = &assignee
		}
		out = append(out, task)
	}
	return out, rows.Err()
}

func (s *Store) listCardComments(ctx context.Context, cardID string) ([]CardComment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cm.id::text, cm.text, COALESCE(cm.created_at, now()),
		       COALESCE(ua.id::text, ''), COALESCE(ua.username, ''), COALESCE(ua.name, ''),
		       COALESCE(ua.avatar->>'uploadedFileId', ''), COALESCE(ua.avatar->>'extension', '')
		  FROM comment cm
		  LEFT JOIN user_account ua ON ua.id = cm.user_id
		 WHERE cm.card_id::text = $1
		 ORDER BY cm.created_at DESC
		 LIMIT 20`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CardComment
	for rows.Next() {
		var item CardComment
		var id, username, name, avatarFileID, avatarExt string
		if err := rows.Scan(&item.ID, &item.Text, &item.CreatedAt, &id, &username, &name, &avatarFileID, &avatarExt); err != nil {
			return nil, err
		}
		item.Author = miniUserWithAvatar(id, username, name, avatarFileID, avatarExt)
		out = append(out, item)
	}
	return out, rows.Err()
}
