// wizard_section_relations.go provides section-level relation lookup for wizard flows.
// wizard_section_relations.go는 위저드의 섹션 단위 연관/영향 조회 기능을 제공합니다.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agd/internal/agd"
)

type sectionRelationItem struct {
	DocRel    string
	SectionID string
	Via       string
	Note      string
}

type sectionRelationReport struct {
	TargetDocRel string
	TargetSecID  string
	Connected    []sectionRelationItem
	Impacted     []sectionRelationItem
}

func wizardSectionRelationOnDoc(reader *bufio.Reader, filePath string) int {
	doc, parseErrs, err := agd.LoadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "section-relation error: %v\n", err)
		return 1
	}
	if len(parseErrs) > 0 {
		for _, p := range parseErrs {
			fmt.Fprintf(os.Stderr, "section-relation parse error: %s\n", p)
		}
		return 1
	}

	sectionID, err := promptSectionIDChoiceFromDoc(reader, filePath, doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "section-relation error: %v\n", err)
		return 1
	}

	report, warnings, err := buildSectionRelationReport(filePath, doc, sectionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "section-relation error: %v\n", err)
		return 1
	}
	printSectionRelationReport(report, warnings)
	return 0
}

func buildSectionRelationReport(targetPath string, targetDoc *agd.Document, sectionID string) (sectionRelationReport, []string, error) {
	targetAbs, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return sectionRelationReport{}, nil, err
	}
	repoRoot := detectCodePlanRepoRoot()
	targetRel := codePlanRelPathFromRoot(repoRoot, targetAbs)
	targetKey := codePlanPathKey(targetRel)
	targetSec := strings.ToUpper(strings.TrimSpace(sectionID))

	report := sectionRelationReport{
		TargetDocRel: targetRel,
		TargetSecID:  targetSec,
		Connected:    make([]sectionRelationItem, 0),
		Impacted:     make([]sectionRelationItem, 0),
	}
	warnings := make([]string, 0)
	connectedSeen := make(map[string]bool)
	impactedSeen := make(map[string]bool)
	addConnected := func(docRel, secID, via, note string) {
		item := normalizeSectionRelationItem(docRel, secID, via, note)
		key := relationItemKey(item)
		if connectedSeen[key] {
			return
		}
		connectedSeen[key] = true
		report.Connected = append(report.Connected, item)
	}
	addImpacted := func(docRel, secID, via, note string) {
		item := normalizeSectionRelationItem(docRel, secID, via, note)
		key := relationItemKey(item)
		if impactedSeen[key] {
			return
		}
		impactedSeen[key] = true
		report.Impacted = append(report.Impacted, item)
	}

	// 1) Current doc -> source doc mapping (derived depends on source).
	targetSourceRef := strings.TrimSpace(targetDoc.Meta["source_doc"])
	targetMappingsRaw := strings.TrimSpace(targetDoc.Meta["source_sections"])
	if targetSourceRef != "" && targetMappingsRaw != "" {
		sourcePath, resolveErr := resolveSourceDocPathForLint(targetAbs, targetSourceRef)
		if resolveErr != nil {
			warnings = append(warnings, fmt.Sprintf("target source_doc resolve failed: %v", resolveErr))
		} else {
			sourceRel := codePlanRelPathFromRoot(repoRoot, sourcePath)
			mappings, parseErrs := parseSourceSectionMappings(targetMappingsRaw)
			for _, parseErr := range parseErrs {
				warnings = append(warnings, fmt.Sprintf("target source_sections parse warning: %s", parseErr))
			}
			for _, mapping := range mappings {
				srcID := strings.ToUpper(strings.TrimSpace(mapping.SourceID))
				dstID := strings.ToUpper(strings.TrimSpace(mapping.DerivedID))
				if targetSec == dstID {
					addConnected(sourceRel, srcID, "source_sections", "depends on mapped source section")
				}
			}
		}
	}

	// 2) Reverse lookup from other docs that reference target as source_doc.
	if err := ensureDocsRoot(); err != nil {
		warnings = append(warnings, fmt.Sprintf("docs root ensure warning: %v", err))
	}
	files, scanErr := listAGDFiles(docsRootDir)
	if scanErr != nil {
		return report, warnings, scanErr
	}
	for _, file := range files {
		currentAbs := filepath.Clean(file)
		if abs, absErr := filepath.Abs(currentAbs); absErr == nil {
			currentAbs = abs
		}
		if strings.EqualFold(filepath.Clean(currentAbs), filepath.Clean(targetAbs)) {
			continue
		}

		doc, parseErrs, loadErr := agd.LoadFile(currentAbs)
		if loadErr != nil {
			continue
		}
		if len(parseErrs) > 0 {
			continue
		}
		docRel := codePlanRelPathFromRoot(repoRoot, currentAbs)

		sourceRef := strings.TrimSpace(doc.Meta["source_doc"])
		mappingsRaw := strings.TrimSpace(doc.Meta["source_sections"])
		if sourceRef != "" && mappingsRaw != "" {
			sourceAbs, resolveErr := resolveSourceDocPathForLint(currentAbs, sourceRef)
			if resolveErr == nil && strings.EqualFold(filepath.Clean(sourceAbs), filepath.Clean(targetAbs)) {
				mappings, parseErrs := parseSourceSectionMappings(mappingsRaw)
				for _, parseErr := range parseErrs {
					warnings = append(warnings, fmt.Sprintf("%s source_sections warning: %s", docRel, parseErr))
				}
				for _, mapping := range mappings {
					srcID := strings.ToUpper(strings.TrimSpace(mapping.SourceID))
					dstID := strings.ToUpper(strings.TrimSpace(mapping.DerivedID))
					if targetSec == srcID {
						addConnected(docRel, dstID, "source_sections", "mapped derived section linked to target")
						addImpacted(docRel, dstID, "source_sections", "review/update derived section after source change")
					}
				}
			}
		}

		collectMetaTagRelationLink(report.TargetDocRel, targetSec, repoRoot, currentAbs, doc, docRel, addConnected, addImpacted)
	}

	// 3) Explicit bidirectional relation policy: A#SEC <-> B#SEC
	relationPolicyPath := resolveCodePlanPath(repoRoot, codePlanDefaultRelationPolicyPath)
	relationPolicy, relationWarnings, enabled, relationErr := loadCodePlanRelationPolicy(relationPolicyPath)
	if relationErr != nil {
		warnings = append(warnings, fmt.Sprintf("relation policy warning: %v", relationErr))
	} else {
		warnings = append(warnings, relationWarnings...)
	}
	if enabled {
		for _, rule := range relationPolicy.Rules {
			if rule.Left.DocKey == targetKey && rule.Left.SectionID == targetSec {
				addConnected(rule.Right.DocPath, rule.Right.SectionID, "relation-policy", "bidirectional section relation")
				addImpacted(rule.Right.DocPath, rule.Right.SectionID, "relation-policy", "linked section should be reviewed after edit")
				continue
			}
			if rule.Right.DocKey == targetKey && rule.Right.SectionID == targetSec {
				addConnected(rule.Left.DocPath, rule.Left.SectionID, "relation-policy", "bidirectional section relation")
				addImpacted(rule.Left.DocPath, rule.Left.SectionID, "relation-policy", "linked section should be reviewed after edit")
			}
		}
	}

	sortSectionRelationItems(report.Connected)
	sortSectionRelationItems(report.Impacted)
	sort.Strings(warnings)
	return report, warnings, nil
}

