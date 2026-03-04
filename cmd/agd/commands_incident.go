// commands_incident.go handles incident feature-tag updates and related mutations.
// commands_incident.go는 오류,버그 기능 태그와 관련 변경 명령을 처리합니다.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agd/internal/agd"
)

func parseIncidentTagFlags(args []string) ([]string, string, string, string, error) {
	core := make([]string, 0, len(args))
	moduleTokens := make([]string, 0)
	reasonTokens := make([]string, 0)
	impactTokens := make([]string, 0)

	moduleSeen := false
	reasonSeen := false
	impactSeen := false
	current := ""

	for i := 0; i < len(args); i++ {
		token := strings.TrimSpace(args[i])
		switch strings.ToLower(token) {
		case "--module", "-m":
			if moduleSeen {
				return nil, "", "", "", fmt.Errorf("module flag provided multiple times")
			}
			moduleSeen = true
			current = "module"
			continue
		case "--reason", "-r":
			if reasonSeen {
				return nil, "", "", "", fmt.Errorf("reason flag provided multiple times")
			}
			reasonSeen = true
			current = "reason"
			continue
		case "--impact", "-i":
			if impactSeen {
				return nil, "", "", "", fmt.Errorf("impact flag provided multiple times")
			}
			impactSeen = true
			current = "impact"
			continue
		}

		switch current {
		case "module":
			moduleTokens = append(moduleTokens, args[i])
		case "reason":
			reasonTokens = append(reasonTokens, args[i])
		case "impact":
			impactTokens = append(impactTokens, args[i])
		default:
			core = append(core, args[i])
		}
	}

	module := strings.TrimSpace(strings.Join(moduleTokens, " "))
	reason := strings.TrimSpace(strings.Join(reasonTokens, " "))
	impact := strings.TrimSpace(strings.Join(impactTokens, " "))

	if moduleSeen && module == "" {
		return nil, "", "", "", fmt.Errorf("module value is missing")
	}
	if reasonSeen && reason == "" {
		return nil, "", "", "", fmt.Errorf("reason value is missing")
	}
	if impactSeen && impact == "" {
		return nil, "", "", "", fmt.Errorf("impact value is missing")
	}
	return core, module, reason, impact, nil
}

func upsertDelimitedContentBlock(content, startMarker, endMarker, block string) (string, bool) {
	start := strings.TrimSpace(startMarker)
	end := strings.TrimSpace(endMarker)
	nextBlock := strings.TrimSpace(block)
	current := strings.TrimRight(content, "\n")
	if start == "" || end == "" || nextBlock == "" {
		return content, false
	}

	startIdx := strings.Index(current, start)
	if startIdx >= 0 {
		searchFrom := startIdx + len(start)
		relativeEnd := strings.Index(current[searchFrom:], end)
		if relativeEnd >= 0 {
			endIdx := searchFrom + relativeEnd + len(end)
			prefix := strings.TrimRight(current[:startIdx], "\n")
			suffix := strings.TrimLeft(current[endIdx:], "\n")

			merged := nextBlock
			if prefix != "" {
				merged = prefix + "\n\n" + merged
			}
			if strings.TrimSpace(suffix) != "" {
				merged = merged + "\n\n" + strings.TrimLeft(suffix, "\n")
			}
			if merged == current {
				return content, false
			}
			return merged, true
		}

		prefix := strings.TrimRight(current[:startIdx], "\n")
		if prefix == "" {
			return nextBlock, true
		}
		return prefix + "\n\n" + nextBlock, true
	}

	if strings.TrimSpace(current) == "" {
		return nextBlock, true
	}
	return current + "\n\n" + nextBlock, true
}

func incidentProblemTagBlock(featureTag, sourceDoc, sectionID, module string) string {
	lines := []string{
		"[INCIDENT-PROBLEM-TAG]",
		fmt.Sprintf("- feature_tag: %s", strings.TrimSpace(featureTag)),
		fmt.Sprintf("- source_doc: %s", strings.TrimSpace(sourceDoc)),
		fmt.Sprintf("- section_id: %s", strings.TrimSpace(sectionID)),
	}
	if strings.TrimSpace(module) != "" {
		lines = append(lines, fmt.Sprintf("- module: %s", strings.TrimSpace(module)))
	}
	lines = append(lines,
		fmt.Sprintf("- tagged_at: %s", time.Now().Format("2006-01-02")),
		"[/INCIDENT-PROBLEM-TAG]",
	)
	return strings.Join(lines, "\n")
}

