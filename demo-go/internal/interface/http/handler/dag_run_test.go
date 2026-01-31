package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"dag-observatory/demo-go/internal/application/dagruntime/usecase"
	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/driver"
	"dag-observatory/demo-go/internal/domain/dagruntime/engine"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/policy"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/state"
	"dag-observatory/demo-go/internal/infrastructure/observability/applog"

	"github.com/labstack/echo/v4"
)

type inputCheckNode struct{}

func (n *inputCheckNode) Name() string { return "input.check" }
func (n *inputCheckNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeySymbol, usecase.InputKeyMode}
}
func (n *inputCheckNode) Provides() []artifact.AnyKey { return nil }
func (n *inputCheckNode) Reads() []state.AnyKey       { return nil }
func (n *inputCheckNode) Writes() []state.AnyKey      { return nil }
func (n *inputCheckNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *inputCheckNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = aw
	_ = txn
	if _, ok := artifact.Get(av, usecase.InputKeySymbol); !ok {
		return errors.New("missing symbol")
	}
	if _, ok := artifact.Get(av, usecase.InputKeyMode); !ok {
		return errors.New("missing mode")
	}
	return nil
}

func TestDagRunE2E(t *testing.T) {
	compiled := pipeline.Compiled{
		Name:  "e2e",
		Order: []node.Node{&inputCheckNode{}},
		Nodes: []node.Node{&inputCheckNode{}},
	}

	runner := &engine.Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{},
	}
	drv := &driver.Driver{
		Runner:   runner,
		Compiled: compiled,
	}
	uc := &usecase.RunWorkflow{Driver: drv}

	h := New(Dependencies{
		AppLog:     applog.New("info", "stdout", nil),
		RunWorkflow: uc,
	})

	e := echo.New()
	body, _ := json.Marshal(map[string]any{"symbol": "USDJPY", "mode": "normal"})
	req := httptest.NewRequest(http.MethodPost, "/dag/run", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.DagRun(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
