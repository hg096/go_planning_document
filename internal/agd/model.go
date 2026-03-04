package agd

import "strings"

type Document struct {
	Meta     map[string]string
	Map      []MapEntry
	Sections []*Section
	Changes  []*Change
}

type MapEntry struct {
	ID      string
	Title   string
	Summary string
}

type Section struct {
	ID      string
	Title   string
	Summary string
	Path    string
	Links   []string
	Content string
}

type Change struct {
	ID     string
	Target string
	Date   string
	Author string
	Reason string
	Impact string
}

type SearchHit struct {
	Kind    string
	ID      string
	Field   string
	Snippet string
}

func NewDocument() *Document {
	return &Document{
		Meta:     make(map[string]string),
		Map:      make([]MapEntry, 0),
		Sections: make([]*Section, 0),
		Changes:  make([]*Change, 0),
	}
}

func (d *Document) SectionByID(id string) *Section {
	raw := strings.TrimSpace(id)
	if isSectionPathRef(raw) {
		needlePath := normalizeSectionPath(raw)
		for _, section := range d.Sections {
			if normalizeSectionPath(section.Path) == needlePath {
				return section
			}
		}
	}
	needle := normalizeID(raw)
	for _, section := range d.Sections {
		if normalizeID(section.ID) == needle {
			return section
		}
	}
	return nil
}

func (d *Document) MapByID(id string) *MapEntry {
	needle := normalizeID(id)
	for i := range d.Map {
		if normalizeID(d.Map[i].ID) == needle {
			return &d.Map[i]
		}
	}
	return nil
}

func normalizeID(id string) string {
	return strings.ToUpper(strings.TrimSpace(id))
}

func isSectionPathRef(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "#") || strings.Contains(trimmed, "/") || strings.Contains(trimmed, `\`)
}

func normalizeSectionPath(path string) string {
	normalized := strings.TrimSpace(path)
	normalized = strings.ReplaceAll(normalized, `\`, "/")
	normalized = strings.TrimPrefix(normalized, "./")
	return strings.ToLower(normalized)
}
