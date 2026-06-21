# MatchLab Frontend MVP Design

## Goal

Build the first complete MatchLab web client in `frontend/` without changing the Go backend. The product remains a campus activity and project collaboration platform centered on activity partners, project teammates, questionnaires, and intelligent team recommendations.

## Scope

The frontend provides these routes:

- `/`: brand landing page and three-step product flow.
- `/login`: login and registration tabs.
- `/dashboard`: current-user overview and role-aware shortcuts.
- `/activities`: public activity list.
- `/activities/create`: authenticated activity publishing form.
- `/activities/[id]`: activity detail and authenticated application form.
- `/questionnaire`: authenticated questionnaire submission and generated profile.
- `/match`: recommendation generation and current saved recommendations.
- `/applications`: authenticated current-user applications.
- `/admin`: admin-only statistics and data tables.

Payment, chat, maps, frontend role management, complex animation, and social or dating positioning are out of scope.

## Architecture

Use Next.js App Router, TypeScript, React, and Tailwind CSS. Pages that access `localStorage` or call the public Go API are client components. Shared layouts and presentational components remain small and focused.

`frontend/lib/api.ts` owns the API base URL, token retrieval, authorization header, JSON handling, and friendly error extraction. `frontend/lib/types.ts` defines the response shapes used across pages. Pages unwrap the backend's `{ "data": ... }` envelope through the shared request helper rather than duplicating fetch logic.

Authentication stays intentionally simple for the MVP: successful login or registration stores the returned JWT in `localStorage`; protected pages redirect unauthenticated visitors to `/login`; the navbar can clear the token for local logout. Authorization remains enforced by the backend.

## Visual System

The approved direction is “indigo aurora”: a pale cream, misty cyan, and cool gray background; deep slate-blue body text; deep indigo headings; indigo primary actions; cyan accents; and restrained coral-orange highlights. Body copy uses strong contrast and avoids pale purple text on light surfaces.

The homepage uses the approved original wide artwork: two East Asian university students collaborating on an electronics project in a surreal environment of translucent bubbles, underwater light, orange flowers, and campus architecture. The illustration stays on the right while the left remains calm enough for readable headline text. The login page may reuse a cropped version. Data-heavy and form pages use only faint gradients and geometric accents so content remains easy to scan.

Responsive behavior uses a single-column mobile layout, collapsible or wrapping navigation, horizontally scrollable admin tables, and grid cards that expand on tablet and desktop widths.

## Components

- `Navbar`: brand navigation, authenticated shortcuts, and local logout.
- `ActivityCard`: activity summary, creator metadata, capacity, status, and detail link.
- `StatCard`: compact administrator metric.
- `EmptyState`: consistent empty-list presentation.
- `TagList`: accessible wrapping tag pills.
- Small page-local form controls and notices keep MVP implementation direct without adding a component framework.

## Data Flow

- Login and registration call their respective auth endpoints, store the JWT, and navigate to `/dashboard`.
- Dashboard and admin first load `/me`; admin data loads only after confirming `role === "admin"`.
- Activity list and detail pages load public activity endpoints. Creation and application attach the JWT automatically.
- Questionnaire submission converts comma-separated values into arrays and renders the returned profile.
- Recommendation generation posts `{ "target_type": "activity", "limit": 10 }`, then refreshes `/me/matches` so the screen reflects persisted results.
- Administrator data requests stats, users, activities, applications, and feedbacks in parallel after role validation.

The UI accepts empty arrays and missing optional fields without crashing. Creator, profile, recommendation, and application renderers use the backend's documented nested data while tolerating an extra data envelope.

## Error and Loading States

Every API-driven page shows an explicit loading state. Request failures render a friendly inline message derived from the backend `message` field when available. HTTP 401 prompts the user to log in again; HTTP 403 on `/admin` displays “无权限访问”. Successful form actions show a visible confirmation before navigation or refreshed data.

Forms prevent duplicate submission while pending and validate required fields in the browser. The API helper never hardcodes or logs tokens.

## Testing and Verification

Use the framework's lint and production build as whole-project gates. Add lightweight unit tests for API URL/header/error behavior and pure data conversion helpers. Verify the main routes at mobile and desktop widths in a browser, including empty/error states and role-gated admin navigation.

Required completion commands are:

```bash
cd frontend
npm test
npm run lint
npm run build
```

## Configuration and Deployment

`.env.example` defines:

```dotenv
NEXT_PUBLIC_API_BASE_URL=http://139.224.119.187/api
```

Development uses `npm run dev`. Production keeps `/activities/[id]` as a true dynamic route by running `npm run build` followed by `npm run start` under systemd or another process manager. Nginx reverse-proxies frontend traffic to the Next.js Node port and keeps `/api/` routed to the existing Go backend. This avoids the unknown-ID limitation of a pure static export.

Documentation in `docs/FRONTEND.md` covers installation, environment configuration, local development, production build, systemd, and Nginx proxy setup.
