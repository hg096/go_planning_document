package agdtemplates

import (
	"strings"
	"testing"

	"agd/internal/agd"
)

func TestTemplatesParseValidateAndUseSectionBraceWrapper(t *testing.T) {
	names := Names()
	if len(names) == 0 {
		t.Fatalf("expected non-empty template names")
	}

	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			raw, err := Load(name)
			if err != nil {
				t.Fatalf("load failed: %v", err)
			}
			assertSectionBraceWrapper(t, raw)

			rendered, err := Render(name, RenderOptions{
				Title:   "Template Smoke Test",
				Owner:   "qa",
				Version: "0.1",
				Date:    "2026-03-05",
			})
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			doc, parseErrs := agd.Parse(rendered)
			if len(parseErrs) > 0 {
				t.Fatalf("parse errs: %s", strings.Join(parseErrs, "; "))
			}

			validateErrs := agd.Validate(doc)
			if len(validateErrs) > 0 {
				t.Fatalf("validate errs: %s", strings.Join(validateErrs, "; "))
			}
		})
	}
}

func assertSectionBraceWrapper(t *testing.T, raw string) {
	t.Helper()

	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	sectionCount := 0
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, "@section ") {
			continue
		}
		sectionCount++

		// Rule 1: after section title line, the next non-empty line must be "{"
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		if j >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[j]), "title:") {
			t.Fatalf("section must include title line before brace wrapper: section header line=%d", i+1)
		}

		legacyIdx := j + 1
		for legacyIdx < len(lines) && strings.TrimSpace(lines[legacyIdx]) == "" {
			legacyIdx++
		}
		if legacyIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[legacyIdx]), "summary:") {
			t.Fatalf("legacy section format detected (summary immediately after title) near line=%d", legacyIdx+1)
		}
		if legacyIdx >= len(lines) || strings.TrimSpace(lines[legacyIdx]) != "{" {
			t.Fatalf("section brace wrapper missing near line=%d", legacyIdx+1)
		}

		// Rule 2: each section block must include a closing "}" before next block header.
		hasClose := false
		inContentBody := false
		for k := legacyIdx + 1; k < len(lines); k++ {
			lineRaw := lines[k]
			lineTrim := strings.TrimSpace(lines[k])
			if strings.HasPrefix(lineTrim, "@section ") || strings.HasPrefix(lineTrim, "@change ") || lineTrim == "@meta" || lineTrim == "@map" {
				break
			}
			if lineTrim == "" {
				continue
			}

			if lineTrim == "}" {
				hasClose = true
				break
			}

			if inContentBody {
				if lineTrim == ">>>" {
					assertTabbedSectionLine(t, lineRaw, lineTrim, k+1)
					inContentBody = false
					continue
				}
				if lineTrim != "" {
					assertTabbedSectionLine(t, lineRaw, lineTrim, k+1)
				}
				continue
			}

			if isSectionFieldOrFence(lineTrim) {
				assertTabbedSectionLine(t, lineRaw, lineTrim, k+1)
				if lineTrim == "<<<" {
					inContentBody = true
				}
			}
		}
		if !hasClose {
			t.Fatalf("section brace wrapper closing '}' missing: section header line=%d", i+1)
		}
	}

	if sectionCount == 0 {
		t.Fatalf("template must contain at least one @section block")
	}
}

func isSectionFieldOrFence(trimmed string) bool {
	return strings.HasPrefix(trimmed, "summary:") ||
		strings.HasPrefix(trimmed, "path:") ||
		strings.HasPrefix(trimmed, "links:") ||
		trimmed == "content:" ||
		trimmed == "<<<" ||
		trimmed == ">>>"
}

func assertTabbedSectionLine(t *testing.T, rawLine, trimmed string, lineNo int) {
	t.Helper()

	if strings.HasPrefix(rawLine, "  ") {
		t.Fatalf("legacy 2-space indentation detected at line=%d (%s)", lineNo, trimmed)
	}
	if !strings.HasPrefix(rawLine, "\t") {
		t.Fatalf("section line must start with tab indentation at line=%d (%s)", lineNo, trimmed)
	}
}
