package events

import (
	"bytes"
	"text/template"
)

// Render — підготовлений опис однієї події для відправлення.
type Render struct {
	EventKey        string            // ключ для шаблонів (cardCreate, cardMove, taskComplete, ...)
	Data            map[string]string // готові значення (HTML-escaped/anchor-подібні)
	DefaultChannel  string            // дефолтний шаблон для загального чату (компактний рядок)
	DefaultPersonal string            // дефолтний шаблон для DM (порожньо — DM не шлемо)
	Details         string            // деталі (HTML, без шаблонів) — приховуються під кнопкою «Деталі»
	TargetUsername  string            // planka username для DM (порожньо — DM не шлемо)
	TargetUserID    string            // planka user id для резолву username через БД, якщо немає в included
	DMCardID        string            // якщо не порожньо — DM усім учасникам цієї картки (крім актора)
	ActorUsername   string            // planka username актора (виключити з DM)

	// Навігаційні ID для побудови inline-кнопок.
	CardID    string
	BoardID   string
	ProjectID string
}

// RenderTemplate — застосувати шаблон до даних.
//
// missingkey=zero — щоб відсутні поля не давали помилку «<no value>», а
// просто рендерилися як порожні рядки.
func RenderTemplate(tpl string, data map[string]string) (string, error) {
	t, err := template.New("msg").Option("missingkey=zero").Parse(tpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
