# AGD 템플릿 가이드 (최신)

이 가이드는 처음 `agd.exe`를 사용하는 사용자를 위해 템플릿 생성과 운영 흐름을 설명합니다.
기본 문서 루트는 `00_agd\agd_docs`입니다.

## 0) 가장 쉬운 시작: Wizard (`agd.exe`)

```cmd
agd.exe
REM 또는
agd.exe wizard
```

상위 메뉴:

```txt
[1] 문서 선택
[2] 문서 키트 생성 (starter/bridge/change/incident/quality)
[3] 문서 트리 전체 검증
[4] 새 문서 만들기
[5] source/derived 관계 그래프 보기
[0] 뒤로/종료
```

문서를 선택한 뒤:

```txt
선택 문서: <문서명>
[1] 키워드 검색
[2] 새 섹션 추가 (map도 함께 갱신)
[3] 로직 변경 이력 추가
[4] map만 동기화
[5] 문서 검증
[6] Markdown 내보내기
[7] source/derived 역할 지정
[0] 뒤로
```

선택형 Wizard 화면(문서 타입/프로필/폴더/문서/섹션/역할/포맷)은 모두 `[0] Back`을 지원합니다.

Wizard `New document` 기본 세트(문서 수 감축 중심):

- `core-spec`
- `delivery-plan`
- `meeting`
- `experiment`
- `roadmap`
- `handoff`
- `policy`

CLI 생성(`agd.exe new`, `agd.exe init`)도 동일한 7개 통합 우선 타입으로 제한됩니다.

## 1) 빠른 생성: `agd new`

```cmd
agd.exe new list
agd.exe new core-spec core_spec_checkout "Checkout Core Spec" "product-team"
```

`agd new`에서 지원하는 문서 타입:

- `core-spec`: 통합 코어 스펙(PRD+로직+ADR+AI 가이드)
- `delivery-plan`: 통합 전달 계획(frontend+QA+release)
- `meeting`: 회의 노트 및 의사결정
- `experiment`: 실험 및 A/B 테스트 계획
- `roadmap`: 분기/반기 로드맵 계획
- `handoff`: 팀 간 인수인계 문서
- `policy`: 정책 및 가이드라인 문서

Wizard `New document`도 기본적으로 위 축약 세트를 사용합니다.

## 2) 상세 생성: `agd init`

```cmd
agd.exe init --list
agd.exe init --template core-spec-ko --out docs/core_spec_checkout_ko.agd --title "체크아웃 코어 스펙 (KO)" --owner "product-team"
agd.exe init --template core-spec-en --out docs/core_spec_checkout_en.agd --title "Checkout Core Spec" --owner "product-team"
```

주요 옵션:

- `--template`: 템플릿 이름(`core-spec-ko`, `core-spec-en`, `delivery-plan-ko`, ...)
- `--out`: 출력 파일 경로
- `--title`: 문서 제목(`@meta.title`)
- `--owner`: 문서 소유자(`@meta.owner`)
- `--version`: 문서 버전(`@meta.version`, 기본값 `0.1`)
- `--force`: 기존 파일 덮어쓰기

참고:

- `agd.exe init --list`는 이제 `*-ko`, `*-en` 템플릿을 모두 보여줍니다.

## 3) 경로 입력 규칙 (`00_agd\agd_docs` 기본)

기존 문서를 열 때 파일명만 입력하면 `00_agd\agd_docs` 하위 폴더 전체에서 검색합니다.

- 입력: `core_spec_checkout`
- 검색 대상: `00_agd\agd_docs\**\core_spec_checkout.agd`

하위 폴더에 동일 파일명이 여러 개면 `front/home` 같은 상대 경로를 사용합니다.

새 문서(`agd new`, wizard 상위 메뉴 `4`)는 파일명만 주면 문서 타입 기준으로 자동 배치됩니다.

- `core-spec`: `00_agd\agd_docs\10_source\product\<file>.agd`
- `delivery-plan`: `00_agd\agd_docs\10_source\product\<file>.agd`
- `policy`: `00_agd\agd_docs\10_source\policy\<file>.agd`
- `meeting`: `00_agd\agd_docs\30_shared\meeting\<file>.agd`
- `experiment`: `00_agd\agd_docs\30_shared\experiment\<file>.agd`
- `roadmap`: `00_agd\agd_docs\30_shared\roadmap\<file>.agd`
- `handoff`: `00_agd\agd_docs\30_shared\handoff\<file>.agd`

