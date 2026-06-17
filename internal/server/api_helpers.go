package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"planka-monitor/internal/store"
)

func normalizeAPIUsername(username string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
}

func (s *Server) plankaUserExists(ctx context.Context, username string) bool {
	exists, err := s.store.PlankaUserExists(ctx, username)
	return err == nil && exists
}

func (s *Server) writeAPIStoreError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "не зареєстрований") {
		writeAPIError(w, http.StatusBadRequest, "needs_start")
		return
	}
	writeAPIError(w, http.StatusInternalServerError, "internal_error")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(out); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_json")
		return false
	}
	return true
}

func parseRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	from := now.AddDate(0, 0, -14)
	to := now.AddDate(0, 0, 60)
	if raw := r.URL.Query().Get("from"); raw != "" {
		if parsed, err := parseAPITime(raw); err == nil {
			from = parsed
		}
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		if parsed, err := parseAPITime(raw); err == nil {
			to = parsed
		}
	}
	if to.Before(from) {
		to = from.AddDate(0, 0, 60)
	}
	return from, to
}

func parseStatsRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 0, 1).Add(-time.Millisecond)
	if raw := r.URL.Query().Get("statsFrom"); raw != "" {
		if parsed, err := parseAPITime(raw); err == nil {
			from = parsed
		}
	}
	if raw := r.URL.Query().Get("statsTo"); raw != "" {
		if parsed, err := parseAPITime(raw); err == nil {
			to = parsed
		}
	}
	if to.Before(from) {
		to = from.AddDate(0, 0, 1).Add(-time.Millisecond)
	}
	return from, to
}

func parseAPITime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	if d, err := time.Parse("2006-01-02", raw); err == nil {
		return d, nil
	}
	return time.Time{}, errors.New("bad time")
}

func parsePositiveInt(raw string, fallback int) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func splitActionPath(path, prefix string) (id, action string, ok bool) {
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if rest == "" || rest == path {
		return "", "", false
	}
	parts := strings.Split(rest, "/")
	id = parts[0]
	if id == "" {
		return "", "", false
	}
	if len(parts) > 1 {
		action = parts[1]
	}
	return id, action, true
}

func parseAvatarPath(path string) (fileID, variant, ext string, ok bool) {
	rest := strings.Trim(strings.TrimPrefix(path, "/api/v1/avatars/"), "/")
	if rest == "" || rest == path {
		return "", "", "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		return "", "", "", false
	}
	decodedFileID, err := url.PathUnescape(parts[0])
	if err != nil || decodedFileID == "" {
		return "", "", "", false
	}
	name := parts[1]
	dot := strings.LastIndex(name, ".")
	if dot <= 0 || dot == len(name)-1 {
		return "", "", "", false
	}
	decodedVariant, err := url.PathUnescape(name[:dot])
	if err != nil || decodedVariant != "cover-180" {
		return "", "", "", false
	}
	decodedExt, err := url.PathUnescape(name[dot+1:])
	if err != nil || decodedExt == "" || strings.Contains(decodedExt, "/") {
		return "", "", "", false
	}
	return decodedFileID, decodedVariant, strings.ToLower(decodedExt), true
}

func avatarContentType(ext string) string {
	switch strings.ToLower(strings.TrimPrefix(ext, ".")) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func (s *Server) workspaceAvatarURL(ctx context.Context, username string) string {
	fileID, ext, ok := s.store.GetPlankaAvatar(ctx, username)
	if !ok {
		return ""
	}
	return avatarProxyURL(fileID, ext)
}

func (s *Server) workspaceURL() string {
	return strings.TrimRight(s.baseURL, "/")
}

func avatarProxyURL(fileID, ext string) string {
	fileID = strings.TrimSpace(fileID)
	ext = strings.Trim(strings.TrimSpace(ext), ".")
	if fileID == "" || ext == "" {
		return ""
	}
	return "/api/v1/avatars/" + url.PathEscape(fileID) + "/cover-180." + url.PathEscape(ext)
}

func (s *Server) attachWorkspaceURLs(items []store.TimelineItem) {
	base := strings.TrimRight(s.baseURL, "/")
	for i := range items {
		if base != "" {
			items[i].WorkspaceURL = base
			items[i].CardWorkspaceURL = base + "/cards/" + items[i].CardID
			items[i].BoardURL = base + "/boards/" + items[i].BoardID
			items[i].ProjectURL = base + "/projects/" + items[i].ProjectID
		}
		attachMiniUserAvatarURLs(items[i].Members)
	}
}

func (s *Server) attachCardWorkspaceURLs(card *store.CardDetail) {
	base := strings.TrimRight(s.baseURL, "/")
	if card == nil {
		return
	}
	if base != "" {
		card.WorkspaceURL = base
		card.CardURL = base + "/cards/" + card.ID
		card.BoardURL = base + "/boards/" + card.BoardID
		card.ProjectURL = base + "/projects/" + card.ProjectID
	}
	attachMiniUserAvatarURLs(card.Members)
	for i := range card.Tasks {
		if card.Tasks[i].Assignee != nil {
			attachMiniUserAvatarURL(card.Tasks[i].Assignee)
		}
	}
	for i := range card.Comments {
		attachMiniUserAvatarURL(&card.Comments[i].Author)
	}
}

func (s *Server) attachFeedWorkspaceURLs(items []store.FeedEvent) {
	base := strings.TrimRight(s.baseURL, "/")
	for i := range items {
		if base != "" {
			items[i].WorkspaceURL = base
			if items[i].CardID != "" {
				items[i].CardWorkspaceURL = base + "/cards/" + items[i].CardID
			}
			if items[i].BoardID != "" {
				items[i].BoardURL = base + "/boards/" + items[i].BoardID
			}
			if items[i].ProjectID != "" {
				items[i].ProjectURL = base + "/projects/" + items[i].ProjectID
			}
		}
		attachMiniUserAvatarURL(&items[i].Actor)
	}
}

func attachAdminSubscriberAvatarURLs(items []store.AdminSubscriber) {
	for i := range items {
		items[i].AvatarURL = avatarProxyURL(items[i].AvatarUploadedFileID, items[i].AvatarExtension)
	}
}

func attachMiniUserAvatarURLs(users []store.MiniUser) {
	for i := range users {
		attachMiniUserAvatarURL(&users[i])
	}
}

func attachMiniUserAvatarURL(user *store.MiniUser) {
	if user == nil {
		return
	}
	user.AvatarURL = avatarProxyURL(user.AvatarUploadedFileID, user.AvatarExtension)
}

func writeAPIError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
