// commands_check_tree.go handles single-file and tree-wide validation commands.
// commands_check_tree.go는 단일 문서와 트리 전체 검증 명령을 처리합니다.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agd/internal/agd"
)

func commandCheckEasy(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, text("usage: agd check <file.agd>", "usage: agd check <file.agd>"))
		return 2
	}
	path, err := resolveAGDPathEasy(args[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	return commandLint([]string{path})
}

type treeCheckResult struct {
	Path     string
	Errors   []string
	Warnings []string
}

func lintAndAuditFile(root, filePath string) treeCheckResult {
	result := treeCheckResult{
		Path:     filePath,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	doc, parseErrs, err := agd.LoadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("read failed: %v", err))
		return result
	}
	if len(parseErrs) > 0 {
		for _, parseErr := range parseErrs {
			result.Errors = append(result.Errors, fmt.Sprintf("parse error: %s", parseErr))
		}
		return result
	}

	for _, validationErr := range agd.Validate(doc) {
		result.Errors = append(result.Errors, fmt.Sprintf("lint error: %s", validationErr))
	}
	for _, crossErr := range lintAuthorityAndConflicts(filePath, doc) {
		result.Errors = append(result.Errors, fmt.Sprintf("lint error: %s", crossErr))
	}

	result.Warnings = append(result.Warnings, collectDesignWarnings(root, filePath, doc)...)
	return result
}

func collectDesignWarnings(root, filePath string, doc *agd.Document) []string {
	warnings := make([]string, 0)
	if doc == nil {
		return warnings
	}

	rel := relativePathForReport(root, filePath)
	rel = filepath.ToSlash(filepath.Clean(rel))
	authority := strings.ToLower(strings.TrimSpace(doc.Meta["authority"]))
	sourceDoc := strings.TrimSpace(doc.Meta["source_doc"])
	sourceSections := strings.TrimSpace(doc.Meta["source_sections"])

	if strings.HasPrefix(rel, "10_source/") {
		if authority == "" {
			warnings = append(warnings, "10_source document should set meta.authority: source")
		} else if authority != "source" {
			warnings = append(warnings, fmt.Sprintf("10_source document should be authority=source (got %q)", authority))
		}
		if strings.HasPrefix(rel, "10_source/service/") && doc.SectionByID("SYS-070") == nil {
			warnings = append(warnings, "service source document should include SYS-070 (Abort Metrics and Guardrails)")
		}
	}

	if strings.HasPrefix(rel, "20_derived/") {
		if authority != "derived" {
			warnings = append(warnings, "20_derived document should set meta.authority: derived")
		}
		if sourceDoc == "" {
			warnings = append(warnings, "20_derived document should set meta.source_doc")
		}
		if sourceSections == "" {
			warnings = append(warnings, "20_derived document should set meta.source_sections (auto/strict-auto/smart-auto recommended)")
		}
		if sourceDoc != "" {
			if sourcePath, err := resolveSourceDocPathForLint(filePath, sourceDoc); err == nil {
				sourceRel := relativePathForReport(root, sourcePath)
				sourceRel = filepath.ToSlash(filepath.Clean(sourceRel))
				if !strings.HasPrefix(sourceRel, "10_source/") {
					warnings = append(warnings, fmt.Sprintf("derived source_doc should point to 10_source (got %s)", sourceDoc))
				}
			}
		}
		if strings.HasPrefix(rel, "20_derived/frontend/") && doc.SectionByID("FP-070") == nil {
			warnings = append(warnings, "frontend derived document should include FP-070 (Assumptions and Risk Signals)")
		}
		if strings.HasPrefix(rel, "20_derived/qa/") && doc.SectionByID("QA-050") == nil {
			warnings = append(warnings, "qa derived document should include QA-050 (Assumptions and Risk Signals)")
		}
	}

	if strings.HasPrefix(rel, "30_shared/postmortem/") {
		if doc.SectionByID("PM-015") == nil {
			warnings = append(warnings, "postmortem should include PM-015 (Decision Chain and Assumptions)")
		}
		if doc.SectionByID("PM-050") == nil {
			warnings = append(warnings, "postmortem should include PM-050 (Source Feedback Completion)")
		}
	}

	missingImpact := 0
	for _, change := range doc.Changes {
		if change == nil {
			continue
		}
		if strings.TrimSpace(change.Impact) == "" {
			missingImpact++
		}
	}
	if missingImpact > 0 {
		warnings = append(warnings, fmt.Sprintf("%d change entries are missing impact", missingImpact))
	}

	return warnings
}

func isUnderArchiveDir(root, filePath string) bool {
	rel := relativePathForReport(root, filePath)
	rel = filepath.ToSlash(filepath.Clean(rel))
	return rel == "90_archive" || strings.HasPrefix(rel, "90_archive/")
}

func relativePathForReport(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(rel)
}
