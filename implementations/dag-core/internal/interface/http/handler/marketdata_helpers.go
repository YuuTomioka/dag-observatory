package handler

import (
	"errors"
	"net/http"

	"dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

func mapSymbolItem(item marketdata.Symbol) dto.SymbolItem {
	return dto.SymbolItem{
		ID:          int64(item.ID),
		Code:        item.Code,
		Base:        item.Base,
		Quote:       item.Quote,
		PriceScale:  item.PriceScale,
		TickSizeRaw: item.TickSizeRaw,
		PipSizeRaw:  item.PipSizeRaw,
	}
}

func mapTickItem(item marketdata.Tick) dto.TickItem {
	return dto.TickItem{
		SymbolID: int64(item.SymbolID),
		Time:     item.Time.String(),
		BidRaw:   item.Bid.Raw(),
		AskRaw:   item.Ask.Raw(),
	}
}

func writeMarketDataError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidArgument):
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	case errors.Is(err, repository.ErrNotFound):
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:  "not found",
			Status: "error",
		})
	case errors.Is(err, repository.ErrConflict):
		return c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	case errors.Is(err, repository.ErrNotConfigured):
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	default:
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
}
