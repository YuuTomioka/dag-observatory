package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"

	"github.com/labstack/echo/v4"
)

type fakeBacktestSymbolGetter struct{}

func (fakeBacktestSymbolGetter) Execute(ctx context.Context, code string) (marketdata.Symbol, error) {
	_ = ctx
	_ = code
	return marketdata.Symbol{ID: 10, Code: "USDJPY"}, nil
}

type fakeBacktestBarLister struct{}

func (fakeBacktestBarLister) Execute(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode timeframe.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]timeframe.TimeframeBar, error) {
	_ = ctx
	_ = symbolID
	_ = timeframeCode
	_ = from
	_ = to
	return []timeframe.TimeframeBar{
		{
			SymbolID:      10,
			TimeframeCode: "m1",
			OHLCV: ohlc.OHLCV{
				Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
				Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
				Open:      marketdata.NewPriceFromRaw(1000),
				High:      marketdata.NewPriceFromRaw(1002),
				Low:       marketdata.NewPriceFromRaw(999),
				Close:     marketdata.NewPriceFromRaw(1001),
				Volume:    10,
			},
		},
		{
			SymbolID:      10,
			TimeframeCode: "m1",
			OHLCV: ohlc.OHLCV{
				Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
				Closetime: marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
				Open:      marketdata.NewPriceFromRaw(1001),
				High:      marketdata.NewPriceFromRaw(1003),
				Low:       marketdata.NewPriceFromRaw(1000),
				Close:     marketdata.NewPriceFromRaw(1002),
				Volume:    11,
			},
		},
		{
			SymbolID:      10,
			TimeframeCode: "m1",
			OHLCV: ohlc.OHLCV{
				Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
				Closetime: marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
				Open:      marketdata.NewPriceFromRaw(1002),
				High:      marketdata.NewPriceFromRaw(1004),
				Low:       marketdata.NewPriceFromRaw(1001),
				Close:     marketdata.NewPriceFromRaw(1003),
				Volume:    12,
			},
		},
	}, nil
}

type fakeBacktestRunner struct{}

func (fakeBacktestRunner) Execute(ctx context.Context, req dagruntimeusecase.RunWorkflowRequest) (dagruntimeusecase.RunWorkflowResult, error) {
	_ = ctx
	return dagruntimeusecase.RunWorkflowResult{
		RunID:  req.RunID,
		Symbol: req.Symbol,
		Mode:   req.Mode,
	}, nil
}

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

func TestRunBacktestUsesPeriodLoopWhenConfigured(t *testing.T) {
	t.Parallel()

	h := New(Dependencies{
		StateStore: stateinfra.NewMemoryStore(),
		RunBacktest: &dagruntimeusecase.RunBacktest{
			RunWorkflow:       fakeBacktestRunner{},
			GetSymbolByCode:   fakeBacktestSymbolGetter{},
			ListTimeframeBars: fakeBacktestBarLister{},
			WindowSizeBars:    2,
		},
	})

	e := echo.New()
	body := `{
		"partition":"bt:USDJPY:m1",
		"run_id":"run-bt-loop",
		"symbol_code":"USDJPY",
		"timeframe_code":"m1",
		"from":"2026-04-01T00:00:00Z",
		"to":"2026-04-01T00:03:00Z"
	}`
	req := httptest.NewRequest(http.MethodPost, "/algotrade/backtests:run", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RunBacktest(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"cycle_count":3`) {
		t.Fatalf("expected cycle_count=3 in response: %s", rec.Body.String())
	}
}
