---
name: cistern-architect
description: Design brief creation for Cistern architect cataractae. Produces a DESIGN_BRIEF.md that constrains implementation to fit the existing codebase's patterns, conventions, and idioms. Every item in the brief must be verifiable — citing specific files, lines, and patterns found in the codebase.
---

# Cistern Architect — Design Brief Protocol

## The One Principle

**Every constraint in the brief must be verifiable with a specific command, file path, or line number.**

A brief that says "follow existing patterns" gives the implementer no clear standard and the reviewer nothing concrete to verify. A brief that says "SQL identifiers must be backtick-quoted — see V135__add_organization_settings.kt" gives both a clear, testable standard.

If you cannot name the file and line that establishes a pattern, you have not investigated deeply enough.

## Investigation Steps

1. Read CONTEXT.md and all revision notes
2. Read the full requirements from the droplet description
3. For each area the requirements touch, investigate the codebase using the method below
4. Write DESIGN_BRIEF.md (see format below)
5. Commit (see cistern-git skill — exclude CONTEXT.md)
6. Signal outcome

## Investigation Method

Do not guess. Do not write "the codebase uses X" without naming the file that proves it.

### Pattern Evidence

For every pattern you prescribe, find at least one file that demonstrates it:

| What to find | How to find it | What to write in the brief |
|---|---|---|
| ORM/query patterns | Search for existing column types, query builders, DAO methods in the target package | "Use Exposed's Exists/NotExists DSL — see `OrganizationTable.kt:45` for an EXISTS subquery precedent" |
| Naming conventions | Check file names, class names, constant names, column names | "Constants live in `OrganizationPermissionNames` object — see `OrganizationPermissionNames.kt`" |
| Error handling | Search for error types, `requireNotNull`, custom exceptions | "Use `requireCatalogPermissionId` pattern — see `CatalogPermissionCache.kt:23`" |
| Collection types | Search for Set vs List vs Map in similar contexts | "`loadPermissionsForOrgs` returns `Map<Long, Map<String, Set<String>>>` — Set because the UNIQUE constraint on `organization_permission` means no duplicates" |
| Migration conventions | List existing migration files, read DDL/DML patterns | "Quote identifiers with backticks — see `V135__add_organization_settings.kt` lines 8-12" |

### Abstraction Boundary Analysis

For every new class, function, or utility the implementation will create:

**"Could another entity use the same pattern?"**

If yes → the implementation must accept its context as a constructor parameter, not hardcode a reference to a specific entity. Name the base class or interface it should extend, with file and line.

If no → state explicitly: "This is specific to Organization and will not be reused." The reviewer will then know not to flag it as over-coupled.

### Repeated Pattern Detection

Search the codebase for repeated inline expressions. When the same pattern appears 3+ times, it must be extracted. Name the exact locations:

"Extract `boolPerm(orgId: Long, perm: String): Boolean` from `OrganizationDAO.kt` lines 45, 52, 59, 66, 73, 80, 87, 94, 101, 108, 115, 122, 129"

Not: "extract common patterns" (useless).

## Brief Format

Write the design brief as `DESIGN_BRIEF.md` in the repository root:

```markdown
# Design Brief: <feature title>

## Requirements Summary
<One-paragraph summary of what needs to be built>

## Existing Patterns to Follow

### ORM / Query
<Specific patterns found — name the files and lines>

### Naming Conventions
<File names, class names, column names, constant names, migration numbering>

### Error Handling
<How the codebase handles errors — specific patterns, specific files>

### Collection Types
<Where Set vs List vs Map is used and why — specific files, specific constraints>

### Migrations
<Numbering, quoting, DDL/DML separation, description quality — with file evidence>

### Testing
<Test patterns, integration test locations, naming conventions — with file evidence>

## Reusability Requirements

<For each new class/utility: is it entity-specific or generic? If generic,
what parameter makes it reusable? If specific, state that explicitly.>

## DRY Requirements

<Any repeated pattern identified by 3+ occurrences. Name the helper, specify
its signature, and list the exact file:line locations where the pattern appears.>

## Migration Requirements

<Specific: file naming, identifier quoting (with dialect), DDL/DML separation,
description quality for reference data inserts.>

## Test Requirements

<Specific: which test files need new tests, what kind (unit vs integration),
exact naming convention for new test functions, and precise coverage gaps.>

## Forbidden Patterns

<Anti-patterns to exclude. Each entry must reference an existing example in the
codebase and explain why the new implementation must not repeat it.>

## API Surface Checklist

<Every item must be individually verifiable — a reviewer can check each with
grep or by reading a named file.>

- [ ] <specific, verifiable constraint>
- [ ] <specific, verifiable constraint>
- [ ] ...
```

## What the Brief Is NOT

- It is NOT a full implementation. Do not write production code.
- It is NOT a test file. Do not write test cases.
- It is NOT a review. Do not review code that does not exist yet.
- It IS a contract document. The implementer must satisfy every item. The
  reviewer must verify every item.

## Quality Bar

The brief is complete when:
1. Every pattern reference includes a specific file path (and line number where possible)
2. Every constraint in the API Surface Checklist is individually verifiable
3. There are no "TBD" or "determine during implementation" items
4. The DRY requirements name exact file:line locations, not vague "similar patterns"

A brief that fails any of these checks is incomplete. Signal recirculate with a
note explaining what you cannot determine from the codebase.