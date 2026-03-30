package handler

import "github.com/labstack/echo/v4"

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

// Healthz
// @Summary Health check
// @Description Returns OK when server is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]any
// @Router /healthz [get]
func (h *Handler) Healthz(c echo.Context) error {
	return c.JSON(200, map[string]any{"ok": true})
}
