package events

import "testing"

func TestPersonalCategory(t *testing.T) {
	tests := map[string]string{
		"cardMembershipCreate": CategoryAssignments,
		"taskAssign":           CategoryAssignments,
		"commentCreate":        CategoryComments,
		"cardEdit":             CategoryChanges,
		"taskRename":           CategoryChanges,
		"taskComplete":         CategoryDone,
		"taskUncomplete":       CategoryDone,
		"projectCreate":        "",
	}
	for eventKey, want := range tests {
		if got := PersonalCategory(eventKey); got != want {
			t.Fatalf("PersonalCategory(%q)=%q, want %q", eventKey, got, want)
		}
	}
}
