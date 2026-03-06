# AGENTS.md (Major Constraints / Contract)
# BasePro Agent Guide

BasePro is both:

1. A **reusable application skeleton**
2. A **foundation for real applications built on top of it**

Agents must recognise when work concerns:

- platform capabilities (skeleton improvements)
- domain modules (application features)

## Desktop Skeleton App (Wails + Gin + React/MUI)

This file defines the project contract for automation agents (Codex CLI) and human contributors.

---

## 1) Source of Truth
- `docs/requirements.md` is the authoritative specification.
- Before starting any milestone, the agent **must read** `docs/requirements.md`.

---

## 2) Milestone Contract (Hard Rule)
- Work is milestone-based.
- **No milestone may begin until the previous milestone is complete.**
- A milestone is only complete when:
  - All required implementation tasks for that milestone are done,
  - **All tests pass**, including:
    - backend tests (`go test ./...`)
    - frontend route tests / smoke tests
  - `docs/status.md` is updated to reflect completion.

If any tests fail, the agent must fix them before claiming completion.

---

## 3) Prompts Hygiene
- A copy of each milestone prompt must be saved under `docs/prompts/` for traceability.
- **Do not commit** the prompt copies.
- Ensure `.gitignore` includes:
  - `docs/prompts/`

---

## 4) Status Updates
After a successful milestone:
- Update `docs/status.md`:
  - milestone name
  - what changed (high-level)
  - how to run tests / what passed
  - any known follow-ups

---

## 5) Commit Discipline
After a successful milestone (tests passing + status updated):
- The agent must propose a commit message in conventional style, e.g.:
  - `feat(backend): add jwt login + refresh rotation`
  - `feat(desktop): add setup screen + api base url persistence`
  - `test(frontend): add route smoke tests`
- The agent must explicitly prompt the user to commit (but must not claim the commit happened).

Do not propose commits when tests are failing.

---

## 6) Logging / Secrets
- Never log secrets (JWTs, refresh tokens, API tokens, passwords).
- Mask tokens if they must appear in debug output.
- Config files should not be committed if they contain secrets.

---

## 7) Graceful Shutdown
- Backend must use `signal.NotifyContext` and `http.Server.Shutdown(...)`.
- Any background goroutine must exit on context cancellation.

---

## 8) Config Hot Reload (Backend)
- Backend config is Viper-based and supports hot reload.
- Runtime readers must use an atomic snapshot (`atomic.Value`).
- If writing config to disk:
  - write to a temp file then atomic rename
  - avoid reload loops
  - keep “runtime state” separate from “config”

---

## 9) UI Constraints
- MUI-based admin dashboard layout:
  - Drawer + AppBar + content + footer
- Use MUI X Data Grid with advanced features.
- Themes: light/dark/system + accent palette presets persisted locally.

---

## 10) API-Only Desktop
- Desktop must never access DB directly.
- All domain data must go through Gin APIs.

---

## 11) Platform Skeleton vs Application Projects

This repository is a **reusable platform skeleton** first, and later may become a concrete application.

The skeleton phase focuses on shared platform capabilities:

* authentication
* RBAC (roles and permissions)
* audit logging
* settings management
* shared application shell
* backend API contracts
* desktop/web clients consuming the same backend

During skeleton work:

* prioritize platform features over domain features
* avoid domain placeholders that are not defined by requirements
* keep default navigation focused on:
  * Dashboard
  * Administration
  * Settings

When a project intentionally transitions to domain modules:

1. Define domain scope in `docs/requirements.md`.
2. Update navigation structure intentionally.
3. Record the transition and scope in `docs/status.md`.

Agents must allow domain expansion when requirements explicitly define it.

---

## 12) Navigation Architecture

Navigation must be grouped and scalable.

Baseline skeleton grouping:

* Dashboard
* Administration
  * Users
  * Roles
  * Permissions
  * Audit Log
* Settings

Applications may add groups (for example `Operations`, `Reporting`, `Finance`) only when defined in requirements.

Rules:

* keep navigation grouped, not flat
* keep platform/system management under `Administration` by default
* apply RBAC visibility rules to navigation items
* desktop and web may differ in responsive layout, but grouping and permission intent must remain consistent

---

## 13) Cross-Client Parity Contract

BasePro has three core layers:

* Gin backend API
* Wails desktop client
* React/TypeScript web client

Shared platform capabilities are expected in both desktop and web.
Temporary parity gaps are allowed only when explicitly documented in `docs/status.md` with follow-up scope.

When adding shared platform capabilities:

* define/update backend API contract first
* implement client consumption in desktop and web
* keep behavior aligned unless a documented temporary gap exists

For shared authentication-entry capabilities:

* keep login, forgot-password, and reset-password views behaviorally aligned across desktop and web
* keep configurable login branding settings (display name, login image, and related presentation settings) aligned across clients
* any temporary parity gap for authentication flows must be explicitly documented in `docs/status.md` with follow-up scope

---

## 14) Shared Administration UX Patterns

Administration UX should remain consistent and reusable across clients.

Preferred patterns:

* DataGrid list pages
* shared actions column behavior
* confirmation flows for destructive actions
* dialogs/drawers for detailed views
* consistent validation display and submission behavior

Prefer extending shared components/utilities for:

* DataGrid wrappers
* row actions renderers
* confirmation dialogs
* metadata/JSON viewers
* shared form helpers and validation adapters

Avoid one-off patterns unless there is a clear requirement gap.

---

## 15) DataGrid Contract

DataGrid usage must follow a common baseline:

* support horizontal and vertical scrolling
* keep action columns pinned right where supported
* degrade gracefully when pinning is unavailable
* honor global UI preferences when configured (for example density, radius, action pinning)
* do not silently ignore global preferences at table level

Grid containers must remain scroll-safe and must not clip:

* menus
* dialogs
* pinned columns
* scrollbars

---

## 16) RBAC UI / Backend Coupling

Frontend permission checks are UX-level only.
Backend authorization is mandatory and authoritative.

Rules:

* enforce permissions server-side for all protected actions
* use UI permission checks only to hide/disable affordances
* validate role/permission assignments server-side
* return typed validation errors for invalid identifiers
* surface validation errors clearly in clients

---

## 17) Audit UX Rules

Audit metadata can contain structured JSON and may be large.

Rules:

* show compact/truncated metadata previews in tables
* provide full metadata in a details dialog/drawer
* metadata viewer must support:
  * pretty print
  * scroll
  * copy
* avoid rendering full raw JSON directly in grid cells

---

## 18) Extension Rule

When extending the skeleton, prefer additive changes over risky core rewrites.

Prefer:

* adding modules in clear directories
* extending navigation via shared patterns/configuration
* adding backend modules without breaking existing API contracts

Treat these as core platform modules:

* authentication
* RBAC
* audit logging
* settings
* application shell

Changes to core modules should be intentional, minimal, and documented.

---

## 19) Agent Behavior

Agents must:

* read `docs/requirements.md` and `docs/status.md` before milestone work
* obey milestone sequencing and completion gates
* preserve backend/desktop/web contract alignment
* avoid breaking API changes without coordinated client updates
* keep changes incremental unless requirements demand broader refactor
* ensure required tests pass before claiming milestone completion

---

## 20) Prompt Scope Discipline

Milestone prompts must explicitly state impacted layers:

* backend
* desktop
* web
* or all three

Prompts should also explicitly require updates to:

* tests
* `docs/status.md`
* relevant documentation

Do not replace working behavior with temporary placeholders that reduce parity, quality, or architectural consistency unless requirements explicitly permit it.


## END
