package agd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type block struct {
	Kind   string
	Arg    string
	Lines  []string
	LineNo int
}

var blockHeaderPattern = regexp.MustCompile(`^@([A-Za-z]+)(?:\s+(.+))?$`)

func LoadFile(path string) (*Document, []string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	doc, parseErrs := Parse(string(raw))
	return doc, parseErrs, nil
}

func Parse(raw string) (*Document, []string) {
	doc := NewDocument()
	errors := make([]string, 0)

	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.TrimPrefix(normalized, "\uFEFF")
	lines := strings.Split(normalized, "\n")

	blocks := make([]*block, 0)
	var current *block

	inSectionContent := false
	expectSectionFence := false

	startBlock := func(kind, arg string, lineNo int) {
		current = &block{
			Kind:   strings.ToLower(strings.TrimSpace(kind)),
			Arg:    strings.TrimSpace(arg),
			Lines:  make([]string, 0),
			LineNo: lineNo,
		}
		blocks = append(blocks, current)
		inSectionContent = false
		expectSectionFence = false
	}

	for idx, line := range lines {
		lineNo := idx + 1
		trimmed := strings.TrimSpace(line)

		if current != nil && current.Kind == "section" {
			if inSectionContent {
				current.Lines = append(current.Lines, line)
				if trimmed == ">>>" {
					inSectionContent = false
				}
				continue
			}
			if expectSectionFence {
				current.Lines = append(current.Lines, line)
				if trimmed == "<<<" {
					inSectionContent = true
				}
				expectSectionFence = false
				continue
			}
		}

		if len(line) > 0 && line[0] == '@' {
			match := blockHeaderPattern.FindStringSubmatch(strings.TrimSpace(line))
			if len(match) == 0 {
				errors = append(errors, fmt.Sprintf("line %d: invalid block header %q", lineNo, line))
				continue
			}
			arg := ""
			if len(match) > 2 {
				arg = match[2]
			}
			startBlock(match[1], arg, lineNo)
			continue
		}

		if current == nil {
			if trimmed == "" {
				continue
			}
			errors = append(errors, fmt.Sprintf("line %d: content outside block", lineNo))
			continue
		}

		current.Lines = append(current.Lines, line)
		if current.Kind == "section" && strings.EqualFold(trimmed, "content:") {
			expectSectionFence = true
		}
	}

	for _, b := range blocks {
		switch b.Kind {
		case "meta":
			for _, e := range parseMetaBlock(doc, b) {
				errors = append(errors, e)
			}
		case "map":
			for _, e := range parseMapBlock(doc, b) {
				errors = append(errors, e)
			}
		case "section":
			for _, e := range parseSectionBlock(doc, b) {
				errors = append(errors, e)
			}
		case "change":
			for _, e := range parseChangeBlock(doc, b) {
				errors = append(errors, e)
			}
		default:
			errors = append(errors, fmt.Sprintf("line %d: unsupported block @%s", b.LineNo, b.Kind))
		}
	}

	return doc, errors
}

func Serialize(doc *Document) string {
	var sb strings.Builder

	sb.WriteString("@meta\n")
	for _, key := range orderedMetaKeys(doc.Meta) {
		sb.WriteString(fmt.Sprintf("%s: %s\n", key, strings.TrimSpace(doc.Meta[key])))
	}
	sb.WriteString("\n")

	sb.WriteString("@map\n")
	for _, entry := range doc.Map {
		sb.WriteString(fmt.Sprintf("- %s | %s | %s\n", strings.TrimSpace(entry.ID), strings.TrimSpace(entry.Title), strings.TrimSpace(entry.Summary)))
	}
	sb.WriteString("\n")

	for _, section := range doc.Sections {
		sb.WriteString(fmt.Sprintf("@section %s\n", strings.TrimSpace(section.ID)))
		sb.WriteString(fmt.Sprintf("title: %s\n", strings.TrimSpace(section.Title)))
		sb.WriteString("{\n")
		sb.WriteString(fmt.Sprintf("\tsummary: %s\n", strings.TrimSpace(section.Summary)))
		path := strings.TrimSpace(section.Path)
		if path == "" {
			sb.WriteString("\tpath:\n")
		} else {
			sb.WriteString(fmt.Sprintf("\tpath: %s\n", path))
		}
		sb.WriteString(fmt.Sprintf("\tlinks: %s\n", strings.Join(section.Links, ", ")))
		sb.WriteString("\tcontent:\n")
		sb.WriteString("\t<<<\n")
		serializedContent := serializeSectionContentWithTab(section.Content)
		if strings.TrimSpace(serializedContent) != "" {
			sb.WriteString(serializedContent)
			sb.WriteString("\n")
		}
		sb.WriteString("\t>>>\n")
		sb.WriteString("}\n\n")
	}

	for _, change := range doc.Changes {
		sb.WriteString(fmt.Sprintf("@change %s\n", strings.TrimSpace(change.ID)))
		sb.WriteString(fmt.Sprintf("target: %s\n", strings.TrimSpace(change.Target)))
		sb.WriteString(fmt.Sprintf("date: %s\n", strings.TrimSpace(change.Date)))
		sb.WriteString(fmt.Sprintf("author: %s\n", strings.TrimSpace(change.Author)))
		sb.WriteString(fmt.Sprintf("reason: %s\n", strings.TrimSpace(change.Reason)))
		if strings.TrimSpace(change.Impact) != "" {
			sb.WriteString(fmt.Sprintf("impact: %s\n", strings.TrimSpace(change.Impact)))
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n") + "\n"
}

func parseMetaBlock(doc *Document, b *block) []string {
	errors := make([]string, 0)
	for i, line := range b.Lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := splitKeyValue(trimmed)
		if !ok {
			errors = append(errors, fmt.Sprintf("line %d: invalid meta entry %q", b.LineNo+i+1, line))
			continue
		}
		doc.Meta[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	return errors
}

func parseMapBlock(doc *Document, b *block) []string {
	errors := make([]string, 0)
	for i, line := range b.Lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			errors = append(errors, fmt.Sprintf("line %d: invalid map row %q", b.LineNo+i+1, line))
			continue
		}
		parts := strings.SplitN(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), "|", 3)
		if len(parts) < 2 {
			errors = append(errors, fmt.Sprintf("line %d: map row requires at least id and title", b.LineNo+i+1))
			continue
		}
		entry := MapEntry{
			ID:    strings.TrimSpace(parts[0]),
			Title: strings.TrimSpace(parts[1]),
		}
		if len(parts) == 3 {
			entry.Summary = strings.TrimSpace(parts[2])
		}
		doc.Map = append(doc.Map, entry)
	}
	return errors
}

