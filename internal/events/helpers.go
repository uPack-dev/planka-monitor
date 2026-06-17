package events

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

func cardURL(baseURL, id string) string  { return fmt.Sprintf("%s/cards/%s", baseURL, id) }
func boardURL(baseURL, id string) string { return fmt.Sprintf("%s/boards/%s", baseURL, id) }

func cardLink(baseURL string, c Card) string {
	if c.ID == "" {
		return ""
	}
	name := c.Name
	if name == "" {
		name = "картка"
	}
	return fmt.Sprintf(`<a href="%s">%s</a>`, cardURL(baseURL, c.ID), html.EscapeString(name))
}

func boardLink(baseURL string, b Board) string {
	if b.ID == "" {
		return ""
	}
	name := b.Name
	if name == "" {
		name = "дошка"
	}
	return fmt.Sprintf(`<a href="%s">%s</a>`, boardURL(baseURL, b.ID), html.EscapeString(name))
}

func mergeCard(item, fallback Card) Card {
	if item.ID == "" {
		item.ID = fallback.ID
	}
	if item.Name == "" {
		item.Name = fallback.Name
	}
	return item
}

func firstBoard(p *Payload) Board {
	if len(p.Data.Included.Boards) > 0 {
		return p.Data.Included.Boards[0]
	}
	return Board{}
}

func firstList(p *Payload) List {
	if len(p.Data.Included.Lists) > 0 {
		return p.Data.Included.Lists[0]
	}
	return List{}
}

func firstCard(p *Payload) Card {
	if len(p.Data.Included.Cards) > 0 {
		return p.Data.Included.Cards[0]
	}
	return Card{}
}

func userDisplay(u *User) (name, username string) {
	name = "користувача"
	if u == nil {
		return
	}
	username = u.Username
	if u.Name != "" {
		name = u.Name
	} else if u.Username != "" {
		name = u.Username
	}
	return
}

func userByID(p *Payload, id string) *User {
	for i := range p.Data.Included.Users {
		if p.Data.Included.Users[i].ID == id {
			return &p.Data.Included.Users[i]
		}
	}
	return nil
}

func contextSuffix(baseURL string, b Board, l List) string {
	parts := []string{}
	if b.ID != "" {
		parts = append(parts, "📋 "+boardLink(baseURL, b))
	}
	if l.Name != "" {
		parts = append(parts, "📑 "+html.EscapeString(l.Name))
	}
	if len(parts) == 0 {
		return ""
	}
	return "\n" + strings.Join(parts, " · ")
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// strFromItem — витягує строкове значення поля з сирого JSON item-а.
func strFromItem(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// prevItem — у цьому форку Planka PrevData має вигляд {"item": {...}, "included": {...}},
// тобто це повне попереднє представлення payload, а не плоский diff. Дістаємо саме item.
// Fallback на «плоский» формат залишено для сумісності.
func prevItem(p *Payload) map[string]any {
	if len(p.PrevData) == 0 {
		return nil
	}
	var wrapper struct {
		Item json.RawMessage `json:"item"`
	}
	if err := json.Unmarshal(p.PrevData, &wrapper); err != nil {
		return nil
	}
	if len(wrapper.Item) == 0 {
		var m map[string]any
		_ = json.Unmarshal(p.PrevData, &m)
		return m
	}
	var m map[string]any
	_ = json.Unmarshal(wrapper.Item, &m)
	return m
}

// changed повертає (prevValue, true) якщо значення в prev відрізняється від new.
// Нормалізує nil ↔ "" для рядків.
func changed(prev map[string]any, key string, newVal any) (any, bool) {
	if prev == nil {
		return nil, false
	}
	pv, ok := prev[key]
	if !ok {
		return nil, false
	}
	if pv == nil && newVal == "" {
		return nil, false
	}
	if ns, isStr := newVal.(string); isStr {
		if ps, ok := pv.(string); ok && ps == ns {
			return nil, false
		}
	} else if pv == newVal {
		return nil, false
	}
	return pv, true
}

// escapeHTMLBlockquote — ескейпить текст і обгортає в <blockquote>, обрізаючи до limit.
func escapeHTMLBlockquote(s string, limit int) string {
	return "<blockquote>" + html.EscapeString(truncate(s, limit)) + "</blockquote>"
}
