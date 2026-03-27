package view

import (
	"strings"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

// RunView is a presentation-friendly model derived from domain events.
type RunView struct {
	RunID  string `json:"run_id"`
	Status string `json:"status"`
	Symbol string `json:"symbol"`
	Mode   string `json:"mode"`
	Phase  string `json:"phase"`
}

func FromEvent(event events.Event) RunView {
	view := RunView{
		RunID:  event.EventID,
		Status: event.Type,
		Phase:  phaseFromEventType(event.Type),
	}
	payload, ok := event.Payload.(map[string]any)
	if !ok {
		return view
	}
	if raw, ok := payload["symbol"].(string); ok {
		view.Symbol = raw
	}
	if raw, ok := payload["mode"].(string); ok {
		view.Mode = raw
	}
	if rawInput, ok := payload["input"].(map[string]any); ok {
		if raw, ok := rawInput["symbol"].(string); ok {
			view.Symbol = raw
		}
		if raw, ok := rawInput["mode"].(string); ok {
			view.Mode = raw
		}
	}
	return view
}

func phaseFromEventType(eventType string) string {
	parts := strings.Split(eventType, ".")
	if len(parts) == 0 {
		return ""
	}
	switch parts[len(parts)-1] {
	case "requested", "started", "running":
		return "in_progress"
	case "completed", "succeeded":
		return "succeeded"
	case "failed", "error":
		return "failed"
	case "cancelled", "canceled":
		return "cancelled"
	default:
		return "unknown"
	}
}
