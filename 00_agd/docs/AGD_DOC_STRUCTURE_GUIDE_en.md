# AGD Document Structure Guide (EN)

This guide explains where to place AGD files under `00_agd\agd_docs` and how to run a structure that stays manageable as documents grow.

## 1) Standard Folder Layout

```text
00_agd/agd_docs/
  00_inbox/
  10_source/
    product/
    service/
    policy/
    architecture/
  20_derived/
    frontend/
    qa/
    ops/
  30_shared/
    meeting/
    handoff/
    roadmap/
    postmortem/
    experiment/
    maintenance/
    errFix/
  90_archive/
```

## 2) Folder Intent

- `00_inbox`: temporary drafts before classification
- `10_source/*`: source-of-truth docs by topic
- `20_derived/*`: downstream docs derived from source docs
- `30_shared/*`: team coordination docs plus single-case maintenance/incident workspaces
- `90_archive`: retired docs kept for history

## 3) Type-to-Folder Auto Routing

When you run `agd.exe new <type> <name>` with filename-only `<name>`, AGD auto-routes by type.

| Type | Auto Folder | Recommended Role |
|---|---|---|
| `prd` | `10_source/product` | source |
| `service-logic` | `10_source/service` | source |
| `policy` | `10_source/policy` | source |
| `adr` | `10_source/architecture` | source |
| `frontend-page` | `20_derived/frontend` | derived |
| `qa-plan` | `20_derived/qa` | derived |
| `runbook` | `20_derived/ops` | derived |
| `meeting` | `30_shared/meeting` | shared |
| `handoff` | `30_shared/handoff` | shared |
| `roadmap` | `30_shared/roadmap` | shared |
| `postmortem` | `30_shared/postmortem` | shared |
| `experiment` | `30_shared/experiment` | shared |
| `maintenance-case` | `30_shared/maintenance` | source (single case) |
| `incident-case` | `30_shared/errFix` | source (single case) |
| other | `00_inbox` | triage |

To override, pass an explicit path:

```cmd
agd.exe new prd 10_source/product/checkout_v3
```

## 4) source / derived Operating Rules

- Keep exactly one source doc per topic.
- Mark downstream docs with `authority: derived`.
- Always set `source_doc` and `source_sections` on derived docs.
- Run `agd.exe check <file>` after changes.

Example:

```cmd
agd.exe role-set service_logic_checkout_core_en source
agd.exe role-set frontend_page_checkout_en derived service_logic_checkout_core_en "SYS-020->FP-020,SYS-030=>FP-030"
agd.exe check frontend_page_checkout_en
```

## 5) Recommended Authoring Order

1. `10_source/service`: service-wide logic
2. `10_source/product`: requirements and scope
3. `10_source/policy`: release/ops rules
4. `20_derived/frontend`: page-level implementation logic
5. `20_derived/qa`, `20_derived/ops`: QA and runbook derivatives
6. `30_shared/*`: meeting, handoff, roadmap, postmortem + maintenance/incident case accumulation

## 5-1) Auto-Exclude Rule for Closed Docs (`END__`)

- `30_shared/maintenance/END__*_maintenance_case.agd`
- `30_shared/errFix/END__*_incident_case.agd`

These two patterns are auto-excluded from scan/select/role-graph flows.

## 6) Routine to Prevent Document Sprawl

Daily:

1. Edit by explicit `SECTION-ID`
2. Require both `--reason` and `--impact` for mutations
3. Run `map-sync -> check`
4. Before service-logic code edits, pass `agd.exe service-gate`
5. Wrap code execution with `run-safe.cmd` to prevent gate bypass

Weekly:

1. Triage `00_inbox`
2. Resolve duplicates/conflicts
3. Move stale docs into `90_archive`

## 7) Conflict Prevention Checklist

- Do we have more than one source doc for the same topic?
- Are `source_doc` and `source_sections` set on every derived doc?
- Are `@map` summaries aligned with `@section` content?
- Was `agd.exe check` run after modifications?
- Was `agd.exe service-gate` run before service code edits?
- Were stale docs archived instead of left ambiguous?

## 8) How to Read with Examples

- Source examples: `examples/en/10_source/*`
- Derived examples: `examples/en/20_derived/*`
- Shared examples: `examples/en/30_shared/*`
- Template index: `examples/TEMPLATE_SHOWCASE_INDEX_en.md`

Korean counterpart: `docs/AGD_DOC_STRUCTURE_GUIDE_ko.md`
