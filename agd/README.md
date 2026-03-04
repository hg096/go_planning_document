# AGD 패키지 (격리 모드)

영문 문서는 `agd/README.en.md`를 참고하세요.

이 패키지는 기존 진행 중인 저장소에 충돌을 최소화하며 추가할 수 있도록 설계되었습니다.
AGD 관련 자산은 모두 `agd/` 폴더 내부에서 관리합니다.

## 1) 포함 구성

- 실행 파일: `agd/agd.exe`, `agd/agd_en.exe`, `agd/agd_ko.exe`
- 문서: `agd/docs/*`
- 예시: `agd/examples/*`
- 작업 문서 루트: `agd/agd_docs/*`
- 세팅 진입점: `agd/setup.cmd`
- 로컬 훅: `agd/.githooks/pre-commit`
- 훅 로직: `agd/scripts/git-hooks/pre-commit.ps1`
- 선택 템플릿(CI/PR): `agd/templates/*`

`agd/` 폴더에는 `.go` 소스 파일을 포함하지 않습니다.

## 2) 원클릭 세팅 (소스 제외 최소 설치)

`cmd` 한 줄로 `agd/`만 내려받아 세팅:

```cmd
cmd /c "git clone --depth 1 --filter=blob:none --no-checkout <repo-url> <repo-folder> && cd /d <repo-folder> && git sparse-checkout init --no-cone && git sparse-checkout set /agd/ && git checkout && agd\setup.cmd"
```

이미 전체 저장소를 클론했다면:

```cmd
agd\setup.cmd -SlimCheckout
```

`-SlimCheckout`은 이 저장소 클론에서 `agd/`만 남기고 작업 트리를 축소합니다.

자동 수행 항목:

- `git config core.hooksPath agd/.githooks` 설정
- `agd/agd_docs`, `agd/examples` strict 검증 실행
- CI 워크플로/PR 템플릿 설치(기존 파일이 있으면 백업 후 갱신)

옵션:

```cmd
agd\setup.cmd -SkipCheck
agd\setup.cmd -SkipTemplates
agd\setup.cmd -InstallCiTemplate
agd\setup.cmd -InstallPrTemplate
agd\setup.cmd -NoTemplateBackup
agd\setup.cmd -SlimCheckout
```

## 3) 일상 사용

```cmd
agd\agd_en.exe quick
agd\agd_en.exe wizard
agd\agd_en.exe check-all agd\agd_docs --strict
```

## 4) 선택 CI 연동

호스트 저장소에 템플릿 설치:

```cmd
agd\setup.cmd -InstallCiTemplate -InstallPrTemplate
```

복사 대상:

- `agd/templates/agd-guard.yml` -> `.github/workflows/agd-guard.yml`
- `agd/templates/pull_request_template.md` -> `.github/pull_request_template.md`
