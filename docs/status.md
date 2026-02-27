# Status

## Milestone 1 — Repo Bootstrap + Baseline Build (Complete)

### What changed
- Bootstrapped repository structure for `backend/`, `desktop/` (generated via `wails init`), `web/`, and `docs/`.
- Added root project files: `.gitignore`, `Makefile`, and `docker-compose.yml`.
- Added prompt traceability folder `docs/prompts/` and stored the Milestone 1 prompt copy (kept ignored from git).
- Bootstrapped backend Go API with Gin under `/api/v1`.
- Implemented `GET /api/v1/health` returning `{ "status": "ok" }`.
- Implemented backend graceful shutdown with `signal.NotifyContext` + `http.Server.Shutdown` timeout.
- Added backend route test for `/api/v1/health`.
- Bootstrapped desktop app using `wails init` (React + TypeScript template).
- Added required frontend dependencies: MUI, MUI Icons, Emotion, TanStack Router, TanStack Query, and MUI X Data Grid.
- Implemented TanStack Router root/outlet + `/` route + configured NotFound route.
- Implemented minimal MUI centered card UI:
  - Title: `Skeleton App Ready`
  - Subtitle: `Wails + React + MUI + TanStack Router`
- Added frontend route tests:
  - `/` renders expected content
  - unknown route renders NotFound

### How to run
- Backend API: `make backend-run`
- Desktop app (dev): `make desktop-dev`

### How to test
- Backend tests: `make backend-test`
- Desktop route tests: `make desktop-test`

### Verification summary
- Backend builds: PASS (`go build ./...` in `backend/`)
- Backend tests pass: PASS (`go test ./...` in `backend/`)
- Desktop builds: PASS (`wails build -compiler /usr/local/go/bin/go -skipbindings` in `desktop/`)
- Desktop route tests pass: PASS (`npm test` in `desktop/frontend`)

### Known follow-ups
- `wails build` in this environment needs explicit Go toolchain env (`GOROOT=/usr/local/go`) and `-skipbindings` due local Wails binding-generation issues with the shell Go configuration and module scanning.

## Milestone 2 — Backend Foundation (Complete)

### What changed
- Added `backend/internal/config` with typed config struct and Viper loading using precedence: config file, environment variables, then CLI flag overrides.
- Enabled Viper hot reload via `WatchConfig` and `OnConfigChange`.
- Added atomic runtime snapshot (`atomic.Value`) with `config.Get()` so runtime code reads typed config without direct Viper access.
- Added reload validation logic: invalid config reloads are rejected and previous snapshot remains active.
- Added safe write helper (`SafeWriteFile`) that writes temp file then atomic rename.
- Added `backend/internal/db` with SQLX-based DB initialization, pool sizing, and startup ping.
- Added migration SQL files under `backend/migrations` for:
  - users
  - roles
  - permissions
  - role_permissions
  - user_roles
  - api_tokens
  - refresh_tokens
  - audit_logs
- Added migration runner commands in `backend/cmd/migrate` and shared helpers in `backend/internal/migrateutil`.
- Added Makefile migration targets:
  - `make migrate-up`
  - `make migrate-down`
  - `make migrate-create name=<migration_name>`
- Hardened API startup/shutdown:
  - `signal.NotifyContext` in `cmd/api/main.go`
  - shutdown timeout sourced from config
  - `http.Server.Shutdown(...)` on cancellation
  - clean DB close on process exit
  - startup server lifecycle helper with unit test coverage
- Updated API routes to include:
  - `GET /api/v1/health` returning `{ "status": "ok", "version": "...", "db": "up" }` when DB is healthy
  - `GET /api/v1/version` returning backend version
- Added tests:
  - config hot reload atomic swap + invalid reload retention test
  - health route test with DB ping mock
  - server graceful startup/shutdown lifecycle test
  - integration-style DB open test that skips when `BASEPRO_TEST_DSN` is not set

### How config reload works
- Config is loaded via Viper from `backend/config/config.yaml` (or `--config` path), env vars (`BASEPRO_*`), and CLI overrides.
- Hot reload watches config file changes.
- On each change, config is decoded + validated.
- If valid, new config is atomically swapped into `config.Get()`.
- If invalid, reload is logged and ignored; previous config remains active.

### How to run backend
- Ensure PostgreSQL is running (example: `docker compose up -d postgres`).
- From repo root run:
  - `make migrate-up`
  - `make backend-run`

