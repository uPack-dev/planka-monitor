package server

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"planka-monitor/internal/store"
)

func (s *Server) authenticateAPI(w http.ResponseWriter, r *http.Request) (telegramInitUser, bool) {
	if user, attempted, ok := s.authenticateDevAPI(r); attempted {
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized")
			return telegramInitUser{}, false
		}
		return user, true
	}

	user, err := validateTelegramInitData(s.botToken, r.Header.Get(telegramInitDataHeader), time.Now())
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized")
		return telegramInitUser{}, false
	}
	if user.Username != "" && s.store.IsBlocked(r.Context(), user.Username) {
		writeAPIError(w, http.StatusForbidden, "blocked")
		return telegramInitUser{}, false
	}
	return user, true
}

func (s *Server) authenticateDevAPI(r *http.Request) (telegramInitUser, bool, bool) {
	token := strings.TrimSpace(r.Header.Get(monitorDevAuthHeader))
	if token == "" {
		token = strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	}
	if token == "" {
		return telegramInitUser{}, false, false
	}
	if s.devAuth.Token == "" {
		return telegramInitUser{}, false, false
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.devAuth.Token)) != 1 {
		return telegramInitUser{}, true, false
	}
	username := normalizeAPIUsername(s.devAuth.TelegramUsername)
	if username == "" {
		username = normalizeAPIUsername(s.devAuth.WorkspaceUsername)
	}
	if username == "" {
		return telegramInitUser{}, true, false
	}
	return telegramInitUser{
		ID:        -1,
		FirstName: "Monitor Dev",
		Username:  username,
		IsDev:     true,
	}, true, true
}

func (s *Server) requireWorkspace(w http.ResponseWriter, r *http.Request) (*apiSession, bool) {
	user, ok := s.authenticateAPI(w, r)
	if !ok {
		return nil, false
	}
	if user.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "username_required")
		return nil, false
	}
	b, workspaceUsername, err := s.bindingForAPIUser(r.Context(), user)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return nil, false
	}
	if b == nil {
		writeAPIError(w, http.StatusBadRequest, "needs_start")
		return nil, false
	}
	if !s.plankaUserExists(r.Context(), workspaceUsername) {
		writeAPIError(w, http.StatusBadRequest, "no_workspace")
		return nil, false
	}
	return &apiSession{
		User:              user,
		Binding:           b,
		WorkspaceUsername: workspaceUsername,
		IsAdmin:           s.isAPIAdmin(r.Context(), user),
	}, true
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) (*apiSession, bool) {
	user, ok := s.authenticateAPI(w, r)
	if !ok {
		return nil, false
	}
	if user.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "username_required")
		return nil, false
	}
	b, workspaceUsername, err := s.bindingForAPIUser(r.Context(), user)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return nil, false
	}
	if b == nil {
		writeAPIError(w, http.StatusBadRequest, "needs_start")
		return nil, false
	}
	isAdmin := s.isAPIAdmin(r.Context(), user)
	if !isAdmin {
		writeAPIError(w, http.StatusForbidden, "forbidden")
		return nil, false
	}
	return &apiSession{
		User:              user,
		Binding:           b,
		WorkspaceUsername: workspaceUsername,
		IsAdmin:           isAdmin,
	}, true
}

func (s *Server) requireRegistered(w http.ResponseWriter, ctx context.Context, username string) bool {
	b, err := s.store.GetByTelegram(ctx, username)
	if err != nil {
		s.writeAPIStoreError(w, err)
		return false
	}
	if b == nil {
		writeAPIError(w, http.StatusBadRequest, "needs_start")
		return false
	}
	return true
}

func (s *Server) bindingForAPIUser(ctx context.Context, user telegramInitUser) (*store.UserBinding, string, error) {
	b, err := s.store.GetByTelegram(ctx, user.Username)
	if err != nil {
		return nil, "", err
	}

	devWorkspace := ""
	if user.IsDev {
		devWorkspace = s.devAuth.WorkspaceUsername
	}
	if devWorkspace != "" {
		if b == nil {
			return &store.UserBinding{
				TelegramUsername:     user.Username,
				TelegramUserID:       user.ID,
				PlankaUsername:       devWorkspace,
				NotificationsEnabled: true,
				OnboardingCompleted:  true,
				Preferences:          store.DefaultNotificationPreferences(),
			}, devWorkspace, nil
		}
		devBinding := *b
		devBinding.PlankaUsername = devWorkspace
		return &devBinding, devWorkspace, nil
	}

	if b == nil {
		return nil, "", nil
	}
	workspaceUsername := b.PlankaUsername
	if workspaceUsername == "" {
		workspaceUsername = b.TelegramUsername
	}
	return b, workspaceUsername, nil
}

func (s *Server) isAPIAdmin(ctx context.Context, user telegramInitUser) bool {
	return s.store.IsAdmin(ctx, user.Username) || (user.IsDev && s.devAuth.IsAdmin)
}
