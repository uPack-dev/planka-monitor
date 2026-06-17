package store

import "time"

type MiniUser struct {
	ID                   string `json:"id"`
	Username             string `json:"username"`
	Name                 string `json:"name"`
	Initials             string `json:"initials"`
	AvatarURL            string `json:"avatarUrl,omitempty"`
	AvatarUploadedFileID string `json:"-"`
	AvatarExtension      string `json:"-"`
}

type TimelineItem struct {
	ID               string     `json:"id"`
	Kind             string     `json:"kind"`
	CardID           string     `json:"cardId"`
	TaskID           string     `json:"taskId,omitempty"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	DueAt            *time.Time `json:"dueAt,omitempty"`
	StartAt          *time.Time `json:"startAt,omitempty"`
	EndAt            *time.Time `json:"endAt,omitempty"`
	BoardID          string     `json:"boardId"`
	BoardName        string     `json:"boardName"`
	ProjectID        string     `json:"projectId"`
	ProjectName      string     `json:"projectName"`
	ListName         string     `json:"listName"`
	IsCompleted      bool       `json:"isCompleted"`
	IsMuted          bool       `json:"isMuted"`
	Members          []MiniUser `json:"members"`
	WorkspaceURL     string     `json:"workspaceUrl,omitempty"`
	BoardURL         string     `json:"boardUrl,omitempty"`
	ProjectURL       string     `json:"projectUrl,omitempty"`
	CardWorkspaceURL string     `json:"cardWorkspaceUrl,omitempty"`
}

type TimelineStats struct {
	Completed int `json:"completed"`
}

type CardDetail struct {
	ID             string        `json:"id"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	DueAt          *time.Time    `json:"dueAt,omitempty"`
	StartAt        *time.Time    `json:"startAt,omitempty"`
	EndAt          *time.Time    `json:"endAt,omitempty"`
	IsDueCompleted bool          `json:"isDueCompleted"`
	BoardID        string        `json:"boardId"`
	BoardName      string        `json:"boardName"`
	ProjectID      string        `json:"projectId"`
	ProjectName    string        `json:"projectName"`
	ListName       string        `json:"listName"`
	IsMuted        bool          `json:"isMuted"`
	Members        []MiniUser    `json:"members"`
	Tasks          []CardTask    `json:"tasks"`
	Comments       []CardComment `json:"comments"`
	WorkspaceURL   string        `json:"workspaceUrl,omitempty"`
	BoardURL       string        `json:"boardUrl,omitempty"`
	ProjectURL     string        `json:"projectUrl,omitempty"`
	CardURL        string        `json:"cardWorkspaceUrl,omitempty"`
}

type CardCompletionTarget struct {
	UseDoneLabel      bool
	DoneLabelID       string
	DoneLabelAttached bool
}

type CardTask struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IsCompleted bool      `json:"isCompleted"`
	Assignee    *MiniUser `json:"assignee,omitempty"`
}

type CardComment struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	Author    MiniUser  `json:"author"`
}

type FeedEvent struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`
	CreatedAt        time.Time `json:"createdAt"`
	IsRead           bool      `json:"isRead"`
	IsMuted          bool      `json:"isMuted"`
	CardID           string    `json:"cardId"`
	CardTitle        string    `json:"cardTitle"`
	BoardID          string    `json:"boardId"`
	BoardName        string    `json:"boardName"`
	ProjectID        string    `json:"projectId"`
	ProjectName      string    `json:"projectName"`
	Text             string    `json:"text,omitempty"`
	Actor            MiniUser  `json:"actor"`
	WorkspaceURL     string    `json:"workspaceUrl,omitempty"`
	BoardURL         string    `json:"boardUrl,omitempty"`
	ProjectURL       string    `json:"projectUrl,omitempty"`
	CardWorkspaceURL string    `json:"cardWorkspaceUrl,omitempty"`
}

type AdminSubscriber struct {
	TelegramUsername     string                  `json:"telegramUsername"`
	TelegramUserID       int64                   `json:"telegramUserId"`
	WorkspaceUsername    string                  `json:"workspaceUsername"`
	WorkspaceExists      bool                    `json:"workspaceExists"`
	DisplayName          string                  `json:"displayName"`
	AvatarURL            string                  `json:"avatarUrl,omitempty"`
	AvatarUploadedFileID string                  `json:"-"`
	AvatarExtension      string                  `json:"-"`
	Role                 string                  `json:"role"`
	IsAdmin              bool                    `json:"isAdmin"`
	IsBlocked            bool                    `json:"isBlocked"`
	NotificationsEnabled bool                    `json:"notificationsEnabled"`
	Preferences          NotificationPreferences `json:"preferences"`
	UpdatedAt            time.Time               `json:"updatedAt"`
}
