package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRunsUIReturnsHTML(t *testing.T) {
	h := New(Dependencies{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/runs/ui", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RunsUI(c); err != nil {
		t.Fatalf("runs ui handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderContentType); !strings.Contains(got, echo.MIMETextHTML) {
		t.Fatalf("expected html content type, got %q", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "DAG Run Inspector") {
		t.Fatalf("expected page title in response body")
	}
	if !strings.Contains(body, "/runs/") {
		t.Fatalf("expected ui to reference run APIs")
	}
}
