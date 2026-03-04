package agd

import (
	"strings"
	"testing"
)

func TestParseAndResolveSectionPath(t *testing.T) {
	raw := `@meta
title: path test
version: 0.1
doc_base_path: 10_source/product/core_spec_checkout.agd

@map
- CORE-010 | Intro | Intro summary

@section CORE-010
title: Intro
summary: Intro summary
path: 10_source/product/core_spec_checkout.agd#CORE-010
links: DOC-001
content:
<<<
hello
>>>

@change CHG-001
target: CORE-010
date: 2026-03-04
author: ai-agent
reason: init
impact: none
`

	doc, errs := Parse(raw)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errs: %v", errs)
	}

	if len(doc.Sections) != 1 {
		t.Fatalf("sections len=%d want=1", len(doc.Sections))
	}
	sec := doc.Sections[0]
	if sec.Path != "10_source/product/core_spec_checkout.agd#CORE-010" {
		t.Fatalf("section path=%q", sec.Path)
	}

	byPath := doc.SectionByID("10_source/product/core_spec_checkout.agd#CORE-010")
	if byPath == nil || byPath.ID != "CORE-010" {
		t.Fatalf("SectionByID(path) failed: %#v", byPath)
	}

	byID := doc.SectionByID("CORE-010")
	if byID == nil || byID.ID != "CORE-010" {
		t.Fatalf("SectionByID(id) failed: %#v", byID)
	}
}

func TestSerializeIncludesPathWhenPresent(t *testing.T) {
	doc := NewDocument()
	doc.Meta["title"] = "serialize path"
	doc.Meta["version"] = "0.1"
	doc.Map = append(doc.Map, MapEntry{ID: "SEC-001", Title: "T", Summary: "S"})
	doc.Sections = append(doc.Sections, &Section{
		ID:      "SEC-001",
		Title:   "T",
		Summary: "S",
		Path:    "10_source/product/x.agd#SEC-001",
		Content: "body",
	})
	doc.Changes = append(doc.Changes, &Change{
		ID:     "CHG-001",
		Target: "SEC-001",
		Date:   "2026-03-04",
		Author: "ai-agent",
		Reason: "r",
		Impact: "i",
	})

	out := Serialize(doc)
	if !strings.Contains(out, "path: 10_source/product/x.agd#SEC-001") {
		t.Fatalf("serialized output missing path field:\n%s", out)
	}
}
