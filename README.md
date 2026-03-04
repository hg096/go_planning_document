# AGD 시작 가이드 (처음 사용자용)

AGD는 **사람과 AI가 같은 문서를 기준으로 협업**하기 위한 문서 포맷입니다.
목표는 단순합니다.

1. 긴 문서를 빠르게 찾고
2. 수정 지점을 정확히 지정하고
3. 변경 이유까지 문서 안에 남기는 것

## 1. 어떤 문제를 해결하나요?

기획 문서가 커질수록 이런 문제가 생깁니다.

- 어디를 고쳐야 할지 찾기 어렵다
- 누가 왜 바꿨는지 추적이 어렵다
- AI에게 수정 요청해도 지점이 애매하다

AGD는 문서를 `@map`(요약 인덱스), `@section`(본문), `@change`(변경 이력)로 분리해서 이 문제를 해결합니다.

### 문서가 늘수록 복잡해지는 한계, 그리고 AGD의 해결 방향

이 프로젝트는 아래 질문에서 시작했습니다.

- "성공한 기획에 모두가 바라봐야 하는 이상적인 지표가 있는가?"
- "문서를 계속 만들고 고치다 보면 파일만 늘고, 관리가 오히려 더 어려워지지 않는가?"
- "문서끼리 서로 다른 로직을 말해 충돌하면, 이 과정을 무한 반복해야 하는가?"

AGD의 답은 "완전한 문서 1개"를 만드는 것이 아니라, **충돌을 빠르게 발견하고 기준 문서로 수렴시키는 운영 체계**를 만드는 것입니다.

해결하려는 핵심 목표:

1. 문서 수가 늘어도 기준이 흔들리지 않게 하기
2. 사람과 AI가 같은 수정 지점을 보게 하기
3. 변경 이유를 끝까지 추적 가능하게 남기기

이를 위한 운영 방안:

- 주제별 기준 문서를 `source`로 1개 지정
- 파생 문서는 `derived`로 연결하고 `source_doc`, `source_sections`로 추적
- `check`를 기본 게이트로 사용해 문서 충돌을 조기에 탐지
- `@map`(요약), `@section`(본문), `@change`(이력)를 항상 함께 갱신
- 주간 루틴(갱신 -> 검증 -> 정리)으로 문서 스프롤 제어

결국 성공 지표는 문서 개수보다 아래에 가깝습니다.

- 충돌 감지까지 걸리는 시간
- 변경 이유를 복원할 수 있는 비율
- 팀/AI가 같은 질문에 같은 답을 내는 일관성

## 2. 5분 시작

### 1) 설치/세팅

- 권장 Go 버전은 `1.25.1`입니다.
- 프로젝트 루트에는 `00_agd/` 폴더만 배치해 사용하세요.
- 초기 세팅 시 setup 스크립트로 훅 경로, 문서 검증, 템플릿 설치를 한 번에 맞추는 것을 권장합니다.
- 세부 세팅 절차는 격리형 패키지 가이드([README.md](/D:/codePack/go/new_read/00_agd/README.md))를 참고하세요.

필수 커맨드(복구):

```cmd
REM git: 현재 프로젝트 루트에 00_agd 폴더만 내려받기
cmd /c "git clone -n --depth 1 --filter=blob:none --sparse https://github.com/hg096/go_planning_document.git .a&&git -C .a sparse-checkout set agd&&git -C .a checkout -q&&xcopy .a\agd 00_agd /e /i /y>nul&&rd /s /q .a"

REM build: 실행 파일 생성
go build -o 00_agd\agd.exe ./cmd/agd
go build -ldflags "-X main.defaultLang=en" -o 00_agd\agd_en.exe ./cmd/agd
go build -ldflags "-X main.defaultLang=ko" -o 00_agd\agd_ko.exe ./cmd/agd

REM exe 실행
00_agd\agd.exe
REM 또는
00_agd\agd.exe wizard
```

### 2) 가장 쉬운 시작: 질문형 모드

실행 후 메뉴에서 번호를 고르고, 질문에 답만 입력하면 됩니다.
문서 경로는 기본적으로 `00_agd\agd_docs`를 기준으로 처리됩니다.
기존 문서를 열 때 파일명만 입력하면 `00_agd\agd_docs` 전체 하위 폴더에서 자동 탐색합니다.
`00_agd\agd_docs` 아래 하위 폴더가 있으면 위자드에서 폴더 선택지가 함께 표시됩니다.
`00_agd/agd_docs/README.md`는 스캐폴드 생성 시 언어(한국어/영어)에 맞춰 자동 생성되며, AI 작성 철학/운영 원칙 가이드가 포함됩니다.

