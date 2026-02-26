package agd

import (
	"fmt"
	"strings"
)

func ToMarkdown(doc *Document) string {
	title := doc.Meta["title"]
	if strings.TrimSpace(title) == "" {
		title = "Untitled AGD"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if len(doc.Meta) > 0 {
		sb.WriteString("## Meta\n\n")
		for _, key := range orderedMetaKeys(doc.Meta) {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", key, doc.Meta[key]))
		}
		sb.WriteString("\n")
	}

	if len(doc.Map) > 0 {
		sb.WriteString("## Map\n\n")
		sb.WriteString("| ID | Title | Summary |\n")
		sb.WriteString("| --- | --- | --- |\n")
		for _, row := range doc.Map {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", row.ID, row.Title, row.Summary))
		}
		sb.WriteString("\n")
	}

	if len(doc.Sections) > 0 {
		sb.WriteString("## Sections\n\n")
		for _, section := range doc.Sections {
			sb.WriteString(fmt.Sprintf("### %s - %s\n\n", section.ID, section.Title))
			sb.WriteString(fmt.Sprintf("**Summary**: %s\n\n", section.Summary))
			if len(section.Links) > 0 {
				sb.WriteString(fmt.Sprintf("**Links**: %s\n\n", strings.Join(section.Links, ", ")))
			}
			if strings.TrimSpace(section.Content) != "" {
				sb.WriteString(section.Content + "\n\n")
			}
		}
	}

	if len(doc.Changes) > 0 {
		sb.WriteString("## Changes\n\n")
		for _, change := range doc.Changes {
			sb.WriteString(fmt.Sprintf("- **%s** (%s) target `%s`: %s\n", change.ID, change.Date, change.Target, change.Reason))
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n") + "\n"
}
