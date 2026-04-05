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

func TestListTradesReturnsClosedTradeSnapshot(t *testing.T) {
	t.Parallel()

	store := stateinfra.NewMemoryStore()
	txn := store.BeginTxn(domainstate.Partition("strategy-1:USDJPY"))
	domainstate.StageWrite(txn, algotrade.StateClosedTrades, algotrade.ClosedTradesState{
		Items: []algotrade.ClosedTrade{
			{
				TradeID:         "older",
				IntentID:        "intent-1",
				PositionID:      "position-1",
				Symbol:          "USDJPY",
				Side:            algotrade.PositionSideLong,
				Size:            1.25,
				EntryTime:       marketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
				ExitTime:        marketdata.MustParseUTCTime("2026-03-01T00:10:00Z"),
				EntryPrice:      marketdata.Price(1000),
				ExitPrice:       marketdata.Price(1010),
				NetPnL:          10.5,
				ExitReason:      "tp",
				WorkflowName:    "breakout",
				WorkflowVersion: "v1",
				ParameterSetID:  "params-a",
			},
			{
				TradeID:         "newer",
				IntentID:        "intent-2",
				PositionID:      "position-2",
				Symbol:          "USDJPY",
				Side:            algotrade.PositionSideLong,
				Size:            1.0,
				EntryTime:       marketdata.MustParseUTCTime("2026-03-01T01:00:00Z"),
				ExitTime:        marketdata.MustParseUTCTime("2026-03-01T01:10:00Z"),
				EntryPrice:      marketdata.Price(1020),
				ExitPrice:       marketdata.Price(1030),
				NetPnL:          8.0,
				ExitReason:      "close",
				WorkflowName:    "breakout",
				WorkflowVersion: "v1",
				ParameterSetID:  "params-a",
			},
		},
	})
	if err := txn.Commit(); err != nil {
		t.Fatalf("commit state: %v", err)
	}

	h := New(Dependencies{
		StateStore:       store,
		ListTradeResults: &dagruntimeusecase.ListTradeResults{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/algotrade/trades?partition=strategy-1:USDJPY", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListTrades(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Partition string `json:"partition"`
		Items     []struct {
			TradeID string `json:"trade_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Partition != "strategy-1:USDJPY" {
		t.Fatalf("unexpected partition: %s", resp.Partition)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].TradeID != "newer" {
		t.Fatalf("expected newest trade first, got %s", resp.Items[0].TradeID)
	}
}

func TestListTradesRequiresPartition(t *testing.T) {
	t.Parallel()

	h := New(Dependencies{
		StateStore:       stateinfra.NewMemoryStore(),
		ListTradeResults: &dagruntimeusecase.ListTradeResults{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/algotrade/trades", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListTrades(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
