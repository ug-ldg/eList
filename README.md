# eList

A hierarchical task management REST API built in Go, designed as a portfolio project to demonstrate backend engineering practices.

## Tech Stack

- **Go** — main language
- **PostgreSQL** — persistent storage
- **Redis** — in-memory cache
- **Chi** — HTTP router + middleware
- **pgx v5** — PostgreSQL driver with connection pooling
- **Docker** — local infrastructure
- **OAuth 2.0** — Google authentication (`golang.org/x/oauth2`)
- **JWT** — stateless authentication tokens (`golang-jwt/jwt`)

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

### OAuth 2.0 + JWT Authentication
`GET /auth/google` redirects to Google. After consent, Google calls back `/auth/google/callback`:

1. Exchange the OAuth code for a Google access token
2. Fetch the user's profile (email, name, provider ID)
3. Upsert the user in PostgreSQL (`ON CONFLICT DO UPDATE`)
4. Generate a signed JWT containing `user_id` (TTL: 24h)
5. Return the JWT to the client

All task endpoints require `Authorization: Bearer <token>`. A Chi middleware validates the JWT and injects `user_id` into the request context — handlers extract it and pass it down to the service and repository layers.

Each task is owned by a user via a `user_id` foreign key. All SQL queries filter by `user_id`, so users can only see and modify their own tasks. The Redis cache key also includes `user_id` (`task:{userID}:{taskID}`) to prevent cross-user cache leaks.

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

### Recursive Tree Traversal
`GET /tasks/{id}/tree` returns a full nested tree using a PostgreSQL `WITH RECURSIVE` CTE — a single SQL query traverses the entire hierarchy regardless of depth:

```
Task 1 (root)
├── Task 2
│   ├── Task 4
│   └── Task 5
└── Task 3
    └── Task 6
```

The flat SQL result is then assembled into a nested structure in Go using a map for O(n) tree construction.

## Project Structure

```
eList/
├── cmd/
│   └── api/
│       └── main.go          # Entry point — wires all layers together
├── internal/
│   ├── model/
│   │   ├── task.go          # Task, TaskNode, Breadcrumb structs
│   │   ├── user.go          # User struct
│   │   └── stats.go         # Stats struct
│   ├── repository/
│   │   ├── db.go            # PostgreSQL connection pool
│   │   ├── task.go          # SQL queries for tasks
│   │   ├── user.go          # User upsert
│   │   └── stats.go         # Concurrent stats queries
│   ├── service/
│   │   └── task.go          # Business logic and validation
│   ├── cache/
│   │   └── task.go          # Redis cache layer (key: task:{userID}:{id})
│   ├── auth/
│   │   └── jwt.go           # JWT generation and validation
│   ├── middleware/
│   │   └── auth.go          # Bearer token validation + userID injection
│   └── handler/
│       ├── task.go          # HTTP handlers for tasks
│       ├── auth.go          # Google OAuth handlers
│       └── stats.go         # HTTP handler for stats
├── migrations/
│   ├── 001_create_tasks.sql # Database schema
│   ├── 002_add_users.sql    # Users table + user_id FK on tasks
│   └── 003_add_note_icon.sql # note (TEXT) and icon (VARCHAR) columns on tasks
├── docker-compose.yml
├── .env
└── go.mod
```

## API Endpoints

### Authentication (public)

| Method | URL                        | Description                          |
|--------|----------------------------|--------------------------------------|
| `GET`  | `/auth/google`             | Redirect to Google OAuth             |
| `GET`  | `/auth/google/callback`    | OAuth callback — returns JWT         |

### Tasks & Stats (requires `Authorization: Bearer <token>`)

| Method   | URL                      | Description                        |
|----------|--------------------------|------------------------------------|
| `POST`   | `/tasks`                 | Create a task or subtask           |
| `GET`    | `/tasks/{id}`            | Get a task by ID (cached)          |
| `GET`    | `/tasks/{id}/children`   | Get direct subtasks                |
| `PATCH`  | `/tasks/{id}/status`     | Update task status                 |
| `DELETE` | `/tasks/{id}`            | Delete a task and its subtasks     |
| `GET`    | `/tasks/{id}/tree`       | Get full nested task tree          |
| `GET`    | `/tasks/{id}/ancestors`  | Get breadcrumb path to root        |
| `GET`    | `/tasks`                 | Get root tasks (no parent)         |
| `PATCH`  | `/tasks/{id}/parent`     | Move a task to a new parent        |
| `PATCH`  | `/tasks/{id}`            | Update title, status, note, icon   |
| `GET`    | `/stats`                 | Get task statistics (concurrent)   |

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

**Get full tree**
```json
GET /tasks/1/tree

{
    "id": 1,
    "title": "My project",
    "status": "pending",
    "created_at": "...",
    "updated_at": "...",
    "children": [
        {
            "id": 2,
            "title": "Phase 1",
            "status": "pending",
            "children": [
                { "id": 4, "title": "Task A", "status": "done", "children": [] }
            ]
        }
    ]
}
```

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
Get-Content migrations/002_add_users.sql | docker exec -i elist-postgres-1 psql -U elist_user -d elist_db

# Run the API
go run ./cmd/api
```

Add the following to your `.env` file:

```env
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
JWT_SECRET=a-long-random-string
```

**Authentication flow:**
1. Open `http://localhost:8080/auth/google` in your browser
2. Sign in with Google
3. Copy the returned `token`
4. Add `Authorization: Bearer <token>` to all subsequent requests

The API will be available at `http://localhost:8080`.
