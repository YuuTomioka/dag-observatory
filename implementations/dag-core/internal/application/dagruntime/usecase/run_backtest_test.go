package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

type fakeBacktestSymbolGetter struct {
	out marketdata.Symbol
	err error
}

func (f fakeBacktestSymbolGetter) Execute(ctx context.Context, code string) (marketdata.Symbol, error) {
	_ = ctx
	_ = code
	return f.out, f.err
}

type fakeBacktestTimeframeBarLister struct {
	out []marketdata.TimeframeBar
	err error
}

func (f fakeBacktestTimeframeBarLister) Execute(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.TimeframeBar, error) {
	_ = ctx
	_ = symbolID
	_ = timeframeCode
	_ = from
	_ = to
	return f.out, f.err
}

type fakeBacktestWorkflowRunner struct {
	calls        []RunWorkflowRequest
	errAt        int
	syntheticPnL map[string]float64
}

func (f *fakeBacktestWorkflowRunner) Execute(ctx context.Context, req RunWorkflowRequest) (RunWorkflowResult, error) {
	_ = ctx
	f.calls = append(f.calls, req)
	if f.syntheticPnL != nil {
		basePnL := 10.0
		costPenalty := req.SpreadBps + req.FeeBps + req.SlippageBps
		sizePenalty := 0.0
		if req.MinLot > 0 {
			sizePenalty = req.MinLot * 0.1
		}
		f.syntheticPnL[req.RunID] = basePnL - costPenalty - sizePenalty
	}
	if f.errAt > 0 && len(f.calls) == f.errAt {
		return RunWorkflowResult{}, fmt.Errorf("run failed at call %d", f.errAt)
	}
	return RunWorkflowResult{
		RunID:       req.RunID,
		Symbol:      req.Symbol,
		Mode:        req.Mode,
		EnqueueMode: false,
	}, nil
}

func TestRunBacktestExecutionAssumptionsAffectSyntheticPnL(t *testing.T) {
	t.Parallel()

	runner := &fakeBacktestWorkflowRunner{
		syntheticPnL: map[string]float64{},
	}
	u := &RunBacktest{
		RunWorkflow: runner,
		GetSymbolByCode: fakeBacktestSymbolGetter{
			out: marketdata.Symbol{ID: 100, Code: "USDJPY"},
		},
		ListTimeframeBars: fakeBacktestTimeframeBarLister{
			out: []marketdata.TimeframeBar{
				{
					SymbolID:      100,
					TimeframeCode: "m1",
					OHLCV: marketdata.OHLCV{
						Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
						Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
						Open:      marketdata.NewPriceFromRaw(1000),
						High:      marketdata.NewPriceFromRaw(1002),
						Low:       marketdata.NewPriceFromRaw(999),
						Close:     marketdata.NewPriceFromRaw(1001),
						Volume:    10,
					},
				},
			},
		},
		WindowSizeBars: 1,
	}

	_, err := u.Execute(context.Background(), RunBacktestRequest{
		RunID:         "bt-low-cost",
		Partition:     "bt:USDJPY:m1",
		SymbolCode:    "USDJPY",
		TimeframeCode: "m1",
		From:          marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
		To:            marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
		Mode:          "backtest",
		SpreadBps:     0.5,
		FeeBps:        0.2,
		SlippageBps:   0.3,
		MinLot:        0.01,
	})
	if err != nil {
		t.Fatalf("low-cost run failed: %v", err)
	}
	_, err = u.Execute(context.Background(), RunBacktestRequest{
		RunID:         "bt-high-cost",
		Partition:     "bt:USDJPY:m1",
		SymbolCode:    "USDJPY",
		TimeframeCode: "m1",
		From:          marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
		To:            marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
		Mode:          "backtest",
		SpreadBps:     2.0,
		FeeBps:        1.0,
		SlippageBps:   1.2,
		MinLot:        0.10,
	})
	if err != nil {
		t.Fatalf("high-cost run failed: %v", err)
	}

	low := runner.syntheticPnL["bt-low-cost:000001"]
	high := runner.syntheticPnL["bt-high-cost:000001"]
	if !(high < low) {
		t.Fatalf("expected higher costs to reduce synthetic pnl: low=%f high=%f", low, high)
	}
}

func TestRunBacktestValidatesRange(t *testing.T) {
	t.Parallel()

	u := &RunBacktest{
		RunWorkflow:       &fakeBacktestWorkflowRunner{},
		GetSymbolByCode:   fakeBacktestSymbolGetter{},
		ListTimeframeBars: fakeBacktestTimeframeBarLister{},
	}
	_, err := u.Execute(context.Background(), RunBacktestRequest{
		Partition:     "bt:USDJPY:m1",
		SymbolCode:    "USDJPY",
		TimeframeCode: "m1",
		From:          marketdata.MustParseUTCTime("2026-04-01T01:00:00Z"),
		To:            marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
	})
	if err == nil {
		t.Fatal("expected range validation error")
	}
	if !errors.Is(err, ErrBacktestInput) {
		t.Fatalf("expected ErrBacktestInput, got %v", err)
	}
}

