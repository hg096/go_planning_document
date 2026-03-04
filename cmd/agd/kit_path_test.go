// kit_path_test.go verifies project-scoped kit path rules.
// kit_path_test.go는 프로젝트 범위 kit 경로 규칙을 검증합니다.
package main

import (
	"path/filepath"
	"testing"
)

func TestKitProjectScopedSubdir(t *testing.T) {
	got := kitProjectScopedSubdir("change-flow", filepath.Join("20_derived", "frontend"), "checkout")
	want := filepath.Join("10_source", "product", "checkout")
	if got != want {
		t.Fatalf("unexpected scoped subdir: got=%q want=%q", got, want)
	}

	keepDerived := kitProjectScopedSubdir("starter-kit", filepath.Join("20_derived", "frontend"), "checkout")
	if keepDerived != filepath.Join("20_derived", "frontend") {
		t.Fatalf("starter-kit should not append project key: got=%q", keepDerived)
	}

	keepBridge := kitProjectScopedSubdir("bridge-lite", filepath.Join("20_derived", "ops"), "checkout")
	if keepBridge != filepath.Join("20_derived", "ops") {
		t.Fatalf("bridge-lite should not append project key: got=%q", keepBridge)
	}

	keepMaintenance := kitProjectScopedSubdir("change-flow", filepath.Join("30_shared", "maintenance"), "checkout")
	if keepMaintenance != filepath.Join("30_shared", "maintenance") {
		t.Fatalf("change-flow maintenance doc should stay single-file path: got=%q", keepMaintenance)
	}

	keepIncident := kitProjectScopedSubdir("incident-lifecycle", filepath.Join("30_shared", "errFix"), "checkout")
	if keepIncident != filepath.Join("30_shared", "errFix") {
		t.Fatalf("incident-lifecycle should not append project key: got=%q", keepIncident)
	}

	keepQuality := kitProjectScopedSubdir("quality-gate", filepath.Join("20_derived", "qa"), "checkout")
	if keepQuality != filepath.Join("20_derived", "qa") {
		t.Fatalf("quality-gate should not append project key: got=%q", keepQuality)
	}

	if keep := kitProjectScopedSubdir("change-flow", filepath.Join("10_source", "product"), ""); keep != filepath.Join("10_source", "product") {
		t.Fatalf("empty key should keep base subdir, got=%q", keep)
	}
}
