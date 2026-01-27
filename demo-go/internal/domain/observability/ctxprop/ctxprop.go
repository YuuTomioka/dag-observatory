package ctxprop

import "context"

type key string

const runIDKey key = "dag.run_id"

func WithRunID(ctx context.Context, runID string) context.Context {
	if runID == "" {
		return ctx
	}
	return context.WithValue(ctx, runIDKey, runID)
}

func RunID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(runIDKey).(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}
