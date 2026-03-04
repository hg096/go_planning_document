# AGD Package (Isolated Mode)

For Korean, see `agd/README.md`.

This package is designed to be dropped into an existing repository with minimal conflict.
All AGD assets are managed under the `agd/` folder.

## 1) What is inside

- executables: `agd/agd.exe`, `agd/agd_en.exe`, `agd/agd_ko.exe`
- docs: `agd/docs/*`
- examples: `agd/examples/*`
- working docs root: `agd/agd_docs/*`
- setup entry: `agd/setup.cmd`
- local hook: `agd/.githooks/pre-commit`
- hook logic: `agd/scripts/git-hooks/pre-commit.ps1`
- optional CI/PR templates: `agd/templates/*`

No `.go` source files are included in `agd/`.

## 2) One-command setup (minimal, no source tree)

Single `cmd` line to fetch only `agd/` and run setup:

```cmd
cmd /c "git clone --depth 1 --filter=blob:none --no-checkout <repo-url> <repo-folder> && cd /d <repo-folder> && git sparse-checkout init --no-cone && git sparse-checkout set /agd/ && git checkout && agd\setup.cmd"
```

If you already cloned the full repository:

```cmd
agd\setup.cmd -SlimCheckout
```

`-SlimCheckout` keeps only `agd/` in this repository clone.

What setup does:

- sets `git config core.hooksPath agd/.githooks`
- validates `agd/agd_docs` and `agd/examples` (strict mode)
- installs CI workflow and PR template (backs up existing files before overwrite)

Options:

```cmd
agd\setup.cmd -SkipCheck
agd\setup.cmd -SkipTemplates
agd\setup.cmd -InstallCiTemplate
agd\setup.cmd -InstallPrTemplate
agd\setup.cmd -NoTemplateBackup
agd\setup.cmd -SlimCheckout
```

## 3) Daily usage

```cmd
agd\agd_en.exe quick
agd\agd_en.exe wizard
agd\agd_en.exe check-all agd\agd_docs --strict
```

## 4) Optional CI integration

Install template files into host repo:

```cmd
agd\setup.cmd -InstallCiTemplate -InstallPrTemplate
```

This copies:

- `agd/templates/agd-guard.yml` -> `.github/workflows/agd-guard.yml`
- `agd/templates/pull_request_template.md` -> `.github/pull_request_template.md`
