package events

const (
	CategoryAssignments = "assignments"
	CategoryComments    = "comments"
	CategoryChanges     = "changes"
	CategoryDone        = "done"
)

func PersonalCategory(eventKey string) string {
	switch eventKey {
	case "cardMembershipCreate", "cardMembershipDelete", "taskAssign", "taskUnassign":
		return CategoryAssignments
	case "commentCreate":
		return CategoryComments
	case "cardEdit", "cardRename", "cardMove", "cardLabelCreate", "cardLabelDelete", "attachmentCreate", "taskRename", "taskEdit":
		return CategoryChanges
	case "taskComplete", "taskUncomplete":
		return CategoryDone
	default:
		return ""
	}
}
