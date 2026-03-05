# AGD AI 실행 가이드 (KO, Token-Light)

이 문서는 AI가 `00_agd/agd_docs`만 읽고, 문서 누락 없이 빠르게 작성/수정하기 위한 최소 운영 규약입니다.

## 1) 작업 범위

- 읽기/쓰기 루트: `00_agd/agd_docs`
- 우선 읽기 순서:
  1. `10_source/*`
  2. `20_derived/*`
  3. `30_shared/*`
- `END__*` 문서는 기본 제외

## 2) AI 입력 계약 (필수)

아래 5개가 없으면 수정하지 않습니다.

- `target_docs`: 대상 문서 경로/이름
- `target_sections`: 대상 섹션 ID 또는 section path 목록
- `goal`: 변경 목적 1~2문장
- `constraints`: 금지사항/범위 제한
- `done`: 완료 기준(검증 조건)

## 3) 경로 규칙 (중요)

- 기본 루트는 항상 `00_agd/agd_docs`
- 템플릿/예제 문서는 기본적으로 `@meta.doc_base_path`와 각 `@section.path` 키를 포함
- `@meta.doc_base_path` 초기값은 빈값 유지
- 템플릿/예제의 `@section.path` 기본값은 빈값 유지
- `doc_base_path`는 "실제 소스 코드 경로를 명시할 때만" 채움
- 자동 상대경로 변환/자동 채움 금지
- `@section.path`가 있으면 `doc_base_path`보다 우선 사용
- section path 권장 형식:
  - `src/.../file.ext#SEC-ID`
  - `backend/.../handler.go#CORE-020`

## 4) 실행 순서 (고정 루프)

1. 기준 `source` 문서 확인
2. 대상 섹션(`target_sections`) 확정
3. `@section` 수정
4. `@change`에 `reason` + `impact` 기록
5. `@map`/`@section` 정합성 확인
6. 검증 실행 (`check` 또는 `check-all`)
7. 실패 시 원인/수정 후 재검증

## 5) 필수 규칙

- 같은 주제의 기준 문서는 `source` 1개 유지
- 파생 문서는 `authority: derived`, `source_doc`, `source_sections` 필수
- 모든 변경은 `@change(reason/impact)` 필수
- `@map`과 `@section` 불일치 상태로 종료 금지
- source와 충돌 시 source 우선
- `target_sections` 누락 시 수정 금지
- `check`/`check-all` 실패 상태에서 완료 보고 금지
- 새 파일 생성은 명시 지시가 있을 때만 허용

## 6) 유지보수/오류 문서 규칙

- 유지보수 문서:
  - `maintenance_feature_tag`, `maintenance_source_doc`, `maintenance_source_section`
  - `[MAINTENANCE-ROOT-TAG]...[/MAINTENANCE-ROOT-TAG]`
- 오류 문서:
  - `incident_feature_tag`, `incident_source_doc`, `incident_source_section`
  - `[INCIDENT-PROBLEM-TAG]...[/INCIDENT-PROBLEM-TAG]`

## 7) 코드 변경 연계 검증 (권장 기본)

문서 누락 방지를 위해 수정 후 아래를 기본 실행합니다.

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
```

연관 정책(서비스↔프론트 양방향 섹션 매핑):

- 파일: `00_agd/policy/code_plan_relations.txt`
- 문법: `<left-doc.agd>#<SECTION-ID> <-> <right-doc.agd>#<SECTION-ID>`
- 한 줄에는 관계 1개만 작성합니다.
- 여러 관계는 줄을 나눠 여러 줄로 작성합니다.
- 1:N 관계는 왼쪽 endpoint를 각 줄에 반복합니다.
- N:1 관계는 오른쪽 endpoint를 각 줄에 반복합니다.
- `,`(콤마)로 한 줄에 여러 endpoint를 묶는 형식은 지원하지 않습니다.
- 연관 문서는 `@change` 갱신이 있어야 통과

예시:

```txt
# 1:N (서비스 섹션 1개 -> 프론트 섹션 여러 개)
00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-010
00_agd/agd_docs/10_source/service/checkout_service.agd#SYS-020 <-> 00_agd/agd_docs/20_derived/frontend/checkout_page.agd#FP-020

# N:1 (프론트 섹션 여러 개 -> 서비스 섹션 1개)
00_agd/agd_docs/10_source/service/payment_service.agd#SYS-040 <-> 00_agd/agd_docs/20_derived/frontend/payment_page.agd#FP-030
00_agd/agd_docs/10_source/service/refund_service.agd#SYS-050 <-> 00_agd/agd_docs/20_derived/frontend/payment_page.agd#FP-030
```

강화 모드:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto --strict-mapping
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

## 8) 빠른 명령 최소 세트

```cmd
00_agd\agd_en.exe check <file.agd>
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
00_agd\agd_en.exe map-sync <file.agd> --reason "..." --impact "..."
00_agd\agd_en.exe role-set <file.agd> source
00_agd\agd_en.exe role-set <file.agd> derived <source-doc> "<source_sections>"
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
```

## 9) AI 출력 형식 (권장, 6줄 고정)

```txt
Changed files: <경로1>, <경로2>
Changed sections: <SEC-ID>, <SEC-ID>
Reason: <핵심 변경 이유>
Impact: <영향/후속 조치>
Validation: <check/code-plan-check 결과 요약>
Risks: <없음|요약>
```

## 10) 자주 실패하는 패턴 (즉시 점검)

- `@change`에 `impact` 누락
- `derived`인데 `source_doc/source_sections` 누락
- `target_sections` 없이 광범위 수정
- `@map` 항목과 `@section` ID 불일치
- `doc_base_path`를 자동 추정값으로 채움
- 검증 실패 상태에서 완료 보고
