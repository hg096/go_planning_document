# AGD 문서 구조 운영 가이드 (KO)

이 가이드는 `00_agd\agd_docs`를 기준으로 문서를 어디에 두고, 어떤 순서로 운영하면 문서가 늘어나도 충돌 없이 관리할 수 있는지 설명합니다.

## 1) 표준 폴더 구조

```text
00_agd/agd_docs/
  00_inbox/
  10_source/
    product/
    service/
    policy/
    architecture/
  20_derived/
    frontend/
    qa/
    ops/
  30_shared/
    meeting/
    handoff/
    roadmap/
    postmortem/
    experiment/
    maintenance/
    errFix/
  90_archive/
```

## 2) 폴더 역할

- `00_inbox`: 분류 전 임시 문서
- `10_source/*`: 주제별 기준 문서(source)
- `20_derived/*`: 기준 문서를 바탕으로 만든 파생 문서(derived)
- `30_shared/*`: 회의/인수인계/로드맵/회고 + 유지보수/장애 단일 문서
- `90_archive`: 사용 종료 문서 보관

## 3) 문서 타입별 자동 배치

`agd.exe new <type> <name>`에서 `<name>`에 파일명만 입력하면 아래 경로로 자동 배치됩니다.

| 타입 | 자동 경로 | 권장 역할 |
|---|---|---|
| `prd` | `10_source/product` | source |
| `service-logic` | `10_source/service` | source |
| `policy` | `10_source/policy` | source |
| `adr` | `10_source/architecture` | source |
| `frontend-page` | `20_derived/frontend` | derived |
| `qa-plan` | `20_derived/qa` | derived |
| `runbook` | `20_derived/ops` | derived |
| `meeting` | `30_shared/meeting` | shared |
| `handoff` | `30_shared/handoff` | shared |
| `roadmap` | `30_shared/roadmap` | shared |
| `postmortem` | `30_shared/postmortem` | shared |
| `experiment` | `30_shared/experiment` | shared |
| `maintenance-case` | `30_shared/maintenance` | source (single case) |
| `incident-case` | `30_shared/errFix` | source (single case) |
| 기타 | `00_inbox` | triage |

원하는 위치를 직접 지정하려면 경로를 함께 입력하면 됩니다.

```cmd
agd.exe new prd 10_source/product/checkout_v3
```

## 4) source / derived 운영 규칙

- 주제 하나당 `source` 문서는 1개만 둡니다.
- 파생 문서는 `authority: derived`를 사용합니다.
- 파생 문서에는 반드시 `source_doc`, `source_sections`를 설정합니다.
- 변경 후에는 `agd.exe check <file>`로 충돌을 확인합니다.

예시:

```cmd
agd.exe role-set service_logic_checkout_core_ko source
agd.exe role-set frontend_page_checkout_ko derived service_logic_checkout_core_ko "SYS-020->FP-020,SYS-030=>FP-030"
agd.exe check frontend_page_checkout_ko
```

## 5) 추천 작성 순서

1. `10_source/service`: 서비스 전체 로직 작성
2. `10_source/product`: 요구사항/범위 확정
3. `10_source/policy`: 릴리즈/운영 정책 확정
4. `20_derived/frontend`: 페이지별 상세 로직 파생
5. `20_derived/qa`, `20_derived/ops`: 테스트/운영 문서 파생
6. `30_shared/*`: 회의, 인수인계, 로드맵, 회고 + 유지보수/장애 단일 케이스 축적

## 5-1) 종료 문서(`END__`) 자동 제외 규칙

- `30_shared/maintenance/END__*_maintenance_case.agd`
- `30_shared/errFix/END__*_incident_case.agd`

위 두 패턴은 스캔/선택/관계도 출력에서 자동 제외됩니다.

## 6) 운영 루틴 (문서 스프롤 방지)

일일 루틴:

1. 수정은 `SECTION-ID` 기준으로만 수행
2. `edit/add/section-add` 시 `--reason` + `--impact` 필수
3. 작업 직후 `map-sync -> check`
4. 서비스 로직 코드 변경 전 `agd.exe service-gate` 통과 확인
5. 코드 실행은 `run-safe.cmd`로 감싸서 게이트 우회 방지

주간 루틴:

1. `00_inbox` 문서 분류
2. 중복/충돌 문서 정리
3. 오래된 문서를 `90_archive`로 이동

## 7) 충돌 방지 체크리스트

- 같은 주제의 source 문서가 2개 이상인가?
- derived 문서에 `source_doc`/`source_sections`가 비어있는가?
- `@map`과 `@section` 제목/요약이 어긋났는가?
- 변경 후 `agd.exe check`를 생략했는가?
- 서비스 코드 변경 전에 `agd.exe service-gate`를 실행했는가?
- 더 이상 쓰지 않는 문서를 archive로 이동했는가?

## 8) examples와 연결해서 보는 법

- source 예시: `examples/ko/10_source/*`
- derived 예시: `examples/ko/20_derived/*`
- shared 예시: `examples/ko/30_shared/*`
- 템플릿 인덱스: `examples/TEMPLATE_SHOWCASE_INDEX_ko.md`

영문 대응 가이드: `docs/AGD_DOC_STRUCTURE_GUIDE_en.md`