새 문서 생성 시(위자드 첫 화면 메뉴 `4`) 파일명만 입력하면 문서 타입 기준 폴더로 자동 배치됩니다.

- `core-spec` -> `00_agd\agd_docs\10_source\product\<file>.agd`
- `delivery-plan` -> `00_agd\agd_docs\20_derived\frontend\<file>.agd`
- `policy` -> `00_agd\agd_docs\10_source\policy\<file>.agd`
- `meeting` -> `00_agd\agd_docs\30_shared\meeting\<file>.agd`
- `experiment` -> `00_agd\agd_docs\30_shared\experiment\<file>.agd`
- `roadmap` -> `00_agd\agd_docs\30_shared\roadmap\<file>.agd`
- `handoff` -> `00_agd\agd_docs\30_shared\handoff\<file>.agd`

원하는 위치를 직접 지정해 문서를 생성할 수 있습니다.
`delivery-plan`은 프런트 페이지 단위 관리가 쉽도록 `20_derived\frontend` 하위(예: `checkout`, `home`)로 나눠 작성하는 것을 권장합니다.

현재 위자드 메뉴(한국어 빌드 기준):

```txt
[1] 문서 선택
[2] 문서 킷 생성 (스타터/유지보수/신규/오류대응)
[3] 문서 전체 구조 검증하기
[4] 새 문서 만들기
[5] 권위/파생 관계도 보기
[0] 뒤로가기/종료
```

문서 선택 후 문서 작업 메뉴:

```txt
선택한 문서: <문서명>
[1] 키워드 검색하기
[2] 새 섹션 추가 (map도 함께 갱신)
[3] 로직 수정 변경 이력 추가
[4] map만 갱신
[5] 문서 검증하기
[6] Markdown으로 내보내기
[7] 권위/파생 문서 역할 지정
[0] 뒤로가기
```

추가로 위자드의 선택형 화면(문서 타입/킷 프로필/폴더/문서/섹션/역할/출력 형식) 하단에는 공통으로 `[0] 뒤로가기`가 제공됩니다.
`[4] 새 문서 만들기`의 문서 타입 목록은 7개 통합 타입(코어 스펙/전달 계획/회의/실험/로드맵/핸드오프/정책) 순서로 정렬됩니다.

위자드의 수정 작업(메뉴 `3`)에는 가이드드 플로우가 적용됩니다.

- `reason` + `impact` 입력 필수
- 작업 후 자동 `map-sync`
- 이어서 자동 `check`

또한 새 문서 생성(첫 화면 메뉴 `4`) 직후에는 `source/derived/later` 역할 선택 단계를 거쳐, 기준 문서 운영을 빠르게 시작할 수 있습니다.
새 문서 생성 시 폴더 선택은 문서 타입별 허용 경로(및 해당 경로 하위 폴더)로 제한됩니다.
위자드(첫 화면 메뉴 `5`)의 기본 문서 타입은 문서 수 절감을 위해 아래 7개로 축소되어 있습니다.

- `core-spec`
- `delivery-plan`
- `meeting`
- `experiment`
- `roadmap`
- `handoff`
- `policy`

문서 생성 기능은 위 7개 통합 타입 중심으로 운영됩니다.

## 3. 핵심 개념 (이것만 기억하면 됩니다)

- `@meta`: 문서 정보 (`title`, `version`, `owner`)
- `@map`: 문서의 한 페이지 요약 인덱스
- `@section <ID>`: 실제 본문 단위
- `@change <ID>`: 무엇을 왜 바꿨는지 기록

좋은 AGD 문서는 **`@map`만 읽어도 전체 구조가 보이고**, `@section`으로 들어가면 상세를 확인할 수 있어야 합니다.

## 4. 기획을 잘 관리하는 운영법

### 원칙 1) 문서를 "섹션 단위"로 관리

- 기능/정책/지표를 섹션 ID로 분리하세요.
- 예: `PRD-010`, `PRD-020`, `RUN-040`
- 한 섹션에는 한 주제만 담으세요.

### 원칙 2) 항상 `@map`을 먼저 고치고 본문을 고치기

- 요약(`@map`)과 본문(`@section`)이 어긋나면 문서가 빠르게 망가집니다.
- 작업 순서: `@map` 업데이트 -> `@section` 수정 -> 문서 정합성 확인
- 새 섹션 추가 시에는 `@map`이 함께 갱신되도록 유지하세요.
- 기존 문서의 map이 어긋났다면 map 동기화를 먼저 수행하세요.

