package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dag-observatory/dag-core/internal/domain/marketdata"

	"github.com/google/uuid"
)

const (
	defaultBacktestWindowSizeBars = 200
	defaultBacktestMode           = "backtest"
)

var ErrBacktestInput = errors.New("dagruntime backtest: invalid input")

type BacktestSymbolGetter interface {
	Execute(ctx context.Context, code string) (marketdata.Symbol, error)
}

type BacktestTimeframeBarLister interface {
	Execute(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		timeframeCode marketdata.TimeframeCode,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) ([]marketdata.TimeframeBar, error)
}

type BacktestWorkflowRunner interface {
	Execute(ctx context.Context, req RunWorkflowRequest) (RunWorkflowResult, error)
}

type RunBacktest struct {
	RunWorkflow       BacktestWorkflowRunner
	GetSymbolByCode   BacktestSymbolGetter
	ListTimeframeBars BacktestTimeframeBarLister
	WindowSizeBars    int
}

type RunBacktestRequest struct {
	RunID          string
	Partition      string
	SymbolCode     string
	TimeframeCode  string
	From           marketdata.UTCTime
	To             marketdata.UTCTime
	Mode           string
	SpreadBps      float64
	FeeBps         float64
	SlippageBps    float64
	MinLot         float64
	AccountBalance float64
}

type RunBacktestResult struct {
	RunID          string
	Partition      string
	Symbol         string
	Mode           string
	TimeframeCode  string
	From           marketdata.UTCTime
	To             marketdata.UTCTime
	CycleCount     int
	WindowSizeBars int
	EnqueueMode    bool
}

func (u *RunBacktest) Execute(ctx context.Context, req RunBacktestRequest) (RunBacktestResult, error) {
	if u == nil || u.RunWorkflow == nil || u.GetSymbolByCode == nil || u.ListTimeframeBars == nil {
		return RunBacktestResult{}, errors.New("dagruntime backtest: not configured")
	}

	req = normalizeBacktestRequest(req)
	if err := validateBacktestRequest(req); err != nil {
		return RunBacktestResult{}, err
	}

	symbol, err := u.GetSymbolByCode.Execute(ctx, req.SymbolCode)
	if err != nil {
		return RunBacktestResult{}, err
	}

	bars, err := u.ListTimeframeBars.Execute(
		ctx,
		symbol.ID,
		marketdata.TimeframeCode(req.TimeframeCode),
		req.From,
		req.To,
	)
	if err != nil {
		return RunBacktestResult{}, err
	}
	if len(bars) == 0 {
		return RunBacktestResult{}, fmt.Errorf("%w: no timeframe bars found in range", ErrBacktestInput)
	}

	windowSizeBars := u.WindowSizeBars
	if windowSizeBars <= 0 {
		windowSizeBars = defaultBacktestWindowSizeBars
	}
	if windowSizeBars > len(bars) {
		windowSizeBars = len(bars)
	}

	baseRunID := req.RunID
	if baseRunID == "" {
		baseRunID = uuid.NewString()
	}

	cycleCount := 0
	lastEnqueueMode := false
	for i := 0; i < len(bars); i++ {
		end := i + 1
		start := 0
		if end > windowSizeBars {
			start = end - windowSizeBars
		}
		cycleBars := make([]marketdata.OHLCV, 0, end-start)
		for _, bar := range bars[start:end] {
			cycleBars = append(cycleBars, bar.OHLCV)
		}
		cycleRunID := fmt.Sprintf("%s:%06d", baseRunID, end)
		result, err := u.RunWorkflow.Execute(ctx, RunWorkflowRequest{
			RunID:          cycleRunID,
			Partition:      req.Partition,
			Symbol:         symbol.Code,
			Mode:           req.Mode,
			OHLCVBars:      cycleBars,
			SpreadBps:      req.SpreadBps,
			FeeBps:         req.FeeBps,
			SlippageBps:    req.SlippageBps,
			MinLot:         req.MinLot,
			AccountBalance: req.AccountBalance,
			Marketdata: &MarketdataRunInput{
				SymbolID:      int64(symbol.ID),
				SymbolCode:    symbol.Code,
				TimeframeCode: req.TimeframeCode,
				From:          cycleBars[0].Opentime,
				To:            cycleBars[len(cycleBars)-1].Closetime,
			},
		})
		if err != nil {
			return RunBacktestResult{}, err
		}
		lastEnqueueMode = result.EnqueueMode
		cycleCount++
	}

	return RunBacktestResult{
		RunID:          baseRunID,
		Partition:      req.Partition,
		Symbol:         symbol.Code,
		Mode:           req.Mode,
		TimeframeCode:  req.TimeframeCode,
		From:           req.From,
		To:             req.To,
		CycleCount:     cycleCount,
		WindowSizeBars: windowSizeBars,
		EnqueueMode:    lastEnqueueMode,
	}, nil
}

func normalizeBacktestRequest(req RunBacktestRequest) RunBacktestRequest {
	req.SymbolCode = strings.ToUpper(strings.TrimSpace(req.SymbolCode))
	req.TimeframeCode = strings.ToLower(strings.TrimSpace(req.TimeframeCode))
	req.Mode = strings.TrimSpace(req.Mode)
	if req.Mode == "" {
		req.Mode = defaultBacktestMode
	}
	return req
}

func validateBacktestRequest(req RunBacktestRequest) error {
	if req.Partition == "" {
		return fmt.Errorf("%w: partition is required", ErrBacktestInput)
	}
	if req.SymbolCode == "" {
		return fmt.Errorf("%w: symbol_code is required", ErrBacktestInput)
	}
	if req.TimeframeCode == "" {
		return fmt.Errorf("%w: timeframe_code is required", ErrBacktestInput)
	}
	if req.From.IsZero() || req.To.IsZero() {
		return fmt.Errorf("%w: from/to are required", ErrBacktestInput)
	}
	if !req.From.Before(req.To) {
		return fmt.Errorf("%w: from must be before to", ErrBacktestInput)
	}
	return nil
}
