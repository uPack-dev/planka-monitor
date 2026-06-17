package telegram

import (
	"context"
	"fmt"
	"html"
	"strings"

	"planka-monitor/internal/events"
)

// handleCommand — точка входу для слеш-команд у приваті.
// Парсить команду й аргументи, перевіряє права, оновлює бот-меню,
// потім диспатчить на конкретний обробник.
func (c *Client) handleCommand(ctx context.Context, fromUsername string, fromUserID, chatID int64, text string) {
	cmd, args, rest, ok := parseCommand(text)
	if !ok {
		return
	}

	isAdmin := c.store.IsAdmin(ctx, fromUsername)
	c.ensureChatMenu(ctx, chatID, isAdmin)

	// Тиха ігнорація адмін-команд для звичайних користувачів —
	// щоб не розкривати адмін-функціонал.
	if !isAdmin {
		if _, isAdminCmd := adminCommandSet[cmd]; isAdminCmd {
			return
		}
	}

	switch cmd {
	// ---- user ----
	case "/start":
		c.cmdStart(ctx, fromUsername, fromUserID, chatID)
	case "/notify":
		c.cmdNotify(ctx, fromUsername, fromUserID, chatID)
	case "/stop":
		c.cmdStop(ctx, fromUsername, chatID)
	case "/whoami":
		c.cmdWhoami(ctx, fromUsername, chatID, isAdmin)
	case "/app":
		c.cmdApp(chatID)
	case "/help":
		c.replyWithMain(ctx, fromUsername, chatID, helpText(isAdmin))
	case "/avatar":
		c.handleAvatar(ctx, fromUsername, fromUserID, chatID, args, isAdmin)
	// ---- admin ----
	case "/link":
		c.cmdLink(ctx, fromUsername, chatID, args)
	case "/users":
		c.cmdUsers(ctx, chatID)
	case "/bind":
		c.cmdBind(ctx, chatID, args)
	case "/unbind":
		c.cmdUnbind(ctx, chatID, args)
	case "/block", "/unblock":
		c.cmdBlock(ctx, chatID, cmd, args)
	case "/setpusername":
		c.cmdSetPlankaUsername(ctx, chatID, args)
	case "/setavatar":
		c.send(chatID, "Надішліть боту фото з підписом <code>/setavatar &lt;workspace_user&gt;</code>.\n"+
			"Бот завантажить його через Workspace API — перевʼю та БД оновляться коректно.")
	case "/templates":
		c.send(chatID, templatesList(ctx, c.store))
	case "/template":
		c.cmdTemplate(ctx, chatID, args, rest)
	}
}

// parseCommand розбирає вхідний text на (cmd, args, rest, ok).
// rest — це ввесь хвіст після команди (зберігає пробіли всередині).
// strip-ить @botname суфікс у першому токені.
func parseCommand(text string) (cmd string, args []string, rest string, ok bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil, "", false
	}
	// strip @botname у першому токені
	if i := strings.Index(text, "@"); i > 0 && i < len(text) && !strings.Contains(text[:i], " ") {
		if j := strings.Index(text[i:], " "); j > 0 {
			text = text[:i] + text[i+j:]
		} else {
			text = text[:i]
		}
	}
	parts := strings.Fields(text)
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "/") {
		return "", nil, "", false
	}
	cmd = strings.ToLower(parts[0])
	args = parts[1:]
	rest = strings.TrimSpace(strings.TrimPrefix(text, parts[0]))
	return cmd, args, rest, true
}

// ---------- user commands ----------

func (c *Client) cmdStart(ctx context.Context, fromUsername string, fromUserID, chatID int64) {
	if fromUsername == "" {
		c.send(chatID, "❗ Будь ласка, встановіть Telegram username у налаштуваннях.")
		return
	}
	if err := c.store.SetUser(ctx, fromUsername, chatID, fromUserID); err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	b, _ := c.store.GetByTelegram(ctx, fromUsername)
	enabled := b != nil && b.NotificationsEnabled
	c.sendWithMarkup(chatID, startText(fromUsername, enabled), startKeyboard(enabled, c.webAppURL))
}

