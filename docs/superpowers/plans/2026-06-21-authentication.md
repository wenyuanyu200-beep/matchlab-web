# MatchLab Authentication MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add registration, login, seven-day JWT authentication, and `/api/me` without changing the frontend or breaking `/api/health`.

**Architecture:** The `user` package owns the model and repository contract; `auth` owns business rules, bcrypt, JWT, and handlers; `middleware` authenticates Bearer tokens; `router` composes explicit dependencies. Router tests use an in-memory repository to exercise the complete HTTP flow without modifying PostgreSQL.

**Tech Stack:** Go 1.22, Gin, GORM/PostgreSQL, bcrypt, golang-jwt/jwt v5, UUID, PostgreSQL SQL

---

### Task 1: Configuration and user model

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Create: `backend/internal/user/model.go`
- Create: `backend/internal/user/model_test.go`

- [ ] **Step 1: Write failing tests**

Test that `JWT_SECRET` defaults to the development value, environment configuration overrides it, default use is detectable, and converting a user to its public representation omits `password_hash` while retaining UUID, email, nickname, role, school, and timestamps.

- [ ] **Step 2: Verify RED**

Run: `go test ./internal/config ./internal/user`

Expected: compilation failures for the missing JWT fields and user model.

- [ ] **Step 3: Implement minimal configuration and model**

Add `JWTSecret`, `UsesDevelopmentJWTSecret()`, and the development constant. Define a GORM `User` with UUID primary key, hidden password hash, role default, and `Public()` conversion.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./internal/config ./internal/user`

Expected: all tests pass.

### Task 2: User repository

**Files:**
- Create: `backend/internal/user/repository.go`
- Create: `backend/internal/user/repository_test.go`

- [ ] **Step 1: Write failing repository-contract tests**

Test PostgreSQL unique-violation mapping to `ErrEmailExists`, record-not-found mapping to `ErrNotFound`, and nil database behavior as `ErrUnavailable`.

- [ ] **Step 2: Verify RED**

Run: `go test ./internal/user`

Expected: compilation failures for repository types and errors.

- [ ] **Step 3: Implement repository contract**

Define `Repository` with `Create`, `FindByEmail`, and `FindByID`. Implement GORM methods with context, map PostgreSQL code `23505`, map `gorm.ErrRecordNotFound`, and guard a nil database.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./internal/user`

Expected: all tests pass.

### Task 3: JWT token manager

**Files:**
- Create: `backend/internal/auth/token.go`
- Create: `backend/internal/auth/token_test.go`

- [ ] **Step 1: Write failing token tests**

Test HS256 issuance with `sub`, `role`, `iss=matchlab`, and a seven-day expiry; test parsing valid tokens and rejecting expired, tampered, and non-HS256 tokens.

- [ ] **Step 2: Verify RED**

Run: `go test ./internal/auth`

Expected: compilation failure for the missing token manager.

- [ ] **Step 3: Implement minimal token manager**

