package telegram

import (
	"context"
	"log"
)

var userCommands = []botCommand{
	{"start", "реєстрація / привітальний екран"},
	{"notify", "увімкнути сповіщення"},
	{"stop", "вимкнути сповіщення"},
	{"whoami", "показати мою привʼязку"},
	{"app", "відкрити mini-app"},
	{"avatar", "поставити моє Telegram-фото в Workspace"},
	{"help", "довідка"},
}

var adminExtraCommands = []botCommand{
	{"users", "список підписників"},
	{"bind", "адмін: привʼязати чужий акаунт"},
	{"unbind", "адмін: зняти привʼязку"},
	{"block", "адмін: заблокувати користувача"},
	{"unblock", "адмін: розблокувати користувача"},
	{"setpusername", "адмін: змінити username у Workspace"},
	{"setavatar", "адмін: змінити аватар (фото з підписом)"},
	{"templates", "адмін: список шаблонів"},
	{"template", "адмін: керувати шаблоном"},
}

// adminCommandSet — для тихої ігнорації адмін-команд від не-адмінів,
// щоб не розкривати наявність адмін-функціоналу.
var adminCommandSet = map[string]struct{}{
	"/users":        {},
	"/bind":         {},
	"/unbind":       {},
	"/block":        {},
	"/unblock":      {},
	"/setpusername": {},
	"/setavatar":    {},
	"/templates":    {},
	"/template":     {},
	"/link":         {}, // привʼязка тепер лише через адміна
}

// ensureChatMenu ставить меню для конкретного чату відповідно до прав користувача.
// Перезаписує adminCmd<->userCmd при зміні ролі та робить це лише коли потрібно.
func (c *Client) ensureChatMenu(ctx context.Context, chatID int64, isAdmin bool) {
	c.mu.Lock()
	known := c.chatMenuKnown[chatID]
	prev := c.chatMenuAdmin[chatID]
	if known && prev == isAdmin {
		c.mu.Unlock()
		return
	}
	c.chatMenuKnown[chatID] = true
	c.chatMenuAdmin[chatID] = isAdmin
	c.mu.Unlock()

	cmds := append([]botCommand{}, userCommands...)
	if isAdmin {
		cmds = append(cmds, adminExtraCommands...)
	}
	c.setMyCommands(ctx, cmds, setCmdScope{Type: "chat", ChatID: chatID})
}

func (c *Client) ensureDefaultMenuButton(ctx context.Context) {
	if c.webAppURL == "" {
		return
	}
	if err := c.callJSON(ctx, "setChatMenuButton", map[string]any{
		"menu_button": menuButton{
			Type:   "web_app",
			Text:   "Mini-app",
			WebApp: &webAppInfo{URL: c.webAppURL},
		},
	}, nil); err != nil {
		log.Printf("telegram setChatMenuButton: %v", err)
	}
}
