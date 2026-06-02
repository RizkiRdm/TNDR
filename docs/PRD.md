# PRD.md — buruh
# Product Requirements Document

Document Version: 1.0.0
Project: buruh — Distributed Task Queue Engine
Status: MVP Definition
Author: solo developer
License: MIT

---

## 1. Project Summary

### Overview

buruh is an open-source distributed task queue engine written in Go.
It provides a reliable, observable mechanism for enqueueing, distributing,
and executing background tasks across a configurable pool of concurrent workers.

buruh ships with two primary artifacts:

1. **Core engine** — embeddable Go library + standalone binary
2. **Web dashboard** — realtime HTML/CSS/JS visualizer served by the engine itself

### Objective

Build a production-usable task queue engine that:
- developers can embed in their Go projects as a library
- developers can run as a standalone daemon process
- non-technical stakeholders can observe via a web dashboard
- junior developers can study to understand how task queues work

### Value Proposition

| For whom          | Value delivered                                                   |
|-------------------|-------------------------------------------------------------------|
| Junior developers | Concrete, readable reference implementation of a task queue       |
| Backend engineers | Embeddable Go library with clean API, no magic                    |
| Open source users | MIT-licensed, zero paid dependency, single binary deployment      |
| Non-technical     | Visual dashboard — see workers, tasks, and queues in realtime     |

---

## 2. Target Users

### Primary — Junior Backend Developer

- **Profile:** 0–2 years experience, learning Go or already writing Go
- **Motivation:** Understand how task queues work by reading and running real code
- **Frustration:** Existing tools (BullMQ, Sidekiq, Asynq) are complex, opaque, or language-locked
- **Technical sophistication:** Medium — can run CLI tools, understands HTTP, knows basic Go syntax
- **Use case:** Study reference, portfolio inspiration, small project background jobs

### Secondary — Solo / Indie Developer

- **Profile:** Building side projects or small products solo
- **Motivation:** Need background job processing without operational overhead of full message broker
- **Frustration:** RabbitMQ/Kafka overkill; hosted solutions require paid accounts
- **Technical sophistication:** High — comfortable with Go, Docker, self-hosting
- **Use case:** Production background job processing for small-to-medium workload

### Tertiary — Non-Technical Observer

- **Profile:** HR, PM, student, or curious non-developer
- **Motivation:** Understand "what does a task queue actually do" visually
- **Frustration:** Documentation is abstract, diagrams are static
- **Technical sophistication:** Low — cannot read code, needs visual explanation
- **Use case:** Open the dashboard URL, watch tasks move through workers in realtime

---

## 3. Problem Statement

### The Problem

Background job processing is a foundational backend pattern.
Every production application eventually needs it.

Existing solutions have one of three problems:

**Problem A — Too complex for learning:**
Asynq, BullMQ, Sidekiq are mature tools but their source code is optimized
for production, not comprehension. A junior developer reading Asynq source
will drown in abstractions before understanding the core concept.

**Problem B — Requires paid infrastructure:**
Many hosted queue solutions (Inngest, Trigger.dev free tier) require account
creation, credit card, or have usage limits that break free experimentation.

**Problem C — No visual feedback:**
All existing OSS task queues are pure code — no built-in visualizer.
Understanding worker behavior requires reading logs or external APM tools.

### Why It Matters

A developer who cannot visualize how a task queue works will:
- misuse it (wrong retry logic, wrong concurrency settings)
- over-engineer alternatives (rolling their own with channels)
- underestimate failure modes (no dead letter handling)

buruh solves all three problems simultaneously:
- readable, idiomatic Go source code
- zero paid dependency (Valkey is free + OSS)
- built-in realtime dashboard

### Measurable Impact

- A developer can run buruh locally within 5 minutes of cloning the repo
- A non-technical person can understand worker behavior within 60 seconds of opening dashboard
- A junior developer can trace a task from enqueue to completion in the source code within 30 minutes

---

## 4. Success Metrics

All metrics are measurable. Invalid metrics are excluded.

