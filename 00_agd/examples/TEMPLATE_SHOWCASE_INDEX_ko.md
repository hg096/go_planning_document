# 템플릿 활용 예시 인덱스 (KO)

국문 예시는 최신 템플릿 9종 기준으로 `examples/ko` 아래에 정리되어 있습니다.

파일명 규칙은 `project_docType_lang`입니다.

예: `checkout_core_spec_ko.agd`

## 템플릿별 예시 파일

- `core-spec` -> `examples/ko/10_source/product/checkout_core_spec_ko.agd`
- `policy` -> `examples/ko/10_source/policy/checkout_policy_ko.agd`
- `delivery-plan` -> `examples/ko/20_derived/frontend/checkout_delivery_plan_ko.agd`
- `meeting` -> `examples/ko/30_shared/meeting/checkout_meeting_ko.agd`
- `handoff` -> `examples/ko/30_shared/handoff/checkout_handoff_ko.agd`
- `roadmap` -> `examples/ko/30_shared/roadmap/checkout_roadmap_ko.agd`
- `experiment` -> `examples/ko/30_shared/experiment/checkout_experiment_ko.agd`
- `maintenance-case` -> `examples/ko/30_shared/maintenance/checkout_maintenance_case_ko.agd`
- `incident-case` -> `examples/ko/30_shared/errFix/checkout_incident_case_ko.agd`

## source/derived 관계 예시

`core_spec -> delivery_plan` 연결 예시:

- source: `examples/ko/10_source/product/checkout_core_spec_ko.agd`
- derived: `examples/ko/20_derived/frontend/checkout_delivery_plan_ko.agd`
- derived meta:
- `authority: derived`
- `source_doc: ../../10_source/product/checkout_core_spec_ko.agd`
- `source_sections: CORE-010->DEL-001,CORE-030->DEL-020`

## 태그 기반 연결 예시

유지보수 문서(`maintenance_case`)는 다음 메타/블록을 포함합니다.

- `maintenance_feature_tag: FT-CHECKOUT`
- `maintenance_source_doc: 00_agd\examples\ko\10_source\product\checkout_core_spec_ko.agd`
- `maintenance_source_section: CORE-010`
- `[MAINTENANCE-ROOT-TAG] ... [/MAINTENANCE-ROOT-TAG]`

오류 문서(`incident_case`)는 다음 메타/블록을 포함합니다.

- `incident_feature_tag: FT-CHECKOUT`
- `incident_source_doc: 00_agd\examples\ko\10_source\product\checkout_core_spec_ko.agd`
- `incident_source_section: CORE-010`
- `[INCIDENT-PROBLEM-TAG] ... [/INCIDENT-PROBLEM-TAG]`

## 검증 명령

```cmd
00_agd\agd_en.exe check-all 00_agd\examples --strict
00_agd\agd_en.exe role-graph 00_agd\examples --scope all
00_agd\agd_en.exe role-graph 00_agd\examples\ko --scope maintenance
00_agd\agd_en.exe role-graph 00_agd\examples\ko --scope incident
```
