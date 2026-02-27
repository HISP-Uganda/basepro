# Desktop Skeleton App (Wails UI + Gin API Backend)
## Authoritative Requirements Specification
## Phase S1 – Skeleton/Foundation

Last Updated: 2026-02-27

---

# 0. Execution Rules (Global)
These rules apply to every milestone and every change.

1) **Milestone gating**
- No milestone may start until the previous milestone is fully complete.
- “Complete” means: all required work is done, **all tests pass (including route tests)**, and `docs/status.md` is updated.

2) **Always read requirements first**
- Before implementing any milestone, the agent must read this `docs/requirements.md`.
- The agent must treat it as the single source of truth.

3) **Status + prompts hygiene**
- At the end of each successful milestone, update `docs/status.md` with what changed and what is now complete.
- Keep a copy of milestone prompts under `docs/prompts/` for traceability **but do not commit prompts**.
  - (Recommended: add `docs/prompts/` to `.gitignore`.)

4) **Commit discipline**
- At the end of every successful milestone, the agent must propose an appropriate **git commit message** and prompt the user to commit.
- Do not suggest committing if tests are failing.

---

# 1. Vision
Build a professional desktop application shell (Wails) that never connects to the database directly.
All data access must happen through a Gin HTTP API backend.

This skeleton must be reusable for multiple future business apps, not only HR.

---

# 2. Target Architecture

## 2.1 Components

### A) Desktop Client (Wails)
- Wails v2 + Go 1.22+
- Frontend: React + TypeScript + MUI v5+
- TanStack Router + TanStack Query
- MUI X Data Grid (advanced features required; see section 3.7)
- Responsibilities:
  - UI shell, navigation, theming
  - login + refresh token flow
  - local settings storage (API base URL and app preferences)
  - HTTP client wrapper for calling the backend API
  - file dialogs (Save As / Open) via Wails bindings if needed later

### B) Backend API (Gin)
- Go 1.22+
- Gin REST API (/api/v1)
- SQLX for DB access
- golang-migrate with up/down SQL files
- Viper-based config file loading with hot reload (see section 4.7)
- PostgreSQL 13+
- Responsibilities:
  - Authentication (JWT access + refresh rotation)
  - Optional API-token authentication (service/machine use)
  - RBAC: roles + permissions + scoped permissions
  - User management
  - Audit logging (writes to DB)
  - Future domain modules (employees, payroll, etc.)

### C) Optional Web Frontend (React)
Because Wails apps are not a web deployment target, the system must allow a **separate web frontend**
that talks to the same Gin API.
- Reuse as much UI code as practical (shared packages allowed later).
- Web frontend must support the same auth flow and theming behavior, as feasible.

---

# 3. Desktop Client Requirements (Wails)

