# Design Brief: Web UI — Integration, routing, and peek

## Requirements Summary

Wire the React SPA into the Go server (verify the existing integration is complete and add missing pieces), establish all routing with proper 404 and error boundaries, add the live session peek viewer with search and auto-scroll toggle, finalize navigation links between old and new UI, add polish components (ErrorBoundary, LoadingSkeleton, Toast system, network status indicator), add keyboard navigation (focus management, Escape to close, command palette stretch goal), make the layout responsive for mobile, and add missing test coverage. The Go server integration and SPA handler already exist — this brief covers completion, polish, and testing.

## Existing Patterns to Follow

### ORM / Query

Not applicable — no database changes in this feature.

### Naming Conventions

- **Go files:** `snake_case.go` matching the command or feature — `dashboard_web.go`, `dashboard_web_spa.go`, `dashboard_web_test.go`.
- **Go test functions:** `TestXxx_YyyScenario` — see `TestDashboardWebMux_RootServesHTML` at `dashboard_web_test.go:24`, `TestSPAHandler_ServesIndexHTML` at `dashboard_web_spa_test.go:9`, `TestSPAHandler_InjectsAuthMetaTag` at `dashboard_web_spa_test.go:86`.
- **Go test helpers:** `tempCfg(t)`, `tempDB(t)` at `dashboard_test.go:21-31`. Use `t.Helper()` — see `dashboard_test.go:22`.
- **React component files:** `PascalCase.tsx` matching the component name — see `PeekPanel.tsx`, `Header.tsx`, `Sidebar.tsx`, `ModalOverlay.tsx`.
- **React component names:** Named export `export function ComponentName` — see `DropletDetail.tsx:30`, `PeekPanel.tsx:10`, `Header.tsx:9`.
- **React hooks:** `useXxx` prefix — see `useDashboardEvents.ts:11`, `useAuth.ts:34`, `useApi.ts:19`.
- **Unexported Go structs:** Unexported fields only. See `aqueductSessionInfo` at `dashboard_web.go:113-118` (`sessionID`, `dropletID`, `title`, `elapsed` — all unexported). See `spaHandler` at `dashboard_web_spa.go:15-18` (`indexHTML`, `assetsFileServer` — unexported).
- **Do not shadow Go builtins:** `min`, `max`, `any`.

### Error Handling

- **Go:** `fmt.Errorf("pkg: context: %w", err)` for wrapping. `slog.Error`/`slog.Warn` for operational errors. See `dashboard_web.go:388` (Origin validation wraps errors), `dashboard_web.go:823` (log.Println for missing API key). For 500 responses, sanitize to "internal error" — see `writeAPIError` at `dashboard_web.go:1196-1203`.
- **React:** `apiFetch<T>` throws `Error('API error ${status}: ${body}')` — see `shared.ts:9`. Components catch and display as `text-sm text-cistern-red font-mono` — see `DropletDetail.tsx:96`.

### Collection Types

- **Go:** Slices throughout — `[]CataractaeInfo`, `[]FlowActivity`, `[]*cistern.Droplet`. Maps for lookups — `map[string]string` for `blockedByMap` at `dashboard.go:69`.
- **React:** Arrays — `Droplet[]`, `DropletIssue[]`, `string[]`. No `Set` or `Map` in frontend code — see `Dashboard.tsx:39` uses `new Set()` only for lookup purposes within a render.

### Migrations

Not applicable — no database schema changes.

### Idiom Fit

