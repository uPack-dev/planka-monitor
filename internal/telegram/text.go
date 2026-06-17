package telegram

import (
	"context"
	"fmt"
	"html"
	"strings"

	"planka-monitor/internal/events"
	"planka-monitor/internal/store"
)

func startText(tgUsername string, enabled bool) string {
	state := "🔕 сповіщення вимкнено"
	if enabled {
		state = "🔔 сповіщення увімкнено"
	}
	return "Привіт, <b>@" + html.EscapeString(tgUsername) + "</b>!\n" +
		"Це бот сповіщень Workspace.\n" +
		"Поточний стан: " + state + "."
}

// notifyToggleButton — головна кнопка вмикання/вимикання сповіщень.
func notifyToggleButton(enabled bool) inlineButton {
	if enabled {
		return inlineButton{Text: "🔕 Вимкнути сповіщення", CallbackData: "sub:off"}
	}
	return inlineButton{Text: "🔔 Увімкнути сповіщення", CallbackData: "sub:on"}
}

// mainKeyboard — універсальна клавіатура під будь-яку відповідь бота.
// Містить toggle сповіщень, /whoami та /help для швидкого доступу.
func miniAppButton(webAppURL string) inlineButton {
	return inlineButton{Text: "📱 Відкрити mini-app", WebApp: &webAppInfo{URL: webAppURL}}
}

func miniAppKeyboard(webAppURL string) *inlineKeyboard {
	if webAppURL == "" {
		return nil
	}
	return &inlineKeyboard{InlineKeyboard: [][]inlineButton{{miniAppButton(webAppURL)}}}
}

func mainKeyboard(enabled bool, webAppURL ...string) *inlineKeyboard {
	rows := [][]inlineButton{{notifyToggleButton(enabled)}}
	if len(webAppURL) > 0 && webAppURL[0] != "" {
		rows = append(rows, []inlineButton{miniAppButton(webAppURL[0])})
	}
	rows = append(rows, []inlineButton{
		{Text: "👤 Хто я", CallbackData: "whoami"},
		{Text: "❓ Допомога", CallbackData: "help"},
	})
	return &inlineKeyboard{InlineKeyboard: rows}
}

// startKeyboard — для екрана привітання (історично — той самий набір, що mainKeyboard).
func startKeyboard(enabled bool, webAppURL ...string) *inlineKeyboard {
	return mainKeyboard(enabled, webAppURL...)
}

func helpText(isAdmin bool) string {
	var b strings.Builder
	b.WriteString("<b>Доступні команди:</b>\n")
	b.WriteString("/start — реєстрація / привітальний екран\n")
	b.WriteString("/notify — увімкнути сповіщення\n")
	b.WriteString("/stop — вимкнути сповіщення\n")
	b.WriteString("/whoami — показати поточну привʼязку\n")
	b.WriteString("/app — відкрити mini-app\n")
	b.WriteString("/avatar — поставити поточне Telegram-фото профілю як аватар у Workspace\n")
	b.WriteString("/help — ця довідка\n")
	if isAdmin {
		b.WriteString("\n<b>Адмін-команди:</b>\n")
		b.WriteString("/users — список підписників\n")
		b.WriteString("/bind &lt;tg&gt; &lt;workspace&gt; — привʼязати чужий акаунт\n")
		b.WriteString("/unbind &lt;tg&gt; — зняти привʼязку\n")
		b.WriteString("/block &lt;tg&gt; — заблокувати користувача\n")
		b.WriteString("/unblock &lt;tg&gt; — розблокувати користувача\n")
		b.WriteString("/link &lt;workspace&gt; — привʼязати свій telegram до Workspace-юзера\n")
		b.WriteString("/setpusername &lt;old&gt; &lt;new&gt; — змінити username у Workspace\n")
		b.WriteString("/setavatar — інфо про зміну аватара\n")
		b.WriteString("/templates — список шаблонів\n")
		b.WriteString("/template show|set|reset — керування шаблонами\n")
	}
	return b.String()
}

func templateHelp() string {
	return "<b>Команди шаблонів:</b>\n" +
		"<code>/template show &lt;event&gt; [channel|personal]</code>\n" +
		"<code>/template set &lt;event&gt; &lt;channel|personal&gt; &lt;text...&gt;</code>\n" +
		"<code>/template reset &lt;event&gt; &lt;channel|personal&gt;</code>\n\n" +
		"Шаблони — це Go text/template. Доступні змінні залежать від події; див. /templates."
}

func templatesList(ctx context.Context, st *store.Store) string {
	var b strings.Builder
	b.WriteString("<b>Усі події:</b>\n")
	for _, ev := range events.KnownEvents() {
		for _, kind := range events.KnownKinds(ev) {
			cur, custom := st.GetTemplate(ctx, ev, kind)
			marker := "  "
			if custom {
				marker = "✏️"
			}
			def := events.DefaultTemplates[ev][kind]
			fmt.Fprintf(&b, "\n%s <b>%s</b>/<b>%s</b>\n<code>%s</code>", marker,
				html.EscapeString(ev), html.EscapeString(kind),
				html.EscapeString(strOr(cur, def)))
		}
	}
	return b.String()
}

func strOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
