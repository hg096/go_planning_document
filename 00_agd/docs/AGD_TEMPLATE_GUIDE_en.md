# AGD AI Execution Guide (EN, Hybrid v2)

This guide is the guide-first operating contract for keeping AGD outputs consistent even when user prompts are low-detail.

## 1) Operating Defaults

- Operating mode: `Hybrid mode`
- Automation surface: `Guide-first`
- Language policy: `Bilingual (ko/en)`
- Principle: improve AGD documentation quality and reproducibility without changing runtime code/infrastructure.

## 2) Work Scope

- Read/write root: `00_agd/agd_docs`
- Read priority:
  1. `10_source/*`
  2. `20_derived/*`
  3. `30_shared/*`
- Exclude `END__*` docs by default

## 3) AI Input Contract v2

### 3.1 Required inputs (fail if missing)

- `goal`: 1-2 sentence objective
- `constraints`: scope limits and forbidden actions
- `done`: validation-based completion criteria

### 3.2 Semi-required inputs (warn and proceed if missing)

- `target_docs`: target document path(s)
- `target_sections`: target section IDs or section paths

### 3.3 Missing-input handling

1) Missing required inputs:
- stop mutation
- return only missing key list

2) Missing semi-required inputs:
- explore repo and propose candidate scope
- proceed with explicit assumptions
- record as warnings in output

## 4) Path Rules

- Always use `00_agd/agd_docs` as the base root
- `@meta.doc_base_path` may remain empty initially
- Fill `doc_base_path` only when a real source path is confirmed
- If `@section.path` exists, it takes precedence
- No auto-relative conversion and no guessed auto-fill

## 5) Fixed Execution Loop

1. Confirm source baseline document
2. Confirm target docs/sections
3. Update `@section`
4. Append `@change` with `reason` and `impact`
5. Validate `@map`/`@section` consistency
6. Run validations (`check-all`, `role-graph`, `code-plan-check`)
7. Apply minimal fixes and rerun on failure

## 6) Hybrid Mode Decision Table

### 6.1 Fail (completion report prohibited)

- `@map`/`@section` ID mismatch
- derived doc missing `source_doc` or `source_sections`
- reporting completion while any validation fails

### 6.2 Warn (allowed, must be reported)

- `doc_base_path` intentionally left empty
- partial missing `@section.path`
- minimum metadata missing in unassigned shared docs

### 6.3 Pass (completion allowed)

- `check-all --strict` passed
- `role-graph` unresolved=0
- `code-plan-check` passed
- changed sections, evidence, and impacts reported

## 7) Mandatory Coverage by Doc Type

### 7.1 `10_source/service`

Each domain section must include 9 blocks:
1) Responsibility and boundary
2) Call sequence
3) Input DTO contract
4) Branching/state transitions
5) Transaction/atomicity
6) Cache key/invalidation
7) Error code/HTTP mapping
8) Pseudocode
9) Test criteria

### 7.2 `10_source/policy`

Must include 3 blocks:
1) Policy decision table
2) Execution template
3) Evidence checklist

### 7.3 `20_derived/frontend/public`

Must include 8 blocks:
1) Rendering model
2) State variable contract
3) Initialization sequence
4) Event-state transitions
5) API contract (request/response)
6) Error/empty fallback
7) Browser storage contract
8) Backend mapping

### 7.4 `20_derived/frontend/admin`

Must match public-level detail with:
- state/transition/API contracts
- permission codes
- failure branches and access-block rules

### 7.5 `20_derived/qa`

Required axes:
- API regression
- permission matrix
- recovery/observability
- release smoke

### 7.6 `20_derived/ops`

Required axes:
- env baseline
- boot failure diagnosis
- deploy/rollback
- DB/collection operation checks

## 8) Relation-Mapping Rules

- relation file: `00_agd/policy/code_plan_relations.txt`
- syntax: `<left>#<SEC> <-> <right>#<SEC>`
- one relation per line; repeat endpoint for 1:N and N:1

### 8.1 Mandatory synchronized updates

- when source sections change, update linked derived `source_sections` in the same changeset
- update linked `@change` entries together

### 8.2 Service section renumbering checklist

1) update source `@map`/`@section` first
2) replace derived `source_sections` globally
3) replace relation endpoints
4) run: `check-all` -> `role-graph` -> `code-plan-check --strict-relation`

## 9) Block Style Standard (aligned with examples)

Reference examples: `00_agd/examples/ko/*`

- use block-style `@section` with `{ ... }`
- fixed key order inside blocks: `summary -> links -> path -> content`
- keep tab indentation (`\t`) inside `content`

Example:

```agd
@section SEC-001
title: Title
{
	summary: Summary
	links: LINK-001
	path: src/path/file.go
	content:
	<<<
		line 1
		line 2
	>>>
}
```

## 10) Minimal User Prompt Templates (5 lines)

### 10.1 New setup

1) Goal: <goal>
2) Scope: <target_docs>
3) Constraints: <constraints>
4) Done: <done>
5) Language: <ko|en>

### 10.2 Partial enhancement

1) Goal: <goal>
2) Scope: <target_sections>
3) Constraints: <constraints>
4) Done: <done>
5) Language: <ko|en>

### 10.3 Section migration

1) Goal: <section renumber/restructure>
2) Scope: <source + derived + relation>
3) Constraints: <constraints>
4) Done: <done>
5) Language: <ko|en>

### 10.4 Validation failure recovery

1) Goal: <failed rule recovery>
2) Scope: <failing docs>
3) Constraints: <constraints>
4) Done: <done>
5) Language: <ko|en>

## 11) Minimum Validation Commands

```cmd
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

## 12) Blind Reconstruction Test (Required)

- Input: only `10_source/service` doc, with code reference prohibited
- Output: domain-level pseudo implementation (handler/service/repository order, DTOs, failure branches, state transitions)
- Pass criteria:
  - zero missing required blocks
  - zero missing DTO field/validation details
  - zero missing error-code/HTTP mapping in failure branches
  - zero mismatch in routing-permission-rate-limit mapping

If failed:
1) patch missing source sections
2) record `@change(reason/impact)`
3) rerun validation commands and blind reconstruction test

## 13) Output Quality Rubric (10 points)

- Axis 1 Structural integrity: 0/1
- Axis 2 Implementability: 0-3
- Axis 3 Operability: 0-3
- Axis 4 Validation evidence: 0-3

Pass threshold:
- total >= 9
- structural integrity must be 1

[Deduction rules]
- pseudocode fewer than 10 steps per domain section
- missing error-code/HTTP mapping in failure branches
- missing type/required/constraint in DTO definitions

## 14) Recommended AI Output Format

```txt
Changed files: <path1>, <path2>
Changed sections: <SEC-ID>, <SEC-ID>
Reason: <main reason>
Impact: <impact/follow-up>
Validation: <check/role-graph/code-plan-check summary>
Risks: <none|summary>
```

## 15) Frequent Failure Patterns

- mutating while required inputs (`goal/constraints/done`) are missing
- derived doc missing `source_doc/source_sections`
- `@map` and `@section` ID mismatch
- source changed without derived/relation synchronization
- reporting completion while validations are failing
- reporting completion without blind reconstruction proof
