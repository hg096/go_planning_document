# Examples Directory Guide

This directory uses the latest 9-template example set.

Legacy example set (`prd/runbook/qa/service/adr/postmortem`) has been removed.

## Naming Rule

All files use:

`project_docType_lang`

Example:

`checkout_core_spec_en.agd`

## Language Structure

- `en/10_source/*`
- `en/20_derived/*`
- `en/30_shared/*`
- `ko/10_source/*`
- `ko/20_derived/*`
- `ko/30_shared/*`

## Document Type Matrix

| doc_type | en path | ko path |
|---|---|---|
| core_spec | `en/10_source/product/checkout_core_spec_en.agd` | `ko/10_source/product/checkout_core_spec_ko.agd` |
| policy | `en/10_source/policy/checkout_policy_en.agd` | `ko/10_source/policy/checkout_policy_ko.agd` |
| delivery_plan | `en/20_derived/frontend/checkout_delivery_plan_en.agd` | `ko/20_derived/frontend/checkout_delivery_plan_ko.agd` |
| meeting | `en/30_shared/meeting/checkout_meeting_en.agd` | `ko/30_shared/meeting/checkout_meeting_ko.agd` |
| handoff | `en/30_shared/handoff/checkout_handoff_en.agd` | `ko/30_shared/handoff/checkout_handoff_ko.agd` |
| roadmap | `en/30_shared/roadmap/checkout_roadmap_en.agd` | `ko/30_shared/roadmap/checkout_roadmap_ko.agd` |
| experiment | `en/30_shared/experiment/checkout_experiment_en.agd` | `ko/30_shared/experiment/checkout_experiment_ko.agd` |
| maintenance_case | `en/30_shared/maintenance/checkout_maintenance_case_en.agd` | `ko/30_shared/maintenance/checkout_maintenance_case_ko.agd` |
| incident_case | `en/30_shared/errFix/checkout_incident_case_en.agd` | `ko/30_shared/errFix/checkout_incident_case_ko.agd` |

## Index Files

- Korean index: `TEMPLATE_SHOWCASE_INDEX_ko.md`
- English index: `TEMPLATE_SHOWCASE_INDEX_en.md`

## Validation Commands

```cmd
00_agd\agd_en.exe check-all 00_agd\examples --strict
00_agd\agd_en.exe role-graph 00_agd\examples --scope all
00_agd\agd_en.exe role-graph 00_agd\examples\en --scope maintenance
00_agd\agd_en.exe role-graph 00_agd\examples\en --scope incident
00_agd\agd_en.exe role-graph 00_agd\examples\en --scope maintenance-incident
00_agd\agd_en.exe role-graph 00_agd\examples\en --scope exclude-maintenance-incident
```