### 원칙 3) 변경은 반드시 이유와 함께 남기기

- 문서 수정이 발생하면 `@change`의 `reason`, `impact`를 **반드시** 남기세요.
- 나중에 "왜 이렇게 됐는지"를 복원할 수 있습니다.
- `@change` ID는 **문서 내에서 유일**하면 됩니다.
- 권장 패턴은 아래 두 가지 중 하나를 사용하세요.
- `CHG-2026-02-23-01` (가독성 중심)
- `CHG-20260223-173953` (시간 기반 고유성 중심)

### 원칙 4) 권위 문서(source)와 파생 문서(derived) 분리

- 한 주제에는 기준이 되는 `source` 문서를 1개만 두세요.
- 파생 문서는 `derived`로 표시하고 기준 문서를 연결하세요.
- 검증 단계에서 `source_doc`와 섹션 `title/summary` 충돌을 자동 확인합니다.
- `source_sections` 매핑 규칙:
- `SRC->DST`: 연결/존재 검증
- `SRC=>DST`: 강한 충돌 검증(제목/요약 비교)
- `SEC-ID`: 같은 ID끼리 강한 충돌 검증
- `auto`: 추천 매핑(`->`) 자동 생성
- `strict-auto`: 추천 매핑을 엄격 매핑(`=>`)으로 자동 생성
- `smart-auto`: 엄격 매핑을 시도하되, 제목/요약 불일치 항목은 `->`로 자동 완화
### 원칙 5) 주간 운영 루틴 만들기

추천 루틴:

1. 월요일: `@map` 기준 우선순위 정리
2. 수시: `edit`/`add`로 섹션 단위 변경
3. 금요일: 전체 검증과 관계도 리뷰 공유

### 원칙 6) Done 기준을 문서에 포함

각 섹션에 아래 항목을 최소로 넣으면 운영이 쉬워집니다.

- 목표 지표 (숫자)
- 담당자/팀
- 완료 기준
- 관련 링크 (`links:`)

### 원칙 7) 실패사례 분석 프레임을 운영 루틴에 넣기

- 실패는 기능명이 아니라 메커니즘(문제정의/의존성/실행가능성/의사결정/학습루프)으로 분류하세요.
- incident-case(30_shared/errFix) 대응 완료 후 필요 시 postmortem을 작성하고 source 반영 증거를 남기세요.
- 주간 리뷰에서 문서 검증 실패 원인 Top3를 집계해 재발 방지 액션으로 연결하세요.
- 릴리즈 전 `@change`의 `reason`, `impact` 누락 건수를 0으로 맞추세요.
- 상세 규약: `00_agd/docs/AGD_TEMPLATE_GUIDE_ko.md`

### 원칙 8) 서비스 로직 변경은 문서 우선

- 서비스 로직 코드 수정 전, 서비스 로직 문서(`10_source/service`)를 먼저 갱신하세요.
- `@change`의 `reason`/`impact`를 누락 없이 남긴 뒤 구현 작업을 진행하세요.
- 변경 후 문서 전체 검증으로 정합성을 확인하세요.

## 5. AI에게 요청할 때 잘 되는 방식

좋은 요청은 **ID를 명시**합니다.

좋은 예:

- "`PRD-020`에 실패 복구 흐름 3단계 추가해줘"
- "`RUN-030` 요약을 1문장으로 압축해줘"
- "`SEC-014` 수정하고 변경 이력(`@change`)도 남겨줘"

나쁜 예:

- "문서 좀 고쳐줘"
- "어딘가 업데이트해줘"

## 6. 템플릿 선택 가이드

템플릿 목록 기능은 7개 통합 타입의 한글/영문 템플릿만 표시합니다.

- 한글 템플릿: `*-ko` (예: `core-spec-ko`)
- 영문 템플릿: `*-en` (예: `core-spec-en`)

- `core-spec-ko`, `core-spec-en`: 통합 코어 스펙 (PRD+로직+ADR+AI 가이드)
- `delivery-plan-ko`, `delivery-plan-en`: 통합 전달 계획 (프론트+QA+릴리즈)
- `meeting-ko`, `meeting-en`: 회의록 및 의사결정 기록
- `experiment-ko`, `experiment-en`: 실험 및 A/B 테스트 계획
- `roadmap-ko`, `roadmap-en`: 분기/반기 로드맵 계획
- `handoff-ko`, `handoff-en`: 팀 간 인수인계 문서
- `policy-ko`, `policy-en`: 정책 및 가이드 문서


