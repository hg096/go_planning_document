// docs_scaffold.go defines default docs folder scaffold and README templates.
// docs_scaffold.go는 기본 문서 폴더 스캐폴드와 README 템플릿을 정의합니다.
package main

import (
	"path/filepath"
	"strings"
)

const docsRootLeafDir = "agd_docs"

var docsScaffoldFolders = []string{
	filepath.Join("10_core_logic", "service"),
	filepath.Join("10_core_logic", "policy"),
	filepath.Join("10_core_logic", "architecture"),
	"20_new_project",
}

const docsScaffoldReadmeEN = `# AGD Docs Folder Guide

This directory uses a focused hierarchy for planning work.

- 10_core_logic/*: source-of-truth logic documents
- 20_new_project/*: new project planning documents

Recommended practice:

1) Put existing product/service rules under 10_core_logic first.
2) Put new initiatives under 20_new_project and link them back to core logic.
3) Keep one core logic source per topic, then let project docs reference it with source_doc/source_sections.
4) Keep outdated or paused material inside the nearest project folder with an explicit status.

Folder details:

- 10_core_logic/service: backend/service logic, domain rules, failure branches, state transitions
- 10_core_logic/policy: rules, standards, approval/release policy
- 10_core_logic/architecture: architecture decisions (ADR) and trade-offs

- 20_new_project/<project-key>: project scope, policy, delivery plan, and roadmap in one folder

AI planning/writing guide (merged into this README):

1) Core philosophy: control conflict/traceability first, not raw file count.
2) Authority rule: core logic is the baseline; new project docs must point back to it when they depend on existing behavior.
3) Prompt contract (required): target section IDs, purpose, constraints, done criteria.
4) AI output (required): updated section IDs, reason, impact, follow-up checklist.
5) Every mutation must append @change with both reason and impact.
6) Keep @map, @section, and @change aligned after each edit.
7) Standard loop: edit -> @change -> map-sync -> check -> core-logic-first conflict resolution.
`

const docsScaffoldReadmeKO = `# AGD 문서 폴더 가이드

이 디렉터리는 기획 문서의 핵심을 두 축으로 좁혀 관리합니다.

- 10_core_logic/*: 핵심 로직 기준 문서
- 20_new_project/*: 신규 프로젝트 기획 문서

권장 방식:

1) 기존 제품/서비스의 판단 기준은 먼저 10_core_logic 아래에 둡니다.
2) 새로 시작하는 일은 20_new_project 아래에 두고, 필요한 경우 핵심 로직 문서를 source_doc/source_sections로 연결합니다.
3) 주제별 핵심 로직 source는 1개만 유지하고 신규 프로젝트 문서는 그 기준으로 수렴시킵니다.
4) 중단/보류된 문서는 별도 보관 폴더를 만들지 말고 가장 가까운 프로젝트 폴더 안에서 상태를 명시합니다.

폴더별 상세 설명:

- 10_core_logic/service: 서비스/백엔드 로직, 도메인 규칙, 실패 분기, 상태 전이
- 10_core_logic/policy: 규칙, 표준, 승인/배포 정책
- 10_core_logic/architecture: 아키텍처 의사결정(ADR)과 트레이드오프 기록

- 20_new_project/<project-key>: 프로젝트 범위, 정책, 전달 계획, 로드맵을 한 폴더에 모음

AI 기획/작성 가이드 (이 README에 통합):

1) 핵심 철학: 파일 수보다 충돌 통제와 추적 가능성을 우선합니다.
2) 권위 규칙: 핵심 로직을 기준선으로 두고 신규 프로젝트 문서는 필요한 경우 그 기준선을 참조합니다.
3) 프롬프트 계약(필수): 대상 섹션 ID, 목적, 제약 조건, 완료 기준.
4) AI 출력(필수): 수정 섹션 ID, reason, impact, 후속 체크리스트.
5) 모든 수정은 @change에 reason과 impact를 함께 기록합니다.
6) 수정 후 @map, @section, @change 정합성을 확인합니다.
7) 표준 루프: 수정 -> @change -> map-sync -> check -> 핵심 로직 우선 충돌 해결.
`

func docsScaffoldReadmeByLang(lang string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "ko") {
		if !isBrokenKOText(docsScaffoldReadmeKO) {
			return docsScaffoldReadmeKO
		}
	}
	return docsScaffoldReadmeEN
}
