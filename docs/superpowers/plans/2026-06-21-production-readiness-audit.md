# Production Readiness Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct production-readiness defects found in the MatchLab frontend/backend audit without adding product features.

**Architecture:** Keep the existing Next.js client and Gin/GORM server boundaries. Add configurable origin policy at the Gin edge, enforce production configuration before server startup, keep API envelopes consistent, and harden the existing frontend request layer so expired credentials cannot leave a stale authenticated UI. Preserve PostgreSQL repositories and the established `/api/me/matches` contract while adding the explicitly requested authenticated compatibility path `/api/matches`.

**Tech Stack:** Go 1.22, Gin, GORM/PostgreSQL, Next.js 16, React 19, TypeScript, Vitest

---

### Task 1: Add configurable CORS at the API edge

**Files:**
- Create: `backend/internal/middleware/cors.go`
- Create: `backend/internal/middleware/cors_test.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`
- Modify: `backend/.env.example`

- [ ] Write a failing middleware test proving an allowed frontend origin receives `Access-Control-Allow-Origin`, `Vary: Origin`, allowed methods/headers, and a 204 preflight response while an unlisted origin receives no allow-origin header.
- [ ] Run `go test ./internal/middleware -run CORS -v` and confirm failure because the middleware is absent.
- [ ] Implement exact-origin matching from a parsed comma-separated `CORS_ALLOWED_ORIGINS` configuration and register it before API routes.
- [ ] Add router coverage proving configured origins work through the assembled application.
- [ ] Run middleware, config, and router tests and confirm they pass.

### Task 2: Standardize edge responses and resolve the matches route inconsistency

**Files:**
- Modify: `backend/internal/health/handler.go`
- Modify: `backend/internal/health/handler_test.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`
- Modify: `docs/API_DOC.md`

- [ ] Change the existing health test expectation first so success must be `{ "data": { "ok": true, "message": "MatchLab API running" } }` and verify it fails.
- [ ] Add failing router tests for JSON 404, JSON 405, and authenticated `/api/matches` compatibility behavior.
- [ ] Wrap health data, add `NoRoute`/`NoMethod` JSON errors, enable method-not-allowed handling, and map `/api/matches` to the same authenticated handler as `/api/me/matches`.
- [ ] Document `/api/me/matches` as canonical and `/api/matches` as a compatibility alias.
- [ ] Run health and router tests and confirm they pass.

### Task 3: Refuse unsafe production startup

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/.env.example`
- Modify: `docs/DEPLOY.md`

- [ ] Write failing tests proving release mode rejects the development JWT fallback and incomplete PostgreSQL configuration while debug mode retains health-only degraded startup support.
- [ ] Add `ValidateRuntime` and call it before opening the database; make database connection failure fatal in release mode.
- [ ] Set production-oriented example values (`GIN_MODE=release`, explicit CORS origin) and document the startup checks.
- [ ] Run config tests and `go test ./...`.

### Task 4: Harden the frontend API/session layer

**Files:**
- Modify: `frontend/lib/api.test.ts`
- Modify: `frontend/lib/api.ts`
- Modify: `frontend/next.config.ts`
- Modify: `frontend/.env.example`
- Modify: `docs/FRONTEND.md`

- [ ] Write failing tests proving a protected 401 removes the stored token and notifies auth subscribers, malformed JSON becomes a friendly `ApiError`, and the no-env default is same-origin `/api`.
- [ ] Run `npm test -- lib/api.test.ts` and confirm the new assertions fail for the expected reasons.
- [ ] Implement 401 token invalidation, safe JSON parsing, and a same-origin `/api` fallback while retaining the configurable full-URL development proxy.
- [ ] Update environment/deployment documentation and run focused tests.

### Task 5: Align deploy artifacts with the frontend/backend topology

**Files:**
- Modify: `deploy/nginx-matchlab.conf`
- Create: `deploy/matchlab-frontend.service`
- Modify: `docs/DEPLOY.md`
- Modify: `docs/FRONTEND.md`

- [ ] Add the Next.js `/` proxy while preserving `/api/` routing to Go and standard forwarded headers.
- [ ] Add a hardened systemd unit for the Next.js production service on port 3000.
- [ ] Verify production browser traffic uses same-origin `/api` and no production frontend code depends on localhost.

### Task 6: Full verification and end-to-end evidence

**Files:**
- Modify: `docs/API_DOC.md` only if validation commands reveal a documentation mismatch.

- [ ] Run `gofmt` on changed Go files, `go test ./...`, `go vet ./...`, and build `cmd/server/main.go`.
- [ ] Run `npm ci`, `npm test`, `npm run lint`, and `npm run build`.
- [ ] Enumerate Gin routes and compare all frontend request paths against registered backend paths.
- [ ] Verify bcrypt hashing, JWT Bearer propagation, admin middleware ordering, PostgreSQL GORM repositories, activity/application transactions, and match upserts from code and tests.
- [ ] Execute the authorized two-user API flow against the deployed API with unique audit accounts: register A/B, login A/B, create activity, submit B questionnaire, apply as B, approve as A, generate B recommendations, and read both `/api/me/matches` and `/api/matches` where available.
- [ ] Record any external deployment-version limitation separately from source-code verification; do not claim undeployed fixes are live.

