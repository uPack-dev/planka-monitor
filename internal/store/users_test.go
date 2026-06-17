package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBindPlankaIfAvailableIntegration(t *testing.T) {
	ctx := context.Background()
	st := newIntegrationStore(t)
	seedPlankaUser(t, st, "1", "yan")
	seedPlankaUser(t, st, "2", "ira")

	if err := st.SetUser(ctx, "telegram_yan", 100, 10); err != nil {
		t.Fatalf("set yan: %v", err)
	}
	if err := st.SetUser(ctx, "telegram_ira", 101, 11); err != nil {
		t.Fatalf("set ira: %v", err)
	}
	if err := st.BindPlankaIfAvailable(ctx, "telegram_yan", "yan"); err != nil {
		t.Fatalf("bind yan: %v", err)
	}
	if err := st.BindPlankaIfAvailable(ctx, "telegram_ira", "yan"); !errors.Is(err, ErrWorkspaceTaken) {
		t.Fatalf("expected workspace conflict, got %v", err)
	}
	if err := st.BindPlankaIfAvailable(ctx, "telegram_ira", "missing"); !errors.Is(err, ErrPlankaUserNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestNotificationPreferencesIntegration(t *testing.T) {
	ctx := context.Background()
	st := newIntegrationStore(t)
	seedPlankaUser(t, st, "1", "yan")

	if err := st.SetUser(ctx, "telegram_yan", 100, 10); err != nil {
		t.Fatalf("set user: %v", err)
	}
	if err := st.BindPlankaIfAvailable(ctx, "telegram_yan", "yan"); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if err := st.SetNotifications(ctx, "telegram_yan", true); err != nil {
		t.Fatalf("enable notifications: %v", err)
	}

	b, err := st.GetByTelegram(ctx, "telegram_yan")
	if err != nil {
		t.Fatalf("get binding: %v", err)
	}
	if b == nil || b.Preferences != DefaultNotificationPreferences() {
		t.Fatalf("unexpected defaults: %+v", b)
	}
	if b.OnboardingCompleted {
		t.Fatalf("new user should start with incomplete onboarding")
	}
	if err := st.SetOnboardingCompleted(ctx, "telegram_yan", true); err != nil {
		t.Fatalf("complete onboarding: %v", err)
	}
	b, err = st.GetByTelegram(ctx, "telegram_yan")
	if err != nil {
		t.Fatalf("get completed onboarding binding: %v", err)
	}
	if b == nil || !b.OnboardingCompleted {
		t.Fatalf("onboarding completion was not persisted: %+v", b)
	}
	if !st.PersonalCategoryEnabled(ctx, "yan", "assignments") {
		t.Fatalf("assignments should be enabled by default")
	}
	if st.PersonalCategoryEnabled(ctx, "yan", "changes") {
		t.Fatalf("changes should be disabled by default for new users")
	}

	prefs := NotificationPreferences{
		Assignments: false,
		Comments:    true,
		Changes:     true,
		Done:        false,
	}
	if err := st.SetNotificationPreferences(ctx, "telegram_yan", prefs); err != nil {
		t.Fatalf("save preferences: %v", err)
	}
	b, err = st.GetByTelegram(ctx, "telegram_yan")
	if err != nil {
		t.Fatalf("get updated binding: %v", err)
	}
	if b == nil || b.Preferences != prefs {
		t.Fatalf("preferences were not persisted: %+v", b)
	}
	if st.PersonalCategoryEnabled(ctx, "yan", "assignments") {
		t.Fatalf("assignments should be disabled after update")
	}
	if !st.PersonalCategoryEnabled(ctx, "yan", "changes") {
		t.Fatalf("changes should be enabled after update")
	}

	if err := st.SetNotifications(ctx, "telegram_yan", false); err != nil {
		t.Fatalf("disable notifications: %v", err)
	}
	if st.PersonalCategoryEnabled(ctx, "yan", "changes") {
		t.Fatalf("master notification toggle should suppress personal categories")
	}
}

func TestCardMuteToggleIntegration(t *testing.T) {
	ctx := context.Background()
	st := newIntegrationStore(t)
	seedPlankaUser(t, st, "1", "yan")

	if err := st.SetUser(ctx, "telegram_yan", 100, 10); err != nil {
		t.Fatalf("set user: %v", err)
	}
	if err := st.BindPlankaIfAvailable(ctx, "telegram_yan", "yan"); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if err := st.MuteCard(ctx, "telegram_yan", "card_1"); err != nil {
		t.Fatalf("mute card: %v", err)
	}
	if !st.CardMutedForPlankaUser(ctx, "yan", "card_1") {
		t.Fatalf("card should be muted after MuteCard")
	}
	if err := st.UnmuteCard(ctx, "telegram_yan", "card_1"); err != nil {
		t.Fatalf("unmute card: %v", err)
	}
	if st.CardMutedForPlankaUser(ctx, "yan", "card_1") {
		t.Fatalf("card should not be muted after UnmuteCard")
	}
}

func TestListTimelineItemsTreatsUndatedDoneLabelAsCompletedIntegration(t *testing.T) {
	ctx := context.Background()
	st := newIntegrationStore(t)
	createMiniAppTimelineSchema(t, st)
	seedPlankaUser(t, st, "1", "yan")

	if _, err := st.pool.Exec(ctx, `
		INSERT INTO project (id, name) VALUES ('p1', 'Project');
		INSERT INTO board (id, project_id, name) VALUES ('b1', 'p1', 'Board');
		INSERT INTO list (id, board_id, name) VALUES ('l1', 'b1', 'Backlog');
		INSERT INTO board_membership (board_id, user_id, role) VALUES ('b1', '1', 'editor');
		INSERT INTO label (id, board_id, name, color) VALUES ('lb_done', 'b1', 'Выполнено', 'green');
		INSERT INTO card (id, board_id, list_id, name, description, is_closed, is_due_completed)
		  VALUES
		    ('c_done', 'b1', 'l1', 'Done undated', '', false, false),
		    ('c_open', 'b1', 'l1', 'Open undated', '', false, false),
		    ('c_due_done', 'b1', 'l1', 'Due with done label', '', false, false);
		UPDATE card SET due_date = '2026-06-13 12:00:00' WHERE id = 'c_due_done';
		INSERT INTO card_membership (card_id, user_id)
		  VALUES ('c_done', '1'), ('c_open', '1'), ('c_due_done', '1');
		INSERT INTO card_label (id, card_id, label_id)
		  VALUES ('cl_done', 'c_done', 'lb_done'), ('cl_due_done', 'c_due_done', 'lb_done');
	`); err != nil {
		t.Fatalf("seed timeline: %v", err)
	}

	items, err := st.ListTimelineItems(
		ctx,
		"yan",
		"",
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC),
		20,
		true,
	)
	if err != nil {
		t.Fatalf("list timeline items: %v", err)
	}

	got := map[string]bool{}
	for _, item := range items {
		got[item.CardID] = true
	}
	if got["c_done"] {
		t.Fatalf("undated card with done label should be hidden: %+v", items)
	}
	if !got["c_open"] {
		t.Fatalf("undated card without done label should stay visible: %+v", items)
	}
	if !got["c_due_done"] {
		t.Fatalf("dated card with done label should stay visible: %+v", items)
	}

	doneDetail, err := st.GetCardDetail(ctx, "c_done", "yan")
	if err != nil {
		t.Fatalf("get done detail: %v", err)
	}
	if !doneDetail.IsDueCompleted {
		t.Fatalf("undated card with done label should be treated as completed")
	}

	dueDetail, err := st.GetCardDetail(ctx, "c_due_done", "yan")
	if err != nil {
		t.Fatalf("get due detail: %v", err)
	}
	if dueDetail.IsDueCompleted {
		t.Fatalf("dated card with done label should not be auto-completed")
	}
}

func TestCardCompletionTargetUsesDoneLabelForUndatedEmployeeIntegration(t *testing.T) {
	ctx := context.Background()
	st := newIntegrationStore(t)
	createMiniAppTimelineSchema(t, st)
	seedPlankaUser(t, st, "1", "yan")

	if _, err := st.pool.Exec(ctx, `
		INSERT INTO project (id, name) VALUES ('p1', 'Project');
		INSERT INTO board (id, project_id, name) VALUES ('b1', 'p1', 'Board');
		INSERT INTO list (id, board_id, name) VALUES ('l1', 'b1', 'Backlog');
		INSERT INTO board_membership (board_id, user_id, role) VALUES ('b1', '1', 'employee');
		INSERT INTO label (id, board_id, name, color, position) VALUES ('lb_done', 'b1', 'Виконано', 'wet-moss', 1);
		INSERT INTO card (id, board_id, list_id, name, description, is_closed, is_due_completed)
		  VALUES
		    ('c_undated', 'b1', 'l1', 'Undated', '', false, false),
		    ('c_dated', 'b1', 'l1', 'Dated', '', false, false);
		UPDATE card SET due_date = '2026-06-13 12:00:00' WHERE id = 'c_dated';
		INSERT INTO card_membership (card_id, user_id)
		  VALUES ('c_undated', '1'), ('c_dated', '1');
	`); err != nil {
		t.Fatalf("seed completion target: %v", err)
	}

	target, ok := st.GetCardCompletionTarget(ctx, "yan", "c_undated", false)
	if !ok {
		t.Fatalf("employee should be allowed to complete undated card through done label")
	}
	if !target.UseDoneLabel || target.DoneLabelID != "lb_done" || target.DoneLabelAttached {
		t.Fatalf("unexpected undated target: %+v", target)
	}

	if st.UserCanCompleteCard(ctx, "yan", "c_dated", false) {
		t.Fatalf("employee should not complete dated card through due completion")
	}

	if _, err := st.pool.Exec(ctx, `
		UPDATE board_membership SET role = 'editor' WHERE board_id = 'b1' AND user_id = '1'
	`); err != nil {
		t.Fatalf("promote to editor: %v", err)
	}
	target, ok = st.GetCardCompletionTarget(ctx, "yan", "c_dated", false)
	if !ok {
		t.Fatalf("editor should be allowed to complete dated card")
	}
	if target.UseDoneLabel || target.DoneLabelID != "" {
		t.Fatalf("dated card should use due completion, got: %+v", target)
	}
}

func TestMiniAppExcludesArchivedBoardsIntegration(t *testing.T) {
	ctx := context.Background()
	st := newIntegrationStore(t)
	createMiniAppTimelineSchema(t, st)
	seedPlankaUser(t, st, "1", "yan")

	if _, err := st.pool.Exec(ctx, `
		INSERT INTO project (id, name) VALUES ('p1', 'Project');
		INSERT INTO board (id, project_id, name, is_archived)
		  VALUES ('b_active', 'p1', 'Active board', false),
		         ('b_archived', 'p1', 'Archived board', true);
		INSERT INTO list (id, board_id, name)
		  VALUES ('l_active', 'b_active', 'Backlog'),
		         ('l_archived', 'b_archived', 'Backlog');
		INSERT INTO board_membership (board_id, user_id, role)
		  VALUES ('b_active', '1', 'editor'),
		         ('b_archived', '1', 'editor');
		INSERT INTO card (id, board_id, list_id, name, description, is_closed, is_due_completed)
		  VALUES ('c_active', 'b_active', 'l_active', 'Active card', '', false, false),
		         ('c_archived', 'b_archived', 'l_archived', 'Archived card', '', false, false);
		INSERT INTO card_membership (card_id, user_id)
		  VALUES ('c_active', '1'),
		         ('c_archived', '1');
		INSERT INTO notification (
			id, user_id, creator_user_id, board_id, card_id, type, data, is_read, created_at
		) VALUES (
			'n_active', '1', '1', 'b_active', 'c_active', 'commentCard', '{"text":"active"}', false, now()
		), (
			'n_archived', '1', '1', 'b_archived', 'c_archived', 'commentCard', '{"text":"archived"}', false, now()
		);
	`); err != nil {
		t.Fatalf("seed archived boards: %v", err)
	}

	items, err := st.ListTimelineItems(
		ctx,
		"yan",
		"",
		time.Now().Add(-24*time.Hour),
		time.Now().Add(24*time.Hour),
		20,
		true,
	)
	if err != nil {
		t.Fatalf("list timeline items: %v", err)
	}
	gotItems := map[string]bool{}
	for _, item := range items {
		gotItems[item.CardID] = true
	}
	if !gotItems["c_active"] {
		t.Fatalf("active board card should be visible: %+v", items)
	}
	if gotItems["c_archived"] {
		t.Fatalf("archived board card should be hidden: %+v", items)
	}

	if !st.UserCanAccessCard(ctx, "yan", "c_active", false) {
		t.Fatalf("active board card should be directly accessible")
	}
	if st.UserCanAccessCard(ctx, "yan", "c_archived", false) {
		t.Fatalf("archived board card should not be directly accessible")
	}
	if st.UserCanCommentCard(ctx, "yan", "c_archived", false) {
		t.Fatalf("archived board card should not allow comments")
	}
	if st.UserCanCompleteCard(ctx, "yan", "c_archived", false) {
		t.Fatalf("archived board card should not allow completion")
	}
	if _, err := st.GetCardDetail(ctx, "c_archived", "yan"); err == nil {
		t.Fatalf("archived board card detail should not load")
	}

	events, err := st.ListFeedEvents(ctx, "yan", "all", 20, nil)
	if err != nil {
		t.Fatalf("list feed events: %v", err)
	}
	gotEvents := map[string]bool{}
	for _, event := range events {
		gotEvents[event.CardID] = true
	}
	if !gotEvents["c_active"] {
		t.Fatalf("active board feed event should be visible: %+v", events)
	}
	if gotEvents["c_archived"] {
		t.Fatalf("archived board feed event should be hidden: %+v", events)
	}
}

func newIntegrationStore(t *testing.T) *Store {
	t.Helper()

	dsn := os.Getenv("MONITOR_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MONITOR_TEST_DATABASE_URL to run store integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	base, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	schema := "monitor_test_" + strings.ReplaceAll(fmt.Sprint(time.Now().UnixNano()), "-", "_")
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := base.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = base.Close(ctx)
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		_, _ = base.Exec(cleanupCtx, "DROP SCHEMA "+quotedSchema+" CASCADE")
		_ = base.Close(cleanupCtx)
	})

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	st := &Store{pool: pool, envAdmins: map[string]struct{}{}}
	t.Cleanup(st.Close)

	if err := migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		CREATE TABLE user_account (
			id       text PRIMARY KEY,
			username text NOT NULL UNIQUE,
			name     text NOT NULL DEFAULT '',
			avatar   jsonb NOT NULL DEFAULT '{}'::jsonb
		)`); err != nil {
		t.Fatalf("create user_account: %v", err)
	}

	return st
}

func createMiniAppTimelineSchema(t *testing.T, st *Store) {
	t.Helper()

	if _, err := st.pool.Exec(context.Background(), `
		CREATE TABLE project (
			id          text PRIMARY KEY,
			name        text NOT NULL,
			is_archived boolean NOT NULL DEFAULT false
		);
		CREATE TABLE board (
			id          text PRIMARY KEY,
			project_id  text NOT NULL,
			name        text NOT NULL,
			is_archived boolean NOT NULL DEFAULT false
		);
		CREATE TABLE list (
			id       text PRIMARY KEY,
			board_id text NOT NULL,
			name     text NOT NULL
		);
		CREATE TABLE board_membership (
			board_id text NOT NULL,
			user_id  text NOT NULL,
			role     text NOT NULL DEFAULT 'editor',
			can_comment boolean NOT NULL DEFAULT true
		);
		CREATE TABLE card (
			id               text PRIMARY KEY,
			board_id         text NOT NULL,
			list_id          text NOT NULL,
			creator_user_id  text,
			name             text NOT NULL,
			description      text,
			due_date         timestamp,
			start_date       timestamp,
			end_date         timestamp,
			is_closed        boolean NOT NULL DEFAULT false,
			is_due_completed boolean NOT NULL DEFAULT false
		);
		CREATE TABLE card_membership (
			id         text NOT NULL DEFAULT '',
			card_id    text NOT NULL,
			user_id    text NOT NULL,
			created_at timestamp NOT NULL DEFAULT now()
		);
		CREATE TABLE task_list (
			id       text PRIMARY KEY,
			card_id  text NOT NULL,
			position double precision NOT NULL DEFAULT 0
		);
		CREATE TABLE task (
			id               text PRIMARY KEY,
			task_list_id     text NOT NULL,
			assignee_user_id text,
			name             text NOT NULL,
			is_completed     boolean NOT NULL DEFAULT false,
			position         double precision NOT NULL DEFAULT 0
		);
		CREATE TABLE label (
			id       text PRIMARY KEY,
			board_id text NOT NULL,
			name     text,
			color    text,
			position double precision NOT NULL DEFAULT 0
		);
		CREATE TABLE card_label (
			id         text PRIMARY KEY,
			card_id    text NOT NULL,
			label_id   text NOT NULL,
			created_at timestamp NOT NULL DEFAULT now()
		);
		CREATE TABLE comment (
			id         text PRIMARY KEY,
			card_id    text NOT NULL,
			user_id    text,
			text       text NOT NULL,
			created_at timestamp NOT NULL DEFAULT now()
		);
		CREATE TABLE action (
			id         text PRIMARY KEY,
			card_id    text NOT NULL,
			board_id   text,
			user_id    text,
			type       text NOT NULL,
			data       jsonb NOT NULL DEFAULT '{}'::jsonb,
			created_at timestamp NOT NULL DEFAULT now(),
			updated_at timestamp NOT NULL DEFAULT now()
		);
		CREATE TABLE notification (
			id              text PRIMARY KEY,
			user_id         text NOT NULL,
			creator_user_id text,
			board_id        text,
			card_id         text,
			comment_id      text,
			action_id       text,
			type            text NOT NULL,
			data            jsonb NOT NULL DEFAULT '{}'::jsonb,
			is_read         boolean NOT NULL DEFAULT false,
			created_at      timestamp NOT NULL DEFAULT now(),
			updated_at      timestamp NOT NULL DEFAULT now()
		);
	`); err != nil {
		t.Fatalf("create mini app timeline schema: %v", err)
	}
}

func seedPlankaUser(t *testing.T, st *Store, id, username string) {
	t.Helper()

	if _, err := st.pool.Exec(context.Background(),
		`INSERT INTO user_account (id, username) VALUES ($1, $2)`,
		id, username); err != nil {
		t.Fatalf("seed planka user %q: %v", username, err)
	}
}
