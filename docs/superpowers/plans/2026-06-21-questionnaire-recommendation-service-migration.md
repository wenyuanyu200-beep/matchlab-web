# Questionnaire Recommendation Service Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicit service layers, migrate matches to independent recommendation fields without data loss, and prove the four JWT routes compile and do not return 404.

**Architecture:** Handlers own HTTP/JWT concerns, services own validation and business orchestration, and repositories own GORM persistence. Existing scoring/profile pure functions remain unchanged. The schema migration adds and backfills fields, preserves legacy columns and rows, repairs duplicate legacy algorithm versions, and then creates an idempotent unique index.

**Tech Stack:** Go 1.22, Gin, GORM, PostgreSQL JSONB, Go testing package.

---

## File Structure

- Create `backend/internal/questionnaire/service.go`: questionnaire validation, profile generation, and repository orchestration.
- Create `backend/internal/questionnaire/service_test.go`: service validation and repository-call tests.
- Modify `backend/internal/questionnaire/handler.go`: depend on service interface and expose `Profile`.
- Modify `backend/internal/questionnaire/handler_test.go`: inject service and preserve HTTP contract coverage.
- Create `backend/internal/match/service.go`: recommendation validation, filtering, ranking, and persistence orchestration.
- Create `backend/internal/match/service_test.go`: own-activity exclusion, limit, ranking, and persistence tests.
- Modify `backend/internal/match/model.go`: independent matches persistence fields.
- Modify `backend/internal/match/repository.go`: read/write new columns only.
- Modify `backend/internal/match/handler.go`: depend on service and expose `MyMatches`.
- Modify `backend/internal/match/handler_test.go`: inject service and preserve HTTP contract coverage.
- Modify `backend/internal/router/router.go`, `router_test.go`: construct services and prove authenticated routes are registered.
- Modify `database/schema.sql`: repeatable additive migration and safe unique-index preparation.
- Modify `docs/API_DOC.md`: new field layout plus deployment commit/file checks.

### Task 1: Questionnaire Service Layer

**Files:**
- Create: `backend/internal/questionnaire/service.go`
- Create: `backend/internal/questionnaire/service_test.go`
- Modify: `backend/internal/questionnaire/handler.go`
- Modify: `backend/internal/questionnaire/handler_test.go`

- [ ] **Step 1: Write failing service tests**

Define the desired service API and verify validation happens before persistence:

```go
type Service interface {
    Submit(ctx context.Context, userID uuid.UUID, mode string, answers Answers) (*Questionnaire, *Profile, error)
    Profile(ctx context.Context, userID uuid.UUID) (*Profile, error)
}

func TestServiceSubmitGeneratesProfileAndPersists(t *testing.T) {
    repository := newServiceRepository()
    service := NewService(repository)
    questionnaire, profile, err := service.Submit(context.Background(), uuid.New(), " activity ", Answers{Interests: StringList{"硬件"}})
    if err != nil || questionnaire.Mode != "activity" || profile.Tags[0] != "硬件" {
        t.Fatalf("questionnaire=%#v profile=%#v err=%v", questionnaire, profile, err)
    }
}

func TestServiceSubmitRejectsUnsupportedMode(t *testing.T) {
    repository := newServiceRepository()
    _, _, err := NewService(repository).Submit(context.Background(), uuid.New(), "team", Answers{})
    if !errors.Is(err, ErrInvalidMode) || repository.submitCalls != 0 {
        t.Fatalf("err=%v calls=%d", err, repository.submitCalls)
    }
}
```

- [ ] **Step 2: Run RED test**

Run: `go test ./internal/questionnaire -run TestService -v`

Expected: build failure because `Service`, `NewService`, and `ErrInvalidMode` do not exist.

- [ ] **Step 3: Implement service and thin handler**

Create `service.go` with `Service`, concrete service, `ErrInvalidMode`, normalized mode/answers, `GenerateProfile`, and repository calls. Change handler construction to `NewHandler(service Service)`, keep JWT and JSON binding in handler, map `ErrInvalidMode` to 400, and rename `CurrentProfile` to `Profile`.

- [ ] **Step 4: Run GREEN tests**

Run: `go test ./internal/questionnaire -v`

Expected: all questionnaire tests pass.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/questionnaire
git commit -m "refactor: add questionnaire service layer"
```

### Task 2: Match Service and Independent Persistence Model

**Files:**
- Create: `backend/internal/match/service.go`
- Create: `backend/internal/match/service_test.go`
- Modify: `backend/internal/match/model.go`
- Modify: `backend/internal/match/handler.go`
- Modify: `backend/internal/match/handler_test.go`

- [ ] **Step 1: Write failing service tests**

Define:

```go
type Service interface {
    Recommend(ctx context.Context, userID uuid.UUID, targetType string, limit *int) ([]Recommendation, error)
    MyMatches(ctx context.Context, userID uuid.UUID) ([]SavedRecommendation, error)
}
```

Test that `Recommend` defaults to 10, rejects non-activity targets and limits outside 1–50, excludes `CreatorID == userID`, ranks results, and passes current questionnaire ID plus recommendations to `UpsertMatches`. Test that `MyMatches` delegates with the authenticated user ID.

- [ ] **Step 2: Run RED test**

Run: `go test ./internal/match -run TestService -v`

Expected: build failure because match `Service` and `NewService` do not exist.

- [ ] **Step 3: Update persistence model**

Replace `Record.Explanation` as the active persistence source with:

```go
TargetID        uuid.UUID    `gorm:"column:target_id;type:uuid"`
TargetType      string       `gorm:"column:target_type"`
Algorithm       string       `gorm:"column:algorithm"`
AlgorithmVersion string      `gorm:"column:algorithm_version"`
DetailScores    DetailScores `gorm:"column:detail_scores;type:jsonb"`
Reason          string       `gorm:"column:reason"`
```

Add JSONB `Value` and `Scan` methods to `DetailScores`. Keep legacy fields out of active writes.

- [ ] **Step 4: Implement service and thin handler**

Move validation, signal/candidate reads, own-activity filtering, `RankActivities`, and `UpsertMatches` from handler into service. Handler binds input and maps service errors only. Rename `CurrentMatches` to `MyMatches`.

- [ ] **Step 5: Run GREEN tests**

Run: `go test ./internal/match -v`

Expected: all match tests pass.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/match
git commit -m "refactor: add match service and direct result fields"
```

