// docs_scaffold.go defines default docs folder scaffold and README templates.
// docs_scaffold.go는 기본 문서 폴더 스캐폴드와 README 템플릿을 정의합니다.
package main

import (
	"path/filepath"
	"strings"
)

const docsRootLeafDir = "agd_docs"

var docsScaffoldFolders = []string{
	"00_inbox",
	filepath.Join("10_source", "product"),
	filepath.Join("10_source", "service"),
	filepath.Join("10_source", "policy"),
	filepath.Join("10_source", "architecture"),
	filepath.Join("20_derived", "frontend"),
	filepath.Join("20_derived", "qa"),
	filepath.Join("20_derived", "ops"),
	filepath.Join("30_shared", "meeting"),
	filepath.Join("30_shared", "handoff"),
	filepath.Join("30_shared", "roadmap"),
	filepath.Join("30_shared", "postmortem"),
	filepath.Join("30_shared", "experiment"),
	filepath.Join("30_shared", "maintenance"),
	filepath.Join("30_shared", "errFix"),
	"90_archive",
}

const docsScaffoldReadmeEN = `# AGD Docs Folder Guide

This directory uses a default hierarchy for scalable document management.

- 00_inbox: temporary drafts and triage
- 10_source/*: source-of-truth documents
- 20_derived/*: derived documents linked to source docs
- 30_shared/*: shared planning/coordination docs
- 90_archive: retired documents

Recommended practice:

1) Keep one source doc per topic under 10_source.
2) Keep derived docs under 20_derived and connect source_doc/source_sections.
3) Move outdated docs to 90_archive instead of deleting immediately.

Path references (updated):

- postmortem: 30_shared/postmortem/<file>.agd
- maintenance-case: 30_shared/maintenance/<project>_maintenance_case.agd
- incident-case: 30_shared/errFix/<project>_incident_case.agd

AI planning/writing guide (merged into this README):

1) Core philosophy: control conflict/traceability first, not raw file count.
2) Authority rule: keep one source doc per topic and converge derived docs to source.
3) Prompt contract (required): target section IDs, purpose, constraints, done criteria.
4) AI output (required): updated section IDs, reason, impact, follow-up checklist.
5) Every mutation must append @change with both reason and impact.
6) Keep @map, @section, and @change aligned after each edit.
7) Standard loop: edit -> @change -> map-sync -> check -> source-first conflict resolution.

END__ close rule:

- maintenance close file: 30_shared/maintenance/END__*_maintenance_case.agd
- incident close file: 30_shared/errFix/END__*_incident_case.agd
- these files are excluded from scan/select/role-graph flows
- remove END__ to reactivate a closed case
`

const docsScaffoldReadmeKO = `# AGD 문서 폴더 가이드

이 디렉터리는 확장 가능한 문서 관리를 위한 기본 계층 구조를 사용합니다.

- 00_inbox: 임시 초안 및 분류
- 10_source/*: 기준(source) 문서
- 20_derived/*: source와 연결된 파생 문서
- 30_shared/*: 공유 계획/협업 문서
- 90_archive: 보관 문서

권장 방식:

1) 주제마다 10_source 아래에 기준 문서를 1개 유지합니다.
2) 파생 문서는 20_derived 아래에 두고 source_doc/source_sections를 연결합니다.
3) 오래된 문서는 바로 삭제하지 말고 90_archive로 이동합니다.

경로 참고 (업데이트):

- postmortem: 30_shared/postmortem/<file>.agd
- maintenance-case: 30_shared/maintenance/<project>_maintenance_case.agd
- incident-case: 30_shared/errFix/<project>_incident_case.agd

AI 기획/작성 가이드 (이 README에 통합):

1) 핵심 철학: 파일 수보다 충돌 통제와 추적 가능성을 우선합니다.
2) 권위 규칙: 주제별 source 문서 1개를 유지하고 파생 문서를 source로 수렴합니다.
3) 프롬프트 계약(필수): 대상 섹션 ID, 목적, 제약 조건, 완료 기준.
4) AI 출력(필수): 수정 섹션 ID, reason, impact, 후속 체크리스트.
5) 모든 수정은 @change에 reason과 impact를 함께 기록합니다.
6) 수정 후 @map, @section, @change 정합성을 확인합니다.
7) 표준 루프: 수정 -> @change -> map-sync -> check -> source 우선 충돌 해결.

END__ 종료 규칙:

- 유지보수 종료 문서: 30_shared/maintenance/END__*_maintenance_case.agd
- 오류,버그 종료 문서: 30_shared/errFix/END__*_incident_case.agd
- 이 파일들은 스캔/선택/역할 그래프 흐름에서 자동 제외됩니다.
- 종료 해제 시 END__ 접두어를 제거하면 됩니다.
`

func docsScaffoldReadmeByLang(lang string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "ko") {
		if !isBrokenKOText(docsScaffoldReadmeKO) {
			return docsScaffoldReadmeKO
		}
	}
	return docsScaffoldReadmeEN
}
