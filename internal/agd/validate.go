package agd

import "fmt"

func Validate(doc *Document) []string {
	errors := make([]string, 0)

	if doc == nil {
		return []string{"document is nil"}
	}
	if doc.Meta["title"] == "" {
		errors = append(errors, "meta.title is required")
	}
	if doc.Meta["version"] == "" {
		errors = append(errors, "meta.version is required")
	}

	mapSeen := make(map[string]bool)
	for i, row := range doc.Map {
		id := normalizeID(row.ID)
		if id == "" {
			errors = append(errors, fmt.Sprintf("map row #%d has empty id", i+1))
			continue
		}
		if mapSeen[id] {
			errors = append(errors, fmt.Sprintf("duplicate map id: %s", row.ID))
		}
		mapSeen[id] = true
	}

	sectionSeen := make(map[string]bool)
	for _, section := range doc.Sections {
		id := normalizeID(section.ID)
		if id == "" {
			errors = append(errors, "section id cannot be empty")
			continue
		}
		if sectionSeen[id] {
			errors = append(errors, fmt.Sprintf("duplicate section id: %s", section.ID))
		}
		sectionSeen[id] = true

		if section.Title == "" {
			errors = append(errors, fmt.Sprintf("section %s is missing title", section.ID))
		}
		if section.Summary == "" {
			errors = append(errors, fmt.Sprintf("section %s is missing summary", section.ID))
		}
	}

	for _, row := range doc.Map {
		if !sectionSeen[normalizeID(row.ID)] {
			errors = append(errors, fmt.Sprintf("map id %s does not match any section", row.ID))
		}
	}
	for _, section := range doc.Sections {
		if !mapSeen[normalizeID(section.ID)] {
			errors = append(errors, fmt.Sprintf("section %s is not listed in @map", section.ID))
		}
	}

	changeSeen := make(map[string]bool)
	for _, change := range doc.Changes {
		id := normalizeID(change.ID)
		if id == "" {
			errors = append(errors, "change id cannot be empty")
			continue
		}
		if changeSeen[id] {
			errors = append(errors, fmt.Sprintf("duplicate change id: %s", change.ID))
		}
		changeSeen[id] = true

		targetID := normalizeID(change.Target)
		if targetID == "" {
			errors = append(errors, fmt.Sprintf("change %s has empty target", change.ID))
		} else if !sectionSeen[targetID] {
			errors = append(errors, fmt.Sprintf("change %s targets unknown section %s", change.ID, change.Target))
		}

		if change.Reason == "" {
			errors = append(errors, fmt.Sprintf("change %s is missing reason", change.ID))
		}
		if change.Impact == "" {
			errors = append(errors, fmt.Sprintf("change %s is missing impact", change.ID))
		}
	}

	return errors
}
