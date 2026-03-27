package view

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

func TestPhaseFromEventType(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		want      string
	}{
		{name: "requested", eventType: "task.requested", want: "in_progress"},
		{name: "started", eventType: "task.started", want: "in_progress"},
		{name: "running", eventType: "task.running", want: "in_progress"},
		{name: "completed", eventType: "task.completed", want: "succeeded"},
		{name: "succeeded", eventType: "task.succeeded", want: "succeeded"},
		{name: "failed", eventType: "task.failed", want: "failed"},
		{name: "error", eventType: "task.error", want: "failed"},
		{name: "cancelled", eventType: "task.cancelled", want: "cancelled"},
		{name: "canceled", eventType: "task.canceled", want: "cancelled"},
		{name: "unknown", eventType: "task.whatever", want: "unknown"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := phaseFromEventType(tc.eventType); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestRunViewFromEvent(t *testing.T) {
	event := events.Event{
		EventID: "run-1",
		Type:    "task.requested",
		Payload: map[string]any{
			"symbol": "USDJPY",
			"mode":   "normal",
		},
	}

	view := FromEvent(event)

	if view.RunID != "run-1" {
		t.Fatalf("expected run id run-1, got %q", view.RunID)
	}
	if view.Status != "task.requested" {
		t.Fatalf("expected status task.requested, got %q", view.Status)
	}
	if view.Phase != "in_progress" {
		t.Fatalf("expected phase in_progress, got %q", view.Phase)
	}
	if view.Symbol != "USDJPY" {
		t.Fatalf("expected symbol USDJPY, got %q", view.Symbol)
	}
	if view.Mode != "normal" {
		t.Fatalf("expected mode normal, got %q", view.Mode)
	}
}
