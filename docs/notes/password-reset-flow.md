# Password Reset Flow (Backend Contract)

## Endpoints
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`

## Forgot Password
Request body:
```json
{
  "identifier": "username-or-email",
  "resetUrl": "https://example.com/reset-password"
}
```

Behavior:
- Accepts username or email identifier.
- Always returns `202` with `{ "status": "accepted" }` to avoid account enumeration.
- If account exists and is active, backend generates a secure random token, stores only token hash, and sets expiry.
- Optional `resetUrl` is validated as absolute `http(s)` URL and receives `token` query param for downstream clients/email providers.

## Reset Password
Request body:
```json
{
  "token": "token-from-link",
  "password": "new-password"
}
```

Behavior:
- Validates token exists, is unexpired, and unused.
- Marks token used, updates password hash, revokes active refresh sessions, and invalidates remaining active reset tokens for the same user.
- Returns `200` with `{ "status": "ok" }` on success.

## Security Notes
- Token material is stored hashed (`sha256`) in `password_reset_tokens`.
- Used and expired token paths are treated as invalid for reset requests.
- Audit events emitted:
  - `auth.password_reset.requested`
  - `auth.password_reset.success`
