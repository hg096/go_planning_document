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

## 6) 핵심 로직 수정 주의 (Soft Guard)

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

## 7) 코드 기획 검증 (Code Plan Check)

```cmd
00_agd\agd_en.exe code-plan-check --mode auto
00_agd\agd_en.exe code-plan-check --mode git
00_agd\agd_en.exe code-plan-check --mode cache
```

동작 요약:
- 변경 코드 파일 수집 (`auto`: git 우선, 실패 시 cache fallback)
- `doc_base_path`/`section.path`와 코드 경로 겹침 검사
- 매칭 문서는 `유지보수/오류 tag` 또는 `@change 업데이트` 중 하나 충족 시 통과

정책/캐시 경로:
- `00_agd/policy/code_plan_scope.txt`
- `00_agd/.cache/code_plan_validation.json`

강화 옵션:

```cmd
00_agd\agd_en.exe code-plan-check --mode auto --strict-mapping
```
