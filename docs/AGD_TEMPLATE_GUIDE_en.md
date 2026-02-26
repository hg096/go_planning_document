# AGD Template Guide (Latest)

This guide explains template creation and operation flow for first-time `agd.exe` users.
Default document root is `agd_docs`.

## 0) Easiest Start: Wizard (`agd.exe`)

```cmd
agd.exe
REM or
agd.exe wizard
```

Top-level menu:

```txt
[1] Select document
[2] Generate doc kit (starter/maintenance/new/incident)
[3] Validate whole docs tree
[4] New document
[5] Show source/derived relation graph
[0] Back/Exit
```

After selecting a document:

```txt
Selected document: <name>
[1] Search keyword
[2] Add new section (also updates map)
[3] Add logic-change record
[4] Sync map only
[5] Check document
[6] Export to markdown
[7] Set source/derived role
[0] Back
```

All selection-style wizard screens (doc type/profile/folder/document/section/role/format) support `[0] Back`.

Wizard `New document` default set (document-count reduction focus):

- `core-spec`
- `delivery-plan`
- `meeting`
- `experiment`
- `roadmap`
- `handoff`
- `policy`

CLI creation (`agd.exe new`, `agd.exe init`) is also limited to the same 7 integrated-first types.

## 1) Fast Creation: `agd new`

```cmd
agd.exe new list
agd.exe new core-spec core_spec_checkout "Checkout Core Spec" "product-team"
```

Document types for `agd new`:

- `core-spec`: Unified core spec (PRD+logic+ADR+AI guide)
- `delivery-plan`: Unified delivery plan (frontend+QA+release)
- `meeting`: Meeting notes and decisions
- `experiment`: Experiment and A/B test plan
- `roadmap`: Quarter/half roadmap plan
- `handoff`: Cross-team handoff document
- `policy`: Policy and guideline document

Wizard `New document` uses the reduced set above by default.

## 2) Detailed Creation: `agd init`

```cmd
agd.exe init --list
agd.exe init --template core-spec-ko --out docs/core_spec_checkout_ko.agd --title "Checkout Core Spec (KO)" --owner "product-team"
agd.exe init --template core-spec-en --out docs/core_spec_checkout_en.agd --title "Checkout Core Spec" --owner "product-team"
```

Main options:

- `--template`: template name (`core-spec-ko`, `core-spec-en`, `delivery-plan-ko`, ...)
- `--out`: output file path
- `--title`: document title (`@meta.title`)
- `--owner`: document owner (`@meta.owner`)
- `--version`: document version (`@meta.version`, default `0.1`)
- `--force`: overwrite existing file

Note:

- `agd.exe init --list` now shows both `*-ko` and `*-en` templates.

## 3) Path Input Rules (`agd_docs` Default)

When opening existing docs, filename-only input is searched across `agd_docs` subfolders.

- Input: `core_spec_checkout`
- Search target: `agd_docs\**\core_spec_checkout.agd`

If duplicate filenames exist in subfolders, use a relative path like `front/home`.

For new docs (`agd new`, wizard top-level menu `4`), filename-only output is auto-routed by doc type:

- `core-spec`: `agd_docs\10_source\product\<file>.agd`
- `delivery-plan`: `agd_docs\10_source\product\<file>.agd`
- `policy`: `agd_docs\10_source\policy\<file>.agd`
- `meeting`: `agd_docs\30_shared\meeting\<file>.agd`
- `experiment`: `agd_docs\30_shared\experiment\<file>.agd`
- `roadmap`: `agd_docs\30_shared\roadmap\<file>.agd`
- `handoff`: `agd_docs\30_shared\handoff\<file>.agd`

To override, pass an explicit path.

`END__` close-state rule:

- `30_shared/maintenance/END__*_maintenance_case.agd`
- `30_shared/errFix/END__*_incident_case.agd`

These files are auto-excluded from scan/select/role-graph flows.

## 4) Section/Map Commands

```cmd
REM add section + update map
agd.exe section-add <file.agd> <SECTION-ID> "<section title>" "<summary>" [links] [content] --reason "<why>" --impact "<effect>"

REM sync map from sections
agd.exe map-sync <file.agd> --reason "<why>" --impact "<effect>"

REM edit section summary
agd.exe edit <file.agd> <SECTION-ID> "<new summary>" --reason "<why>" --impact "<effect>"

REM append one line to section content
agd.exe add <file.agd> <SECTION-ID> "<line to append>" --reason "<why>" --impact "<effect>"
```

Every document mutation must log `@change(reason/impact)`.
For `edit`/`add`/`section-add`, `map-sync -> check` runs automatically after mutation.

Whole-tree validation:

```cmd
REM validate full document tree (default: agd_docs)
agd.exe check-all
agd.exe check-all examples

REM treat warnings as failures
agd.exe check-all examples --strict

REM include archived documents
agd.exe check-all --include-archive
```

Recommended order for service-logic changes:

```cmd
REM update service doc first
agd.exe check 10_source/service/<service_doc>

REM verify full-tree consistency
agd.exe check-all
```

## 5) Source/Derived Document Operation

```cmd
REM source (authoritative) document
agd.exe role-set service_overview source

REM derived document with explicit mapping
agd.exe role-set frontend_pages derived service_overview "SYS-020->FP-020,SYS-030=>FP-030"

REM auto mapping modes
agd.exe role-set frontend_pages derived service_overview auto
agd.exe role-set frontend_pages derived service_overview strict-auto
agd.exe role-set frontend_pages derived service_overview smart-auto

REM preview mapping suggestions only
agd.exe map-suggest frontend_pages service_overview
agd.exe map-suggest frontend_pages service_overview strict-auto
agd.exe map-suggest frontend_pages service_overview smart-auto

REM validate
agd.exe check frontend_pages
```

`source_sections` rules:

- `SRC->DST`: relation/existence check
- `SRC=>DST`: strict title/summary conflict check
- `SEC-ID`: strict same-ID conflict check
- `auto`: generate `->` mappings
- `strict-auto`: generate `=>` mappings
- `smart-auto`: try strict first, then relax mismatches to `->`

## 6) Change Log (`@change`) Operation

```cmd
agd.exe logic-log <file.agd> <SECTION-ID> --reason "<change reason>" --impact "<change impact>"
```

`@change` IDs only need to be unique inside a document.
Recommended patterns:

- `CHG-2026-02-23-01` (readability)
- `CHG-20260223-173953` (timestamp uniqueness)

## 7) Recommended Flow

1. Create a document with wizard or `agd.exe new ...`.
2. Edit by section (`section-add`/`edit`/`add`).
3. Re-sync map with `map-sync`.
4. Record reason with `logic-log`.
5. Validate with `check`, then share with `view`.

## 8) Apply Failure Analysis Framework

Add the following items to your weekly operating loop.

1. Classify failures by mechanism (problem definition/dependency/execution/governance/learning loop).
2. Review Top3 recurring `agd.exe check` failure causes.
3. Keep release-candidate `@change reason/impact` missing count at zero.
4. Confirm source-doc feedback evidence before closing incident-cases/postmortems.

Detailed framework: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_en.md`
