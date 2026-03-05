# AGD 패키지 (격리 모드)

영문 문서는 `00_agd/README.en.md`를 참고하세요.
설치/세팅 절차는 루트 [README.md](../README.md)를 참고하세요.

이 패키지는 기존 저장소에 충돌을 최소화하면서, AI 문서 운영을 `00_agd/` 내부로 격리하기 위해 설계되었습니다.

## 0) AI 운영 사이클 (핵심)

아래 1~7 단계를 반복하면 AI가 문서 누락 없이 안정적으로 운영됩니다.

1. 요청 정규화
- AI 입력 5항목을 먼저 고정합니다.
- `target_docs`, `target_sections`, `goal`, `constraints`, `done`

2. 기준선 확인
- `00_agd/agd_docs/10_source/*` 기준 문서를 먼저 읽습니다.
- 같은 주제의 source 문서가 여러 개인지 충돌을 점검합니다.

3. 섹션 단위 수정
- 지정된 `target_sections`만 수정합니다.
- 새 파일은 명시 지시가 있을 때만 생성합니다.

4. 변경 추적 기록
- 모든 변경에 `@change(reason/impact)`를 남깁니다.
- 유지보수/오류 문서는 태그 메타/블록을 유지합니다.

5. 정합성 동기화
- `@map`과 `@section` ID 일치를 확인합니다.
- 필요 시 `map-sync`를 실행합니다.

6. 검증 실행
- 기본: `code-plan-check --mode auto`
- 필수: `check-all 00_agd/agd_docs --strict`

7. 승인/병합
- 로컬 훅 경고 확인 후 커밋
- 최종 강제는 CI(`AGD Guard`) Required check로 처리

실패 분기 규칙:
- `check-all` 실패: 문서 정합성 수정 후 재검증
- `code-plan-check` 실패: `doc_base_path`/`section.path` 또는 태그/`@change` 보강 후 재검증

관련 상세 규약:
- `00_agd/docs/AGD_TEMPLATE_GUIDE_ko.md`

## 1) AI 요청 템플릿 (복붙용)

```txt
target_docs: 00_agd/agd_docs/...
target_sections: CORE-020, DEL-010
goal: 이번 변경 목적 1~2문장
constraints: 수정 금지 범위/조건
done: 어떤 검증 결과가 나오면 완료인지
```

## 2) 일상 운영 루틴 (권장)

작업 시작 전:

