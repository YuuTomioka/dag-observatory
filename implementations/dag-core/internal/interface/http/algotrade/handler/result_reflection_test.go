package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"

	"github.com/labstack/echo/v4"
)

type recordingEnqueuer struct {
	last events.Event
}

func (r *recordingEnqueuer) Enqueue(ctx context.Context, event events.Event) error {
	_ = ctx
	r.last = event
	return nil
}

func TestRunResultReflectionRequiresPartition(t *testing.T) {
	t.Parallel()

	enqueuer := &recordingEnqueuer{}
	h := New(Dependencies{
		StateStore: stateinfra.NewMemoryStore(),
		RunWorkflow: &dagruntimeusecase.RunWorkflow{
			Enqueuer: enqueuer,
		},
	})

	e := echo.New()
	body := `{"ohlcv_bars":[{"opentime":"2026-04-01T00:00:00Z","closetime":"2026-04-01T00:01:00Z","open":1000,"high":1010,"low":990,"close":1005,"volume":10}]}`
	req := httptest.NewRequest(http.MethodPost, "/algotrade/result-reflection:run", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RunResultReflection(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRunResultReflectionEnqueuesWithExplicitPartition(t *testing.T) {
	t.Parallel()

	enqueuer := &recordingEnqueuer{}
	h := New(Dependencies{
		StateStore: stateinfra.NewMemoryStore(),
		RunWorkflow: &dagruntimeusecase.RunWorkflow{
			Enqueuer: enqueuer,
		},
	})

	e := echo.New()
	body := `{
		"partition":"strategy-1:USDJPY",
		"run_id":"run-reflect-1",
		"symbol":"USDJPY",
		"ohlcv_bars":[{"opentime":"2026-04-01T00:00:00Z","closetime":"2026-04-01T00:01:00Z","open":1000,"high":1010,"low":990,"close":1005,"volume":10}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/algotrade/result-reflection:run", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RunResultReflection(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if enqueuer.last.Partition != "strategy-1:USDJPY" {
		t.Fatalf("expected partition strategy-1:USDJPY, got %q", enqueuer.last.Partition)
	}
	if enqueuer.last.EventID != "run-reflect-1" {
		t.Fatalf("expected event id run-reflect-1, got %q", enqueuer.last.EventID)
	}
}
