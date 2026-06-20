# MatchLab Backend MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a runnable, deployable Go backend foundation whose `/api/health` endpoint works without PostgreSQL.

**Architecture:** A small `cmd/server` composition root loads typed configuration, optionally opens PostgreSQL through GORM, constructs a Gin router, and runs the HTTP server. The health package is independent of database state and is tested through the public HTTP route.

**Tech Stack:** Go 1.22, Gin, GORM PostgreSQL driver, godotenv, PostgreSQL SQL, Nginx, systemd

---

### Task 1: Module and health endpoint test

**Files:**
- Create: `matchlab-web/backend/go.mod`
- Create: `matchlab-web/backend/internal/health/handler_test.go`
- Create: `matchlab-web/backend/internal/router/router_test.go`

- [ ] **Step 1: Initialize the Go module**

Create module `matchlab/backend` targeting Go 1.22 with Gin, godotenv, GORM, and the GORM PostgreSQL driver.

- [ ] **Step 2: Write failing handler and route tests**

Test that the health handler returns status 200, `ok: true`, and `message: MatchLab API running`; test that the assembled router exposes it at `GET /api/health`.

- [ ] **Step 3: Run tests and verify RED**

Run: `go test ./internal/health ./internal/router`

Expected: compilation failure because the handler and router constructors do not exist.

### Task 2: Minimal HTTP implementation

**Files:**
- Create: `matchlab-web/backend/internal/health/handler.go`
- Create: `matchlab-web/backend/internal/router/router.go`
- Create: `matchlab-web/backend/internal/middleware/doc.go`

- [ ] **Step 1: Implement the health handler**

Expose `health.Handler(c *gin.Context)` and return the exact required JSON body with HTTP 200.

- [ ] **Step 2: Implement the router**

Expose `router.New() *gin.Engine`, attach Gin recovery and logger middleware, create `/api`, and register `GET /health`.

- [ ] **Step 3: Run tests and verify GREEN**

Run: `go test ./internal/health ./internal/router`

Expected: both packages pass.

### Task 3: Configuration and optional database initialization

**Files:**
- Create: `matchlab-web/backend/internal/config/config.go`
- Create: `matchlab-web/backend/internal/config/config_test.go`
- Create: `matchlab-web/backend/internal/database/database.go`
- Create: `matchlab-web/backend/internal/database/database_test.go`
- Create: `matchlab-web/backend/.env.example`
- Create: `matchlab-web/backend/.gitignore`

- [ ] **Step 1: Write failing configuration tests**

Test default address `127.0.0.1:8080`, environment overrides, incomplete database configuration detection, and a PostgreSQL DSN without logging the password.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `go test ./internal/config ./internal/database`

Expected: compilation failure because the configuration and database functions are absent.

- [ ] **Step 3: Implement typed environment configuration**

Load `.env` with godotenv, read server and PostgreSQL variables, provide `Address()`, and expose whether all database fields needed for a connection are present.

- [ ] **Step 4: Implement GORM connection construction**

Build a PostgreSQL DSN and call `gorm.Open(postgres.Open(dsn))` only when startup requests it. Return clear errors and do not auto-migrate.

- [ ] **Step 5: Add safe environment examples and ignore rules**

Document local values in `.env.example`; ignore `.env`, binaries, coverage, and temporary files.

- [ ] **Step 6: Run focused tests and verify GREEN**

Run: `go test ./internal/config ./internal/database`

Expected: both packages pass.

### Task 4: Server composition and graceful shutdown

**Files:**
- Create: `matchlab-web/backend/cmd/server/main.go`

- [ ] **Step 1: Compose startup dependencies**

Load configuration, set Gin mode, attempt the optional database connection, log database unavailability without exiting, create the router, and listen on the configured address.

- [ ] **Step 2: Add graceful shutdown**

Use `http.Server`, listen for SIGINT/SIGTERM, and allow a ten-second shutdown timeout.

- [ ] **Step 3: Run all Go tests**

Run: `go test ./...`

Expected: all packages pass.

### Task 5: Initial PostgreSQL schema

**Files:**
- Create: `matchlab-web/database/schema.sql`

