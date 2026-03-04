# AGD 패키지 (격리 모드)

영문 문서는 `agd/README.en.md`를 참고하세요.

이 패키지는 기존 진행 중인 저장소에 충돌을 최소화하며 추가할 수 있도록 설계되었습니다.
AGD 관련 자산은 모두 `agd/` 폴더 내부에서 관리합니다.

## 1) 포함 구성

- 실행 파일: `agd/agd.exe`, `agd/agd_en.exe`, `agd/agd_ko.exe`
- 문서: `agd/docs/*`
- 예시: `agd/examples/*`
- 작업 문서 루트: `00_agd/agd_docs/*`
- 세팅 진입점: `agd/setup.cmd`
- 로컬 훅: `agd/.githooks/pre-commit`
- 훅 로직: `agd/scripts/git-hooks/pre-commit.ps1`
- 선택 템플릿(CI/PR): `agd/templates/*`

`agd/` 폴더에는 `.go` 소스 파일을 포함하지 않습니다.

## 2) 원클릭 세팅 (소스 제외 최소 설치)

`cmd` 한 줄로 프로젝트 루트에 `00_agd/`만 내려받기:

```cmd
cmd /c "git clone --depth 1 --filter=blob:none --no-checkout <repo-url> .agd_tmp && git -C .agd_tmp sparse-checkout init --no-cone && git -C .agd_tmp sparse-checkout set /agd/ && git -C .agd_tmp checkout --quiet && xcopy .agd_tmp\agd 00_agd /E /I /Y >nul && rmdir /s /q .agd_tmp"
```

운영 훅/검증까지 적용하려면:

```cmd
00_agd\setup.cmd -SkipTemplates
```

자동 수행 항목:

- `git config core.hooksPath 00_agd/.githooks` 설정
- `00_agd/agd_docs`, `00_agd/examples` strict 검증 실행
- CI 워크플로/PR 템플릿 설치(기존 파일이 있으면 백업 후 갱신)

옵션:

```cmd
00_agd\setup.cmd -SkipCheck
00_agd\setup.cmd -SkipTemplates
00_agd\setup.cmd -InstallCiTemplate
00_agd\setup.cmd -InstallPrTemplate
00_agd\setup.cmd -NoTemplateBackup
00_agd\setup.cmd -SlimCheckout
```

## 3) 일상 사용

```cmd
00_agd\agd_en.exe quick
00_agd\agd_en.exe wizard
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

## 4) 선택 CI 연동

호스트 저장소에 템플릿 설치:

```cmd
00_agd\setup.cmd -InstallCiTemplate -InstallPrTemplate
```

복사 대상:

- `agd/templates/agd-guard.yml` -> `.github/workflows/agd-guard.yml`
- `agd/templates/pull_request_template.md` -> `.github/pull_request_template.md`
