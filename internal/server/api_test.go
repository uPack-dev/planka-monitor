package server

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseAPITimeAcceptsRFC3339AndDateOnly(t *testing.T) {
	t.Parallel()

	rfc, err := parseAPITime("2026-06-09T08:30:00Z")
	if err != nil {
		t.Fatalf("parse RFC3339: %v", err)
	}
	if rfc.Format(time.RFC3339) != "2026-06-09T08:30:00Z" {
		t.Fatalf("unexpected RFC3339 value: %s", rfc.Format(time.RFC3339))
	}

	dateOnly, err := parseAPITime("2026-06-09")
	if err != nil {
		t.Fatalf("parse date-only: %v", err)
	}
	if dateOnly.Format("2006-01-02") != "2026-06-09" {
		t.Fatalf("unexpected date-only value: %s", dateOnly.Format("2006-01-02"))
	}
}

func TestParseRangeFallsBackWhenToBeforeFrom(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(
		"GET",
		"/api/v1/tasks?from=2026-06-09&to=2026-06-01",
		nil,
	)
	from, to := parseRange(request)

	if from.Format("2006-01-02") != "2026-06-09" {
		t.Fatalf("unexpected from: %s", from.Format("2006-01-02"))
	}
	if to.Format("2006-01-02") != "2026-08-08" {
		t.Fatalf("unexpected to fallback: %s", to.Format("2006-01-02"))
	}
}

func TestParseStatsRangeUsesExplicitBounds(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(
		"GET",
		"/api/v1/tasks?statsFrom=2026-06-13T00:00:00Z&statsTo=2026-06-13T23:59:59Z",
		nil,
	)
	from, to := parseStatsRange(request)

	if from.Format(time.RFC3339) != "2026-06-13T00:00:00Z" {
		t.Fatalf("unexpected stats from: %s", from.Format(time.RFC3339))
	}
	if to.Format(time.RFC3339) != "2026-06-13T23:59:59Z" {
		t.Fatalf("unexpected stats to: %s", to.Format(time.RFC3339))
	}
}

func TestSplitActionPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		prefix     string
		wantID     string
		wantAction string
		wantOK     bool
	}{
		{
			name:       "card comment action",
			path:       "/api/v1/cards/card-1/comments",
			prefix:     "/api/v1/cards/",
			wantID:     "card-1",
			wantAction: "comments",
			wantOK:     true,
		},
		{
			name:       "task complete action",
			path:       "/api/v1/tasks/task-1/complete",
			prefix:     "/api/v1/tasks/",
			wantID:     "task-1",
			wantAction: "complete",
			wantOK:     true,
		},
		{
			name:   "missing id",
			path:   "/api/v1/tasks/",
			prefix: "/api/v1/tasks/",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, action, ok := splitActionPath(tt.path, tt.prefix)
			if id != tt.wantID || action != tt.wantAction || ok != tt.wantOK {
				t.Fatalf(
					"splitActionPath() = (%q, %q, %t), want (%q, %q, %t)",
					id,
					action,
					ok,
					tt.wantID,
					tt.wantAction,
					tt.wantOK,
				)
			}
		})
	}
}
