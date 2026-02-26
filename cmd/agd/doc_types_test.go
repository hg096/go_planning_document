// doc_types_test.go verifies document type allowlist behavior.
// doc_types_test.go는 문서 타입 허용 목록 동작을 검증합니다.
package main

import "testing"

func TestAllowedCreateDocTypeOptionsMatchAllowedKeys(t *testing.T) {
	keys := allowedCreateDocTypeKeys()
	options := allowedCreateDocTypeOptions()

	if len(options) != len(keys) {
		t.Fatalf("allowed options count mismatch: got %d, want %d", len(options), len(keys))
	}

	seen := make(map[string]bool, len(options))
	for _, opt := range options {
		seen[opt.Key] = true
	}
	for _, key := range keys {
		if !seen[key] {
			t.Fatalf("allowed key missing in options: %s", key)
		}
	}
}

func TestExperimentDocTypeIsSelectable(t *testing.T) {
	if !isAllowedCreateDocType("experiment") {
		t.Fatalf("experiment must be allowed")
	}
}

