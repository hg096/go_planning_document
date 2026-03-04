package main

import (
	"strings"
	"testing"
)

func TestKitProfileSpecsBridgeLite(t *testing.T) {
	specs := kitProfileSpecs("bridge-lite")
	if len(specs) != 4 {
		t.Fatalf("bridge-lite should generate 4 docs, got=%d", len(specs))
	}

	byKey := map[string]kitDocSpec{}
	for _, spec := range specs {
		byKey[spec.Key] = spec
	}

	service, ok := byKey["service"]
	if !ok || service.Role != "source" {
		t.Fatalf("service doc must exist as source, got=%+v", service)
	}
	core, ok := byKey["core_spec"]
	if !ok || core.Role != "source" {
		t.Fatalf("core_spec doc must exist as source, got=%+v", core)
	}

	delivery, ok := byKey["delivery_plan"]
	if !ok || delivery.Role != "derived" || delivery.SourceKey != "core_spec" {
		t.Fatalf("delivery_plan must derive from core_spec, got=%+v", delivery)
	}
	if !strings.Contains(delivery.SourceSections, "CORE-010->DEL-001") {
		t.Fatalf("delivery_plan source mapping missing CORE-010->DEL-001: %q", delivery.SourceSections)
	}

	runbook, ok := byKey["runbook"]
	if !ok || runbook.Role != "derived" || runbook.SourceKey != "service" {
		t.Fatalf("runbook must derive from service, got=%+v", runbook)
	}
	if !strings.Contains(runbook.SourceSections, "SYS-050->RUN-020") {
		t.Fatalf("runbook source mapping missing SYS-050->RUN-020: %q", runbook.SourceSections)
	}
}
