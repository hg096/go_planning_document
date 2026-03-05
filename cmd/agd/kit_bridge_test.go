package main

import (
	"path/filepath"
	"testing"
)

func TestKitProfileSpecsStarterKitUsesUnifiedTemplates(t *testing.T) {
	specs := kitProfileSpecs("starter-kit")
	if len(specs) != 3 {
		t.Fatalf("starter-kit should generate 3 docs, got=%d", len(specs))
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

	if _, ok := byKey["prd"]; ok {
		t.Fatalf("legacy prd doc should not be generated in starter-kit")
	}
	if _, ok := byKey["core_spec"]; ok {
		t.Fatalf("starter-kit should not generate product core_spec")
	}
}

func TestKitProfileSpecsNewProjectUnifiesUnderProductFolder(t *testing.T) {
	specs := kitProfileSpecs("new-project")
	if len(specs) != 4 {
		t.Fatalf("new-project should generate 4 docs, got=%d", len(specs))
	}

	byKey := map[string]kitDocSpec{}
	for _, spec := range specs {
		byKey[spec.Key] = spec
	}

	delivery, ok := byKey["delivery_plan"]
	if !ok {
		t.Fatalf("new-project delivery_plan doc must exist")
	}
	if delivery.Subdir != filepath.Join("10_source", "product") {
		t.Fatalf("new-project delivery_plan should be under product folder, got=%q", delivery.Subdir)
	}
	if delivery.Role != "derived" {
		t.Fatalf("new-project delivery_plan role must remain derived, got=%q", delivery.Role)
	}
}