Use `github.com/golang-jwt/jwt/v5`, registered claims, an injectable clock for deterministic tests, and explicit HS256 method validation.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./internal/auth`

Expected: all token tests pass.

### Task 4: Registration and login service

**Files:**
- Create: `backend/internal/auth/service.go`
- Create: `backend/internal/auth/service_test.go`

- [ ] **Step 1: Write failing registration tests**

Test required/valid email, eight-character password minimum, lowercase normalization, bcrypt hashing, default `user` role, safe returned data, and duplicate-email propagation.

- [ ] **Step 2: Write failing login/current-user tests**

Test successful password verification and token issuance, identical `ErrInvalidCredentials` for unknown email and wrong password, and current-user lookup by UUID.

- [ ] **Step 3: Verify RED**

Run: `go test ./internal/auth`

Expected: compilation failures for service inputs and methods.

- [ ] **Step 4: Implement service behavior**

Normalize email, validate with `net/mail`, count password Unicode characters, hash with bcrypt default cost, persist safe user fields, compare password hashes, issue JWT, and expose `CurrentUser`.

- [ ] **Step 5: Verify GREEN**

Run: `go test ./internal/auth`

Expected: all service tests pass.

### Task 5: JWT middleware and HTTP handlers

**Files:**
- Replace: `backend/internal/middleware/doc.go`
- Create: `backend/internal/middleware/auth.go`
- Create: `backend/internal/middleware/auth_test.go`
- Create: `backend/internal/auth/handler.go`
- Create: `backend/internal/auth/handler_test.go`

- [ ] **Step 1: Write failing middleware tests**

Test missing header, wrong scheme, invalid token, valid context `user_id`/`role`, and aborted 401 responses.

- [ ] **Step 2: Write failing handler tests**

Test request JSON binding, 201 registration, 409 duplicate, 401 login failure, successful login envelope, current-user envelope, 404 deleted user, and 503 unavailable repository.

- [ ] **Step 3: Verify RED**

Run: `go test ./internal/auth ./internal/middleware`

Expected: compilation failures for middleware and handler constructors.

- [ ] **Step 4: Implement middleware and handlers**

Parse `Authorization: Bearer <token>`, set context values, return uniform errors, bind DTOs, call the service, map known errors to required statuses, and never serialize password hashes.

- [ ] **Step 5: Verify GREEN**

Run: `go test ./internal/auth ./internal/middleware`

Expected: all tests pass.

### Task 6: Router and server composition

**Files:**
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`
- Create: `backend/internal/router/auth_flow_test.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Write failing full-flow router test**

Using an in-memory `user.Repository`, register a user, assert stored bcrypt rather than plaintext, log in, extract the JWT, call `/api/me`, and verify the safe user. Retain the existing health test and add database-unavailable 503 coverage.

- [ ] **Step 2: Verify RED**

Run: `go test ./internal/router`

Expected: compilation or route-not-found failures.

- [ ] **Step 3: Compose dependencies**

Change `router.New` to accept a `Dependencies` value containing `user.Repository` and JWT secret. Register public auth routes and a middleware-protected `/api/me`. In `main`, retain the GORM connection, create the repository even when DB is nil, pass the JWT secret, and log a warning for the development secret.

- [ ] **Step 4: Verify GREEN and health regression**

Run: `go test ./internal/router ./internal/health`

Expected: full auth flow and original health tests pass.

### Task 7: Idempotent schema and environment documentation

**Files:**
- Modify: `database/schema.sql`
- Modify: `backend/.env.example`
- Modify: `README.md`
- Replace: `docs/API_DOC.md`
- Modify: `docs/DEPLOY.md`

- [ ] **Step 1: Update idempotent users migration**

Enable `uuid-ossp`, include nickname and school in new-table creation, and add both columns with `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` for deployed databases. Preserve existing data, status, constraints, and indexes.

- [ ] **Step 2: Document JWT secret**

Add a placeholder `JWT_SECRET` to `.env.example` and prominent production warnings to README and deployment documentation, including an OpenSSL generation command.

- [ ] **Step 3: Document API and curl tests**

Write complete register, login, and `/api/me` request/response/status documentation. Include curl commands and shell token extraction without real secrets.

### Task 8: Final verification

**Files:**
- Modify only when verification reveals a defect.

- [ ] **Step 1: Format and tidy**

Run: `gofmt -w cmd internal`

Run: `GOTOOLCHAIN=local go mod tidy`

- [ ] **Step 2: Verify code**

Run: `go test ./...`

Run: `go vet ./...`

Run: `go build ./cmd/server`

Expected: all commands exit zero under Go 1.22.5.

- [ ] **Step 3: Run HTTP smoke test**

Run the router full-flow test uncached and start the binary without database configuration to confirm `/api/health` still returns HTTP 200 while registration returns HTTP 503.

- [ ] **Step 4: Audit requirements and diff**

Check every requested field, route, status, SQL statement, curl example, and production JWT warning. Run `git diff --check` and inspect `git status --short` before handoff.

## Repository Note

The current branch is `main` and clean except for the approved design/plan documents. Implementation should occur in an isolated feature worktree or, if the user explicitly prefers, on a feature branch in place. No production database or real `.env` is modified by this plan.
