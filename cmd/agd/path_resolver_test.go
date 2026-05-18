package main

import (
	"path/filepath"
	"testing"
)

func TestDefaultSubdirForDocTypeUsesFocusedDocsStructure(t *testing.T) {
	cases := map[string]string{
		"service":          filepath.Join("10_core_logic", "service"),
		"policy":           filepath.Join("10_core_logic", "policy"),
		"architecture":     filepath.Join("10_core_logic", "architecture"),
		"adr":              filepath.Join("10_core_logic", "architecture"),
		"core-spec":        filepath.Join("20_new_project", "product"),
		"delivery-plan":    filepath.Join("20_new_project", "delivery"),
		"maintenance-case": filepath.Join("20_new_project", "delivery"),
		"incident-case":    filepath.Join("20_new_project", "delivery"),
		"unknown":          filepath.Join("20_new_project", "product"),
	}

	for docType, want := range cases {
		if got := defaultSubdirForDocType(docType); got != want {
			t.Fatalf("defaultSubdirForDocType(%q)=%q want %q", docType, got, want)
		}
	}
}
