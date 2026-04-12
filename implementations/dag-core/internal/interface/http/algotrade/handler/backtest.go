package handler

import (
	"errors"
	"net/http"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/interface/http/algotrade/response"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

const (
	backtestModeDefault       = "backtest"
	backtestSourceDefault     = "tsdb.timeframe_bars.v1"
	backtestTimezoneDefault   = "UTC"
	backtestGapHandlingStrict = "strict"
	backtestGapHandlingSkip   = "skip"
)

type backtestRunRequest struct {
	RunID           string             `json:"run_id,omitempty"`
	Partition       string             `json:"partition"`
	SymbolCode      string             `json:"symbol_code"`
	TimeframeCode   string             `json:"timeframe_code"`
	From            marketdata.UTCTime `json:"from"`
	To              marketdata.UTCTime `json:"to"`
	Source          string             `json:"source,omitempty"`
	Timezone        string             `json:"timezone,omitempty"`
	GapHandling     string             `json:"gap_handling,omitempty"`
	Mode            string             `json:"mode,omitempty"`
	SpreadBps       float64            `json:"spread_bps,omitempty"`
	FeeBps          float64            `json:"fee_bps,omitempty"`
	SlippageBps     float64            `json:"slippage_bps,omitempty"`
	MinLot          float64            `json:"min_lot,omitempty"`
	WorkflowName    string             `json:"workflow_name,omitempty"`
	WorkflowVersion string             `json:"workflow_version,omitempty"`
	ParameterSetID  string             `json:"parameter_set_id,omitempty"`
	AccountBalance  float64            `json:"account_balance,omitempty"`
}

// RunBacktest
// @Summary Run real-data backtest workflow
// @Description Triggers the configured runtime workflow as a backtest entrypoint over an explicit historical range and returns run correlation metadata.
// @Tags algotrade
// @Accept json
// @Produce json
// @Param request body backtestRunRequest true "Backtest run request"
// @Success 200 {object} response.BacktestRunResponse
// @Success 202 {object} response.BacktestRunResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /algotrade/backtests:run [post]
func (h *Handler) RunBacktest(c echo.Context) error {
	canUsePeriodLoop := h.runBacktest != nil &&
		h.runBacktest.RunWorkflow != nil &&
		h.runBacktest.GetSymbolByCode != nil &&
		h.runBacktest.ListTimeframeBars != nil

	if h.runWF == nil && !canUsePeriodLoop {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "backtest API not configured",
			Status: "error",
		})
	}

	var req backtestRunRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "invalid request",
			Status: "error",
		})
	}

	source := req.Source
	if source == "" {
		source = backtestSourceDefault
	}
	timezone := req.Timezone
	if timezone == "" {
		timezone = backtestTimezoneDefault
	}
	gapHandling := req.GapHandling
	if gapHandling == "" {
		gapHandling = backtestGapHandlingStrict
	}
	mode := req.Mode
	if mode == "" {
		mode = backtestModeDefault
	}

	if err := validateBacktestRunRequest(req, timezone, gapHandling); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}

	resp := response.BacktestRunResponse{
		Status:          "completed",
		Entrypoint:      "backtest",
		Partition:       req.Partition,
		Symbol:          req.SymbolCode,
		Mode:            mode,
		CycleCount:      1,
		WindowSize:      1,
		WorkflowName:    req.WorkflowName,
		WorkflowVersion: req.WorkflowVersion,
		ParameterSetID:  req.ParameterSetID,
		Marketdata: response.BacktestMarketdataInput{
			SymbolCode:    req.SymbolCode,
			TimeframeCode: req.TimeframeCode,
			From:          req.From.String(),
			To:            req.To.String(),
			Source:        source,
			Timezone:      timezone,
			GapHandling:   gapHandling,
		},
	}

	if canUsePeriodLoop {
		result, err := h.runBacktest.Execute(c.Request().Context(), dagruntimeusecase.RunBacktestRequest{
			RunID:           req.RunID,
			Partition:       req.Partition,
			SymbolCode:      req.SymbolCode,
			TimeframeCode:   req.TimeframeCode,
			From:            req.From,
			To:              req.To,
			Mode:            mode,
			SpreadBps:       req.SpreadBps,
			FeeBps:          req.FeeBps,
			SlippageBps:     req.SlippageBps,
			MinLot:          req.MinLot,
			WorkflowName:    req.WorkflowName,
			WorkflowVersion: req.WorkflowVersion,
			ParameterSetID:  req.ParameterSetID,
			AccountBalance:  req.AccountBalance,
		})
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, dagruntimeusecase.ErrBacktestInput) {
				status = http.StatusBadRequest
			}
			return c.JSON(status, dto.ErrorResponse{
				Error:  err.Error(),
				Status: "failed",
				Symbol: result.Symbol,
				RunID:  result.RunID,
			})
		}
		resp.RunID = result.RunID
		resp.Partition = result.Partition
		resp.Symbol = result.Symbol
		resp.Mode = result.Mode
		resp.CycleCount = result.CycleCount
		resp.WindowSize = result.WindowSizeBars
		if result.EnqueueMode {
			resp.Status = "enqueued"
			return c.JSON(http.StatusAccepted, resp)
		}
		return c.JSON(http.StatusOK, resp)
	}

	result, err := h.runWF.Execute(c.Request().Context(), dagruntimeusecase.RunWorkflowRequest{
		RunID:           req.RunID,
		Partition:       req.Partition,
		Symbol:          req.SymbolCode,
		Mode:            mode,
		SpreadBps:       req.SpreadBps,
		FeeBps:          req.FeeBps,
		SlippageBps:     req.SlippageBps,
		MinLot:          req.MinLot,
		WorkflowName:    req.WorkflowName,
		WorkflowVersion: req.WorkflowVersion,
		ParameterSetID:  req.ParameterSetID,
		AccountBalance:  req.AccountBalance,
		Marketdata: &dagruntimeusecase.MarketdataRunInput{
			SymbolCode:    req.SymbolCode,
			TimeframeCode: req.TimeframeCode,
			From:          req.From,
			To:            req.To,
		},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "failed",
			Symbol: result.Symbol,
			RunID:  result.RunID,
		})
	}
	resp.RunID = result.RunID
	resp.Symbol = result.Symbol
	if result.EnqueueMode {
		resp.Status = "enqueued"
		return c.JSON(http.StatusAccepted, resp)
	}
	return c.JSON(http.StatusOK, resp)
}

func validateBacktestRunRequest(req backtestRunRequest, timezone, gapHandling string) error {
	if req.Partition == "" {
		return errors.New("partition is required")
	}
	if req.SymbolCode == "" {
		return errors.New("symbol_code is required")
	}
	if req.TimeframeCode == "" {
		return errors.New("timeframe_code is required")
	}
	if req.From.IsZero() {
		return errors.New("from is required")
	}
	if req.To.IsZero() {
		return errors.New("to is required")
	}
	if req.From.Compare(req.To) >= 0 {
		return errors.New("from must be before to")
	}
	if timezone != backtestTimezoneDefault {
		return errors.New("timezone must be UTC")
	}
	switch gapHandling {
	case backtestGapHandlingStrict, backtestGapHandlingSkip:
	default:
		return errors.New("gap_handling must be strict or skip")
	}
	if req.FeeBps < 0 {
		return errors.New("fee_bps must be >= 0")
	}
	if req.SlippageBps < 0 {
		return errors.New("slippage_bps must be >= 0")
	}
	if req.MinLot < 0 {
		return errors.New("min_lot must be >= 0")
	}
	return nil
}
