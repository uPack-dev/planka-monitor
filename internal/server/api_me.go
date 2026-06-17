package server

import (
	"context"
	"errors"
	"net/http"

	"planka-monitor/internal/store"
)

func (s *Server) handleAPIMe(w http.ResponseWriter, r *http.Request) {
	user, ok := s.authenticateAPI(w, r)
	if !ok {
		return
	}
	me, err := s.buildMe(r.Context(), user)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, me)
}

func (s *Server) handleAPIAvatar(w http.ResponseWriter, r *http.Request) {
	if s.planka == nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}
	fileID, variant, ext, ok := parseAvatarPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found")
		return
	}
	data, err := s.planka.DownloadAvatar(r.Context(), fileID, ext, variant)
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, "planka_unavailable")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Type", avatarContentType(ext))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleAPIBind(w http.ResponseWriter, r *http.Request) {
	user, ok := s.authenticateAPI(w, r)
	if !ok {
		return
	}
	if user.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "username_required")
		return
	}
	if !s.requireRegistered(w, r.Context(), user.Username) {
		return
	}

	var in bindRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := s.store.BindPlankaIfAvailable(r.Context(), user.Username, in.WorkspaceUsername); err != nil {
		switch {
		case errors.Is(err, store.ErrPlankaUserNotFound):
			writeAPIError(w, http.StatusNotFound, "not_found")
		case errors.Is(err, store.ErrWorkspaceTaken):
			writeAPIError(w, http.StatusConflict, "already_bound")
		default:
			s.writeAPIStoreError(w, err)
		}
		return
	}
	s.writeUpdatedMe(w, r, user)
}

func (s *Server) handleAPIPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := s.authenticateAPI(w, r)
	if !ok {
		return
	}
	if user.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "username_required")
		return
	}
	if !s.requireRegistered(w, r.Context(), user.Username) {
		return
	}

	var prefs store.NotificationPreferences
	if !decodeJSON(w, r, &prefs) {
		return
	}
	if err := s.store.SetNotificationPreferences(r.Context(), user.Username, prefs); err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	s.writeUpdatedMe(w, r, user)
}

func (s *Server) handleAPINotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := s.authenticateAPI(w, r)
	if !ok {
		return
	}
	if user.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "username_required")
		return
	}
	if !s.requireRegistered(w, r.Context(), user.Username) {
		return
	}

	var in notificationsRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := s.store.SetNotifications(r.Context(), user.Username, in.Enabled); err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	s.writeUpdatedMe(w, r, user)
}

func (s *Server) handleAPIOnboarding(w http.ResponseWriter, r *http.Request) {
	user, ok := s.authenticateAPI(w, r)
	if !ok {
		return
	}
	if user.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "username_required")
		return
	}
	if !s.requireRegistered(w, r.Context(), user.Username) {
		return
	}

	var in onboardingRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := s.store.SetOnboardingCompleted(r.Context(), user.Username, in.Completed); err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	s.writeUpdatedMe(w, r, user)
}

func (s *Server) writeUpdatedMe(w http.ResponseWriter, r *http.Request, user telegramInitUser) {
	me, err := s.buildMe(r.Context(), user)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, me)
}

func (s *Server) buildMe(ctx context.Context, user telegramInitUser) (meResponse, error) {
	resp := meResponse{
		TelegramUsername: user.Username,
		TelegramUserID:   user.ID,
		IsDevAuth:        user.IsDev,
		Preferences:      store.DefaultNotificationPreferences(),
		NeedsStart:       true,
	}
	if user.Username == "" {
		return resp, nil
	}

	b, workspaceUsername, err := s.bindingForAPIUser(ctx, user)
	if err != nil {
		return resp, err
	}
	if b == nil {
		resp.WorkspaceUsername = user.Username
		resp.WorkspaceExists = s.plankaUserExists(ctx, user.Username)
		resp.WorkspaceDisplayName = s.store.GetPlankaDisplayName(ctx, user.Username)
		resp.WorkspaceAvatarURL = s.workspaceAvatarURL(ctx, user.Username)
		resp.WorkspaceURL = s.workspaceURL()
		return resp, nil
	}

	resp.TelegramUsername = b.TelegramUsername
	if b.TelegramUserID != 0 {
		resp.TelegramUserID = b.TelegramUserID
	}
	resp.WorkspaceUsername = workspaceUsername
	resp.WorkspaceExists = s.plankaUserExists(ctx, workspaceUsername)
	resp.WorkspaceDisplayName = s.store.GetPlankaDisplayName(ctx, workspaceUsername)
	resp.WorkspaceAvatarURL = s.workspaceAvatarURL(ctx, workspaceUsername)
	resp.WorkspaceURL = s.workspaceURL()
	resp.IsAdmin = s.isAPIAdmin(ctx, user)
	resp.NotificationsEnabled = b.NotificationsEnabled
	resp.Preferences = b.Preferences
	resp.NeedsStart = false
	resp.OnboardingCompleted = b.OnboardingCompleted
	return resp, nil
}
