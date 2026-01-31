package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"dag-observatory/demo-go/internal/application/dagruntime/usecase"
	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/driver"
	"dag-observatory/demo-go/internal/domain/dagruntime/engine"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/policy"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/artifact"
	eventstore "dag-observatory/demo-go/internal/infrastructure/dagruntime/eventstore"
	stateinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/state"
	"dag-observatory/demo-go/internal/infrastructure/observability/applog"

	"github.com/labstack/echo/v4"
)

type countNode struct {
	mu    sync.Mutex
	count int
}

func (n *countNode) Name() string { return "count.node" }
func (n *countNode) Requires() []artifact.AnyKey { return nil }
func (n *countNode) Provides() []artifact.AnyKey { return nil }
func (n *countNode) Reads() []state.AnyKey       { return nil }
func (n *countNode) Writes() []state.AnyKey      { return nil }
func (n *countNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *countNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	_ = txn
	n.mu.Lock()
	n.count++
	n.mu.Unlock()
	return nil
}

func (n *countNode) Count() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.count
}

func TestHTTPToKafkaToDriver(t *testing.T) {
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

	runWF := &usecase.RunWorkflow{Enqueuer: producer}
	h := New(Dependencies{
		AppLog:     applog.New("info", "stdout", nil),
		RunWorkflow: runWF,
	})

	payload := map[string]any{
		"symbol": "USDJPY",
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

	stream := make(chan events.Event, 1)
	stream <- got
	close(stream)

	countNodeInst := &countNode{}
	driver := &driver.Driver{
		Runner: &engine.Runner{
			ArtifactStore: artifactinfra.NewMemoryStore(),
			StateStore:    stateinfra.NewMemoryStore(),
			Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
		},
		Compiled: pipeline.Compiled{
			Name:  "test",
			Order: []node.Node{countNodeInst},
			Nodes: []node.Node{countNodeInst},
		},
	}

	if err := driver.Run(context.Background(), stream); err != nil {
		t.Fatalf("driver run failed: %v", err)
	}
	if countNodeInst.Count() != 1 {
		t.Fatalf("expected node to run once, got %d", countNodeInst.Count())
	}
}
