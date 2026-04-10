# TaskFlow Milestone-Based Implementation Plan

## Summary

- Build TaskFlow in dependency order: backend foundation first, domain/API second, verification third, frontend demo fourth, and Docker/README/final QA last.
- Treat backend bonus scope as required baseline: pagination, `GET /projects/:id/stats`, and at least 3 integration tests.
- Define each milestone by deliverables and exit criteria so implementation can progress without hidden gaps.

## Milestones

### Milestone 0: Project Skeleton And Tooling

- Create the monorepo structure under `backend/`, `frontend/`, and root infra files.
- Initialize the Go module, React app, shared env strategy, and root-level docs/README placeholders.
- Lock library choices already defined in the architecture:
  - `chi`
  - `pgxpool`
  - `golang-jwt/jwt/v5`
  - `bcrypt`
  - `slog`
  - `golang-migrate`
  - React
  - Vite
  - React Router
  - TanStack Query
  - MUI
- Exit criteria:
  - Repo layout matches the architecture doc.
  - `.env.example` lists every required variable.
  - No unresolved tooling decisions remain.

### Milestone 1: Database Schema, Migrations, And Seed Data

- Create SQL migrations for enums, tables, constraints, indexes, and down migrations.
- Include `tasks.created_by` and `ON DELETE CASCADE` from projects to tasks.
- Add idempotent seed data for 2 users, 1 project, and 3 tasks with different statuses.
- Use `test@example.com / password123` as the primary reviewer credential.
- Exit criteria:
  - Every schema change has matching up/down migration files.
  - Schema matches the PRD and technical architecture exactly.
  - Seed data supports assignment-to-other-user demos and delete-permission demos.

### Milestone 2: Backend Platform And Cross-Cutting Concerns

- Implement config loading, DB connection, router setup, middleware, structured logging, JSON response helpers, and graceful shutdown.
- Add auth middleware that validates bearer tokens and injects current user context.
- Centralize error handling for `400`, `401`, `403`, `404`, and internal errors with standardized JSON bodies.
- Exit criteria:
  - Server boots cleanly with env-driven config.
  - Protected routes distinguish unauthenticated from unauthorized requests.
  - All responses follow JSON conventions.

### Milestone 3: Auth Module

- Implement `POST /auth/register` and `POST /auth/login`.
- Hash passwords with bcrypt cost 12 or higher.
- Issue 24-hour JWTs containing `user_id` and `email`.
- Return token plus minimal user payload from auth endpoints.
- Exit criteria:
  - Registration creates a valid user record.
  - Login succeeds with valid credentials and fails cleanly otherwise.
  - JWT-protected endpoints accept valid tokens and reject invalid or missing tokens with `401`.

### Milestone 4: Projects And Tasks Domain APIs

- Implement repositories, services, and handlers for projects and tasks in the modular monolith structure.
- Enforce the access model exactly:
  - Project access: owner, any task assignee in project, or any task creator in project
  - Project update/delete: owner only
  - Task create/update: any user with project access
  - Task delete: project owner or task creator
- Implement the API surface:
  - `GET /projects`
  - `POST /projects`
  - `GET /projects/:id`
  - `PATCH /projects/:id`
  - `DELETE /projects/:id`
  - `GET /projects/:id/stats`
  - `GET /projects/:id/tasks`
  - `POST /projects/:id/tasks`
  - `PATCH /tasks/:id`
  - `DELETE /tasks/:id`
- Lock response contracts:
  - List endpoints return `{ projects|tasks, pagination }`
  - `GET /projects/:id` returns project data, tasks, and `available_assignees`
  - `GET /projects/:id/stats` returns `by_status` and `by_assignee`
  - Delete endpoints return `204 No Content`
- Exit criteria:
  - Core CRUD works end to end through the service layer.
  - Pagination and filters work on list endpoints.
  - Authorization rules are enforced consistently on every mutation.

### Milestone 5: Integration Tests And API Verification

- Add at least 3 integration tests that exercise real HTTP handlers against a test database.
- Required scenarios:
  - Auth happy path: register/login and protected access
  - Auth distinction: missing token gives `401`, forbidden mutation gives `403`
  - Task flow: create, update, filter, paginate, stats, and delete-permission checks
- Add a lightweight API collection or examples if useful for manual review, but tests remain the primary backend proof.
- Exit criteria:
  - Test suite covers the required scenarios and passes reliably.
  - Tests prove the most failure-prone authz and task flows, not just happy paths.

### Milestone 6: Minimal React Demo Frontend

- Build the frontend after backend contracts are stable.
- Implement routes:
  - `/login`
  - `/register`
  - `/projects`
  - `/projects/:id`
- Add auth persistence, protected routing, loading/error/empty states, navbar, project creation, project detail, task create/edit modal, task filters, stats display, and optimistic task status changes with rollback.
- Keep the UI minimal and presentable rather than feature-rich.
- Exit criteria:
  - A reviewer can complete the main flows without API tooling.
  - No blank screens, obvious state bugs, or console errors in the production build.
  - Layout works at `375px` and `1280px`.

### Milestone 7: Docker, Startup Automation, And README

- Create root `docker-compose.yml` for PostgreSQL, backend, and frontend.
- Use a multi-stage backend Dockerfile and production-ready frontend container build.
- Ensure backend container startup runs migrations automatically, then seed logic, then the API server.
- Write README sections required by the assignment:
  - Overview
  - Architecture decisions
  - Local run steps
  - Migration behavior
  - Test credentials
  - API reference
  - Honest tradeoffs and future work
- Exit criteria:
  - `docker compose up` works from a fresh clone after copying `.env.example` to `.env`.
  - README is sufficient for a reviewer with Docker and nothing else installed.
  - No secrets are hardcoded; JWT secret is env-only.

### Milestone 8: Final Acceptance And Submission Hardening

- Run the full manual smoke test using seeded credentials.
- Confirm required endpoints, structured errors, stats, pagination, responsive UI, and graceful shutdown behavior.
- Check that migrations, seeding, Docker flow, README, and test coverage satisfy all assignment disqualifier conditions.
- Exit criteria:
  - All rubric-critical requirements are explicitly verified.
  - Nothing required is left as a hidden follow-up or undocumented assumption.

## Test Plan

### Backend Integration

- Register and login return a JWT and valid user payload.
- Invalid or missing bearer token returns `401`.
- Authenticated but disallowed project/task mutation returns `403`.
- Project and task CRUD succeed for authorized users.
- Task list filters and pagination return correct slices and metadata.
- Project stats return correct counts by status and assignee.

### Frontend Verification

- Login persists across refresh.
- Protected routes redirect unauthenticated users to `/login`.
- Project detail renders tasks, filters, stats, and task form states.
- Optimistic task status updates revert on API failure.

### Infrastructure Verification

- Fresh `docker compose up` initializes the database, runs migrations, seeds data, and serves frontend and API without manual intervention.

## Assumptions And Defaults

- Build order is backend-first; frontend work starts only after API contracts are implemented and tested.
- The React app remains intentionally minimal; no frontend bonus features are included.
- `GET /projects` and `GET /projects/:id/tasks` both support `page` and `limit`.
- `GET /projects/:id` includes embedded tasks plus `available_assignees`, so no separate users endpoint is required.
- Seed data includes 2 users so assignment and permission scenarios are demonstrable.
- The implementation favors a clean modular monolith over additional abstraction unless a concrete gap appears during development.
