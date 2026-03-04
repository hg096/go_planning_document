package agd

import "testing"

func TestParseAcceptsUTF8BOM(t *testing.T) {
	raw := "\uFEFF@meta\n" +
		"title: BOM test\n" +
		"version: 0.1\n" +
		"owner: qa\n\n" +
		"@map\n" +
		"- SEC-001 | Sample | Summary\n\n" +
		"@section SEC-001\n" +
		"title: Sample\n" +
		"summary: Sample summary\n" +
		"links: LINK-001\n" +
		"content:\n" +
		"<<<\n" +
		"hello\n" +
		">>>\n"

	doc, errs := Parse(raw)
	if len(errs) != 0 {
		t.Fatalf("expected no parse errors, got: %v", errs)
	}
	if got := doc.Meta["title"]; got != "BOM test" {
		t.Fatalf("unexpected title: got=%q", got)
	}
	if len(doc.Sections) != 1 || doc.Sections[0].ID != "SEC-001" {
		t.Fatalf("expected one section SEC-001, got=%v", doc.Sections)
	}
}
