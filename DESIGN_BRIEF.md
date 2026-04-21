# Design Brief: Fix Web SPA Crash on Empty Data — Null Array Regression

## Requirements Summary

Fix the complete SPA crash when the Go API returns JSON `null` for empty collection fields. The root cause is threefold: (1) Go nil slices serialize to JSON `null`, (2) React components call `.length` on potentially null arrays without null-coalescing, and (3) no integration test exercises the empty/null path. The fix must ensure the Go API always returns `[]` for array fields, the frontend null-coalesces as defense-in-depth, and both layers have regression tests.

## Existing Patterns to Follow

### ORM / Query

The codebase uses raw `database/sql` with `?` placeholders for SQLite — no ORM. Queries use `c.db.Query` with `rows.Scan`. See `internal/cistern/client.go:982-1015` for the `List` method pattern. The `fetchDashboardData` function already initializes empty slices in its constructor (`cmd/ct/dashboard.go:81-90`) — this is the correct pattern that must be extended to all collection fields unconditionally.

### Naming Conventions

Dashboard data struct fields use PascalCase for exported struct fields with `json` tags using snake_case — see `DashboardData` at `cmd/ct/dashboard.go:59-73`:

```go
Cataractae      []CataractaeInfo   `json:"cataractae"`
UnassignedItems []*cistern.Droplet `json:"unassigned_items"`
```

Unexported structs use unexported fields. All structs here are exported because they cross package boundaries (serialized to JSON in the HTTP handler at `cmd/ct/dashboard_web.go:876`).

### Error Handling

Errors are wrapped with `fmt.Errorf("context: %w", err)` — see `cmd/ct/dashboard.go:94`. The SSE handler at `cmd/ct/dashboard_web.go:874` silently drops fetch errors (`data, _ := fetcher(...)`) which is acceptable since SSE is a best-effort stream. No change needed to error handling patterns.

### Collection Types

The codebase uses slices (`[]T`) for all collection fields, not maps, because JSON serialization of arrays is ordered and positional. See `DashboardData` at `cmd/ct/dashboard.go:59-73` — seven slice/map fields all use `[]T` or `map[string]string`. The existing pattern for empty-slice initialization is at `cmd/ct/dashboard.go:81-90`:

```go
data := &DashboardData{
    Cataractae:      []CataractaeInfo{},
    UnassignedItems: []*cistern.Droplet{},
    CisternItems:    []*cistern.Droplet{},
    PooledItems:     []*cistern.Droplet{},
    BlockedByMap:    map[string]string{},
    FlowActivities:  []FlowActivity{},
}
```

This already covers UnassignedItems, CisternItems, PooledItems, BlockedByMap, FlowActivities, and Cataractae. The bug is that `RecentItems` is **not** initialized in this block — so when the `fetchDashboardData` code path at `cmd/ct/dashboard.go:240` assigns `data.RecentItems = recent` and `recent` is nil (no delivered/pooled droplets), it becomes nil and serializes to JSON `null`.

Additionally, the `FlowActivity.RecentNotes` field at `cmd/ct/dashboard.go:55` has a partial nil guard at line 249 (`if err != nil || notes == nil { notes = []cistern.CataractaeNote{} }`), but the struct field itself is never pre-initialized, so when `FlowActivities` entries are constructed at line 256, `RecentNotes` depends on this runtime guard.

### Idiom Fit

The standard library `encoding/json` marshals nil slices as `null` and empty slices as `[]` — this is documented Go behavior. The fix is to ensure all slice fields are initialized to empty slices (not nil). No custom JSON marshaler is needed. The existing `fetchDashboardData` constructor already demonstrates the correct pattern at `cmd/ct/dashboard.go:81-90`.

### Testing

Go tests use table-driven patterns with `t.Run` subtests. See `cmd/ct/dashboard_test.go:119-210` for `TestFetchDashboardData_FeedsDataCorrectly` and `cmd/ct/dashboard_web_test.go:123-144` for `TestDashboardWebMux_APIReturnsJSON`. React tests use Vitest with `@testing-library/react` — see `web/src/__tests__/DashboardContext.test.ts` and `web/src/__tests__/useDashboardEvents.test.ts`.

