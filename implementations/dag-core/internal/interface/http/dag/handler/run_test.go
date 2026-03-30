package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"

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

func TestDagRunHTTP(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	uc := &usecase.RunWorkflow{Enqueuer: enqueuer}

	h := New(Dependencies{
		AppLog:      applog.New("info", "stdout", nil),
		RunWorkflow: uc,
	})

	e := echo.New()
	body, _ := json.Marshal(map[string]any{"symbol": "USDJPY", "mode": "normal"})
	req := httptest.NewRequest(http.MethodPost, "/dag/run", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.DagRun(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestDagRunHTTPWithMarketdataInput(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	uc := &usecase.RunWorkflow{Enqueuer: enqueuer}

	h := New(Dependencies{
		AppLog:      applog.New("info", "stdout", nil),
		RunWorkflow: uc,
	})

	e := echo.New()
	body, _ := json.Marshal(map[string]any{
		"marketdata": map[string]any{
			"symbol_code":    "USDJPY",
			"timeframe_code": "M1",
			"from":           "2026-03-01T00:00:00Z",
			"to":             "2026-03-01T01:00:00Z",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/dag/run", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.DagRun(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}
