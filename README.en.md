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

## 2. Build

Recommended Go version: `1.25.1`

```cmd
go build -o agd.exe ./cmd/agd
```

Build separate language defaults:

```cmd
go build -ldflags "-X main.defaultLang=en" -o agd_en.exe ./cmd/agd
go build -ldflags "-X main.defaultLang=ko" -o agd_ko.exe ./cmd/agd
```

## 3. Easiest Start: Wizard

```cmd
agd.exe
REM or
agd.exe wizard
```

English wizard menu:

```txt
1.
Select document

[1] Select document
[2] Generate doc kit (starter/maintenance/new/incident)
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
In wizard menu `4` (New document), the default type set is reduced to these 7 integrated-first types:

- `core-spec`
- `delivery-plan`
- `meeting`
- `experiment`
- `roadmap`
- `handoff`
- `policy`

CLI creation (`agd.exe new`, `agd.exe init`) is limited to the same 7 integrated-first types. Split/legacy types are blocked.

Path input is resolved against `agd_docs` by default.
When opening existing docs, filename-only input is searched across subfolders.
If duplicate filenames exist under subfolders, use a relative path like `front/home`.
`agd_docs/README.md` is scaffold-generated in the selected language and includes the AI writing philosophy/operating guide.

For new document creation (`agd new`, wizard top menu `4`), filename-only output is auto-routed by doc type:

- `core-spec` -> `agd_docs\10_source\product\<file>.agd`
- `delivery-plan` -> `agd_docs\10_source\product\<file>.agd`
- `policy` -> `agd_docs\10_source\policy\<file>.agd`
- `meeting` -> `agd_docs\30_shared\meeting\<file>.agd`
- `experiment` -> `agd_docs\30_shared\experiment\<file>.agd`
- `roadmap` -> `agd_docs\30_shared\roadmap\<file>.agd`
- `handoff` -> `agd_docs\30_shared\handoff\<file>.agd`

To override, pass an explicit path.
Example: `agd.exe new core-spec 10_source/product/checkout_v2`

## 4. Common Commands

```cmd
agd.exe new core-spec core_spec_checkout "Checkout Core Spec" "product-team"
agd.exe check core_spec_checkout
agd.exe check-all
agd.exe check-all --strict
agd.exe check-all --include-archive
agd.exe search core_spec_checkout payment
agd.exe edit core_spec_checkout CORE-020 "Clarify must-have feature priorities" --reason "re-align feature priority" --impact "reduce priority interpretation mismatch"
agd.exe add core_spec_checkout CORE-020 "- Add failure recovery flow" --reason "improve operational recovery guidance" --impact "faster recovery action during incidents"
agd.exe section-add core_spec_checkout CORE-050 "Payment Failure Recovery" "Define recovery procedure" "RUN-001,LOG-002" "- Retry guidance" --reason "document missing recovery policy" --impact "consistent failure handling communication"
agd.exe logic-log core_spec_checkout CORE-020 --reason "Applied retry branch update" --impact "lower drop rate on transient payment failures"
agd.exe incident-tag checkout_incident_case FT-CHECKOUT service_logic_checkout_core [section-id] --reason "set issue root to exact feature section" --impact "AI can track the target section precisely"
agd.exe kit starter-kit checkout
agd.exe maintenance checkout --owner ops-team
agd.exe incident-response checkout --feature-tag FT-CHECKOUT
agd.exe kit starter-kit checkout --no-graph
agd.exe role-graph
agd.exe role-graph --format mermaid --out agd_docs\role_graph.mmd
agd.exe view core_spec_checkout
```

`edit`/`add`/`section-add` require both `--reason` and `--impact`, and then automatically run `map-sync -> check`.

Kit profile intent:

- `starter-kit`: minimal source+derived baseline for first setup
- `maintenance`: single maintenance case doc (`agd_docs/30_shared/maintenance/<project>_maintenance_case.agd`)
- `new-project`: set for adding or improving new feature scope
- `incident-response`: single incident case doc (`agd_docs/30_shared/errFix/<project>_incident_case.agd`)
- `maintenance`/`incident-response`: do not create a `<project>` subfolder; only create the single file at the path above
- manual `postmortem` type path: `agd_docs/30_shared/postmortem/<file>.agd`
- `incident-response` auto-injects the rooted trace block into the generated `incident-case` doc using `--feature-tag`.
- `incident-case` default flow: `INC-001(tag issue) -> INC-010(capture bug) -> INC-020(quick RCA) -> INC-030(fix direction) -> INC-040(AI handoff) -> INC-050(AI result) -> INC-060(validate/close)`
- If `--feature-tag` is omitted, an automatic tag is generated from the project key (example: `checkout` -> `FT-CHECKOUT`).
- Close-state rule: `END__*_maintenance_case.agd` (maintenance) and `END__*_incident_case.agd` (errFix) are auto-excluded from scan/select/role-graph.

`ai_planning_guide` content is now merged into `agd_docs/README.md` instead of being auto-generated as a separate file.
After kit creation, use `agd_docs/README.md` as the policy anchor so AI output stays aligned with source/derived rules and rationale standards.

## 5. Document Roles: source / derived

```cmd
agd.exe role-set service_overview source
agd.exe role-set frontend_pages derived service_overview "SYS-020->FP-020,SYS-030=>FP-030"
agd.exe role-set frontend_pages derived service_overview auto
agd.exe role-set frontend_pages derived service_overview strict-auto
agd.exe role-set frontend_pages derived service_overview smart-auto