func collectMetaTagRelationLink(targetDocRel, targetSec, repoRoot, currentAbs string, doc *agd.Document, docRel string, addConnected func(string, string, string, string), addImpacted func(string, string, string, string)) {
	if doc == nil {
		return
	}
	type metaPair struct {
		docKey string
		secKey string
		label  string
		marker string
	}
	pairs := []metaPair{
		{docKey: "incident_source_doc", secKey: "incident_source_section", label: "incident-link", marker: "INCIDENT-PROBLEM-TAG"},
		{docKey: "maintenance_source_doc", secKey: "maintenance_source_section", label: "maintenance-link", marker: "MAINTENANCE-ROOT-TAG"},
	}
	for _, pair := range pairs {
		refDoc := strings.TrimSpace(doc.Meta[pair.docKey])
		refSec := strings.ToUpper(strings.TrimSpace(doc.Meta[pair.secKey]))
		if refDoc == "" || refSec == "" || refSec != targetSec {
			continue
		}
		resolvedSource, err := resolveSourceDocPathForLint(currentAbs, refDoc)
		if err != nil {
			continue
		}
		resolvedRel := codePlanRelPathFromRoot(repoRoot, resolvedSource)
		if codePlanPathKey(resolvedRel) != codePlanPathKey(targetDocRel) {
			continue
		}
		rootSectionID := findSectionIDByMarker(doc, pair.marker)
		addConnected(docRel, rootSectionID, pair.label, "meta tag traces to target section")
		addImpacted(docRel, rootSectionID, pair.label, "trace/case document should be reviewed")
	}
}

