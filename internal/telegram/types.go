package telegram

// JSON-DTO для Telegram Bot API. Не плутати з типами в internal/events
// (там моделі Planka).

type inlineButton struct {
	Text         string      `json:"text"`
	CallbackData string      `json:"callback_data,omitempty"`
	URL          string      `json:"url,omitempty"`
	WebApp       *webAppInfo `json:"web_app,omitempty"`
}

type webAppInfo struct {
	URL string `json:"url"`
}

type inlineKeyboard struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

type menuButton struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	WebApp *webAppInfo `json:"web_app,omitempty"`
}

type sendReq struct {
	ChatID                any             `json:"chat_id"`
	Text                  string          `json:"text"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool            `json:"disable_web_page_preview,omitempty"`
	ReplyMarkup           *inlineKeyboard `json:"reply_markup,omitempty"`
}

type editReq struct {
	ChatID                int64           `json:"chat_id"`
	MessageID             int64           `json:"message_id"`
	Text                  string          `json:"text"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool            `json:"disable_web_page_preview,omitempty"`
	ReplyMarkup           *inlineKeyboard `json:"reply_markup,omitempty"`
}

type photoSize struct {
	FileID   string `json:"file_id"`
	FileSize int    `json:"file_size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type tgMessage struct {
	MessageID int64       `json:"message_id"`
	Text      string      `json:"text"`
	Caption   string      `json:"caption"`
	Photo     []photoSize `json:"photo"`
	Document  *struct {
		FileID   string `json:"file_id"`
		FileName string `json:"file_name"`
		MimeType string `json:"mime_type"`
	} `json:"document"`
	From struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	} `json:"from"`
	Chat struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	} `json:"chat"`
}

type callbackQuery struct {
	ID   string `json:"id"`
	From struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	} `json:"from"`
	Message *tgMessage `json:"message"`
	Data    string     `json:"data"`
}

type update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *tgMessage     `json:"message"`
	CallbackQuery *callbackQuery `json:"callback_query"`
}

type updatesResp struct {
	OK     bool     `json:"ok"`
	Result []update `json:"result"`
}

type botCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type setCmdScope struct {
	Type   string `json:"type"`
	ChatID int64  `json:"chat_id,omitempty"`
}

type sendMessageResp struct {
	OK     bool `json:"ok"`
	Result struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

type profilePhotosResp struct {
	OK     bool `json:"ok"`
	Result struct {
		TotalCount int           `json:"total_count"`
		Photos     [][]photoSize `json:"photos"`
	} `json:"result"`
}

type fileResp struct {
	OK     bool `json:"ok"`
	Result struct {
		FilePath string `json:"file_path"`
	} `json:"result"`
}
