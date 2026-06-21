# Questionnaire, Profile, and Activity Recommendation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add JWT-protected questionnaire submission, generated user profiles, explainable activity recommendations with latest-result persistence, and reusable KM/greedy matching algorithms.

**Architecture:** Keep questionnaire/profile and recommendation concerns in separate internal packages with repository interfaces around GORM. Keep profile generation and scoring as deterministic pure functions, while repository transactions own database consistency and match upserts. Register the packages through the existing router without changing current health, auth, activity, or application behavior.

**Tech Stack:** Go 1.22, Gin, GORM, PostgreSQL JSONB, google/uuid, Go standard testing package.

---

## File Structure

- Create `backend/internal/questionnaire/model.go`: JSONB value types plus questionnaire/profile models and API input types.
- Create `backend/internal/questionnaire/profile.go`: deterministic tag, score, and summary generation.
- Create `backend/internal/questionnaire/repository.go`: questionnaire/profile repository contract and transactional GORM implementation.
- Create `backend/internal/questionnaire/handler.go`: submit and current-profile endpoints.
- Create `backend/internal/questionnaire/profile_test.go`, `handler_test.go`: pure rule and HTTP behavior tests.
- Create `backend/internal/match/model.go`: detail scores, explanation, match persistence, and response models.
- Create `backend/internal/match/scorer.go`: deterministic weighted scoring, keyword matching, sorting inputs, and reason generation.
- Create `backend/internal/match/repository.go`: candidate reads, latest questionnaire reads, match upserts, and saved-result reads.
- Create `backend/internal/match/handler.go`: recommendation and current-match endpoints.
- Create `backend/internal/match/scorer_test.go`, `handler_test.go`: scoring and HTTP behavior tests.
- Create `backend/internal/algorithm/km.go`, `km_test.go`: maximum-weight assignment and row-order greedy baseline.
- Modify `backend/internal/router/router.go`, `router_test.go`: wire dependencies and protected routes.
- Modify `database/schema.sql`: additive compatible columns, constraints, and indexes.
- Modify `docs/API_DOC.md`: endpoint examples, smoke test, SQL verification, and KM test instructions.

### Task 1: Profile Generation Rules

**Files:**
- Create: `backend/internal/questionnaire/model.go`
- Create: `backend/internal/questionnaire/profile.go`
- Create: `backend/internal/questionnaire/profile_test.go`

- [ ] **Step 1: Write failing profile tests**

Add table-driven tests that call the desired API:

```go
func TestGenerateProfile(t *testing.T) {
    answers := Answers{
        Interests: []string{"电赛", "STM32", "硬件", "电赛"},
        Skills: []string{"嵌入式", "焊接"},
        ActivityTypes: []string{"competition", "project"},
        AvailableTime: "周末下午",
        Goal: "找队友一起参加比赛",
        CommunicationStyle: "稳定沟通",
    }
    profile := GenerateProfile("activity", answers)
    wantTags := []string{"电赛", "STM32", "硬件", "嵌入式", "焊接", "competition", "project"}
    if !reflect.DeepEqual(profile.Tags, wantTags) {
        t.Fatalf("tags=%v, want %v", profile.Tags, wantTags)
    }
    if profile.Scores.InterestScore != 80 || profile.Scores.SkillScore != 75 || profile.Scores.TimeScore != 70 || profile.Scores.GoalScore != 80 || profile.Scores.CommunicationScore != 75 {
        t.Fatalf("unexpected scores: %#v", profile.Scores)
    }
    if !strings.Contains(profile.Summary, "电赛") || !strings.Contains(profile.Summary, "嵌入式") {
        t.Fatalf("unexpected summary: %s", profile.Summary)
    }
}
```

Also test blank entries, case-insensitive duplicate English tags, missing dimensions receiving their documented base scores, and absence of prohibited product terms.

- [ ] **Step 2: Run the test and verify RED**

Run: `go test ./internal/questionnaire -run TestGenerateProfile -v`

Expected: compile failure because `Answers` and `GenerateProfile` do not exist.

- [ ] **Step 3: Implement JSONB models and minimal profile generator**

