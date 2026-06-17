package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
)

// methodURL будує URL для виклику довільного методу Bot API.
func (c *Client) methodURL(method string) string {
	return fmt.Sprintf("%s/bot%s/%s", telegramAPIBase, c.token, method)
}

// callJSON — універсальна точка POST для Telegram Bot API.
//   - method: ім'я Bot API методу (наприклад "sendMessage").
//   - in:     payload, який буде marshall-нутий у JSON. nil → порожнє тіло.
//   - out:    куди декодити відповідь. nil → відповідь дропається.
//
// Помилки HTTP/мережі логуються та повертаються; >=300 декодується в текст
// для діагностики.
func (c *Client) callJSON(ctx context.Context, method string, in, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshal %s: %w", method, err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.methodURL(method), body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("telegram %s: %v", method, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("telegram %s status=%d body=%s", method, resp.StatusCode, string(b))
		return fmt.Errorf("telegram %s: status %d", method, resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		log.Printf("telegram %s decode: %v", method, err)
		return err
	}
	return nil
}

// callGET виконує GET для Bot API методів типу getUpdates/getFile/getUserProfilePhotos.
func (c *Client) callGET(ctx context.Context, method, query string, out any) error {
	u := c.methodURL(method)
	if query != "" {
		u += "?" + query
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("telegram %s: %v", method, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("telegram %s status=%d body=%s", method, resp.StatusCode, string(b))
		return fmt.Errorf("telegram %s: status %d", method, resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// send — POST sendMessage без захоплення message_id.
func (c *Client) send(chatID any, text string) {
	c.sendWithMarkup(chatID, text, nil)
}

func (c *Client) sendWithMarkup(chatID any, text string, markup *inlineKeyboard) {
	_ = c.callJSON(context.Background(), "sendMessage", sendReq{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
		ReplyMarkup:           markup,
	}, nil)
}

// sendCapture — POST sendMessage із захопленням message_id для подальшого edit-у.
// Повертає 0 при помилці.
func (c *Client) sendCapture(chatID any, text string, markup *inlineKeyboard) int64 {
	var r sendMessageResp
	err := c.callJSON(context.Background(), "sendMessage", sendReq{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
		ReplyMarkup:           markup,
	}, &r)
	if err != nil {
		return 0
	}
	return r.Result.MessageID
}

func (c *Client) editMessage(chatID, messageID int64, text string, markup *inlineKeyboard) {
	_ = c.callJSON(context.Background(), "editMessageText", editReq{
		ChatID:                chatID,
		MessageID:             messageID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
		ReplyMarkup:           markup,
	}, nil)
}

func (c *Client) answerCallback(callbackID, text string, alert bool) {
	_ = c.callJSON(context.Background(), "answerCallbackQuery", map[string]any{
		"callback_query_id": callbackID,
		"text":              text,
		"show_alert":        alert,
	}, nil)
}

func (c *Client) setMyCommands(ctx context.Context, cmds []botCommand, scope setCmdScope) {
	_ = c.callJSON(ctx, "setMyCommands", map[string]any{
		"commands": cmds,
		"scope":    scope,
	}, nil)
}

// getProfilePhotoFileID — повертає file_id найбільшого розміру першого фото профілю.
func (c *Client) getProfilePhotoFileID(ctx context.Context, userID int64) (string, error) {
	q := fmt.Sprintf("user_id=%d&limit=1", userID)
	var pr profilePhotosResp
	if err := c.callGET(ctx, "getUserProfilePhotos", q, &pr); err != nil {
		return "", err
	}
	if !pr.OK || pr.Result.TotalCount == 0 || len(pr.Result.Photos) == 0 {
		return "", fmt.Errorf("у користувача немає фото профілю або воно приховане налаштуваннями приватності")
	}
	sizes := pr.Result.Photos[0]
	if len(sizes) == 0 {
		return "", fmt.Errorf("порожній набір розмірів")
	}
	return sizes[len(sizes)-1].FileID, nil
}

// sendPhotoMultipart — POST sendPhoto з multipart-завантаженням файлу.
// Caption підтримує HTML.
func (c *Client) sendPhotoMultipart(ctx context.Context, chatID any, photo []byte, filename, caption string, markup *inlineKeyboard) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("chat_id", fmt.Sprint(chatID)); err != nil {
		return err
	}
	if caption != "" {
		_ = mw.WriteField("caption", caption)
		_ = mw.WriteField("parse_mode", "HTML")
	}
	if markup != nil {
		raw, err := json.Marshal(markup)
		if err != nil {
			return err
		}
		_ = mw.WriteField("reply_markup", string(raw))
	}
	fw, err := mw.CreateFormFile("photo", filename)
	if err != nil {
		return err
	}
	if _, err := fw.Write(photo); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.methodURL("sendPhoto"), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("telegram sendPhoto: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("telegram sendPhoto status=%d body=%s", resp.StatusCode, string(b))
		return fmt.Errorf("telegram sendPhoto: status %d", resp.StatusCode)
	}
	return nil
}

// downloadFile — getFile + завантаження вмісту по file_path.
func (c *Client) downloadFile(ctx context.Context, fileID string) ([]byte, error) {
	var fr fileResp
	if err := c.callGET(ctx, "getFile", "file_id="+url.QueryEscape(fileID), &fr); err != nil {
		return nil, err
	}
	if !fr.OK || fr.Result.FilePath == "" {
		return nil, fmt.Errorf("getFile: bad response")
	}
	dl := fmt.Sprintf("%s/file/bot%s/%s", telegramAPIBase, c.token, fr.Result.FilePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
