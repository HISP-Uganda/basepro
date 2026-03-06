# Error Handling Foundation (Desktop + Web)

## Notification Contract

Both clients now support a shared `AppNotification` shape:

```ts
type AppNotification = {
  kind: 'success' | 'error' | 'warning' | 'info'
  message: string
  title?: string
  autoHideDuration?: number
  requestId?: string
  persistent?: boolean
}
```

Application code should dispatch through facade methods:

- `notify.success(message, options?)`
- `notify.error(message, options?)`
- `notify.warning(message, options?)`
- `notify.info(message, options?)`

Desktop uses a pub/sub store consumed by the root snackbar.
Web uses the snackbar provider context and a notify hook facade.

## Error Normalization

Both clients now provide:

- `normalizeError(error, fallbackMessage?)`

Normalized shape:

```ts
type NormalizedAppError = {
  type:
    | 'validation'
    | 'unauthorized'
    | 'forbidden'
    | 'not_found'
    | 'conflict'
    | 'network'
    | 'timeout'
    | 'server'
    | 'unknown'
  message: string
  fieldErrors?: Record<string, string[]>
  requestId?: string
}
```

## Error Handling Helper

Both clients now provide:

- `handleAppError(error, options?)`

Helper behavior:

- normalizes unknown/backend/network errors
- optionally maps `validation` field errors to form callbacks
- optionally triggers session-expiry callbacks on `unauthorized`
- optionally dispatches user-facing notifications through notify facades
- returns structured normalized results for page-level decisions

## Session Expiry Standard

Unauthorized/expired-session handling now follows a shared behavior:

1. Detect unauthorized/expired auth responses.
2. Clear local session/auth state.
3. Notify user: `Session expired. Please log in again.`
4. Redirect to `/login`.
5. Preserve intended destination in session storage where practical, then consume it on the next successful login.

## Error Boundary Fallback

Both clients now wrap root rendering with `AppErrorBoundary`:

- friendly fallback message
- reload action
- no stack traces shown in production
- optional error message detail in development mode only
