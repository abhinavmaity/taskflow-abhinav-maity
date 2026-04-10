# TaskFlow Technical Architecture

## 1. Architecture Summary

TaskFlow will be implemented as a modular monolith optimized for a take-home assignment: fast to ship, easy to review, and structured enough to discuss design decisions clearly.

- Backend: Go REST API
- Database: PostgreSQL
- Frontend: React + TypeScript
- Local infrastructure: Docker Compose

The system is intentionally split into clear layers:

- Transport layer: HTTP routing, request parsing, response formatting, auth middleware
- Service layer: business rules, authorization checks, validation orchestration
- Repository layer: hand-written SQL queries against PostgreSQL
- Shared infrastructure: config, logging, DB connection, migrations, seeding, graceful shutdown

## 2. Repository Layout

```text
taskflow/
  docs/
  backend/
    cmd/api/
    internal/
      auth/
      projects/
      tasks/
      platform/
    migrations/
    seed/
  frontend/
    src/
      app/
      components/
      features/
      lib/
  docker-compose.yml
  .env.example
  README.md
```

## 3. Backend Design

### 3.1 Stack Choices

- Router: `chi`
- Database driver/pool: `pgx/v5` + `pgxpool`
- JWT: `golang-jwt/jwt/v5`
- Password hashing: `golang.org/x/crypto/bcrypt`
- Logging: `log/slog`
- Migrations: `golang-migrate`

These choices keep the stack lightweight, explicit, and idiomatic for a small Go service.

### 3.2 Module Boundaries

#### Auth
- Register user
- Log in user
- Create JWT with `user_id` and `email` claims
- Validate password and token

#### Projects
- List accessible projects
- Create project
- Fetch project details and tasks
- Update/delete project for owners only
- Fetch project-level task stats

#### Tasks
- List tasks with filters and pagination
- Create task inside a project
- Update task fields
- Delete task if requester is project owner or task creator

### 3.3 Request Lifecycle

1. HTTP request enters router.
2. Middleware enforces JSON conventions, request ID propagation, logging, and auth where required.
3. Handler decodes input and delegates to a service.
4. Service validates business rules and authorization.
5. Repository executes SQL queries.
6. Handler writes JSON success or standardized error response.

## 4. Data Model

### 4.1 Tables

#### `users`
- `id uuid primary key`
- `name text not null`
- `email text not null unique`
- `password text not null`
- `created_at timestamptz not null default now()`

#### `projects`
- `id uuid primary key`
- `name text not null`
- `description text null`
- `owner_id uuid not null references users(id)`
- `created_at timestamptz not null default now()`

#### `tasks`
- `id uuid primary key`
- `title text not null`
- `description text null`
- `status task_status not null`
- `priority task_priority not null`
- `project_id uuid not null references projects(id) on delete cascade`
- `assignee_id uuid null references users(id)`
- `created_by uuid not null references users(id)`
- `due_date date null`
- `created_at timestamptz not null default now()`
- `updated_at timestamptz not null default now()`

### 4.2 Enums

- `task_status`: `todo`, `in_progress`, `done`
- `task_priority`: `low`, `medium`, `high`

### 4.3 Indexing

- `users(email)` unique index
- `projects(owner_id)`
- `tasks(project_id)`
- `tasks(assignee_id)`
- `tasks(created_by)`
- `tasks(project_id, status)`
- `tasks(project_id, assignee_id)`

### 4.4 Access Model

- Project access is granted if the user:
  - owns the project, or
  - is assigned to any task in the project, or
  - created any task in the project
- Project mutation is owner-only.
- Task creation and updates are allowed for any user with project access.
- Task deletion is allowed for the project owner or the task creator.

## 5. API Contract

### 5.1 Authentication

#### `POST /auth/register`
- Input: `name`, `email`, `password`
- Output: JWT token and basic user object

#### `POST /auth/login`
- Input: `email`, `password`
- Output: JWT token and basic user object

### 5.2 Projects

#### `GET /projects`
- Returns projects accessible to the current user
- Supports `page` and `limit`

#### `POST /projects`
- Creates a project with the authenticated user as owner

