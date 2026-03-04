# AGD AI Execution Guide (EN, Token-Light)

This file is the minimal runbook for AI to write/update planning docs by reading only `00_agd/agd_docs`.

## 1) Scope

- Default read/write root: `00_agd/agd_docs`
- Read order:
  1. `10_source/*`
  2. `20_derived/*`
  3. `30_shared/*`
- Exclude `END__*` docs by default.

## 2) Required Input Contract

- `target_docs`: target file paths/names
- `target_sections`: section IDs (ex: `CORE-020`, `SYS-070`)
- `goal`: 1-2 sentences
- `constraints`: limits / do-not-touch rules
- `done`: validation criteria
- Keep `@meta.doc_base_path` (default is empty)
- Keep section `path:` in each `@section` (example: `10_source/product/core_spec.agd#CORE-020`)

## 3) Fixed Execution Loop

1. Read source baseline
2. Confirm target sections
3. Update `@section`
4. Append `@change` with both `reason` and `impact`
5. Keep `@map` and `@section` aligned
6. Pass `check` or `check-all`

## 4) Must Rules

- Keep one `source` doc per topic
- For derived docs, require `authority: derived` + `source_doc` + `source_sections`
- Every mutation must include `@change reason/impact`
- Do not finish with map/section mismatch
- If conflict exists, source wins
- If `target_sections` is missing, do not mutate
- If `10_source` baseline is missing/conflicting, ask one question and wait
- Do not report completion when `check`/`check-all` fails
- Create new files only when explicitly requested
- If both section `path` and `ID` are available, use `path` first
- Fill `doc_base_path` only when a real source path is explicitly known (no auto-relative default)

## 5) Recommended AI Output Format

Use this 6-line output only.

```txt
Changed files: <path1>, <path2>
Changed sections: <SEC-ID>, <SEC-ID>
Reason: <main reason>
Impact: <impact/follow-up>
Validation: <check summary>
Risks: <none|summary>
```

## 6) Minimal Commands

```cmd
agd.exe check <file.agd>
agd.exe check-all
agd.exe map-sync <file.agd> --reason "..." --impact "..."
agd.exe role-set <file.agd> source
agd.exe role-set <file.agd> derived <source-doc> "<source_sections>"
agd.exe role-graph --scope all
```

## 7) Role-Graph Scope Filter

- `all`: all docs
- `maintenance`: maintenance only (with linked nodes)
- `incident`: incident only (with linked nodes)
- `maintenance-incident`: maintenance + incident only
- `exclude-maintenance-incident`: all except maintenance/incident
