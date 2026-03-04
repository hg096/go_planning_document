// commands_map.go handles map-sync and source-map suggestion commands.
// commands_map.go는 map-sync와 source-map 제안 명령을 처리합니다.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"agd/internal/agd"
)

func mapEntriesEqual(a, b []agd.MapEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimSpace(a[i].ID) != strings.TrimSpace(b[i].ID) {
			return false
		}
		if strings.TrimSpace(a[i].Title) != strings.TrimSpace(b[i].Title) {
			return false
		}
		if strings.TrimSpace(a[i].Summary) != strings.TrimSpace(b[i].Summary) {
			return false
		}
	}
	return true
}

func commandMapSyncEasy(args []string) int {
	parsedArgs, reason, impact, parseErr := parseReasonImpactFlags(args)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "map-sync error: %v\n", parseErr)
		fmt.Fprintln(os.Stderr, text(
			`usage: agd map-sync <file.agd> [--reason "<why>" --impact "<effect>"]`, `usage: agd map-sync <file.agd> [--reason "<why>" --impact "<effect>"]`,
		))
		return 2
	}
	if len(parsedArgs) != 1 {
		fmt.Fprintln(os.Stderr, text(
			`usage: agd map-sync <file.agd> [--reason "<why>" --impact "<effect>"]`, `usage: agd map-sync <file.agd> [--reason "<why>" --impact "<effect>"]`,
		))
		return 2
	}

	filePath, err := resolveAGDPathEasy(parsedArgs[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
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

	newMap := make([]agd.MapEntry, 0, len(doc.Sections))
	for _, section := range doc.Sections {
		newMap = append(newMap, agd.MapEntry{
			ID:      strings.TrimSpace(section.ID),
			Title:   strings.TrimSpace(section.Title),
			Summary: strings.TrimSpace(section.Summary),
		})
	}

	if mapEntriesEqual(doc.Map, newMap) {
		fmt.Printf(text(
			"Map already in sync (%d rows) -> %s\n",
			"Map already in sync (%d rows) -> %s\n",
		), len(newMap), filePath)
		return 0
	}

	if strings.TrimSpace(reason) == "" {
		fmt.Fprintln(os.Stderr, text(
			"map-sync error: --reason is required when map changes",
			"map-sync error: --reason is required when map changes",
		))
		fmt.Fprintln(os.Stderr, text(
			`usage: agd map-sync <file.agd> --reason "<why>" --impact "<effect>"`, `usage: agd map-sync <file.agd> --reason "<why>" --impact "<effect>"`,
		))
		return 2
	}
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, text(
			"map-sync error: --impact is required when map changes",
			"map-sync error: --impact is required when map changes",
		))
		fmt.Fprintln(os.Stderr, text(
			`usage: agd map-sync <file.agd> --reason "<why>" --impact "<effect>"`, `usage: agd map-sync <file.agd> --reason "<why>" --impact "<effect>"`,
		))
		return 2
	}
	if len(doc.Sections) == 0 {
		fmt.Fprintln(os.Stderr, text(
			"map-sync error: cannot record @change because document has no section target",
			"map-sync error: cannot record @change because document has no section target",
		))
		return 1
	}

	doc.Map = newMap
	doc.Changes = append(doc.Changes, &agd.Change{
		ID:     nextChangeID(doc),
		Target: strings.TrimSpace(doc.Sections[0].ID),
		Date:   time.Now().Format("2006-01-02"),
		Author: "ai-agent",
		Reason: strings.TrimSpace(reason),
		Impact: strings.TrimSpace(impact),
	})
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
		"Map synced from sections (%d rows) and @change logged -> %s\n",
		"Map synced from sections (%d rows) and @change logged -> %s\n",
	), len(newMap), filePath)
	return 0
}
func commandMapSuggestEasy(args []string) int {
	if len(args) < 2 || len(args) > 3 {
		fmt.Fprintln(os.Stderr, text(
			"usage: agd map-suggest <derived-file.agd> <source-doc.agd> [auto|strict-auto|smart-auto]", "usage: agd map-suggest <derived-file.agd> <source-doc.agd> [auto|strict-auto|smart-auto]",
		))
		return 2
	}

	mode := "auto"
	if len(args) == 3 {
		mode = strings.ToLower(strings.TrimSpace(args[2]))
	}
	strictMode := false
	smartMode := false
	switch mode {
	case "", "auto", "suggest":
		strictMode = false
	case "strict-auto", "auto-strict", "strict", "suggest-strict":
		strictMode = true
	case "smart-auto", "auto-smart", "smart", "suggest-smart":
		strictMode = true
		smartMode = true
	default:
		fmt.Fprintln(os.Stderr, text(
			"map-suggest error: mode must be auto, strict-auto, or smart-auto",
			"map-suggest error: mode must be auto, strict-auto, or smart-auto",
		))
		return 2
	}

	derivedPath, err := resolveAGDPathEasy(args[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	sourcePath, err := resolveSourceDocPathForLint(derivedPath, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}

	derivedDoc, parseErrs, err := agd.LoadFile(derivedPath)
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

	suggested := suggestSourceSectionMappingsDetailed(sourceDoc, derivedDoc, strictMode)
	if smartMode {
		suggested = applySmartAutoRelaxation(suggested, sourceDoc, derivedDoc)
	}
	if len(suggested) == 0 {
		fmt.Println(text(
			"No mapping suggestions. Please set source_sections manually if needed.",
			"No mapping suggestions. Please set source_sections manually if needed.",
		))
		return 0
	}

	csv := formatSourceSectionMappings(mappingsFromSuggestions(suggested))
	fmt.Println(text("Suggested source_sections CSV:", "Suggested source_sections CSV:"))
	fmt.Println(csv)
	fmt.Println(text("Suggestion details:", "Suggestion details:"))
	for _, suggestion := range suggested {
		token := formatSingleSourceSectionMapping(suggestion.Mapping)
		if token == "" {
			continue
		}
		if suggestion.Relaxed {
			fmt.Printf("  %s [%s; %s]\n", token, suggestionRuleLabel(suggestion.Rule), text("relaxed by smart-auto", "relaxed by smart-auto"))
			continue
		}
		fmt.Printf("  %s [%s]\n", token, suggestionRuleLabel(suggestion.Rule))
	}
	return 0
}
