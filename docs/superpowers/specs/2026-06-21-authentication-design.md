# MatchLab Authentication MVP Design

## Scope

This phase adds email/password registration, login, seven-day JWT access tokens, authenticated current-user lookup, and an idempotent users-table migration. It changes only the Go backend, SQL, environment example, and documentation. It does not add frontend work, verification codes, refresh tokens, password recovery, logout state, or a broader permission system.

## Package Boundaries

- `internal/user` owns the GORM `User` model and PostgreSQL repository.
- `internal/auth` owns request validation, registration/login services, bcrypt password handling, JWT creation, and HTTP handlers.
- `internal/middleware` validates Bearer tokens and writes `user_id` and `role` to the Gin context.
- `internal/router` receives explicit database and JWT dependencies and registers public and protected routes.
- `internal/config` loads `JWT_SECRET`, using a documented development-only fallback when absent.

The health route remains independent of PostgreSQL and authentication. If PostgreSQL is unavailable at startup, `/api/health` remains available while authentication routes return HTTP 503.

## User Model

The API-facing user fields are:

- `id`: UUID
- `email`: normalized lowercase email, unique and required
- `nickname`: optional string
- `role`: `user` by default; allowed values are `user` and `admin`
- `school`: optional string
- `created_at` and `updated_at`: timezone-aware timestamps

`password_hash` is stored but never serialized. The existing database `status` column is preserved for compatibility but is not exposed by this MVP model.

## API Contract

### Register

`POST /api/auth/register` accepts `email`, `password`, `nickname`, and `school`. Email is trimmed, lowercased, and validated as an email address. Password must contain at least eight Unicode characters. Passwords are hashed with bcrypt's default cost before storage.

Success returns HTTP 201:

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "test@example.com",
      "nickname": "测试用户",
      "role": "user",
      "school": "西南大学",
      "created_at": "2026-06-21T00:00:00Z",
      "updated_at": "2026-06-21T00:00:00Z"
    }
  }
}
```

Invalid input returns 400. A duplicate normalized email returns 409.

### Login

`POST /api/auth/login` accepts email and password. Invalid credentials always return the same 401 response so the API does not reveal whether an email exists. Success returns HTTP 200 with a JWT and the safe user object:

```json
{
  "data": {
    "token": "jwt",
    "user": {}
  }
}
```

### Current User

`GET /api/me` requires `Authorization: Bearer <token>`. The middleware validates signature, algorithm, and expiry, then stores `user_id` and `role` in Gin context. The handler reloads the user from PostgreSQL and returns `{"data":{"user":{...}}}`. Missing, malformed, expired, or invalid tokens return 401.

### Errors

All errors use:

```json
{
  "error": "machine_readable_code",
  "message": "human-readable Chinese message"
}
```

The status mapping is 400 for validation, 401 for credentials/token failures, 404 when the token's user no longer exists, 409 for duplicate email, 500 for unexpected failures, and 503 when authentication storage is unavailable.

## JWT

Tokens use HMAC SHA-256 and include:

- `sub`: user UUID
- `role`: current role
- `iat`: issued-at time
- `exp`: issued-at plus seven days
- `iss`: `matchlab`

`JWT_SECRET` comes from the environment. When absent, development uses a visible fallback and startup logs a warning. `.env.example`, README, and deployment documentation require a long random production value.

## Database Migration

`database/schema.sql` enables both `uuid-ossp` and the existing `pgcrypto` extension. The `users` creation statement includes all requested fields and remains guarded by `IF NOT EXISTS`. For an already deployed table, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` adds `nickname` and `school` without destroying data. The existing case-insensitive unique email index remains the source of duplicate-email enforcement.

## Testing

Tests are written before production behavior and cover:

- registration validation, normalization, bcrypt storage, safe response, and duplicate conflict;
- login success and indistinguishable invalid-credential failures;
- seven-day JWT issuance, expiry, signature, and algorithm validation;
- Bearer middleware context values and 401 paths;
- current-user lookup and password-hash omission;
- router availability, database-unavailable behavior, and unchanged `/api/health`;
- configuration fallback and `JWT_SECRET` override.

Final verification runs `go test ./...`, `go vet ./...`, `go build`, and HTTP smoke tests against a disposable test server. PostgreSQL schema execution remains an explicit deployment step because no local production database is modified automatically.
