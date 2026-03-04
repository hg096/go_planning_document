# AGD 패키지 (격리 모드)

영문 문서는 `00_agd/README.en.md`를 참고하세요.

이 패키지는 기존 진행 중인 저장소에 충돌을 최소화하며 추가할 수 있도록 설계되었습니다.
AGD 관련 자산은 모두 `00_agd/` 폴더 내부에서 관리합니다.

## 1) 포함 구성

- 실행 파일: `00_agd/agd.exe`, `00_agd/agd_en.exe`, `00_agd/agd_ko.exe`
- 문서: `00_agd/docs/*`
- 예시: `00_agd/examples/*`
- 작업 문서 루트: `00_agd/agd_docs/*`
- 세팅 진입점: `00_agd/setup.cmd`
- 로컬 훅: `00_agd/.githooks/pre-commit`
- 훅 로직: `00_agd/scripts/git-hooks/pre-commit.ps1`
- 선택 템플릿(CI/PR): `00_agd/templates/*`

`00_agd/` 폴더에는 `.go` 소스 파일을 포함하지 않습니다.

## 2) 원클릭 세팅 (소스 제외 최소 설치)

`cmd` 한 줄로 프로젝트 루트에 `00_agd/`만 내려받기:

```cmd
cmd /v:on /c "set T=%TEMP%\agd_tmp_%RANDOM%%RANDOM%&&(git clone -n --depth 1 --filter=blob:none --sparse https://github.com/hg096/go_planning_document.git "!T!" && git -C "!T!" sparse-checkout set 00_agd && git -C "!T!" checkout -q && xcopy "!T!\00_agd" "00_agd" /e /i /y >nul) & set EC=!ERRORLEVEL! & if exist "!T!" rd /s /q "!T!" & exit /b !EC!"
```

운영 훅/검증까지 적용하려면:

```cmd
00_agd\setup.cmd
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

## 2-1) 강제 운영 세팅 (로컬+CI)

아래 순서로 설정하면 문서 누락을 실무에서 강하게 줄일 수 있습니다.

1. 초기 세팅 실행

```cmd
00_agd\setup.cmd
```

2. 훅 적용 확인 (`core.hooksPath`)

```cmd
git config --get core.hooksPath
```

기대값:

```txt
00_agd/.githooks
```

3. 로컬 검증 기본 명령 실행

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

4. CI 템플릿 설치

```cmd
00_agd\setup.cmd -InstallCiTemplate -InstallPrTemplate
```

설치 후 `.github/workflows/agd-guard.yml`에서 실행 경로가 실제 구조와 일치하는지 확인하세요.

- 기대 경로: `./00_agd/agd_en.exe`, `00_agd/agd_docs`, `00_agd/examples`
- 템플릿에 `./agd/...`가 남아 있으면 `./00_agd/...`로 보정

5. 브랜치 보호 설정

- `main` 브랜치 보호 활성화
- Required status checks에 `AGD Guard` 추가

강제력 한계/권장:

- 로컬 훅은 `git commit --no-verify`로 우회 가능합니다.
- 최종 강제는 반드시 GitHub Required CI check로 설정하세요.

## 2-2) 핵심 로직 수정 주의 (Soft Guard)

핵심 로직 경로 정책 파일:

- `00_agd/policy/core_logic_paths.txt`

경로 포맷 규칙:

- 저장소 루트 기준 상대 경로
- `#` 주석 허용
- 빈 줄 무시
- `/`, `\` 경로 구분자 혼용 허용(훅에서 정규화)

커밋 훅 동작:

- staged 변경 파일이 정책 경로와 일치하면 경고 배너와 파일 목록을 출력합니다.
- 대화형 터미널에서는 `Continue commit? (y/N)` 확인을 받고 `y/yes`일 때만 진행합니다.
- 비대화형 환경(CI/자동화)에서는 경고만 출력하고 차단하지 않습니다.

## 2-3) 코드 기획 검증 (Code Plan Check)

코드 변경과 문서 누락을 함께 점검하는 명령입니다.

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe code-plan-check --mode git
00_agd\agd_en.exe code-plan-check --mode cache
```

동작 요약:

- 변경 코드 파일을 수집합니다. (`auto`: git 우선, 실패 시 cache fallback)
- `doc_base_path`, `section.path`와 코드 경로 겹침을 검사합니다.
- 매칭 문서는 `유지보수/오류 tag` 또는 `@change 업데이트` 중 하나를 만족해야 통과합니다.
- 실패 시 누락 원인과 수정 가이드를 출력하고 종료 코드 `1`을 반환합니다.

정책/캐시 경로:

- 범위 정책: `00_agd/policy/code_plan_scope.txt`
- 로컬 캐시: `00_agd/.cache/code_plan_validation.json`

범위 정책 형식:

- `ext:.vue` (확장자 추가)
- `include:src/generated/` (강제 포함)
- `exclude:vendor/` (제외)

선택 강화 옵션:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto --strict-mapping
```

- `--strict-mapping`은 "매칭 문서 미검출"도 실패로 처리하고 싶을 때만 사용하세요.

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

- `00_agd/templates/agd-guard.yml` -> `.github/workflows/agd-guard.yml`
- `00_agd/templates/pull_request_template.md` -> `.github/pull_request_template.md`

검증 체크리스트:

- `.github/workflows/agd-guard.yml`가 `00_agd` 경로를 사용한다.
- `AGD Guard`가 PR 상태 체크로 표시된다.
- 브랜치 보호에서 `AGD Guard`가 Required로 설정되어 있다.