func (c *Client) cmdStop(ctx context.Context, fromUsername string, chatID int64) {
	if fromUsername == "" {
		c.send(chatID, "❗ Не вдалося визначити ваш username.")
		return
	}
	if err := c.store.SetNotifications(ctx, fromUsername, false); err != nil {
		c.sendWithMarkup(chatID, "Помилка: "+html.EscapeString(err.Error()), mainKeyboard(false, c.webAppURL))
		return
	}
	c.sendWithMarkup(chatID, "🔕 Сповіщення вимкнено. Щоб увімкнути знову — /notify.", mainKeyboard(false, c.webAppURL))
}

// cmdNotify — увімкнення сповіщень. Окрема команда від /start, щоб не плутати
// привітальний екран зі станом підписки.
func (c *Client) cmdNotify(ctx context.Context, fromUsername string, fromUserID, chatID int64) {
	if fromUsername == "" {
		c.send(chatID, "❗ Будь ласка, встановіть Telegram username у налаштуваннях.")
		return
	}
	if err := c.store.SetUser(ctx, fromUsername, chatID, fromUserID); err != nil {
		c.sendWithMarkup(chatID, "Помилка: "+html.EscapeString(err.Error()), mainKeyboard(false, c.webAppURL))
		return
	}
	if err := c.store.SetNotifications(ctx, fromUsername, true); err != nil {
		c.sendWithMarkup(chatID, "Помилка: "+html.EscapeString(err.Error()), mainKeyboard(false, c.webAppURL))
		return
	}
	c.sendWithMarkup(chatID, "🔔 Сповіщення увімкнено. Щоб вимкнути — /stop.", mainKeyboard(true, c.webAppURL))
}

// replyWithMain — звичайний текст з універсальною mainKeyboard. Стан toggle
// підтягується з БД, щоб лейбл кнопки був актуальним.
func (c *Client) replyWithMain(ctx context.Context, fromUsername string, chatID int64, text string) {
	c.sendWithMarkup(chatID, text, mainKeyboard(c.notificationsEnabled(ctx, fromUsername), c.webAppURL))
}

func (c *Client) notificationsEnabled(ctx context.Context, fromUsername string) bool {
	if fromUsername == "" {
		return false
	}
	b, _ := c.store.GetByTelegram(ctx, fromUsername)
	return b != nil && b.NotificationsEnabled
}

func (c *Client) cmdWhoami(ctx context.Context, fromUsername string, chatID int64, isAdmin bool) {
	b, _ := c.store.GetByTelegram(ctx, fromUsername)
	if b == nil {
		c.sendWithMarkup(chatID, "Ви не підписані. Надішліть /start.", mainKeyboard(false, c.webAppURL))
		return
	}
	wsUsername := b.PlankaUsername
	wsLabel := wsUsername
	if wsUsername == "" {
		wsUsername = b.TelegramUsername
		wsLabel = b.TelegramUsername + " <i>(автоматично, без явної привʼязки)</i>"
	} else {
		wsLabel = html.EscapeString(wsLabel)
	}
	role := ""
	if isAdmin {
		role = "\nРоль: <b>адмін</b>"
	}
	caption := fmt.Sprintf("Telegram: <b>@%s</b>\nWorkspace: <b>%s</b>%s",
		html.EscapeString(b.TelegramUsername), wsLabel, role)

	markup := mainKeyboard(b.NotificationsEnabled, c.webAppURL)

	// Намагаємось показати фото профілю. Будь-яка помилка → fallback на текст.
	if photo, ok := c.fetchWorkspaceAvatar(ctx, wsUsername); ok {
		if err := c.sendPhotoMultipart(ctx, chatID, photo, "avatar.jpg", caption, markup); err == nil {
			return
		}
	}
	c.sendWithMarkup(chatID, caption, markup)
}

func (c *Client) cmdApp(chatID int64) {
	if c.webAppURL == "" {
		c.send(chatID, "Mini-app поки не налаштовано.")
		return
	}
	c.sendWithMarkup(chatID, "Відкрийте mini-app кнопкою нижче. Важливо: авторизація Telegram працює тільки з кнопки у приватному чаті з ботом, не зі звичайного посилання.", miniAppKeyboard(c.webAppURL))
}