### How to run migrations
- Apply all up migrations: `make migrate-up`
- Roll back one migration: `make migrate-down`
- Create migration pair: `make migrate-create name=<migration_name>`

### How to test
- Backend tests: `make backend-test`
- Frontend route/smoke tests: `make desktop-test`

### Verification summary
- Backend builds/tests (`go test ./...` in `backend/`): PASS
- Health route test: PASS
- Config reload test: PASS
- Frontend route tests: PASS

### Known follow-ups
- Auto-migration is available via `database.auto_migrate` config / `--database-auto-migrate`; keep disabled in production.
- Authentication/RBAC remains intentionally unimplemented for Milestone 2.

## Milestone 3 — Backend Auth (JWT + Refresh Rotation + Typed Errors + Audit) (Complete)

### What changed
- Added auth-specific migration `000009_auth_token_chain_and_audit_indexes` to extend existing tables without editing prior migrations.
- `refresh_tokens` now supports rotation/reuse chain tracking fields:
  - `issued_at`
  - `replaced_by_token_id`
  - `updated_at`
- `audit_logs` now supports `timestamp` for ordered event querying.
- Added required indexes:
  - `refresh_tokens(user_id)`
  - `refresh_tokens(token_hash)` unique index
  - `audit_logs(timestamp DESC)`
- Extended backend config with auth settings:
  - `auth.access_token_ttl_seconds`
  - `auth.refresh_token_ttl_seconds`
  - `auth.jwt_signing_key`
  - `auth.password_hash_cost`
- Added config validation to fail fast when JWT signing key is empty.
- Added `internal/apperror` for standardized typed error responses.
- Added `internal/audit` (SQLX repository + service).
- Added `internal/auth`:
  - SQLX repository for users/refresh tokens
  - password hashing/verify helpers (bcrypt)
  - JWT manager
  - auth service implementing login, refresh rotation, refresh reuse detection, logout, and `me`
  - HTTP handlers for `/api/v1/auth/*`
- Added `internal/middleware` JWT auth middleware with request context claims injection.
- Wired auth + audit dependencies in `cmd/api/main.go` via dependency injection (no global DB state).
- Added backend auth endpoints:
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout`
  - `GET /api/v1/auth/me` (JWT protected)

### Typed error shape
- Handlers now return auth errors in standardized shape:
  - `{ "error": { "code": "...", "message": "..." } }`
- Implemented codes:
  - `AUTH_UNAUTHORIZED`
  - `AUTH_EXPIRED`
  - `AUTH_REFRESH_REUSED`
  - `AUTH_REFRESH_INVALID`

### Refresh rotation and reuse detection
- Refresh tokens are generated as random opaque strings and only SHA-256 hashes are stored in DB.
- On refresh success:
  - old token row is marked `revoked_at`
  - old token row links `replaced_by_token_id` to the new token row
  - new access + refresh tokens are issued
- On refresh reuse (revoked token presented):
  - response returns `AUTH_REFRESH_REUSED` (401)
  - active refresh tokens for that user are revoked to contain the token family/session.

### Auth endpoint examples (curl)
- Login:
  - `curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login -H 'Content-Type: application/json' -d '{"username":"alice","password":"secret"}'`
- Refresh:
  - `curl -s -X POST http://127.0.0.1:8080/api/v1/auth/refresh -H 'Content-Type: application/json' -d '{"refreshToken":"<refresh-token>"}'`
- Logout with refresh token:
  - `curl -s -X POST http://127.0.0.1:8080/api/v1/auth/logout -H 'Content-Type: application/json' -d '{"refreshToken":"<refresh-token>"}'`
- Me with access token:
  - `curl -s http://127.0.0.1:8080/api/v1/auth/me -H 'Authorization: Bearer <access-token>'`

### How to test
- Backend tests: `make backend-test`
- Frontend route/smoke tests: `make desktop-test`

### Verification summary
- Backend tests (`go test ./...`): PASS
- Login success test (token issuance + hashed refresh storage): PASS
- Login failure typed error test: PASS
- Refresh rotation test (old revoked + new works): PASS
- Refresh reuse detection + active token revocation test: PASS
- JWT middleware missing token -> `AUTH_UNAUTHORIZED`: PASS
- JWT middleware expired token -> `AUTH_EXPIRED`: PASS
- Frontend route tests: PASS

### Milestone scope guard
- API-token authentication was not implemented in this milestone.
- RBAC roles/permission enforcement was not implemented in this milestone.
