// commands_help.go prints quick-start help for AGD CLI commands.
// commands_help.go는 AGD CLI 빠른 도움말을 출력합니다.
package main

import "fmt"

func commandHelp(_ []string) int {
	fmt.Print(text(`AGD quick guide

1) Create a new document:
   agd new core-spec core_spec_checkout "Checkout Core Spec" "product-team"

2) Search by keyword:
   agd search core_spec_checkout payment

3) Add a section (map updates together):
   agd section-add core_spec_checkout CORE-050 "Payment Failure Recovery" "Define recovery flow" --reason "fill missing flow" --impact "operators can recover consistently"

4) Update section summary:
   agd edit core_spec_checkout CORE-020 "Prioritize must-have features" --reason "align priority" --impact "reduce ambiguity"

5) Append sentence to section body:
   agd add core_spec_checkout CORE-020 "- Add failure recovery flow chart" --reason "add actionable detail" --impact "faster incident response"

6) Add logic-change history:
   agd logic-log core_spec_checkout CORE-020 --reason "retry branch updated" --impact "lower transient failure drop"

7) Tag incident feature root:
   agd incident-tag checkout_incident_case FT-CHECKOUT service_logic_checkout_core SYS-020 --reason "set exact root" --impact "AI finds target section fast"

8) Generate kit in one shot:
   agd starter-kit checkout
   agd new-project checkout --owner product-team
   agd maintenance checkout --tag-source service_logic_checkout_core --tag-section SYS-020
   agd incident checkout --tag-source service_logic_checkout_core --tag-section SYS-020

9) Show source/derived graph:
   agd role-graph
   agd role-graph --format mermaid --out 00_agd\agd_docs\role_graph.mmd

10) Interactive mode:
    agd wizard
`, `AGD 빠른 가이드

1) 새 문서 만들기:
   agd new core-spec core_spec_checkout "Checkout Core Spec" "product-team"

2) 키워드 검색:
   agd search core_spec_checkout payment

3) 섹션 추가 (map 동시 갱신):
   agd section-add core_spec_checkout CORE-050 "Payment Failure Recovery" "Define recovery flow" --reason "fill missing flow" --impact "operators can recover consistently"

4) 섹션 요약 수정:
   agd edit core_spec_checkout CORE-020 "Prioritize must-have features" --reason "align priority" --impact "reduce ambiguity"

5) 섹션 본문에 문장 추가:
   agd add core_spec_checkout CORE-020 "- Add failure recovery flow chart" --reason "add actionable detail" --impact "faster incident response"

6) 로직 변경 이력 추가:
   agd logic-log core_spec_checkout CORE-020 --reason "retry branch updated" --impact "lower transient failure drop"

7) 오류,버그 기능 루트 태그 지정:
   agd incident-tag checkout_incident_case FT-CHECKOUT service_logic_checkout_core SYS-020 --reason "set exact root" --impact "AI finds target section fast"

8) 문서 키트 한 번에 생성:
   agd starter-kit checkout
   agd new-project checkout --owner product-team
   agd maintenance checkout --tag-source service_logic_checkout_core --tag-section SYS-020
   agd incident checkout --tag-source service_logic_checkout_core --tag-section SYS-020

9) source/derived 그래프 보기:
   agd role-graph
   agd role-graph --format mermaid --out 00_agd\agd_docs\role_graph.mmd

10) 대화형 모드:
    agd wizard
`))
	return 0
}
