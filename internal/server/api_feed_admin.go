package server

import (
	"net/http"
	"time"
)

func (s *Server) handleAPIFeed(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	var before *time.Time
	if raw := r.URL.Query().Get("before"); raw != "" {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			before = &parsed
		}
	}
	events, err := s.store.ListFeedEvents(
		r.Context(),
		session.WorkspaceUsername,
		r.URL.Query().Get("type"),
		parsePositiveInt(r.URL.Query().Get("limit"), 50),
		before,
	)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	s.attachFeedWorkspaceURLs(events)
	writeJSON(w, http.StatusOK, map[string]any{"items": events})
}

func (s *Server) handleAPIFeedRead(w http.ResponseWriter, r *http.Request) {
	feedID, action, ok := splitActionPath(r.URL.Path, "/api/v1/feed/")
	if !ok || action != "read" {
		writeAPIError(w, http.StatusNotFound, "not_found")
		return
	}
	session, ok := s.requireWorkspace(w, r)
	if !ok {
		return
	}
	if err := s.store.MarkFeedRead(r.Context(), session.WorkspaceUsername, feedID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAPIAdminSubscribers(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAdmin(w, r)
	if !ok || session == nil {
		return
	}
	items, err := s.store.ListAdminSubscribers(
		r.Context(),
		r.URL.Query().Get("q"),
		r.URL.Query().Get("status"),
	)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	attachAdminSubscriberAvatarURLs(items)
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleAPIAdminSubscriberAction(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	target, action, ok := splitActionPath(r.URL.Path, "/api/v1/admin/subscribers/")
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found")
		return
	}
	var in enabledRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	switch action {
	case "block":
		if s.store.IsAdmin(r.Context(), target) && in.Enabled {
			writeAPIError(w, http.StatusForbidden, "cannot_block_admin")
			return
		}
		if err := s.store.SetBlocked(r.Context(), target, in.Enabled); err != nil {
			s.writeAPIStoreError(w, err)
			return
		}
	case "admin":
		if err := s.store.SetMonitorAdmin(r.Context(), target, in.Enabled); err != nil {
			s.writeAPIStoreError(w, err)
			return
		}
	default:
		writeAPIError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
