package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestHealthzIncludesObservationReplayCounters(t *testing.T) {
	t.Parallel()

	h := New(Dependencies{
		ReplaySkippedLines: 3,
		ReplayErrors:       2,
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Healthz(c); err != nil {
		t.Fatalf("healthz handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode healthz response: %v", err)
	}
	observation, _ := payload["observation_replay"].(map[string]any)
	if observation == nil {
		t.Fatalf("expected observation_replay section: %#v", payload)
	}
	if observation["skipped_lines"] != float64(3) || observation["errors"] != float64(2) {
		t.Fatalf("unexpected counters: %#v", observation)
	}
}
