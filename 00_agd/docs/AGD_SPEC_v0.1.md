# AGD v0.1 (Agent Guided Document)

AGD는 사람과 AI 에이전트가 함께 읽고 수정할 수 있도록 설계된 대규모 기획 문서용 텍스트 포맷입니다.

## 1. 목표

1. 방대한 문서를 빠르게 훑어볼 수 있어야 한다.
2. 안정적인 섹션 ID 기반으로 AI가 정확한 수정 지점을 찾을 수 있어야 한다.
3. 변경 이력을 동일 문서 내부에 유지할 수 있어야 한다.
4. 엄격한 lint 규칙으로 기계 검증이 가능해야 한다.

## 2. 파일 확장자

- 권장 확장자: `.agd`

## 3. 핵심 블록

AGD는 최상위 `@block` 단위로 구성된다. 모든 블록 헤더는 줄 맨 앞(1열)에서 시작해야 한다.

- `@meta`: 문서 메타데이터
- `@map`: 한눈에 보는 섹션 인덱스
- `@section <SECTION-ID>`: 섹션 본문
- `@change <CHANGE-ID>`: 변경 이력 항목

## 4. 블록 문법

### 4.1 `@meta`

`key: value` 형식을 사용한다.

필수 키:
- `title`
- `version`

예시:
```txt
@meta
title: 결제 플랫폼 개편
version: 0.1
owner: product-team
```

### 4.2 `@map`

섹션당 한 줄씩 작성한다.

형식:
- `- <SECTION-ID> | <TITLE> | <SUMMARY>`

예시:
```txt
@map
- SEC-001 | 목표 | 결제 성공률 99.9 퍼센트 달성
- SEC-014 | 장애 롤백 | 장애 발생 시 5분 내 복구
```

### 4.3 `@section <SECTION-ID>`

필드:
- `title:`
- `summary:`
- `links:` (쉼표 구분)
- `content:` (멀티라인 본문)

`content:`는 반드시 `<<<` 와 `>>>` 펜스를 사용해야 한다.

예시:
```txt
@section SEC-014
title: 장애 자동 롤백
summary: 임계치 초과 시 결제 트래픽을 안정 버전으로 전환
links: TEST-022, CODE-payment-rollback
content:
<<<
트리거:
1) 최근 5분 실패율 3 퍼센트 초과
2) 승인 지연 P95 2.5초 초과
>>>
```

### 4.4 `@change <CHANGE-ID>`

필드:
- `target:` (섹션 ID)
- `date:` (`YYYY-MM-DD`)
- `author:`
- `reason:`
- `impact:` (필수)

`CHANGE-ID` 규칙:
- 문서 내부에서 유일해야 한다.
- v0.1에서는 특정 정규식을 강제하지 않는다.
- 권장 패턴 예시:
- `CHG-2026-02-23-01`
- `CHG-20260223-173953`

예시:
```txt
@change CHG-20260223-01
target: SEC-014
date: 2026-02-23
author: ai-agent
reason: 롤백 임계치를 수치로 명확히 추가
impact: 런북과 테스트 시나리오 동기화 필요
```

## 5. 권장 Lint 규칙

1. `meta.title`과 `meta.version`은 필수다.
2. `@map` 내 ID 중복은 허용하지 않는다.
3. `@section` 내 ID 중복은 허용하지 않는다.
4. `@map`의 모든 ID는 `@section`에 존재해야 한다.
5. `@section`의 모든 ID는 `@map`에 등록되어야 한다.
6. `@change.target`은 실제 존재하는 `@section`을 참조해야 한다.
7. `@change`의 `reason`은 비어 있으면 안 된다.
8. `@change`의 `impact`는 비어 있으면 안 된다.

## 6. v0.1 범위 밖

1. 엄격한 ID 정규식 강제 (`SEC-001`, `REQ-120` 등)
2. 블록별 사용자 정의 스키마
3. 섹션 계층(중첩 구조)
4. JSON/YAML 양방향 변환
