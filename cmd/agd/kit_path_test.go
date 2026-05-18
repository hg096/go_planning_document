// kit_path_test.go verifies project-scoped kit path rules.
// kit_path_test.go는 프로젝트 범위 kit 경로 규칙을 검증합니다.
package main

import (
	"path/filepath"
	"testing"
)

func TestKitProjectScopedSubdir(t *testing.T) {
	got := kitProjectScopedSubdir("new-project", filepath.Join("20_new_project", "delivery"), "checkout")
	want := filepath.Join("20_new_project", "checkout")
	if got != want {
		t.Fatalf("unexpected scoped subdir: got=%q want=%q", got, want)
	}

	gotProduct := kitProjectScopedSubdir("new-project", filepath.Join("20_new_project", "product"), "checkout")
	if gotProduct != want {
		t.Fatalf("new-project docs should share one project folder: got=%q want=%q", gotProduct, want)
	}

	gotRoadmap := kitProjectScopedSubdir("new-project", filepath.Join("20_new_project", "roadmap"), "checkout")
	if gotRoadmap != want {
		t.Fatalf("new-project docs should share one project folder: got=%q want=%q", gotRoadmap, want)
	}

	keepDerived := kitProjectScopedSubdir("starter-kit", filepath.Join("20_new_project", "delivery"), "checkout")
	if keepDerived != filepath.Join("20_new_project", "delivery") {
		t.Fatalf("starter-kit should not append project key: got=%q", keepDerived)
	}

	keepMaintenance := kitProjectScopedSubdir("maintenance", filepath.Join("20_new_project", "delivery"), "checkout")
	if keepMaintenance != filepath.Join("20_new_project", "delivery") {
		t.Fatalf("maintenance doc should stay single-file path: got=%q", keepMaintenance)
	}

	keepIncident := kitProjectScopedSubdir("incident", filepath.Join("20_new_project", "delivery"), "checkout")
	if keepIncident != filepath.Join("20_new_project", "delivery") {
		t.Fatalf("incident should not append project key: got=%q", keepIncident)
	}

	if keep := kitProjectScopedSubdir("maintenance", filepath.Join("20_new_project", "product"), ""); keep != filepath.Join("20_new_project", "product") {
		t.Fatalf("empty key should keep base subdir, got=%q", keep)
	}

	if keep := kitProjectScopedSubdir("new-project", filepath.Join("20_new_project", "delivery"), ""); keep != filepath.Join("20_new_project", "delivery") {
		t.Fatalf("new-project with empty key should keep base subdir, got=%q", keep)
	}
}
