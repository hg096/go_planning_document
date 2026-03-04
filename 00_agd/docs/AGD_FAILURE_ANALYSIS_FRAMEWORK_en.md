# AGD Failure Analysis Framework (EN)

This guide defines how to convert planning failures into concrete updates in the `source/derived` document system.

## 1) Failure Mechanism Taxonomy

Classify failures by mechanism, not by feature area.

- Problem-definition failure: unclear problem or missing success metric
- User-journey failure: internal logic is correct but user flow breaks
- Dependency failure: missing cross-team/system/approval dependencies
- Execution-feasibility failure: plan ignores timeline/resource constraints
- Decision-governance failure: conflicting source documents and drift
- Learning-loop failure: incident-case/postmortem output never updates source docs

## 2) Planning Failure Analysis Frame (6 Steps)

1. Define the incident  
   - one sentence: what failed, when, and where
2. Reconstruct decision chain  
   - timeline of assumptions, evidence, and approvals
3. Identify missed signals  
   - early warnings, test gaps, review misses
4. Classify root cause  
   - map to one or more of the six failure mechanisms
5. Design recurrence prevention  
   - decide whether to fix via template/rule/check/process
6. Record implementation evidence  
   - list updated `source/derived` files and section IDs

## 3) Immediate Hardening Added to This Project

These are now the default hardening actions.

1. Add abort-metric section to source docs  
   - templates: `service-logic-ko`, `service-logic-en`  
   - objective: stop wrong-direction rollout early
2. Add assumptions/risk-signals section to derived docs  
   - templates: `frontend-page-ko/en`, `qa-plan-ko/en`  
   - objective: make hidden assumptions explicit and monitorable
3. Weekly Top3 `check` failure-cause review  
   - objective: prevent repeating the same conflict patterns
4. Release gate checks `@change reason/impact` completeness  
   - objective: block rationale-free changes
5. Incident-case/postmortem exit condition includes source-feedback evidence  
   - templates: `incident-case-ko/en`, `postmortem-ko/en`  
   - objective: enforce learning-loop closure

## 4) Weekly Checklist

1. Did we classify all `agd.exe check` failures by cause category?
2. Did we assign owner/due-date to Top3 recurring causes?
3. Are all derived docs linked with `source_doc/source_sections`?
4. Are `@change reason` and `impact` complete for release candidates?
5. Do all completed incident-cases/postmortems include source-doc update evidence?

## 5) Command Examples

```cmd
REM validate docs
agd.exe check service_logic_checkout_core_en
agd.exe check frontend_page_checkout_en

REM log changes with rationale
agd.exe logic-log service_logic_checkout_core_en SYS-070 "Adjusted abort thresholds"

REM review source/derived consistency
agd.exe map-suggest frontend_page_checkout_en service_logic_checkout_core_en smart-auto
agd.exe check frontend_page_checkout_en
```

## 6) Related Documents

- Start guide: `README.en.md`
- Document structure guide: `docs/AGD_DOC_STRUCTURE_GUIDE_en.md`
- Template guide: `docs/AGD_TEMPLATE_GUIDE_en.md`
- Korean version: `docs/AGD_FAILURE_ANALYSIS_FRAMEWORK_ko.md`
