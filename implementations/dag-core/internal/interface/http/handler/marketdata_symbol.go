package handler

import (
	"net/http"

	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

// CreateSymbol
// @Summary Create symbol
// @Description Creates a market symbol master record
// @Tags marketdata
// @Accept json
// @Produce json
// @Param request body dto.CreateSymbolRequest true "Create symbol request"
// @Success 201 {object} dto.SymbolItem
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /marketdata/symbols [post]
func (h *Handlers) CreateSymbol(c echo.Context) error {
	if h.createSymbol == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	}

	var req dto.CreateSymbolRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "invalid request",
			Status: "error",
		})
	}

	created, err := h.createSymbol.Execute(c.Request().Context(), marketdatausecase.CreateSymbolRequest{
		Code:        req.Code,
		Base:        req.Base,
		Quote:       req.Quote,
		PriceScale:  req.PriceScale,
		TickSizeRaw: req.TickSizeRaw,
		PipSizeRaw:  req.PipSizeRaw,
	})
	if err != nil {
		return writeMarketDataError(c, err)
	}
	return c.JSON(http.StatusCreated, mapSymbolItem(created))
}

// GetSymbolByCode
// @Summary Get symbol by code
// @Description Returns symbol metadata by symbol code
// @Tags marketdata
// @Produce json
// @Param code path string true "Symbol code"
// @Success 200 {object} dto.SymbolItem
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /marketdata/symbols/{code} [get]
func (h *Handlers) GetSymbolByCode(c echo.Context) error {
	if h.getSymbolByCode == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	}

	item, err := h.getSymbolByCode.Execute(c.Request().Context(), c.Param("code"))
	if err != nil {
		return writeMarketDataError(c, err)
	}
	return c.JSON(http.StatusOK, mapSymbolItem(item))
}

// ListSymbols
// @Summary List symbols
// @Description Returns market symbol master records
// @Tags marketdata
// @Produce json
// @Success 200 {object} dto.ListSymbolsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /marketdata/symbols [get]
func (h *Handlers) ListSymbols(c echo.Context) error {
	if h.listSymbols == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	}

	items, err := h.listSymbols.Execute(c.Request().Context())
	if err != nil {
		return writeMarketDataError(c, err)
	}

	resp := dto.ListSymbolsResponse{Items: make([]dto.SymbolItem, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, mapSymbolItem(item))
	}
	return c.JSON(http.StatusOK, resp)
}
