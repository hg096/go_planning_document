package main

import (
	"path/filepath"
	"testing"
)

func TestKitProfileSpecsStarterKitUsesUnifiedTemplates(t *testing.T) {
	specs := kitProfileSpecs("starter-kit")
	if len(specs) != 4 {
		t.Fatalf("starter-kit should generate 4 docs, got=%d", len(specs))
	}

	byKey := map[string]kitDocSpec{}
	for _, spec := range specs {
		byKey[spec.Key] = spec
	}

	serviceDoc, ok := byKey["service"]
	if !ok || serviceDoc.Role != "source" {
		t.Fatalf("service doc must exist as source, got=%+v", serviceDoc)
	}

	if _, ok := byKey["delivery_plan"]; ok {
		t.Fatalf("starter-kit should not generate delivery_plan")
	}
	architectureDoc, ok := byKey["architecture"]
	if !ok {
		t.Fatalf("starter-kit should generate architecture example doc")
	}
	if architectureDoc.Subdir != filepath.Join("10_core_logic", "architecture") {
		t.Fatalf("architecture doc should be under core logic architecture folder, got=%q", architectureDoc.Subdir)
	}
	if architectureDoc.Role != "source" {
		t.Fatalf("architecture doc must be source, got=%q", architectureDoc.Role)
	}

	if _, ok := byKey["prd"]; ok {
		t.Fatalf("legacy prd doc should not be generated in starter-kit")
	}
	if _, ok := byKey["core_spec"]; ok {
		t.Fatalf("starter-kit should not generate product core_spec")
	}
}

func TestKitProfileSpecsNewProjectUsesFocusedDocTypes(t *testing.T) {
	specs := kitProfileSpecs("new-project")
	if len(specs) != 7 {
		t.Fatalf("new-project should generate 7 docs, got=%d", len(specs))
	}

	byKey := map[string]kitDocSpec{}
	for _, spec := range specs {
		byKey[spec.Key] = spec
	}

	plan, ok := byKey["plan"]
	if !ok {
		t.Fatalf("new-project plan doc must exist")
	}
	if plan.Role != "source" {
		t.Fatalf("new-project plan doc must be source, got=%q", plan.Role)
	}
	checklist, ok := byKey["ai_checklist"]
	if !ok {
		t.Fatalf("new-project AI checklist doc must exist")
	}
	if checklist.SourceKey != "plan" {
		t.Fatalf("AI checklist should derive from project plan source, got=%q", checklist.SourceKey)
	}
	for _, spec := range specs {
		if spec.Role == "derived" && spec.SourceKey != "plan" {
			t.Fatalf("%s should derive from project plan source, got=%q", spec.Key, spec.SourceKey)
		}
	}
}

func TestNewProjectKitScopesEveryDocIntoProjectFolder(t *testing.T) {
	for _, spec := range kitProfileSpecs("new-project") {
		got := kitProjectScopedSubdir("new-project", spec.Subdir, "checkout")
		want := filepath.Join("20_new_project", "checkout")
		if got != want {
			t.Fatalf("%s scoped folder=%q want %q", spec.Key, got, want)
		}
	}
}

func TestKitProfileSpecsMaintenanceIncidentAvoidSharedFolder(t *testing.T) {
	for _, profile := range []string{"maintenance", "incident"} {
		specs := kitProfileSpecs(profile)
		if len(specs) != 1 {
			t.Fatalf("%s should generate one doc, got=%d", profile, len(specs))
		}
		if specs[0].Subdir != filepath.Join("20_new_project", "delivery") {
			t.Fatalf("%s should use new project delivery folder, got=%q", profile, specs[0].Subdir)
		}
	}
}
