package events

import (
	"html"
	"strings"
)

// prodCtx — контекст рендеру для всіх обробників: спільні поля payload-у +
// preсчитаний `base` (актор, посилання на картку/дошку, контекст-суфікс).
type prodCtx struct {
	p             *Payload
	baseURL       string
	base          map[string]string
	board         Board
	list          List
	card          Card
	actorUsername string
}

func newProdCtx(p *Payload, baseURL string) *prodCtx {
	baseURL = strings.TrimRight(baseURL, "/")

	actorRaw := "Хтось"
	actorUsername := ""
	if p.User != nil {
		actorUsername = p.User.Username
		switch {
		case p.User.Name != "":
			actorRaw = p.User.Name
		case p.User.Username != "":
			actorRaw = "@" + p.User.Username
		}
	}

	board := firstBoard(p)
	list := firstList(p)
	card := firstCard(p)

	return &prodCtx{
		p:             p,
		baseURL:       baseURL,
		board:         board,
		list:          list,
		card:          card,
		actorUsername: actorUsername,
		base: map[string]string{
			"Actor":     "<b>" + html.EscapeString(actorRaw) + "</b>",
			"BoardLink": boardLink(baseURL, board),
			"CardLink":  cardLink(baseURL, card),
			"ListName":  html.EscapeString(list.Name),
			"Context":   contextSuffix(baseURL, board, list),
		},
	}
}

// mk будує Render із спільних полів base + extra. target — planka username для DM.
func (c *prodCtx) mk(eventKey string, extra map[string]string, target string) Render {
	data := make(map[string]string, len(c.base)+len(extra))
	for k, v := range c.base {
		data[k] = v
	}
	for k, v := range extra {
		data[k] = v
	}
	def := DefaultTemplates[eventKey]
	return Render{
		EventKey:        eventKey,
		Data:            data,
		DefaultChannel:  def["channel"],
		DefaultPersonal: def["personal"],
		TargetUsername:  target,
		ActorUsername:   c.actorUsername,
		CardID:          c.card.ID,
		BoardID:         c.board.ID,
	}
}

// Produce — головний диспетчер. Повертає 0..N Render-ів для одного payload-у.
func Produce(p *Payload, baseURL string) []Render {
	c := newProdCtx(p, baseURL)

	switch p.Event {
	case "cardCreate", "cardDelete":
		return produceCardCreateDelete(c, p.Event)
	case "cardUpdate":
		return produceCardUpdate(c)
	case "cardMembershipCreate", "cardMembershipDelete":
		return produceCardMembership(c, p.Event)
	case "cardLabelCreate", "cardLabelDelete":
		return produceCardLabel(c, p.Event)

	case "taskCreate":
		return produceTaskCreate(c)
	case "taskDelete":
		return produceTaskDelete(c)
	case "taskUpdate":
		return produceTaskUpdate(c)
	case "taskMembershipCreate", "taskMembershipDelete":
		return produceTaskMembership(c, p.Event)

	case "commentCreate", "actionCreate":
		return produceComment(c)
	case "attachmentCreate":
		return produceAttachment(c)
	case "listCreate", "listDelete":
		return produceList(c, p.Event)
	case "boardCreate", "boardDelete":
		return produceBoard(c, p.Event)
	case "projectCreate":
		return produceProject(c)
	}
	return nil
}