func maintenanceRootTagBlock(featureTag, sourceDoc, sectionID string) string {
	lines := []string{
		"[MAINTENANCE-ROOT-TAG]",
		fmt.Sprintf("- feature_tag: %s", strings.TrimSpace(featureTag)),
	}
	if strings.TrimSpace(sourceDoc) != "" {
		lines = append(lines, fmt.Sprintf("- source_doc: %s", strings.TrimSpace(sourceDoc)))
	}
	if strings.TrimSpace(sectionID) != "" {
		lines = append(lines, fmt.Sprintf("- section_id: %s", strings.TrimSpace(sectionID)))
	}
	lines = append(lines,
		fmt.Sprintf("- tagged_at: %s", time.Now().Format("2006-01-02")),
		"[/MAINTENANCE-ROOT-TAG]",
	)
	return strings.Join(lines, "\n")
}

func applyIncidentTagBySection(targetPath, sourcePath, featureTag, sectionID, module, reason, impact, author string) int {
	if strings.TrimSpace(featureTag) == "" {
		fmt.Fprintln(os.Stderr, text("incident-tag error: feature tag is required", "incident-tag error: feature tag is required"))
		return 2
	}
	if strings.TrimSpace(sectionID) == "" {
		fmt.Fprintln(os.Stderr, text("incident-tag error: section id is required", "incident-tag error: section id is required"))
		return 2
	}
	if strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, text("incident-tag error: --reason is required", "incident-tag error: --reason is required"))
		return 2
	}
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, text("incident-tag error: --impact is required", "incident-tag error: --impact is required"))
		return 2
	}

	targetDoc, parseErrs, err := agd.LoadFile(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read failed: %v\n", err)
		return 1
	}
	if len(parseErrs) > 0 {
		for _, e := range parseErrs {
			fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
		}
		return 1
	}
	if len(targetDoc.Sections) == 0 {
		fmt.Fprintln(os.Stderr, text(
			"incident-tag error: target document has no sections",
			"incident-tag error: target document has no sections",
		))
		return 1
	}

	storedSource := filepath.Clean(sourcePath)
	if rel, relErr := filepath.Rel(filepath.Dir(targetPath), sourcePath); relErr == nil {
		storedSource = filepath.Clean(rel)
	}
	traceChain := fmt.Sprintf(
		"%s -> INC-001 -> INC-010 -> INC-020 -> INC-030 -> INC-040 -> INC-050 -> INC-060",
		strings.TrimSpace(featureTag),
	)

	targetDoc.Meta["incident_feature_tag"] = strings.TrimSpace(featureTag)
	targetDoc.Meta["incident_mode"] = "feature-root-trace"
	targetDoc.Meta["incident_trace_root"] = traceChain
	targetDoc.Meta["incident_source_doc"] = storedSource
	targetDoc.Meta["incident_source_section"] = strings.TrimSpace(sectionID)
	if strings.TrimSpace(module) != "" {
		targetDoc.Meta["incident_source_module"] = strings.TrimSpace(module)
	} else {
		delete(targetDoc.Meta, "incident_source_module")
	}

	targetSection := targetDoc.SectionByID("INC-001")
	if targetSection == nil {
		targetSection = targetDoc.Sections[0]
	}

	block := incidentProblemTagBlock(featureTag, storedSource, sectionID, module)
	nextContent, changed := upsertDelimitedContentBlock(
		targetSection.Content,
		"[INCIDENT-PROBLEM-TAG]",
		"[/INCIDENT-PROBLEM-TAG]",
		block,
	)
	if changed {
		targetSection.Content = nextContent
	}

	targetDoc.Changes = append(targetDoc.Changes, &agd.Change{
		ID:     nextChangeID(targetDoc),
		Target: strings.TrimSpace(targetSection.ID),
		Date:   time.Now().Format("2006-01-02"),
		Author: strings.TrimSpace(author),
		Reason: strings.TrimSpace(reason),
		Impact: strings.TrimSpace(impact),
	})
	ensureDocPathMetadata(targetDoc, targetPath)

	validationErrors := agd.Validate(targetDoc)
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}
	crossDocErrors := lintAuthorityAndConflicts(targetPath, targetDoc)
	if len(crossDocErrors) > 0 {
		for _, e := range crossDocErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}
	if err := os.WriteFile(targetPath, []byte(agd.Serialize(targetDoc)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		return 1
	}

	fmt.Printf(text(
		"Tagged incident feature %s on section %s (source=%s) -> %s\n", "Tagged incident feature %s on section %s (source=%s) -> %s\n",
	), strings.TrimSpace(featureTag), strings.TrimSpace(sectionID), storedSource, targetPath)

	return runGuidedFlowAfterMutation(targetPath)
}

