package events

import (
	"encoding/json"
	"html"
)

func produceComment(c *prodCtx) []Render {
	// commentCreate (Planka v2) кладе text безпосередньо в item.
	// actionCreate (legacy) — у data.text, разом із data.type="commentCard".
	var item struct {
		Type string `json:"type"`
		Text string `json:"text"`
		Data struct {
			Text string `json:"text"`
		} `json:"data"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)
	if item.Type != "" && item.Type != "commentCard" {
		return nil
	}
	text := item.Text
	if text == "" {
		text = item.Data.Text
	}
	text = truncate(text, 1500)

	r := c.mk("commentCreate", nil, "")
	r.DMCardID = c.card.ID
	if text != "" {
		r.Details = "<blockquote>" + html.EscapeString(text) + "</blockquote>"
	}
	return []Render{r}
}

func produceAttachment(c *prodCtx) []Render {
	var item struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)
	return []Render{c.mk("attachmentCreate", map[string]string{"AttachName": html.EscapeString(item.Name)}, "")}
}

func produceList(c *prodCtx, event string) []Render {
	var item List
	_ = json.Unmarshal(c.p.Data.Item, &item)
	return []Render{c.mk(event, map[string]string{"ListName": html.EscapeString(item.Name)}, "")}
}

func produceBoard(c *prodCtx, event string) []Render {
	var item Board
	_ = json.Unmarshal(c.p.Data.Item, &item)
	c.base["BoardLink"] = boardLink(c.baseURL, item)
	r := c.mk(event, nil, "")
	r.BoardID = item.ID
	r.CardID = ""
	return []Render{r}
}

func produceProject(c *prodCtx) []Render {
	var item Project
	_ = json.Unmarshal(c.p.Data.Item, &item)
	r := c.mk("projectCreate", map[string]string{"ProjectName": html.EscapeString(item.Name)}, "")
	r.ProjectID = item.ID
	r.CardID = ""
	r.BoardID = ""
	return []Render{r}
}
