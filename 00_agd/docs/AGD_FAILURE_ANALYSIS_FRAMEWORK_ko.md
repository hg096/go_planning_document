# AGD 기획 실패사례 분석 프레임 (KO)

이 문서는 기획 실패를 회고로 끝내지 않고, `source/derived` 문서 체계에 바로 반영하기 위한 운영 프레임입니다.

## 1) 실패를 보는 기준: 실패 메커니즘 분류

실패사례는 기능 단위가 아니라 아래 메커니즘으로 분류합니다.

- 문제정의 실패: 해결하려는 문제가 불명확하거나 성공 지표가 없음
- 사용자 여정 실패: 내부 기능은 맞지만 실제 사용자 흐름이 끊김
- 의존성 실패: 타 팀/외부 시스템/승인 경로가 누락됨
- 실행가능성 실패: 일정/리소스/기술 제약을 반영하지 못함
- 의사결정 실패: 문서 기준이 분산되어 충돌이 발생함
- 학습루프 실패: incident-case/postmortem 결과가 다음 source 문서에 반영되지 않음

## 2) 기획 실패사례 분석 프레임 (6단계)

1. 사건 정의  
   - 무엇이 실패했는지, 언제/어디서 발생했는지 한 문장으로 정의
2. 의사결정 체인 복원  
   - 어떤 가정과 근거로 어떤 결정을 했는지 시간순으로 정리
3. 놓친 신호 식별  
   - 당시 있었지만 무시된 경고 지표, 테스트 공백, 리뷰 누락 기록
4. 근본 원인 분류  
   - 위 6개 실패 메커니즘 중 어디에 속하는지 분류
5. 재발 방지 장치 설계  
   - 문서 규칙, 검증 게이트, 템플릿, 운영 루틴 중 어디에 반영할지 지정
6. 반영 증거 남기기  
   - 어떤 `source/derived` 문서를 실제로 수정했는지 경로와 섹션 ID로 기록

## 3) 지금 프로젝트에 바로 넣은 보완

아래 항목은 즉시 운영 가능한 기본 보완안입니다.

1. source 문서에 중단 기준 섹션 추가  
   - 대상 템플릿: `service-logic-ko`, `service-logic-en`  
   - 목적: 잘못된 방향으로 계속 가는 리스크를 조기 차단
2. derived 문서에 가정/리스크 섹션 추가  
   - 대상 템플릿: `frontend-page-ko/en`, `qa-plan-ko/en`  
   - 목적: 숨은 전제를 문서화하고 깨질 때 알림 조건을 명시
3. 주간 리뷰에서 `check` 실패 원인 Top3 집계  
   - 목표: 동일 유형 충돌 반복 방지
4. 릴리즈 게이트에서 `@change reason/impact` 완결성 점검  
   - 목표: "왜 바꿨는지"가 없는 변경 차단
5. incident-case/postmortem 완료 조건에 source 반영 증거 추가  
   - 대상 템플릿: `incident-case-ko/en`, `postmortem-ko/en`  
   - 목표: 회고 결과가 실제 기준 문서로 역반영되도록 강제

## 4) 운영 체크리스트

주간 체크:

1. `agd.exe check` 실패 로그를 유형별로 분류했는가
2. Top3 실패 유형에 대해 소유자와 마감일을 지정했는가
3. 파생 문서의 `source_doc/source_sections` 누락 건수를 0으로 유지했는가
4. 변경 건의 `@change reason/impact` 누락이 없는가
5. 이번 주 incident-case/postmortem의 source 반영 링크가 남아 있는가

## 5) 실행 예시

```cmd
REM 문서 검증
agd.exe check service_logic_checkout_core_ko
agd.exe check frontend_page_checkout_ko

REM 변경 이력 기록 (reason/impact 추적)
agd.exe logic-log service_logic_checkout_core_ko SYS-070 "중단 기준 임계치 조정"

REM source/derived 연결 점검
agd.exe map-suggest frontend_page_checkout_ko service_logic_checkout_core_ko smart-auto
agd.exe check frontend_page_checkout_ko
```

## 6) 연계 문서

- 시작 가이드: `README.md`
- 문서 구조 가이드: `docs/AGD_DOC_STRUCTURE_GUIDE_ko.md`
- 템플릿 가이드: `docs/AGD_TEMPLATE_GUIDE_ko.md`
- 영문 버전: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_en.md`
