# AGENTS.md (Major Constraints / Contract)
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

## END
