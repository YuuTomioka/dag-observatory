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

func TestGetSummaryReturnsProjectedSummary(t *testing.T) {
	t.Parallel()

	store := stateinfra.NewMemoryStore()
	txn := store.BeginTxn(domainstate.Partition("strategy-1:USDJPY"))
	domainstate.StageWrite(txn, algotrade.StateStrategySummary, algotrade.StrategySummaryState{
		Summary: algotrade.StrategySummary{
			StrategyID:      "breakout",
			WorkflowName:    "dagruntime.breakout_long_result_reflection_minimal",
			WorkflowVersion: "v1",
			TradeCount:      3,
			WinCount:        2,
			LossCount:       1,
			WinRate:         2.0 / 3.0,
			TotalNetPnL:     12.0,
			UpdatedAt:       marketdata.MustParseUTCTime("2026-04-12T01:00:00Z"),
		},
	})
	if err := txn.Commit(); err != nil {
		t.Fatalf("commit state: %v", err)
	}

	h := New(Dependencies{
		StateStore:         store,
		GetStrategySummary: &dagruntimeusecase.GetStrategySummary{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/algotrade/summary?partition=strategy-1:USDJPY", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetSummary(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Partition string `json:"partition"`
		Summary   struct {
			StrategyID  string  `json:"strategy_id"`
			TradeCount  int     `json:"trade_count"`
			TotalNetPnL float64 `json:"total_net_pnl"`
			UpdatedAt   string  `json:"updated_at"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Partition != "strategy-1:USDJPY" {
		t.Fatalf("unexpected partition: %s", resp.Partition)
	}
	if resp.Summary.StrategyID != "breakout" || resp.Summary.TradeCount != 3 || resp.Summary.TotalNetPnL != 12 {
		t.Fatalf("unexpected summary response: %#v", resp.Summary)
	}
	if resp.Summary.UpdatedAt == "" {
		t.Fatalf("expected updated_at, got %#v", resp.Summary)
	}
}

func TestGetSummaryRequiresPartition(t *testing.T) {
	t.Parallel()

	h := New(Dependencies{
		StateStore:         stateinfra.NewMemoryStore(),
		GetStrategySummary: &dagruntimeusecase.GetStrategySummary{},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/algotrade/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetSummary(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
