# TaskFlow Product Requirements Document

## 1. Overview

TaskFlow is a minimal but production-minded task management application for small teams. Users can register, log in, create projects, add tasks to projects, assign tasks to themselves or other users, and track progress by status and priority.

This implementation is intentionally backend-first. The primary goal is to deliver a reliable Go + PostgreSQL REST API with clean auth, authorization, migrations, Docker-based local setup, and a lightweight React frontend that demonstrates the core user flows without over-investing in visual complexity.

## 2. Product Goals

- Demonstrate a complete end-to-end task management flow.
- Meet every required backend and infrastructure constraint from the assignment.
- Include backend bonus features as part of the baseline scope:
  - Pagination on list endpoints
  - `GET /projects/:id/stats`
  - At least 3 integration tests
- Provide a small but presentable React UI for reviewer demos.

## 3. Non-Goals

- Real-time collaboration
- Drag-and-drop task management
- Dark mode
- Comments, attachments, notifications, or activity history
- Organization/workspace concepts beyond a single-user-owned project model

## 4. Primary Users

- Reviewer evaluating backend correctness, data modeling, and infrastructure quality
- Reviewer validating that the UI can exercise the major workflows without relying solely on API tooling

## 5. Core User Flows

### 5.1 Authentication
- A new user can register with name, email, and password.
- A returning user can log in with email and password.
- On successful authentication, the system returns a JWT access token valid for 24 hours.
- The frontend persists auth state across refreshes.

### 5.2 Project Management
- An authenticated user can view all projects they own or participate in through tasks.
- An authenticated user can create a project and automatically becomes the project owner.
- A user can view project details and its tasks.
- Only the project owner can update or delete the project.

### 5.3 Task Management
- A user with access to a project can list its tasks.
- Task listing supports filtering by status and assignee.
- Task listing supports pagination.
- A user with access to a project can create tasks in that project.
- A user with access to a project can update task title, description, status, priority, assignee, and due date.
- A task can be deleted only by the project owner or the task creator.

### 5.4 Reporting
- A user with access to a project can fetch task statistics for that project.
- Statistics include counts by status and counts by assignee.

## 6. Functional Requirements

### 6.1 Backend
- Use Go for the API implementation.
- Use PostgreSQL for persistence.
- Use SQL migrations for schema management.
- Hash passwords with bcrypt using cost 12 or higher.
- Protect all non-auth routes with bearer-token authentication.
- Distinguish `401 Unauthorized` from `403 Forbidden`.
- Return structured JSON error responses.
- Emit structured logs.
- Shut down gracefully on `SIGTERM`.

### 6.2 Frontend
- Use React with TypeScript.
- Use React Router for navigation.
- Provide routes for login, register, projects list, and project detail.
- Provide create/edit task UI via modal or side panel.
- Show loading, error, and empty states explicitly.
- Persist JWT-based auth state across refreshes.
- Support optimistic task status updates and rollback on failure.
- Work cleanly at 375px and 1280px widths.

### 6.3 Infrastructure
- Root `docker-compose.yml` must start PostgreSQL, API server, and frontend.
- `docker compose up` should work with no manual setup beyond copying `.env.example`.
- PostgreSQL credentials and JWT secret must be environment-driven.
- Backend Docker image must use a multi-stage build.
- Migrations should run automatically on backend container startup.
- Seed data must create at least:
  - 1 known test user
  - 1 project
  - 3 tasks with different statuses

## 7. API Scope

### Auth
- `POST /auth/register`
- `POST /auth/login`

### Projects
- `GET /projects`
- `POST /projects`
- `GET /projects/:id`
- `PATCH /projects/:id`
- `DELETE /projects/:id`
- `GET /projects/:id/stats`

### Tasks
- `GET /projects/:id/tasks`
- `POST /projects/:id/tasks`
- `PATCH /tasks/:id`
- `DELETE /tasks/:id`

## 8. Success Criteria

- A reviewer can clone the repo, copy `.env.example` to `.env`, run `docker compose up`, and use the app immediately.
- Seed credentials work without additional setup.
- Auth, authorization, CRUD operations, filtering, pagination, and stats work end to end.
- Backend responses follow the required status-code and JSON conventions.
- The frontend is stable, responsive, and free from blank/error-prone states in the demo flow.
- At least 3 integration tests verify core backend behavior.

## 9. Quality Bar

- Clear separation of concerns in backend modules
- Hand-written SQL with explicit migrations
- Predictable error handling and validation
- Sensible indexing and relational integrity
- Minimal but coherent UI with visible state transitions
- README that explains setup, architecture, API usage, credentials, and tradeoffs honestly

## 10. Risks And Mitigations

- Risk: Overbuilding the frontend and under-delivering backend polish
  - Mitigation: Keep the UI intentionally minimal and centered on demoing API-backed flows.
- Risk: Ambiguity around project membership and task deletion permissions
  - Mitigation: Define access rules explicitly in the architecture and enforce them consistently in the service layer.
- Risk: Docker startup friction
  - Mitigation: Automate migrations and seeding on backend startup and document all env vars up front.

## 11. Out-of-Scope Follow-Ups

If more time were available, likely next steps would be invitations/collaboration, richer audit history, stronger observability, and realtime task updates.