func parseSectionBlock(doc *Document, b *block) []string {
	errors := make([]string, 0)
	sectionID := strings.TrimSpace(b.Arg)
	if sectionID == "" {
		return []string{fmt.Sprintf("line %d: section block requires id, e.g. @section SEC-001", b.LineNo)}
	}

	fields, content, fieldErrs := parseFieldsWithContent(b)
	errors = append(errors, fieldErrs...)

	section := &Section{
		ID:      sectionID,
		Title:   fields["title"],
		Summary: fields["summary"],
		Path:    resolveSectionPathField(fields),
		Links:   parseCSV(fields["links"]),
		Content: content,
	}
	doc.Sections = append(doc.Sections, section)
	return errors
}

func resolveSectionPathField(fields map[string]string) string {
	if fields == nil {
		return ""
	}
	path := strings.TrimSpace(fields["path"])
	if path != "" {
		return path
	}
	return strings.TrimSpace(fields["section_path"])
}

func parseChangeBlock(doc *Document, b *block) []string {
	errors := make([]string, 0)
	changeID := strings.TrimSpace(b.Arg)
	if changeID == "" {
		return []string{fmt.Sprintf("line %d: change block requires id, e.g. @change CHG-20260223-01", b.LineNo)}
	}

	fields, _, fieldErrs := parseFieldsWithContent(b)
	errors = append(errors, fieldErrs...)

	change := &Change{
		ID:     changeID,
		Target: fields["target"],
		Date:   fields["date"],
		Author: fields["author"],
		Reason: fields["reason"],
		Impact: fields["impact"],
	}
	doc.Changes = append(doc.Changes, change)
	return errors
}

func parseFieldsWithContent(b *block) (map[string]string, string, []string) {
	fields := make(map[string]string)
	errors := make([]string, 0)
	content := ""

	i := 0
	for i < len(b.Lines) {
		line := b.Lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			i++
			continue
		}
		if trimmed == "{" || trimmed == "}" {
			i++
			continue
		}

		if strings.EqualFold(trimmed, "content:") {
			i++
			for i < len(b.Lines) && strings.TrimSpace(b.Lines[i]) == "" {
				i++
			}
			if i >= len(b.Lines) || strings.TrimSpace(b.Lines[i]) != "<<<" {
				errors = append(errors, fmt.Sprintf("line %d: content block must start with <<<", b.LineNo+i+1))
				continue
			}
			i++

			contentLines := make([]string, 0)
			for i < len(b.Lines) && strings.TrimSpace(b.Lines[i]) != ">>>" {
				contentLines = append(contentLines, b.Lines[i])
				i++
			}
			if i >= len(b.Lines) {
				errors = append(errors, fmt.Sprintf("line %d: content block missing >>> terminator", b.LineNo))
				break
			}
			content = strings.TrimRight(strings.Join(contentLines, "\n"), "\n")
			i++
			continue
		}

		key, value, ok := splitKeyValue(trimmed)
		if !ok {
			errors = append(errors, fmt.Sprintf("line %d: invalid field %q", b.LineNo+i+1, line))
			i++
			continue
		}
		fields[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
		i++
	}

	return fields, content, errors
}

func splitKeyValue(line string) (string, string, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false
	}
	return key, strings.TrimSpace(parts[1]), true
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func serializeSectionContentWithTab(content string) string {
	trimmed := strings.TrimRight(content, "\n")
	if strings.TrimSpace(trimmed) == "" {
		return ""
	}

	lines := strings.Split(trimmed, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			normalized = append(normalized, line)
			continue
		}
		if strings.HasPrefix(line, "\t") {
			normalized = append(normalized, line)
			continue
		}
		normalized = append(normalized, "\t"+line)
	}
	return strings.Join(normalized, "\n")
}

func orderedMetaKeys(meta map[string]string) []string {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	priority := []string{"title", "version", "owner"}
	ordered := make([]string, 0, len(keys))

	seen := make(map[string]bool)
	for _, key := range priority {
		if _, ok := meta[key]; ok {
			ordered = append(ordered, key)
			seen[key] = true
		}
	}
	for _, key := range keys {
		if !seen[key] {
			ordered = append(ordered, key)
		}
	}
	return ordered
}