| Metric                                      | Target        | Measurement Method                        |
|---------------------------------------------|---------------|-------------------------------------------|
| Time-to-first-task (clone → task processed) | < 5 minutes   | Manual stopwatch test on fresh machine    |
| Dashboard initial load time                 | < 1 second    | Browser DevTools Network tab              |
| SSE first event delivery latency            | < 500ms       | Timestamp delta: enqueue → dashboard event|
| Worker throughput (single node)             | > 100 tasks/s | Benchmark test: 10k tasks, 10 workers     |
| Dashboard CPU usage (browser, idle)         | < 5% CPU      | Browser Task Manager during idle stream   |
| Enqueue API p95 latency                     | < 10ms        | Go benchmark test                         |
| Task state accuracy                         | 100%          | Integration test: all state transitions   |
| Retry correctness                           | 100%          | Unit test: max retry → DLQ transition     |
| Dashboard readability (non-tech user)       | Understands worker state in < 60s | Manual user test |

---

## 5. Core Capabilities (MVP)

These are SYSTEM capabilities, not UI features.

### 5.1 Task Enqueueing

- Accept task definitions: name, payload (JSON), queue name, max retries, delay
- Persist task to Valkey with assigned ID and initial state (pending)
- Return task ID synchronously to caller

### 5.2 Worker Pool Management

- Initialize N concurrent workers (goroutines) at startup
- Each worker pulls tasks from assigned queue(s)
- Worker executes registered handler function for task type
- Worker reports state transitions: idle → active → success/failed

### 5.3 Task State Machine

Tasks MUST move through defined states only:

```
pending → active → success
pending → active → failed → retrying → active (retry loop)
retrying → active → failed (max retries) → dead
```

No other transitions are valid.

### 5.4 Retry Engine

- Configurable max retry count per task
- Exponential backoff between retries: base 1s, multiplier 2x, max 60s
- Retry count tracked in task record
- On max retry exceeded: move task to Dead Letter Queue (DLQ)

### 5.5 Dead Letter Queue

- Separate queue for tasks that exceeded max retries
- Tasks in DLQ are not re-processed automatically
- DLQ visible in dashboard
- DLQ count exposed via metrics endpoint

### 5.6 HTTP API

- Enqueue endpoint: `POST /tasks`
- Task status endpoint: `GET /tasks/{id}`
- Queue stats endpoint: `GET /queues`
- Metrics endpoint: `GET /metrics`
- Health check endpoint: `GET /health`

### 5.7 Realtime SSE Stream

- Single SSE endpoint: `GET /stream`
- Broadcasts all state change events to connected dashboard clients
- Event payload: JSON with task ID, worker ID, from_state, to_state, timestamp
- Reconnect handled by browser EventSource API natively

### 5.8 Web Dashboard

- Served by engine itself at `GET /`
- Static HTML/CSS/JS — no build step, no framework
- Displays: worker lanes, queue depth bars, metric cards, event log
- Worker lane drag-to-reorder (layout only, no queue mutation)
- Queue bar panel zoom (visual only)
- No dashboard action modifies queue state

### 5.9 Handler Registration

- Developer registers handlers: `queue.Register("task-name", handlerFunc)`
- Handler signature: `func(ctx context.Context, task *Task) error`
- Unregistered task type: move to DLQ immediately with error log

---

## 6. Non-MVP Features

Explicitly deferred. Will NOT be built in v1.

| Feature                        | Reason deferred                                      |
|--------------------------------|------------------------------------------------------|
| Task scheduling (cron)         | Increases engine complexity, separate concern        |
| Task cancellation via dashboard| Requires bidirectional SSE or WebSocket              |
| Multi-node / distributed workers| Requires distributed lock — separate architecture   |
| Authentication on dashboard    | Not needed for local/trusted network deployment      |
| Task payload editing           | Dashboard is read-only observer                      |
| Plugin system                  | Premature abstraction                                |
| Mobile dashboard               | Ops tool — desktop only is defensible                |
| Persistent task history (DB)   | Valkey TTL sufficient for MVP                        |
| Priority queues                | FIFO sufficient for MVP                              |
| Rate limiting per queue        | Deferred to v2                                       |
| Webhook on task completion     | Deferred to v2                                       |
| Dashboard authentication       | Deferred to v2                                       |
| gRPC API                       | HTTP sufficient for MVP                              |
| Prometheus metrics export      | /metrics endpoint sufficient for MVP                 |

---

## 7. User Flows

### 7.1 Primary Flow — Developer Enqueues a Task