agd.exe map-suggest frontend_pages service_overview
agd.exe map-suggest frontend_pages service_overview strict-auto
agd.exe map-suggest frontend_pages service_overview smart-auto

agd.exe starter-kit checkout
agd.exe maintenance checkout
agd.exe new-project checkout
agd.exe incident-response checkout --feature-tag FT-CHECKOUT
```

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
- In weekly reviews, aggregate Top3 `agd.exe check` failure causes and assign actions.
- Before release, keep `@change reason/impact` missing count at zero.
- Full framework: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_en.md`

### Service-Logic Changes: Doc-First Rule

- Before service-code edits, update the service source doc (`10_source/service`) first.
- Keep `@change reason/impact` complete before implementation.
- After edits, run `agd.exe check`/`agd.exe check-all` to verify document consistency.

## 7. New Document Types (`agd new`)

- `core-spec`: Unified core spec (PRD+logic+ADR+AI guide)
- `delivery-plan`: Unified delivery plan (frontend+QA+release)
- `meeting`: Meeting notes and decisions
- `experiment`: Experiment and A/B test plan
- `roadmap`: Quarter/half roadmap plan
- `handoff`: Cross-team handoff document
- `policy`: Policy and guideline document

For `agd init --list`, Korean and English template names are available for the same 7 types:

- Korean templates: `*-ko` (example: `core-spec-ko`)
- English templates: `*-en` (example: `core-spec-en`)

Template showcase indexes:

- `examples/TEMPLATE_SHOWCASE_INDEX_ko.md`
- `examples/TEMPLATE_SHOWCASE_INDEX_en.md`
- `examples/README.md`

Examples folder structure:

- `examples/ko/10_source/*`: Korean source examples
- `examples/ko/20_derived/*`: Korean derived examples
- `examples/ko/30_shared/*`: Korean shared examples
- `examples/en/10_source/*`: English source examples
- `examples/en/20_derived/*`: English derived examples
- `examples/en/30_shared/*`: English shared examples

For single-case maintenance/incident flows, use kit-generated paths instead of fixed example files:

- `agd_docs/30_shared/maintenance/<project>_maintenance_case.agd`
- `agd_docs/30_shared/errFix/<project>_incident_case.agd`

## 8. References

- Korean start guide: `README.md`
- Korean template guide: `docs/AGD_TEMPLATE_GUIDE_ko.md`
- English template guide: `docs/AGD_TEMPLATE_GUIDE_en.md`
- Korean document structure guide: `docs/AGD_DOC_STRUCTURE_GUIDE_ko.md`
- English document structure guide: `docs/AGD_DOC_STRUCTURE_GUIDE_en.md`
- Korean failure analysis framework: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_ko.md`
- English failure analysis framework: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_en.md`
- Docs root folder guide: `agd_docs/README.md`
- Gate-enforced execution wrapper (cmd): `run-safe.cmd`
- Spec (KR): `docs/AGD_SPEC_v0.1.md`
- Spec (EN): `docs/AGD_SPEC_v0.1.en.md`
