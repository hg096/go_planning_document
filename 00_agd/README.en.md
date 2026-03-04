# AGD Package (Isolated Mode)

For Korean, see `agd/README.md`.

This package is designed to be dropped into an existing repository with minimal conflict.
All AGD assets are managed under the `agd/` folder.

## 1) What is inside

- executables: `agd/agd.exe`, `agd/agd_en.exe`, `agd/agd_ko.exe`
- docs: `agd/docs/*`
- examples: `agd/examples/*`
- working docs root: `00_agd/agd_docs/*`
- setup entry: `agd/setup.cmd`
- local hook: `agd/.githooks/pre-commit`
- hook logic: `agd/scripts/git-hooks/pre-commit.ps1`
- optional CI/PR templates: `agd/templates/*`

No `.go` source files are included in `agd/`.

## 2) One-command setup (minimal, no source tree)

Single `cmd` line to fetch only `00_agd/` into project root:

```cmd
cmd /c "git clone --depth 1 --filter=blob:none --no-checkout <repo-url> .agd_tmp && git -C .agd_tmp sparse-checkout init --no-cone && git -C .agd_tmp sparse-checkout set /agd/ && git -C .agd_tmp checkout --quiet && xcopy .agd_tmp\agd 00_agd /E /I /Y >nul && rmdir /s /q .agd_tmp"
```

To apply hooks/validation setup after download:

```cmd
00_agd\setup.cmd -SkipTemplates
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

- `agd/templates/agd-guard.yml` -> `.github/workflows/agd-guard.yml`
- `agd/templates/pull_request_template.md` -> `.github/pull_request_template.md`
