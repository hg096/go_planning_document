package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agd/internal/agd"
)

func TestBuildSectionRelationReportSourceImpact(t *testing.T) {
	repoRoot := detectCodePlanRepoRoot()
	tmpRoot, err := os.MkdirTemp(repoRoot, "tmp_wiz_rel_")
	if err != nil {
		t.Fatalf("mkdir temp failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpRoot)
	})

	docsRoot := filepath.Join(tmpRoot, "00_agd", "agd_docs")
	sourcePath := filepath.Join(docsRoot, "10_source", "service", "checkout_service.agd")
	derivedPath := filepath.Join(docsRoot, "20_derived", "frontend", "checkout_page.agd")
	writeAGDForRelationTest(t, sourcePath, `@meta
title: source
version: 0.1
authority: source
doc_base_path:

@map
- SYS-020 | core flow | source

@section SYS-020
title: core flow
summary: source
links: L
content:
<<<
source
>>>

@change CHG-001
target: SYS-020
date: 2026-03-05
author: test
reason: init
impact: init
`)
	writeAGDForRelationTest(t, derivedPath, `@meta
title: derived
version: 0.1
authority: derived
source_doc: ..\..\10_source\service\checkout_service.agd
source_sections: SYS-020->FP-010
doc_base_path:

@map
- FP-010 | page flow | derived

@section FP-010
title: page flow
summary: derived
links: L
content:
<<<
derived
>>>

@change CHG-001
target: FP-010
date: 2026-03-05
author: test
reason: init
impact: init
`)

	oldDocsRoot := docsRootDir
	docsRootDir = docsRoot
	t.Cleanup(func() { docsRootDir = oldDocsRoot })

	sourceDoc, parseErrs, err := agd.LoadFile(sourcePath)
	if err != nil || len(parseErrs) > 0 {
		t.Fatalf("load source failed err=%v parse=%v", err, parseErrs)
	}

	report, _, err := buildSectionRelationReport(sourcePath, sourceDoc, "SYS-020")
	if err != nil {
		t.Fatalf("build report failed: %v", err)
	}

	if !hasRelationItem(report.Impacted, "20_derived/frontend/checkout_page.agd", "FP-010", "source_sections") {
		t.Fatalf("expected derived impact item, got=%+v", report.Impacted)
	}
}

func TestBuildSectionRelationReportPolicyBidirectional(t *testing.T) {
	repoRoot := detectCodePlanRepoRoot()
	tmpRoot, err := os.MkdirTemp(repoRoot, "tmp_wiz_rel_policy_")
	if err != nil {
		t.Fatalf("mkdir temp failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpRoot)
	})

	docsRoot := filepath.Join(tmpRoot, "00_agd", "agd_docs")
	sourcePath := filepath.Join(docsRoot, "10_source", "service", "checkout_service.agd")
	derivedPath := filepath.Join(docsRoot, "20_derived", "frontend", "checkout_page.agd")
	policyPath := filepath.Join(tmpRoot, "00_agd", "policy", "code_plan_relations.txt")
	writeAGDForRelationTest(t, sourcePath, `@meta
title: source
version: 0.1
authority: source
doc_base_path:

@map
- SYS-020 | core flow | source

@section SYS-020
title: core flow
summary: source
links: L
content:
<<<
source
>>>

@change CHG-001
target: SYS-020
date: 2026-03-05
author: test
reason: init
impact: init
`)
	writeAGDForRelationTest(t, derivedPath, `@meta
title: derived
version: 0.1
authority: derived
source_doc: ..\..\10_source\service\checkout_service.agd
source_sections: SYS-020->FP-010
doc_base_path:

@map
- FP-010 | page flow | derived

@section FP-010
title: page flow
summary: derived
links: L
content:
<<<
derived
>>>

@change CHG-001
target: FP-010
date: 2026-03-05
author: test
reason: init
impact: init
`)
	if err := os.MkdirAll(filepath.Dir(policyPath), 0o755); err != nil {
		t.Fatalf("mkdir policy dir failed: %v", err)
	}
	sourceRel := codePlanRelPathFromRoot(repoRoot, sourcePath)
	derivedRel := codePlanRelPathFromRoot(repoRoot, derivedPath)
	line := sourceRel + "#SYS-020 <-> " + derivedRel + "#FP-010\n"
	if err := os.WriteFile(policyPath, []byte(line), 0o644); err != nil {
		t.Fatalf("write relation policy failed: %v", err)
	}

	oldDocsRoot := docsRootDir
	oldPolicyPath := codePlanDefaultRelationPolicyPath
	docsRootDir = docsRoot
	codePlanDefaultRelationPolicyPath = policyPath
	t.Cleanup(func() {
		docsRootDir = oldDocsRoot
		codePlanDefaultRelationPolicyPath = oldPolicyPath
	})

	derivedDoc, parseErrs, err := agd.LoadFile(derivedPath)
	if err != nil || len(parseErrs) > 0 {
		t.Fatalf("load derived failed err=%v parse=%v", err, parseErrs)
	}

	report, _, err := buildSectionRelationReport(derivedPath, derivedDoc, "FP-010")
	if err != nil {
		t.Fatalf("build report failed: %v", err)
	}

	if !hasRelationItem(report.Connected, "10_source/service/checkout_service.agd", "SYS-020", "relation-policy") {
		t.Fatalf("expected policy connection item, got=%+v", report.Connected)
	}
	if !hasRelationItem(report.Impacted, "10_source/service/checkout_service.agd", "SYS-020", "relation-policy") {
		t.Fatalf("expected policy impact item, got=%+v", report.Impacted)
	}
}

func writeAGDForRelationTest(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func hasRelationItem(items []sectionRelationItem, docSuffix, sectionID, via string) bool {
	wantSec := strings.ToUpper(strings.TrimSpace(sectionID))
	wantVia := strings.TrimSpace(via)
	for _, item := range items {
		if !strings.HasSuffix(strings.ToLower(filepath.ToSlash(item.DocRel)), strings.ToLower(filepath.ToSlash(docSuffix))) {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(item.SectionID)) != wantSec {
			continue
		}
		if strings.TrimSpace(item.Via) != wantVia {
			continue
		}
		return true
	}
	return false
}
