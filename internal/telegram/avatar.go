package telegram

import (
	"context"
	"html"
	"strings"
)

// handleSetAvatar — адмін надсилає фото/документ-зображення з підписом
// "/setavatar <planka_user>". Бот заливає файл у Planka API.
func (c *Client) handleSetAvatar(ctx context.Context, m *tgMessage) {
	chatID := m.Chat.ID
	if !c.store.IsAdmin(ctx, m.From.Username) {
		c.send(chatID, "⛔ Лише для адмінів.")
		return
	}
	if !c.planka.Configured() {
		c.send(chatID, "❗ Workspace API не налаштоване (PLANKA_API_EMAIL/PASSWORD).")
		return
	}
	parts := strings.Fields(strings.TrimSpace(m.Caption))
	if len(parts) != 2 {
		c.send(chatID, "Використання: надішліть фото з підписом <code>/setavatar &lt;workspace_user&gt;</code>")
		return
	}
	username := parts[1]

	userID, err := c.store.GetPlankaUserID(ctx, username)
	if err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}

	fileID, filename := pickAvatarFile(m)

	data, err := c.downloadFile(ctx, fileID)
	if err != nil {
		c.send(chatID, "Не вдалося завантажити файл: "+html.EscapeString(err.Error()))
		return
	}
	if err := c.planka.UploadAvatar(ctx, userID, filename, data); err != nil {
		c.send(chatID, "Workspace API помилка: "+html.EscapeString(err.Error()))
		return
	}
	c.send(chatID, "🖼 Аватар оновлено для <b>"+html.EscapeString(username)+"</b>.")
}

// pickAvatarFile — обирає file_id (найбільший photo size або document) і ім'я файлу.
func pickAvatarFile(m *tgMessage) (fileID, filename string) {
	if len(m.Photo) > 0 {
		// Найбільший варіант — останній в масиві.
		return m.Photo[len(m.Photo)-1].FileID, "avatar.jpg"
	}
	filename = m.Document.FileName
	if filename == "" {
		filename = "avatar.png"
	}
	return m.Document.FileID, filename
}

// handleAvatar — /avatar [planka_user].
// Без аргументів — зливаємо власне Telegram-фото в Planka. З аргументом (адмін) —
// зливаємо фото профілю вказаного користувача.
func (c *Client) handleAvatar(ctx context.Context, fromUsername string, fromUserID, chatID int64, args []string, isAdmin bool) {
	if !c.planka.Configured() {
		c.send(chatID, "❗ Workspace API не налаштоване (PLANKA_API_EMAIL/PASSWORD).")
		return
	}

	plankaUsername, tgUserID, ok := c.resolveAvatarTarget(ctx, fromUsername, fromUserID, chatID, args, isAdmin)
	if !ok {
		return
	}

	userID, err := c.store.GetPlankaUserID(ctx, plankaUsername)
	if err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	fileID, err := c.getProfilePhotoFileID(ctx, tgUserID)
	if err != nil {
		c.send(chatID, "Не вдалося отримати фото профілю: "+html.EscapeString(err.Error()))
		return
	}
	data, err := c.downloadFile(ctx, fileID)
	if err != nil {
		c.send(chatID, "Не вдалося завантажити файл: "+html.EscapeString(err.Error()))
		return
	}
	if err := c.planka.UploadAvatar(ctx, userID, "avatar.jpg", data); err != nil {
		c.send(chatID, "Workspace API помилка: "+html.EscapeString(err.Error()))
		return
	}
	c.send(chatID, "🖼 Аватар оновлено для <b>"+html.EscapeString(plankaUsername)+"</b>.")
}

// resolveAvatarTarget — визначає (plankaUsername, telegramUserID) для /avatar за аргументами.
func (c *Client) resolveAvatarTarget(ctx context.Context, fromUsername string, fromUserID, chatID int64, args []string, isAdmin bool) (plankaUsername string, tgUserID int64, ok bool) {
	switch len(args) {
	case 0:
		// Власне фото → власний planka-акаунт.
		if fromUsername == "" {
			c.send(chatID, "❗ Спочатку /start.")
			return "", 0, false
		}
		b, _ := c.store.GetByTelegram(ctx, fromUsername)
		if b == nil {
			c.send(chatID, "❗ Спочатку /start.")
			return "", 0, false
		}
		plankaUsername = b.PlankaUsername
		if plankaUsername == "" {
			plankaUsername = b.TelegramUsername
		}
		return plankaUsername, fromUserID, true

	case 1:
		if !isAdmin {
			c.send(chatID, "⛔ Лише для адмінів. Без аргументу — фото для себе.")
			return "", 0, false
		}
		plankaUsername = args[0]
		id, found := c.store.GetTelegramIDForPlanka(ctx, plankaUsername)
		if !found {
			c.send(chatID, "Користувач не привʼязаний до Telegram. Нехай зробить /start (або /link).")
			return "", 0, false
		}
		return plankaUsername, id, true

	default:
		c.send(chatID, "Використання: <code>/avatar</code> (для себе) або <code>/avatar &lt;workspace_user&gt;</code> (адмін).")
		return "", 0, false
	}
}