```
1. Developer imports buruh library OR runs buruh binary
2. Developer registers handler: queue.Register("send-email", handler)
3. Developer starts engine: engine.Start()
4. Developer calls: client.Enqueue("send-email", payload)
5. buruh assigns task ID, persists to Valkey, state = pending
6. Available worker picks up task, state = active
7. Worker executes handler
8. On success: state = success, task TTL set (default 24h)
9. On failure: state = failed → retry logic evaluates
10. On max retry: state = dead, moved to DLQ
```

### 7.2 Secondary Flow — Non-Technical Observer Opens Dashboard

```
1. Observer opens browser: http://localhost:8080
2. Dashboard loads (< 1s)
3. SSE connection established automatically
4. Worker lanes render — showing worker count and current state
5. Observer watches tasks appear in worker lanes (active state = amber)
6. Observer watches tasks complete (success state = green, fade out)
7. Observer sees queue depth bars rise/fall in realtime
8. Observer reads event log to understand what happened
```

### 7.3 Edge Case — Worker Handler Panics

```
1. Worker goroutine executes handler
2. Handler panics
3. buruh recover() catches panic
4. Task state → failed
5. Error message logged: "handler panic: {message}"
6. Retry logic evaluates normally
7. Worker goroutine recovers and picks next task (does not crash engine)
```

### 7.4 Edge Case — Valkey Connection Lost

```
1. Engine loses connection to Valkey
2. In-flight tasks continue executing (already in worker memory)
3. New enqueue attempts return error: "storage unavailable"
4. Engine attempts reconnect with exponential backoff
5. Dashboard SSE continues streaming (engine process still alive)
6. Dashboard topbar shows: "storage disconnected" warning
7. On reconnect: engine resumes normal operation
```

### 7.5 Edge Case — SSE Client Disconnects

```
1. Browser tab closes or network drops
2. Server detects broken SSE connection
3. Server removes client from broadcast list
4. No error, no crash
5. Browser EventSource API automatically reconnects on tab reopen
6. Dashboard receives full state snapshot on reconnect (not just delta)
```

### 7.6 Edge Case — Unregistered Task Type

```
1. Task enqueued with type "unknown-handler"
2. Worker picks up task, state = active
3. Worker looks up handler registry — not found
4. Task state → dead immediately (no retry)
5. Error logged: "no handler registered for task type: unknown-handler"
6. DLQ count incremented
7. Dashboard DLQ badge updates
```

---

## 8. High-Level Tech Stack

| Layer           | Technology              | Justification                                           |
|-----------------|-------------------------|---------------------------------------------------------|
| Language        | Go 1.22+                | Native goroutines, single binary, strong std library    |
| Broker/Storage  | Valkey 7.x              | Redis-compatible, BSD licensed, free, run local         |
| Valkey client   | `github.com/valkey-io/valkey-go` | Official client, maintained by Valkey project |
| HTTP server     | `net/http` (stdlib)     | No framework needed for 5 endpoints                     |
| Dashboard       | HTML + CSS + Vanilla JS | No build step, no framework, no npm                     |
| SSE             | `net/http` + `text/event-stream` | Native, no library needed                    |
| Configuration   | YAML file + env vars    | Standard ops pattern                                    |
| Testing         | `testing` (stdlib) + `testify` | Standard Go testing stack                      |
| CI/CD           | GitHub Actions          | Free tier, sufficient for OSS project                   |
| Binary release  | `goreleaser`            | Multi-platform binary distribution                      |

---

## 9. Technical Assumptions

### Constraints

- Single-node deployment only (v1)
- Valkey MUST be running and reachable before engine starts
- Dashboard MUST work without JavaScript frameworks
- Engine MUST compile to a single static binary
- Engine MUST run on CPU without AVX support (no AVX-dependent dependencies)
- All dependencies MUST be MIT or BSD licensed

### Scale Assumptions

- Expected workload: 1–1,000 tasks/minute
- Worker count: 1–50 workers
- Queue count: 1–10 queues
- Concurrent dashboard clients: 1–10
- No assumption of multi-tenancy
- No assumption of task payload > 1MB

### Integration Assumptions

- No authentication required in v1
- No TLS in v1 (assumed behind reverse proxy if public)
- No external metrics system in v1 (self-contained /metrics endpoint)
- No persistent task history beyond Valkey TTL
