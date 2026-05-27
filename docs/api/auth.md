# Authentication

## Model

The backend uses short-lived bearer access tokens plus an opaque refresh token in an `HttpOnly` cookie.

Send the access token in the `Authorization` header.

```text
Authorization: Bearer <token>
```

## `POST /api/v1/auth/register`

| Field | Value |
| --- | --- |
| Auth | No |
| Success | `201` auth result `{ access_token, user }` plus `Set-Cookie: refresh_token=...` |
| Errors | `400 INVALID_INPUT`, `400 INVALID_EMAIL`, `400 INVALID_PASSWORD`, `400 EMAIL_ALREADY_EXISTS`, `500 INTERNAL_SERVER_ERROR` |
| Notes | Password must be at least 8 characters |

Request:

```json
{
  "email": "demo@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "access_token": "eyJhbGciOi...",
  "user": {
    "id": "7d7efb4d-0f1c-4f69-9b74-5e2a0d5d5a50",
    "email": "demo@example.com",
    "created_at": "2026-05-05T12:00:00Z"
  }
}
```

Validation errors use the shared envelope, for example:

```json
{
  "error": {
    "code": "INVALID_EMAIL",
    "message": "Invalid email",
    "details": {
      "field": "email"
    }
  }
}
```

## `POST /api/v1/auth/login`

| Field | Value |
| --- | --- |
| Auth | No |
| Success | `200` auth result `{ access_token, user }` plus `Set-Cookie: refresh_token=...` |
| Errors | `400 INVALID_INPUT`, `401 INVALID_CREDENTIALS`, `500 INTERNAL_SERVER_ERROR` |
| Notes | Missing users and bad passwords both map to `INVALID_CREDENTIALS` |

Request:

```json
{
  "email": "demo@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "access_token": "eyJhbGciOi...",
  "user": {
    "id": "7d7efb4d-0f1c-4f69-9b74-5e2a0d5d5a50",
    "email": "demo@example.com",
    "created_at": "2026-05-05T12:00:00Z"
  }
}
```

## `POST /api/v1/auth/refresh`

| Field | Value |
| --- | --- |
| Auth | Refresh cookie |
| Success | `200` auth result `{ access_token, user }` plus rotated `Set-Cookie: refresh_token=...` |
| Errors | `401 UNAUTHORIZED` |
| Notes | Refresh tokens are single-use; a successful refresh revokes the previous refresh token |

Request body is empty. The browser/client must send the `refresh_token` cookie scoped to `/api/v1/auth`.

## `POST /api/v1/auth/logout`

| Field | Value |
| --- | --- |
| Auth | Optional bearer token and optional refresh cookie |
| Success | `200` `{ "status": "ok" }` and cleared refresh cookie |
| Errors | None expected for invalid or missing tokens |
| Notes | If a valid access token is provided, its JWT ID is blacklisted until expiry; the refresh token is revoked when present |

## `GET /api/v1/auth/me`

| Field | Value |
| --- | --- |
| Auth | Yes |
| Success | `200` user object `{ id, email, created_at }` |
| Errors | `401 UNAUTHORIZED`, `500 INTERNAL_SERVER_ERROR` |
| Notes | Returns the current token subject |

Response:

```json
{
  "id": "7d7efb4d-0f1c-4f69-9b74-5e2a0d5d5a50",
  "email": "demo@example.com",
  "created_at": "2026-05-05T12:00:00Z"
}
```

Missing or invalid bearer tokens return `401` with `UNAUTHORIZED`.