- **SPA handler:** The existing `spaHandler` at `dashboard_web_spa.go:15` uses constructor injection via `newSPAHandler(apiKey)` — no package-level mutable state. All new Go code must follow this pattern.
- **Embedded FS:** `embed.FS` pattern already established — `//go:embed assets/static` at `dashboard_web.go:42` and `//go:embed assets/web` at `dashboard_web_spa.go:11`. Build output goes to `cmd/ct/assets/web/`.
- **Vite config:** `base: '/app/'` at `vite.config.ts:6`, `outDir: '../cmd/ct/assets/web'` at `vite.config.ts:8`. No changes needed to these — the build pipeline is already correct.
- **Tests:** Use `httptest.NewServer` and `httptest.NewRecorder` — see `dashboard_web_spa_test.go:11`, `dashboard_web_test.go:27`. `t.Helper()` in test helpers — see `dashboard_test.go:22`.
- **React state:** All state lives in component `useState` or hooks. No shared mutable package-level state — see `DashboardContext.tsx:17` (provider wraps hook, no module-level state).
- **Constructor pattern:** `newSPAHandler(apiKey string) *spaHandler` at `dashboard_web_spa.go:20` — eager initialization, fully usable after construction. All new Go structs must follow this. No `SetXxx` mutation methods, no `initClient()` pattern.

### Testing

- **Go tests:** `dashboard_web_spa_test.go` uses `httptest.NewServer(handler)` then `http.Get(server.URL + "/app/")` — see line 11-14. Table-driven tests — see `dashboard_web_test.go:565-587` (cross-origin test cases). Test helpers `tempCfg`, `tempDB` at `dashboard_test.go:21-31`.
- **React tests:** Vitest with `@testing-library/react` and `jsdom` environment — `vitest.config.ts:10`. `describe('ComponentName', () => { it('description') })` — see `PeekPanel.test.ts:4`. Mock fetch pattern — see `useApi.test.ts`. `beforeEach/afterEach` with `localStorage.clear()` + `vi.restoreAllMocks()` — universal pattern.
- **React test location:** `web/src/__tests__/ComponentName.test.tsx` — see existing 39 test files in `web/src/__tests__/`.

## Reusability Requirements

### New components and their reusability

| Component | Reusable? | Reason |
|-----------|-----------|--------|
| `ErrorBoundary.tsx` | Yes — wraps the entire app | Catches render errors anywhere in the component tree |
| `LoadingSkeleton.tsx` | Yes — used by every page on initial load | Multiple card/row variants reuse the same animation |
| `Toast.tsx` | Yes — replaces inline `showToast` per page | Currently duplicated in `CastellariusPage.tsx:6-9` |
| `NetworkStatusBar.tsx` | Yes — shows in `Header` | Already partially exists as the green/red dot in `Header.tsx:33-37` |
| `TerminalView.tsx` | Specific to `PeekPanel` | Monospace renderer is peek-specific but could be reused if a log viewer needed it |
| `NotFoundPage.tsx` | Yes — React Router catch-all | Generic 404 for all unknown `/app/*` routes |
| `CommandPalette.tsx` | Yes — global keyboard shortcut | Stretch goal; would be invoked from `AppLayout` |

### Toast — must be extracted

The `showToast` pattern with timer-based auto-dismiss appears in `CastellariusPage.tsx:6-25` (inline `Toast` interface + `showToast` callback + `toastTimerRef`). The requirements specify toast notifications for all API errors, which means every page will need this. Extract to a shared component and context:

```tsx
// web/src/context/ToastContext.tsx
// Wraps the app, provides useToast() hook
```

See the inline pattern at `CastellariusPage.tsx:6-24` for the exact behavior (3-second auto-dismiss, replaces previous toast, timer cleanup on unmount).

### LoadingSkeleton — must be parameterized

The requirements specify "skeleton screens for initial page loads (matching the dark theme)". The existing loading pattern is `text-center py-4 text-cistern-muted font-mono text-sm` with "Loading..." — see `DropletDetail.tsx:104-107`, `IssuesList.tsx:29`. The skeleton must accept a `variant` prop (`card`, `row`, `table`) to match different page sections.

## Coupling Requirements

- No shared mutable package-level state in Go or React.
- `ToastContext` must be a React context wrapping `AppLayout`, not a module-level singleton. Toast state lives in `useState` inside the provider — see `DashboardContext.tsx:17-25` for the established context pattern.
- `ErrorBoundary` must be a class component (React requirement — `componentDidCatch`), wrapping `<RouterProvider>` in `main.tsx`. It does not depend on any context.
- `LoadingSkeleton` accepts `variant` and optional `count` props — no dependency on data shapes.
- `NetworkStatusBar` reads `connected` from `DashboardContext` — already available in `AppLayoutInner` at `App.tsx:26`.
- `CommandPalette` (stretch goal) accepts `onSelect` callback — no dependency on specific routes or actions.

