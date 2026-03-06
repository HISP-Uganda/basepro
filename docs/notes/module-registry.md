# Module Registry Workflow (Registry-First)

This note documents the recommended workflow for adding a new module using BasePro's registry-first architecture.

Status: educational/reference-only guidance.  
This does not mean a new business module is implemented.

## 1) Define the Module in Client Registries

Add the module in:

- `desktop/frontend/src/registry/modules.ts`
- `web/src/registry/modules.ts`

Each module entry should define:

- `id` (module key)
- `label`
- `navGroup`
- `basePath`
- `permissions`
- `navItems`
- optional `flags`/`metadata`

Example (reference only):

```ts
{
  id: 'example_module',
  label: 'Example Module',
  navGroup: 'administration',
  basePath: '/example-module',
  permissions: ['example_module.read', 'example_module.write'],
  navItems: ['example-module'],
  metadata: {
    description: 'Reference-only module scaffold example.',
  },
}
```

## 2) Define Navigation Entries

Add navigation entries in:

- `desktop/frontend/src/registry/navigation.ts`
- `web/src/registry/navigation.ts`

Use grouped navigation and RBAC-aware visibility via `requiredPermissions`.

Example child item under `Administration` (reference only):

```ts
{
  id: 'example-module',
  label: 'Example Module',
  icon: 'administration',
  path: '/example-module',
  group: 'administration',
  requiredPermissions: ['example_module.read', 'example_module.write'],
}
```

Notes:

- Frontend checks are UX-only (hide/disable affordances).
- Backend authorization remains authoritative.

## 3) Define Permissions (Backend + Clients)

Add permission metadata in:

- `desktop/frontend/src/registry/permissions.ts`
- `web/src/registry/permissions.ts`
- `backend/internal/rbac/registry.go`

Use predictable naming:

- `example_module.read`
- `example_module.write`
- `example_module.delete`
- `example_module.admin`

Example client permission entries (reference only):

```ts
{
  key: 'example_module.read',
  label: 'Example Module: Read',
  description: 'View example module records.',
  module: 'example_module',
  category: 'Administration',
},
{
  key: 'example_module.write',
  label: 'Example Module: Write',
  description: 'Create and update example module records.',
  module: 'example_module',
  category: 'Administration',
}
```

Example backend constants + registry entries (reference only):

```go
const (
	PermissionExampleModuleRead  = "example_module.read"
	PermissionExampleModuleWrite = "example_module.write"
)
```

## 4) Align Backend Endpoints With Permission Keys

When a module introduces API actions:

1. Define/update backend API contract first.
2. Add permission keys to backend RBAC registry/constants (`backend/internal/rbac/registry.go`).
3. Protect endpoints with the same permission keys in router middleware wiring.
4. Ensure typed/clear authorization validation errors are returned.

Reference pattern (illustrative only):

```go
group.GET("/example-module", authMW.RequirePermission(rbac.PermissionExampleModuleRead), handler.List)
group.POST("/example-module", authMW.RequirePermission(rbac.PermissionExampleModuleWrite), handler.Create)
```

## 5) Maintain Desktop/Web Parity

If module scope is shared, keep parity across desktop and web for:

- route existence
- navigation visibility intent
- major CRUD capability surface
- permission checks and backend contract usage

Temporary gaps are allowed only if explicitly documented in `docs/status.md` with follow-up scope.

## 6) Document the Module Work in Status

For each module milestone, update `docs/status.md` with:

- module name and intent
- layers affected (`backend`, `desktop`, `web`)
- routes/pages added
- permission keys added
- backend endpoints added/changed
- tests run and results
- parity notes and any temporary gaps

## 7) Recommended Order (Checklist)

1. Add permission keys/metadata in backend and both clients.
2. Add module definitions in desktop/web module registries.
3. Add navigation entries in desktop/web navigation registries.
4. Add backend endpoints and permission middleware.
5. Add desktop/web route wiring and UI pages.
6. Run tests (`go test ./...`, desktop/frontend tests, web tests).
7. Update `docs/status.md`.