Define `Answers`, `ProfileScores`, `Questionnaire`, and `Profile`. Implement `driver.Valuer`/`sql.Scanner` JSONB wrappers, stable case-insensitive tag de-duplication, the requested target scores for non-empty dimensions, lower base scores for missing dimensions, and a deterministic Chinese summary built from the first few interest and skill tags.

- [ ] **Step 4: Run profile tests and verify GREEN**

Run: `go test ./internal/questionnaire -v`

Expected: all questionnaire package tests pass.

- [ ] **Step 5: Commit the unit**

```bash
git add backend/internal/questionnaire/model.go backend/internal/questionnaire/profile.go backend/internal/questionnaire/profile_test.go
git commit -m "feat: add questionnaire profile generation"
```

### Task 2: Questionnaire Persistence and HTTP Endpoints

**Files:**
- Create: `backend/internal/questionnaire/repository.go`
- Create: `backend/internal/questionnaire/handler.go`
- Create: `backend/internal/questionnaire/handler_test.go`

- [ ] **Step 1: Write failing handler tests with an in-memory repository**

Specify these behaviors:

```go
func TestSubmitUsesAuthenticatedUserAndReturnsQuestionnaireAndProfile(t *testing.T) { /* POST, assert 201 and repository user ID */ }
func TestSubmitUpdatesExistingProfile(t *testing.T) { /* submit twice, assert one profile */ }
func TestCurrentProfileReturnsProfile(t *testing.T) { /* GET, assert 200 */ }
func TestSubmitRejectsInvalidBody(t *testing.T) { /* POST malformed/empty mode, assert 400 */ }
func TestCurrentProfileMapsNotFound(t *testing.T) { /* GET, assert 404 profile_not_found */ }
```

The fake repository implements:

```go
type Repository interface {
    Submit(context.Context, uuid.UUID, string, Answers, GeneratedProfile) (*Questionnaire, *Profile, error)
    GetProfile(context.Context, uuid.UUID) (*Profile, error)
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/questionnaire -run 'TestSubmit|TestCurrentProfile' -v`

Expected: compile failure because repository and handler APIs do not exist.

- [ ] **Step 3: Implement transactional GORM repository**

Inside one transaction: lock/read the user's maximum questionnaire version, create the completed questionnaire, read the user's nickname, and upsert profile on `user_id`. Use `clause.OnConflict` to update only `profile_type`, `tags`, `scores`, `summary`, and `updated_at`; preserve legacy profile fields.

- [ ] **Step 4: Implement handlers and error mapping**

`Submit` reads the JWT subject from Gin context, binds `{mode, answers}`, normalizes/validates `mode`, calls `GenerateProfile`, persists both records, and returns `201` with both objects. `CurrentProfile` returns `200`, `404 profile_not_found`, or `503 service_unavailable` as appropriate.

- [ ] **Step 5: Run package tests and verify GREEN**

Run: `go test ./internal/questionnaire -v`

Expected: all tests pass.

- [ ] **Step 6: Commit the unit**

```bash
git add backend/internal/questionnaire/repository.go backend/internal/questionnaire/handler.go backend/internal/questionnaire/handler_test.go
git commit -m "feat: add questionnaire and profile endpoints"
```

### Task 3: Recommendation Scoring and Reasons

**Files:**
- Create: `backend/internal/match/model.go`
- Create: `backend/internal/match/scorer.go`
- Create: `backend/internal/match/scorer_test.go`

- [ ] **Step 1: Write failing scoring tests**

Use a profile/questionnaire that fully matches an activity and assert exact caps:

```go
func TestScoreActivityUsesRequestedWeights(t *testing.T) {
    got := ScoreActivity(UserSignals{
        Interests: []string{"电赛", "STM32", "硬件"},
        Skills: []string{"嵌入式", "焊接"},
        ActivityTypes: []string{"competition"},
        AvailableTime: "周末下午",
        Goal: "参加电赛比赛",
    }, activity.Activity{Type: "competition", Title: "电赛组队", Description: "参加比赛", Tags: activity.StringList{"电赛", "STM32", "硬件"}, PreferredTags: activity.StringList{"嵌入式", "焊接"}, TimeText: "周末下午"})
    if got.DetailScores != (DetailScores{Interest: 30, Skill: 25, Type: 20, Time: 10, Goal: 15}) || got.Score != 100 {
        t.Fatalf("unexpected score: %#v", got)
    }
}
```

