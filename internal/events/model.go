// Package events перетворює сирі payload-и Planka webhook-ів у Render-и.
//
// Файлова структура:
//
//	model.go         — типи payload-у та entity-моделі.
//	render.go        — Render-структура + RenderTemplate (text/template wrapper).
//	templates.go     — DefaultTemplates, KnownEvents, KnownKinds.
//	helpers.go       — спільні утиліти (URL/HTML-формування, lookup в Included).
//	produce.go       — точка входу Produce + диспетчер по event-ах.
//	handlers_card.go — обробники cardCreate / cardDelete / cardUpdate / cardMembership* / cardLabel*.
//	handlers_task.go — taskCreate / taskDelete / taskUpdate / taskMembership*.
//	handlers_misc.go — commentCreate / attachmentCreate / list*/board*/projectCreate.
package events

import "encoding/json"

type Payload struct {
	Event    string          `json:"event"`
	Data     Data            `json:"data"`
	PrevData json.RawMessage `json:"prevData"`
	User     *User           `json:"user"`
}

type Data struct {
	Item     json.RawMessage `json:"item"`
	Included Included        `json:"included"`
}

type Included struct {
	Users    []User    `json:"users"`
	Projects []Project `json:"projects"`
	Boards   []Board   `json:"boards"`
	Lists    []List    `json:"lists"`
	Cards    []Card    `json:"cards"`
	Labels   []Label   `json:"labels"`
	Tasks    []Task    `json:"tasks"`
}

type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Task struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	CardID string `json:"cardId"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Board struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
}

type List struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BoardID string `json:"boardId"`
}

type Card struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BoardID string `json:"boardId"`
	ListID  string `json:"listId"`
}