## 8. 바로 열어볼 예시 문서

`examples` 폴더는 아래 구조로 정리되어 있습니다.

- `examples/ko/10_source/*`: 한국어 기준(source) 예시
- `examples/ko/20_derived/*`: 한국어 파생(derived) 예시
- `examples/ko/30_shared/*`: 한국어 공유(shared) 예시
- `examples/en/10_source/*`: 영어 기준(source) 예시
- `examples/en/20_derived/*`: 영어 파생(derived) 예시
- `examples/en/30_shared/*`: 영어 공유(shared) 예시

한국어 source 예시:

- `examples/ko/10_source/product/prd_subscription_checkout_ko.agd`
- `examples/ko/10_source/service/service_logic_checkout_core_ko.agd`
- `examples/ko/10_source/policy/policy_release_gate_ko.agd`
- `examples/ko/10_source/architecture/adr_checkout_retry_policy_ko.agd`

한국어 derived/shared 예시:

- `examples/ko/20_derived/frontend/frontend_page_checkout_ko.agd`
- `examples/ko/20_derived/qa/qa_plan_checkout_v2_ko.agd`
- `examples/ko/20_derived/ops/runbook_payment_incident_ko.agd`
- `examples/ko/30_shared/roadmap/roadmap_ai_doc_platform_ko.agd`
- `examples/ko/30_shared/postmortem/postmortem_api_timeout_ko.agd`
- `examples/ko/30_shared/handoff/handoff_backend_frontend_ko.agd`
- `examples/ko/30_shared/meeting/meeting_checkout_weekly_ko.agd`
- `examples/ko/30_shared/experiment/experiment_oneclick_checkout_ko.agd`

유지보수/장애 단일 케이스는 examples 고정 파일 대신 킷 생성 경로를 사용합니다.

- `00_agd/agd_docs/30_shared/maintenance/<project>_maintenance_case.agd`
- `00_agd/agd_docs/30_shared/errFix/<project>_incident_case.agd`

템플릿별 활용 예시 인덱스:

- `examples/TEMPLATE_SHOWCASE_INDEX_ko.md`
- `examples/TEMPLATE_SHOWCASE_INDEX_en.md`
- `examples/README.md`

## 9. 운영 강제 세팅 (Git Hook + CI)

기존 프로젝트와 충돌을 줄이기 위해 AGD는 `agd/` 폴더 안에서만 동작하도록 분리합니다.
호스트 프로젝트 루트에는 훅/워크플로를 직접 고정 생성하지 않고, 필요 시 템플릿 설치 방식으로 적용합니다.

### 9-1) 로컬 pre-commit 훅 설정

1) AGD 폴더 기준 세팅 스크립트를 실행해 로컬 훅을 활성화합니다.

2) 저장소 포함 훅 파일:

- `agd/.githooks/pre-commit`
- `agd/scripts/git-hooks/pre-commit.ps1`

3) 커밋 시 자동 강제 규칙:

- `agd/` 하위에서 변경된 `.agd` 파일만 검사
- `00_agd\agd_docs` 전체 strict 검증
- `00_agd\examples` 변경 시 examples strict 검증

### 9-2) GitHub Actions 머지 게이트

기본 제공 위치(템플릿):

- `agd/templates/agd-guard.yml`

기본 세팅 시 자동 설치됩니다.

CI 템플릿은 `.github/workflows/agd-guard.yml`로 복사됩니다.

### 9-3) PR 체크리스트 강제

기본 제공 위치(템플릿):

- `agd/templates/pull_request_template.md`

기본 세팅 시 자동 설치됩니다.

PR 템플릿은 `.github/pull_request_template.md`로 복사됩니다.

## 10. 참고 문서

- AI 실행 가이드(국문): `00_agd/docs/AGD_TEMPLATE_GUIDE_ko.md`
- AI Execution Guide (English): `00_agd/docs/AGD_TEMPLATE_GUIDE_en.md`
- Start Guide (English): `README.en.md`
- 문서 루트 구조 가이드: `00_agd/agd_docs/README.md`
- 격리형 패키지 가이드: `agd/README.md`
- 원클릭 부트스트랩 스크립트: `agd/setup.cmd`
- 부트스트랩 스크립트(PowerShell): `agd/scripts/setup.ps1`
- 게이트 강제 실행 래퍼(cmd): `run-safe.cmd`
- 코어 로직: `internal/agd/*`
- 템플릿 소스: `internal/agdtemplates/*`