## DRY Requirements

### Toast pattern (1 explicit occurrence, N pending)

The `showToast` + auto-dismiss timer pattern appears in:
1. `web/src/pages/CastellariusPage.tsx:6-25` (inline `Toast` interface, `showToast` callback, `toastTimerRef`)

With the requirement that "toast notifications for all API errors" appear everywhere, every page would duplicate this pattern without extraction. Extract:
- `web/src/context/ToastContext.tsx` — provider + `useToast()` hook
- `web/src/components/Toast.tsx` — render component (fixed at bottom-right, auto-dismiss)
- Replace the inline pattern in `CastellariusPage.tsx:6-25`

### Loading/empty state pattern (5+ occurrences)

The "Loading..." text pattern appears in:
1. `web/src/pages/DropletDetail.tsx:104-107`
2. `web/src/components/IssuesList.tsx:29`
3. `web/src/components/DependenciesList.tsx:43-44`
4. `web/src/pages/Dashboard.tsx:33-35`
5. `web/src/components/DropletTable.tsx` (spinner pattern)

Each currently renders a plain "Loading..." string. The `LoadingSkeleton` component replaces these with animated placeholder rectangles matching the dark theme. The component does NOT replace the `useApi` hook loading states — it's purely a presentational component that receives `loading: boolean` and renders skeleton or children.

### PeekPanel auto-scroll + search (new feature, no existing duplication)

The current `PeekPanel.tsx:59-63` auto-scrolls unconditionally via `scrollTop = scrollHeight`. The requirements add:
- Toggle auto-scroll (pin/unpin button)
- Search within peek output

These are additions to the existing component, not a new component, because the peek viewer is entity-specific to aqueduct monitoring.

### Header network status (partial duplication)

The `Header.tsx:33-37` already shows a green/red connection dot with "Live"/"Disconnected" text. The requirements add "Network status indicator in the header" with auto-reconnect display. Enhance the existing `Header` component rather than creating a separate `NetworkStatusBar`. The current `connected` prop from `DashboardContext` already drives this.

## Migration Requirements

Not applicable — no database changes.

## Test Requirements

### New test files

| Test file | What it covers |
|-----------|----------------|
| `web/src/__tests__/ErrorBoundary.test.tsx` | Catches render errors, displays fallback UI, does not crash app |
| `web/src/__tests__/LoadingSkeleton.test.tsx` | Renders card/row/table variants, applies pulse animation, hides when loading=false |
| `web/src/__tests__/Toast.test.tsx` | Renders toast message, auto-dismisses after 3s, replaces previous toast |
| `web/src/__tests__/NotFoundPage.test.tsx` | Renders 404 message, shows link back to dashboard |
| `web/src/__tests__/TerminalView.test.tsx` | Renders monospace text lines, handles empty input, respects max buffer size |
| `cmd/ct/dashboard_web_integration_test.go` | Verifies /app/ serves index.html, /app/assets/ serves assets, / redirects to xterm, /api/ endpoints still work, /app/unknown serves index.html for client routing |

### Updated test files

| Test file | What changes |
|-----------|-------------|
| `web/src/__tests__/PeekPanel.test.ts` | Add tests for auto-scroll toggle, search within output |
| `web/src/__tests__/DashboardContext.test.ts` | Add test for network status propagation to Header |
| `web/src/__tests__/useAuth.test.ts` | Add test for 401 redirect interceptor pattern |
| `cmd/ct/dashboard_web_spa_test.go` | Add test for CSP header on index.html (already partially tested), add test for /app/unknown falling back to index.html |

### Specific test cases

