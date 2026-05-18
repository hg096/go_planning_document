# AGD Package (Isolated Mode)

For Korean, see `00_agd/README.md`.
For installation and setup steps, see root [README.md](../README.md).

This package is designed to isolate AI document operations under `00_agd/` while minimizing conflicts with existing repositories.

## 0) AI Operation Cycle (Core)

Repeat steps 1-7 below to run AI workflows consistently without document gaps.

1. Normalize request
- Fix the 5 AI input fields first.
- `target_docs`, `target_sections`, `goal`, `constraints`, `done`

2. Check baseline
- Read core logic docs under `00_agd/agd_docs/10_core_logic/*` first.
- Check whether multiple source docs exist for the same topic.

3. Edit by section scope
- Edit only the specified `target_sections`.
- Create new files only when explicitly requested.

4. Record change trace
- Add `@change(reason/impact)` for every change.
- Keep tag metadata/blocks for maintenance and incident docs.

5. Sync consistency
- Verify `@map` and `@section` IDs are aligned.
- Run `map-sync` when needed.

6. Run validation
- default: `code-plan-check --mode auto`
- required: `check-all 00_agd/agd_docs --strict`

7. Approve and merge
- review local hook warnings before commit
- final enforcement should rely on CI (`AGD Guard`) required check

Failure branch rules:
- If `check-all` fails: fix document consistency and re-validate.
- If `code-plan-check` fails: reinforce `doc_base_path`/`section.path` or tags/`@change`, then re-validate.

Detailed guide:
- `00_agd/docs/AGD_TEMPLATE_GUIDE_en.md`

## 1) AI Request Template (Copy/Paste)

```txt
target_docs: 00_agd/agd_docs/...
target_sections: CORE-020, DEL-010
goal: state the change objective in 1-2 sentences
constraints: forbidden scope or limits
done: completion criteria
```

## 2) Daily Operation Routine (Recommended)

Before starting work:

```cmd
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

After edits:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

Check relations:

```cmd
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
```

## 3) Included Components

- executables: `00_agd/agd.exe`, `00_agd/agd_en.exe`, `00_agd/agd_ko.exe`
- docs: `00_agd/docs/*`
- examples: `00_agd/examples/*`
- working docs root: `00_agd/agd_docs/*`
- local hook: `00_agd/.githooks/pre-commit`
- hook logic: `00_agd/scripts/git-hooks/pre-commit.ps1`
- optional templates (CI/PR): `00_agd/templates/*`

No `.go` source files are included in the `00_agd/` folder.

## 4) Core Logic Change Notice (Soft Guard)

Core path policy file:
- `00_agd/policy/core_logic_paths.txt`

Rules:
- repository-root relative paths
- `#` comments allowed
- empty lines ignored
- both `/` and `\` are accepted (normalized by the hook)

Commit hook behavior:
- If staged files match policy paths, a warning is printed.
- In interactive terminals, it asks `Continue commit? (y/N)`.
- In non-interactive environments (CI), it prints warning only and continues.

## 5) Code Plan Validation (Code Plan Check)

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe code-plan-check --mode git
00_agd\agd_en.exe code-plan-check --mode cache
00_agd\agd_en.exe code-plan-check --mode git --git-source staged --strict-relation
```

Behavior summary:
- Collect changed code files (`auto`: git first, cache fallback on failure).
- Check overlap between code paths and AGD `doc_base_path`/`section.path`.
- A matched doc passes when either maintenance/incident tags exist or `@change` is updated.
- If relation policy (`A#SEC <-> B#SEC`) exists, linked service/frontend docs are additionally checked for `@change` updates in both directions.

Policy/cache paths:
- `00_agd/policy/code_plan_scope.txt`
- `00_agd/policy/code_plan_relations.txt`
- `00_agd/.cache/code_plan_validation.json`

Relation policy syntax:
- `<left-doc.agd>#<SECTION-ID> <-> <right-doc.agd>#<SECTION-ID>`
- write one relation per line
- for multiple relations, add multiple lines
- for 1:N, repeat the left endpoint; for N:1, repeat the right endpoint
- comma-separated multi-endpoint syntax in one line is not supported
- example: `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010`
- example (1:N):
  - `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010`
  - `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-020`
- default is WARN; `--strict-relation` upgrades it to FAIL

Strict options:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto --strict-mapping
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

Hook enforcement behavior:
- pre-commit runs `code-plan-check --mode git --git-source staged --strict-relation` first to block missing linked-doc updates.

## 6) agd_docs README Merge (No-Omission Integration)

The following section integrates the former `00_agd/agd_docs/README.md` operational content without omission.

### 6-1) Folder Structure

- `10_core_logic/*`: source-of-truth core logic documents
- `20_new_project/*`: new project planning documents

### 6-2) Recommended Placement by Topic

- `10_core_logic/service`: service/backend logic, domain rules, failure branches
- `10_core_logic/policy`: rules/standards/approval/release policy
- `10_core_logic/architecture`: ADR/trade-off records

- `20_new_project/<project-key>`: project scope, policy, delivery plan, and roadmap in one folder

### 6-3) Guide-First Operation

1) Read `00_agd/docs/AGD_TEMPLATE_GUIDE_ko.md` or `AGD_TEMPLATE_GUIDE_en.md` first.
2) For core logic, keep `10_core_logic/service` as the single authoritative flow.
3) Link new project docs to core logic with `source_doc`/`source_sections` when needed.
4) When core logic changes, sync linked new project docs in the same change set.
5) Always record `@change(reason/impact)`.
6) Do not report completion before passing 3 validations + blind reconstruction test.

### 6-4) Minimal Command Set

```cmd
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

### 6-5) Blind Reconstruction Test

- input: only `10_core_logic/service` document(s) (no code reference)
- output: domain pseudo-implementation (call order, DTO, failure branches, state transitions)
- pass criteria: 0 missing required blocks + 0 mapping mismatch

### 6-6) Quality Rubric (9/10)

- rubric score must be 9 or higher
- structural consistency point is mandatory
- deduction criteria:
  - fewer than 10 domain pseudo-code steps
  - missing error-code/HTTP mapping in failure branches
  - missing DTO type/required/validation constraints

### 6-7) Document Style Rules (Example Alignment)

Reference examples: `00_agd/examples/ko/*`

- Use block-style `@section { ... }` as the standard.
- Keep block key order as `summary -> links -> path -> content`.
- Keep tab (`\t`) indentation for `content` body lines.
- Style changes must preserve semantics (format-only normalization).

### 6-8) General Quality Standards

- core logic docs: executable contract quality
- new project docs: traceable to core logic through `source_doc`/`source_sections` when they depend on existing behavior
- keep `@map`/`@section`/`@change` consistent
