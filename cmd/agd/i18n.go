// i18n.go handles language selection and fallback text localization.
// i18n.go는 언어 선택과 텍스트 fallback 현지화를 처리합니다.
package main

import (
	"os"
	"path/filepath"
	"strings"
)

// defaultLang is overridden at build time via:
// go build -ldflags "-X main.defaultLang=ko" ...
var defaultLang = "en"

func activeLang() string {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("AGD_LANG")))
	if env == "ko" || env == "en" {
		return env
	}

	if exeLang := detectLangFromExecutable(); exeLang != "" {
		return exeLang
	}

	lang := strings.ToLower(strings.TrimSpace(defaultLang))
	if lang == "ko" || lang == "en" {
		return lang
	}
	return "en"
}

func detectLangFromExecutable() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	base := strings.ToLower(strings.TrimSpace(filepath.Base(execPath)))
	ext := strings.ToLower(filepath.Ext(base))
	base = strings.TrimSuffix(base, ext)
	switch {
	case strings.HasSuffix(base, "_ko"), strings.HasSuffix(base, "-ko"):
		return "ko"
	case strings.HasSuffix(base, "_en"), strings.HasSuffix(base, "-en"):
		return "en"
	default:
		return ""
	}
}

func isKO() bool {
	return activeLang() == "ko"
}

func isBrokenKOText(value string) bool {
	v := strings.TrimSpace(value)
	if v == "" {
		return true
	}
	if strings.ContainsRune(v, '\uFFFD') {
		return true
	}
	return strings.Count(v, "?") >= 3
}

