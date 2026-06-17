package telegram

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// RunPolling — long-poll цикл getUpdates. Завершується по ctx.Done().
func (c *Client) RunPolling(ctx context.Context) {
	// Дефолтне меню — для всіх приватних чатів.
	c.setMyCommands(ctx, userCommands, setCmdScope{Type: "all_private_chats"})
	c.ensureDefaultMenuButton(ctx)

	const allowedUpdates = `["message","callback_query"]`
	var offset int64
	for {
		if ctx.Err() != nil {
			return
		}
		query := fmt.Sprintf("timeout=%d&offset=%d&allowed_updates=%s",
			pollTimeout, offset, url.QueryEscape(allowedUpdates))

		var ur updatesResp
		if err := c.callGET(ctx, "getUpdates", query, &ur); err != nil {
			if ctx.Err() != nil {
				return
			}
			// Простий backoff — не флудимо API при тимчасових проблемах.
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		for _, up := range ur.Result {
			offset = up.UpdateID + 1
			switch {
			case up.CallbackQuery != nil:
				c.handleCallback(ctx, up.CallbackQuery)
			case up.Message != nil:
				c.handleMessage(ctx, up.Message)
			}
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, m *tgMessage) {
	// Бот працює лише у приватних чатах. У групах/каналах ігноруємо все,
	// щоб уникнути зливу особистих сповіщень або реєстрації group chat_id.
	if m.Chat.Type != "private" {
		text := strings.TrimSpace(m.Text)
		if strings.HasPrefix(text, "/start") || strings.HasPrefix(text, "/start@") {
			c.send(m.Chat.ID, "Будь ласка, напишіть мені у приват — я працюю лише в особистих повідомленнях.")
		}
		return
	}
	if c.store.IsBlocked(ctx, m.From.Username) {
		return
	}

	caption := strings.TrimSpace(m.Caption)
	hasImage := len(m.Photo) > 0 || (m.Document != nil && strings.HasPrefix(m.Document.MimeType, "image/"))
	if hasImage && strings.HasPrefix(caption, "/setavatar") {
		c.handleSetAvatar(ctx, m)
		return
	}
	c.handleCommand(ctx, m.From.Username, m.From.ID, m.Chat.ID, m.Text)
}

func (c *Client) handleCallback(ctx context.Context, cb *callbackQuery) {
	if cb.Message == nil {
		c.answerCallback(cb.ID, "", false)
		return
	}
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID
	fromUsername := cb.From.Username

	// exp/col доступні всюди (зокрема у каналі) — це лише toggle деталей.
	if cb.Data == "exp" || cb.Data == "col" {
		c.handleToggleDetails(ctx, cb, chatID, msgID, fromUsername)
		return
	}

	// Решта callback-ів — лише в DM.
	if cb.Message.Chat.Type != "private" {
		c.answerCallback(cb.ID, "Бот працює лише у приваті.", true)
		return
	}
	if c.store.IsBlocked(ctx, fromUsername) {
		c.answerCallback(cb.ID, "", false)
		return
	}

	switch cb.Data {
	case "sub:on", "sub:off":
		c.handleSubscriptionToggle(ctx, cb, chatID, msgID, fromUsername)
	case "help":
		c.answerCallback(cb.ID, "", false)
		isAdmin := c.store.IsAdmin(ctx, fromUsername)
		c.ensureChatMenu(ctx, chatID, isAdmin)
		c.replyWithMain(ctx, fromUsername, chatID, helpText(isAdmin))
	case "whoami":
		c.answerCallback(cb.ID, "", false)
		isAdmin := c.store.IsAdmin(ctx, fromUsername)
		c.ensureChatMenu(ctx, chatID, isAdmin)
		c.cmdWhoami(ctx, fromUsername, chatID, isAdmin)
	default:
		c.answerCallback(cb.ID, "", false)
	}
}

func (c *Client) handleToggleDetails(ctx context.Context, cb *callbackQuery, chatID, msgID int64, fromUsername string) {
	if c.store.IsBlocked(ctx, fromUsername) {
		c.answerCallback(cb.ID, "", false)
		return
	}
	key := notifKey(chatID, msgID)
	n, ok := c.notifCache.get(ctx, key)
	if !ok {
		c.answerCallback(cb.ID, "Деталі недоступні (можливо, бот був перезапущений).", true)
		return
	}
	c.answerCallback(cb.ID, "", false)
	expanded := cb.Data == "exp"
	text := n.Heading
	if expanded {
		text = n.Heading + "\n\n" + n.Details
	}
	c.editMessage(chatID, msgID, text, buildNotifKeyboard(n, expanded))
}

func (c *Client) handleSubscriptionToggle(ctx context.Context, cb *callbackQuery, chatID, msgID int64, fromUsername string) {
	enable := cb.Data == "sub:on"
	if fromUsername == "" {
		c.answerCallback(cb.ID, "Спочатку встановіть Telegram username", true)
		return
	}
	if err := c.store.SetUser(ctx, fromUsername, chatID, cb.From.ID); err != nil {
		c.answerCallback(cb.ID, "Помилка: "+err.Error(), true)
		return
	}
	if err := c.store.SetNotifications(ctx, fromUsername, enable); err != nil {
		c.answerCallback(cb.ID, "Помилка: "+err.Error(), true)
		return
	}
	notice := "🔕 Сповіщення вимкнено"
	if enable {
		notice = "🔔 Сповіщення увімкнено"
	}
	c.answerCallback(cb.ID, notice, false)
	c.editMessage(chatID, msgID, startText(fromUsername, enable), startKeyboard(enable, c.webAppURL))
}