The existing Go API test at `cmd/ct/dashboard_web_test.go:123-144` (`TestDashboardWebMux_APIReturnsJSON`) tests with an empty database but does not verify that array fields are `[]` not `null`. This is the exact coverage gap that allowed the bug to ship.

## Reusability Requirements

- The Go-side fix (empty-slice initialization) is specific to `DashboardData` — no other struct in the codebase serializes dashboard data. Not reusable.
- The React-side null-coalescing pattern (`items ?? []`) is a local defensive measure in `Dashboard.tsx` — it should remain inline, not extracted into a utility, since each component section directly consumes a specific prop.
- Any integration test helper (e.g., `tempDB` and `tempCfg` at `cmd/ct/dashboard_test.go:21-115`) is already reused across test files and should continue to be reused.

## Coupling Requirements

No shared mutable package-level state is involved. `DashboardData` is constructed fresh per request in `fetchDashboardData`. The fix touches only struct initialization and JSX null-coalescing — no new shared state, no new packages, no package-level vars.

## DRY Requirements

### Repeated `items.length === 0` pattern in Dashboard.tsx

The pattern `if (items.length === 0) return null;` appears 4 times in Dashboard.tsx:

- `cmd/ct/dashboard.go` line 168 (QueueSection)
- `cmd/ct/dashboard.go` line 188 (PooledSection)
- `cmd/ct/dashboard.go` line 208 (UnassignedSection)
- `cmd/ct/dashboard.go` line 228 (RecentSection)

These are file-local component functions that each receive a typed `Droplet[]` prop. Extracting a shared wrapper would not reduce meaningful complexity — each section has different styling and header content. **Do not extract a helper.** The null-coalescing fix (`items ?? []`) at the call site in `SummarySection` is the correct DRY approach: apply the coalesce once at the point where `data.*_items` is passed as a prop, not inside each subcomponent.

### Go-side nil-slice pattern

The pattern of initializing empty slices in the `DashboardData` constructor already exists at `cmd/ct/dashboard.go:81-90`. The fix is to add `RecentItems: []*cistern.Droplet{}` to this constructor. No new helper needed — the constructor pattern is the standard.

## Migration Requirements

Not applicable — no database schema changes. This is a serialization and frontend fix only.

## Test Requirements

### Go Tests

**Must use existing test patterns** (table-driven with `t.Run`, httptest.NewRecorder, tempDB/tempCfg helpers at `cmd/ct/dashboard_test.go:21-115`).

New Go test functions required:

1. **`TestAPI_Dashboard_EmptyDB_ReturnsEmptyArraysNotNull`** in `cmd/ct/dashboard_web_test.go`
   - Creates a mux with `tempCfg(t)` and `tempDB(t)` (empty DB)
   - GET /api/dashboard
   - Decodes response into `map[string]interface{}` (not `DashboardData` struct, to verify raw JSON)
   - Asserts every array field (`cataractae`, `unassigned_items`, `cistern_items`, `pooled_items`, `recent_items`, `flow_activities`) is `[]` not `null`
   - Pattern: follows `TestDashboardWebMux_APIReturnsJSON` at `cmd/ct/dashboard_web_test.go:123-144`

2. **`TestFetchDashboardData_EmptyDB_AllSliceFieldsNonNil`** in `cmd/ct/dashboard_test.go`
   - Calls `fetchDashboardData(cfgPath, dbPath)` with empty DB
   - Asserts `data.RecentItems != nil`, `data.UnassignedItems != nil`, `data.CisternItems != nil`, `data.PooledItems != nil`, `data.Cataractae != nil`, `data.FlowActivities != nil`
   - Pattern: follows `TestFetchDashboardData_PooledItems_EmptyWhenNonePooled` at `cmd/ct/dashboard_test.go:626-644`

3. **`TestFetchDashboardData_FlowActivity_RecentNotesNonNil`** in `cmd/ct/dashboard_test.go`
   - Seeds an in-progress droplet assigned to an aqueduct with NO notes
   - Calls `fetchDashboardData`
   - Asserts `data.FlowActivities[0].RecentNotes != nil`
   - Pattern: follows `TestDashboardWebMux_NoteFieldsRoundTrip` at `cmd/ct/dashboard_web_test.go:219-257`

