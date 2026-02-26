// wizard_folder_scope_test.go verifies folder scope boundary checks.
// wizard_folder_scope_test.go는 폴더 범위 경계 체크를 검증합니다.
package main

import (
	"path/filepath"
	"testing"
)

func TestIsPathWithinBase(t *testing.T) {
	base := filepath.Clean(filepath.Join("agd_docs", "10_source", "product"))

	if !isPathWithinBase(base, base) {
		t.Fatalf("base path must be allowed")
	}
	if !isPathWithinBase(base, filepath.Join(base, "checkout")) {
		t.Fatalf("child path must be allowed")
	}
	if isPathWithinBase(base, filepath.Join("agd_docs", "10_source", "service")) {
		t.Fatalf("sibling path must be blocked")
	}
	if isPathWithinBase(base, filepath.Join("agd_docs", "30_shared", "roadmap")) {
		t.Fatalf("other root path must be blocked")
	}
}