func findSectionIDByMarker(doc *agd.Document, marker string) string {
	if doc == nil {
		return ""
	}
	needle := "[" + strings.ToUpper(strings.TrimSpace(marker)) + "]"
	for _, section := range doc.Sections {
		if section == nil {
			continue
		}
		if strings.Contains(strings.ToUpper(section.Content), needle) {
			return strings.ToUpper(strings.TrimSpace(section.ID))
		}
	}
	if len(doc.Sections) > 0 && doc.Sections[0] != nil {
		return strings.ToUpper(strings.TrimSpace(doc.Sections[0].ID))
	}
	return ""
}

func normalizeSectionRelationItem(docRel, secID, via, note string) sectionRelationItem {
	return sectionRelationItem{
		DocRel:    codePlanNormalizeRelPath(docRel),
		SectionID: strings.ToUpper(strings.TrimSpace(secID)),
		Via:       strings.TrimSpace(via),
		Note:      strings.TrimSpace(note),
	}
}

func relationItemKey(item sectionRelationItem) string {
	return strings.ToLower(item.DocRel) + "|" + strings.ToUpper(item.SectionID) + "|" + strings.ToLower(item.Via)
}

func sortSectionRelationItems(items []sectionRelationItem) {
	sort.Slice(items, func(i, j int) bool {
		if strings.ToLower(items[i].DocRel) != strings.ToLower(items[j].DocRel) {
			return strings.ToLower(items[i].DocRel) < strings.ToLower(items[j].DocRel)
		}
		if strings.ToUpper(items[i].SectionID) != strings.ToUpper(items[j].SectionID) {
			return strings.ToUpper(items[i].SectionID) < strings.ToUpper(items[j].SectionID)
		}
		return strings.ToLower(items[i].Via) < strings.ToLower(items[j].Via)
	})
}

func printSectionRelationReport(report sectionRelationReport, warnings []string) {
	fmt.Println("")
	fmt.Println(text("SECTION RELATION LOOKUP", "섹션 연관 검색"))
	fmt.Printf(text("Target: %s#%s\n", "대상: %s#%s\n"), report.TargetDocRel, report.TargetSecID)

	fmt.Println("")
	fmt.Println(text("Connected sections:", "연결된 섹션:"))
	if len(report.Connected) == 0 {
		fmt.Println(text("  (none)", "  (없음)"))
	} else {
		for _, item := range report.Connected {
			fmt.Printf("  - %s (via=%s)\n", formatSectionRef(item.DocRel, item.SectionID), item.Via)
			if item.Note != "" {
				fmt.Printf("    %s\n", item.Note)
			}
		}
	}

	fmt.Println("")
	fmt.Println(text("Impact sections when editing target:", "대상 섹션 수정 시 영향 후보:"))
	if len(report.Impacted) == 0 {
		fmt.Println(text("  (none)", "  (없음)"))
	} else {
		for _, item := range report.Impacted {
			fmt.Printf("  - %s (via=%s)\n", formatSectionRef(item.DocRel, item.SectionID), item.Via)
			if item.Note != "" {
				fmt.Printf("    %s\n", item.Note)
			}
		}
	}

	if len(warnings) > 0 {
		fmt.Println("")
		fmt.Println(text("Warnings:", "경고:"))
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
}

func formatSectionRef(docRel, secID string) string {
	doc := strings.TrimSpace(docRel)
	sec := strings.TrimSpace(secID)
	if doc == "" {
		doc = text("(unknown-doc)", "(알 수 없는 문서)")
	}
	if sec == "" {
		return doc
	}
	return doc + "#" + sec
}
