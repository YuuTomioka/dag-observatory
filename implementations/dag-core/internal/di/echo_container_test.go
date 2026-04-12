package di

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRegisterSwaggerUIRoute(t *testing.T) {
	t.Run("registers route when enabled", func(t *testing.T) {
		e := echo.New()
		registerSwaggerUIRoute(e, true, "/swagger/*")
		if !hasRoute(e, "/swagger/*") {
			t.Fatalf("expected /swagger/* route to be registered")
		}
	})

	t.Run("does not register route when disabled", func(t *testing.T) {
		e := echo.New()
		registerSwaggerUIRoute(e, false, "/swagger/*")
		if hasRoute(e, "/swagger/*") {
			t.Fatalf("expected /swagger/* route not to be registered")
		}
	})

	t.Run("does not register route when route is empty", func(t *testing.T) {
		e := echo.New()
		registerSwaggerUIRoute(e, true, "")
		if len(e.Routes()) != 0 {
			t.Fatalf("expected no routes to be registered, got %d", len(e.Routes()))
		}
	})
}

func hasRoute(e *echo.Echo, path string) bool {
	for _, route := range e.Routes() {
		if route.Path == path {
			return true
		}
	}
	return false
}
