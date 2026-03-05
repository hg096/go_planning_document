# AGD AI 실행 가이드 (KO, Hybrid v2)

이 문서는 사용자 지시 디테일이 낮아도 AI가 `00_agd/agd_docs`를 일관되게 생성/수정하도록 만드는 가이드 우선 운영 규약입니다.

## 1) 운영 기본값

- 운영 모드: `혼합 모드`
- 자동화 표면: `가이드 우선`
- 언어 정책: `ko/en 병행`
- 원칙: 코드/인프라 변경 없이 AGD 문서 품질과 재현성을 높인다.

## 2) 작업 범위

- 읽기/쓰기 루트: `00_agd/agd_docs`
- 우선 읽기 순서:
  1. `10_source/*`
  2. `20_derived/*`
  3. `30_shared/*`
- `END__*` 문서는 기본 제외

## 3) AI 입력 계약 v2

### 3.1 필수 입력 (없으면 실패)

- `goal`: 변경 목적 1~2문장
- `constraints`: 금지사항/범위 제한
- `done`: 완료 기준(검증 조건)

### 3.2 준필수 입력 (없으면 경고 후 진행)

- `target_docs`: 대상 문서 경로/이름
- `target_sections`: 대상 섹션 ID 또는 section path 목록

### 3.3 누락 처리 규칙

1) 필수 입력 누락:
- 작업 중단
- 누락 키 목록만 간결히 반환

2) 준필수 입력 누락:
- repo 탐색으로 후보 범위 제안
- 적용 가정(assumption) 명시 후 진행
- 결과에 경고(warn)로 기록

## 4) 경로 규칙

- 기본 루트는 항상 `00_agd/agd_docs`
- `@meta.doc_base_path`는 초기값 빈값 허용
- `doc_base_path`는 실제 소스 경로가 명확할 때만 채움
- `@section.path`가 있으면 `doc_base_path`보다 우선
- 자동 상대경로 변환/자동 채움 금지

## 5) 실행 순서 (고정 루프)

1. source 기준 문서 확인
2. 대상 문서/섹션 확정
3. `@section` 수정
4. `@change`에 `reason` + `impact` 기록
5. `@map`/`@section` 정합성 확인
6. 검증 실행 (`check-all`, `role-graph`, `code-plan-check`)
7. 실패 원인 최소 수정 후 재검증

## 6) 혼합 모드 판정표

### 6.1 Fail (완료 보고 금지)

- `@map`/`@section` ID 불일치
- derived 문서의 `source_doc`/`source_sections` 누락
- 검증 실패 상태에서 완료 보고

### 6.2 Warn (진행 가능, 결과에 경고 표기)

- `doc_base_path` 빈값 유지
- 일부 `@section.path` 누락
- unassigned 공유 문서의 메타 최소값 일부 누락

### 6.3 Pass (완료 가능)

- `check-all --strict` 통과
- `role-graph` unresolved=0
- `code-plan-check` 통과
- 변경 섹션/근거/영향 보고 완료

## 7) 문서 타입별 필수 커버리지 계약

### 7.1 `10_source/service`

도메인 섹션마다 아래 9블록 필수:
1) 책임/경계
2) 호출 순서
3) 입력 DTO 계약
4) 핵심 분기/상태 전이
5) 트랜잭션/원자성
6) 캐시 키/무효화
7) 에러 코드/HTTP 매핑
8) 의사코드
9) 테스트 기준

### 7.2 `10_source/policy`

아래 3블록 필수:
1) 정책 결정 테이블
2) 실행 템플릿
3) 증거 체크리스트

### 7.3 `20_derived/frontend/public`

아래 8블록 필수:
1) 렌더링 모델
2) 상태 변수 계약
3) 초기화 순서
4) 이벤트-상태 전이
5) API 계약(요청/응답)
6) 오류/빈데이터 fallback
7) 브라우저 저장소 계약
8) 백엔드 매핑

### 7.4 `20_derived/frontend/admin`

public과 동일 수준으로 아래 항목 필수:
- 상태/전이/API 계약
- 권한 코드
- 실패 분기/접근 차단 규칙

### 7.5 `20_derived/qa`

필수 축:
- API 회귀
- 권한 매트릭스
- 복구/관측성
- 릴리즈 스모크

### 7.6 `20_derived/ops`

