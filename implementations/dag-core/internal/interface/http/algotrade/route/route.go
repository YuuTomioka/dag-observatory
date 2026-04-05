package route

import (
	"dag-observatory/dag-core/internal/interface/http/algotrade/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *handler.Handler) {
	e.GET("/algotrade/trades", h.ListTrades)
}
