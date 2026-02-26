# 템플릿 활용 예시 인덱스 (KO)

국문 예시는 `agd_docs`와 같은 계층(`10_source / 20_derived / 30_shared`)으로 정리되어 있습니다.

## 템플릿별 예시 파일

- `prd` -> `examples/ko/10_source/product/prd_subscription_checkout_ko.agd`
- `runbook` -> `examples/ko/20_derived/ops/runbook_payment_incident_ko.agd`
- `postmortem` -> `examples/ko/30_shared/postmortem/postmortem_api_timeout_ko.agd`
- `roadmap` -> `examples/ko/30_shared/roadmap/roadmap_ai_doc_platform_ko.agd`
- `handoff` -> `examples/ko/30_shared/handoff/handoff_backend_frontend_ko.agd`
- `meeting` -> `examples/ko/30_shared/meeting/meeting_checkout_weekly_ko.agd`
- `adr` -> `examples/ko/10_source/architecture/adr_checkout_retry_policy_ko.agd`
- `policy` -> `examples/ko/10_source/policy/policy_release_gate_ko.agd`
- `experiment` -> `examples/ko/30_shared/experiment/experiment_oneclick_checkout_ko.agd`
- `qa-plan` -> `examples/ko/20_derived/qa/qa_plan_checkout_v2_ko.agd`
- `service-logic` -> `examples/ko/10_source/service/service_logic_checkout_core_ko.agd`
- `frontend-page` -> `examples/ko/20_derived/frontend/frontend_page_checkout_ko.agd`
- `maintenance-case` -> `agd_docs/30_shared/maintenance/<project>_maintenance_case.agd` (킷 생성 경로)
- `incident-case` -> `agd_docs/30_shared/errFix/<project>_incident_case.agd` (킷 생성 경로)

## 권위/파생(source/derived) 예시

### 1) 서비스 전체 로직(source) -> 프론트 페이지(derived)

- source: `examples/ko/10_source/service/service_logic_checkout_core_ko.agd`
- derived: `examples/ko/20_derived/frontend/frontend_page_checkout_ko.agd`
- derived meta:
- `authority: derived`
- `source_doc: ../../10_source/service/service_logic_checkout_core_ko.agd`
- `source_sections: SYS-020=>FP-020,SYS-060=>FP-060,SYS-030->FP-030,SYS-050->FP-040`

### 2) 정책(source) -> QA 계획(derived)

- source: `examples/ko/10_source/policy/policy_release_gate_ko.agd`
- derived: `examples/ko/20_derived/qa/qa_plan_checkout_v2_ko.agd`
- derived meta:
- `authority: derived`
- `source_doc: ../../10_source/policy/policy_release_gate_ko.agd`
- `source_sections: POL-010->QA-030,POL-030->QA-040`

## 검증 명령

```cmd
agd.exe check examples/ko/10_source/service/service_logic_checkout_core_ko.agd
agd.exe check examples/ko/20_derived/frontend/frontend_page_checkout_ko.agd
agd.exe check examples/ko/10_source/policy/policy_release_gate_ko.agd
agd.exe check examples/ko/20_derived/qa/qa_plan_checkout_v2_ko.agd
```
