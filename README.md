# GoFlow — Distributed Task Processing and Intelligent Worker Orchestration Platform

GoFlow is a final-year software engineering project that aims to design and evaluate a scalable distributed task-processing platform implemented primarily in Go.

## Current phase

Phase 2 adds a local concurrent worker pool to the Phase 1 Gin REST API and concurrency-safe in-memory task repository. Tasks are accepted immediately and processed asynchronously by worker goroutines in the same application process.

```text
Client -> Gin handler -> Task service -> In-memory repository
                              |
                              v
                    Buffered task channel
                              |
                 +------------+------------+
                 v            v            v
              worker-1     worker-2     worker-3
                 |            |            |
                 +------ simulated processing -----+
                              |
                              v
                  Service updates task status
```

This separation keeps HTTP, business rules, queueing, and persistence independent. The queue contract can later be backed by Redis, and a PostgreSQL repository can replace the in-memory store without rewriting HTTP handlers.

## Worker pool and asynchronous processing

`POST /api/v1/tasks` validates and saves a task with `queued` status, submits only its ID to a buffered channel, and returns `201 Created` without waiting. Three worker goroutines listen to the same channel. A worker changes its task to `running`, simulates work for approximately three seconds, and then changes it to `completed`. Processing errors or shutdown cancellation produce `failed` status.

A buffered channel allows a limited number of jobs to wait without requiring a worker to receive each job at the exact instant it is submitted. Its fixed capacity also provides backpressure: a full or closed queue produces a `503 Service Unavailable` response rather than blocking the HTTP request indefinitely.

The application listens for Ctrl+C, `SIGINT`, and `SIGTERM`. Shutdown first stops the HTTP server from accepting new requests, closes the queue to new submissions, cancels worker processing through `context.Context`, and uses `sync.WaitGroup` to wait for worker goroutines to exit.

## Planned architecture

Later phases will introduce Go worker pools, goroutines and channels, Redis, PostgreSQL, distributed workers, task retries, priority scheduling, worker heartbeats, intelligent/adaptive scheduling, real-time monitoring, Docker, Prometheus/Grafana, and performance benchmarking.

## Running the project

Requires Go 1.27 or later.

```bash
go mod download
go run ./cmd/api
```

The API listens on `http://localhost:8080`.

## API

| Method | Path | Purpose | Success |
|---|---|---|---|
| `GET` | `/health` | Check API health | `200` |
| `POST` | `/api/v1/tasks` | Create a queued task | `201` |
| `GET` | `/api/v1/tasks` | List tasks | `200` |
| `GET` | `/api/v1/tasks/:id` | Get a task | `200` |
| `DELETE` | `/api/v1/tasks/:id` | Delete a task | `204` |

```bash
curl http://localhost:8080/health

curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"type":"email","payload":"Send welcome email to the new user","priority":2}'

curl http://localhost:8080/api/v1/tasks
curl http://localhost:8080/api/v1/tasks/<task-id>
curl -X DELETE http://localhost:8080/api/v1/tasks/<task-id>
```

`type` and `payload` are required, and `priority` must be from 1 to 5. New tasks always begin with `queued` status; the client cannot override it.

Errors have a consistent structure:

```json
{"error":{"code":"TASK_NOT_FOUND","message":"Task was not found"}}
```

## Testing

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Current limitations

- The queue and repository are in memory, so tasks are lost on restart.
- Workers and the queue exist only inside one application process.
- Processing is simulated; there is no real task-specific work yet.
- There is no distributed queue or persistent database.
- Authentication, pagination, observability, and deployment packaging are not included.

The recommended next phase focuses on task retries, timeouts, cancellation, and priority queue behavior before introducing Redis.