// Keep exact fallback table for prompts that are often passed with identical en/ko arguments.
var koFallbackExact = map[string]string{
	"Guided flow is on for edit actions: reason+impact required -> map-sync -> check.": "편집 작업에는 가이드 흐름이 적용됩니다: reason+impact 필수 -> map-sync -> check.",
	"AGD Wizard":                             "AGD 마법사",
	"Type a number to choose an action.":     "번호를 입력해 작업을 선택하세요.",
	"[1] Select document":                    "[1] 문서 선택",
	"[2] Generate doc kit":                   "[2] 문서 키트 생성",
	"[3] Validate whole docs tree":           "[3] 전체 문서 트리 검증",
	"[4] New document":                       "[4] 새 문서",
	"[5] Show source/derived relation graph": "[5] source/derived 관계 그래프 보기",
	"[6] Code planning validation":           "[6] 코드 기획 검증",
	"[0] Exit":                               "[0] 종료",
	"Select":                                 "선택",
	"Bye.":                                   "종료합니다.",
	"Invalid choice. Use 0-5.":               "잘못된 선택입니다. 0-5를 사용하세요.",
	"Invalid choice. Use 0-6.":               "잘못된 선택입니다. 0-6을 사용하세요.",

	"Selected document: %s\n":                                 "선택된 문서: %s\n",
	"Selected document -> %s\n":                               "문서 선택 완료 -> %s\n",
	"No document selected. Choose [1] Select document first.": "선택된 문서가 없습니다. 먼저 [1] 문서 선택을 실행하세요.",

	"[1] Search keyword":                     "[1] 키워드 검색",
	"[2] Add new section (also updates map)": "[2] 새 섹션 추가 (map도 함께 갱신)",
	"[3] Add logic-change record":            "[3] 로직 변경 기록 추가",
	"[4] Sync map only":                      "[4] map만 동기화",
	"[5] Check document":                     "[5] 문서 검증",
	"[6] Export to markdown":                 "[6] markdown으로 내보내기",
	"[7] Set source/derived role":            "[7] source/derived 역할 설정",
	"[0] Back":                               "[0] 뒤로가기",
	"Invalid choice. Use 0-7.":               "잘못된 선택입니다. 0-7을 사용하세요.",

	"Document type:": "문서 타입:",
	"Invalid type.":  "잘못된 타입입니다.",
	"types:":         "타입:",

	"File name":                           "파일 이름",
	"Title":                               "제목",
	"Owner":                               "담당자",
	"Root folder (Enter for agd_docs)":    "루트 폴더 (agd_docs 사용 시 Enter)",
	"Strict mode? (y/N)":                  "엄격 모드로 검사할까요? (y/N)",
	"Include 90_archive? (y/N)":           "90_archive를 포함할까요? (y/N)",
	"Service doc (Enter for auto-select)": "서비스 문서 (자동 선택은 Enter)",
	"Max age days (default 0 = must update today)": "최대 허용 경과일 (기본 0 = 오늘 수정 필수)",
	"Allow missing impact? (y/N)":                  "impact 누락을 허용할까요? (y/N)",

	"Kit profile:": "키트 프로필:",
	"  [1] starter-kit         - initial source baseline":                   "  [1] starter-kit         - 초기 기획 시작용 기본 세트",
	"  [2] new-project         - create project-key folder starter docs":    "  [2] new-project         - 프로젝트 키 폴더 기반 신규 기획 세트",
	"  [3] maintenance         - maintenance flow with source section link": "  [3] maintenance         - 유지보수 흐름 + 소스 섹션 파생 연결",
	"  [4] incident            - incident flow with source section link":    "  [4] incident            - 오류 대응 흐름 + 소스 섹션 파생 연결",
	"  [0] back":                            "  [0] 뒤로가기",
	"Profile (1-4 or name)":                 "프로필 (1-4 또는 이름)",
	"Invalid profile.":                      "잘못된 프로필입니다.",
	"Project key (ex: checkout)":            "프로젝트 키 (예: checkout)",
	"Owner (Enter to leave empty)":          "담당자 (비우려면 Enter)",
	"Overwrite existing files? (y/N)":       "기존 파일을 덮어쓸까요? (y/N)",
	"Show role graph after creation? (Y/n)": "생성 후 역할 그래프를 표시할까요? (Y/n)",

	"Root (Enter for agd_docs)":               "루트 (agd_docs 사용 시 Enter)",
	"Change source mode:":                     "변경 소스 모드:",
	"  [1] auto (git -> cache fallback)":      "  [1] auto (git 우선, 실패 시 cache)",
	"  [2] git":                               "  [2] git (현재 워킹트리 기준)",
	"  [3] cache":                             "  [3] cache (로컬 스냅샷 기준)",
	"Mode (Enter=1)":                          "모드 (Enter=1)",
	"Invalid mode. Use 0-3.":                  "잘못된 모드입니다. 0-3을 사용하세요.",
	"Format: [1] text, [2] mermaid, [0] back": "형식: [1] text, [2] mermaid, [0] 뒤로가기",
	"Format (Enter=1)":                        "형식 (Enter=1)",
	"Output file (Enter to print on screen)":  "출력 파일 (화면 출력은 Enter)",
	"Invalid format. Use text or mermaid.":    "잘못된 형식입니다. text 또는 mermaid를 사용하세요.",
	"Keyword":                                 "키워드",
	"Section ID (ex: PRD-020)":                "섹션 ID (예: PRD-020)",
	"Section ID (ex: PRD-050)":                "섹션 ID (예: PRD-050)",
	"New summary":                             "새 요약",
	"Line to append":                          "추가할 문장",
	"Change reason":                           "변경 사유",
	"Change impact":                           "변경 영향",
	"Output .md path (Enter for auto)":        ".md 출력 경로 (자동은 Enter)",
	"Section title":                           "섹션 제목",
	"Section summary":                         "섹션 요약",
	"Links (comma-separated, optional)":       "링크 (쉼표 구분, 선택)",
	"First content line (optional)":           "첫 본문 줄 (선택)",
	"Map sync reason":                         "map 동기화 사유",
	"Map sync impact":                         "map 동기화 영향",
	"[1] source (authoritative)":              "[1] source (기준 문서)",
	"[2] derived (references source)":         "[2] derived (source 참조)",
	"[3] later (skip for now)":                "[3] 나중에 (지금은 건너뛰기)",
	"Role (1-2)":                              "역할 (1-2)",
	"Role (1-3)":                              "역할 (1-3)",
	"Source document (name/path)":             "source 문서 (이름/경로)",
	"Set document role now (recommended):":    "지금 문서 역할을 설정하세요 (권장):",
	"Invalid role.":                           "잘못된 역할입니다.",
	"Guided flow: sync map":                   "가이드 흐름: map 동기화",
	"Guided flow: check document":             "가이드 흐름: 문서 검증",

	"SERVICE CHANGE BLOCKED": "서비스 변경 차단",
	"SERVICE CHANGE ALLOWED": "서비스 변경 허용",
	"Target doc: %s\n":       "대상 문서: %s\n",

	"Suggested source_sections CSV:": "추천 source_sections CSV:",
	"Suggestion details:":            "추천 상세:",
	"relaxed by smart-auto":          "smart-auto 완화 규칙 적용",

	"Role graph after kit creation:": "키트 생성 후 역할 그래프:",
	"Role graph root: %s\n":          "역할 그래프 루트: %s\n",
	"Sources:\n":                     "소스 문서:\n",
	"  (none)\n":                     "  (없음)\n",
	"title:":                         "제목:",
	"derived docs:":                  "파생 문서:",
	"links: (none)":                  "링크: (없음)",
	"source_sections: (none)":        "source_sections: (없음)",
	"source_sections count:":         "source_sections 개수:",
	"Unresolved derived docs:\n":     "해결되지 않은 파생 문서:\n",
	"(empty)":                        "(비어 있음)",
	"source_doc:":                    "source_doc:",
	"reason:":                        "사유:",
	"Unassigned docs (no source/derived role):\n": "미지정 문서 (source/derived 역할 없음):\n",
	"Parse failures:\n":                           "파싱 실패:\n",
	"Read failures:\n":                            "읽기 실패:\n",
}

func koFallbackText(en string) string {
	if v, ok := koFallbackExact[en]; ok {
		return v
	}
	if strings.HasPrefix(en, "usage: ") {
		return strings.Replace(en, "usage:", "사용법:", 1)
	}
	if strings.HasPrefix(en, "error: ") {
		return strings.Replace(en, "error:", "오류:", 1)
	}
	if strings.Contains(en, " error: ") {
		return strings.Replace(en, " error: ", " 오류: ", 1)
	}
	if strings.HasPrefix(en, "unknown command ") {
		return strings.Replace(en, "unknown command ", "알 수 없는 명령어 ", 1)
	}
	if strings.HasPrefix(en, "path is required") {
		return "경로가 필요합니다"
	}
	if strings.HasPrefix(en, "file not found") {
		return strings.Replace(en, "file not found", "파일을 찾을 수 없음", 1)
	}
	if strings.HasPrefix(en, "multiple files matched") {
		return strings.Replace(en, "multiple files matched", "여러 파일이 일치함", 1)
	}
	return en
}

func text(en, ko string) string {
	if isKO() {
		if !isBrokenKOText(ko) && strings.TrimSpace(ko) != strings.TrimSpace(en) {
			return ko
		}
		return koFallbackText(en)
	}
	return en
}