Add tests for partial overlap, case-insensitive English matching, no overlap, natural reasons containing matched tags, prohibited-term absence, score sorting, and limit behavior.

- [ ] **Step 2: Run scoring tests and verify RED**

Run: `go test ./internal/match -run 'TestScore|TestRank' -v`

Expected: compile failure because scorer types/functions do not exist.

- [ ] **Step 3: Implement minimal scorer**

Implement normalized set overlap for interest and skill, exact normalized activity type membership, rune/letter/digit keyword extraction plus containment for time and goal, integer capped scores, deterministic reason clauses, and stable sorting by score then creation time.

- [ ] **Step 4: Run scoring tests and verify GREEN**

Run: `go test ./internal/match -v`

Expected: all scorer tests pass.

- [ ] **Step 5: Commit the unit**

```bash
git add backend/internal/match/model.go backend/internal/match/scorer.go backend/internal/match/scorer_test.go
git commit -m "feat: add explainable activity scoring"
```

### Task 4: Recommendation Persistence and HTTP Endpoints

**Files:**
- Create: `backend/internal/match/repository.go`
- Create: `backend/internal/match/handler.go`
- Create: `backend/internal/match/handler_test.go`

- [ ] **Step 1: Write failing service/handler tests**

Specify:

```go
func TestRecommendExcludesOwnActivitiesAndPersistsRankedResults(t *testing.T) { /* assert creator exclusion, ranking, upsert calls */ }
func TestRecommendDefaultsLimitAndRejectsUnsupportedTarget(t *testing.T) { /* target validation */ }
func TestRecommendRequiresProfile(t *testing.T) { /* assert 400 profile_required */ }
func TestCurrentMatchesReturnsSavedActivityResults(t *testing.T) { /* assert activity, score, detail_scores, reason */ }
```

The repository contract exposes `LoadSignals`, `ListCandidates`, `UpsertMatches`, and `ListMatches`. Its fake tracks the authenticated user ID and persisted records.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/match -run 'TestRecommend|TestCurrentMatches' -v`

Expected: compile failure because handler/repository APIs do not exist.

- [ ] **Step 3: Implement GORM repository and latest-result upsert**

Read profile plus the latest completed questionnaire. Query recruiting activities with `creator_id <> ?`. In one transaction, upsert every selected result using `(user_id, activity_id, algorithm_version)` and update `questionnaire_id`, `score`, `explanation`, `status`, and `updated_at`. Join activities when listing saved matches and order by score descending then match updated time.

- [ ] **Step 4: Implement handlers**

`Recommend` accepts only `activity`, defaults limit to 10, rejects values outside 1–50, scores/ranks candidates, persists selected recommendations, and returns API response fields at the recommendation item level. `CurrentMatches` returns current saved rows. Map missing profile to `400 profile_required`, repository absence to `503`, and other failures to `500`.

- [ ] **Step 5: Run package tests and verify GREEN**

Run: `go test ./internal/match -v`

Expected: all match tests pass.

- [ ] **Step 6: Commit the unit**

```bash
git add backend/internal/match/repository.go backend/internal/match/handler.go backend/internal/match/handler_test.go
git commit -m "feat: persist and expose activity recommendations"
```

### Task 5: KM and Greedy Algorithms

**Files:**
- Create: `backend/internal/algorithm/km.go`
- Create: `backend/internal/algorithm/km_test.go`

- [ ] **Step 1: Write failing algorithm tests**

```go
func TestKMOutperformsRowGreedy(t *testing.T) {
    weights := [][]int{{9, 8, 1}, {8, 1, 1}, {1, 8, 8}}
    _, kmTotal, err := MaximumWeightMatching(weights)
    if err != nil || kmTotal != 24 { t.Fatalf("KM: total=%d err=%v", kmTotal, err) }
    _, greedyTotal, err := GreedyMatching(weights)
    if err != nil || greedyTotal != 18 { t.Fatalf("greedy: total=%d err=%v", greedyTotal, err) }
}
```

Also test empty, rectangular row<column, rectangular row>column, negative weights, and ragged matrices.

- [ ] **Step 2: Run algorithm tests and verify RED**

Run: `go test ./internal/algorithm -v`

Expected: compile failure because matching functions do not exist.

- [ ] **Step 3: Implement KM and greedy**

Validate rectangular input, pad the smaller dimension with zero-weight dummy nodes, run a maximum-weight Hungarian/KM assignment using labels and augmenting paths, remove dummy pairs from the result, and sum only real matrix weights. Greedy processes rows in input order and picks the highest unused real column, leaving surplus rows unmatched.

- [ ] **Step 4: Run algorithm tests and verify GREEN**

Run: `go test ./internal/algorithm -v`

Expected: all tests pass, with totals 24 and 18 for the required matrix.

- [ ] **Step 5: Commit the unit**

```bash
git add backend/internal/algorithm/km.go backend/internal/algorithm/km_test.go
git commit -m "feat: add KM maximum-weight matching"
```

### Task 6: Schema and Router Integration

**Files:**
- Modify: `database/schema.sql`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`