- [ ] **Step 1: Define schema transaction and UUID support**

Use `BEGIN`, `CREATE EXTENSION IF NOT EXISTS pgcrypto`, and `COMMIT`.

- [ ] **Step 2: Define the eight MVP tables**

Create `users`, `profiles`, `questionnaires`, `activities`, `applications`, `matches`, `events`, and `feedbacks` with UUID primary keys, timestamps, required foreign keys, unique constraints, JSONB fields, and conservative status checks.

- [ ] **Step 3: Add lookup indexes**

Add indexes for activity schedule/status, applications, match lookups, events, and feedback ownership.

### Task 6: Future package boundaries and deployment assets

**Files:**
- Create: `matchlab-web/backend/internal/auth/doc.go`
- Create: `matchlab-web/backend/internal/user/doc.go`
- Create: `matchlab-web/backend/internal/activity/doc.go`
- Create: `matchlab-web/backend/internal/application/doc.go`
- Create: `matchlab-web/backend/internal/questionnaire/doc.go`
- Create: `matchlab-web/backend/internal/match/doc.go`
- Create: `matchlab-web/backend/internal/admin/doc.go`
- Create: `matchlab-web/backend/internal/algorithm/doc.go`
- Create: `matchlab-web/frontend/.gitkeep`
- Create: `matchlab-web/deploy/nginx-matchlab.conf`
- Create: `matchlab-web/deploy/matchlab-api.service`
- Create: `matchlab-web/deploy/deploy.sh`

- [ ] **Step 1: Create documented package placeholders**

Use valid `doc.go` package declarations so `go test ./...` covers every planned boundary without speculative domain code.

- [ ] **Step 2: Add Nginx reverse proxy**

Proxy `/api/` to `http://127.0.0.1:8080` and forward host, client address, forwarded-for, and forwarded-proto headers.

- [ ] **Step 3: Add systemd service**

Run `/opt/matchlab/backend/bin/matchlab-api` as `matchlab`, load `/opt/matchlab/backend/.env`, restart on failure, and apply basic service hardening.

- [ ] **Step 4: Add deployment script**

Build a Linux amd64 binary into `backend/bin` and print explicit installation/restart commands. The script must use `set -eu` and must not embed secrets.

### Task 7: Operating documentation

**Files:**
- Create: `matchlab-web/README.md`
- Create: `matchlab-web/docs/MVP_PLAN.md`
- Create: `matchlab-web/docs/API_DOC.md`
- Create: `matchlab-web/docs/DEPLOY.md`

- [ ] **Step 1: Document local startup and health testing**

Include `.env` creation, dependency download, `go run ./cmd/server`, curl, and PowerShell examples.

- [ ] **Step 2: Document database setup**

Include safe placeholder SQL for the `matchlab_user` password and the exact `psql` schema import command.

- [ ] **Step 3: Document Ubuntu deployment**

Cover upload paths, building, service user/directories, environment permissions, systemd installation, Nginx installation, validation, firewall considerations, logs, and rollback.

- [ ] **Step 4: Document four-day scope and next authentication phase**

Describe MVP exclusions and the concrete register/login sequence: password hashing, JWT, middleware, request validation, conflict handling, and tests.

### Task 8: Final verification

**Files:**
- Modify only if verification finds a defect.

- [ ] **Step 1: Format and tidy**

Run: `gofmt -w cmd internal`

Run: `go mod tidy`

- [ ] **Step 2: Verify tests, static analysis, and build**

Run: `go test ./...`

Run: `go vet ./...`

Run: `go build ./cmd/server`

Expected: every command exits zero.

- [ ] **Step 3: Verify database-independent runtime behavior**

Start the server with database variables absent, request `http://127.0.0.1:8080/api/health`, and confirm status 200 with the exact required JSON fields.

- [ ] **Step 4: Audit requested file structure**

List all files under `matchlab-web` and compare them against the requested structure before reporting completion.

## Repository Note

The containing workspace is not a Git repository, so the commit steps normally required by this workflow cannot be performed. This plan does not initialize Git implicitly because the user only authorized project creation.
