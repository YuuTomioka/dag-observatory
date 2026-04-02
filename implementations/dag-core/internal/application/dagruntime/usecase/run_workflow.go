package usecase

import (
	"context"
	"errors"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/observability/ctxprop"

	"github.com/google/uuid"
)

const (
	defaultSymbol                = "USDJPY"
	defaultMode                  = "normal"
	defaultTaskID                = "heavy_calc"
	defaultTaskName              = "heavy_calc"
	defaultAttempt               = 1
	defaultEventType             = "task.requested"
	defaultDriverDestinationName = "dagruntime-driver"
)

type RunWorkflow struct {
	Driver        *driver.Driver
	Enqueuer      port.EventEnqueuer
	Clock         port.Clock
	ProducerTopic string
}

func (u *RunWorkflow) Execute(ctx context.Context, req RunWorkflowRequest) (RunWorkflowResult, error) {
	result, event, partition := buildRunEvent(req, u.now())
	ctx = ctxprop.WithRunID(ctx, result.RunID)
	if err := u.Handle(ctx, partition, event); err != nil {
		return result, err
	}
	result.EnqueueMode = u.IsEnqueueMode()
	result.MessagingDestination = u.messagingDestination(result.EnqueueMode)
	return result, nil
}

func (u *RunWorkflow) Handle(ctx context.Context, partition state.Partition, event events.Event) error {
	if u.Enqueuer != nil {
		payload, err := toPayloadEnvelopes(event.Payload)
		if err != nil {
			return err
		}
		event.Payload = payload
		event.Partition = partition
		return u.Enqueuer.Enqueue(ctx, event)
	}
	if u.Driver == nil {
		return errors.New("dagruntime: driver not configured")
	}
	inputs, err := toInputMap(event.Payload)
	if err != nil {
		return err
	}
	event.Payload = inputs
	event.Partition = partition
	stream := make(chan events.Event, 1)
	stream <- event
	close(stream)
	return u.Driver.Run(ctx, stream)
}

func (u *RunWorkflow) IsEnqueueMode() bool {
	return u != nil && u.Enqueuer != nil
}

func (u *RunWorkflow) messagingDestination(enqueueMode bool) string {
	if enqueueMode {
		if u != nil && u.ProducerTopic != "" {
			return u.ProducerTopic
		}
		return "unknown"
	}
	return defaultDriverDestinationName
}

func (u *RunWorkflow) now() time.Time {
	if u != nil && u.Clock != nil {
		return u.Clock.Now()
	}
	return time.Now()
}

func buildRunEvent(req RunWorkflowRequest, now time.Time) (RunWorkflowResult, events.Event, state.Partition) {
	runID := req.RunID
	if runID == "" {
		runID = uuid.NewString()
	}
	symbol := req.Symbol
	if symbol == "" && req.Marketdata != nil && req.Marketdata.SymbolCode != "" {
		symbol = req.Marketdata.SymbolCode
	}
	if symbol == "" && req.Marketdata == nil {
		symbol = defaultSymbol
	}
	mode := req.Mode
	if mode == "" {
		mode = defaultMode
	}

	result := RunWorkflowResult{
		RunID:      runID,
		Symbol:     symbol,
		Mode:       mode,
		Marketdata: cloneMarketdataInput(req.Marketdata),
		TaskID:     defaultTaskID,
		TaskName:   defaultTaskName,
		Attempt:    defaultAttempt,
	}

	payload := map[string]any{
		"run_id":    runID,
		"task_id":   defaultTaskID,
		"attempt":   defaultAttempt,
		"task_name": defaultTaskName,
		"input":     map[string]any{},
		// Keep top-level keys for direct execution compatibility.
		"mode": mode,
	}
	payload["input"].(map[string]any)["mode"] = mode
	if symbol != "" {
		payload["input"].(map[string]any)["symbol"] = symbol
		payload["symbol"] = symbol
	}
	if len(req.Bars) > 0 {
		payload["input"].(map[string]any)["bars"] = req.Bars
		payload["market_bars"] = req.Bars
	}
	if len(req.OHLCVBars) > 0 {
		payload["input"].(map[string]any)["ohlcv_bars"] = req.OHLCVBars
		payload["market_ohlcv_bars"] = req.OHLCVBars
	}
	if req.SpreadBps > 0 {
		payload["input"].(map[string]any)["market.spread_bps"] = req.SpreadBps
		payload["market_spread_bps"] = req.SpreadBps
	}
	if req.AccountBalance > 0 {
		payload["input"].(map[string]any)["account.balance"] = req.AccountBalance
		payload["account_balance"] = req.AccountBalance
	}
	if req.Marketdata != nil {
		if req.Marketdata.SymbolID > 0 {
			payload["input"].(map[string]any)["marketdata.symbol_id"] = req.Marketdata.SymbolID
			payload["marketdata.symbol_id"] = req.Marketdata.SymbolID
		}
		if req.Marketdata.SymbolCode != "" {
			payload["input"].(map[string]any)["marketdata.symbol_code"] = req.Marketdata.SymbolCode
			payload["marketdata.symbol_code"] = req.Marketdata.SymbolCode
		}
		if req.Marketdata.TimeframeCode != "" {
			payload["input"].(map[string]any)["marketdata.timeframe_code"] = req.Marketdata.TimeframeCode
			payload["marketdata.timeframe_code"] = req.Marketdata.TimeframeCode
		}
		if !req.Marketdata.From.IsZero() {
			payload["input"].(map[string]any)["marketdata.from"] = req.Marketdata.From
			payload["marketdata.from"] = req.Marketdata.From
		}
		if !req.Marketdata.To.IsZero() {
			payload["input"].(map[string]any)["marketdata.to"] = req.Marketdata.To
			payload["marketdata.to"] = req.Marketdata.To
		}
	}

	partition := state.Partition(runID)
	event := events.Event{
		EventID:   runID,
		EventTime: now,
		Partition: partition,
		Type:      defaultEventType,
		Payload:   payload,
	}

	return result, event, partition
}

func cloneMarketdataInput(input *MarketdataRunInput) *MarketdataRunInput {
	if input == nil {
		return nil
	}
	cloned := *input
	return &cloned
}
