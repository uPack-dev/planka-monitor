package store

import "strings"

func miniUser(id, username, name string) MiniUser {
	return miniUserWithAvatar(id, username, name, "", "")
}

func miniUserWithAvatar(id, username, name, avatarFileID, avatarExt string) MiniUser {
	display := name
	if display == "" {
		display = username
	}
	return MiniUser{
		ID:                   id,
		Username:             username,
		Name:                 name,
		Initials:             initials(display),
		AvatarUploadedFileID: avatarFileID,
		AvatarExtension:      avatarExt,
	}
}

func initials(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "?"
	}
	parts := strings.Fields(s)
	first := strings.ToUpper(string([]rune(parts[0])[0]))
	if len(parts) == 1 {
		return first
	}
	return first + strings.ToUpper(string([]rune(parts[1])[0]))
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
