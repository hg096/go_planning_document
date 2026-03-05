# AGD Quick Start Guide (Beginner)

AGD is a document format for **human + AI collaboration on one source of truth**.

Core goals:

1. Find edit points quickly in large documents.
2. Make edits by stable section IDs.
3. Keep change reasons in the same file.

## 1. When Documents Grow, Complexity Grows Too

This project started from these questions:

- Is there an ideal metric that every successful plan should align to?
- As we keep creating and editing documents, do files multiply while management gets harder?
- If two documents describe different logic, do we end up in an endless rework loop?

AGD's answer is not "one perfect document." It is an **operating system that detects conflicts early and converges to a source document**.

What this project tries to solve:

1. Keep a stable source of truth even when document count grows.
2. Make humans and AI edit the same explicit section targets.
3. Preserve decision-change reasons end to end.

How we operate:

- Keep one `source` document per topic.
- Mark downstream documents as `derived` and connect them with `source_doc` and `source_sections`.
- Use `check` as a default quality gate to detect document conflicts early.
- Update `@map`, `@section`, and `@change` together.
- Run a weekly routine (update -> validate -> reconcile) to control document sprawl.

Success metrics are not document counts:

- time to detect conflicts
- percentage of changes with recoverable rationale
- consistency: humans and AI produce the same answer to the same question

## 2. Quick Start

Recommended Go version: `1.25.1`

### Setup (root README first)

Use this sequence in project root:

```cmd
REM 1) download only 00_agd into current project root
cmd /v:on /c "set T=%TEMP%\agd_tmp_%RANDOM%%RANDOM%&&(git clone -n --depth 1 --filter=blob:none --sparse https://github.com/hg096/go_planning_document.git "!T!" && git -C "!T!" sparse-checkout set 00_agd && git -C "!T!" checkout -q && xcopy "!T!\00_agd" "00_agd" /e /i /y >nul) & set EC=!ERRORLEVEL! & if exist "!T!" rd /s /q "!T!" & exit /b !EC!"

REM 1-1) preserve current 00_agd/agd_docs, 00_agd/policy and download the rest into project root
cmd /v:on /c "set T=%TEMP%\agd_tmp_%RANDOM%%RANDOM%&&(git clone -n --depth 1 --filter=blob:none --sparse https://github.com/hg096/go_planning_document.git "!T!" && git -C "!T!" sparse-checkout set 00_agd && git -C "!T!" checkout -q && (if not exist "00_agd" mkdir "00_agd") && robocopy "!T!\00_agd" "00_agd" /E /XD "!T!\00_agd\agd_docs" "!T!\00_agd\policy" /R:1 /W:1 /NFL /NDL /NJH /NJS /NP >nul) & set EC=!ERRORLEVEL! & if !EC! LSS 8 set EC=0 & if exist "!T!" rd /s /q "!T!" & exit /b !EC!"

REM 2) local setup (hooks + baseline checks)
00_agd\setup.cmd

REM 3) verify hooks path
git config --get core.hooksPath

REM 4) baseline validation
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict

REM 5) install CI/PR templates (optional)
00_agd\setup.cmd -InstallCiTemplate -InstallPrTemplate
```

Expected hooks path:

```txt
00_agd/.githooks
```

Build only when needed:

```cmd
go build -o 00_agd\agd.exe ./cmd/agd
go build -ldflags "-X main.defaultLang=en" -o 00_agd\agd_en.exe ./cmd/agd
go build -ldflags "-X main.defaultLang=ko" -o 00_agd\agd_ko.exe ./cmd/agd
```

Run:

```cmd
00_agd\agd_en.exe
REM or
00_agd\agd_ko.exe
```

Enforcement note:

- local hooks can be bypassed with `--no-verify`
- final enforcement should be Required CI check (`AGD Guard`)
- package details: [README.en.md](/D:/codePack/go/new_read/00_agd/README.en.md)

## 3. Easiest Start: Wizard


English wizard menu:

```txt
1.
Select document

[1] Select document
[2] Generate doc kit (starter/new-project/maintenance/error)
[3] Validate whole docs tree
[4] New document
[5] Show source/derived relation graph
[0] Back/Exit
```

After selecting a document, the per-document menu is:

```txt
2.
Selected document: <document>
[1] Search keyword
[2] Add new section (also updates map)
[3] Add logic-change record
[4] Sync map only
[5] Check document
[6] Export to markdown
[7] Set source/derived role
[0] Back
```

Wizard mutation action (`3`) now runs a guided flow:

- `reason` is required
- automatic `map-sync` after mutation
- automatic `check` right after sync

