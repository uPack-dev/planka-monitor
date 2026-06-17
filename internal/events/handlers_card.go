package events

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

func produceCardCreateDelete(c *prodCtx, event string) []Render {
	var item Card
	_ = json.Unmarshal(c.p.Data.Item, &item)
	merged := mergeCard(item, c.card)
	c.base["CardLink"] = cardLink(c.baseURL, merged)
	r := c.mk(event, nil, "")
	r.CardID = merged.ID
	return []Render{r}
}

func produceCardUpdate(c *prodCtx) []Render {
	var item Card
	_ = json.Unmarshal(c.p.Data.Item, &item)
	prev := prevItem(c.p)
	merged := mergeCard(item, c.card)
	c.base["CardLink"] = cardLink(c.baseURL, merged)

	// move
	if pv, ok := changed(prev, "listId", item.ListID); ok {
		if s, _ := pv.(string); s != "" {
			r := c.mk("cardMove", nil, "")
			r.CardID = merged.ID
			return []Render{r}
		}
	}
	// rename
	if pv, ok := changed(prev, "name", item.Name); ok {
		if s, _ := pv.(string); s != item.Name {
			r := c.mk("cardRename", nil, "")
			r.CardID = merged.ID
			r.Details = "було: <i>" + html.EscapeString(s) + "</i>\nстало: <i>" + html.EscapeString(item.Name) + "</i>"
			return []Render{r}
		}
	}
	// generic edit (description / dueDate)
	var diff []string
	newDesc := strFromItem(c.p.Data.Item, "description")
	descChanged := false
	if _, ok := changed(prev, "description", newDesc); ok {
		diff = append(diff, "• опис змінено")
		descChanged = true
	}
	if pv, ok := changed(prev, "dueDate", strFromItem(c.p.Data.Item, "dueDate")); ok {
		diff = append(diff, fmt.Sprintf("• термін: %v", pv))
	}
	if len(diff) == 0 {
		return nil
	}
	r := c.mk("cardEdit", nil, "")
	r.CardID = merged.ID
	r.DMCardID = merged.ID
	details := strings.Join(diff, "\n")
	if descChanged && newDesc != "" {
		details += "\n\n<b>Новий опис:</b>\n" + escapeHTMLBlockquote(newDesc, 1500)
	}
	r.Details = details
	return []Render{r}
}

func produceCardMembership(c *prodCtx, event string) []Render {
	var item struct {
		CardID string `json:"cardId"`
		UserID string `json:"userId"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)

	target := userByID(c.p, item.UserID)
	targetName := "користувача"
	targetUsername := ""
	if target != nil {
		switch {
		case target.Name != "":
			targetName = target.Name
		case target.Username != "":
			targetName = target.Username
		}
		targetUsername = target.Username
	}
	return []Render{c.mk(event, map[string]string{
		"TargetName": html.EscapeString(targetName),
	}, targetUsername)}
}

func produceCardLabel(c *prodCtx, event string) []Render {
	var item struct {
		CardID  string `json:"cardId"`
		LabelID string `json:"labelId"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)

	labelName, labelColor := "", ""
	for _, l := range c.p.Data.Included.Labels {
		if l.ID == item.LabelID {
			labelName = l.Name
			labelColor = l.Color
			break
		}
	}
	if labelName == "" {
		labelName = labelColor
	}
	// LabelID лишаємо для пост-резолву через DB (server.dispatch),
	// якщо included порожній (типово для delete-подій).
	return []Render{c.mk(event, map[string]string{
		"LabelName": html.EscapeString(labelName),
		"LabelID":   item.LabelID,
	}, "")}
}