func TestRunBacktestRequiresBars(t *testing.T) {
	t.Parallel()

	u := &RunBacktest{
		RunWorkflow: &fakeBacktestWorkflowRunner{},
		GetSymbolByCode: fakeBacktestSymbolGetter{
			out: marketdata.Symbol{ID: 1, Code: "USDJPY"},
		},
		ListTimeframeBars: fakeBacktestTimeframeBarLister{
			out: nil,
		},
	}
	_, err := u.Execute(context.Background(), RunBacktestRequest{
		Partition:     "bt:USDJPY:m1",
		SymbolCode:    "USDJPY",
		TimeframeCode: "m1",
		From:          marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
		To:            marketdata.MustParseUTCTime("2026-04-01T03:00:00Z"),
	})
	if err == nil {
		t.Fatal("expected no bars error")
	}
	if !errors.Is(err, ErrBacktestInput) {
		t.Fatalf("expected ErrBacktestInput, got %v", err)
	}
}

func TestRunBacktestRunsPeriodLoopWithSlidingWindow(t *testing.T) {
	t.Parallel()

	runner := &fakeBacktestWorkflowRunner{}
	u := &RunBacktest{
		RunWorkflow: runner,
		GetSymbolByCode: fakeBacktestSymbolGetter{
			out: marketdata.Symbol{ID: 42, Code: "USDJPY"},
		},
		ListTimeframeBars: fakeBacktestTimeframeBarLister{
			out: []marketdata.TimeframeBar{
				{
					SymbolID:      42,
					TimeframeCode: "m1",
					OHLCV: marketdata.OHLCV{
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
					SymbolID:      42,
					TimeframeCode: "m1",
					OHLCV: marketdata.OHLCV{
						Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
						Closetime: marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
						Open:      marketdata.NewPriceFromRaw(1001),
						High:      marketdata.NewPriceFromRaw(1004),
						Low:       marketdata.NewPriceFromRaw(1000),
						Close:     marketdata.NewPriceFromRaw(1003),
						Volume:    12,
					},
				},
				{
					SymbolID:      42,
					TimeframeCode: "m1",
					OHLCV: marketdata.OHLCV{
						Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
						Closetime: marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
						Open:      marketdata.NewPriceFromRaw(1003),
						High:      marketdata.NewPriceFromRaw(1005),
						Low:       marketdata.NewPriceFromRaw(1002),
						Close:     marketdata.NewPriceFromRaw(1004),
						Volume:    14,
					},
				},
			},
		},
		WindowSizeBars: 2,
	}

	result, err := u.Execute(context.Background(), RunBacktestRequest{
		RunID:          "bt-run-1",
		Partition:      "bt:USDJPY:m1",
		SymbolCode:     "usdjpy",
		TimeframeCode:  "m1",
		From:           marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
		To:             marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
		Mode:           "backtest",
		SpreadBps:      2.5,
		AccountBalance: 100000,
	})
	if err != nil {
		t.Fatalf("execute run backtest: %v", err)
	}
	if result.CycleCount != 3 {
		t.Fatalf("expected 3 cycles, got %d", result.CycleCount)
	}
	if result.WindowSizeBars != 2 {
		t.Fatalf("expected window_size_bars=2, got %d", result.WindowSizeBars)
	}
	if result.RunID != "bt-run-1" {
		t.Fatalf("expected run id bt-run-1, got %q", result.RunID)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("expected 3 workflow calls, got %d", len(runner.calls))
	}
	if runner.calls[0].RunID != "bt-run-1:000001" || len(runner.calls[0].OHLCVBars) != 1 {
		t.Fatalf("unexpected first call: run_id=%q bars=%d", runner.calls[0].RunID, len(runner.calls[0].OHLCVBars))
	}
	if runner.calls[1].RunID != "bt-run-1:000002" || len(runner.calls[1].OHLCVBars) != 2 {
		t.Fatalf("unexpected second call: run_id=%q bars=%d", runner.calls[1].RunID, len(runner.calls[1].OHLCVBars))
	}
	if runner.calls[2].RunID != "bt-run-1:000003" || len(runner.calls[2].OHLCVBars) != 2 {
		t.Fatalf("unexpected third call: run_id=%q bars=%d", runner.calls[2].RunID, len(runner.calls[2].OHLCVBars))
	}
	if runner.calls[2].Partition != "bt:USDJPY:m1" {
		t.Fatalf("expected partition bt:USDJPY:m1, got %q", runner.calls[2].Partition)
	}
}
