package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const runsUIPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>DAG Run Inspector</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f4f1e8;
      --panel: #fffdf8;
      --panel-strong: #f8f4ea;
      --border: #d8cdb7;
      --text: #1f1d1a;
      --muted: #6f6759;
      --accent: #1f6f78;
      --accent-soft: #d8ecee;
      --danger: #b3432f;
      --warn: #9b6c18;
      --ok: #2b6e3f;
      --shadow: 0 18px 40px rgba(58, 41, 22, 0.08);
      --radius: 18px;
      --mono: "IBM Plex Mono", "SFMono-Regular", Consolas, monospace;
      --sans: "IBM Plex Sans", "Segoe UI", sans-serif;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: var(--sans);
      background:
        radial-gradient(circle at top right, rgba(31,111,120,0.18), transparent 28%),
        radial-gradient(circle at bottom left, rgba(179,67,47,0.12), transparent 22%),
        var(--bg);
      color: var(--text);
    }
    .page {
      min-height: 100vh;
      padding: 24px;
    }
    .shell {
      max-width: 1440px;
      margin: 0 auto;
      display: grid;
      gap: 18px;
    }
    .hero, .pane {
      background: var(--panel);
      border: 1px solid var(--border);
      border-radius: var(--radius);
      box-shadow: var(--shadow);
    }
    .hero {
      padding: 22px 24px;
      display: grid;
      gap: 16px;
    }
    .hero-top {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: start;
      flex-wrap: wrap;
    }
    .hero h1 {
      margin: 0;
      font-size: 2rem;
      letter-spacing: -0.04em;
    }
    .hero p {
      margin: 6px 0 0;
      color: var(--muted);
      max-width: 64ch;
    }
    .controls {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
      align-items: center;
    }
    .controls input, .controls button {
      font: inherit;
      border-radius: 999px;
      border: 1px solid var(--border);
      background: #fff;
      color: var(--text);
      padding: 10px 14px;
    }
    .controls input {
      min-width: 240px;
    }
    .controls button {
      background: var(--accent);
      color: #fff;
      border-color: var(--accent);
      cursor: pointer;
    }
    .meta-strip {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
    }
    .meta-card {
      background: var(--panel-strong);
      border: 1px solid var(--border);
      border-radius: 14px;
      padding: 14px;
    }
    .meta-card .label {
      color: var(--muted);
      font-size: 0.8rem;
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    .meta-card .value {
      display: block;
      margin-top: 8px;
      font-size: 1.05rem;
      font-weight: 600;
    }
    .content {
      display: grid;
      grid-template-columns: minmax(320px, 0.95fr) minmax(420px, 1.35fr);
      gap: 18px;
    }
    .pane {
      overflow: hidden;
      min-height: 65vh;
    }
    .pane-header {
      padding: 18px 20px 14px;
      border-bottom: 1px solid var(--border);
      background: linear-gradient(180deg, rgba(248,244,234,0.8), rgba(255,253,248,0.6));
    }
    .pane-header h2, .pane-header h3 {
      margin: 0;
      font-size: 1rem;
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    .pane-header p {
      margin: 8px 0 0;
      color: var(--muted);
      font-size: 0.92rem;
    }
    .timeline {
      display: grid;
      gap: 10px;
      padding: 16px;
      max-height: calc(65vh - 70px);
      overflow: auto;
    }
    .step {
      width: 100%;
      border: 1px solid var(--border);
      background: #fff;
      color: inherit;
      text-align: left;
      border-radius: 16px;
      padding: 14px;
      cursor: pointer;
      transition: transform 120ms ease, border-color 120ms ease, background 120ms ease;
    }
    .step:hover, .step:focus-visible {
      transform: translateY(-1px);
      border-color: var(--accent);
      outline: none;
    }
    .step.active {
      background: var(--accent-soft);
      border-color: var(--accent);
    }
    .step-top, .detail-grid {
      display: grid;
      gap: 8px;
    }
    .step-top {
      grid-template-columns: auto 1fr auto;
      align-items: center;
    }
    .sequence {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-width: 36px;
      height: 36px;
      border-radius: 50%;
      background: var(--panel-strong);
      font-family: var(--mono);
      font-size: 0.85rem;
    }
    .node-name {
      font-weight: 600;
      font-size: 1rem;
    }
    .node-id {
      font-family: var(--mono);
      color: var(--muted);
      font-size: 0.8rem;
    }
    .status {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      border-radius: 999px;
      padding: 6px 10px;
      font-size: 0.8rem;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      border: 1px solid currentColor;
    }
    .status.succeeded { color: var(--ok); }
    .status.failed { color: var(--danger); }
    .status.skipped { color: var(--warn); }
    .status.running, .status.unknown { color: var(--accent); }
    .micro {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
      margin-top: 10px;
    }
    .pill {
      padding: 5px 8px;
      border-radius: 999px;
      background: var(--panel-strong);
      color: var(--muted);
      font-size: 0.8rem;
    }
    .detail {
      padding: 20px;
      display: grid;
      gap: 18px;
    }
    .detail-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .field {
      padding: 12px 14px;
      border-radius: 14px;
      background: var(--panel-strong);
      border: 1px solid var(--border);
    }
    .field .label {
      display: block;
      color: var(--muted);
      font-size: 0.78rem;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      margin-bottom: 8px;
    }
    .field .value {
      font-size: 0.94rem;
      line-height: 1.5;
      overflow-wrap: anywhere;
    }
    .field pre {
      margin: 0;
      font: 0.86rem/1.5 var(--mono);
      white-space: pre-wrap;
    }
    .section {
      display: grid;
      gap: 10px;
    }
    .section h3 {
      margin: 0;
      font-size: 0.95rem;
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    .diff-list {
      display: grid;
      gap: 10px;
    }
    .diff-item {
      border: 1px solid var(--border);
      border-radius: 14px;
      padding: 14px;
      background: #fff;
    }
    .diff-item strong {
      display: block;
      font-family: var(--mono);
      margin-bottom: 10px;
    }
    .diff-columns {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
    }
    .diff-box {
      background: var(--panel-strong);
      border-radius: 12px;
      padding: 10px;
      border: 1px solid var(--border);
    }
    .diff-box .label {
      display: block;
      color: var(--muted);
      font-size: 0.75rem;
      text-transform: uppercase;
      margin-bottom: 8px;
    }
    .diff-box pre {
      margin: 0;
      font: 0.84rem/1.45 var(--mono);
      white-space: pre-wrap;
      overflow-wrap: anywhere;
    }
    .hint, .error {
      padding: 14px;
      border-radius: 14px;
      border: 1px solid var(--border);
      background: var(--panel-strong);
      color: var(--muted);
    }
    .error {
      color: var(--danger);
      border-color: rgba(179,67,47,0.35);
      background: rgba(179,67,47,0.08);
    }
    @media (max-width: 980px) {
      .meta-strip, .content, .detail-grid, .diff-columns {
        grid-template-columns: 1fr;
      }
      .timeline {
        max-height: none;
      }
      .pane {
        min-height: auto;
      }
    }
  </style>
</head>
<body>
  <div class="page">
    <div class="shell">
      <section class="hero">
        <div class="hero-top">
          <div>
            <h1>DAG Run Inspector</h1>
            <p>Execution observability first. Inspect run order, node outcomes, and state diff without committing to a graph-heavy UI.</p>
          </div>
          <form class="controls" id="controls">
            <input id="partitionInput" name="partition" type="text" placeholder="partition filter">
            <select id="compareRunSelect" name="target_run_id" title="compare target run">
              <option value="">Compare target (optional)</option>
            </select>
            <button type="submit">Load Runs</button>
          </form>
        </div>
        <div class="meta-strip">
          <div class="meta-card">
            <span class="label">Run</span>
            <span class="value" id="runLabel">No run selected</span>
          </div>
          <div class="meta-card">
            <span class="label">Partition</span>
            <span class="value" id="partitionLabel">-</span>
          </div>
          <div class="meta-card">
            <span class="label">Summary</span>
            <span class="value" id="summaryLabel">-</span>
          </div>
          <div class="meta-card">
            <span class="label">Detail</span>
            <span class="value" id="detailLabel">Select a node</span>
          </div>
        </div>
      </section>

      <section class="content">
        <div class="pane">
          <div class="pane-header">
            <h2>Execution Timeline</h2>
            <p>Runs are listed from the read model. Click a step to inspect state diff and refs.</p>
          </div>
          <div class="timeline" id="timeline">
            <div class="hint">Loading runs...</div>
          </div>
        </div>

        <div class="pane">
          <div class="pane-header">
            <h2>Node Detail</h2>
            <p>Selected node execution with refs, reasons, and compact diff.</p>
          </div>
          <div class="detail" id="detail">
            <div class="hint">Select a run and node execution.</div>
          </div>
        </div>
      </section>
    </div>
  </div>

  <script>
    const state = {
      partition: "",
      runs: [],
      runsNextCursor: "",
      selectedRun: null,
      steps: [],
      stepsSyncing: false,
      stepsSyncToken: 0,
      selectedSequenceNo: "",
      selectedNode: null,
      summary: null,
      compareTargetRunID: "",
      compare: null
    };
    const RUNS_PAGE_LIMIT = 200;
    const STEP_RENDER_LIMIT = 300;

    const timelineEl = document.getElementById("timeline");
    const detailEl = document.getElementById("detail");
    const partitionInput = document.getElementById("partitionInput");
    const compareRunSelect = document.getElementById("compareRunSelect");
    const runLabelEl = document.getElementById("runLabel");
    const partitionLabelEl = document.getElementById("partitionLabel");
    const summaryLabelEl = document.getElementById("summaryLabel");
    const detailLabelEl = document.getElementById("detailLabel");

    function qs(key) {
      return new URLSearchParams(window.location.search).get(key) || "";
    }

    function setQuery(params) {
      const next = new URLSearchParams();
      Object.entries(params).forEach(function(entry) {
        const key = entry[0];
        const value = entry[1];
        if (value) next.set(key, value);
      });
      const nextURL = next.toString() ? "?" + next.toString() : window.location.pathname;
      window.history.replaceState({}, "", nextURL);
    }

    function escapeHTML(value) {
      return String(value)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
    }

    function pretty(value) {
      if (value === null || value === undefined || value === "") return "-";
      if (typeof value === "string") return value;
      return JSON.stringify(value, null, 2);
    }

    function formatDurationNS(value) {
      if (!value) return "0 ms";
      const ms = value / 1000000;
      if (ms >= 1000) return (ms / 1000).toFixed(2) + " s";
      if (ms >= 1) return ms.toFixed(2) + " ms";
      return value + " ns";
    }

    async function getJSON(url) {
      const res = await fetch(url);
      if (!res.ok) {
        let message = res.status + " " + res.statusText;
        try {
          const body = await res.json();
          if (body && body.error) message = body.error;
        } catch (err) {}
        throw new Error(message);
      }
      return res.json();
    }

    async function loadRuns() {
      const params = new URLSearchParams();
      params.set("limit", String(RUNS_PAGE_LIMIT));
      if (state.partition) params.set("partition", state.partition);
      const payload = await getJSON("/runs?" + params.toString());
      state.runs = payload.items || [];
      state.runsNextCursor = payload.next_cursor || "";
      if (!state.runs.length) {
        state.selectedRun = null;
        state.steps = [];
        state.selectedNode = null;
        state.summary = null;
        state.compare = null;
        state.compareTargetRunID = "";
        state.selectedSequenceNo = "";
        render();
        return;
      }
      const requestedRunID = qs("run_id");
      state.selectedRun = state.runs.find(function(run) {
        return run.run_id === requestedRunID;
      }) || state.runs[0];
      const requestedTargetRunID = qs("target_run_id");
      state.compareTargetRunID = requestedTargetRunID;
      await loadSelectedRun();
    }

    async function loadSelectedRun() {
      if (!state.selectedRun) {
        render();
        return;
      }
      state.stepsSyncToken += 1;
      const syncToken = state.stepsSyncToken;
      const runID = state.selectedRun.run_id;
      const expectedCount = Number(state.selectedRun.step_count || 0);
      state.stepsSyncing = false;
      state.steps = await fetchRunStepsWithRetry(runID, expectedCount);
      if (!state.steps.length && expectedCount > 0) {
        state.stepsSyncing = true;
        render();
        hydrateStepsEventually(runID, expectedCount, syncToken);
      }
      const requestedSequenceNo = qs("sequence_no");
      const initialStep = state.steps.find(function(step) {
        return step.node_execution_id === requestedSequenceNo;
      }) || state.steps[0] || null;
      state.selectedSequenceNo = initialStep ? initialStep.node_execution_id : "";
      if (!state.compareTargetRunID || !state.runs.find(function(run) { return run.run_id === state.compareTargetRunID; })) {
        const fallback = state.runs.find(function(run) {
          return run.run_id !== state.selectedRun.run_id;
        });
        state.compareTargetRunID = fallback ? fallback.run_id : "";
      }
      await Promise.all([
        loadSelectedNode(),
        loadSummary(),
        loadCompare()
      ]);
    }

    async function fetchRunStepsWithRetry(runID, expectedCount) {
      const maxRetry = expectedCount > 0 ? 3 : 0;
      for (let attempt = 0; attempt <= maxRetry; attempt++) {
        const stepsPayload = await getJSON("/runs/" + encodeURIComponent(runID) + "/steps");
        const items = stepsPayload.items || [];
        if (items.length > 0 || expectedCount <= 0 || attempt === maxRetry) {
          return items;
        }
        await new Promise(function(resolve) { setTimeout(resolve, 200); });
      }
      return [];
    }

    async function hydrateStepsEventually(runID, expectedCount, syncToken) {
      const maxRetry = expectedCount > 0 ? 30 : 0;
      for (let attempt = 0; attempt < maxRetry; attempt++) {
        if (state.stepsSyncToken !== syncToken) return;
        await new Promise(function(resolve) { setTimeout(resolve, 300); });
        if (state.stepsSyncToken !== syncToken) return;
        const stepsPayload = await getJSON("/runs/" + encodeURIComponent(runID) + "/steps");
        const items = stepsPayload.items || [];
        if (!items.length) continue;
        if (state.stepsSyncToken !== syncToken) return;
        state.steps = items;
        const requestedSequenceNo = qs("sequence_no");
        const initialStep = state.steps.find(function(step) {
          return step.node_execution_id === requestedSequenceNo;
        }) || state.steps[0] || null;
        state.selectedSequenceNo = initialStep ? initialStep.node_execution_id : "";
        state.stepsSyncing = false;
        await loadSelectedNode();
        return;
      }
      if (state.stepsSyncToken !== syncToken) return;
      state.stepsSyncing = false;
      render();
    }

    async function loadSelectedNode() {
      if (!state.selectedRun || !state.selectedSequenceNo) {
        state.selectedNode = null;
        render();
        return;
      }
      const runID = encodeURIComponent(state.selectedRun.run_id);
      const sequenceNo = encodeURIComponent(state.selectedSequenceNo);
      const payload = await getJSON("/runs/" + runID + "/steps/" + sequenceNo);
      state.selectedNode = payload.node || null;
      render();
    }

    async function loadSummary() {
      if (!state.selectedRun || !state.selectedRun.partition) {
        state.summary = null;
        render();
        return;
      }
      try {
        const payload = await getJSON("/algotrade/summary?partition=" + encodeURIComponent(state.selectedRun.partition));
        state.summary = payload.summary || null;
      } catch (err) {
        state.summary = null;
      }
      render();
    }

    async function loadCompare() {
      if (!state.selectedRun || !state.compareTargetRunID || state.selectedRun.run_id === state.compareTargetRunID) {
        state.compare = null;
        render();
        return;
      }
      try {
        const payload = await getJSON(
          "/runs/compare?base_run_id=" + encodeURIComponent(state.selectedRun.run_id) +
          "&target_run_id=" + encodeURIComponent(state.compareTargetRunID)
        );
        state.compare = payload.compare || null;
      } catch (err) {
        state.compare = null;
      }
      render();
    }

    function renderTimeline() {
      if (!state.runs.length) {
        timelineEl.innerHTML = '<div class="hint">No runs found for the current filter.</div>';
        return;
      }
      const runHeader = [
        '<div class="section">',
        '<h3>Runs</h3>',
        state.runs.map(function(run) {
          const active = state.selectedRun && run.run_id === state.selectedRun.run_id ? " active" : "";
          return [
            '<button class="step' + active + '" data-run-id="' + escapeHTML(run.run_id) + '">',
            '<div class="step-top">',
            '<span class="sequence">R</span>',
            '<div><div class="node-name">' + escapeHTML(run.run_id) + '</div><div class="node-id">' + escapeHTML(run.partition) + "</div></div>",
            '<span class="status ' + escapeHTML(run.status || "unknown") + '">' + escapeHTML(run.status || "unknown") + "</span>",
            "</div>",
            '<div class="micro">',
            '<span class="pill">' + escapeHTML(run.event_type || "-") + "</span>",
            '<span class="pill">' + escapeHTML(String(run.step_count || 0)) + " steps</span>",
            '<span class="pill">' + escapeHTML(String(run.duration_ms || 0)) + " ms</span>",
            "</div>",
            "</button>"
          ].join("");
        }).join(""),
        state.runsNextCursor ? '<div class="hint">Showing first ' + escapeHTML(String(RUNS_PAGE_LIMIT)) + ' runs. Apply partition filter to narrow results.</div>' : '',
        "</div>"
      ].join("");

      const renderedSteps = state.steps.slice(0, STEP_RENDER_LIMIT);
      const stepHeader = !state.selectedRun ? '<div class="hint">Select a run.</div>' : [
        '<div class="section">',
        '<h3>Steps</h3>',
        renderedSteps.length ? renderedSteps.map(function(step) {
          const active = step.node_execution_id === state.selectedSequenceNo ? " active" : "";
          return [
            '<button class="step' + active + '" data-sequence-no="' + escapeHTML(step.node_execution_id) + '">',
            '<div class="step-top">',
            '<span class="sequence">' + escapeHTML(String(step.sequence_no)) + "</span>",
            '<div><div class="node-name">' + escapeHTML(step.node_name || step.node_id) + '</div><div class="node-id">' + escapeHTML(step.node_id) + "</div></div>",
            '<span class="status ' + escapeHTML(step.status || "unknown") + '">' + escapeHTML(step.status || "unknown") + "</span>",
            "</div>",
            '<div class="micro">',
            '<span class="pill">trigger: ' + escapeHTML(step.trigger_reason || "-") + "</span>",
            '<span class="pill">skip: ' + escapeHTML(step.skip_reason || "-") + "</span>",
            '<span class="pill">' + escapeHTML(formatDurationNS(step.duration_ns)) + "</span>",
            "</div>",
            "</button>"
          ].join("");
        }).join("") : ('<div class="hint">' + (state.stepsSyncing ? "Waiting for step records to become available..." : "No steps recorded for this run.") + '</div>'),
        state.steps.length > renderedSteps.length ? '<div class="hint">Showing first ' + escapeHTML(String(STEP_RENDER_LIMIT)) + ' steps.</div>' : '',
        "</div>"
      ].join("");

      timelineEl.innerHTML = runHeader + stepHeader;
      compareRunSelect.innerHTML = ['<option value="">Compare target (optional)</option>'].concat(
        state.runs.filter(function(run) {
          return !state.selectedRun || run.run_id !== state.selectedRun.run_id;
        }).map(function(run) {
          const selected = run.run_id === state.compareTargetRunID ? ' selected' : '';
          return '<option value="' + escapeHTML(run.run_id) + '"' + selected + '>' +
            escapeHTML(run.run_id + " (" + (run.status || "unknown") + ")") + "</option>";
        })
      ).join("");
      timelineEl.querySelectorAll("[data-run-id]").forEach(function(button) {
        button.addEventListener("click", async function() {
          const runID = button.getAttribute("data-run-id");
          const next = state.runs.find(function(run) { return run.run_id === runID; }) || null;
          if (!next) return;
          state.selectedRun = next;
          state.selectedSequenceNo = "";
          setQuery({
            partition: state.partition,
            run_id: next.run_id,
            sequence_no: "",
            target_run_id: state.compareTargetRunID
          });
          timelineEl.innerHTML = '<div class="hint">Loading steps...</div>';
          await loadSelectedRun();
        });
      });
      timelineEl.querySelectorAll("[data-sequence-no]").forEach(function(button) {
        button.addEventListener("click", async function() {
          state.selectedSequenceNo = button.getAttribute("data-sequence-no") || "";
          setQuery({
            partition: state.partition,
            run_id: state.selectedRun ? state.selectedRun.run_id : "",
            sequence_no: state.selectedSequenceNo,
            target_run_id: state.compareTargetRunID
          });
          detailEl.innerHTML = '<div class="hint">Loading node detail...</div>';
          await loadSelectedNode();
        });
      });
    }

    function renderDetail() {
      updateMeta();
      if (!state.selectedNode) {
        detailEl.innerHTML = '<div class="hint">Select a node execution to inspect detail and diff.</div>';
        return;
      }
      const node = state.selectedNode;
      const refs = [
        field("Trigger Reason", node.trigger_reason || "-"),
        field("Skip Reason", node.skip_reason || "-"),
        field("Input Ref", node.input_ref || "-"),
        field("Output Ref", node.output_ref || "-"),
        field("Snapshot Before", node.snapshot_before_ref || "-"),
        field("Snapshot After", node.snapshot_after_ref || "-"),
        field("Duration", formatDurationNS(node.duration_ns)),
        field("Error", node.error || "-")
      ].join("");
      const diffs = node.state_diff && node.state_diff.length ? node.state_diff.map(function(item) {
        return [
          '<div class="diff-item">',
          "<strong>" + escapeHTML(item.field || "-") + "</strong>",
          '<div class="diff-columns">',
          '<div class="diff-box"><span class="label">Before</span><pre>' + escapeHTML(pretty(item.before)) + "</pre></div>",
          '<div class="diff-box"><span class="label">After</span><pre>' + escapeHTML(pretty(item.after)) + "</pre></div>",
          "</div>",
          "</div>"
        ].join("");
      }).join("") : '<div class="hint">No state diff fields were recorded for this node.</div>';

      detailEl.innerHTML = [
        '<div class="section">',
        '<h3>' + escapeHTML(node.node_name || node.node_id || "Node") + "</h3>",
        '<div class="micro">',
        '<span class="status ' + escapeHTML(node.status || "unknown") + '">' + escapeHTML(node.status || "unknown") + "</span>",
        '<span class="pill">sequence ' + escapeHTML(String(node.sequence_no || "-")) + "</span>",
        '<span class="pill">intent ' + escapeHTML(node.intent_id || "-") + "</span>",
        '<span class="pill">execution ' + escapeHTML(node.execution_id || "-") + "</span>",
        '<span class="pill">trade ' + escapeHTML(node.trade_id || "-") + "</span>",
        "</div>",
        "</div>",
        '<div class="detail-grid">' + refs + "</div>",
        '<div class="section"><h3>State Diff</h3><div class="diff-list">' + diffs + "</div></div>",
        state.compare ? renderCompare() : "",
        state.summary ? renderSummary() : ""
      ].join("");
    }

    function renderCompare() {
      const c = state.compare;
      const changed = (c.state_field_deltas || []).filter(function(item) { return item.changed; });
      const changedPreview = changed.slice(0, 8).map(function(item) {
        return [
          '<div class="diff-item">',
          "<strong>" + escapeHTML(item.field || "-") + "</strong>",
          '<div class="diff-columns">',
          '<div class="diff-box"><span class="label">Base After</span><pre>' + escapeHTML(pretty(item.base_after)) + "</pre></div>",
          '<div class="diff-box"><span class="label">Target After</span><pre>' + escapeHTML(pretty(item.target_after)) + "</pre></div>",
          "</div>",
          "</div>"
        ].join("");
      }).join("");
      return [
        '<div class="section">',
        "<h3>Run Compare</h3>",
        '<div class="detail-grid">',
        field("Base Run", c.base ? c.base.run_id : "-"),
        field("Target Run", c.target ? c.target.run_id : "-"),
        field("Duration Delta (ms)", c.summary_delta ? String(c.summary_delta.duration_ms_delta || 0) : "0"),
        field("Failed Step Delta", c.summary_delta ? String(c.summary_delta.failed_step_delta || 0) : "0"),
        field("Skipped Step Delta", c.summary_delta ? String(c.summary_delta.skipped_step_delta || 0) : "0"),
        field("Changed State Fields", String(changed.length)),
        "</div>",
        '<div class="section"><h3>Changed State Fields (Top 8)</h3><div class="diff-list">' + (changedPreview || '<div class="hint">No changed fields.</div>') + "</div></div>",
        "</div>"
      ].join("");
    }

    function renderSummary() {
      const s = state.summary;
      return [
        '<div class="section">',
        "<h3>Partition Summary</h3>",
        '<div class="detail-grid">',
        field("Strategy", s.strategy_id || "-"),
        field("Workflow", [s.workflow_name || "-", s.workflow_version || ""].join(" ").trim()),
        field("Trades", String(s.trade_count || 0)),
        field("Win Rate", String(s.win_rate || 0)),
        field("Total Net PnL", String(s.total_net_pnl || 0)),
        field("Updated", s.updated_at || "-"),
        "</div>",
        "</div>"
      ].join("");
    }

    function field(label, value) {
      return '<div class="field"><span class="label">' + escapeHTML(label) + '</span><div class="value"><pre>' + escapeHTML(pretty(value)) + "</pre></div></div>";
    }

    function updateMeta() {
      runLabelEl.textContent = state.selectedRun ? state.selectedRun.run_id : "No run selected";
      partitionLabelEl.textContent = state.selectedRun ? state.selectedRun.partition || "-" : "-";
      if (state.summary) {
        let text = "trades " + String(state.summary.trade_count || 0) + " / pnl " + String(state.summary.total_net_pnl || 0);
        if (state.compare && state.compare.summary_delta) {
          text += " / failΔ " + String(state.compare.summary_delta.failed_step_delta || 0);
        }
        summaryLabelEl.textContent = text;
      } else {
        summaryLabelEl.textContent = "-";
      }
      if (state.selectedNode) {
        detailLabelEl.textContent = (state.selectedNode.node_name || state.selectedNode.node_id || "node") + " #" + state.selectedNode.sequence_no;
      } else {
        detailLabelEl.textContent = "Select a node";
      }
    }

    function render() {
      renderTimeline();
      renderDetail();
    }

    async function boot() {
      state.partition = qs("partition");
      partitionInput.value = state.partition;
      document.getElementById("controls").addEventListener("submit", async function(event) {
        event.preventDefault();
        state.partition = partitionInput.value.trim();
        setQuery({
          partition: state.partition,
          run_id: "",
          sequence_no: "",
          target_run_id: state.compareTargetRunID
        });
        timelineEl.innerHTML = '<div class="hint">Loading runs...</div>';
        detailEl.innerHTML = '<div class="hint">Select a node execution to inspect detail and diff.</div>';
        try {
          await loadRuns();
        } catch (err) {
          timelineEl.innerHTML = '<div class="error">' + escapeHTML(err.message) + "</div>";
        }
      });
      compareRunSelect.addEventListener("change", async function() {
        state.compareTargetRunID = compareRunSelect.value || "";
        setQuery({
          partition: state.partition,
          run_id: state.selectedRun ? state.selectedRun.run_id : "",
          sequence_no: state.selectedSequenceNo,
          target_run_id: state.compareTargetRunID
        });
        await loadCompare();
      });
      try {
        await loadRuns();
      } catch (err) {
        timelineEl.innerHTML = '<div class="error">' + escapeHTML(err.message) + "</div>";
      }
    }

    boot();
  </script>
</body>
</html>`

func (h *Handler) RunsUI(c echo.Context) error {
	return c.HTML(http.StatusOK, runsUIPage)
}
