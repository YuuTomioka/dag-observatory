package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"

	"github.com/labstack/echo/v4"
)

func TestGetEquitySeriesReturnsPoints(t *testing.T) {
	t.Parallel()

	store := stateinfra.NewMemoryStore()
	txn := store.BeginTxn(domainstate.Partition("strategy-1:USDJPY"))
	domainstate.StageWrite(txn, algotrade.StateClosedTrades, algotrade.ClosedTradesState{
		Items: []algotrade.ClosedTrade{
			{
				TradeID:  "trade:1",
				NetPnL:   5,
				ExitTime: marketdata.MustParseUTCTime("2026-03-01T00:10:00Z"),
			},
			{
				TradeID:  "trade:2",
				NetPnL:   -2,
				ExitTime: marketdata.MustParseUTCTime("2026-03-01T00:20:00Z"),
			},
		},
	})
	if err := txn.Commit(); err != nil {
		t.Fatalf("commit state: %v", err)
	}

	h := New(Dependencies{
		StateStore:      store,
		GetEquitySeries: &dagruntimeusecase.GetEquitySeries{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/algotrade/equity?partition=strategy-1:USDJPY", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetEquitySeries(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		MaxDrawdown float64 `json:"max_drawdown"`
		TotalNetPnL float64 `json:"total_net_pnl"`
		Points      []struct {
			TradeID string  `json:"trade_id"`
			Equity  float64 `json:"equity"`
		} `json:"points"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(resp.Points))
	}
	if resp.MaxDrawdown != 2 || resp.TotalNetPnL != 3 {
		t.Fatalf("unexpected series summary: %#v", resp)
	}
}
