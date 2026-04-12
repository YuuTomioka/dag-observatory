package handler

import "github.com/labstack/echo/v4"

type Handler struct {
	replaySkippedLines int64
	replayErrors       int64
}

type Dependencies struct {
	ReplaySkippedLines int64
	ReplayErrors       int64
}

func New(d Dependencies) *Handler {
	return &Handler{
		replaySkippedLines: d.ReplaySkippedLines,
		replayErrors:       d.ReplayErrors,
	}
}

// Healthz
// @Summary Health check
// @Description Returns OK when server is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]any
// @Router /healthz [get]
func (h *Handler) Healthz(c echo.Context) error {
	return c.JSON(200, map[string]any{
		"ok": true,
		"observation_replay": map[string]any{
			"skipped_lines": h.replaySkippedLines,
			"errors":        h.replayErrors,
		},
	})
}
