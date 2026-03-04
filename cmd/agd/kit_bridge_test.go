package main

import (
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

	core, ok := byKey["core_spec"]
	if !ok || core.Role != "source" {
		t.Fatalf("core_spec doc must exist as source, got=%+v", core)
	}

	if _, ok := byKey["delivery_plan"]; ok {
		t.Fatalf("starter-kit should not generate delivery_plan")
	}

	if _, ok := byKey["prd"]; ok {
		t.Fatalf("legacy prd doc should not be generated in starter-kit")
	}
	if _, ok := byKey["service"]; ok {
		t.Fatalf("legacy service doc should not be generated in starter-kit")
	}
}
