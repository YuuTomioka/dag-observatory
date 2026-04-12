package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
		RunID:       "run-1",
		Partition:   state.Partition("p1"),
		EventTime:   runEvent.EventTime,
		SequenceNo:  1,
		IntentID:    "intent:entry",
		ExecutionID: "exec:intent:entry:1",
		TradeID:     "trade:intent:entry:1",
		NodeID:      "n1",
		NodeName:    "node.one",
		Status:      events.NodeExecutionStatusSucceeded,
		DurationNS:  1000,
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
		Items      []usecase.RunView `json:"items"`
		NextCursor string            `json:"next_cursor"`
	}
	if err := json.Unmarshal(rec1.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Items) != 1 || listResp.Items[0].RunID != "run-1" {
		t.Fatalf("unexpected list runs response: %#v", listResp)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/runs/run-1/steps/1", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetPath("/runs/:run_id/steps/:sequence_no")
	c2.SetParamNames("run_id", "sequence_no")
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
	if nodeResp.Node.NodeExecutionID != "1" || nodeResp.Node.ExecutionID != "exec:intent:entry:1" || nodeResp.Node.NodeName != "node.one" {
		t.Fatalf("unexpected node response: %#v", nodeResp.Node)
	}
}

func TestListRunsSupportsFiltersAndCursor(t *testing.T) {
	recorder := recorderinfra.NewInMemoryRecorder()
	base := time.Date(2026, 4, 12, 1, 0, 0, 0, time.UTC)
	for i, runID := range []string{"run-a", "run-b", "run-c"} {
		eventTime := base.Add(time.Duration(i) * time.Minute)
		eventType := "task.completed"
		if runID == "run-c" {
			eventType = "task.failed"
		}
		event := events.Event{
			EventID:   runID,
			EventTime: eventTime,
			Partition: state.Partition("p-cursor"),
			Type:      "task.requested",
		}
		recorder.RecordEvent(context.Background(), event)
		recorder.RecordCycleResult(context.Background(), port.CycleResult{
			Partition: state.Partition("p-cursor"),
			Event: events.Event{
				EventID:   runID,
				EventTime: eventTime,
				Partition: state.Partition("p-cursor"),
				Type:      eventType,
			},
			Duration: time.Second,
			Err: func() error {
				if runID == "run-c" {
					return errors.New("synthetic fail")
				}
				return nil
			}(),
		})
	}

	h := New(Dependencies{
		AppLog:   applog.New("info", "stdout", nil),
		ListRuns: &usecase.ListRuns{Reader: recorder},
	})
	e := echo.New()
	req := httptest.NewRequest(
		http.MethodGet,
		"/runs?partition=p-cursor&status=succeeded&since=2026-04-12T01:00:00Z&until=2026-04-12T01:03:00Z&limit=1&cursor=0",
		nil,
	)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListRuns(c); err != nil {
		t.Fatalf("list runs handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp1 struct {
		Items      []usecase.RunView `json:"items"`
		NextCursor string            `json:"next_cursor"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("decode first page: %v", err)
	}
	if len(resp1.Items) != 1 || resp1.Items[0].RunID != "run-b" {
		t.Fatalf("unexpected first page: %#v", resp1)
	}
	if resp1.NextCursor != "1" {
		t.Fatalf("expected next_cursor=1, got %q", resp1.NextCursor)
	}

	req2 := httptest.NewRequest(
		http.MethodGet,
		"/runs?partition=p-cursor&status=succeeded&since=2026-04-12T01:00:00Z&until=2026-04-12T01:03:00Z&limit=1&cursor="+resp1.NextCursor,
		nil,
	)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := h.ListRuns(c2); err != nil {
		t.Fatalf("list runs second page handler error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var resp2 struct {
		Items      []usecase.RunView `json:"items"`
		NextCursor string            `json:"next_cursor"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("decode second page: %v", err)
	}
	if len(resp2.Items) != 1 || resp2.Items[0].RunID != "run-a" {
		t.Fatalf("unexpected second page: %#v", resp2)
	}
	if resp2.NextCursor != "" {
		t.Fatalf("expected empty next_cursor on last page, got %q", resp2.NextCursor)
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

func TestListRunsRejectsInvalidCursor(t *testing.T) {
	h := New(Dependencies{
		AppLog:   applog.New("info", "stdout", nil),
		ListRuns: &usecase.ListRuns{Reader: recorderinfra.NewInMemoryRecorder()},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/runs?cursor=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListRuns(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetRunNodeByDeprecatedExecutionIDPath(t *testing.T) {
	recorder := recorderinfra.NewInMemoryRecorder()
	runEvent := events.Event{
		EventID:   "run-compat",
		EventTime: time.Date(2026, 4, 12, 2, 0, 0, 0, time.UTC),
		Partition: state.Partition("p-compat"),
		Type:      "task.requested",
	}
	recorder.RecordEvent(context.Background(), runEvent)
	recorder.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:       "run-compat",
		Partition:   state.Partition("p-compat"),
		EventTime:   runEvent.EventTime,
		SequenceNo:  10,
		ExecutionID: "exec:intent:entry:10",
		NodeID:      "n-compat",
		NodeName:    "node.compat",
		Status:      events.NodeExecutionStatusSucceeded,
	})

	h := New(Dependencies{
		AppLog:     applog.New("info", "stdout", nil),
		GetRunNode: &usecase.GetRunNode{Reader: recorder},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/runs/run-compat/nodes/exec:intent:entry:10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/runs/:run_id/nodes/:execution_id")
	c.SetParamNames("run_id", "execution_id")
	c.SetParamValues("run-compat", "exec:intent:entry:10")

	if err := h.GetRunNode(c); err != nil {
		t.Fatalf("get run node handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var nodeResp struct {
		Node usecase.RunStepView `json:"node"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &nodeResp); err != nil {
		t.Fatalf("decode node response: %v", err)
	}
	if nodeResp.Node.ExecutionID != "exec:intent:entry:10" {
		t.Fatalf("unexpected node response: %#v", nodeResp.Node)
	}
}

func TestListRunStepsDecodesEncodedRunID(t *testing.T) {
	recorder := recorderinfra.NewInMemoryRecorder()
	runID := "partition-1:heavy_calc:1"
	runEvent := events.Event{
		EventID:   runID,
		EventTime: time.Date(2026, 4, 12, 3, 0, 0, 0, time.UTC),
		Partition: state.Partition("partition-1"),
		Type:      "task.requested",
	}
	recorder.RecordEvent(context.Background(), runEvent)
	recorder.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:       runID,
		Partition:   state.Partition("partition-1"),
		EventTime:   runEvent.EventTime,
		SequenceNo:  1,
		ExecutionID: "exec:intent:entry:1",
		NodeID:      "n1",
		NodeName:    "node.one",
		Status:      events.NodeExecutionStatusSucceeded,
	})

	h := New(Dependencies{
		AppLog:       applog.New("info", "stdout", nil),
		ListRunSteps: &usecase.ListRunSteps{Reader: recorder},
	})
	e := echo.New()
	encodedRunID := url.PathEscape(runID)
	req := httptest.NewRequest(http.MethodGet, "/runs/"+encodedRunID+"/steps", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/runs/:run_id/steps")
	c.SetParamNames("run_id")
	c.SetParamValues(encodedRunID)

	if err := h.ListRunSteps(c); err != nil {
		t.Fatalf("list run steps handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		RunID string                `json:"run_id"`
		Items []usecase.RunStepView `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.RunID != runID {
		t.Fatalf("expected decoded run_id %q, got %q", runID, resp.RunID)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestCompareRunsHTTP(t *testing.T) {
	reader := recorderinfra.NewInMemoryRecorder()
	baseEvent := events.Event{
		EventID:   "base",
		EventTime: time.Date(2026, 4, 12, 1, 0, 0, 0, time.UTC),
		Partition: state.Partition("p1"),
		Type:      "task.requested",
	}
	targetEvent := events.Event{
		EventID:   "target",
		EventTime: time.Date(2026, 4, 12, 1, 1, 0, 0, time.UTC),
		Partition: state.Partition("p1"),
		Type:      "task.requested",
	}
	reader.RecordEvent(context.Background(), baseEvent)
	reader.RecordEvent(context.Background(), targetEvent)
	reader.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:      "base",
		Partition:  state.Partition("p1"),
		SequenceNo: 1,
		Status:     events.NodeExecutionStatusSucceeded,
		StateDiff: []events.StateDiffField{
			{Field: "strategy_summary.trade_count", After: 1.0},
		},
	})
	reader.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:      "target",
		Partition:  state.Partition("p1"),
		SequenceNo: 1,
		Status:     events.NodeExecutionStatusFailed,
		StateDiff: []events.StateDiffField{
			{Field: "strategy_summary.trade_count", After: 3.0},
		},
	})
	reader.RecordCycleResult(context.Background(), port.CycleResult{
		Partition: state.Partition("p1"),
		Event:     baseEvent,
		Duration:  2 * time.Second,
	})
	reader.RecordCycleResult(context.Background(), port.CycleResult{
		Partition: state.Partition("p1"),
		Event:     targetEvent,
		Duration:  4 * time.Second,
	})

	h := New(Dependencies{
		AppLog:      applog.New("info", "stdout", nil),
		CompareRuns: &usecase.CompareRuns{Reader: reader},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/runs/compare?base_run_id=base&target_run_id=target", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CompareRuns(c); err != nil {
		t.Fatalf("compare runs handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"failed_step_delta":1`) {
		t.Fatalf("expected failed_step_delta=1 in response: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"strategy_summary.trade_count"`) {
		t.Fatalf("expected state field delta in response: %s", rec.Body.String())
	}
}

func TestGetRunBacktestSummaryHTTP(t *testing.T) {
	recorder := recorderinfra.NewInMemoryRecorder()
	runEvent := events.Event{
		EventID:   "run-summary",
		EventTime: time.Date(2026, 4, 12, 4, 0, 0, 0, time.UTC),
		Partition: state.Partition("p-summary"),
		Type:      "task.requested",
	}
	recorder.RecordEvent(context.Background(), runEvent)
	recorder.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:      "run-summary",
		Partition:  state.Partition("p-summary"),
		SequenceNo: 1,
		StateDiff: []events.StateDiffField{
			{Field: "strategy_summary.trade_count", After: 4.0},
			{Field: "strategy_summary.total_net_pnl", After: 15.0},
		},
	})
	recorder.RecordCycleResult(context.Background(), port.CycleResult{
		Partition: state.Partition("p-summary"),
		Event:     runEvent,
		Duration:  2 * time.Second,
	})

	h := New(Dependencies{
		AppLog:             applog.New("info", "stdout", nil),
		GetBacktestSummary: &usecase.GetRunBacktestSummary{Reader: recorder},
	})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/runs/run-summary/backtest-summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/runs/:run_id/backtest-summary")
	c.SetParamNames("run_id")
	c.SetParamValues("run-summary")

	if err := h.GetRunBacktestSummary(c); err != nil {
		t.Fatalf("get run backtest summary handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"trade_count":4`) {
		t.Fatalf("expected trade_count in response: %s", rec.Body.String())
	}
}