func commandIncidentTagEasy(args []string) int {
	usageText := text(
		`usage: agd incident-tag <target.agd> <feature-tag> [source-doc.agd] [section-id] [--module "<name>"] --reason "<why>" --impact "<effect>"`, `usage: agd incident-tag <target.agd> <feature-tag> [source-doc.agd] [section-id] [--module "<name>"] --reason "<why>" --impact "<effect>"`,
	)
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, usageText)
		return 2
	}

	coreArgs, module, reason, impact, parseErr := parseIncidentTagFlags(args)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "incident-tag error: %v\n", parseErr)
		fmt.Fprintln(os.Stderr, usageText)
		return 2
	}
	if len(coreArgs) < 2 || len(coreArgs) > 4 {
		fmt.Fprintln(os.Stderr, usageText)
		return 2
	}
	if strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, text("incident-tag error: --reason is required", "incident-tag error: --reason is required"))
		fmt.Fprintln(os.Stderr, usageText)
		return 2
	}
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, text("incident-tag error: --impact is required", "incident-tag error: --impact is required"))
		fmt.Fprintln(os.Stderr, usageText)
		return 2
	}

	targetPath, err := resolveAGDPathEasy(coreArgs[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	featureTag := normalizeIncidentFeatureTag(coreArgs[1], "")
	if featureTag == "" {
		fmt.Fprintln(os.Stderr, text("incident-tag error: invalid feature tag format", "incident-tag error: invalid feature tag format"))
		return 2
	}

	sourcePath := targetPath
	sectionChoice := ""
	if len(coreArgs) >= 3 {
		third := strings.TrimSpace(coreArgs[2])
		if len(coreArgs) == 4 {
			resolvedSource, resolveErr := resolveAGDPathEasy(third, true)
			if resolveErr != nil {
				fmt.Fprintf(os.Stderr, "%s\n", resolveErr.Error())
				return 1
			}
			sourcePath = resolvedSource
			sectionChoice = strings.TrimSpace(coreArgs[3])
		} else {
			if resolvedSource, resolveErr := resolveAGDPathEasy(third, true); resolveErr == nil {
				sourcePath = resolvedSource
			} else {
				sectionChoice = third
			}
		}
	}

	sourceDoc, parseErrs, err := agd.LoadFile(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read failed: %v\n", err)
		return 1
	}
	if len(parseErrs) > 0 {
		for _, e := range parseErrs {
			fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
		}
		return 1
	}
	if len(sourceDoc.Sections) == 0 {
		fmt.Fprintln(os.Stderr, text(
			"incident-tag error: source document has no sections",
			"incident-tag error: source document has no sections",
		))
		return 1
	}

	if strings.TrimSpace(sectionChoice) == "" {
		reader := bufio.NewReader(os.Stdin)
		selectedSectionID, promptErr := promptSectionIDChoiceFromDoc(reader, sourcePath, sourceDoc)
		if promptErr != nil {
			fmt.Fprintf(os.Stderr, "incident-tag error: %v\n", promptErr)
			return 1
		}
		sectionChoice = selectedSectionID
	} else {
		resolvedSectionID, resolveErr := resolveSectionIDChoice(sourceDoc, sectionChoice)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "incident-tag error: %v\n", resolveErr)
			return 1
		}
		sectionChoice = resolvedSectionID
	}

	return applyIncidentTagBySection(targetPath, sourcePath, featureTag, sectionChoice, module, reason, impact, "ai-agent")
}
