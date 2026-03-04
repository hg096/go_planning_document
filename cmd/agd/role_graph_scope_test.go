package main

import "testing"

func TestParseRoleGraphScope(t *testing.T) {
	cases := []struct {
		in   string
		want roleGraphScope
		ok   bool
	}{
		{in: "", want: roleGraphScopeAll, ok: true},
		{in: "all", want: roleGraphScopeAll, ok: true},
		{in: "maintenance", want: roleGraphScopeMaintenanceOnly, ok: true},
		{in: "maint", want: roleGraphScopeMaintenanceOnly, ok: true},
		{in: "incident", want: roleGraphScopeIncidentOnly, ok: true},
		{in: "error", want: roleGraphScopeIncidentOnly, ok: true},
		{in: "maintenance+incident", want: roleGraphScopeMaintenanceIncidentOnly, ok: true},
		{in: "exclude-maintenance-incident", want: roleGraphScopeExcludeMaintenanceIncident, ok: true},
		{in: "unknown", want: "", ok: false},
	}

	for _, tc := range cases {
		got, ok := parseRoleGraphScope(tc.in)
		if ok != tc.ok {
			t.Fatalf("parseRoleGraphScope(%q) ok=%v want=%v", tc.in, ok, tc.ok)
		}
		if got != tc.want {
			t.Fatalf("parseRoleGraphScope(%q)=%q want=%q", tc.in, got, tc.want)
		}
	}
}

func TestApplyRoleGraphScopeFilter(t *testing.T) {
	sourcePath := `C:\tmp\10_source\service\service_logic.agd`
	maintenancePath := `C:\tmp\30_shared\maintenance\project_maintenance_case.agd`
	incidentPath := `C:\tmp\30_shared\errFix\project_incident_case.agd`
	derivedPath := `C:\tmp\20_derived\frontend\frontend_page.agd`

	records := map[string]roleGraphRecord{
		sourcePath: {
			Path:      sourcePath,
			Rel:       `10_source\service\service_logic.agd`,
			Authority: "source",
		},
		maintenancePath: {
			Path:      maintenancePath,
			Rel:       `30_shared\maintenance\project_maintenance_case.agd`,
			Authority: "derived",
		},
		incidentPath: {
			Path:      incidentPath,
			Rel:       `30_shared\errFix\project_incident_case.agd`,
			Authority: "derived",
		},
		derivedPath: {
			Path:      derivedPath,
			Rel:       `20_derived\frontend\frontend_page.agd`,
			Authority: "derived",
		},
	}

	edges := []roleGraphEdge{
		{SourcePath: sourcePath, DerivedPath: maintenancePath, Mapping: "SYS-020"},
		{SourcePath: sourcePath, DerivedPath: incidentPath, Mapping: "SYS-030"},
		{SourcePath: sourcePath, DerivedPath: derivedPath, Mapping: "SYS-040"},
	}
	unresolved := []unresolvedRoleLink{
		{DerivedPath: maintenancePath, SourceDocRef: "", Reason: "source_doc is empty"},
		{DerivedPath: derivedPath, SourceDocRef: "", Reason: "source_doc is empty"},
	}
	unassigned := []string{maintenancePath, derivedPath}

	filteredRecords, filteredEdges, filteredUnresolved, filteredUnassigned := applyRoleGraphScopeFilter(
		roleGraphScopeMaintenanceOnly,
		records,
		edges,
		unresolved,
		unassigned,
	)

	if len(filteredRecords) != 2 {
		t.Fatalf("maintenance scope records=%d want=2", len(filteredRecords))
	}
	if _, ok := filteredRecords[sourcePath]; !ok {
		t.Fatalf("maintenance scope should include linked source node")
	}
	if _, ok := filteredRecords[maintenancePath]; !ok {
		t.Fatalf("maintenance scope should include maintenance node")
	}
	if len(filteredEdges) != 1 {
		t.Fatalf("maintenance scope edges=%d want=1", len(filteredEdges))
	}
	if filteredEdges[0].DerivedPath != maintenancePath {
		t.Fatalf("maintenance scope edge target=%s want=%s", filteredEdges[0].DerivedPath, maintenancePath)
	}
	if len(filteredUnresolved) != 1 || filteredUnresolved[0].DerivedPath != maintenancePath {
		t.Fatalf("maintenance scope unresolved mismatch: %+v", filteredUnresolved)
	}
	if len(filteredUnassigned) != 1 || filteredUnassigned[0] != maintenancePath {
		t.Fatalf("maintenance scope unassigned mismatch: %+v", filteredUnassigned)
	}

	excludedRecords, excludedEdges, _, _ := applyRoleGraphScopeFilter(
		roleGraphScopeExcludeMaintenanceIncident,
		records,
		edges,
		nil,
		nil,
	)
	if _, ok := excludedRecords[maintenancePath]; ok {
		t.Fatalf("exclude scope should remove maintenance doc")
	}
	if _, ok := excludedRecords[incidentPath]; ok {
		t.Fatalf("exclude scope should remove incident doc")
	}
	if _, ok := excludedRecords[sourcePath]; !ok {
		t.Fatalf("exclude scope should keep source doc")
	}
	if _, ok := excludedRecords[derivedPath]; !ok {
		t.Fatalf("exclude scope should keep non-maintenance/non-incident doc")
	}
	if len(excludedEdges) != 1 || excludedEdges[0].DerivedPath != derivedPath {
		t.Fatalf("exclude scope edges mismatch: %+v", excludedEdges)
	}
}
