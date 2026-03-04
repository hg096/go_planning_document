// section_paths.go manages doc/section path metadata defaults.
// section_paths.go는 문서/섹션 경로 메타데이터 기본값을 관리합니다.
package main

import (
	"strings"

	"agd/internal/agd"
)

func defaultSectionPath(docBasePath, sectionID string) string {
	base := strings.TrimSpace(docBasePath)
	sec := strings.ToUpper(strings.TrimSpace(sectionID))
	if base == "" || sec == "" {
		return ""
	}
	return base + "#" + sec
}

func ensureDocPathMetadata(doc *agd.Document, filePath string) bool {
	_ = filePath
	if doc == nil {
		return false
	}
	if doc.Meta == nil {
		doc.Meta = map[string]string{}
	}

	changed := false
	if _, ok := doc.Meta["doc_base_path"]; !ok {
		// Keep initial value empty by default.
		doc.Meta["doc_base_path"] = ""
		changed = true
	}
	base := strings.TrimSpace(doc.Meta["doc_base_path"])

	for _, section := range doc.Sections {
		if section == nil {
			continue
		}
		if strings.TrimSpace(section.Path) == "" {
			section.Path = defaultSectionPath(base, section.ID)
			if strings.TrimSpace(section.Path) != "" {
				changed = true
			}
		}
	}
	return changed
}
