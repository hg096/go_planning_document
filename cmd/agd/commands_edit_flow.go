// commands_edit_flow.go handles edit/add/section-add/logic-log mutation command flows.
// commands_edit_flow.go는 edit/add/section-add/logic-log 변경 흐름을 처리합니다.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"agd/internal/agd"
)

func parseReasonImpactFlags(args []string) ([]string, string, string, error) {
	core := make([]string, 0, len(args))
	reasonTokens := make([]string, 0)
	impactTokens := make([]string, 0)
	reasonSeen := false
	impactSeen := false
	current := ""

	for i := 0; i < len(args); i++ {
		token := strings.TrimSpace(args[i])
		switch strings.ToLower(token) {
		case "--reason", "-r":
			if reasonSeen {
				return nil, "", "", fmt.Errorf("reason flag provided multiple times")
			}
			reasonSeen = true
			current = "reason"
			continue
		case "--impact", "-i":
			if impactSeen {
				return nil, "", "", fmt.Errorf("impact flag provided multiple times")
			}
			impactSeen = true
			current = "impact"
			continue
		}
		switch current {
		case "reason":
			reasonTokens = append(reasonTokens, args[i])
		case "impact":
			impactTokens = append(impactTokens, args[i])
		default:
			core = append(core, args[i])
		}
	}

	reason := strings.TrimSpace(strings.Join(reasonTokens, " "))
	impact := strings.TrimSpace(strings.Join(impactTokens, " "))

	if reasonSeen && reason == "" {
		return nil, "", "", fmt.Errorf("reason value is missing")
	}
	if impactSeen && impact == "" {
		return nil, "", "", fmt.Errorf("impact value is missing")
	}
	return core, reason, impact, nil
}

func printReasonImpactUsageHint(command string) {
	switch command {
	case "edit":
		fmt.Fprintln(os.Stderr, text(
			`usage: agd edit <file.agd> <SECTION-ID> <new-summary> --reason "<why>" --impact "<effect>"`, `usage: agd edit <file.agd> <SECTION-ID> <new-summary> --reason "<why>" --impact "<effect>"`,
		))
	case "add":
		fmt.Fprintln(os.Stderr, text(
			`usage: agd add <file.agd> <SECTION-ID> <text-to-append> --reason "<why>" --impact "<effect>"`, `usage: agd add <file.agd> <SECTION-ID> <text-to-append> --reason "<why>" --impact "<effect>"`,
		))
	case "section-add":
		fmt.Fprintln(os.Stderr, text(
			`usage: agd section-add <file.agd> <SECTION-ID> <title> <summary> [links] [content] --reason "<why>" --impact "<effect>"`, `usage: agd section-add <file.agd> <SECTION-ID> <title> <summary> [links] [content] --reason "<why>" --impact "<effect>"`,
		))
	}
}

func commandEditEasy(args []string) int {
	parsedArgs, reason, impact, parseErr := parseReasonImpactFlags(args)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "edit error: %v\n", parseErr)
		printReasonImpactUsageHint("edit")
		return 2
	}
	if len(parsedArgs) < 3 {
		printReasonImpactUsageHint("edit")
		return 2
	}
	if strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, text(
			"edit error: --reason is required",
			"edit error: --reason is required",
		))
		printReasonImpactUsageHint("edit")
		return 2
	}
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, text(
			"edit error: --impact is required",
			"edit error: --impact is required",
		))
		printReasonImpactUsageHint("edit")
		return 2
	}

	filePath, err := resolveAGDPathEasy(parsedArgs[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	sectionID := parsedArgs[1]
	summary := strings.Join(parsedArgs[2:], " ")

	if code := commandPatch([]string{
		filePath,
		"--section", sectionID,
		"--summary", summary,
		"--reason", reason,
		"--impact", impact,
	}); code != 0 {
		return code
	}
	return runGuidedFlowAfterMutation(filePath)
}

func commandAddEasy(args []string) int {
	parsedArgs, reason, impact, parseErr := parseReasonImpactFlags(args)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "add error: %v\n", parseErr)
		printReasonImpactUsageHint("add")
		return 2
	}
	if len(parsedArgs) < 3 {
		printReasonImpactUsageHint("add")
		return 2
	}
	if strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, text(
			"add error: --reason is required",
			"add error: --reason is required",
		))
		printReasonImpactUsageHint("add")
		return 2
	}
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, text(
			"add error: --impact is required",
			"add error: --impact is required",
		))
		printReasonImpactUsageHint("add")
		return 2
	}

	filePath, err := resolveAGDPathEasy(parsedArgs[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	sectionID := parsedArgs[1]
	appendText := strings.Join(parsedArgs[2:], " ")

	if code := commandPatch([]string{
		filePath,
		"--section", sectionID,
		"--append-content", appendText,
		"--reason", reason,
		"--impact", impact,
	}); code != 0 {
		return code
	}
	return runGuidedFlowAfterMutation(filePath)
}

func commandSectionAddEasy(args []string) int {
	parsedArgs, reason, impact, parseErr := parseReasonImpactFlags(args)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "section-add error: %v\n", parseErr)
		printReasonImpactUsageHint("section-add")
		return 2
	}
	if len(parsedArgs) < 4 {
		printReasonImpactUsageHint("section-add")
		return 2
	}
	if strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, text(
			"section-add error: --reason is required",
			"section-add error: --reason is required",
		))
		printReasonImpactUsageHint("section-add")
		return 2
	}
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, text(
			"section-add error: --impact is required",
			"section-add error: --impact is required",
		))
		printReasonImpactUsageHint("section-add")
		return 2
	}

	filePath, err := resolveAGDPathEasy(parsedArgs[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}

	sectionID := strings.TrimSpace(parsedArgs[1])
	title := strings.TrimSpace(parsedArgs[2])
	summary := strings.TrimSpace(parsedArgs[3])
	linksRaw := ""
	content := ""
	if len(parsedArgs) >= 5 {
		linksRaw = strings.TrimSpace(parsedArgs[4])
	}
	if len(parsedArgs) >= 6 {
		content = strings.TrimSpace(strings.Join(parsedArgs[5:], " "))
	}

	if code := addSectionWithMapUpdate(filePath, sectionID, title, summary, linksRaw, content, "ai-agent", reason, impact); code != 0 {
		return code
	}
	return runGuidedFlowAfterMutation(filePath)
}

