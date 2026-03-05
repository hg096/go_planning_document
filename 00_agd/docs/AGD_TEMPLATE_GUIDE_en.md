# AGD AI Execution Guide (EN, Token-Light)

This file is the minimum operating guide for AI to read only `00_agd/agd_docs` and update planning docs without missing documentation steps.

## 1) Work Scope

- Read/write root: `00_agd/agd_docs`
- Read priority:
  1. `10_source/*`
  2. `20_derived/*`
  3. `30_shared/*`
- Exclude `END__*` docs by default

## 2) AI Input Contract (Required)

Do not mutate if any of the following is missing.

- `target_docs`: target document path(s)/name(s)
- `target_sections`: section IDs or section paths
- `goal`: 1-2 sentence change goal
- `constraints`: forbidden scope / limits
- `done`: validation-based completion criteria

## 3) Path Rules (Critical)

- Always use `00_agd/agd_docs` as the base root
- Template/example docs include `@meta.doc_base_path` and each `@section.path` key by default
- Keep initial `@meta.doc_base_path` empty
- Keep default `@section.path` empty in template/example docs
- Fill `doc_base_path` only when an actual source code path is explicitly known
- No auto-relative path conversion and no auto-fill
- If `@section.path` exists, it takes precedence over `doc_base_path`
- Recommended `section.path` format:
  - `src/.../file.ext#SEC-ID`
  - `backend/.../handler.go#CORE-020`

## 4) Fixed Execution Loop

1. Confirm source baseline document
2. Confirm target sections (`target_sections`)
3. Update `@section`
4. Append `@change` with both `reason` and `impact`
5. Validate `@map` and `@section` consistency
6. Run validation (`check` or `check-all`)
7. If failed, fix and rerun

## 5) Must Rules

- Keep one `source` doc per topic
- Derived docs must include `authority: derived`, `source_doc`, `source_sections`
- Every mutation must record `@change(reason/impact)`
- Never finish with `@map`/`@section` mismatch
- Source has priority when conflict exists
- Do not mutate when `target_sections` is missing
- Do not report completion if `check`/`check-all` fails
- Create new files only when explicitly requested

## 6) Maintenance / Incident Doc Rules

- Maintenance doc:
  - `maintenance_feature_tag`, `maintenance_source_doc`, `maintenance_source_section`
  - `[MAINTENANCE-ROOT-TAG]...[/MAINTENANCE-ROOT-TAG]`
- Incident doc:
  - `incident_feature_tag`, `incident_source_doc`, `incident_source_section`
  - `[INCIDENT-PROBLEM-TAG]...[/INCIDENT-PROBLEM-TAG]`

## 7) Code-Change Link Validation (Recommended Default)

After mutation, run both commands to prevent doc omissions.

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

Relation policy (bidirectional service/frontend section mapping):

- file: `00_agd/policy/code_plan_relations.txt`
- syntax: `<left-doc.agd>#<SECTION-ID> <-> <right-doc.agd>#<SECTION-ID>`
- one relation per line
- for multiple relations, add multiple lines
- for 1:N, repeat the left endpoint on each line
- for N:1, repeat the right endpoint on each line
- comma-separated multi-endpoint syntax in a single line is not supported
- linked docs pass only when `@change` is updated

Examples:

```txt
# 1:N (one service section -> multiple frontend sections)
00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010
00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-020

# N:1 (multiple frontend/service sections -> one target section)
00_agd/agd_docs/10_source/service/payment_service.agd#SYS-040 <-> 00_agd/agd_docs/20_derived/frontend/payment_page.agd#FP-030
00_agd/agd_docs/10_source/service/refund_service.agd#SYS-050 <-> 00_agd/agd_docs/20_derived/frontend/payment_page.agd#FP-030
```

Strict modes:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto --strict-mapping
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

## 8) Minimal Command Set

```cmd
00_agd\agd_en.exe check <file.agd>
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
00_agd\agd_en.exe map-sync <file.agd> --reason "..." --impact "..."
00_agd\agd_en.exe role-set <file.agd> source
00_agd\agd_en.exe role-set <file.agd> derived <source-doc> "<source_sections>"
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
```

## 9) AI Output Format (Recommended, 6 lines)

```txt
Changed files: <path1>, <path2>
Changed sections: <SEC-ID>, <SEC-ID>
Reason: <main reason>
Impact: <impact/follow-up>
Validation: <check/code-plan-check summary>
Risks: <none|summary>
```

## 10) Frequent Failure Patterns (Quick Check)

- `@change` missing `impact`
- `derived` doc missing `source_doc/source_sections`
- Broad mutation without `target_sections`
- `@map` and `@section` ID mismatch
- Auto-filled `doc_base_path` from guessed path
- Reporting completion while validation is failing
