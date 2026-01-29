package otelecho

import (
	"fmt"
	"time"

	"dag-observatory/demo-go/internal/infrastructure/observability/metrics"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func Middleware(tr trace.Tracer, inst *metrics.Instruments) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			spanName := fmt.Sprintf("%s %s", c.Request().Method, c.Path())
			start := time.Now()
			ctx, span := tr.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(c.Request().Method),
					semconv.HTTPRouteKey.String(c.Path()),
					semconv.HTTPTargetKey.String(c.Request().URL.RequestURI()),
					semconv.HTTPSchemeKey.String(c.Scheme()),
					semconv.HTTPHostKey.String(c.Request().Host),
					semconv.HTTPUserAgentKey.String(c.Request().UserAgent()),
				),
			)
			defer span.End()

			c.SetRequest(c.Request().WithContext(ctx))
			err := next(c)

			status := c.Response().Status
			if err != nil {
				if httpErr, ok := err.(*echo.HTTPError); ok && httpErr.Code != 0 {
					status = httpErr.Code
				}
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else if status >= 500 {
				span.SetStatus(codes.Error, fmt.Sprintf("http %d", status))
			}

			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(status),
				attribute.Bool("http.server_error", status >= 500),
			)

			if inst != nil {
				attrs := []attribute.KeyValue{
					attribute.String("http.route", c.Path()),
					attribute.String("http.method", c.Request().Method),
					attribute.Int("http.status_code", status),
				}
				inst.HTTPRequestCounter.Add(ctx, 1, attrs...)
				inst.HTTPRequestDuration.Record(ctx, float64(time.Since(start).Milliseconds()), attrs...)
			}

			return err
		}
	}
}
