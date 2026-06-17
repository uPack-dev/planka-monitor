package events

import "sort"

// DefaultTemplates — карта усіх відомих шаблонів (для /templates).
//
// Компактні «заголовки» подій. Деталі (коментар, diff, old/new назви тощо)
// живуть у Render.Details і відображаються лише після натискання кнопки «Деталі».
//
// kind: "channel" | "personal".
var DefaultTemplates = map[string]map[string]string{
	"cardCreate": {
		"channel": "🆕 {{.Actor}} створив(-ла) картку {{.CardLink}}{{.Context}}",
	},
	"cardDelete": {
		"channel": "🗑 {{.Actor}} видалив(-ла) картку {{.CardLink}}{{.Context}}",
	},
	"cardRename": {
		"channel": "✏️ {{.Actor}} перейменував(-ла) картку {{.CardLink}}{{.Context}}",
	},
	"cardEdit": {
		"channel":  "✏️ {{.Actor}} оновив(-ла) картку {{.CardLink}}{{.Context}}",
		"personal": "✏️ {{.Actor}} оновив(-ла) вашу картку {{.CardLink}}{{.Context}}",
	},
	"cardMove": {
		"channel": "➡️ {{.Actor}} перемістив(-ла) картку {{.CardLink}} → <b>{{.ListName}}</b>{{.Context}}",
	},
	"cardMembershipCreate": {
		"channel":  "👤 {{.Actor}} призначив(-ла) <b>{{.TargetName}}</b> на {{.CardLink}}{{.Context}}",
		"personal": "👤 {{.Actor}} призначив(-ла) вас на {{.CardLink}}{{.Context}}",
	},
	"cardMembershipDelete": {
		"channel":  "👤❌ {{.Actor}} зняв(-ла) <b>{{.TargetName}}</b> з {{.CardLink}}{{.Context}}",
		"personal": "👤❌ {{.Actor}} зняв(-ла) вас з {{.CardLink}}{{.Context}}",
	},
	"commentCreate": {
		"channel":  "💬 {{.Actor}} прокоментував(-ла) {{.CardLink}}{{.Context}}",
		"personal": "💬 {{.Actor}} прокоментував(-ла) вашу {{.CardLink}}{{.Context}}",
	},
	"cardLabelCreate": {
		"channel": "🏷 {{.Actor}} додав(-ла) мітку <b>{{.LabelName}}</b> до {{.CardLink}}{{.Context}}",
	},
	"cardLabelDelete": {
		"channel": "🏷❌ {{.Actor}} зняв(-ла) мітку <b>{{.LabelName}}</b> з {{.CardLink}}{{.Context}}",
	},
	"taskAssign": {
		"channel":  "👤 {{.Actor}} призначив(-ла) <b>{{.TargetName}}</b> на <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
		"personal": "👤 {{.Actor}} призначив(-ла) вас на <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
	},
	"taskUnassign": {
		"channel":  "👤❌ {{.Actor}} зняв(-ла) <b>{{.TargetName}}</b> з <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
		"personal": "👤❌ {{.Actor}} зняв(-ла) вас з <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
	},
	"taskCreate": {
		"channel": "➕ {{.Actor}} додав(-ла) задачу <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
	},
	"taskDelete": {
		"channel": "➖ {{.Actor}} видалив(-ла) задачу <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
	},
	"taskComplete": {
		"channel":  "✅ {{.Actor}} виконав(-ла) <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
		"personal": "✅ {{.Actor}} виконав(-ла) вашу <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
	},
	"taskUncomplete": {
		"channel":  "↩️ {{.Actor}} повернув(-ла) <i>{{.TaskName}}</i> у роботу у {{.CardLink}}{{.Context}}",
		"personal": "↩️ {{.Actor}} повернув(-ла) вашу <i>{{.TaskName}}</i> у роботу у {{.CardLink}}{{.Context}}",
	},
	"taskRename": {
		"channel":  "✏️ {{.Actor}} перейменував(-ла) задачу у {{.CardLink}}{{.Context}}",
		"personal": "✏️ {{.Actor}} перейменував(-ла) вашу задачу у {{.CardLink}}{{.Context}}",
	},
	"taskEdit": {
		"channel":  "✏️ {{.Actor}} оновив(-ла) <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
		"personal": "✏️ {{.Actor}} оновив(-ла) вашу <i>{{.TaskName}}</i> у {{.CardLink}}{{.Context}}",
	},
	"attachmentCreate": {
		"channel": "📎 {{.Actor}} додав(-ла) <i>{{.AttachName}}</i> до {{.CardLink}}{{.Context}}",
	},
	"listCreate": {
		"channel": "📋 {{.Actor}} створив(-ла) колонку <b>{{.ListName}}</b>{{.Context}}",
	},
	"listDelete": {
		"channel": "🗑 {{.Actor}} видалив(-ла) колонку <b>{{.ListName}}</b>{{.Context}}",
	},
	"boardCreate": {
		"channel": "🗂 {{.Actor}} створив(-ла) дошку {{.BoardLink}}",
	},
	"boardDelete": {
		"channel": "🗑 {{.Actor}} видалив(-ла) дошку {{.BoardLink}}",
	},
	"projectCreate": {
		"channel": "📁 {{.Actor}} створив(-ла) проєкт <b>{{.ProjectName}}</b>",
	},
}

func KnownEvents() []string {
	keys := make([]string, 0, len(DefaultTemplates))
	for k := range DefaultTemplates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func KnownKinds(event string) []string {
	m, ok := DefaultTemplates[event]
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
