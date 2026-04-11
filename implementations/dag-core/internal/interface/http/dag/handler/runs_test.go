package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	recorderinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/recorder"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"

	"github.com/labstack/echo/v4"
)

func TestListRunsAndGetRunNodeHTTP(t *testing.T) {
	recorder := recorderinfra.NewInMemoryRecorder()
	runEvent := events.Event{
		EventID:   "run-1",
		EventTime: time.Date(2026, 4, 12, 1, 0, 0, 0, time.UTC),
		Partition: state.Partition("p1"),
		Type:      "task.requested",
	}
	recorder.RecordEvent(context.Background(), runEvent)
	recorder.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:      "run-1",
		Partition:  state.Partition("p1"),
		EventTime:  runEvent.EventTime,
		SequenceNo: 1,
		NodeID:     "n1",
		NodeName:   "node.one",
		Status:     events.NodeExecutionStatusSucceeded,
		DurationNS: 1000,
	})
	recorder.RecordCycleResult(context.Background(), port.CycleResult{
		Partition: state.Partition("p1"),
		Event:     runEvent,
		Duration:  time.Second,
	})

	h := New(Dependencies{
		AppLog:       applog.New("info", "stdout", nil),
		ListRuns:     &usecase.ListRuns{Reader: recorder},
		GetRun:       &usecase.GetRun{Reader: recorder},
		ListRunSteps: &usecase.ListRunSteps{Reader: recorder},
		GetRunNode:   &usecase.GetRunNode{Reader: recorder},
	})
	e := echo.New()

	req1 := httptest.NewRequest(http.MethodGet, "/runs?partition=p1", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := h.ListRuns(c1); err != nil {
		t.Fatalf("list runs handler error: %v", err)
	}
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec1.Code, rec1.Body.String())
	}
	var listResp struct {
		Items []usecase.RunView `json:"items"`
	}
	if err := json.Unmarshal(rec1.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Items) != 1 || listResp.Items[0].RunID != "run-1" {
		t.Fatalf("unexpected list runs response: %#v", listResp)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/runs/run-1/nodes/1", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetPath("/runs/:run_id/nodes/:execution_id")
	c2.SetParamNames("run_id", "execution_id")
	c2.SetParamValues("run-1", "1")
	if err := h.GetRunNode(c2); err != nil {
		t.Fatalf("get run node handler error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var nodeResp struct {
		Node usecase.RunStepView `json:"node"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &nodeResp); err != nil {
		t.Fatalf("decode node response: %v", err)
	}
	if nodeResp.Node.ExecutionID != "1" || nodeResp.Node.NodeName != "node.one" {
		t.Fatalf("unexpected node response: %#v", nodeResp.Node)
	}
}

func TestGetRunReturnsNotFound(t *testing.T) {
	h := New(Dependencies{
		AppLog: applog.New("info", "stdout", nil),
		GetRun: &usecase.GetRun{Reader: recorderinfra.NewInMemoryRecorder()},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/runs/missing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/runs/:run_id")
	c.SetParamNames("run_id")
	c.SetParamValues("missing")

	if err := h.GetRun(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
