package server

import (
	"context"
	"net/http"
	"strings"
)

func (s *Server) handleAPITasks(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	from, to := parseRange(r)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 120)
	statsFrom, statsTo := parseStatsRange(r)
	items, err := s.store.ListTimelineItems(
		r.Context(),
		session.WorkspaceUsername,
		r.URL.Query().Get("q"),
		from,
		to,
		limit,
		r.URL.Query().Get("view") != "calendar",
	)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	stats, err := s.store.GetTimelineStats(
		r.Context(),
		session.WorkspaceUsername,
		statsFrom,
		statsTo,
	)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	s.attachWorkspaceURLs(items)
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"stats": stats,
	})
}

func (s *Server) handleAPITaskAction(w http.ResponseWriter, r *http.Request) {
	taskID, action, ok := splitActionPath(r.URL.Path, "/api/v1/tasks/")
	if !ok || action != "complete" {
		writeAPIError(w, http.StatusNotFound, "not_found")
		return
	}
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	if !s.store.UserCanCompleteTask(r.Context(), session.WorkspaceUsername, taskID, session.IsAdmin) {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return
	}
	if s.planka == nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}
	token, ok := s.createPlankaUserToken(w, r.Context(), session)
	if !ok {
		return
	}
	if err := s.planka.CompleteTaskAs(r.Context(), token, taskID); err != nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAPICardRoute(w http.ResponseWriter, r *http.Request) {
	cardID, action, ok := splitActionPath(r.URL.Path, "/api/v1/cards/")
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found")
		return
	}
	switch {
	case action == "" && r.Method == http.MethodGet:
		s.handleAPICard(w, r, cardID)
	case action == "complete" && r.Method == http.MethodPost:
		s.handleAPICardComplete(w, r, cardID)
	case action == "comments" && r.Method == http.MethodPost:
		s.handleAPICardComment(w, r, cardID)
	case action == "mute" && r.Method == http.MethodPost:
		s.handleAPICardMute(w, r, cardID)
	case action == "mute" && r.Method == http.MethodDelete:
		s.handleAPICardUnmute(w, r, cardID)
	default:
		writeAPIError(w, http.StatusNotFound, "not_found")
	}
}

func (s *Server) handleAPICard(w http.ResponseWriter, r *http.Request, cardID string) {
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	if !s.store.UserCanAccessCard(r.Context(), session.WorkspaceUsername, cardID, session.IsAdmin) {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return
	}
	card, err := s.store.GetCardDetail(
		r.Context(),
		cardID,
		session.WorkspaceUsername,
	)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	s.attachCardWorkspaceURLs(card)
	writeJSON(w, http.StatusOK, card)
}

func (s *Server) handleAPICardComplete(w http.ResponseWriter, r *http.Request, cardID string) {
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	target, ok := s.store.GetCardCompletionTarget(r.Context(), session.WorkspaceUsername, cardID, session.IsAdmin)
	if !ok {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return
	}
	if s.planka == nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}
	if target.UseDoneLabel && target.DoneLabelAttached {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	token, ok := s.createPlankaUserToken(w, r.Context(), session)
	if !ok {
		return
	}
	var err error
	if target.UseDoneLabel {
		err = s.planka.AddCardLabelAs(r.Context(), token, cardID, target.DoneLabelID)
	} else {
		err = s.planka.CompleteCardDueAs(r.Context(), token, cardID)
	}
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAPICardMute(w http.ResponseWriter, r *http.Request, cardID string) {
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	if !s.store.UserCanAccessCard(r.Context(), session.WorkspaceUsername, cardID, session.IsAdmin) {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := s.store.MuteCard(r.Context(), session.User.Username, cardID); err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true, "isMuted": true})
}

func (s *Server) handleAPICardUnmute(w http.ResponseWriter, r *http.Request, cardID string) {
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	if !s.store.UserCanAccessCard(r.Context(), session.WorkspaceUsername, cardID, session.IsAdmin) {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := s.store.UnmuteCard(r.Context(), session.User.Username, cardID); err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true, "isMuted": false})
}

func (s *Server) handleAPICardComment(w http.ResponseWriter, r *http.Request, cardID string) {
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	if !s.store.UserCanCommentCard(r.Context(), session.WorkspaceUsername, cardID, session.IsAdmin) {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return
	}
	var in commentRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		writeAPIError(w, http.StatusBadRequest, "empty_comment")
		return
	}
	if s.planka == nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}
	token, ok := s.createPlankaUserToken(w, r.Context(), session)
	if !ok {
		return
	}
	if err := s.planka.CreateCommentAs(r.Context(), token, cardID, text); err != nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) createPlankaUserToken(w http.ResponseWriter, ctx context.Context, session *apiSession) (string, bool) {
	if s.plankaSessionSecret == "" {
		writeAPIError(w, http.StatusBadGateway, "planka_user_auth_unavailable")
		return "", false
	}
	token, err := s.store.CreatePlankaSessionToken(
		ctx,
		session.WorkspaceUsername,
		s.plankaSessionSecret,
		plankaUserSessionTTL,
	)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return "", false
	}
	return token, true
}
