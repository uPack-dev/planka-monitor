package events

import (
	"encoding/json"
	"strings"
	"testing"
)

// helper: build a Payload with given event/item/prevItem/included.
// prevItem відповідає реальному формату цього форка Planka:
// PrevData = {"item": {...}, "included": {...}}.
func mkPayload(t *testing.T, event string, item any, prevItem any, inc Included, actor *User) *Payload {
	t.Helper()
	itemRaw, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal item: %v", err)
	}
	var prevRaw json.RawMessage
	if prevItem != nil {
		b, err := json.Marshal(map[string]any{"item": prevItem})
		if err != nil {
			t.Fatalf("marshal prev: %v", err)
		}
		prevRaw = b
	}
	return &Payload{
		Event: event,
		Data: Data{
			Item:     itemRaw,
			Included: inc,
		},
		PrevData: prevRaw,
		User:     actor,
	}
}

func single(t *testing.T, rs []Render) Render {
	t.Helper()
	if len(rs) != 1 {
		t.Fatalf("expected 1 render, got %d", len(rs))
	}
	return rs[0]
}

func TestProduce_CardLabelCreate(t *testing.T) {
	p := mkPayload(t,
		"cardLabelCreate",
		map[string]string{"cardId": "10", "labelId": "55"},
		nil,
		Included{
			Cards:  []Card{{ID: "10", Name: "Buy milk"}},
			Labels: []Label{{ID: "55", Name: "urgent", Color: "red"}},
		},
		&User{ID: "1", Username: "alice", Name: "Alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "cardLabelCreate" {
		t.Errorf("eventKey=%s", r.EventKey)
	}
	if got := r.Data["LabelName"]; got != "urgent" {
		t.Errorf("LabelName=%q", got)
	}
	if !strings.Contains(r.Data["CardLink"], "/cards/10") {
		t.Errorf("CardLink=%q", r.Data["CardLink"])
	}
	if r.ActorUsername != "alice" {
		t.Errorf("ActorUsername=%q", r.ActorUsername)
	}
	if r.DefaultChannel == "" {
		t.Errorf("default channel template missing")
	}
}

func TestProduce_CardLabelDelete_NoIncludedNamePassesIDForResolve(t *testing.T) {
	p := mkPayload(t,
		"cardLabelDelete",
		map[string]string{"cardId": "10", "labelId": "999"},
		nil,
		Included{Cards: []Card{{ID: "10", Name: "X"}}}, // no labels in included
		&User{Username: "bob"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	// Producer не має ставити fallback "мітка" — лишає порожньо й передає ID
	// для DB-резолву на рівні server.go.
	if r.Data["LabelName"] != "" {
		t.Errorf("expected empty LabelName for delete w/o included, got %q", r.Data["LabelName"])
	}
	if r.Data["LabelID"] != "999" {
		t.Errorf("LabelID must be propagated, got %q", r.Data["LabelID"])
	}
}

func TestProduce_TaskMembershipCreate(t *testing.T) {
	p := mkPayload(t,
		"taskMembershipCreate",
		map[string]string{"taskId": "77", "userId": "u2"},
		nil,
		Included{
			Cards: []Card{{ID: "10", Name: "Card"}},
			Tasks: []Task{{ID: "77", Name: "Subtask A", CardID: "10"}},
			Users: []User{{ID: "u2", Username: "carol", Name: "Carol"}},
		},
		&User{ID: "u1", Username: "alice", Name: "Alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "taskAssign" {
		t.Errorf("eventKey=%s", r.EventKey)
	}
	if r.Data["TaskName"] != "Subtask A" {
		t.Errorf("TaskName=%q", r.Data["TaskName"])
	}
	if r.Data["TargetName"] != "Carol" {
		t.Errorf("TargetName=%q", r.Data["TargetName"])
	}
	if r.TargetUsername != "carol" {
		t.Errorf("TargetUsername=%q (must DM the assignee)", r.TargetUsername)
	}
	if r.DefaultPersonal == "" {
		t.Errorf("personal template missing for taskAssign")
	}
}

func TestProduce_TaskMembershipDelete(t *testing.T) {
	p := mkPayload(t,
		"taskMembershipDelete",
		map[string]string{"taskId": "77", "userId": "u2"},
		nil,
		Included{
			Tasks: []Task{{ID: "77", Name: "Sub"}},
			Users: []User{{ID: "u2", Username: "carol"}},
		},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "taskUnassign" {
		t.Errorf("eventKey=%s", r.EventKey)
	}
	if r.TargetUsername != "carol" {
		t.Errorf("TargetUsername=%q", r.TargetUsername)
	}
}

func TestProduce_TaskUpdate_AssigneeChange(t *testing.T) {
	// new assigneeUserId
	p := mkPayload(t,
		"taskUpdate",
		map[string]any{"name": "Sub", "assigneeUserId": "u9"},
		map[string]any{"assigneeUserId": nil},
		Included{
			Cards: []Card{{ID: "10", Name: "Card"}},
			Users: []User{{ID: "u9", Username: "dan"}},
		},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "taskAssign" {
		t.Fatalf("eventKey=%s", r.EventKey)
	}
	if r.TargetUsername != "dan" {
		t.Errorf("TargetUsername=%q", r.TargetUsername)
	}
	if r.TargetUserID != "u9" {
		t.Errorf("TargetUserID=%q (must be set for DB resolve fallback)", r.TargetUserID)
	}

	// unassign: previously had user, now empty
	p2 := mkPayload(t,
		"taskUpdate",
		map[string]any{"name": "Sub", "assigneeUserId": ""},
		map[string]any{"assigneeUserId": "u9"},
		Included{Users: []User{{ID: "u9", Username: "dan"}}},
		&User{Username: "alice"},
	)
	r2 := single(t, Produce(p2, "https://planka.test"))
	if r2.EventKey != "taskUnassign" {
		t.Errorf("expected taskUnassign, got %s", r2.EventKey)
	}
	if r2.TargetUsername != "dan" {
		t.Errorf("TargetUsername=%q", r2.TargetUsername)
	}
}

func TestProduce_TaskUpdate_DescriptionEdit(t *testing.T) {
	p := mkPayload(t,
		"taskUpdate",
		map[string]any{"name": "Sub", "description": "new"},
		map[string]any{"description": "old"},
		Included{Cards: []Card{{ID: "10", Name: "Card"}}},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "taskEdit" {
		t.Errorf("eventKey=%s", r.EventKey)
	}
	if r.Data["TaskName"] != "Sub" {
		t.Errorf("TaskName=%q", r.Data["TaskName"])
	}
}

func TestProduce_TaskUpdate_CompleteStillWorks(t *testing.T) {
	p := mkPayload(t,
		"taskUpdate",
		map[string]any{"name": "Sub", "isCompleted": true, "assigneeUserId": "u9"},
		map[string]any{"isCompleted": false},
		Included{Users: []User{{ID: "u9", Username: "dan"}}},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "taskComplete" {
		t.Errorf("regression: eventKey=%s", r.EventKey)
	}
	if r.TargetUserID != "u9" {
		t.Errorf("expected TargetUserID=u9 (assignee should be DM'd on complete), got %q", r.TargetUserID)
	}
	if r.TargetUsername != "dan" {
		t.Errorf("expected TargetUsername=dan, got %q", r.TargetUsername)
	}
	if r.DefaultPersonal == "" {
		t.Errorf("personal template missing for taskComplete")
	}
}

func TestProduce_CommentCreate_SetsDMCardID(t *testing.T) {
	p := mkPayload(t,
		"commentCreate",
		map[string]any{
			"type": "commentCard",
			"data": map[string]any{"text": "hello"},
		},
		nil,
		Included{
			Cards: []Card{{ID: "abc", Name: "Card"}},
		},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "commentCreate" {
		t.Fatalf("eventKey=%s", r.EventKey)
	}
	if r.DMCardID != "abc" {
		t.Errorf("DMCardID=%q (should match included card id)", r.DMCardID)
	}
	if r.ActorUsername != "alice" {
		t.Errorf("ActorUsername=%q", r.ActorUsername)
	}
	if r.DefaultPersonal == "" {
		t.Errorf("personal template missing for commentCreate")
	}
	if !strings.Contains(r.Details, "hello") {
		t.Errorf("Details=%q (must contain comment text)", r.Details)
	}
	if r.CardID != "abc" {
		t.Errorf("CardID=%q (must be set for navigation button)", r.CardID)
	}
}

func TestProduce_CardUpdate_DescriptionEdit_DMs(t *testing.T) {
	p := mkPayload(t,
		"cardUpdate",
		map[string]any{"id": "10", "name": "Card", "listId": "L1"},
		map[string]any{"description": "old text"},
		Included{Cards: []Card{{ID: "10", Name: "Card", ListID: "L1"}}},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "cardEdit" {
		t.Fatalf("eventKey=%s", r.EventKey)
	}
	if r.DMCardID != "10" {
		t.Errorf("DMCardID=%q", r.DMCardID)
	}
	if !strings.Contains(r.Details, "опис") {
		t.Errorf("Details missing description marker: %q", r.Details)
	}
}

func TestProduce_CardUpdate_Move_NoRegression(t *testing.T) {
	p := mkPayload(t,
		"cardUpdate",
		map[string]any{"id": "10", "name": "Card", "listId": "L2"},
		map[string]any{"listId": "L1"},
		Included{Cards: []Card{{ID: "10", Name: "Card"}}},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.EventKey != "cardMove" {
		t.Errorf("eventKey=%s", r.EventKey)
	}
}

func TestProduce_CardMembership_DirectTarget(t *testing.T) {
	p := mkPayload(t,
		"cardMembershipCreate",
		map[string]any{"cardId": "10", "userId": "u2"},
		nil,
		Included{
			Cards: []Card{{ID: "10", Name: "Card"}},
			Users: []User{{ID: "u2", Username: "carol", Name: "Carol"}},
		},
		&User{Username: "alice"},
	)
	r := single(t, Produce(p, "https://planka.test"))
	if r.TargetUsername != "carol" {
		t.Errorf("TargetUsername=%q", r.TargetUsername)
	}
	if r.ActorUsername != "alice" {
		t.Errorf("ActorUsername=%q", r.ActorUsername)
	}
	if r.DefaultPersonal == "" {
		t.Errorf("personal template missing")
	}
}

func TestRenderTemplate_NewEventsRender(t *testing.T) {
	cases := []struct {
		event, kind string
		data        map[string]string
		mustHave    []string
	}{
		{"cardLabelCreate", "channel",
			map[string]string{"Actor": "<b>A</b>", "LabelName": "urgent", "CardLink": "<a>X</a>", "Context": ""},
			[]string{"urgent", "X"}},
		{"taskAssign", "personal",
			map[string]string{"Actor": "<b>A</b>", "TaskName": "Sub", "CardLink": "<a>X</a>", "Context": ""},
			[]string{"вас на", "Sub"}},
		{"commentCreate", "personal",
			map[string]string{"Actor": "<b>A</b>", "CardLink": "<a>X</a>", "Context": ""},
			[]string{"прокоментув", "вашу"}},
		{"cardEdit", "personal",
			map[string]string{"Actor": "<b>A</b>", "CardLink": "<a>X</a>", "Context": ""},
			[]string{"вашу картку"}},
	}
	for _, tc := range cases {
		tpl := DefaultTemplates[tc.event][tc.kind]
		if tpl == "" {
			t.Errorf("%s/%s: template missing", tc.event, tc.kind)
			continue
		}
		got, err := RenderTemplate(tpl, tc.data)
		if err != nil {
			t.Errorf("%s/%s: render err %v", tc.event, tc.kind, err)
			continue
		}
		for _, s := range tc.mustHave {
			if !strings.Contains(got, s) {
				t.Errorf("%s/%s: rendered=%q missing %q", tc.event, tc.kind, got, s)
			}
		}
	}
}