### Task 3: Repository Writes and Repeatable Schema Migration

**Files:**
- Modify: `backend/internal/match/repository.go`
- Modify: `database/schema.sql`

- [ ] **Step 1: Write failing model persistence test**

Add a test proving JSONB round-trip for `DetailScores` and that a new record is populated with target/activity metadata:

```go
func TestNewRecordUsesDirectRecommendationFields(t *testing.T) {
    record := newRecord(userID, questionnaireID, recommendation, time.Now())
    if record.TargetID != recommendation.Activity.ID || record.TargetType != "activity" || record.Algorithm != "rules" {
        t.Fatalf("unexpected record: %#v", record)
    }
    if record.DetailScores != recommendation.DetailScores || record.Reason != recommendation.Reason {
        t.Fatalf("missing direct fields: %#v", record)
    }
}
```

- [ ] **Step 2: Run RED test**

Run: `go test ./internal/match -run TestNewRecordUsesDirectRecommendationFields -v`

Expected: build failure because `newRecord` does not exist.

- [ ] **Step 3: Implement direct-column repository mapping**

Extract `newRecord`, update conflict assignments to `target_id`, `target_type`, `algorithm`, `questionnaire_id`, `score`, `detail_scores`, `reason`, `status`, and `updated_at`. Update `ListMatches` to map `record.DetailScores` and `record.Reason` rather than legacy explanation.

- [ ] **Step 4: Implement idempotent schema migration**

Add all requested fields to both new table definition and an `ALTER TABLE matches ... ADD COLUMN IF NOT EXISTS` block. Backfill from legacy explanation. Use a `ROW_NUMBER()` CTE to change only duplicate rows with rank > 1 to `legacy-` plus the first 8 UUID characters. Create `matches_user_activity_algorithm_uq` with `CREATE UNIQUE INDEX IF NOT EXISTS`. Do not drop tables, columns, or rows.

- [ ] **Step 5: Run package and schema checks**

Run: `go test ./internal/match -v`

Expected: all tests pass.

Run: `rg -n "DROP TABLE|TRUNCATE|DELETE FROM matches" database/schema.sql`

Expected: no matches.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/match database/schema.sql
git commit -m "feat: migrate matches to direct recommendation fields"
```

### Task 4: Router Wiring and Non-404 Regression

**Files:**
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`

- [ ] **Step 1: Write failing authenticated-route test**

Issue valid JWT requests against the four URLs with unavailable repositories and assert none returns 404:

```go
tests := []struct{ method, path, body string }{
    {http.MethodPost, "/api/questionnaires", validQuestionnaireBody},
    {http.MethodGet, "/api/me/profile", ""},
    {http.MethodPost, "/api/match/recommend", `{"target_type":"activity"}`},
    {http.MethodGet, "/api/me/matches", ""},
}
```

Also assert the expected service-unavailable status where the request reaches persistence.

- [ ] **Step 2: Run RED test after handler renames**

Run: `go test ./internal/router -run TestQuestionnaireAndMatchRoutesAreRegistered -v`

Expected: build failure or 404 until router constructs services and registers the renamed handler methods.

- [ ] **Step 3: Wire services and exact routes**

Construct questionnaire and match repositories, then services, then handlers. Register `Submit`, `Profile`, `Recommend`, and `MyMatches` behind `middleware.RequireAuth(tokens)`.

- [ ] **Step 4: Run router and full regression tests**

Run: `go test ./internal/router -v`

Run: `go test ./...`

Expected: all tests pass, including existing health/auth/activity/application packages.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/router
git commit -m "fix: wire questionnaire recommendation services"
```

### Task 5: API Documentation, Build, and Deployment Proof

**Files:**
- Modify: `docs/API_DOC.md`

- [ ] **Step 1: Update docs**

Document direct matches fields, four curl commands, SQL verification, `git rev-parse HEAD`, `find backend/internal/{questionnaire,match,algorithm}`, schema execution, exact build command, binary timestamp, service restart, journal check, and public curl verification.

- [ ] **Step 2: Format and verify diff**

Run: `gofmt -w internal/questionnaire internal/match internal/router`

Run from repository root: `git diff --check`

Expected: no errors.

- [ ] **Step 3: Run fresh required verification**

From `backend`:

```bash
go test -count=1 ./...
go vet ./...
go build -o bin/matchlab-api cmd/server/main.go
```

Expected: all commands exit 0. Confirm `bin/matchlab-api` exists, then ensure it is not accidentally committed.

- [ ] **Step 4: Commit documentation**

```bash
git add docs/API_DOC.md
git commit -m "docs: add verifiable recommendation deployment"
```

- [ ] **Step 5: Final requirement audit**

Confirm the required model/repository/service/handler files exist, all four routes are registered, schema contains additive fields and safe unique handling, KM remains 24/18, legacy fields remain, no tables or rows are removed, and the working tree is clean.