func commandLogicLogEasy(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, text(
			`usage: agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"`, `usage: agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"`,
		))
		return 2
	}

	filePath, err := resolveAGDPathEasy(args[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	sectionID := strings.TrimSpace(args[1])
	restArgs, reason, impact, parseErr := parseReasonImpactFlags(args[2:])
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "logic-log error: %v\n", parseErr)
		fmt.Fprintln(os.Stderr, text(
			`usage: agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"`, `usage: agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"`,
		))
		return 2
	}
	if len(restArgs) != 0 {
		fmt.Fprintln(os.Stderr, text(
			`usage: agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"`, `usage: agd logic-log <file.agd> <SECTION-ID> --reason "<why>" --impact "<effect>"`,
		))
		return 2
	}
	if sectionID == "" {
		fmt.Fprintln(os.Stderr, text("logic-log error: section id is required", "logic-log error: section id is required"))
		return 2
	}
	if reason == "" {
		fmt.Fprintln(os.Stderr, text("logic-log error: --reason is required", "logic-log error: --reason is required"))
		return 2
	}
	if impact == "" {
		fmt.Fprintln(os.Stderr, text("logic-log error: --impact is required", "logic-log error: --impact is required"))
		return 2
	}

	doc, parseErrs, err := agd.LoadFile(filePath)
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

	if doc.SectionByID(sectionID) == nil {
		fmt.Fprintf(os.Stderr, text("logic-log error: section not found: %s\n", "logic-log error: section not found: %s\n"), sectionID)
		return 1
	}
	ensureDocPathMetadata(doc, filePath)

	change := &agd.Change{
		ID:     nextChangeID(doc),
		Target: sectionID,
		Date:   time.Now().Format("2006-01-02"),
		Author: "ai-agent",
		Reason: reason,
		Impact: impact,
	}
	doc.Changes = append(doc.Changes, change)

	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}

	if err := os.WriteFile(filePath, []byte(agd.Serialize(doc)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		return 1
	}

	fmt.Printf(text(
		"Added logic-change record for %s -> %s\n", "Added logic-change record for %s -> %s\n",
	), sectionID, filePath)
	return 0
}

func addSectionWithMapUpdate(filePath, sectionID, title, summary, linksRaw, content, author, reason, impact string) int {
	if strings.TrimSpace(sectionID) == "" {
		fmt.Fprintln(os.Stderr, text("section-add error: section id is required", "section-add error: section id is required"))
		return 2
	}
	if strings.TrimSpace(title) == "" {
		fmt.Fprintln(os.Stderr, text("section-add error: title is required", "section-add error: title is required"))
		return 2
	}
	if strings.TrimSpace(summary) == "" {
		fmt.Fprintln(os.Stderr, text("section-add error: summary is required", "section-add error: summary is required"))
		return 2
	}
	if strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, text("section-add error: reason is required", "section-add error: reason is required"))
		return 2
	}
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, text("section-add error: impact is required", "section-add error: impact is required"))
		return 2
	}

	doc, parseErrs, err := agd.LoadFile(filePath)
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

	if doc.SectionByID(sectionID) != nil {
		fmt.Fprintf(os.Stderr, text("section-add error: section already exists: %s\n", "section-add error: section already exists: %s\n"), sectionID)
		return 1
	}

	newSection := &agd.Section{
		ID:      sectionID,
		Title:   title,
		Summary: summary,
		Path:    defaultSectionPath(strings.TrimSpace(doc.Meta["doc_base_path"]), sectionID),
		Links:   splitCSV(linksRaw),
		Content: strings.TrimSpace(content),
	}
	doc.Sections = append(doc.Sections, newSection)

	existingMap := doc.MapByID(sectionID)
	if existingMap == nil {
		doc.Map = append(doc.Map, agd.MapEntry{
			ID:      sectionID,
			Title:   title,
			Summary: summary,
		})
	} else {
		existingMap.Title = title
		existingMap.Summary = summary
	}

	change := &agd.Change{
		ID:     nextChangeID(doc),
		Target: sectionID,
		Date:   time.Now().Format("2006-01-02"),
		Author: strings.TrimSpace(author),
		Reason: strings.TrimSpace(reason),
		Impact: strings.TrimSpace(impact),
	}
	doc.Changes = append(doc.Changes, change)
	ensureDocPathMetadata(doc, filePath)

	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}

	if err := os.WriteFile(filePath, []byte(agd.Serialize(doc)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		return 1
	}

	fmt.Printf(text(
		"Added section %s and updated map -> %s\n", "Added section %s and updated map -> %s\n",
	), sectionID, filePath)
	return 0
}