// fetchWorkspaceAvatar — повертає байти аватара (cover-180), якщо у workspace-юзера
// він є та Planka API доступний.
func (c *Client) fetchWorkspaceAvatar(ctx context.Context, wsUsername string) ([]byte, bool) {
	if wsUsername == "" || !c.planka.Configured() {
		return nil, false
	}
	fileID, ext, ok := c.store.GetPlankaAvatar(ctx, wsUsername)
	if !ok {
		return nil, false
	}
	data, err := c.planka.DownloadAvatar(ctx, fileID, ext, "cover-180")
	if err != nil {
		return nil, false
	}
	return data, true
}

// ---------- admin commands ----------

func (c *Client) cmdLink(ctx context.Context, fromUsername string, chatID int64, args []string) {
	if len(args) != 1 {
		c.send(chatID, "Використання: <code>/link &lt;workspace_username&gt;</code>")
		return
	}
	if exists, _ := c.store.PlankaUserExists(ctx, args[0]); !exists {
		c.send(chatID, "Користувача Workspace <b>"+html.EscapeString(args[0])+"</b> не знайдено.")
		return
	}
	if err := c.store.BindPlanka(ctx, fromUsername, args[0]); err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	c.send(chatID, "🔗 Привʼязано: Telegram <b>@"+html.EscapeString(fromUsername)+
		"</b> ↔ Workspace <b>"+html.EscapeString(args[0])+"</b>.")
}

func (c *Client) cmdUsers(ctx context.Context, chatID int64) {
	us, err := c.store.ListUsers(ctx)
	if err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	if len(us) == 0 {
		c.send(chatID, "Підписників ще немає.")
		return
	}
	var b strings.Builder
	b.WriteString("<b>Підписники:</b>\n")
	for _, u := range us {
		planka := u.PlankaUsername
		if planka == "" {
			planka = "<i>(=" + html.EscapeString(u.TelegramUsername) + ")</i>"
		} else {
			planka = html.EscapeString(planka)
		}
		marker := ""
		switch {
		case u.IsBlocked:
			marker = " 🚫"
		case !u.NotificationsEnabled:
			marker = " 🔕"
		}
		fmt.Fprintf(&b, "• @%s → planka: %s%s\n", html.EscapeString(u.TelegramUsername), planka, marker)
	}
	c.send(chatID, b.String())
}

func (c *Client) cmdBind(ctx context.Context, chatID int64, args []string) {
	if len(args) != 2 {
		c.send(chatID, "Використання: <code>/bind &lt;telegram_user&gt; &lt;workspace_user&gt;</code>")
		return
	}
	tg := strings.TrimPrefix(args[0], "@")
	pk := args[1]
	if exists, _ := c.store.PlankaUserExists(ctx, pk); !exists {
		c.send(chatID, "Workspace-юзера <b>"+html.EscapeString(pk)+"</b> не знайдено.")
		return
	}
	if err := c.store.BindPlanka(ctx, tg, pk); err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	c.send(chatID, "🔗 Привʼязано @"+html.EscapeString(tg)+" ↔ "+html.EscapeString(pk))
}

func (c *Client) cmdUnbind(ctx context.Context, chatID int64, args []string) {
	if len(args) != 1 {
		c.send(chatID, "Використання: <code>/unbind &lt;telegram_user&gt;</code>")
		return
	}
	target := strings.TrimPrefix(args[0], "@")
	if err := c.store.BindPlanka(ctx, target, ""); err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	c.send(chatID, "🔓 Привʼязку знято для @"+html.EscapeString(target))
}

