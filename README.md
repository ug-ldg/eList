# eList

A hierarchical task management REST API built in Go, designed as a portfolio project to demonstrate backend engineering practices.

## Tech Stack

- **Go** — main language
- **PostgreSQL** — persistent storage
- **Redis** — in-memory cache
- **Chi** — HTTP router
- **pgx v5** — PostgreSQL driver with connection pooling
- **Docker** — local infrastructure

## Key Engineering Concepts Demonstrated

### Layered Architecture
Each layer has a single responsibility and only knows about the layer directly below it:

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

Handlers know nothing about SQL. The repository knows nothing about HTTP. This separation makes each layer independently testable and replaceable.

### Cache-Aside Pattern (Redis)
`GET /tasks/{id}` avoids unnecessary database queries:

1. Check Redis first
2. **HIT** → return immediately, no database query
3. **MISS** → query PostgreSQL, store result in Redis (TTL: 5 min)
4. On status update → invalidate the cache entry

The `X-Cache: HIT/MISS` response header makes cache behavior observable — useful for debugging and performance monitoring.

### Concurrent Statistics with Goroutines
`GET /stats` runs 3 independent SQL queries **in parallel** using goroutines:

```
goroutine 1 → SELECT COUNT(*) WHERE status = 'pending'  ─┐
goroutine 2 → SELECT COUNT(*) WHERE status = 'done'      ├─ run concurrently
goroutine 3 → SELECT COUNT(*) WHERE parent_id IS NULL   ─┘
                         │
                    sync.WaitGroup
                         │
                    merge results
```

A `sync.WaitGroup` waits for all goroutines to complete. A `sync.Mutex` protects concurrent writes to the shared result struct. This reduces latency from 3× to ~1× the cost of a single query.

### Self-Referencing Data Model
Tasks form a tree structure via a self-referencing foreign key:

```go
type Task struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    ParentID  *int      `json:"parent_id,omitempty"` // nil = root task
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

`ON DELETE CASCADE` in the schema ensures that deleting a parent task automatically removes all its subtasks.

## Project Structure

```
eList/
├── cmd/
│   └── api/
│       └── main.go          # Entry point — wires all layers together
├── internal/
│   ├── model/
│   │   ├── task.go          # Task struct
│   │   └── stats.go         # Stats struct
│   ├── repository/
│   │   ├── db.go            # PostgreSQL connection pool
│   │   ├── task.go          # SQL queries for tasks
│   │   └── stats.go         # Concurrent stats queries
│   ├── service/
│   │   └── task.go          # Business logic and validation
│   ├── cache/
│   │   └── task.go          # Redis cache layer
│   └── handler/
│       ├── task.go          # HTTP handlers for tasks
│       └── stats.go         # HTTP handler for stats
├── migrations/
│   └── 001_create_tasks.sql # Database schema
├── docker-compose.yml
├── .env
└── go.mod
```

## API Endpoints

| Method  | URL                      | Description                        |
|---------|--------------------------|------------------------------------|
| `POST`  | `/tasks`                 | Create a task or subtask           |
| `GET`   | `/tasks/{id}`            | Get a task by ID (cached)          |
| `GET`   | `/tasks/{id}/children`   | Get direct subtasks                |
| `PATCH` | `/tasks/{id}/status`     | Update task status                 |
| `GET`   | `/stats`                 | Get task statistics (concurrent)   |

### Request & Response Examples

**Create a root task**
```json
POST /tasks
{ "title": "My project" }
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
Status values: `pending`, `in_progress`, `done`

**Get statistics**
```json
GET /stats

{
    "total_tasks": 17,
    "pending": 12,
    "done": 5,
    "root_tasks": 4,
    "sub_tasks": 13
}
```

## Infrastructure

PostgreSQL runs on port `5433` (host) → `5432` (container) to avoid conflicts with a local PostgreSQL installation.

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
