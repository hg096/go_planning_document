package main

import (
	"testing"

	"agd/internal/agd"
)

func TestEnsureDocPathMetadata_DefaultEmptyBasePath(t *testing.T) {
	doc := agd.NewDocument()
	doc.Meta["title"] = "x"
	doc.Meta["version"] = "0.1"
	doc.Sections = append(doc.Sections, &agd.Section{
		ID:      "SEC-001",
		Title:   "t",
		Summary: "s",
	})

	changed := ensureDocPathMetadata(doc, `00_agd\agd_docs\10_source\product\x.agd`)
	if !changed {
		t.Fatalf("expected metadata change")
	}
	if got := doc.Meta["doc_base_path"]; got != "" {
		t.Fatalf("doc_base_path=%q want empty", got)
	}
	if got := doc.Sections[0].Path; got != "" {
		t.Fatalf("section path=%q want empty", got)
	}
}

func TestEnsureDocPathMetadata_FillSectionPathWhenBaseExists(t *testing.T) {
	doc := agd.NewDocument()
	doc.Meta["title"] = "x"
	doc.Meta["version"] = "0.1"
	doc.Meta["doc_base_path"] = "10_source/product/x.agd"
	doc.Sections = append(doc.Sections, &agd.Section{
		ID:      "SEC-001",
		Title:   "t",
		Summary: "s",
	})

	changed := ensureDocPathMetadata(doc, `ignored`)
	if !changed {
		t.Fatalf("expected metadata change")
	}
	if got := doc.Sections[0].Path; got != "10_source/product/x.agd#SEC-001" {
		t.Fatalf("section path=%q", got)
	}
}
