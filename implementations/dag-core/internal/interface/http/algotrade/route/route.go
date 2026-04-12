package route

import (
	"dag-observatory/dag-core/internal/interface/http/algotrade/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *handler.Handler) {
	e.POST("/algotrade/backtests\\:run", h.RunBacktest)
	e.GET("/algotrade/trades", h.ListTrades)
	e.GET("/algotrade/summary", h.GetSummary)
	e.POST("/algotrade/result-reflection:run", h.RunResultReflection)
}