### React Tests

**Must use existing Vitest + @testing-library/react pattern** — see `web/src/__tests__/DashboardContext.test.ts` and `web/src/__tests__/useDashboardEvents.test.ts`.

New React test required:

1. **`Dashboard null fields render without crashing`** in `web/src/__tests__/Dashboard.test.tsx`
   - Renders `<Dashboard>` with `DashboardProvider` providing data where all array fields are `null` (simulating the broken API)
   - Asserts no JavaScript errors are thrown
   - Asserts the component renders (not blank/crashed)
   - Pattern: follows `DashboardContext.test.ts` mock data structure at lines 6-20, but with `null` for array fields instead of `[]`

### Integration Test

Per the acceptance criteria, an integration test that starts the dashboard server and hits `/api/dashboard` with an empty database is already covered by `TestAPI_Dashboard_EmptyDB_ReturnsEmptyArraysNotNull` above, which uses `httptest.NewServer` / `newDashboardMux` to test the full HTTP stack against an empty DB.

A headless-browser test that loads `/app/` and verifies no JS errors is beyond the scope of a unit/integration fix and would require Playwright or similar. The React unit test above provides equivalent coverage for the null-coalescing defense-in-depth layer.

## Forbidden Patterns

- **Do not add a custom `MarshalJSON` method to `DashboardData`** — the fix is struct initialization, not custom serialization. A `MarshalJSON` override would hide future nil-slice bugs instead of preventing them at the source.
- **Do not use `omitempty` on array JSON tags** — `omitempty` would omit the field entirely on empty, which breaks the API contract (callers expect the field to exist).
- **Do not extract a shared React wrapper component for the 4 similar section components** — each has different styling and semantics; the DRY fix is null-coalescing at the call site, not component extraction.
- **Do not use `SetXxx` mutation methods or package-level mutable state** — constructor initialization only.
- **Do not introduce a new package or utility** — the fix is two lines in Go (constructor) and six null-coalescences in React (call sites).

## API Surface Checklist

- [ ] `fetchDashboardData` constructor initializes `RecentItems: []*cistern.Droplet{}` — contract: every `DashboardData` returned by `fetchDashboardData` has non-nil slices for all seven collection fields, regardless of DB content. Verified by `TestFetchDashboardData_EmptyDB_AllSliceFieldsNonNil`.
- [ ] `fetchDashboardData` FlowActivity constructor at `cmd/ct/dashboard.go:256` ensures `RecentNotes` is never nil — the existing guard at line 249 already does this, but the contract must be: when `FlowActivities` is non-empty, each entry's `RecentNotes` is `[]cistern.CataractaeNote{}` not nil. Verified by `TestFetchDashboardData_FlowActivity_RecentNotesNonNil`.
- [ ] `SummarySection` in `Dashboard.tsx` passes null-coalesced arrays to child components — contract: `data.pooled_items ?? []`, `data.cistern_items ?? []`, `data.unassigned_items ?? []`, `data.recent_items ?? []` are passed to `PooledSection`, `QueueSection`, `UnassignedSection`, `RecentSection`. This is defense-in-depth; the Go fix ensures these are already `[]`, but the coalescing guarantees no crash even if a future code path or proxy introduces null.
- [ ] `CisternCountCard` in `Dashboard.tsx:146` uses `(data.pooled_items ?? []).length` instead of `data.pooled_items.length` — contract: this expression never throws TypeError, even if `data.pooled_items` is null/undefined.
- [ ] `AqueductSection` in `Dashboard.tsx:14` iterates `data.flow_activities` with null-coalesce — contract: `data.flow_activities ?? []` used in the `activityMap` construction, or the `.filter` calls on `cataractae` at lines 76-77 are safe because `data.cataractae` is already initialized by the Go constructor. But `data.flow_activities` at line 14 should be null-coalesced: `for (const act of data.flow_activities ?? [])`.
- [ ] `/api/dashboard` never returns JSON `null` for any array field — contract: for an empty database, the response JSON contains `[]` for `cataractae`, `unassigned_items`, `cistern_items`, `pooled_items`, `recent_items`, `flow_activities`. Verified by `TestAPI_Dashboard_EmptyDB_ReturnsEmptyArraysNotNull`.