## 3.1 First-Run Setup (API Base URL)
Before login is possible, the app must allow configuration of:
- API Base URL (e.g., http://127.0.0.1:8080 or https://api.example.com)
- Optional: “API Token” (for environments using token auth instead of username/password)
- Optional: request timeout (seconds)

Storage:
- Persist locally on the machine in an app config file under OS app data dir.
- Do not store secrets in logs.
- MVP: local file with restrictive permissions (best-effort).

## 3.2 Authentication (JWT)
- Login UI exactly as previously defined in the HR requirements (small centered card; no shell visible pre-login).
- Login options:
  - Username + Password (standard)
  - (Optional mode) API Token login (if configured)
- JWT access token stored in memory.
- Refresh token stored in a secure-ish local store (MVP: local file or localStorage; choose one consistent approach).
- Auto-refresh:
  - If access token expires, attempt refresh.
  - If refresh token is invalid/expired/reused, force logout and show:
    "Session expired. Please log in again."

## 3.3 App Shell UI (MUI)
Must look like a sleek MUI admin template:
- Left side navigation (Drawer):
  - supports collapse (mini variant) and mobile overlay behavior
- Top AppBar:
  - user avatar/menu
  - current section title
  - quick access to Appearance and Settings
- Main content area (Outlet)
- Footer (always visible on authenticated pages)

## 3.4 Theme System + Admin Palette
- Theme mode: light | dark | system
- Accent/palette selection:
  - Provide a UI to choose from multiple presets (admin-dashboard feel)
  - Persist locally
- Must be smooth and consistent across pages.

## 3.5 Routes (Skeleton)
Unauthenticated:
- /setup (API base URL + optional API token)
- /login

Authenticated:
- /dashboard (placeholder content)
- /settings (local + server-backed settings later)
- /users (scaffold; only visible if allowed)
- /audit (scaffold; only visible if allowed)

Unknown routes must show a NotFound component (no router warnings).

## 3.6 Settings Page
Settings must include at minimum:
- API Base URL (editable, with Test Connection)
- Auth mode selector (Username/Password vs API Token)
- Appearance: mode + accent preset (can link to the Appearance dialog)
- Placeholder “About” section (version/build)

## 3.7 MUI Data Grid (Advanced)
The app must use **MUI X Data Grid** with advanced capabilities (where applicable):
- Server-side pagination
- Column filtering and sorting (server-driven where needed)
- Column visibility, reordering
- Column pinning (left/right) if available
- Export (CSV at minimum; advanced export when available)
- Density selector
- Persist user table preferences locally (per table)

Note:
- If the project uses DataGridPro/Premium features, ensure the team handles licensing appropriately.
- Skeleton must be structured so tables can “upgrade” to Pro/Premium without refactor.

---

# 4. Backend API Requirements (Gin)

## 4.1 Migrations
- Use golang-migrate with up/down SQL files.
- Use SQLX DB pool.
- Provide a clear migration runner command (Makefile required; see section 7).

## 4.2 Authentication
### 4.2.1 Username/Password + JWT
Endpoints (suggested):
- POST /api/v1/auth/login
- POST /api/v1/auth/refresh
- POST /api/v1/auth/logout
- GET  /api/v1/auth/me

Rules:
- Access token short-lived.
- Refresh token rotation:
  - refresh exchanges old refresh token for a new pair
  - refresh reuse detection must invalidate the session and return a typed error

Error codes (standardized JSON):
- AUTH_UNAUTHORIZED
- AUTH_EXPIRED
- AUTH_REFRESH_REUSED
- AUTH_REFRESH_INVALID

### 4.2.2 API Token Authentication (Machine/Integration)
Backend must also support authenticating requests via an API token:
- Header: X-API-Token: <token>
or:
- Authorization: Bearer <token> (token-type distinguishable server-side)

Use cases:
- automation scripts
- other services calling the API

Token storage:
- Store hashed token in DB.
- Support token rotation (create new, revoke old).
- Tokens may be scoped by permissions.

## 4.3 Authorization: Roles + Permissions + Scoping
Roles (initial):
- Admin
- Manager
- Staff
- Viewer

Permissions:
- Defined as strings, e.g.:
  - users.read, users.write
  - audit.read
  - settings.read, settings.write

Scoping requirement:
- permissions can optionally be scoped to a module (e.g. module=hr, module=payroll)
- MVP: implement “module scope” as an optional column; enforce later as needed

## 4.4 User Management (Server)
Admin-only endpoints:
- GET /api/v1/users
- POST /api/v1/users
- PATCH /api/v1/users/:id (role, active flag)
- POST /api/v1/users/:id/reset-password

## 4.5 Audit Logging
- audit_logs table records:
  - id, timestamp, actor_user_id, action, entity_type, entity_id, metadata_json
- Must record at least:
  - auth.login.success / auth.login.failure
  - auth.refresh
  - auth.logout
  - users.create / users.update / users.reset_password / users.set_active
- Admin-only endpoint:
  - GET /api/v1/audit (with pagination + filters)

## 4.6 Health / Version
- GET /api/v1/health (db connectivity + version)
- GET /api/v1/version

## 4.7 Backend Configuration (Viper + Hot Reload)
The backend must use a config file, loaded via **Viper**, supporting hot reload.

Requirements:
- Config sources:
  1) config file (YAML)
  2) environment variables (override)
  3) flags (override)
- Hot reload:
  - Use fsnotify/Viper WatchConfig.
  - On file change, load and validate the new config.
  - Swap the active config atomically for readers (use `sync/atomic` via `atomic.Value`).
- Read/write caution:
  - Runtime-generated or mutable settings (e.g., secrets generated at runtime) must not “fight” with hot reload.
  - If the app writes any config to disk, implement a safe write strategy (write temp file + atomic rename) and prevent reload loops.
  - Keep “backend runtime state” separate from “configuration”.

---

# 5. Graceful Shutdown + Context
Both desktop and backend processes must:
- Use `context.Context` properly to propagate cancellation.
- Handle SIGTERM and Ctrl+C (SIGINT) using `signal.NotifyContext`.
- Ensure HTTP server shuts down gracefully:
  - stop accepting new requests
  - allow in-flight requests to complete within a timeout
- Ensure background goroutines exit on context cancellation.

---

# 6. Non-Functional Requirements
- Clean architecture:
  - handlers -> services -> repositories
- Parameterized SQL only
- Consistent error response shape
- Tests:
  - backend: auth + RBAC + audit + config reload behavior
  - frontend: routing + auth flow smoke tests (including NotFound route behavior)

---

# 7. Makefiles (Required)
A Makefile must exist (at least at repo root), to simplify routine commands, for example:
- `make backend-run`
- `make backend-test`
- `make migrate-up` / `make migrate-down`
- `make desktop-dev`
- `make desktop-test`
- `make web-dev` (if web frontend exists)

---

# 8. Repository Structure (Recommended)

repo/
  backend/
    cmd/api/
    internal/
      config/
      db/
      auth/
      rbac/
      users/
      audit/
      middleware/
    migrations/
  desktop/
    frontend/
    internal/ (wails bindings: config store, api client, file dialogs)
    wails.json
  web/ (optional, React app targeting browsers)
  docs/
    requirements.md
    status.md
    prompts/   (not committed)
    notes/
  Makefile

---

# END (Authoritative)
