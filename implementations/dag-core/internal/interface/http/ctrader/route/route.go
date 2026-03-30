package route

import (
	ctraderhandler "dag-observatory/dag-core/internal/interface/http/ctrader/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *ctraderhandler.PostTicksHandler) {
	e.POST("/v1/ctrader/ticks", h.PostTicks)
}
