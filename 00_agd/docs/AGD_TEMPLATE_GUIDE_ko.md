# AGD AI 실행 가이드 (KO, Token-Light)

이 문서는 AI가 `00_agd/agd_docs`만 읽고 빠르게 기획 문서를 작성/수정하기 위한 최소 실행 규약입니다.

## 1) 작업 범위

- 읽기/쓰기 기본 루트: `00_agd/agd_docs`
- 우선 읽기 순서:
  1. `10_source/*`
  2. `20_derived/*`
  3. `30_shared/*`
- `END__*` 문서는 기본적으로 제외합니다.

## 2) 입력 규약 (AI에 반드시 제공)

- `target_docs`: 대상 문서 경로/이름
- `target_sections`: 대상 섹션 ID 목록 (예: `CORE-020`, `SYS-070`)
- `goal`: 이번 수정 목적 1~2문장
- `constraints`: 금지사항/범위 제한
- `done`: 완료 기준(검증 조건)
- 문서 메타: `@meta.doc_base_path` 유지 (초기값은 빈값)
- 섹션 메타: 각 `@section`에 `path:` 유지 (예: `10_source/product/core_spec.agd#CORE-020`)

## 3) 실행 순서 (고정)

1. `source` 문서 기준선 확인
2. 수정 대상 섹션 확정
3. `@section` 내용 수정
4. `@change`에 `reason` + `impact` 기록
5. `@map`과 `@section` 정합성 확인
6. `check` 또는 `check-all` 통과 확인

## 4) 필수 규칙

- 같은 주제의 기준 문서는 `source` 1개만 유지
- 파생 문서는 `authority: derived` + `source_doc` + `source_sections` 필수
- 모든 변경은 `@change(reason/impact)` 필수
- `@map`과 `@section` 불일치 상태로 종료 금지
- source 기준과 충돌 시 source 우선
- `target_sections`가 없으면 수정 금지
- `10_source` 기준이 없거나 충돌하면 질문 1개 후 대기
- `check`/`check-all` 실패 상태에서는 완료 보고 금지
- 새 파일 생성은 명시 지시가 있을 때만 허용
- 섹션 참조 입력이 `path`와 `ID` 모두 가능할 때는 `path`를 우선 사용
- `doc_base_path`는 실제 기준 source 경로를 명시할 때만 채움 (자동 상대경로 금지)

## 5) AI 출력 형식 (권장)

아래 6줄 형식으로만 출력하면 됩니다.

```txt
Changed files: <경로1>, <경로2>
Changed sections: <SEC-ID>, <SEC-ID>
Reason: <핵심 변경 이유>
Impact: <영향/후속 조치>
Validation: <check 결과 요약>
Risks: <없음|요약>
```

## 6) 빠른 명령 최소 세트

```cmd
agd.exe check <file.agd>
agd.exe check-all
agd.exe map-sync <file.agd> --reason "..." --impact "..."
agd.exe role-set <file.agd> source
agd.exe role-set <file.agd> derived <source-doc> "<source_sections>"
agd.exe role-graph --scope all
```

## 7) 관계 그래프 범위 필터

- `all`: 전체
- `maintenance`: 유지보수만(연결 노드 포함)
- `incident`: 오류만(연결 노드 포함)
- `maintenance-incident`: 유지보수+오류만
- `exclude-maintenance-incident`: 유지보수/오류 제외
