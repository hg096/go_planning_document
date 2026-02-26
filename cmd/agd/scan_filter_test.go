// scan_filter_test.go verifies scan filters for ended maintenance/incident docs.
// scan_filter_test.go는 종료된 유지보수/오류,버그 문서 스캔 필터를 검증합니다.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListAGDFilesSkipsEndedMaintenanceAndIncidentDocs(t *testing.T) {
	root := t.TempDir()

	mustWrite := func(rel string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(path, []byte("@meta\ntitle: x\nversion: 0.1\n"), 0o644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	mustWrite(filepath.Join("30_shared", "maintenance", "normal_maintenance_case.agd"))
	mustWrite(filepath.Join("30_shared", "maintenance", "END__done_maintenance_case.agd"))
	mustWrite(filepath.Join("30_shared", "errFix", "normal_incident_case.agd"))
	mustWrite(filepath.Join("30_shared", "errFix", "END__done_incident_case.agd"))
	mustWrite(filepath.Join("10_source", "service", "END__service_logic.agd"))

	files, err := listAGDFiles(root)
	if err != nil {
		t.Fatalf("listAGDFiles failed: %v", err)
	}

	has := map[string]bool{}
	for _, f := range files {
		has[filepath.Base(f)] = true
	}

	if has["END__done_maintenance_case.agd"] {
		t.Fatalf("ended maintenance doc should be excluded from scan")
	}
	if has["END__done_incident_case.agd"] {
		t.Fatalf("ended incident doc should be excluded from scan")
	}
	if !has["normal_maintenance_case.agd"] {
		t.Fatalf("normal maintenance doc must remain in scan")
	}
	if !has["normal_incident_case.agd"] {
		t.Fatalf("normal incident doc must remain in scan")
	}
	if !has["END__service_logic.agd"] {
		t.Fatalf("END__ prefix for non-target doc type must not be auto-excluded")
	}
}
