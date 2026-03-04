// commands_role_set.go handles source/derived role assignment commands.
// commands_role_set.go는 source/derived 역할 지정 명령을 처리합니다.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agd/internal/agd"
)

func commandRoleSetEasy(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, text(
			"usage: agd role-set <file.agd> <source|derived> [source-doc] [source-sections]", "usage: agd role-set <file.agd> <source|derived> [source-doc] [source-sections]",
		))
		fmt.Fprintln(os.Stderr, text(
			"source-sections: SEC-ID / SRC-ID->DST-ID / SRC-ID=>DST-ID CSV / auto / strict-auto / smart-auto",
			"source-sections: SEC-ID / SRC-ID->DST-ID / SRC-ID=>DST-ID CSV / auto / strict-auto / smart-auto",
		))
		return 2
	}

	filePath, err := resolveAGDPathEasy(args[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}

	role := strings.ToLower(strings.TrimSpace(args[1]))
	if role != "source" && role != "derived" {
		fmt.Fprintln(os.Stderr, text("role-set error: role must be source or derived", "role-set error: role must be source or derived"))
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

	doc.Meta["authority"] = role
	delete(doc.Meta, "source_doc")
	delete(doc.Meta, "source_sections")

	reason := ""
	impact := ""
	switch role {
	case "source":
		reason = "document authority set to source"
		impact = "document now treated as the authority source for downstream derived docs"
	case "derived":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, text(
				"role-set error: derived requires source-doc",
				"role-set error: derived requires source-doc",
			))
			return 2
		}
		sourceDocRef := strings.TrimSpace(args[2])
		sourcePath, err := resolveSourceDocPathForLint(filePath, sourceDocRef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return 1
		}
		storedSource := sourcePath
		if rel, relErr := filepath.Rel(filepath.Dir(filePath), sourcePath); relErr == nil {
			storedSource = filepath.Clean(rel)
		}
		doc.Meta["source_doc"] = storedSource

		if len(args) >= 4 {
			sourceSections := strings.TrimSpace(strings.Join(args[3:], " "))
			if sourceSections != "" {
				mode := strings.ToLower(sourceSections)
				autoMode := mode == "auto" || mode == "suggest"
				strictAutoMode := mode == "strict-auto" || mode == "auto-strict" || mode == "strict"
				smartAutoMode := mode == "smart-auto" || mode == "auto-smart" || mode == "smart"
				if autoMode || strictAutoMode || smartAutoMode {
					sourceDoc, sourceParseErrs, err := agd.LoadFile(sourcePath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "read failed: %v\n", err)
						return 1
					}
					if len(sourceParseErrs) > 0 {
						for _, e := range sourceParseErrs {
							fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
						}
						return 1
					}

					suggested := suggestSourceSectionMappingsDetailed(sourceDoc, doc, strictAutoMode || smartAutoMode)
					if smartAutoMode {
						suggested = applySmartAutoRelaxation(suggested, sourceDoc, doc)
					}
					if len(suggested) == 0 {
						fmt.Println(text(
							"No auto mapping suggestions found; source_sections left empty.", "No auto mapping suggestions found; source_sections left empty.",
						))
					} else {
						generated := formatSourceSectionMappings(mappingsFromSuggestions(suggested))
						doc.Meta["source_sections"] = generated
						modeName := "auto"
						if strictAutoMode {
							modeName = "strict-auto"
						}
						if smartAutoMode {
							modeName = "smart-auto"
						}
						fmt.Printf(text(
							"Auto-generated source_sections (%s): %s\n", "Auto-generated source_sections (%s): %s\n",
						), modeName, generated)
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
					}
				} else {
					_, parseErrs := parseSourceSectionMappings(sourceSections)
					if len(parseErrs) > 0 {
						for _, parseErr := range parseErrs {
							fmt.Fprintf(os.Stderr, "role-set error: %s\n", parseErr)
						}
						fmt.Fprintln(os.Stderr, text(
							"hint: use auto, strict-auto, or smart-auto",
							"hint: use auto, strict-auto, or smart-auto",
						))
						return 2
					}
					doc.Meta["source_sections"] = sourceSections
				}
			}
		}
		reason = fmt.Sprintf("document authority set to derived (source=%s)", storedSource)
		impact = fmt.Sprintf("derived-sync policy now references source_doc=%s", storedSource)
	}

	if len(doc.Sections) == 0 {
		fmt.Fprintln(os.Stderr, text(
			"role-set error: document has no section to attach @change target",
			"role-set error: document has no section to attach @change target",
		))
		return 1
	}
	change := &agd.Change{
		ID:     nextChangeID(doc),
		Target: strings.TrimSpace(doc.Sections[0].ID),
		Date:   time.Now().Format("2006-01-02"),
		Author: "ai-agent",
		Reason: reason,
		Impact: impact,
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
	crossDocErrors := lintAuthorityAndConflicts(filePath, doc)
	if len(crossDocErrors) > 0 {
		for _, e := range crossDocErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}

	if err := os.WriteFile(filePath, []byte(agd.Serialize(doc)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		return 1
	}

	if role == "source" {
		fmt.Printf(text("Set document role to source -> %s\n", "Set document role to source -> %s\n"), filePath)
		return 0
	}
	fmt.Printf(text("Set document role to derived -> %s\n", "Set document role to derived -> %s\n"), filePath)
	return 0
}
