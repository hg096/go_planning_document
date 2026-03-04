package main

import (
	"strings"
	"testing"
	"time"

	"agd/internal/agd"
)

func TestCodePlanPathHasPrefixBoundary(t *testing.T) {
	if !codePlanPathHasPrefix("src/a/file.go", "src/a") {
		t.Fatalf("expected prefix match")
	}
	if codePlanPathHasPrefix("src/ab/file.go", "src/a") {
		t.Fatalf("unexpected prefix match for boundary case")
	}
	if !codePlanPathHasPrefix("src/a", "src/a") {
		t.Fatalf("expected exact match")
	}
}

func TestCodePlanPathsOverlap(t *testing.T) {
	if !codePlanPathsOverlap("src/service/api/handler.go", "src/service/api") {
		t.Fatalf("expected folder overlap")
	}
	if !codePlanPathsOverlap("src/service/api/handler.go", "src/service/api/handler.go") {
		t.Fatalf("expected exact overlap")
	}
	if codePlanPathsOverlap("src/service/api.go", "src/service/api") {
		t.Fatalf("unexpected overlap for sibling name")
	}
}

func TestHasCodePlanTag(t *testing.T) {
	docMetaTag := agd.NewDocument()
	docMetaTag.Meta["incident_feature_tag"] = "FT-CHECKOUT"
	if !hasCodePlanTag(docMetaTag) {
		t.Fatalf("meta incident tag should be detected")
	}

	docBlockTag := agd.NewDocument()
	docBlockTag.Sections = append(docBlockTag.Sections, &agd.Section{
		ID: "INC-001",
		Content: `[INCIDENT-PROBLEM-TAG]
- feature_tag: FT-X
[/INCIDENT-PROBLEM-TAG]`,
	})
	if !hasCodePlanTag(docBlockTag) {
		t.Fatalf("section block tag should be detected")
	}

	docNoTag := agd.NewDocument()
	docNoTag.Sections = append(docNoTag.Sections, &agd.Section{
		ID:      "A-001",
		Content: "plain content",
	})
	if hasCodePlanTag(docNoTag) {
		t.Fatalf("no tag should not be detected")
	}
}

func TestParseCodePlanScopePolicyLines(t *testing.T) {
	policy := codePlanScopePolicy{
		AllowedExt: defaultCodePlanAllowedExt(),
	}
	warnings := parseCodePlanScopePolicyLines([]string{
		"ext:.vue",
		"include:src/generated/",
		"exclude:vendor/",
		"unknown:token",
	}, &policy)

	if !policy.AllowedExt[".vue"] {
		t.Fatalf("expected .vue extension to be added")
	}
	if len(policy.IncludePrefixes) != 1 || !strings.HasPrefix(policy.IncludePrefixes[0], "src/generated") {
		t.Fatalf("include prefix parse mismatch: %+v", policy.IncludePrefixes)
	}
	if len(policy.ExcludePrefixes) != 1 || !strings.HasPrefix(policy.ExcludePrefixes[0], "vendor") {
		t.Fatalf("exclude prefix parse mismatch: %+v", policy.ExcludePrefixes)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings=%d want=1", len(warnings))
	}
}

func TestIsCodePlanFileInScope(t *testing.T) {
	policy := codePlanScopePolicy{
		AllowedExt:      defaultCodePlanAllowedExt(),
		IncludePrefixes: []string{"src/generated/"},
		ExcludePrefixes: []string{"vendor/"},
	}

	cases := []struct {
		path string
		want bool
	}{
		{path: "src/app/main.go", want: true},
		{path: "src/generated/schema.txt", want: true},
		{path: "vendor/lib/main.go", want: false},
		{path: "00_agd/scripts/setup.ps1", want: false},
		{path: "docs/readme.md", want: false},
		{path: "src/plan/spec.agd", want: false},
	}

	for _, tc := range cases {
		got := isCodePlanFileInScope(tc.path, policy)
		if got != tc.want {
			t.Fatalf("scope(%s)=%t want=%t", tc.path, got, tc.want)
		}
	}
}

func TestParseGitStatusPorcelainZ(t *testing.T) {
	raw := []byte(" M main.go\x00?? new.ts\x00R  old.go\x00new.go\x00D  gone.go\x00")
	entries := parseGitStatusPorcelainZ(raw)
	if len(entries) != 4 {
		t.Fatalf("entries=%d want=4", len(entries))
	}
	if entries[0].Path != "main.go" {
		t.Fatalf("entry0 path=%q", entries[0].Path)
	}
	if entries[1].Status != "??" || entries[1].Path != "new.ts" {
		t.Fatalf("entry1 mismatch: %+v", entries[1])
	}
	if entries[2].Status[0] != 'R' || entries[2].Path != "new.go" {
		t.Fatalf("entry2 rename mismatch: %+v", entries[2])
	}
	if !codePlanIsDeletedStatus(entries[3].Status) {
		t.Fatalf("entry3 should be deleted status")
	}
}

func TestIsCodePlanChangeUpdated(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	doc := codePlanDocInfo{
		RelPath:          "00_agd/agd_docs/10_source/product/core.agd",
		ChangeDigest:     "abc",
		LatestChangeDate: today,
	}
	if !isCodePlanChangeUpdated(doc, nil, today) {
		t.Fatalf("without cache and today change should be updated")
	}

	cacheSame := &codePlanCache{
		DocSnapshot: map[string]codePlanDocSnapshot{
			codePlanPathKey(doc.RelPath): {ChangeDigest: "abc"},
		},
	}
	if isCodePlanChangeUpdated(doc, cacheSame, today) {
		t.Fatalf("same digest should not be updated")
	}

	cacheDiff := &codePlanCache{
		DocSnapshot: map[string]codePlanDocSnapshot{
			codePlanPathKey(doc.RelPath): {ChangeDigest: "xyz"},
		},
	}
	if !isCodePlanChangeUpdated(doc, cacheDiff, today) {
		t.Fatalf("different digest should be updated")
	}
}
