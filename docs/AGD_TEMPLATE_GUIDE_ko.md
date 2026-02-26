# AGD 템플릿 가이드 (최신)

이 문서는 `agd.exe`를 처음 쓰는 사용자 기준으로 템플릿 생성과 운영 절차를 설명합니다.
기본 문서 루트는 `agd_docs`입니다.

## 0) 가장 쉬운 시작: 위자드 (`agd.exe`)

```cmd
agd.exe
REM 또는
agd.exe wizard
```

한국어 빌드(`agd_ko.exe`) 기준 1차 메뉴:

```txt
[1] 문서 선택
[2] 문서 킷 생성 (starter/maintenance/new/incident)
[3] 문서 전체 구조 검증하기
[4] 새 문서 만들기
[5] 권위/파생 관계도 보기
[0] 뒤로가기/종료
```

문서 선택 후 2차 메뉴:

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

선택형 화면(문서 타입/킷 프로필/폴더/문서/섹션/역할/출력 형식)에서는 공통으로 `[0] 뒤로가기`를 지원합니다.

위자드 `새 문서 만들기` 기본 타입(문서 수 절감용):

- `core-spec`
- `delivery-plan`
- `meeting`
- `experiment`
- `roadmap`
- `handoff`
- `policy`

CLI 생성(`agd.exe new`, `agd.exe init`)도 위 7개 통합 타입으로 제한됩니다.

## 1) 빠른 생성: `agd new`

```cmd
agd.exe new list
agd.exe new core-spec core_spec_checkout_ko "체크아웃 코어 스펙" "product-team"
```

`agd new`에서 쓰는 문서 타입:

- `core-spec`: 통합 코어 스펙 (PRD+로직+ADR+AI 가이드)
- `delivery-plan`: 통합 전달 계획 (프론트+QA+릴리즈)
- `meeting`: 회의록/결정사항 문서 (Meeting notes and decisions)
- `experiment`: 실험/AB 테스트 계획 (Experiment and A/B test plan)
- `roadmap`: 분기/반기 로드맵 문서 (Quarter/half roadmap plan)
- `handoff`: 팀 간 연동 핸드오프 문서 (Cross-team handoff document)
- `policy`: 정책/가이드 문서 (Policy and guideline document)

## 2) 상세 생성: `agd init`

```cmd
agd.exe init --list
agd.exe init --template core-spec-ko --out docs/core_spec_checkout_ko.agd --title "체크아웃 코어 스펙" --owner "product-team"
agd.exe init --template core-spec-en --out docs/core_spec_checkout_en.agd --title "Checkout Core Spec" --owner "product-team"
```

주요 옵션:

- `--template`: 템플릿 이름 (`core-spec-ko`, `core-spec-en`, `delivery-plan-ko`, ...)
- `--out`: 생성할 파일 경로
- `--title`: 문서 제목 (`@meta.title`)
- `--owner`: 문서 소유 팀/담당 (`@meta.owner`)
- `--version`: 문서 버전 (`@meta.version`, 기본 `0.1`)
- `--force`: 기존 파일 덮어쓰기

참고:

- `agd.exe init --list`를 실행하면 `*-ko`, `*-en` 템플릿이 모두 출력됩니다.

## 3) 경로 입력 규칙 (`agd_docs` 기본)

기존 문서를 열 때 파일명만 입력해도 `agd_docs` 하위 전체에서 자동 탐색합니다.

- 입력: `core_spec_checkout_ko`
- 탐색 대상: `agd_docs\**\core_spec_checkout_ko.agd`

동일 파일명이 여러 하위 폴더에 있으면 다중 매칭 오류를 반환합니다.
이 경우 `front/home`처럼 하위 경로를 함께 입력하세요.

새 문서 생성(`agd new`, 위자드 1차 메뉴 `4`)에서 파일명만 넣으면 문서 타입 기준 폴더로 자동 배치됩니다.

- `core-spec`: `agd_docs\10_source\product\<file>.agd`
- `delivery-plan`: `agd_docs\10_source\product\<file>.agd`
- `policy`: `agd_docs\10_source\policy\<file>.agd`
- `meeting`: `agd_docs\30_shared\meeting\<file>.agd`
- `experiment`: `agd_docs\30_shared\experiment\<file>.agd`
- `roadmap`: `agd_docs\30_shared\roadmap\<file>.agd`
- `handoff`: `agd_docs\30_shared\handoff\<file>.agd`

