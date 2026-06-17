package telegram

import (
	"context"
	"fmt"
)

func (c *Client) SendChannel(text string) { c.send(c.chatID, text) }

func (c *Client) SendToPlankaUser(ctx context.Context, plankaUsername, text string) bool {
	id, ok := c.store.GetChatIDForPlanka(ctx, plankaUsername)
	if !ok {
		return false
	}
	c.send(id, text)
	return true
}

// SendNotificationChannel — компактний хедер у канал з кнопками навігації + «Деталі».
func (c *Client) SendNotificationChannel(n Notification) {
	c.sendNotification(c.chatID, n)
}

// SendNotificationToPlankaUser — те саме, але в DM. Повертає false, якщо привʼязки немає.
func (c *Client) SendNotificationToPlankaUser(ctx context.Context, plankaUsername string, n Notification) bool {
	id, ok := c.store.GetChatIDForPlanka(ctx, plankaUsername)
	if !ok {
		return false
	}
	c.sendNotification(id, n)
	return true
}

func (c *Client) sendNotification(chatID any, n Notification) {
	markup := buildNotifKeyboard(n, false)
	msgID := c.sendCapture(chatID, n.Heading, markup)
	if msgID == 0 {
		return
	}
	idInt, ok := chatIDToInt64(chatID)
	if !ok {
		return
	}
	c.notifCache.put(context.Background(), notifKey(idInt, msgID), n)
}

// chatIDToInt64 — нормалізує chat_id (може бути int64 або string у головному каналі) у int64.
func chatIDToInt64(chatID any) (int64, bool) {
	switch v := chatID.(type) {
	case int64:
		return v, true
	case string:
		var id int64
		if _, err := fmt.Sscanf(v, "%d", &id); err == nil {
			return id, true
		}
	}
	return 0, false
}

// buildNotifKeyboard — будує клавіатуру з URL-кнопок (Картка/Дошка/Проєкт)
// та опціональної callback-кнопки розгортання деталей.
func buildNotifKeyboard(n Notification, expanded bool) *inlineKeyboard {
	var nav []inlineButton
	if n.CardURL != "" {
		nav = append(nav, inlineButton{Text: "🃏 Картка", URL: n.CardURL})
	}
	if n.BoardURL != "" {
		nav = append(nav, inlineButton{Text: "🗂 Дошка", URL: n.BoardURL})
	}
	if n.ProjectURL != "" {
		nav = append(nav, inlineButton{Text: "📁 Проєкт", URL: n.ProjectURL})
	}

	rows := [][]inlineButton{}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}
	if n.Details != "" {
		btn := inlineButton{Text: "👁 Деталі", CallbackData: "exp"}
		if expanded {
			btn = inlineButton{Text: "🙈 Сховати", CallbackData: "col"}
		}
		rows = append(rows, []inlineButton{btn})
	}
	if len(rows) == 0 {
		return nil
	}
	return &inlineKeyboard{InlineKeyboard: rows}
}
