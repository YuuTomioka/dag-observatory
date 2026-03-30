package route

import (
	"dag-observatory/dag-core/internal/interface/http/marketdata/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *handler.Handlers) {
	e.POST("/marketdata/symbols", h.CreateSymbol)
	e.GET("/marketdata/symbols", h.ListSymbols)
	e.GET("/marketdata/symbols/:code", h.GetSymbolByCode)
	e.POST("/marketdata/ticks\\:upsert-bulk", h.UpsertTicksBulk)
	e.GET("/marketdata/ticks/latest", h.GetLatestTickBySymbol)
	e.GET("/marketdata/ticks", h.ListTicksBySymbolAndRange)
	e.POST("/marketdata/timeframe-bars\\:backfill", h.BackfillTimeframeBars)
	e.POST("/marketdata/timeframe-bars\\:backfill-workflow", h.BackfillTimeframeBarsWorkflow)
}
