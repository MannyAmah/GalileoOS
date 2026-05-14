"use client";

// Client component for the Stage 0 Hello Agent demo.
//
// Flow: operator types a goal + pastes their tenant JWT → submits →
// POST /api/tasks creates a workflow → returned task_id is polled at
// 500ms via GET /api/tasks/{id} until the workflow completes →
// TaskResult renders.
//
// Polling cleanup follows the canonical Next.js 16 / React 19 pattern:
// AbortController cancels the in-flight fetch on unmount AND
// clearInterval prevents the next scheduled poll. Both go in the
// useEffect cleanup function; missing either leaks resources.

import { useEffect, useState } from "react";
import type {
  CreateTaskResponse,
  PollResponse,
  TaskInput,
  TaskResult,
} from "@/lib/api-types";

const POLL_INTERVAL_MS = 500;
const POLL_TIMEOUT_MS = 60_000;

export default function HelloAgent() {
  const [tenantId, setTenantId] = useState("");
  const [token, setToken] = useState("");
  const [goal, setGoal] = useState("Say hello in three words.");
  const [taskId, setTaskId] = useState<string | null>(null);
  const [result, setResult] = useState<TaskResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // Submit handler — POST /api/tasks, capture the task_id.
  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setResult(null);
    setSubmitting(true);
    try {
      const input: TaskInput = {
        tenant: {
          tenant_id: { value: tenantId },
          monthly_budget_cents: 99999,
        },
        department: "hello",
        goal,
      };
      const resp = await fetch("/api/tasks", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(input),
      });
      if (!resp.ok) {
        setError(`create-task failed: ${resp.status} ${await resp.text()}`);
        setSubmitting(false);
        return;
      }
      const body = (await resp.json()) as CreateTaskResponse;
      setTaskId(body.task_id);
    } catch (err) {
      setError(`submit error: ${(err as Error).message}`);
      setSubmitting(false);
    }
  }

  // Polling effect. Re-runs whenever taskId changes; cleanup cancels
  // in-flight fetch (AbortController) and the next-poll timer
  // (clearInterval). Both pieces are load-bearing — without the
  // interval cleanup, polls keep scheduling even when the AbortController
  // cancels each individual request.
  useEffect(() => {
    if (!taskId) return;
    const controller = new AbortController();
    const startedAt = Date.now();
    const id = setInterval(async () => {
      if (Date.now() - startedAt > POLL_TIMEOUT_MS) {
        setError("task did not complete within 60s");
        setSubmitting(false);
        clearInterval(id);
        return;
      }
      try {
        const resp = await fetch(`/api/tasks/${taskId}`, {
          method: "GET",
          headers: { Authorization: `Bearer ${token}` },
          signal: controller.signal,
        });
        if (!resp.ok) return; // transient — keep polling
        const body = (await resp.json()) as PollResponse;
        if (body.status === "running") return;
        setResult(body as TaskResult);
        setSubmitting(false);
        clearInterval(id);
      } catch (err) {
        if ((err as Error).name === "AbortError") return;
        // Other errors are transient; surface only if persistent.
      }
    }, POLL_INTERVAL_MS);
    return () => {
      controller.abort();
      clearInterval(id);
    };
  }, [taskId, token]);

  return (
    <section style={{ marginTop: 24 }}>
      <form onSubmit={onSubmit} style={{ display: "grid", gap: 12, maxWidth: 640 }}>
        <label>
          Tenant ID (UUID v7)
          <input
            type="text"
            value={tenantId}
            onChange={(e) => setTenantId(e.target.value)}
            placeholder="00000000-0000-7000-8000-..."
            style={inputStyle}
            required
          />
        </label>
        <label>
          Tenant JWT (Bearer token from `make stage0-jwt`)
          <textarea
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="eyJ..."
            rows={3}
            style={inputStyle}
            required
          />
        </label>
        <label>
          Goal
          <textarea
            value={goal}
            onChange={(e) => setGoal(e.target.value)}
            rows={2}
            style={inputStyle}
            required
          />
        </label>
        <button type="submit" disabled={submitting} style={buttonStyle}>
          {submitting ? "Running…" : "Run Hello Agent"}
        </button>
      </form>

      {error && (
        <p style={{ color: "crimson", marginTop: 16 }}>
          {error}
        </p>
      )}
      {result && (
        <article style={{ marginTop: 24, padding: 16, background: "#f6f6f6", borderRadius: 6 }}>
          <h2 style={{ margin: "0 0 8px" }}>{result.status}</h2>
          {result.output && (
            <>
              <p style={{ whiteSpace: "pre-wrap" }}>{result.output.body}</p>
              <p style={{ fontSize: 12, color: "#666" }}>
                cost: {result.output.cost_cents}¢ · cost_event_request_ids:{" "}
                {result.output.cost_event_request_ids.join(", ")}
              </p>
            </>
          )}
          {result.error && <p style={{ color: "crimson" }}>{result.error}</p>}
        </article>
      )}
    </section>
  );
}

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: "8px 10px",
  fontFamily: "monospace",
  fontSize: 13,
  marginTop: 4,
  border: "1px solid #ccc",
  borderRadius: 4,
};

const buttonStyle: React.CSSProperties = {
  padding: "10px 16px",
  background: "#222",
  color: "white",
  border: "none",
  borderRadius: 4,
  cursor: "pointer",
};
