# MatchLab Frontend MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and verify the complete MatchLab Next.js frontend MVP in `frontend/` while leaving `backend/` unchanged.

**Architecture:** Use the Next.js App Router with client components for JWT-aware API flows, a typed shared request layer, focused reusable UI components, and page-local forms. Production runs as a Next.js Node service behind Nginx so the dynamic activity route remains valid for arbitrary IDs.

**Tech Stack:** Next.js, React, TypeScript, Tailwind CSS, Vitest, React Testing Library

---

### Task 1: Create the frontend project skeleton

**Files:**
- Delete: `frontend/.gitkeep`
- Create: `frontend/package.json`
- Create: `frontend/package-lock.json`
- Create: `frontend/next.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/postcss.config.mjs`
- Create: `frontend/eslint.config.mjs`
- Create: `frontend/vitest.config.ts`
- Create: `frontend/vitest.setup.ts`
- Create: `frontend/.gitignore`
- Create: `frontend/.env.example`
- Create: `frontend/app/globals.css`
- Create: `frontend/app/layout.tsx`

- [ ] Verify the main repository is clean and create the isolated `feature/frontend-mvp` worktree.
- [ ] Create minimal configuration and package scripts for `dev`, `build`, `start`, `lint`, and `test`.
- [ ] Install dependencies and commit the generated lockfile.
- [ ] Run `npm test`, `npm run lint`, and `npm run build` to establish the frontend baseline.
- [ ] Commit with `chore: scaffold frontend application`.

### Task 2: Build the typed API and data helper layer with TDD

**Files:**
- Create: `frontend/lib/api.test.ts`
- Create: `frontend/lib/api.ts`
- Create: `frontend/lib/types.ts`
- Create: `frontend/lib/forms.test.ts`
- Create: `frontend/lib/forms.ts`

- [ ] Write failing tests for base URL normalization, token headers, successful JSON extraction, backend error messages, and 401 handling.
- [ ] Run the focused test and confirm failure because the API helper is absent.
- [ ] Implement `API_BASE_URL`, `getToken`, `clearToken`, `setToken`, `request`, and `ApiError` minimally.
- [ ] Run the focused test and confirm it passes.
- [ ] Write failing tests for comma-separated tag parsing and optional response-array normalization.
- [ ] Implement the form/data helpers and shared API response types.
- [ ] Run all frontend tests and commit with `feat: add typed frontend API client`.

### Task 3: Add the visual system, approved artwork, and shared components

**Files:**
- Create: `frontend/public/images/matchlab-hero.png`
- Modify: `frontend/app/globals.css`
- Create: `frontend/components/Navbar.test.tsx`
- Create: `frontend/components/Navbar.tsx`
- Create: `frontend/components/TagList.tsx`
- Create: `frontend/components/EmptyState.tsx`
- Create: `frontend/components/StatCard.tsx`
- Create: `frontend/components/ActivityCard.test.tsx`
- Create: `frontend/components/ActivityCard.tsx`

- [ ] Copy the approved second original generated illustration into the project.
- [ ] Define accessible indigo, slate, cyan, cream, and coral design tokens plus responsive page/form/table utilities.
- [ ] Write failing component tests for navbar role visibility/logout and activity-card metadata/detail links.
- [ ] Implement the five requested shared components.
- [ ] Run component tests and verify green.
- [ ] Commit with `feat: add frontend visual system and shared components`.

### Task 4: Implement landing, authentication, and dashboard routes

**Files:**
- Create: `frontend/app/page.test.tsx`
- Create: `frontend/app/page.tsx`
- Create: `frontend/app/login/page.test.tsx`
- Create: `frontend/app/login/page.tsx`
- Create: `frontend/app/dashboard/page.test.tsx`
- Create: `frontend/app/dashboard/page.tsx`

- [ ] Write failing tests for landing copy and calls to action.
- [ ] Implement the artwork-led landing page and three-step flow.
- [ ] Write failing tests for login/register mode switching, token persistence, friendly errors, and dashboard admin visibility.
- [ ] Implement login/register and authenticated dashboard pages.
- [ ] Run focused tests and commit with `feat: add landing authentication and dashboard pages`.

### Task 5: Implement the activity discovery and publishing flow

**Files:**
- Create: `frontend/app/activities/page.test.tsx`
- Create: `frontend/app/activities/page.tsx`
- Create: `frontend/app/activities/create/page.test.tsx`
- Create: `frontend/app/activities/create/page.tsx`
- Create: `frontend/app/activities/[id]/page.test.tsx`
- Create: `frontend/app/activities/[id]/page.tsx`

- [ ] Write failing tests for loading, empty, error, and populated activity-list states.
- [ ] Implement activity listing with cards and the publish shortcut.
- [ ] Write failing tests for create-form array conversion and successful navigation.
- [ ] Implement authenticated activity creation.
- [ ] Write failing tests for detail rendering, unauthenticated state, and application submission.
- [ ] Implement dynamic activity detail and application form.
- [ ] Run focused tests and commit with `feat: add activity browsing creation and application flow`.

### Task 6: Implement questionnaire, recommendation, and applications routes

**Files:**
- Create: `frontend/app/questionnaire/page.test.tsx`
- Create: `frontend/app/questionnaire/page.tsx`
- Create: `frontend/app/match/page.test.tsx`
- Create: `frontend/app/match/page.tsx`
- Create: `frontend/app/applications/page.test.tsx`
- Create: `frontend/app/applications/page.tsx`

- [ ] Write failing questionnaire tests for array conversion, submission, and returned profile rendering.
- [ ] Implement the questionnaire and profile result view.
- [ ] Write failing match tests for recommendation generation and saved-match refresh.
- [ ] Implement score, detail-score, reason, and saved-recommendation cards.
- [ ] Write failing tests for application loading, empty state, and status rendering.
- [ ] Implement the current-user applications page.
- [ ] Run focused tests and commit with `feat: add questionnaire recommendation and application pages`.

### Task 7: Implement the administrator dashboard

**Files:**
- Create: `frontend/app/admin/page.test.tsx`
- Create: `frontend/app/admin/page.tsx`

- [ ] Write failing tests for unauthenticated handling, non-admin denial, stats rendering, table rendering, and empty feedbacks.
- [ ] Implement `/me` role validation followed by parallel admin requests.
- [ ] Render six statistic cards and responsive user, activity, application, and feedback tables.
- [ ] Run focused tests and commit with `feat: add administrator dashboard page`.

### Task 8: Document, verify, and browser-test the complete frontend

**Files:**
- Create: `docs/FRONTEND.md`
- Modify: `README.md`
- Modify: `frontend/README.md` if generated

- [ ] Document `npm install`, `npm run dev`, environment configuration, testing, production build, systemd, and Nginx reverse proxy deployment.
- [ ] Confirm `git diff -- backend database` is empty.
- [ ] Run `npm test` and confirm zero failures.
- [ ] Run `npm run lint` and confirm zero errors.
- [ ] Run `npm run build` and confirm success.
- [ ] Start the development server and inspect all ten routes at desktop and mobile widths.
- [ ] Fix only observed regressions through new failing tests.
- [ ] Commit with `docs: add frontend development and deployment guide`.
