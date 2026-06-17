package server

import (
	"context"
	"html"
	"log"
	"strings"

	"planka-monitor/internal/events"
	"planka-monitor/internal/store"
	"planka-monitor/internal/telegram"
)

// dispatch — перетворює events.Render у Telegram-надсилання
// (channel-повідомлення + 0..N personal DM).
func (s *Server) dispatch(p *events.Payload) {
	ctx, cancel := context.WithTimeout(context.Background(), dispatchTimeout)
	defer cancel()

	for _, r := range events.Produce(p, s.baseURL) {
		s.enrichLabelName(ctx, &r)
		if r.EventKey == "cardLabelDelete" && r.Data["LabelName"] == "" {
			continue
		}
		if r.EventKey == "cardLabelCreate" && r.Data["LabelName"] == "" {
			r.Data["LabelName"] = "мітка"
		}

		urls := s.buildURLs(r)
		s.sendChannel(ctx, r, urls)
		s.sendPersonal(ctx, r, urls)
	}
}

// enrichLabelName — для cardLabel*-подій ім'я мітки часто відсутнє в payload
// (особливо для delete). Підтягуємо з БД як best-effort.
func (s *Server) enrichLabelName(ctx context.Context, r *events.Render) {
	if r.EventKey != "cardLabelCreate" && r.EventKey != "cardLabelDelete" {
		return
	}
	if r.Data["LabelName"] != "" {
		return
	}
	id := r.Data["LabelID"]
	if id == "" {
		return
	}
	if name := s.store.GetLabelName(ctx, id); name != "" {
		r.Data["LabelName"] = html.EscapeString(name)
	}
}

type notifURLs struct{ card, board, project string }

func (s *Server) buildURLs(r events.Render) notifURLs {
	base := strings.TrimRight(s.baseURL, "/")
	var u notifURLs
	if r.CardID != "" {
		u.card = base + "/cards/" + r.CardID
	}
	if r.BoardID != "" {
		u.board = base + "/boards/" + r.BoardID
	}
	if r.ProjectID != "" {
		u.project = base + "/projects/" + r.ProjectID
	}
	return u
}

func (s *Server) sendChannel(ctx context.Context, r events.Render, u notifURLs) {
	tpl := pickTemplate(ctx, s.store, r.EventKey, "channel", r.DefaultChannel)
	if tpl == "" {
		return
	}
	heading, err := events.RenderTemplate(tpl, r.Data)
	if err != nil {
		log.Printf("render %s/channel: %v", r.EventKey, err)
		return
	}
	if heading == "" {
		return
	}
	s.tg.SendNotificationChannel(telegram.Notification{
		Heading:    heading,
		Details:    r.Details,
		CardURL:    u.card,
		BoardURL:   u.board,
		ProjectURL: u.project,
	})
}

func (s *Server) sendPersonal(ctx context.Context, r events.Render, u notifURLs) {
	tpl := pickTemplate(ctx, s.store, r.EventKey, "personal", r.DefaultPersonal)
	if tpl == "" {
		return
	}

	targets := s.collectPersonalTargets(ctx, &r)
	if len(targets) == 0 {
		return
	}

	heading, err := events.RenderTemplate(tpl, r.Data)
	if err != nil {
		log.Printf("render %s/personal: %v", r.EventKey, err)
		return
	}
	if heading == "" {
		return
	}
	notif := telegram.Notification{
		Heading:    heading,
		Details:    r.Details,
		CardURL:    u.card,
		BoardURL:   u.board,
		ProjectURL: u.project,
	}
	category := events.PersonalCategory(r.EventKey)
	cardID := r.CardID
	if cardID == "" {
		cardID = r.DMCardID
	}
	for username := range targets {
		if cardID != "" && s.store.CardMutedForPlankaUser(ctx, username, cardID) {
			continue
		}
		if category != "" && !s.store.PersonalCategoryEnabled(ctx, username, category) {
			continue
		}
		if !s.tg.SendNotificationToPlankaUser(ctx, username, notif) {
			log.Printf("no telegram mapping for planka user %q", username)
		}
	}
}

// collectPersonalTargets — формує множину planka usernames для DM (без актора).
func (s *Server) collectPersonalTargets(ctx context.Context, r *events.Render) map[string]struct{} {
	if r.TargetUsername == "" && r.TargetUserID != "" {
		r.TargetUsername = s.store.GetUsernameByID(ctx, r.TargetUserID)
	}
	targets := map[string]struct{}{}
	if r.TargetUsername != "" {
		targets[strings.ToLower(r.TargetUsername)] = struct{}{}
	}
	if r.DMCardID != "" {
		members, err := s.store.ListCardMemberUsernames(ctx, r.DMCardID)
		if err != nil {
			log.Printf("card members %s: %v", r.DMCardID, err)
		}
		for _, m := range members {
			targets[strings.ToLower(m)] = struct{}{}
		}
	}
	if r.ActorUsername != "" {
		delete(targets, strings.ToLower(r.ActorUsername))
	}
	return targets
}

func pickTemplate(ctx context.Context, st *store.Store, event, kind, def string) string {
	if t, ok := st.GetTemplate(ctx, event, kind); ok {
		return t
	}
	return def
}