자동 배치를 무시하려면 명시 경로를 직접 전달하세요.

`END__` 종료 상태 규칙:

- `30_shared/maintenance/END__*_maintenance_case.agd`
- `30_shared/errFix/END__*_incident_case.agd`

이 파일들은 scan/select/role-graph 흐름에서 자동 제외됩니다.

## 4) 섹션/맵 명령

```cmd
REM 섹션 추가 + 맵 업데이트
agd.exe section-add <file.agd> <SECTION-ID> "<섹션 제목>" "<요약>" [links] [content] --reason "<사유>" --impact "<영향>"

REM 섹션 기반 맵 동기화
agd.exe map-sync <file.agd> --reason "<사유>" --impact "<영향>"

REM 섹션 요약 수정
agd.exe edit <file.agd> <SECTION-ID> "<새 요약>" --reason "<사유>" --impact "<영향>"

REM 섹션 본문에 한 줄 추가
agd.exe add <file.agd> <SECTION-ID> "<추가할 한 줄>" --reason "<사유>" --impact "<영향>"
```

모든 문서 변경은 `@change(reason/impact)`를 남겨야 합니다.
`edit`/`add`/`section-add`는 변경 후 `map-sync -> check`가 자동 실행됩니다.

트리 전체 검증:

```cmd
REM 문서 트리 전체 검증 (기본값: 00_agd\agd_docs)
agd.exe check-all
agd.exe check-all examples

REM 경고도 실패로 처리
agd.exe check-all examples --strict

REM 아카이브 문서 포함
agd.exe check-all --include-archive
```

서비스 로직 변경 시 권장 순서:

```cmd
REM 먼저 서비스 문서 업데이트
agd.exe check 10_source/service/<service_doc>

REM 트리 전체 정합성 확인
agd.exe check-all
```

## 5) Source/Derived(권위/파생) 문서 운영

```cmd
REM source (권위 원본) 문서 지정
agd.exe role-set service_overview source

REM 명시 매핑이 있는 derived 문서 지정
agd.exe role-set frontend_pages derived service_overview "SYS-020->FP-020,SYS-030=>FP-030"

REM 자동 매핑 모드
agd.exe role-set frontend_pages derived service_overview auto
agd.exe role-set frontend_pages derived service_overview strict-auto
agd.exe role-set frontend_pages derived service_overview smart-auto

REM 매핑 추천만 미리 보기
agd.exe map-suggest frontend_pages service_overview
agd.exe map-suggest frontend_pages service_overview strict-auto
agd.exe map-suggest frontend_pages service_overview smart-auto

REM 검증
agd.exe check frontend_pages
```

`source_sections` 규칙:

- `SRC->DST`: 연관/존재성 검사
- `SRC=>DST`: 제목/요약 엄격 충돌 검사
- `SEC-ID`: 동일 ID 엄격 충돌 검사
- `auto`: `->` 매핑 생성
- `strict-auto`: `=>` 매핑 생성
- `smart-auto`: 먼저 strict를 시도하고, 불일치는 `->`로 완화

## 6) 변경 로그(`@change`) 운영

```cmd
agd.exe logic-log <file.agd> <SECTION-ID> --reason "<변경 사유>" --impact "<변경 영향>"
```

`@change` ID는 문서 내부에서만 유일하면 됩니다.
권장 패턴:

- `CHG-2026-02-23-01` (가독성)
- `CHG-20260223-173953` (타임스탬프 유일성)

## 7) 권장 흐름

1. Wizard 또는 `agd.exe new ...`로 문서를 생성합니다.
2. 섹션 단위(`section-add`/`edit`/`add`)로 편집합니다.
3. `map-sync`로 맵을 다시 동기화합니다.
4. `logic-log`로 변경 이유를 남깁니다.
5. `check`로 검증한 뒤 `view`로 공유합니다.

## 8) 실패 분석 프레임워크 적용

아래 항목을 주간 운영 루프에 포함하세요.

1. 실패를 메커니즘(문제정의/의존성/실행/거버넌스/학습 루프) 기준으로 분류합니다.
2. 반복되는 `agd.exe check` 실패 원인 Top3를 검토합니다.
3. 릴리스 후보의 `@change reason/impact` 누락 건수를 0으로 유지합니다.
4. incident-case/postmortem 종료 전에 source 문서 피드백 근거를 확인합니다.

상세 프레임워크: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_ko.md`
