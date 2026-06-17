package events

import (
	"encoding/json"
	"html"
)

func produceTaskCreate(c *prodCtx) []Render {
	var item struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)
	return []Render{c.mk("taskCreate", map[string]string{"TaskName": html.EscapeString(item.Name)}, "")}
}

func produceTaskDelete(c *prodCtx) []Render {
	var item struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)
	return []Render{c.mk("taskDelete", map[string]string{"TaskName": html.EscapeString(item.Name)}, "")}
}

func produceTaskUpdate(c *prodCtx) []Render {
	var item struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		IsCompleted    bool   `json:"isCompleted"`
		AssigneeUserID string `json:"assigneeUserId"`
		UserID         string `json:"userId"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)
	prev := prevItem(c.p)

	curAssignee := item.AssigneeUserID
	if curAssignee == "" {
		curAssignee = item.UserID
	}

	// dmAssignee — для complete/edit/rename DM-имо поточного assignee.
	dmAssignee := func(r Render) Render {
		if curAssignee != "" {
			r.TargetUserID = curAssignee
			if u := userByID(c.p, curAssignee); u != nil {
				r.TargetUsername = u.Username
			}
		}
		return r
	}

	// 1) complete/uncomplete
	if pv, ok := changed(prev, "isCompleted", item.IsCompleted); ok {
		if v, ok2 := pv.(bool); ok2 && v != item.IsCompleted {
			key := "taskComplete"
			if !item.IsCompleted {
				key = "taskUncomplete"
			}
			return []Render{dmAssignee(c.mk(key, map[string]string{"TaskName": html.EscapeString(item.Name)}, ""))}
		}
	}

	// 2) rename
	if pv, ok := changed(prev, "name", item.Name); ok {
		if s, _ := pv.(string); s != item.Name {
			r := dmAssignee(c.mk("taskRename", nil, ""))
			r.Details = "було: <i>" + html.EscapeString(s) + "</i>\nстало: <i>" + html.EscapeString(item.Name) + "</i>"
			return []Render{r}
		}
	}

	// 3) assignee change (single-assignee schema: assigneeUserId / legacy userId)
	if r, ok := taskAssigneeChange(c, prev, item.Name, curAssignee); ok {
		return []Render{r}
	}

	// 4) description edit
	if _, ok := changed(prev, "description", item.Description); ok {
		r := dmAssignee(c.mk("taskEdit", map[string]string{
			"TaskName": html.EscapeString(item.Name),
		}, ""))
		details := "• опис задачі змінено"
		if item.Description != "" {
			details += "\n\n<b>Новий опис:</b>\n" + escapeHTMLBlockquote(item.Description, 1500)
		}
		r.Details = details
		return []Render{r}
	}
	return nil
}

// taskAssigneeChange — повертає Render для taskAssign/taskUnassign або (_, false), якщо assignee не змінився.
func taskAssigneeChange(c *prodCtx, prev map[string]any, taskName, newAssignee string) (Render, bool) {
	var prevAssignee string
	changedFlag := false
	for _, k := range []string{"assigneeUserId", "userId"} {
		if pv, ok := changed(prev, k, newAssignee); ok {
			changedFlag = true
			if s, ok2 := pv.(string); ok2 {
				prevAssignee = s
			}
			break
		}
	}
	if !changedFlag {
		return Render{}, false
	}
	if newAssignee != "" {
		u := userByID(c.p, newAssignee)
		name, uname := userDisplay(u)
		r := c.mk("taskAssign", map[string]string{
			"TaskName":   html.EscapeString(taskName),
			"TargetName": html.EscapeString(name),
		}, uname)
		r.TargetUserID = newAssignee
		return r, true
	}
	if prevAssignee != "" {
		u := userByID(c.p, prevAssignee)
		name, uname := userDisplay(u)
		r := c.mk("taskUnassign", map[string]string{
			"TaskName":   html.EscapeString(taskName),
			"TargetName": html.EscapeString(name),
		}, uname)
		r.TargetUserID = prevAssignee
		return r, true
	}
	return Render{}, false
}

func produceTaskMembership(c *prodCtx, event string) []Render {
	var item struct {
		TaskID string `json:"taskId"`
		UserID string `json:"userId"`
	}
	_ = json.Unmarshal(c.p.Data.Item, &item)

	taskName := ""
	for _, t := range c.p.Data.Included.Tasks {
		if t.ID == item.TaskID {
			taskName = t.Name
			break
		}
	}

	u := userByID(c.p, item.UserID)
	name, uname := userDisplay(u)

	key := "taskAssign"
	if event == "taskMembershipDelete" {
		key = "taskUnassign"
	}
	r := c.mk(key, map[string]string{
		"TaskName":   html.EscapeString(taskName),
		"TargetName": html.EscapeString(name),
	}, uname)
	r.TargetUserID = item.UserID
	return []Render{r}
}
