# MatchLab Backend MVP Design

## Scope

This phase creates a new, independently runnable `matchlab-web` project. It delivers only the backend foundation needed for a four-day MVP: a Gin HTTP server, environment-based configuration, optional PostgreSQL/GORM initialization, a database-independent health endpoint, initial SQL schema, deployment files, and operating documentation. Registration, login, authorization, and frontend behavior are explicitly deferred.

## Architecture

The Go module uses a lightweight modular structure:

- `cmd/server` owns process startup and shutdown.
- `internal/config` loads `.env` and environment variables into a typed configuration value.
- `internal/database` constructs a GORM PostgreSQL connection when database configuration is available.
- `internal/router` assembles Gin routes and middleware.
- `internal/health` owns the health HTTP handler.
- Future domain packages are created as documented placeholders without speculative implementations.

Database initialization is best-effort for this phase. A missing or unavailable PostgreSQL instance is logged, while the HTTP server still starts. Consequently, `GET /api/health` never depends on database availability.

## HTTP Behavior

`GET /api/health` returns HTTP 200 and this JSON body:

```json
{
  "ok": true,
  "message": "MatchLab API running"
}
```

Unknown routes use Gin's standard 404 behavior. The server listens on the configured host and port, defaulting to `127.0.0.1:8080`.

## Configuration

`godotenv` loads a local `.env` file when present. Environment variables remain authoritative and production may provide them through systemd. Configuration covers server host/port, Gin mode, and PostgreSQL host, port, database, user, password, and SSL mode. Secrets are excluded from source control; `.env.example` contains safe placeholders.

## Database Schema

`database/schema.sql` enables UUID generation and defines these initial tables:

- `users`: identity, email, password hash, role, account state, timestamps.
- `profiles`: one-to-one user profile and matching preferences.
- `questionnaires`: questionnaire answers and scoring data as JSONB.
- `activities`: activity ownership, details, capacity, status, and schedule.
- `applications`: user applications to activities and their status.
- `matches`: questionnaire/activity matching results and explanations.
- `events`: auditable domain events with JSONB payloads.
- `feedbacks`: user ratings and comments for activities or matches.

Foreign keys, uniqueness rules, timestamps, useful indexes, and conservative status checks are included. The schema is intentionally sufficient for the next authentication and activity iterations without implementing those APIs now.

## Deployment

Nginx proxies `/api/` to `http://127.0.0.1:8080` and forwards standard proxy headers. The systemd unit runs a compiled binary as a dedicated service user, loads `/opt/matchlab/backend/.env`, restarts after failures, and starts after networking. `deploy.sh` builds the Go binary and prints the privileged installation steps instead of silently modifying the host.

## Testing and Verification

The health endpoint is developed test-first with `net/http/httptest`. Verification consists of:

1. Running `go test ./...`.
2. Running `go vet ./...`.
3. Building the server binary.
4. Starting the server without a database and requesting `/api/health`.
5. Confirming HTTP 200 and the exact JSON fields.

## Deferred Work

The next phase will add password hashing, registration, login, JWT access tokens, authentication middleware, validation, uniqueness conflict handling, and authentication integration tests. Refresh tokens, third-party login, password recovery, and advanced RBAC remain outside the immediate MVP unless later prioritized.
