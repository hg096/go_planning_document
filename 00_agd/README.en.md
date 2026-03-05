# AGD Package (Isolated Mode)

For Korean, see `00_agd/README.md`.

This package is designed to be dropped into an existing repository with minimal conflict.
All AGD assets are managed under the `00_agd/` folder.

## 1) What is inside

- executables: `00_agd/agd.exe`, `00_agd/agd_en.exe`, `00_agd/agd_ko.exe`
- docs: `00_agd/docs/*`
- examples: `00_agd/examples/*`
- working docs root: `00_agd/agd_docs/*`
- setup entry: `00_agd/setup.cmd`
- local hook: `00_agd/.githooks/pre-commit`
- hook logic: `00_agd/scripts/git-hooks/pre-commit.ps1`
- optional CI/PR templates: `00_agd/templates/*`

No `.go` source files are included in `00_agd/`.

## 2) One-command setup (minimal, no source tree)

Single `cmd` line to fetch only `00_agd/` into project root:

```cmd
cmd /v:on /c "set T=%TEMP%\agd_tmp_%RANDOM%%RANDOM%&&(git clone -n --depth 1 --filter=blob:none --sparse https://github.com/hg096/go_planning_document.git "!T!" && git -C "!T!" sparse-checkout set 00_agd && git -C "!T!" checkout -q && xcopy "!T!\00_agd" "00_agd" /e /i /y >nul) & set EC=!ERRORLEVEL! & if exist "!T!" rd /s /q "!T!" & exit /b !EC!"
```

To apply hooks/validation after download:

```cmd
00_agd\setup.cmd
```

What setup does:

- sets `git config core.hooksPath 00_agd/.githooks`
- validates `00_agd/agd_docs` and `00_agd/examples` (strict mode)
- installs CI workflow and PR template (backs up existing files before overwrite)

Options:

```cmd
00_agd\setup.cmd -SkipCheck
00_agd\setup.cmd -SkipTemplates
00_agd\setup.cmd -InstallCiTemplate
00_agd\setup.cmd -InstallPrTemplate
00_agd\setup.cmd -NoTemplateBackup
00_agd\setup.cmd -SlimCheckout
```

## 2-1) Enforcement Setup (Local + CI)

Use the following flow to make document checks operationally strict.

1. Run setup

```cmd
00_agd\setup.cmd
```

2. Verify hooks path (`core.hooksPath`)

```cmd
git config --get core.hooksPath
```

Expected:

```txt
00_agd/.githooks
```

3. Run local baseline checks

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

4. Install CI templates

```cmd
00_agd\setup.cmd -InstallCiTemplate -InstallPrTemplate
```

Then verify `.github/workflows/agd-guard.yml` uses real paths:

- expected: `./00_agd/agd_en.exe`, `00_agd/agd_docs`, `00_agd/examples`
- if the workflow still uses `./agd/...`, replace it with `./00_agd/...`

5. Enable branch protection

- protect `main`
- add `AGD Guard` to Required status checks

Enforcement note:

- local hooks can be bypassed via `git commit --no-verify`
- final enforcement must rely on Required CI checks

## 2-2) Core Logic Change Notice (Soft Guard)

Core logic policy file:

- `00_agd/policy/core_logic_paths.txt`

Path format rules:

- repo-root relative paths
- `#` comments allowed
- empty lines ignored
- both `/` and `\` separators accepted (normalized by hook)

Commit hook behavior:

- If staged files match policy paths, it prints a warning banner and matched file list.
- In interactive terminals, it asks `Continue commit? (y/N)` and proceeds only on `y/yes`.
- In non-interactive environments (CI/automation), it prints warning only and does not block.

## 2-3) Code Planning Validation (Code Plan Check)

Use this command to detect code-change vs planning-doc gaps.

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe code-plan-check --mode git
00_agd\agd_en.exe code-plan-check --mode cache
00_agd\agd_en.exe code-plan-check --mode git --git-source staged --strict-relation
```

Behavior summary:

- Collect changed code files (`auto`: git first, then cache fallback).
- Check overlap against AGD doc `doc_base_path` and `section.path`.
- Matched docs pass only when either maintenance/incident tag exists or `@change` is updated.
- If relation policy (`A#SEC <-> B#SEC`) exists, bidirectional service/frontend linked docs are additionally checked for `@change` updates.
- On failure, it prints remediation steps and exits with code `1`.

Policy/cache paths:

- scope policy: `00_agd/policy/code_plan_scope.txt`
- relation policy: `00_agd/policy/code_plan_relations.txt`
- local cache: `00_agd/.cache/code_plan_validation.json`

Relation policy syntax:

- `<left-doc.agd>#<SECTION-ID> <-> <right-doc.agd>#<SECTION-ID>`
- one relation per line
- for multiple relations, add multiple lines
- for 1:N, repeat the left endpoint; for N:1, repeat the right endpoint
- comma-separated multi-endpoint syntax in one line is not supported
- example: `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010`
- example (1:N):
  - `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010`
  - `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-020`

Scope policy format:

- `ext:.vue` (add extension)
- `include:src/generated/` (force include path)
- `exclude:vendor/` (exclude path)

Optional strict mode:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto --strict-mapping
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

- Use `--strict-mapping` only when you want unmatched doc mapping to fail immediately.
- Use `--strict-relation` when missing linked-doc updates must fail.
- pre-commit runs relation strict mode with staged-only source:
  `code-plan-check --mode git --git-source staged --strict-relation`.

## 3) Daily usage

```cmd
00_agd\agd_en.exe quick
00_agd\agd_en.exe wizard
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

## 4) Optional CI integration

Install template files into host repo:

```cmd
00_agd\setup.cmd -InstallCiTemplate -InstallPrTemplate
```

This copies:

- `00_agd/templates/agd-guard.yml` -> `.github/workflows/agd-guard.yml`
- `00_agd/templates/pull_request_template.md` -> `.github/pull_request_template.md`

Verification checklist:

- `.github/workflows/agd-guard.yml` uses `00_agd` paths.
- `AGD Guard` appears in PR checks.
- branch protection marks `AGD Guard` as required.