원하는 위치를 직접 지정하려면 경로를 함께 입력하세요.

`END__` 종료 처리 규칙:

- `30_shared/maintenance/END__*_maintenance_case.agd`
- `30_shared/errFix/END__*_incident_case.agd`

위 파일은 스캔/선택/관계도 출력에서 자동 제외됩니다.

## 4) 섹션/맵 관리 명령

```cmd
REM 섹션 추가 + map 자동 갱신
agd.exe section-add <file.agd> <SECTION-ID> "<섹션 제목>" "<요약>" [links] [content] --reason "<변경 이유>" --impact "<변경 영향>"

REM section 기준으로 map만 재동기화
agd.exe map-sync <file.agd> --reason "<동기화 사유>" --impact "<동기화 영향>"

REM 섹션 요약 수정
agd.exe edit <file.agd> <SECTION-ID> "<새 요약>" --reason "<변경 이유>" --impact "<변경 영향>"

REM 본문 한 줄 추가
agd.exe add <file.agd> <SECTION-ID> "<추가할 문장>" --reason "<변경 이유>" --impact "<변경 영향>"
```

문서 수정이 발생하면 `@change(reason/impact)` 기록이 필수입니다.
`edit`/`add`/`section-add` 실행 시 `map-sync -> check`가 자동으로 이어서 수행됩니다.

문서 묶음 전체 검증:

```cmd
REM 전체 문서 트리 검사 (기본: agd_docs)
agd.exe check-all
agd.exe check-all examples

REM 경고도 실패로 처리
agd.exe check-all examples --strict

REM archive 문서까지 포함
agd.exe check-all --include-archive
```

서비스 로직 변경 시 권장 순서:

```cmd
REM 서비스 문서 먼저 갱신
agd.exe check 10_source/service/<service_doc>

REM 전체 정합성 검증
agd.exe check-all
```

## 5) 권위/파생 문서 운영

```cmd
REM 기준 문서(권위 문서)
agd.exe role-set service_overview source

REM 파생 문서 지정
agd.exe role-set frontend_pages derived service_overview "SYS-020->FP-020,SYS-030=>FP-030"

REM 매핑 자동 생성
agd.exe role-set frontend_pages derived service_overview auto
agd.exe role-set frontend_pages derived service_overview strict-auto
agd.exe role-set frontend_pages derived service_overview smart-auto

REM 추천 매핑만 조회
agd.exe map-suggest frontend_pages service_overview
agd.exe map-suggest frontend_pages service_overview strict-auto
agd.exe map-suggest frontend_pages service_overview smart-auto

REM 검증
agd.exe check frontend_pages
```

`source_sections` 규칙:

- `SRC->DST`: 연결/존재 검증
- `SRC=>DST`: 제목/요약 강한 충돌 검증
- `SEC-ID`: 같은 ID끼리 강한 충돌 검증
- `auto`: 추천 매핑을 `->`로 자동 생성
- `strict-auto`: 추천 매핑을 `=>`로 자동 생성
- `smart-auto`: 엄격 매핑 시도 후, 불일치 항목만 `->`로 자동 완화

## 6) 변경 이력(`@change`) 운영

```cmd
agd.exe logic-log <file.agd> <SECTION-ID> --reason "<변경 사유>" --impact "<변경 영향>"
```

`@change` ID는 문서 내에서 유일하면 됩니다.
권장 패턴은 아래 둘 중 하나입니다.

- `CHG-2026-02-23-01` (가독성 중심)
- `CHG-20260223-173953` (시간 기반 고유성 중심)

## 7) 추천 운영 순서

1. `agd.exe new ...` 또는 위자드로 문서 생성
2. `section-add`/`edit`/`add`로 섹션 단위로 수정
3. `map-sync`로 map 정합성 재확인
4. `logic-log`로 변경 사유 기록
5. `check`로 검증 후 `view`로 Markdown 공유

## 8) 실패사례 분석 프레임 적용

실패를 줄이려면 문서 생성 이후 운영 루틴에 아래를 포함하세요.

1. 실패 유형을 메커니즘(문제정의/의존성/실행/의사결정/학습루프)으로 분류
2. `agd.exe check` 실패 원인 Top3를 주간 리뷰에서 집계
3. release 후보 문서의 `@change reason/impact` 누락 0건 유지
4. incident-case/postmortem 완료 전 source 문서 반영 증거 확인

상세 운영 프레임: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_ko.md`
