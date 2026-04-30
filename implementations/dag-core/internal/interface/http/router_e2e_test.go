package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dagruntimeport "dag-observatory/dag-core/internal/application/dagruntime/port"
	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatarepo "dag-observatory/dag-core/internal/application/marketdata/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"

	"github.com/labstack/echo/v4"
)

type e2eMarketdataRepositories struct {
	symbols       marketdatarepo.SymbolRepository
	ticks         marketdatarepo.TickRepository
	timeframeBars marketdatarepo.TimeframeBarRepository
}

func (r e2eMarketdataRepositories) Symbols() marketdatarepo.SymbolRepository {
	return r.symbols
}

func (r e2eMarketdataRepositories) Ticks() marketdatarepo.TickRepository {
	return r.ticks
}

func (r e2eMarketdataRepositories) TimeframeBars() marketdatarepo.TimeframeBarRepository {
	return r.timeframeBars
}

type e2eMarketdataUnitOfWork struct {
	repos marketdatarepo.Repositories
}

func (u e2eMarketdataUnitOfWork) Do(ctx context.Context, fn func(repos marketdatarepo.Repositories) error) error {
	return fn(u.repos)
}

func (u e2eMarketdataUnitOfWork) DoReadOnly(ctx context.Context, fn func(repos marketdatarepo.Repositories) error) error {
	return fn(u.repos)
}

type e2eSymbolRepo struct {
	symbol marketdata.Symbol
}

func (r e2eSymbolRepo) Create(ctx context.Context, input marketdatarepo.CreateSymbolInput) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r e2eSymbolRepo) GetByID(ctx context.Context, id marketdata.SymbolID) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r e2eSymbolRepo) GetByCode(ctx context.Context, code string) (marketdata.Symbol, error) {
	return r.symbol, nil
}

func (r e2eSymbolRepo) List(ctx context.Context) ([]marketdata.Symbol, error) {
	return nil, nil
}

type e2eTickRepo struct {
	ticks []marketdata.Tick
}

func (r e2eTickRepo) Insert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r e2eTickRepo) Upsert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r e2eTickRepo) BulkUpsert(ctx context.Context, ticks []marketdata.Tick) error {
	return nil
}
func (r e2eTickRepo) GetLatestBySymbol(ctx context.Context, symbolID marketdata.SymbolID) (marketdata.Tick, error) {
	return marketdata.Tick{}, nil
}
func (r e2eTickRepo) ListBySymbolAndRange(
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
func (r e2eTickRepo) ListBySymbolsAndRange(
	ctx context.Context,
	symbolIDs []marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	return nil, nil
}
func (r e2eTickRepo) ListByRange(ctx context.Context, from marketdata.UTCTime, to marketdata.UTCTime) ([]marketdata.Tick, error) {
	return nil, nil
}

type e2eTimeframeBarRepo struct {
	bars []timeframe.TimeframeBar
}

func (r *e2eTimeframeBarRepo) BulkUpsert(ctx context.Context, bars []timeframe.TimeframeBar) error {
	r.bars = append(r.bars, bars...)
	return nil
}

func (r *e2eTimeframeBarRepo) DeleteBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode timeframe.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) error {
	return nil
}

func (r *e2eTimeframeBarRepo) GetLatestBySymbolAndTimeframe(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode timeframe.TimeframeCode,
) (timeframe.TimeframeBar, error) {
	return timeframe.TimeframeBar{}, nil
}

func (r *e2eTimeframeBarRepo) ListBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode timeframe.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]timeframe.TimeframeBar, error) {
	return nil, nil
}

type e2eRecordingEnqueuer struct {
	last events.Event
}

var _ dagruntimeport.EventEnqueuer = (*e2eRecordingEnqueuer)(nil)

func (r *e2eRecordingEnqueuer) Enqueue(ctx context.Context, event events.Event) error {
	_ = ctx
	r.last = event
	return nil
}

func TestRegisterRoutes_MarketdataBackfillDirectE2E(t *testing.T) {
	t.Parallel()

	barRepo := &e2eTimeframeBarRepo{}
	uow := e2eMarketdataUnitOfWork{
		repos: e2eMarketdataRepositories{
			symbols: e2eSymbolRepo{
				symbol: marketdata.Symbol{ID: 7, Code: "USDJPY"},
			},
			ticks: e2eTickRepo{
				ticks: []marketdata.Tick{
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:10Z"), Bid: 1000, Ask: 1002},
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:20Z"), Bid: 1004, Ask: 1006},
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:01:05Z"), Bid: 1010, Ask: 1014},
				},
			},
			timeframeBars: barRepo,
		},
	}

	e := echo.New()
	RegisterRoutes(e, Dependencies{
		AppLog: applog.New("info", "stdout", nil),
		BackfillTimeframeBars: &marketdatausecase.BackfillTimeframeBars{
			UnitOfWork: uow,
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/marketdata/timeframe-bars:backfill",
		bytes.NewBufferString(`{"symbol_code":"USDJPY","timeframe_code":"M1","from":"2026-03-01T00:00:00Z","to":"2026-03-01T00:02:00Z","chunk_size_bars":1}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

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
	if got := len(barRepo.bars); got != 2 {
		t.Fatalf("expected 2 upserted bars, got %d", got)
	}
}

func TestRegisterRoutes_MarketdataBackfillWorkflowE2E(t *testing.T) {
	t.Parallel()

	enqueuer := &e2eRecordingEnqueuer{}

	e := echo.New()
	RegisterRoutes(e, Dependencies{
		AppLog: applog.New("info", "stdout", nil),
		RunWorkflow: &dagruntimeusecase.RunWorkflow{
			Enqueuer: enqueuer,
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/marketdata/timeframe-bars:backfill-workflow",
		bytes.NewBufferString(`{"symbol_code":"USDJPY","timeframe_code":"M1","from":"2026-03-01T00:00:00Z","to":"2026-03-01T02:00:00Z"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

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
	if enqueuer.last.Type != "task.requested" {
		t.Fatalf("expected enqueued task.requested event, got %q", enqueuer.last.Type)
	}
}