#### `GET /projects/:id`
- Returns project metadata
- Returns project tasks
- Returns `available_assignees` for assignment UI

#### `PATCH /projects/:id`
- Owner-only update of `name` and `description`

#### `DELETE /projects/:id`
- Owner-only delete
- Returns `204 No Content`

#### `GET /projects/:id/stats`
- Returns:
  - task counts by status
  - task counts by assignee

### 5.3 Tasks

#### `GET /projects/:id/tasks`
- Returns task list for the project
- Supports `status`, `assignee`, `page`, and `limit`

#### `POST /projects/:id/tasks`
- Creates a task under the project

#### `PATCH /tasks/:id`
- Supports updates to title, description, status, priority, assignee, and due date

#### `DELETE /tasks/:id`
- Allowed for project owner or task creator
- Returns `204 No Content`

### 5.4 Response Conventions

- All non-204 responses use `application/json`
- Validation failure:

```json
{
  "error": "validation failed",
  "fields": {
    "email": "is required"
  }
}
```

- Unauthenticated:

```json
{ "error": "unauthorized" }
```

- Forbidden:

```json
{ "error": "forbidden" }
```

- Not found:

```json
{ "error": "not found" }
```

### 5.5 Pagination Shape

List endpoints will return:

```json
{
  "tasks": [],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

The same `pagination` object shape will be used for project lists.

## 6. Frontend Architecture

### 6.1 Scope

The frontend is a reviewer-facing demo surface, not the primary complexity center. It should remain minimal but complete enough to validate the backend.

Routes:
- `/login`
- `/register`
- `/projects`
- `/projects/:id`

### 6.2 Stack

- React
- TypeScript
- Vite
- React Router
- TanStack Query
- MUI

### 6.3 UI Capabilities

- Auth forms with client-side validation
- Protected routes
- Persistent auth via local storage
- Project list and create flow
- Project detail with task list, filters, and stats
- Task create/edit modal
- Loading, error, and empty states
- Optimistic task status updates with rollback

## 7. Infrastructure Design

### 7.1 Docker Compose

Services:
- `db`: PostgreSQL with healthcheck
- `backend`: Go API container
- `frontend`: production frontend container serving built assets

Startup expectations:
- `db` becomes healthy
- `backend` runs migrations automatically
- `backend` runs idempotent seeding automatically
- `backend` starts the HTTP server
- `frontend` serves the application on port `3000`

### 7.2 Environment Variables

Required variables in `.env.example`:
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_PORT`
- `DATABASE_URL`
- `JWT_SECRET`
- `API_PORT`
- `FRONTEND_PORT`
- `CORS_ORIGIN`
- `SEED_ENABLED`

### 7.3 Backend Container

- Multi-stage Docker build
- Builder image compiles static Go binary
- Runtime image contains only the binary, migrations/seed assets, and minimal runtime dependencies

## 8. Migrations And Seeding

### 8.1 Migrations

- SQL migrations only
- Every migration includes both up and down scripts
- Initial migrations create enums, tables, constraints, and indexes

### 8.2 Seed Strategy

Seeding must be idempotent and safe to run on container start.

Seed data includes:
- Primary user: `test@example.com / password123`
- Secondary user for assignment demos
- 1 project owned by the primary user
- 3 tasks with distinct statuses and mixed assignees

## 9. Testing Strategy

### 9.1 Integration Tests

At minimum:
- Auth flow: register/login and protected access
- Authz flow: confirm `401` vs `403`
- Task flow: create, update, filter, paginate, fetch stats, and delete rules

### 9.2 Manual Verification

- Start system with Docker
- Log in with seed credentials
- Load projects
- Open project detail
- Create and update tasks
- Filter tasks
- Check stats rendering
- Refresh browser and confirm auth persistence

## 10. Key Tradeoffs

- Modular monolith over heavier architecture:
  - Faster to build and easier to review
  - Still preserves clean separation of concerns
- Hand-written SQL over ORM:
  - More explicit schema and queries
  - Better alignment with migration and data-model evaluation criteria
- Minimal React UI over richer frontend feature work:
  - Keeps implementation energy on backend correctness and infrastructure
  - Still gives reviewers a clear visual demo path
