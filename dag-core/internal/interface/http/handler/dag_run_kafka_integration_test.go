package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	eventstore "dag-observatory/dag-core/internal/infrastructure/dagruntime/eventstore"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"

	"github.com/labstack/echo/v4"
)

func TestHTTPToKafkaEnqueue(t *testing.T) {
	brokers := splitCSV(os.Getenv("KAFKA_BROKERS"))
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if len(brokers) == 0 || topic == "" {
		t.Skip("KAFKA_BROKERS and KAFKA_TOPIC are required for Kafka integration test")
	}
	if groupID == "" {
		groupID = "dagruntime-test"
	}

	producer, err := eventstore.NewKafkaProducer(brokers, topic)
	if err != nil {
		t.Fatalf("producer init failed: %v", err)
	}
	defer producer.Close()

	consumer, err := eventstore.NewKafkaConsumer(brokers, topic, groupID)
	if err != nil {
		t.Fatalf("consumer init failed: %v", err)
	}
	defer consumer.Close()

	runWF := &usecase.RunWorkflow{
		Enqueuer: producer,
	}
	h := New(Dependencies{
		AppLog:     applog.New("info", "stdout", nil),
		RunWorkflow: runWF,
	})

	payload := map[string]any{
		"symbol": "EURUSD",
		"mode":   "debug",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/dag/run", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	if err := h.DagRun(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out := make(chan events.Event, 1)
	go func() {
		_ = consumer.Run(ctx, out)
	}()

	var got events.Event
	select {
	case got = <-out:
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}

	if got.Type != "http.dag.run" {
		t.Fatalf("unexpected event type: %s", got.Type)
	}
	if string(got.Partition) != "EURUSD" {
		t.Fatalf("unexpected partition: %s", got.Partition)
	}
	envelopes, ok := got.Payload.([]events.PayloadEnvelope)
	if !ok || len(envelopes) == 0 {
		t.Fatalf("unexpected payload: %T", got.Payload)
	}
}

func splitCSV(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