- `ErrorBoundary.test.tsx`: `it('catches render error and displays fallback')`, `it('does not crash sibling components')`, `it('provides retry button in fallback')`
- `LoadingSkeleton.test.tsx`: `it('renders card skeleton with pulse animation')`, `it('renders row skeleton')`, `it('renders table skeleton with multiple rows')`, `it('renders nothing when loading is false')`
- `Toast.test.tsx`: `it('displays toast message')`, `it('auto-dismisses after timeout')`, `it('replaces previous toast on new call')`, `it('cleans up timer on unmount')`
- `NotFoundPage.test.tsx`: `it('renders 404 heading')`, `it('renders link back to /app/')`
- `TerminalView.test.tsx`: `it('renders lines in monospace')`, `it('renders "Connecting" when no output')`, `it('handles empty lines array')`
- `PeekPanel.test.ts` (enhanced): `it('defaults to auto-scroll on')`, `it('toggles auto-scroll off when user scrolls up')`, `it('resumes auto-scroll when toggled back on')`, `it('renders search input when search is toggled')`, `it('highlights matching text in search results')`, `it('clears search on Escape')`
- Go integration tests:
  - `TestWebMux_AppServesIndexHTML` — verify /app/ returns 200 with text/html and CSP headers (partially covered by existing `TestSPAHandler_ServesIndexHTML`)
  - `TestWebMux_AppSubRouteServesIndexHTML` — verify /app/droplets, /app/castellarius return index.html (covered by `TestSPAHandler_ServesSubRoutes`)
  - `TestWebMux_AppAssetsServedWithNosniff` — verify /app/assets/*.js has X-Content-Type-Options: nosniff (covered by `TestSPAHandler_SecurityHeadersOnAssets`)
  - `TestWebMux_RootServesXtermDashboard` — verify / serves xterm.js HTML (covered by `TestDashboardWebMux_RootServesHTML`)
  - `TestWebMux_AppRedirectRoot` — verify /app redirects 301 to /app/ (covered by `TestSPAHandler_RedirectsAppRoot`)
  - `TestWebMux_APIEndpointsStillWork` — verify /api/droplets returns JSON (covered by existing API tests)
  - `TestWebMux_ClassicDashboardLink` — NEW: verify the root HTML contains a link to /app/
  - `TestWebMux_AppIndexHTMLContainsClassicLink` — NEW: verify /app/ HTML contains "Classic Dashboard" link to /

## Forbidden Patterns

- **Package-level mutable vars for config.** The existing `currentDropletSSEConnections` and `currentLogSSEConnections` at `dashboard_web.go:95-98` are `int64` atomics used as global counters — this is acceptable for atomic counters but FORBIDDEN for new config-like mutable state. All new config must be struct fields with constructor injection.
- **SetXxx mutation methods.** FORBIDDEN — use constructor injection. See `newSPAHandler(apiKey string)` at `dashboard_web_spa.go:20`.
- **Lazy initialization (initClient pattern).** FORBIDDEN — eager constructor initialization. `newSPAHandler` reads all files and builds the handler in one call.
- **Shared mutable package-level state (maps, vars).** FORBIDDEN — must be struct fields with defensive copies. The `outboundRateLimiter` at `dashboard_web.go:1266-1271` is a struct with constructor `newOutboundRateLimiter()` — correct pattern.
- **PascalCase fields on unexported Go structs.** FORBIDDEN — see `aqueductSessionInfo` at `dashboard_web.go:113-118` for the correct pattern (all unexported fields).
- **fmt.Fprintf(os.Stderr) for errors.** FORBIDDEN in new code — the existing `fmt.Fprintf(os.Stderr, "ct dashboard: spawn error: %v\n", err)` at `dashboard_web.go:517` is legacy; new code must use `slog.Error`.
- **Silently swallowing errors.** FORBIDDEN — at minimum `slog.Debug`.
- **Shadowing Go builtins** (`min`, `max`, `any`).** FORBIDDEN.
- **Inline modal JSX in parent components.** FORBIDDEN — see previous brief. All modals must be separate component files.
- **`any` type usage in TypeScript.** FORBIDDEN — all API responses and form state must use typed interfaces from `api/types.ts`.
- **Third-party component libraries (xterm.js, ansi-to-html, etc.).** The peek viewer must NOT import xterm.js — it already exists for the classic dashboard and is 300KB+. The `TerminalView.tsx` must be a simple `<pre>`-based renderer. The server already strips ANSI on the peek endpoint (`dashboard_web.go:898,957`). See existing `PeekPanel.tsx:94-99` which already uses `<pre>` + `font-mono text-xs text-cistern-green`.

## API Surface Checklist

### Go server — existing endpoints (verify no regressions)

- [ ] **`GET /app/` serves index.html** — Contract: returns 200 with `Content-Type: text/html; charset=utf-8`, CSP header `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss:; img-src 'self'; font-src 'self'`, X-Content-Type-Options: nosniff, X-Frame-Options: DENY, Referrer-Policy: strict-origin-when-cross-origin. Already implemented at `dashboard_web_spa.go:68-74`. Already tested at `dashboard_web_spa_test.go:9-28`.

- [ ] **`GET /app` redirects to /app/** — Contract: returns 301 Moved Permanently with Location: /app/. Already implemented at `dashboard_web.go:848-850`. Already tested at `dashboard_web_spa_test.go:53-84`.

- [ ] **`GET /app/assets/*` serves static assets** — Contract: serves embedded JS/CSS with X-Content-Type-Options: nosniff, correct MIME types via `http.FileServer`. Already implemented at `dashboard_web_spa.go:61-64`. Already tested at `dashboard_web_spa_test.go:155-169`.

- [ ] **`GET /app/*` (non-asset) serves index.html** — Contract: all non-asset routes under /app/ return index.html for client-side routing. Already implemented at `dashboard_web_spa.go:67-74`. Already tested at `dashboard_web_spa_test.go:30-51`.

- [ ] **`GET /` serves xterm.js classic dashboard** — Contract: returns 200 with text/html containing xterm.js references, backward compatible. Already implemented at `dashboard_web.go:852-859`. Already tested at `dashboard_web_test.go:24-45`.

### Go server — new additions

- [ ] **Classic Dashboard link in /app/ index.html** — Contract: the SPA's `Header` component renders a "Classic Dashboard" link pointing to `/`. The link is visible in the header bar next to the connection status. Implementation goes in the React `Header.tsx` component, NOT in Go. The Go server does not inject links into the SPA HTML — it only injects the auth meta tag.

- [ ] **New UI link in / xterm.js dashboard** — Contract: the root HTML (the xterm.js page at `/`) contains a link or badge "New UI →" that navigates to `/app/`. This is implemented in Go by adding an HTML element to `dashboardHTML` (the template string). The link must be unobtrusive — a small fixed-position badge in the top-right corner, not in the terminal viewport.

- [ ] **`/app/unknown-route` still serves index.html** — Contract: client-side routing means any `/app/*` route that isn't an asset request serves index.html. The React Router handles 404 display. Already implemented at `dashboard_web_spa.go:67-74`. Add a React-side 404 catch-all route.

### React — routing additions

- [ ] **404 catch-all route in `main.tsx`** — Contract: a `path: '*'` route as the last child of the `/app` layout renders `<NotFoundPage />`. The `NotFoundPage` displays "404 — Page Not Found" heading in `font-mono text-cistern-muted`, and a link back to `/app/` styled as `text-cistern-accent hover:underline`. Route must come after all other children in the `createBrowserRouter` array — see `main.tsx:21-32` for the existing route structure.

- [ ] **React Router `ErrorBoundary` class component** — Contract: wraps `<RouterProvider>` in `main.tsx:36-39`. Catches render errors from any route component. Renders a fallback UI: `flex items-center justify-center h-screen bg-cistern-bg` with `text-cistern-red font-mono text-lg` heading "Something went wrong" and a "Reload" button that calls `window.location.reload()`. Must be a class component (React's `componentDidCatch` API requirement). Does NOT replace per-page error states (like `DropletDetail.tsx:92-99`) — those handle data-loading errors, not render crashes.

### React — PeekPanel enhancements

- [ ] **`PeekPanel` auto-scroll toggle** — Contract: adds a `autoScroll` boolean state defaulting to `true`. When `autoScroll` is true, the existing `useEffect` at `PeekPanel.tsx:59-63` scrolls to bottom on every output change. When `autoScroll` is false, the panel does NOT auto-scroll. A toggle button in the header bar (next to "Peek" label) shows "↓ Auto" when on and "↓ Manual" when off. The button uses `text-xs px-2 py-0.5 rounded border border-cistern-border text-cistern-muted hover:text-cistern-fg` — matching the existing button style at `PeekPanel.tsx:84`.

- [ ] **`PeekPanel` auto-scroll pause on manual scroll** — Contract: detects when the user manually scrolls the `<pre>` element. If the user scrolls more than 30px from the bottom, `autoScroll` is set to `false`. If the user scrolls to the bottom (within 30px), `autoScroll` is set to `true`. Uses `onScroll` handler on the `<pre>` element at `PeekPanel.tsx:94`.

- [ ] **`PeekPanel` search within output** — Contract: adds a search toggle button (magnifying glass icon) in the header bar. When active, shows a search input field below the header. Input uses existing `bg-cistern-bg border border-cistern-border rounded px-2 py-1.5 text-sm text-cistern-fg` style (see `AddNoteModal.tsx:49-50`). As the user types, matching lines are highlighted with `bg-cistern-yellow/30`. Pressing Enter jumps to the next match. Pressing Escape clears the search and hides the input. The search operates on the existing `output` state string (not a separate data structure) — filter lines containing the query text and render matches. When no search is active, all lines render as before.

### React — polish components

- [ ] **`ErrorBoundary` component** — Contract: class component `ErrorBoundary` in `web/src/components/ErrorBoundary.tsx`. Implements `componentDidCatch(error, errorInfo)`. State: `{ hasError: boolean, error: Error | null }`. When `hasError` is true, renders fallback UI instead of children. Fallback: centered card `bg-cistern-surface border border-cistern-border rounded-lg p-6 max-w-md w-full mx-4` containing error message and "Reload" button. When `hasError` is false, renders `children` unchanged. Props: `{ children: React.ReactNode, fallback?: React.ReactNode }`. If `fallback` is provided, render it instead of the default error UI.

- [ ] **`LoadingSkeleton` component** — Contract: functional component in `web/src/components/LoadingSkeleton.tsx`. Props: `{ variant: 'card' | 'row' | 'table', count?: number, loading: boolean, children: React.ReactNode }`. When `loading` is true, renders animated skeleton placeholders. When `loading` is false, renders `children`. Animation: Tailwind `animate-pulse` on each skeleton element. Card variant: `bg-cistern-surface border border-cistern-border rounded-lg p-4` with 3 horizontal bars inside (heading: `h-4 w-24 bg-cistern-border rounded`, body: `h-3 w-full bg-cistern-border rounded`, body2: `h-3 w-3/4 bg-cistern-border rounded`). Row variant: single row `h-6 w-full bg-cistern-border rounded`. Table variant: `count` rows of `h-8 w-full bg-cistern-border rounded` with header row. Must be added to barrel `web/src/components/index.ts`.

- [ ] **`Toast` component and `ToastContext`** — Contract: `web/src/components/Toast.tsx` renders a single toast notification at `fixed bottom-4 right-4 z-50`. Props: `{ message: string, type: 'success' | 'error', visible: boolean }`. Success: `bg-cistern-green/20 border border-cistern-green/40 text-cistern-green`. Error: `bg-cistern-red/20 border border-cistern-red/40 text-cistern-red`. Uses `font-mono text-sm`. Auto-dismisses after 3000ms (matches existing pattern at `CastellariusPage.tsx:21-24`). `web/src/context/ToastContext.tsx` provides `useToast()` hook returning `{ showToast: (message: string, type: 'success' | 'error') => void }`. Must wrap `AppLayout` children in `App.tsx:16-20`. Replace inline pattern in `CastellariusPage.tsx:6-25`. Must be added to barrel `web/src/components/index.ts`.

- [ ] **Network status indicator in `Header`** — Contract: enhances the existing connection dot at `Header.tsx:33-37`. When `connected` is false, shows "Reconnecting..." text in addition to the red dot. When `connected` transitions from false to true, briefly shows "Connected" in `text-cistern-green` for 2 seconds before reverting to "Live". No new component — modify `Header.tsx` directly. Uses `useEffect` with a timer to handle the "Connected" flash.

- [ ] **API 401 interceptor** — Contract: in `web/src/api/shared.ts`, after `apiFetch` receives a 401 response, check if the app requires auth (via `isAuthRequired()` at `useAuth.ts:5-8`). If auth is required and the key is present, the 401 means the key is invalid — clear the stored key and update auth state. If auth is required and there's no key, this should not happen (SPA routes are auth-exempt). Implementation: in `apiFetch` at `shared.ts:3-13`, add a check after `!resp.ok`: if `resp.status === 401` and auth is required, call `clearStoredKey()` from `useAuth.ts:26-32` and dispatch a custom event `'cistern:auth-expired'`. The `AppLayout` component listens for this event and triggers `logout()` from `useAuth`.

### React — mobile responsive

- [ ] **Sidebar collapses on mobile** — Already implemented at `Sidebar.tsx:20-21` (overlay + `md:hidden`). Verify the hamburger menu button at `Header.tsx:13-19` works correctly on small screens. No changes needed — this pattern is already in place.

- [ ] **Table-to-card switch on mobile** — Contract: `DropletTable.tsx` (if it exists) or the droplets list page must use responsive breakpoints. On screens `< md` (640px), render droplets as vertical cards instead of table rows. Card pattern: `bg-cistern-surface border border-cistern-border rounded-lg p-3 space-y-1`. Table pattern (existing): standard `<table>`. Implementation: use Tailwind `hidden md:table` / `md:hidden` pattern for switching between card and table views.

- [ ] **Touch-friendly tap targets** — Contract: all interactive elements (buttons, links, inputs) must have minimum 44px tap target area. Use `min-h-[44px] min-w-[44px]` on icon-only buttons. Existing button styles at `PeekPanel.tsx:74` and `Header.tsx:14` may fail this bar on mobile — add `min-h-[44px]` or `p-2` to ensure adequate tap area.

### React — keyboard navigation

- [ ] **Focus management for modals** — Contract: when a `ModalOverlay` opens (see `ModalOverlay.tsx`), focus moves to the first focusable element inside the modal. When the modal closes, focus returns to the element that triggered it. Implementation: in `ModalOverlay.tsx`, add `useEffect` that focuses the first `[tabindex], button, input, select, textarea, a[href]` inside the modal when `open` becomes true. Use `autoFocus` or `ref.current.focus()`. Store the previously focused element via `document.activeElement` and restore it in the cleanup. Existing modal components (`AddNoteModal`, `EditMetadataModal`, etc.) inherit this behavior automatically since they use `ModalOverlay`.

- [ ] **Escape closes modals** — Contract: pressing Escape when a modal is open calls `onClose`. Implementation: in `ModalOverlay.tsx`, add a `keydown` event listener on `document` that calls `onClose` when `e.key === 'Escape'`. Use `useEffect` with `open` dependency — add listener when modal opens, remove when it closes. Must check that the event target is not an `<input>` or `<textarea>` where Escape might have a different meaning (e.g., clearing search in PeekPanel). For modals, Escape always closes.

- [ ] **Tab navigation through sidebar** — Contract: sidebar links are focusable and navigable via Tab. Already implemented — `NavLink` at `Sidebar.tsx:49` renders `<a>` tags which are natively focusable. Add `focus:ring-2 focus:ring-cistern-accent focus:outline-none` to the NavLink className to show focus indicator.

- [ ] **Command palette (stretch goal — nice-to-have)** — Contract: `Ctrl+K` opens a command palette overlay. The palette shows a search input and a list of navigable actions (Go to Dashboard, Go to Droplets, etc.). Uses `useNavigate` from React Router. Component: `web/src/components/CommandPalette.tsx`. Accepts `{ open: boolean, onClose: () => void }`. Uses `ModalOverlay` for the backdrop. If not implemented, the `Ctrl+K` key binding should not be registered (no empty stub).

### React — loading states and optimistic updates

- [ ] **`LoadingSkeleton` replaces "Loading..." text** — Contract: every page component that currently renders a "Loading..." message (see DRY Requirements) must be updated to render `<LoadingSkeleton variant="..." loading={loading}>...</LoadingSkeleton>` instead. The `children` prop receives the loaded content.

- [ ] **Spinner for async actions** — Contract: buttons that trigger API mutations (Pass, Recirculate, Pool, Create, Signal, Restart) must show a spinner during the request. Use the existing `submitting` state pattern — see `EditMetadataModal.tsx` (disables button during submit). Add `disabled:opacity-50` and a small inline spinner `<span className="animate-spin">⟳</span>` replacing the button text. Matches existing `disabled:opacity-50` pattern at `AddNoteModal.tsx:55`.

- [ ] **Optimistic UI updates for status changes** — Contract: when a user signals a droplet (pass, recirculate, pool, close, reopen), update the droplet status in the UI immediately before the API response arrives. If the API call fails, revert to the original status and show an error toast. Implementation: in `DropletDetail.tsx`, wrap `handleAction` to set `sseDroplet` with the expected status optimistically, then revert on error. Uses the existing `useDropletMutation.mutate()` at `useApi.ts:238-251`.

### React — cross-UI links

- [ ] **Header "Classic Dashboard" link** — Contract: the React `Header.tsx` component renders a small link "Classic Dashboard" pointing to `/`. Position: after the connection indicator at `Header.tsx:33-37`. Style: `text-xs text-cistern-muted hover:text-cistern-fg font-mono border border-cistern-border rounded px-2 py-0.5`. This link allows users on the SPA to reach the xterm.js TUI dashboard.

- [ ] **Root HTML "New UI" link** — Contract: the Go `dashboardHTML` string (serving the xterm.js classic dashboard at `/`) contains a small fixed-position badge in the top-right corner with text "New UI →" linking to `/app/`. Style: `position:fixed;top:8px;right:8px;background:rgba(0,0,0,0.7);color:#60a5fa;padding:4px 8px;border-radius:4px;font-size:11px;font-family:monospace;z-index:100;text-decoration:none;`. Must be added AFTER the opening `<body>` tag in `dashboardHTML` so it does not interfere with xterm.js terminal layout. The link opens `/app/` in the same tab (no `target="_blank"`).

### Build integration

- [ ] **`make web-build` Makefile target** — Contract: a Makefile target `web-build` runs `cd web && npm install && npm run build`. The Vite config (`vite.config.ts:8`) already outputs to `../cmd/ct/assets/web/`. The Go `//go:embed assets/web` directive (`dashboard_web_spa.go:11`) automatically picks up the built assets. No other build changes needed. The existing `go:build` pipeline does not need modification — embedding happens at compile time.

- [ ] **Build process documentation** — Contract: add a "Web UI Development" section to `CONTRIBUTING.md` explaining: 1) `make web-build` to build the SPA, 2) `cd web && npm run dev` for development with Vite proxy, 3) the SPA is always available at `/app/` when running `ct dashboard --web`, 4) the classic xterm.js dashboard remains at `/`. Keep the section under 15 lines.

### Go server — dashboard command flag

- [ ] **`ct dashboard --web` works as before** — Contract: the `--web` flag starts the HTTP server. Already implemented at `dashboard.go:814-815`. No changes needed.

- [ ] **`ct dashboard --web --addr 0.0.0.0:5737`** — Contract: the `--addr` flag sets the listen address. Already implemented at `dashboard.go:824`. No changes needed.

- [ ] **SPA available at /app/ in --web mode** — Contract: already true. The `newDashboardMuxInternalWith` function at `dashboard_web.go:846-847` mounts the SPA handler at `/app/`. No changes needed.

- [ ] **`ct dashboard --web --new-ui` flag** — Contract: OPTIONAL. If implemented, this flag makes `/` redirect to `/app/` instead of serving the xterm.js dashboard. When NOT set, `/` continues to serve xterm.js (backward compatible). If NOT implemented, add a comment in `dashboard.go` noting this as a future option.