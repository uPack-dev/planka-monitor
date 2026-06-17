package server

import (
	"net/http"
	"strings"
	"time"

	"planka-monitor/internal/store"
)

const (
	telegramInitDataHeader = "X-Telegram-Init-Data"
	monitorDevAuthHeader   = "X-Monitor-Dev-Auth"
	plankaUserSessionTTL   = 10 * time.Minute
)

type meResponse struct {
	TelegramUsername     string                        `json:"telegramUsername"`
	TelegramUserID       int64                         `json:"telegramUserId"`
	WorkspaceUsername    string                        `json:"workspaceUsername"`
	WorkspaceDisplayName string                        `json:"workspaceDisplayName,omitempty"`
	WorkspaceAvatarURL   string                        `json:"workspaceAvatarUrl,omitempty"`
	WorkspaceURL         string                        `json:"workspaceUrl,omitempty"`
	WorkspaceExists      bool                          `json:"workspaceExists"`
	IsAdmin              bool                          `json:"isAdmin"`
	IsDevAuth            bool                          `json:"devAuth"`
	NotificationsEnabled bool                          `json:"notificationsEnabled"`
	Preferences          store.NotificationPreferences `json:"preferences"`
	NeedsStart           bool                          `json:"needsStart"`
	OnboardingCompleted  bool                          `json:"onboardingCompleted"`
}

type bindRequest struct {
	WorkspaceUsername string `json:"workspaceUsername"`
}

type notificationsRequest struct {
	Enabled bool `json:"enabled"`
}

type onboardingRequest struct {
	Completed bool `json:"completed"`
}

type commentRequest struct {
	Text string `json:"text"`
}

type enabledRequest struct {
	Enabled bool `json:"enabled"`
}

type apiSession struct {
	User              telegramInitUser
	Binding           *store.UserBinding
	WorkspaceUsername string
	IsAdmin           bool
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	switch {
	case r.URL.Path == "/api/v1/me" && r.Method == http.MethodGet:
		s.handleAPIMe(w, r)
	case r.URL.Path == "/api/v1/me/bind" && r.Method == http.MethodPost:
		s.handleAPIBind(w, r)
	case r.URL.Path == "/api/v1/me/preferences" && r.Method == http.MethodPatch:
		s.handleAPIPreferences(w, r)
	case r.URL.Path == "/api/v1/me/notifications" && r.Method == http.MethodPatch:
		s.handleAPINotifications(w, r)
	case r.URL.Path == "/api/v1/me/onboarding" && r.Method == http.MethodPatch:
		s.handleAPIOnboarding(w, r)
	case r.URL.Path == "/api/v1/tasks" && r.Method == http.MethodGet:
		s.handleAPITasks(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/tasks/") && r.Method == http.MethodPost:
		s.handleAPITaskAction(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/cards/"):
		s.handleAPICardRoute(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/avatars/") && r.Method == http.MethodGet:
		s.handleAPIAvatar(w, r)
	case r.URL.Path == "/api/v1/feed" && r.Method == http.MethodGet:
		s.handleAPIFeed(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/feed/") && r.Method == http.MethodPatch:
		s.handleAPIFeedRead(w, r)
	case r.URL.Path == "/api/v1/admin/subscribers" && r.Method == http.MethodGet:
		s.handleAPIAdminSubscribers(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/admin/subscribers/") && r.Method == http.MethodPatch:
		s.handleAPIAdminSubscriberAction(w, r)
	default:
		writeAPIError(w, http.StatusNotFound, "not_found")
	}
}
