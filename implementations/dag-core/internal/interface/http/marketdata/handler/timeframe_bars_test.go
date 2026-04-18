package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/marketdata"

	"github.com/labstack/echo/v4"
)

type marketdataRecordingEnqueuer struct {
	last events.Event
}

func (r *marketdataRecordingEnqueuer) Enqueue(ctx context.Context, event events.Event) error {
	_ = ctx
	r.last = event
	return nil
}

type fakeMarketdataRepositories struct {
	symbols       apprepository.SymbolRepository
	ticks         apprepository.TickRepository
	timeframeBars apprepository.TimeframeBarRepository
}

func (r fakeMarketdataRepositories) Symbols() apprepository.SymbolRepository {
	return r.symbols
}

func (r fakeMarketdataRepositories) Ticks() apprepository.TickRepository {
	return r.ticks
}

func (r fakeMarketdataRepositories) TimeframeBars() apprepository.TimeframeBarRepository {
	return r.timeframeBars
}

func (r fakeMarketdataRepositories) PhaseBars() apprepository.PhaseBarRepository {
	return nil
}

type fakeMarketdataUnitOfWork struct {
	repos apprepository.Repositories
}

func (u fakeMarketdataUnitOfWork) Do(ctx context.Context, fn func(repos apprepository.Repositories) error) error {
	return fn(u.repos)
}

func (u fakeMarketdataUnitOfWork) DoReadOnly(ctx context.Context, fn func(repos apprepository.Repositories) error) error {
	return fn(u.repos)
}

type fakeMarketdataSymbolRepo struct {
	symbol marketdata.Symbol
}

func (r fakeMarketdataSymbolRepo) Create(ctx context.Context, input apprepository.CreateSymbolInput) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r fakeMarketdataSymbolRepo) GetByID(ctx context.Context, id marketdata.SymbolID) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r fakeMarketdataSymbolRepo) GetByCode(ctx context.Context, code string) (marketdata.Symbol, error) {
	return r.symbol, nil
}

func (r fakeMarketdataSymbolRepo) List(ctx context.Context) ([]marketdata.Symbol, error) {
	return nil, nil
}

type fakeMarketdataTickRepo struct {
	ticks []marketdata.Tick
}

func (r fakeMarketdataTickRepo) Insert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r fakeMarketdataTickRepo) Upsert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r fakeMarketdataTickRepo) BulkUpsert(ctx context.Context, ticks []marketdata.Tick) error {
	return nil
}
func (r fakeMarketdataTickRepo) GetLatestBySymbol(ctx context.Context, symbolID marketdata.SymbolID) (marketdata.Tick, error) {
	return marketdata.Tick{}, nil
}
func (r fakeMarketdataTickRepo) ListBySymbolAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	out := make([]marketdata.Tick, 0)
	for _, tick := range r.ticks {
		if tick.SymbolID != symbolID {
			continue
		}
		if tick.Time.Compare(from) >= 0 && tick.Time.Compare(to) < 0 {
			out = append(out, tick)
		}
	}
	return out, nil
}
func (r fakeMarketdataTickRepo) ListBySymbolsAndRange(
	ctx context.Context,
	symbolIDs []marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	return nil, nil
}
func (r fakeMarketdataTickRepo) ListByRange(ctx context.Context, from marketdata.UTCTime, to marketdata.UTCTime) ([]marketdata.Tick, error) {
	return nil, nil
}

type fakeMarketdataTimeframeBarRepo struct {
	bars []marketdata.TimeframeBar
}

func (r *fakeMarketdataTimeframeBarRepo) BulkUpsert(ctx context.Context, bars []marketdata.TimeframeBar) error {
	r.bars = append(r.bars, bars...)
	return nil
}

func (r *fakeMarketdataTimeframeBarRepo) DeleteBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) error {
	return nil
}

func (r *fakeMarketdataTimeframeBarRepo) GetLatestBySymbolAndTimeframe(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
) (marketdata.TimeframeBar, error) {
	return marketdata.TimeframeBar{}, nil
}

func (r *fakeMarketdataTimeframeBarRepo) ListBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.TimeframeBar, error) {
	return nil, nil
}

func TestBackfillTimeframeBarsHTTP(t *testing.T) {
	t.Parallel()

	barRepo := &fakeMarketdataTimeframeBarRepo{}
	h := New(Dependencies{
		BackfillTimeframeBars: &marketdatausecase.BackfillTimeframeBars{
			UnitOfWork: fakeMarketdataUnitOfWork{
				repos: fakeMarketdataRepositories{
					symbols: fakeMarketdataSymbolRepo{
						symbol: marketdata.Symbol{ID: 7, Code: "USDJPY"},
					},
					ticks: fakeMarketdataTickRepo{
						ticks: []marketdata.Tick{
							{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:10Z"), Bid: 1000, Ask: 1002},
							{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:20Z"), Bid: 1004, Ask: 1006},
							{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:01:05Z"), Bid: 1010, Ask: 1014},
						},
					},
					timeframeBars: barRepo,
				},
			},
		},
	})

	e := echo.New()
	req := httptest.NewRequest(
		http.MethodPost,
		"/marketdata/timeframe-bars:backfill",
		bytes.NewBufferString(`{"symbol_code":"USDJPY","timeframe_code":"M1","from":"2026-03-01T00:00:00Z","to":"2026-03-01T00:02:00Z","chunk_size_bars":1}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.BackfillTimeframeBars(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["entrypoint"] != "direct" {
		t.Fatalf("expected entrypoint direct, got %v", resp["entrypoint"])
	}
	if resp["status"] != "completed" {
		t.Fatalf("expected status completed, got %v", resp["status"])
	}
	if len(barRepo.bars) != 2 {
		t.Fatalf("expected 2 upserted bars, got %d", len(barRepo.bars))
	}
}

func TestBackfillTimeframeBarsWorkflowHTTP(t *testing.T) {
	t.Parallel()

	enqueuer := &marketdataRecordingEnqueuer{}
	h := New(Dependencies{
		RunWorkflow: &dagruntimeusecase.RunWorkflow{Enqueuer: enqueuer},
	})

	e := echo.New()
	req := httptest.NewRequest(
		http.MethodPost,
		"/marketdata/timeframe-bars:backfill-workflow",
		bytes.NewBufferString(`{"symbol_code":"USDJPY","timeframe_code":"M1","from":"2026-03-01T00:00:00Z","to":"2026-03-01T02:00:00Z"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.BackfillTimeframeBarsWorkflow(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["entrypoint"] != "workflow" {
		t.Fatalf("expected entrypoint workflow, got %v", resp["entrypoint"])
	}
	if resp["status"] != "enqueued" {
		t.Fatalf("expected status enqueued, got %v", resp["status"])
	}
	if resp["symbol"] != "USDJPY" {
		t.Fatalf("expected symbol USDJPY, got %v", resp["symbol"])
	}
	marketdataResp, ok := resp["marketdata"].(map[string]any)
	if !ok {
		t.Fatalf("expected marketdata response object, got %T", resp["marketdata"])
	}
	if marketdataResp["timeframe_code"] != "M1" {
		t.Fatalf("expected timeframe_code M1, got %v", marketdataResp["timeframe_code"])
	}
}
