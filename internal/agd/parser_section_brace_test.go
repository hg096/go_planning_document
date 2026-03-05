package agd

import (
	"strings"
	"testing"
)

func TestParseSectionWithBraceWrapper(t *testing.T) {
	raw := `@meta
title: brace section parse
version: 0.1

@map
- DEL-000 | Guide | Parse braces

@section DEL-000
title: Guide
{
  summary: Parse section fields in brace wrapper
  links: GUIDE-INPUT
  content:
  <<<
line-1
line-2
  >>>
}
`

	doc, errs := Parse(raw)
	if len(errs) != 0 {
		t.Fatalf("expected no parse errors, got: %v", errs)
	}
	if len(doc.Sections) != 1 {
		t.Fatalf("sections len=%d want=1", len(doc.Sections))
	}
	sec := doc.Sections[0]
	if sec.ID != "DEL-000" {
		t.Fatalf("section id=%q", sec.ID)
	}
	if sec.Summary != "Parse section fields in brace wrapper" {
		t.Fatalf("summary=%q", sec.Summary)
	}
	if len(sec.Links) != 1 || sec.Links[0] != "GUIDE-INPUT" {
		t.Fatalf("links=%v", sec.Links)
	}
	if sec.Content != "line-1\nline-2" {
		t.Fatalf("content=%q", sec.Content)
	}
}

func TestSerializeSectionUsesBraceWrapper(t *testing.T) {
	doc := NewDocument()
	doc.Meta["title"] = "serialize brace"
	doc.Meta["version"] = "0.1"
	doc.Map = append(doc.Map, MapEntry{ID: "SEC-001", Title: "Guide", Summary: "Brace output"})
	doc.Sections = append(doc.Sections, &Section{
		ID:      "SEC-001",
		Title:   "Guide",
		Summary: "Brace output",
		Links:   []string{"GUIDE-INPUT"},
		Content: "line-a\nline-b",
	})

	out := Serialize(doc)
	if !strings.Contains(out, "@section SEC-001\ntitle: Guide\n{\n\tsummary: Brace output\n") {
		t.Fatalf("serialized output missing brace section header:\n%s", out)
	}
	if !strings.Contains(out, "\tlinks: GUIDE-INPUT\n\tcontent:\n\t<<<\n\tline-a\n\tline-b\n\t>>>\n}\n") {
		t.Fatalf("serialized output missing brace content block:\n%s", out)
	}
}

func TestSerializeSectionContentTabIndentIsStable(t *testing.T) {
	doc := NewDocument()
	doc.Meta["title"] = "stable tab"
	doc.Meta["version"] = "0.1"
	doc.Map = append(doc.Map, MapEntry{ID: "SEC-001", Title: "Guide", Summary: "Stable"})
	doc.Sections = append(doc.Sections, &Section{
		ID:      "SEC-001",
		Title:   "Guide",
		Summary: "Stable",
		Links:   []string{"GUIDE-INPUT"},
		Content: "line-a\n\tline-b\n\nline-c",
	})

	first := Serialize(doc)
	parsed, errs := Parse(first)
	if len(errs) != 0 {
		t.Fatalf("parse after first serialize failed: %v", errs)
	}
	second := Serialize(parsed)

	if first != second {
		t.Fatalf("serialize should be stable across parse/serialize roundtrip\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if !strings.Contains(first, "\t<<<\n\tline-a\n\tline-b\n\n\tline-c\n\t>>>") {
		t.Fatalf("expected tab-indented content body, got:\n%s", first)
	}
}
