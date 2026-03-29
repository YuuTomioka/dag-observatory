package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

// UpsertTicksBulk
// @Summary Bulk upsert ticks
// @Description Upserts multiple ticks in a single transaction
// @Tags marketdata
// @Accept json
// @Produce json
// @Param request body dto.UpsertTicksBulkRequest true "Bulk tick upsert request"
// @Success 200 {object} dto.UpsertTicksBulkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /marketdata/ticks:upsert-bulk [post]
func (h *Handlers) UpsertTicksBulk(c echo.Context) error {
	if h.upsertTicks == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	}

	var req dto.UpsertTicksBulkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "invalid request",
			Status: "error",
		})
	}

	ticks := make([]marketdata.Tick, 0, len(req.Ticks))
	for _, item := range req.Ticks {
		parsedTime, err := time.Parse(time.RFC3339Nano, item.Time)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:  "time must be RFC3339Nano",
				Status: "error",
			})
		}
		ticks = append(ticks, marketdata.Tick{
			SymbolID: marketdata.SymbolID(item.SymbolID),
			Time:     marketdata.NewUTCTime(parsedTime),
			Bid:      marketdata.NewPriceFromRaw(item.BidRaw),
			Ask:      marketdata.NewPriceFromRaw(item.AskRaw),
		})
	}

	if err := h.upsertTicks.Execute(c.Request().Context(), ticks); err != nil {
		return writeMarketDataError(c, err)
	}
	return c.JSON(http.StatusOK, dto.UpsertTicksBulkResponse{Upserted: len(ticks)})
}

// GetLatestTickBySymbol
// @Summary Get latest tick
// @Description Returns latest tick by symbol id
// @Tags marketdata
// @Produce json
// @Param symbol_id query int true "Symbol ID"
// @Success 200 {object} dto.TickItem
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /marketdata/ticks/latest [get]
func (h *Handlers) GetLatestTickBySymbol(c echo.Context) error {
	if h.getLatestTickBySymbol == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	}
	symbolID, err := parseSymbolID(c.QueryParam("symbol_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}

	item, err := h.getLatestTickBySymbol.Execute(c.Request().Context(), symbolID)
	if err != nil {
		return writeMarketDataError(c, err)
	}
	return c.JSON(http.StatusOK, mapTickItem(item))
}

// ListTicksBySymbolAndRange
// @Summary List ticks by symbol and range
// @Description Returns ticks for symbol_id in [from, to)
// @Tags marketdata
// @Produce json
// @Param symbol_id query int true "Symbol ID"
// @Param from query string true "From (RFC3339Nano)"
// @Param to query string true "To (RFC3339Nano)"
// @Success 200 {object} dto.ListTicksResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /marketdata/ticks [get]
func (h *Handlers) ListTicksBySymbolAndRange(c echo.Context) error {
	if h.listTicksBySymbolAndRange == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	}
	symbolID, err := parseSymbolID(c.QueryParam("symbol_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}

	fromRaw := c.QueryParam("from")
	toRaw := c.QueryParam("to")
	fromParsed, err := time.Parse(time.RFC3339Nano, fromRaw)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "from must be RFC3339Nano",
			Status: "error",
		})
	}
	toParsed, err := time.Parse(time.RFC3339Nano, toRaw)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "to must be RFC3339Nano",
			Status: "error",
		})
	}

	items, err := h.listTicksBySymbolAndRange.Execute(
		c.Request().Context(),
		symbolID,
		marketdata.NewUTCTime(fromParsed),
		marketdata.NewUTCTime(toParsed),
	)
	if err != nil {
		return writeMarketDataError(c, err)
	}

	resp := dto.ListTicksResponse{Items: make([]dto.TickItem, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, mapTickItem(item))
	}
	return c.JSON(http.StatusOK, resp)
}

func parseSymbolID(raw string) (marketdata.SymbolID, error) {
	if raw == "" {
		return 0, fmt.Errorf("symbol_id is required")
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("symbol_id must be integer")
	}
	if n <= 0 {
		return 0, fmt.Errorf("symbol_id must be > 0")
	}
	return marketdata.SymbolID(n), nil
}
