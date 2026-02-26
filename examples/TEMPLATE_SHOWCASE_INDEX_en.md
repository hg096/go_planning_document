# Template Showcase Index (EN)

Examples are now arranged with the same hierarchy as `agd_docs` (`10_source / 20_derived / 30_shared`).

## Template Examples

- `prd` -> `examples/en/10_source/product/prd_subscription_checkout_en.agd`
- `runbook` -> `examples/en/20_derived/ops/runbook_payment_incident_en.agd`
- `postmortem` -> `examples/en/30_shared/postmortem/postmortem_api_timeout_en.agd`
- `roadmap` -> `examples/en/30_shared/roadmap/roadmap_ai_doc_platform_en.agd`
- `handoff` -> `examples/en/30_shared/handoff/handoff_backend_frontend_en.agd`
- `meeting` -> `examples/en/30_shared/meeting/meeting_checkout_weekly_en.agd`
- `adr` -> `examples/en/10_source/architecture/adr_checkout_retry_policy_en.agd`
- `policy` -> `examples/en/10_source/policy/policy_release_gate_en.agd`
- `experiment` -> `examples/en/30_shared/experiment/experiment_oneclick_checkout_en.agd`
- `qa-plan` -> `examples/en/20_derived/qa/qa_plan_checkout_v2_en.agd`
- `service-logic` -> `examples/en/10_source/service/service_logic_checkout_core_en.agd`
- `frontend-page` -> `examples/en/20_derived/frontend/frontend_page_checkout_en.agd`
- `maintenance-case` -> `agd_docs/30_shared/maintenance/<project>_maintenance_case.agd` (kit-generated path)
- `incident-case` -> `agd_docs/30_shared/errFix/<project>_incident_case.agd` (kit-generated path)

## Source/Derived Role Examples

### 1) Service Logic (source) -> Frontend Page (derived)

- source: `examples/en/10_source/service/service_logic_checkout_core_en.agd`
- derived: `examples/en/20_derived/frontend/frontend_page_checkout_en.agd`
- derived meta:
- `authority: derived`
- `source_doc: ../../10_source/service/service_logic_checkout_core_en.agd`
- `source_sections: SYS-020->FP-020,SYS-060->FP-060,SYS-030->FP-030,SYS-050->FP-040`

### 2) Policy (source) -> QA Plan (derived)

- source: `examples/en/10_source/policy/policy_release_gate_en.agd`
- derived: `examples/en/20_derived/qa/qa_plan_checkout_v2_en.agd`
- derived meta:
- `authority: derived`
- `source_doc: ../../10_source/policy/policy_release_gate_en.agd`
- `source_sections: POL-010->QA-030,POL-030->QA-040`

## Validation Commands

```cmd
agd.exe check examples/en/10_source/service/service_logic_checkout_core_en.agd
agd.exe check examples/en/20_derived/frontend/frontend_page_checkout_en.agd
agd.exe check examples/en/10_source/policy/policy_release_gate_en.agd
agd.exe check examples/en/20_derived/qa/qa_plan_checkout_v2_en.agd
```
