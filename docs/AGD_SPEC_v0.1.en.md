# AGD v0.1 (Agent Guided Document)

AGD is a text format for large planning documents that are edited and read by both humans and AI agents.

## 1. Goals

1. Make large documents easy to scan quickly.
2. Let AI target exact edit locations using stable section IDs.
3. Keep change history inside the same document.
4. Enable strict machine validation through lint rules.

## 2. File Extension

- Recommended extension: `.agd`

## 3. Core Blocks

AGD is composed of top-level `@block` units. Every block header must start at column 1.

- `@meta`: document metadata
- `@map`: one-page section index
- `@section <SECTION-ID>`: section body
- `@change <CHANGE-ID>`: change log entry

## 4. Block Syntax

### 4.1 `@meta`

Use `key: value` lines.

Required keys:
- `title`
- `version`

Example:
```txt
@meta
title: Payment Platform Rewrite
version: 0.1
owner: product-team
```

### 4.2 `@map`

Use one row per section.

Format:
- `- <SECTION-ID> | <TITLE> | <SUMMARY>`

Example:
```txt
@map
- SEC-001 | Objective | Reach 99.9 percent payment success
- SEC-014 | Incident Rollback | Recover in under 5 minutes
```

### 4.3 `@section <SECTION-ID>`

Fields:
- `title:`
- `summary:`
- `links:` (comma-separated)
- `content:` (multi-line body)

`content:` must use `<<<` and `>>>` fences.

Example:
```txt
@section SEC-014
title: Incident Auto Rollback
summary: Switch payment traffic to the stable version when thresholds are exceeded
links: TEST-022, CODE-payment-rollback
content:
<<<
Triggers:
1) 5-minute failure rate over 3 percent
2) Approval latency P95 over 2.5 seconds
>>>
```

### 4.4 `@change <CHANGE-ID>`

Fields:
- `target:` (section ID)
- `date:` (`YYYY-MM-DD`)
- `author:`
- `reason:`
- `impact:` (required)

`CHANGE-ID` rules:
- Must be unique within the document.
- v0.1 does not enforce a strict regex pattern.
- Recommended patterns:
- `CHG-2026-02-23-01`
- `CHG-20260223-173953`

Example:
```txt
@change CHG-20260223-01
target: SEC-014
date: 2026-02-23
author: ai-agent
reason: Added explicit numeric rollback thresholds
impact: Runbook and test scenarios need to be synchronized
```

## 5. Recommended Lint Rules

1. `meta.title` and `meta.version` are required.
2. Duplicate IDs in `@map` are not allowed.
3. Duplicate IDs in `@section` are not allowed.
4. Every `@map` ID must exist in `@section`.
5. Every `@section` ID must be listed in `@map`.
6. Every `@change.target` must reference an existing `@section`.
7. `@change.reason` must not be empty.
8. `@change.impact` must not be empty.

## 6. Out of Scope for v0.1

1. Strict ID regex enforcement (`SEC-001`, `REQ-120`, etc.)
2. Per-block custom schemas
3. Section hierarchy (nested structures)
4. JSON/YAML round-trip conversion
