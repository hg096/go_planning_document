# Template Showcase Index (EN)

English examples follow the latest 9-template set under `examples/en`.

Filename convention is `project_docType_lang`.

Example: `checkout_core_spec_en.agd`

## Template Example Files

- `core-spec` -> `examples/en/10_source/product/checkout_core_spec_en.agd`
- `policy` -> `examples/en/10_source/policy/checkout_policy_en.agd`
- `delivery-plan` -> `examples/en/20_derived/frontend/checkout_delivery_plan_en.agd`
- `meeting` -> `examples/en/30_shared/meeting/checkout_meeting_en.agd`
- `handoff` -> `examples/en/30_shared/handoff/checkout_handoff_en.agd`
- `roadmap` -> `examples/en/30_shared/roadmap/checkout_roadmap_en.agd`
- `experiment` -> `examples/en/30_shared/experiment/checkout_experiment_en.agd`
- `maintenance-case` -> `examples/en/30_shared/maintenance/checkout_maintenance_case_en.agd`
- `incident-case` -> `examples/en/30_shared/errFix/checkout_incident_case_en.agd`

## Source/Derived Link Example

`core_spec -> delivery_plan` linkage:

- source: `examples/en/10_source/product/checkout_core_spec_en.agd`
- derived: `examples/en/20_derived/frontend/checkout_delivery_plan_en.agd`
- derived meta:
- `authority: derived`
- `source_doc: ../../10_source/product/checkout_core_spec_en.agd`
- `source_sections: CORE-010->DEL-001,CORE-030->DEL-020`

## Tag-Rooted Link Example

Maintenance (`maintenance_case`) includes:

- `maintenance_feature_tag: FT-CHECKOUT`
- `maintenance_source_doc: 00_agd\examples\en\10_source\product\checkout_core_spec_en.agd`
- `maintenance_source_section: CORE-010`
- `[MAINTENANCE-ROOT-TAG] ... [/MAINTENANCE-ROOT-TAG]`

Incident (`incident_case`) includes:

- `incident_feature_tag: FT-CHECKOUT`
- `incident_source_doc: 00_agd\examples\en\10_source\product\checkout_core_spec_en.agd`
- `incident_source_section: CORE-010`
- `[INCIDENT-PROBLEM-TAG] ... [/INCIDENT-PROBLEM-TAG]`

## Validation Commands

```cmd
00_agd\agd_en.exe check-all 00_agd\examples --strict
00_agd\agd_en.exe role-graph 00_agd\examples --scope all
00_agd\agd_en.exe role-graph 00_agd\examples\en --scope maintenance
00_agd\agd_en.exe role-graph 00_agd\examples\en --scope incident
```