```cmd
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

수정 후:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

관계 확인:

```cmd
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
```

## 3) 포함 구성

- 실행 파일: `00_agd/agd.exe`, `00_agd/agd_en.exe`, `00_agd/agd_ko.exe`
- 문서: `00_agd/docs/*`
- 예시: `00_agd/examples/*`
- 작업 문서 루트: `00_agd/agd_docs/*`
- 로컬 훅: `00_agd/.githooks/pre-commit`
- 훅 로직: `00_agd/scripts/git-hooks/pre-commit.ps1`
- 선택 템플릿(CI/PR): `00_agd/templates/*`

`00_agd/` 폴더에는 `.go` 소스 파일을 포함하지 않습니다.

## 4) 핵심 로직 수정 주의 (Soft Guard)

핵심 경로 정책 파일:
- `00_agd/policy/core_logic_paths.txt`

규칙:
- 저장소 루트 기준 상대 경로
- `#` 주석 허용
- 빈 줄 무시
- `/`, `\` 혼용 허용(훅에서 정규화)

커밋 훅 동작:
- staged 파일이 정책 경로와 일치하면 경고 출력
- 대화형에서는 `Continue commit? (y/N)` 확인
- 비대화형(CI)은 경고만 출력하고 진행

## 5) 코드 기획 검증 (Code Plan Check)

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe code-plan-check --mode git
00_agd\agd_en.exe code-plan-check --mode cache
00_agd\agd_en.exe code-plan-check --mode git --git-source staged --strict-relation
```

동작 요약:
- 변경 코드 파일 수집 (`auto`: git 우선, 실패 시 cache fallback)
- `doc_base_path`/`section.path`와 코드 경로 겹침 검사
- 매칭 문서는 `유지보수/오류 tag` 또는 `@change 업데이트` 중 하나 충족 시 통과
- 연관 정책(`A#SEC <-> B#SEC`)이 있으면 서비스↔프론트 양방향 연관 문서의 `@change` 갱신 여부를 추가 검사

정책/캐시 경로:
- `00_agd/policy/code_plan_scope.txt`
- `00_agd/policy/code_plan_relations.txt`
- `00_agd/.cache/code_plan_validation.json`

연관 정책 문법:
- `<left-doc.agd>#<SECTION-ID> <-> <right-doc.agd>#<SECTION-ID>`
- 한 줄에는 관계 1개만 작성
- 여러 관계는 줄을 나눠 반복 작성
- 1:N은 왼쪽 endpoint를 반복, N:1은 오른쪽 endpoint를 반복
- 한 줄에서 `,`로 여러 endpoint를 나열하는 형식은 지원하지 않음
- 예: `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010`
- 예(1:N):
  - `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010`
  - `00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-020`
- 기본은 WARN, `--strict-relation` 사용 시 FAIL

강화 옵션:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto --strict-mapping
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

훅 강제 동작:
- pre-commit은 `code-plan-check --mode git --git-source staged --strict-relation`을 먼저 실행해 연관 문서 누락을 차단

## 6) agd_docs README 통합 (기존 원문 보존)

아래 내용은 기존 `00_agd/agd_docs/README.md`에서 운영에 필요한 내용을 누락 없이 통합한 것입니다.

### 6-1) 폴더 구조

- `00_inbox`: 임시 초안/분류 전 메모
- `10_source/*`: 기준(source) 문서
- `20_derived/*`: source와 연결된 파생 문서
- `30_shared/*`: 공유 계획/협업 문서
- `90_archive`: 보관 문서

### 6-2) 주제별 권장 배치

- `10_source/product`: 제품 범위/성공기준
- `10_source/service`: 서비스/백엔드 로직 기준
- `10_source/policy`: 규칙/표준/승인/배포 정책
- `10_source/architecture`: ADR/트레이드오프

- `20_derived/frontend`: 화면별 구현/검증/배포 문서
- `20_derived/qa`: 회귀/권한/복구 테스트 문서
- `20_derived/ops`: 운영 런북 문서

- `30_shared/meeting`: 의사결정/액션 로그
- `30_shared/handoff`: 릴리즈 인수인계
- `30_shared/roadmap`: 마일스톤/우선순위
- `30_shared/postmortem`: 장애 회고
- `30_shared/experiment`: 실험 기록
- `30_shared/maintenance`: 유지보수 단일 흐름
- `30_shared/errFix`: 장애/버그 단일 흐름

### 6-3) Guide-First 운용법

1) `00_agd/docs/AGD_TEMPLATE_GUIDE_ko.md` 또는 `AGD_TEMPLATE_GUIDE_en.md`를 먼저 읽는다.
2) 서비스 주제는 `10_source/service` 단일완결 우선 원칙으로 유지한다.
3) source 문서를 기준으로 derived를 연결한다.
4) 서비스 source 변경 시 연결된 derived/source_sections를 같은 변경셋에서 동기화한다.
5) 변경은 항상 `@change(reason/impact)`를 남긴다.
6) 검증 3종 + 블라인드 재구성 테스트 통과 전 완료 보고를 금지한다.

### 6-4) 최소 실행 명령 세트

```cmd
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

### 6-5) 블라인드 재구성 테스트

- 입력: `10_source/service` 문서만 제공(코드 참조 금지)
- 산출: 도메인별 의사구현서(호출순서, DTO, 실패분기, 상태전이)
- 합격: 필수 블록 누락 0건 + 매핑 불일치 0건

### 6-6) 품질 기준 (9/10)

- 루브릭 총점 9점 이상
- 구조 정합성 1점 필수
- 감점 기준:
  - 도메인 의사코드 10단계 미만
  - 실패분기 에러코드/HTTP 매핑 누락
  - DTO 타입/필수/검증 제약 누락

### 6-7) 문서 스타일 규약 (예시 정렬)

예시 기준: `00_agd/examples/ko/*`

- `@section`은 블록형 `{ ... }`를 표준으로 사용한다.
- 블록 키 순서는 `summary -> links -> path -> content`를 유지한다.
- `content` 본문 줄은 탭(`\t`) 들여쓰기를 유지한다.
- 스타일 변경은 의미 변경 없이 포맷만 일관화한다.

### 6-8) 일반 품질 기준

- source 문서: 구현 가능한 계약 수준으로 작성
- derived 문서: source 섹션과 `source_sections`로 추적 가능하게 작성
- shared 문서: 운영/협업 실행에 필요한 의사결정 맥락 유지
- `@map`/`@section`/`@change` 정합성 유지

### 6-9) END__ 종료 규칙

- 유지보수 종료 문서: `30_shared/maintenance/END__*_maintenance_case.agd`
- 장애 종료 문서: `30_shared/errFix/END__*_incident_case.agd`
- `END__*` 문서는 스캔/선택/역할 그래프 기본 흐름에서 제외된다.
- 종료 해제 시 `END__` 접두어를 제거한다.