필수 축:
- env 기준선
- 부팅 실패 진단
- 배포/롤백
- DB/수집 운영 체크

## 8) 관계 매핑 운영 규칙

- relation 파일: `00_agd/policy/code_plan_relations.txt`
- 문법: `<left>#<SEC> <-> <right>#<SEC>`
- 한 줄 1관계, 1:N/N:1은 endpoint 반복

### 8.1 동시 갱신 의무

- source 섹션 변경 시 연결 derived의 `source_sections`를 같은 변경셋에서 갱신
- derived 갱신 시 관련 `@change`를 함께 기록

### 8.2 서비스 섹션 리넘버링 체크리스트

1) source `@map`/`@section` 먼저 갱신
2) derived `source_sections` 전량 치환
3) relation endpoint 치환
4) `check-all` -> `role-graph` -> `code-plan-check --strict-relation`

## 9) 블록 스타일 표준 (예시 정렬)

예시 기준: `00_agd/examples/ko/*`

- `@section`은 블록형 `{ ... }` 사용
- 블록 키 순서: `summary -> links -> path -> content`
- `content` 본문은 탭(`\t`) 들여쓰기 유지

예시:

```agd
@section SEC-001
title: 제목
{
	summary: 요약
	links: LINK-001
	path: src/path/file.go
	content:
	<<<
		본문 1
		본문 2
	>>>
}
```

## 10) 사용자 프롬프트 최소 템플릿 (5줄)

### 10.1 신규 구성

1) 목표: <goal>
2) 범위: <target_docs>
3) 제약: <constraints>
4) 완료조건: <done>
5) 언어: <ko|en>

### 10.2 부분 보강

1) 목표: <goal>
2) 범위: <target_sections>
3) 제약: <constraints>
4) 완료조건: <done>
5) 언어: <ko|en>

### 10.3 섹션 마이그레이션

1) 목표: <section renumber/restructure>
2) 범위: <source + derived + relation>
3) 제약: <constraints>
4) 완료조건: <done>
5) 언어: <ko|en>

### 10.4 검증 실패 복구

1) 목표: <failed rule recovery>
2) 범위: <failing docs>
3) 제약: <constraints>
4) 완료조건: <done>
5) 언어: <ko|en>

## 11) 최소 검증 명령

```cmd
00_agd\agd_en.exe check-all 00_agd\agd_docs --strict
00_agd\agd_en.exe role-graph 00_agd\agd_docs --scope all
00_agd\agd_en.exe code-plan-check --mode auto --strict-relation
```

## 12) 블라인드 재구성 테스트 (필수)

- 입력: `10_source/service` 문서만 제공, 코드 참조 금지
- 산출: 도메인별 의사구현서(핸들러/서비스/저장소 호출순서, DTO, 실패분기, 상태전이)
- 판정:
  - 필수 블록 누락 0건
  - DTO 필드/검증 누락 0건
  - 실패분기의 에러코드/HTTP 매핑 누락 0건
  - 라우팅-권한-레이트리밋 매핑 불일치 0건

실패 시:
1) 누락 섹션을 source 문서에서 보강
2) @change(reason/impact) 기록
3) 최소 검증 명령 + 블라인드 테스트 재실행

## 13) 출력 품질 루브릭 (10점)

- 축 1 구조 정합성: 0/1
- 축 2 구현 가능성: 0~3
- 축 3 운영 가능성: 0~3
- 축 4 검증 증거: 0~3

합격 기준:
- 총점 9점 이상
- 구조 정합성 1점 필수

[감점 기준]
- 도메인 의사코드 10단계 미만
- 실패분기에서 에러코드/HTTP 매핑 누락
- DTO에서 타입/필수/검증 제약 누락

## 14) AI 출력 형식 (권장)

```txt
Changed files: <path1>, <path2>
Changed sections: <SEC-ID>, <SEC-ID>
Reason: <핵심 변경 이유>
Impact: <영향/후속 조치>
Validation: <check/role-graph/code-plan-check 요약>
Risks: <none|summary>
```

## 15) 자주 실패하는 패턴

- 필수 입력(goal/constraints/done) 누락 상태로 진행
- derived의 `source_doc/source_sections` 누락
- `@map` 항목과 `@section` ID 불일치
- source 변경 후 derived/relation 미동기화
- 검증 실패 상태에서 완료 보고
- 블라인드 재구성 테스트를 생략한 채 완료 보고
