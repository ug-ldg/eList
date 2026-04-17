# eList

A hierarchical task management REST API built in Go, designed as a portfolio project to demonstrate backend engineering practices.

## Tech Stack

- **Go** — main language
- **PostgreSQL** — persistent storage
- **Redis** — in-memory cache
- **Chi** — HTTP router
- **pgx v5** — PostgreSQL driver with connection pooling
- **Docker** — local infrastructure

## Architecture

The project follows a layered architecture where each layer has a single responsibility:

```
HTTP Request
     │
     ▼
 Handler        → parses HTTP input, writes HTTP output
     │
     ▼
 Service        → business logic and validation
     │
     ├──────────────────────┐
     ▼                      ▼
Repository               Cache
(PostgreSQL)             (Redis)
```

Each layer only knows about the layer directly below it. Handlers know nothing about SQL; the repository knows nothing about HTTP.

## Project Structure

```
eList/
├── cmd/
│   └── api/
│       └── main.go          # Entry point — wires all layers together
├── internal/
│   ├── model/
│   │   └── task.go          # Task struct definition
│   ├── repository/
│   │   ├── db.go            # PostgreSQL connection pool
│   │   └── task.go          # SQL queries
│   ├── service/
│   │   └── task.go          # Business logic and validation
│   ├── cache/
│   │   └── task.go          # Redis cache layer
│   └── handler/
│       └── task.go          # HTTP handlers
├── migrations/
│   └── 001_create_tasks.sql # Database schema
├── docker-compose.yml
├── .env
└── go.mod
```

## Data Model

A task can contain subtasks, forming a tree structure. The `parent_id` field references another task in the same table (self-referencing foreign key).

```go
type Task struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    ParentID  *int      `json:"parent_id,omitempty"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

Status values: `pending`, `in_progress`, `done`

## API Endpoints

| Method  | URL                      | Description              |
|---------|--------------------------|--------------------------|
| `POST`  | `/tasks`                 | Create a task            |
| `GET`   | `/tasks/{id}`            | Get a task by ID         |
| `GET`   | `/tasks/{id}/children`   | Get subtasks             |
| `PATCH` | `/tasks/{id}/status`     | Update task status       |

### Examples

**Create a root task**
```json
POST /tasks
{ "title": "My task" }
```

**Create a subtask**
```json
POST /tasks
{ "title": "My subtask", "parent_id": 1 }
```

**Update status**
```json
PATCH /tasks/1/status
{ "status": "in_progress" }
```

## Caching Strategy

`GET /tasks/{id}` uses a **cache-aside** pattern:

1. Check Redis first
2. If found (HIT) → return immediately, no database query
3. If not found (MISS) → query PostgreSQL, store result in Redis with a 5-minute TTL
4. On status update → invalidate the cache entry for that task

The response includes an `X-Cache: HIT` or `X-Cache: MISS` header so the cache behavior is observable.

## Infrastructure

PostgreSQL runs on port `5433` (host) mapped to `5432` (container) to avoid conflicts with a locally installed PostgreSQL instance.

Redis runs on the default port `6379`.

## Getting Started

**Prerequisites:** Docker, Go 1.21+

```bash
# Start infrastructure
docker-compose up -d

# Apply database migrations (PowerShell)
Get-Content migrations/001_create_tasks.sql | docker exec -i elist-postgres-1 psql -U elist_user -d elist_db

# Run the API
go run ./cmd/api
```

The API will be available at `http://localhost:8080`.