After new document creation (top menu `4`), wizard also asks you to choose `source/derived/later` so authority setup starts early.
For new document creation, folder selection is limited to each doc type's allowed path (and its subfolders).
All selection-style wizard screens (doc type/profile/folder/document/section/role/format) support `[0] Back`.
In wizard menu `4` (New document), the default type set includes these 8 types:

- `service`
- `core-spec`
- `delivery-plan`
- `meeting`
- `experiment`
- `roadmap`
- `handoff`
- `policy`

Document creation is focused on the same 8 types.

Path input is resolved against `00_agd\agd_docs` by default.
When opening existing docs, filename-only input is searched across subfolders.
If duplicate filenames exist under subfolders, use a relative path like `front/home`.
For AI writing philosophy and operating rules, use `00_agd/README.md` (English: `00_agd/README.en.md`) as the unified guide.

For new document creation (wizard top menu `4`), filename-only output is auto-routed by doc type:

- `service` -> `00_agd\agd_docs\10_source\service\<file>.agd`
- `core-spec` -> `00_agd\agd_docs\10_source\product\<file>.agd`
- `delivery-plan` -> `00_agd\agd_docs\20_derived\frontend\<file>.agd`
- `policy` -> `00_agd\agd_docs\10_source\policy\<file>.agd`
- `meeting` -> `00_agd\agd_docs\30_shared\meeting\<file>.agd`
- `experiment` -> `00_agd\agd_docs\30_shared\experiment\<file>.agd`
- `roadmap` -> `00_agd\agd_docs\30_shared\roadmap\<file>.agd`
- `handoff` -> `00_agd\agd_docs\30_shared\handoff\<file>.agd`

You can also create documents by specifying an explicit target path.
For page-by-page management, place `delivery-plan` docs under `20_derived\frontend` subfolders (for example, `checkout`, `home`).

## 5. Document Roles: source / derived


`source_sections` mapping rules:

- `SRC->DST`: relation/existence check
- `SRC=>DST`: strict title/summary conflict check
- `SEC-ID`: strict same-ID conflict check
- `auto`: generate `->` mappings
- `strict-auto`: generate `=>` mappings
- `smart-auto`: start strict, then relax mismatches to `->`

## 6. Change Log (`@change`) Guidance

Every document mutation must append an `@change` entry with both `reason` and `impact`.

`@change` IDs only need to be unique inside the document.
Recommended patterns:

- `CHG-2026-02-23-01` (readability)
- `CHG-20260223-173953` (timestamp uniqueness)

### Failure Analysis Frame in the Weekly Loop

- Classify failures by mechanism, not by feature name.
- After each incident-case (and postmortem if created), record source-doc feedback evidence before closure.
- In weekly reviews, aggregate Top3 document validation failure causes and assign actions.
- Before release, keep `@change reason/impact` missing count at zero.
- Full rule set: `00_agd/docs/AGD_TEMPLATE_GUIDE_en.md`

### Service-Logic Changes: Doc-First Rule

- Before service-code edits, update the service source doc (`10_source/service`) first.
- Keep `@change reason/impact` complete before implementation.
- After edits, run document-wide validation to verify consistency.


## 8. Enforced Ops Setup (Git Hook + CI)

To avoid conflicts with ongoing host projects, AGD is isolated under `00_agd/` and managed from there.
Root-level hook/workflow files are not forced by default; install is template-based.

### 8-1) Local pre-commit hook

1) Run the package setup script to enable local hooks.

2) Hook files included in this repo:

- `00_agd/.githooks/pre-commit`
- `00_agd/scripts/git-hooks/pre-commit.ps1`

3) Enforced checks at commit time:

- Checks only staged `.agd` changes under `00_agd/`
- Runs strict tree validation for `00_agd\agd_docs`
- Runs strict validation for `00_agd\examples` when examples changed

### 8-2) GitHub Actions required check

Template location:

- `00_agd/templates/agd-guard.yml`

This is installed during the default setup flow.

The CI template is copied to `.github/workflows/agd-guard.yml`.

### 8-3) Pull request checklist

Template location:

- `00_agd/templates/pull_request_template.md`

This is installed during the default setup flow.

The PR template is copied to `.github/pull_request_template.md`.

## 9. References

- Korean start guide: `README.md`
- Korean AI execution guide: `00_agd/docs/AGD_TEMPLATE_GUIDE_ko.md`
- English AI execution guide: `00_agd/docs/AGD_TEMPLATE_GUIDE_en.md`
- Unified docs-root/operations guide: `00_agd/README.en.md` (Korean source: `00_agd/README.md`)
- Isolated package guide (EN): `00_agd/README.en.md`
- Isolated package guide (KO): `00_agd/README.md`
- One-command bootstrap (cmd): `00_agd/setup.cmd`
- Bootstrap script (PowerShell): `00_agd/scripts/setup.ps1`
- Gate-enforced execution wrapper (cmd): `run-safe.cmd`

