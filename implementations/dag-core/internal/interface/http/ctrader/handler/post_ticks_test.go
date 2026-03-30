package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ctraderusecase "dag-observatory/dag-core/internal/application/ctrader/usecase"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"

	"github.com/labstack/echo/v4"
)

type fakeRepositories struct {
	symbols marketdatarepository.SymbolRepository
	ticks   marketdatarepository.TickRepository
}

func (r fakeRepositories) Symbols() marketdatarepository.SymbolRepository { return r.symbols }
func (r fakeRepositories) Ticks() marketdatarepository.TickRepository     { return r.ticks }

type fakeUnitOfWork struct {
	repos marketdatarepository.Repositories
}

func (u fakeUnitOfWork) Do(ctx context.Context, fn func(repos marketdatarepository.Repositories) error) error {
	return fn(u.repos)
}

func (u fakeUnitOfWork) DoReadOnly(ctx context.Context, fn func(repos marketdatarepository.Repositories) error) error {
	return fn(u.repos)
}

type fakeSymbolRepo struct {
	symbol marketdata.Symbol
	err    error
}

func (r fakeSymbolRepo) Create(ctx context.Context, input marketdatarepository.CreateSymbolInput) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r fakeSymbolRepo) GetByID(ctx context.Context, id marketdata.SymbolID) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r fakeSymbolRepo) GetByCode(ctx context.Context, code string) (marketdata.Symbol, error) {
	if r.err != nil {
		return marketdata.Symbol{}, r.err
	}
	return r.symbol, nil
}

func (r fakeSymbolRepo) List(ctx context.Context) ([]marketdata.Symbol, error) {
	return nil, nil
}

type fakeTickRepo struct {
	seen []marketdata.Tick
	err  error
}

func (r *fakeTickRepo) Insert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r *fakeTickRepo) Upsert(ctx context.Context, tick marketdata.Tick) error { return nil }

func (r *fakeTickRepo) BulkUpsert(ctx context.Context, ticks []marketdata.Tick) error {
	if r.err != nil {
		return r.err
	}
	r.seen = append(r.seen, ticks...)
	return nil
}

func (r *fakeTickRepo) GetLatestBySymbol(ctx context.Context, symbolID marketdata.SymbolID) (marketdata.Tick, error) {
	return marketdata.Tick{}, nil
}

func (r *fakeTickRepo) ListBySymbolAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	return nil, nil
}

func (r *fakeTickRepo) ListBySymbolsAndRange(
	ctx context.Context,
	symbolIDs []marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	return nil, nil
}

func (r *fakeTickRepo) ListByRange(ctx context.Context, from marketdata.UTCTime, to marketdata.UTCTime) ([]marketdata.Tick, error) {
	return nil, nil
}

func TestPostTicksGzipHTTP(t *testing.T) {
	tickRepo := &fakeTickRepo{}
	uc := ctraderusecase.PostTicksUsecase{
		UnitOfWork: fakeUnitOfWork{
			repos: fakeRepositories{
				symbols: fakeSymbolRepo{
					symbol: marketdata.Symbol{ID: 7, Code: "EURUSD", PriceScale: 5},
				},
				ticks: tickRepo,
			},
		},
	}

	h := NewPostTicksHandler(uc)
	e := echo.New()

	body := map[string]any{
		"symbol":      "EURUSD",
		"price_scale": 5,
		"request_id":  "EURUSD-2026-03-29",
		"day":         "2026-03-29",
		"batch_seq":   0,
		"ticks": []map[string]any{
			{"time": "2026-03-29T00:00:00.1234567Z", "bid": 108765, "ask": 108777},
			{"time": "2026-03-29T00:00:01.1234567Z", "bid": 108766, "ask": 108778},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/ctrader/ticks", gzipJSON(t, body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderContentEncoding, "gzip")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.PostTicks(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(tickRepo.seen) != 2 {
		t.Fatalf("expected 2 upserted ticks, got %d", len(tickRepo.seen))
	}
	if tickRepo.seen[0].SymbolID != 7 {
		t.Fatalf("expected symbol_id=7, got %d", tickRepo.seen[0].SymbolID)
	}
}

func TestPostTicksRejectsBadTime(t *testing.T) {
	h := NewPostTicksHandler(ctraderusecase.PostTicksUsecase{})
	e := echo.New()

	body := bytes.NewBufferString(`{"symbol":"EURUSD","price_scale":5,"batch_seq":0,"ticks":[{"time":"bad","bid":1,"ask":2}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ctrader/ticks", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = h.PostTicks(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPostTicksRejectsUnsupportedEncoding(t *testing.T) {
	h := NewPostTicksHandler(ctraderusecase.PostTicksUsecase{})
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/ctrader/ticks", bytes.NewBufferString(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderContentEncoding, "br")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = h.PostTicks(c)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPostTicksRejectsPriceScaleMismatch(t *testing.T) {
	tickRepo := &fakeTickRepo{}
	uc := ctraderusecase.PostTicksUsecase{
		UnitOfWork: fakeUnitOfWork{
			repos: fakeRepositories{
				symbols: fakeSymbolRepo{
					symbol: marketdata.Symbol{ID: 7, Code: "EURUSD", PriceScale: 3},
				},
				ticks: tickRepo,
			},
		},
	}

	h := NewPostTicksHandler(uc)
	e := echo.New()

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/ctrader/ticks",
		bytes.NewBufferString(`{"symbol":"EURUSD","price_scale":5,"batch_seq":0,"ticks":[{"time":"2026-03-29T00:00:00Z","bid":1,"ask":2}]}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = h.PostTicks(c)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(tickRepo.seen) != 0 {
		t.Fatalf("expected no tick writes on conflict")
	}
}

func gzipJSON(t *testing.T, payload any) *bytes.Reader {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	return bytes.NewReader(buf.Bytes())
}
