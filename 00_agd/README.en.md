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
- Read source docs under `00_agd/agd_docs/10_source/*` first.
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

- `00_inbox`: temporary draft and pre-triage notes
- `10_source/*`: source-of-truth documents
- `20_derived/*`: derived documents linked to source docs
- `30_shared/*`: shared planning/collaboration docs
- `90_archive`: archived documents

### 6-2) Recommended Placement by Topic

- `10_source/product`: product scope and success criteria
- `10_source/service`: service/backend logic baseline
- `10_source/policy`: rules/standards/approval/release policy
- `10_source/architecture`: ADR/trade-off records

- `20_derived/frontend`: page-level implementation/validation/release docs
- `20_derived/qa`: regression/permission/recovery test docs
- `20_derived/ops`: operations runbook docs

- `30_shared/meeting`: decision/action logs
- `30_shared/handoff`: release handoff package
- `30_shared/roadmap`: milestones/priorities
- `30_shared/postmortem`: incident retrospectives
- `30_shared/experiment`: experiment records
- `30_shared/maintenance`: single-flow maintenance workspace
- `30_shared/errFix`: single-flow incident/bug workspace

### 6-3) Guide-First Operation

1) Read `00_agd/docs/AGD_TEMPLATE_GUIDE_ko.md` or `AGD_TEMPLATE_GUIDE_en.md` first.
2) For service topics, keep `10_source/service` as the single authoritative flow.
3) Link derived docs from the source baseline.
4) When service source changes, sync linked derived/source_sections in the same change set.
5) Always record `@change(reason/impact)`.
6) Do not report completion before passing 3 validations + blind reconstruction test.

### 6-4) Minimal Command Set

```cmd
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

### 6-5) Blind Reconstruction Test

- input: only `10_source/service` document(s) (no code reference)
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

- source docs: executable contract quality
- derived docs: traceable via source sections and `source_sections`
- shared docs: preserve decision context needed for operations/collaboration
- keep `@map`/`@section`/`@change` consistent

### 6-9) END__ Close Rules

- maintenance close docs: `30_shared/maintenance/END__*_maintenance_case.agd`
- incident close docs: `30_shared/errFix/END__*_incident_case.agd`
- `END__*` docs are excluded from default scan/select/role-graph flows
- remove the `END__` prefix to reopen
