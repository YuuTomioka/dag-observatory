package events

import "context"

type contextKey string

const eventKey contextKey = "dagruntime.event"

func WithEvent(ctx context.Context, event Event) context.Context {
	return context.WithValue(ctx, eventKey, event)
}

func EventFromContext(ctx context.Context) (Event, bool) {
	value, ok := ctx.Value(eventKey).(Event)
	return value, ok
}

func EventIDFromContext(ctx context.Context) (string, bool) {
	event, ok := EventFromContext(ctx)
	if !ok || event.EventID == "" {
		return "", false
	}
	return event.EventID, true
}