func (c *Client) cmdBlock(ctx context.Context, chatID int64, cmd string, args []string) {
	if len(args) != 1 {
		c.send(chatID, "Використання: <code>"+cmd+" &lt;telegram_user&gt;</code>")
		return
	}
	target := strings.TrimPrefix(args[0], "@")
	if c.store.IsAdmin(ctx, target) {
		c.send(chatID, "⛔ Не можна (роз)блокувати адміна.")
		return
	}
	blocked := cmd == "/block"
	if err := c.store.SetBlocked(ctx, target, blocked); err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	if blocked {
		c.send(chatID, "🚫 Користувача @"+html.EscapeString(target)+" заблоковано. Доступ до бота закрито.")
	} else {
		c.send(chatID, "✅ Користувача @"+html.EscapeString(target)+" розблоковано.")
	}
}

func (c *Client) cmdSetPlankaUsername(ctx context.Context, chatID int64, args []string) {
	if len(args) != 2 {
		c.send(chatID, "Використання: <code>/setpusername &lt;old&gt; &lt;new&gt;</code> — змінює username у Workspace.")
		return
	}
	if err := c.store.UpdatePlankaUsername(ctx, args[0], args[1]); err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	c.send(chatID, "✏️ Workspace username: <b>"+html.EscapeString(args[0])+"</b> → <b>"+html.EscapeString(args[1])+"</b>")
}

// ---------- /template ----------

func (c *Client) cmdTemplate(ctx context.Context, chatID int64, args []string, rest string) {
	if len(args) < 1 {
		c.send(chatID, templateHelp())
		return
	}
	switch strings.ToLower(args[0]) {
	case "show":
		c.cmdTemplateShow(ctx, chatID, args)
	case "set":
		c.cmdTemplateSet(ctx, chatID, args, rest)
	case "reset":
		c.cmdTemplateReset(ctx, chatID, args)
	default:
		c.send(chatID, templateHelp())
	}
}

func (c *Client) cmdTemplateShow(ctx context.Context, chatID int64, args []string) {
	if len(args) < 2 {
		c.send(chatID, "Використання: <code>/template show &lt;event&gt; [channel|personal]</code>")
		return
	}
	event := args[1]
	kind := "channel"
	if len(args) >= 3 {
		kind = args[2]
	}
	cur, custom := c.store.GetTemplate(ctx, event, kind)
	def := ""
	if m, ok := events.DefaultTemplates[event]; ok {
		def = m[kind]
	}
	c.send(chatID, fmt.Sprintf(
		"Подія: <b>%s</b> / <b>%s</b>\n\n<b>Поточний:</b>\n<code>%s</code>\n\n<b>Дефолт:</b>\n<code>%s</code>\n\nКастомний: %v",
		html.EscapeString(event), html.EscapeString(kind),
		html.EscapeString(strOr(cur, def)), html.EscapeString(def), custom))
}

func (c *Client) cmdTemplateSet(ctx context.Context, chatID int64, args []string, rest string) {
	// /template set <event> <kind> <text...>
	tail := strings.TrimSpace(strings.TrimPrefix(rest, args[0]))
	f := strings.SplitN(tail, " ", 3)
	if len(f) < 3 {
		c.send(chatID, "Використання: <code>/template set &lt;event&gt; &lt;channel|personal&gt; &lt;text...&gt;</code>")
		return
	}
	event, kind, tpl := f[0], f[1], f[2]
	if _, err := events.RenderTemplate(tpl, map[string]string{}); err != nil {
		c.send(chatID, "Помилка шаблону: "+html.EscapeString(err.Error()))
		return
	}
	if err := c.store.SetTemplate(ctx, event, kind, tpl); err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	c.send(chatID, "✅ Шаблон збережено для <b>"+html.EscapeString(event)+"</b>/<b>"+html.EscapeString(kind)+"</b>")
}

func (c *Client) cmdTemplateReset(ctx context.Context, chatID int64, args []string) {
	if len(args) < 3 {
		c.send(chatID, "Використання: <code>/template reset &lt;event&gt; &lt;channel|personal&gt;</code>")
		return
	}
	ok, err := c.store.DeleteTemplate(ctx, args[1], args[2])
	if err != nil {
		c.send(chatID, "Помилка: "+html.EscapeString(err.Error()))
		return
	}
	if !ok {
		c.send(chatID, "Кастомного шаблону не було, повертатиметься дефолт.")
		return
	}
	c.send(chatID, "♻️ Скинуто до дефолту.")
}
