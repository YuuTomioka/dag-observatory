package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"

	"github.com/labstack/echo/v4"
)

func TestRunBacktestRequiresPartition(t *testing.T) {
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
		"symbol_code":"USDJPY",
		"timeframe_code":"m1",
		"from":"2026-04-01T00:00:00Z",
		"to":"2026-04-01T01:00:00Z"
	}`
	req := httptest.NewRequest(http.MethodPost, "/algotrade/backtests:run", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RunBacktest(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "partition is required") {
		t.Fatalf("expected partition validation message, got: %s", rec.Body.String())
	}
}

func TestRunBacktestValidatesRange(t *testing.T) {
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
		"partition":"bt:USDJPY:m1",
		"symbol_code":"USDJPY",
		"timeframe_code":"m1",
		"from":"2026-04-01T01:00:00Z",
		"to":"2026-04-01T00:00:00Z"
	}`
	req := httptest.NewRequest(http.MethodPost, "/algotrade/backtests:run", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RunBacktest(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "from must be before to") {
		t.Fatalf("expected range validation message, got: %s", rec.Body.String())
	}
}

func TestRunBacktestEnqueuesWithExplicitContract(t *testing.T) {
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
		"partition":"bt:USDJPY:m1",
		"run_id":"run-bt-1",
		"symbol_code":"USDJPY",
		"timeframe_code":"m1",
		"from":"2026-04-01T00:00:00Z",
		"to":"2026-04-01T01:00:00Z",
		"source":"tsdb.timeframe_bars.v1",
		"timezone":"UTC",
		"gap_handling":"strict",
		"mode":"backtest"
	}`
	req := httptest.NewRequest(http.MethodPost, "/algotrade/backtests:run", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RunBacktest(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if enqueuer.last.Partition != "bt:USDJPY:m1" {
		t.Fatalf("expected partition bt:USDJPY:m1, got %q", enqueuer.last.Partition)
	}
	if enqueuer.last.EventID != "run-bt-1" {
		t.Fatalf("expected event id run-bt-1, got %q", enqueuer.last.EventID)
	}
	if !strings.Contains(rec.Body.String(), `"entrypoint":"backtest"`) {
		t.Fatalf("expected backtest entrypoint in response: %s", rec.Body.String())
	}
}