- [ ] **Step 1: Write failing route-protection tests**

Add a table-driven test that requests all four routes without a token and expects `401`:

```go
tests := []struct{ method, path string }{
    {http.MethodPost, "/api/questionnaires"},
    {http.MethodGet, "/api/me/profile"},
    {http.MethodPost, "/api/match/recommend"},
    {http.MethodGet, "/api/me/matches"},
}
```

- [ ] **Step 2: Run router test and verify RED**

Run: `go test ./internal/router -run TestRecommendationRoutesRequireAuthentication -v`

Expected: routes return 404 because they are not registered.

- [ ] **Step 3: Add compatible schema migration**

Use `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` for questionnaire `mode` and profile fields. Backfill safe defaults before applying non-null constraints. Keep existing tables and columns intact. Retain the existing matches unique constraint and add any missing index needed for latest user result ordering.

- [ ] **Step 4: Wire dependencies and routes**

Extend router dependencies only where tests benefit from injected repositories; otherwise construct GORM repositories from `Dependencies.DB`. Create both handlers and register all four routes behind one existing `RequireAuth(tokens)` middleware instance.

- [ ] **Step 5: Run router and full tests**

Run: `go test ./internal/router -v`

Expected: router tests pass.

Run: `go test ./...`

Expected: all packages pass, including existing health/auth/activity/application tests.

- [ ] **Step 6: Commit the integration**

```bash
git add database/schema.sql backend/internal/router/router.go backend/internal/router/router_test.go
git commit -m "feat: register questionnaire and recommendation routes"
```

### Task 7: API Documentation and Final Verification

**Files:**
- Modify: `docs/API_DOC.md`

- [ ] **Step 1: Update API documentation**

Document exact curl commands for user B login, questionnaire submit, profile fetch, recommendation request, and current matches. Include assertions using `jq`, this SQL check:

```sql
SELECT user_id, activity_id, algorithm_version, score,
       explanation->'detail_scores' AS detail_scores,
       explanation->>'reason' AS reason, updated_at
FROM matches
WHERE user_id = '<USER_B_UUID>'
ORDER BY score DESC, updated_at DESC;
```

Document `go test ./internal/algorithm -run TestKMOutperformsRowGreedy -v` and the expected KM 24 / greedy 18 totals. Keep all product copy within the campus activity and project collaboration positioning.

- [ ] **Step 2: Run formatting and diff checks**

Run: `gofmt -w internal/questionnaire internal/match internal/algorithm internal/router`

Run from repository root: `git diff --check`

Expected: no whitespace errors.

- [ ] **Step 3: Run fresh verification suite**

Run from `backend`:

```bash
go test ./...
go vet ./...
go build ./cmd/server
```

Expected: every command exits 0, all tests pass, and the server builds.

- [ ] **Step 4: Review requirement coverage**

Confirm every requested endpoint, all five score dimensions, reason output, self-created activity exclusion, latest-result upsert, profile upsert, schema compatibility, KM/greedy totals, API curls, deployment order, and prohibited wording constraints are represented in code or tests.

- [ ] **Step 5: Commit documentation**

```bash
git add docs/API_DOC.md
git commit -m "docs: add recommendation API and deployment tests"
```

- [ ] **Step 6: Prepare handoff**

Report the modified file list, whether schema changes are required, why production must run the idempotent schema migration rather than recreate the database, deployment commands, and the full curl flow requested by the user.
