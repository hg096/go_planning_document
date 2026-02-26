// commands_lint_mapping.go handles linting and source-section mapping suggestions.
// commands_lint_mapping.go는 lint와 source 섹션 매핑 제안을 처리합니다.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agd/internal/agd"
)

func commandLint(args []string) int {
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "usage: agd lint <file.agd>\n")
		return 2
	}

	doc, parseErrs, err := agd.LoadFile(args[0])
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

	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}

	crossDocErrors := lintAuthorityAndConflicts(args[0], doc)
	if len(crossDocErrors) > 0 {
		for _, e := range crossDocErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}

	fmt.Printf("OK: %s is valid AGD v0.1\n", args[0])
	return 0
}

func lintAuthorityAndConflicts(filePath string, doc *agd.Document) []string {
	errors := make([]string, 0)
	authority := strings.ToLower(strings.TrimSpace(doc.Meta["authority"]))
	sourceDocRef := strings.TrimSpace(doc.Meta["source_doc"])
	sourceSectionsRaw := strings.TrimSpace(doc.Meta["source_sections"])

	if authority == "" {
		if sourceDocRef != "" || sourceSectionsRaw != "" {
			errors = append(errors, "meta.source_doc/source_sections is set but meta.authority is missing (use source or derived)")
		}
		return errors
	}

	if authority != "source" && authority != "derived" {
		errors = append(errors, fmt.Sprintf("meta.authority must be source or derived (got %q)", authority))
		return errors
	}

	if authority == "source" {
		if sourceDocRef != "" {
			errors = append(errors, "source document must not define meta.source_doc")
		}
		if sourceSectionsRaw != "" {
			errors = append(errors, "source document must not define meta.source_sections")
		}
		return errors
	}

	if sourceDocRef == "" {
		errors = append(errors, "derived document requires meta.source_doc")
		return errors
	}

	sourcePath, err := resolveSourceDocPathForLint(filePath, sourceDocRef)
	if err != nil {
		errors = append(errors, err.Error())
		return errors
	}
	if strings.EqualFold(filepath.Clean(sourcePath), filepath.Clean(filePath)) {
		errors = append(errors, "meta.source_doc cannot point to itself")
		return errors
	}

	sourceDoc, parseErrs, err := agd.LoadFile(sourcePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("cannot read source_doc %q: %v", sourceDocRef, err))
		return errors
	}
	if len(parseErrs) > 0 {
		errors = append(errors, fmt.Sprintf("source_doc %q has parse errors", sourceDocRef))
		for _, e := range parseErrs {
			errors = append(errors, fmt.Sprintf("source_doc parse error: %s", e))
		}
		return errors
	}

	sourceAuthority := strings.ToLower(strings.TrimSpace(sourceDoc.Meta["authority"]))
	if sourceAuthority == "derived" {
		errors = append(errors, fmt.Sprintf("source_doc %q is also derived; source_doc should point to a source document", sourceDocRef))
	}

	explicitMapping := strings.TrimSpace(sourceSectionsRaw) != ""
	mappings := make([]sectionMapping, 0)
	if explicitMapping {
		parsedMappings, parseErrs := parseSourceSectionMappings(sourceSectionsRaw)
		if len(parseErrs) > 0 {
			errors = append(errors, parseErrs...)
		}
		mappings = append(mappings, parsedMappings...)
	}
	if len(mappings) == 0 {
		derivedByID := make(map[string]bool)
		for _, section := range doc.Sections {
			derivedByID[strings.ToUpper(strings.TrimSpace(section.ID))] = true
		}
		for _, section := range sourceDoc.Sections {
			id := strings.TrimSpace(section.ID)
			if id == "" {
				continue
			}
			if derivedByID[strings.ToUpper(id)] {
				mappings = append(mappings, sectionMapping{SourceID: id, DerivedID: id, Strict: true})
			}
		}
	}
	if len(mappings) == 0 {
		return errors
	}

	for _, mapping := range mappings {
		src := sourceDoc.SectionByID(mapping.SourceID)
		if src == nil {
			errors = append(errors, fmt.Sprintf("source_doc %q is missing section %s (mapped to %s)", sourceDocRef, mapping.SourceID, mapping.DerivedID))
			continue
		}
		dst := doc.SectionByID(mapping.DerivedID)
		if dst == nil {
			if explicitMapping {
				errors = append(errors, fmt.Sprintf("derived document is missing mapped section %s (from %s)", mapping.DerivedID, mapping.SourceID))
			}
			continue
		}

		if mapping.Strict {
			if normalizeComparableText(src.Title) != normalizeComparableText(dst.Title) {
				errors = append(errors, fmt.Sprintf("section mapping %s=>%s title conflicts with source_doc", mapping.SourceID, mapping.DerivedID))
			}
			if normalizeComparableText(src.Summary) != normalizeComparableText(dst.Summary) {
				errors = append(errors, fmt.Sprintf("section mapping %s=>%s summary conflicts with source_doc", mapping.SourceID, mapping.DerivedID))
			}
		}
	}
	return errors
}

func resolveSourceDocPathForLint(currentFilePath, sourceDocRef string) (string, error) {
	ref := strings.TrimSpace(sourceDocRef)
	if ref == "" {
		return "", fmt.Errorf("meta.source_doc is empty")
	}

	if fileExists(ref) {
		return filepath.Clean(ref), nil
	}
	siblingCandidate := ensureAGDExt(filepath.Join(filepath.Dir(currentFilePath), ref))
	if fileExists(siblingCandidate) {
		return filepath.Clean(siblingCandidate), nil
	}

	if hasPathSeparator(ref) {
		candidate := ensureAGDExt(ref)
		if fileExists(candidate) {
			return filepath.Clean(candidate), nil
		}
		if !filepath.IsAbs(ref) {
			relFromDoc := ensureAGDExt(filepath.Join(filepath.Dir(currentFilePath), ref))
			if fileExists(relFromDoc) {
				return filepath.Clean(relFromDoc), nil
			}
		}
	}

	resolved, err := resolveAGDPathEasy(ref, true)
	if err != nil {
		return "", fmt.Errorf("cannot resolve meta.source_doc %q: %v", sourceDocRef, err)
	}
	return resolved, nil
}

func normalizeComparableText(value string) string {
	lowered := strings.ToLower(strings.TrimSpace(value))
	return strings.Join(strings.Fields(lowered), " ")
}

type sectionMapping struct {
	SourceID  string
	DerivedID string
	Strict    bool
}

type sectionSuggestion struct {
	Mapping sectionMapping
	Rule    string
	Relaxed bool
}

func parseSourceSectionMappings(raw string) ([]sectionMapping, []string) {
	mappings := make([]sectionMapping, 0)
	errors := make([]string, 0)
	seen := make(map[string]bool)

	for _, token := range splitCSV(raw) {
		rel := strings.TrimSpace(token)
		if rel == "" {
			continue
		}

		var sourceID string
		var derivedID string
		strict := false
		if strings.Contains(rel, "=>") {
			parts := strings.SplitN(rel, "=>", 2)
			sourceID = strings.TrimSpace(parts[0])
			derivedID = strings.TrimSpace(parts[1])
			strict = true
		} else if strings.Contains(rel, "=<") {
			parts := strings.SplitN(rel, "=<", 2)
			derivedID = strings.TrimSpace(parts[0])
			sourceID = strings.TrimSpace(parts[1])
			strict = true
		} else if strings.Contains(rel, "->") {
			parts := strings.SplitN(rel, "->", 2)
			sourceID = strings.TrimSpace(parts[0])
			derivedID = strings.TrimSpace(parts[1])
		} else if strings.Contains(rel, "<-") {
			parts := strings.SplitN(rel, "<-", 2)
			derivedID = strings.TrimSpace(parts[0])
			sourceID = strings.TrimSpace(parts[1])
		} else {
			sourceID = rel
			derivedID = rel
			strict = true
		}

		if sourceID == "" || derivedID == "" {
			errors = append(errors, fmt.Sprintf("invalid source_sections mapping %q (use SEC-ID, SRC-ID->DST-ID, or SRC-ID=>DST-ID)", rel))
			continue
		}

		key := strings.ToUpper(sourceID) + "->" + strings.ToUpper(derivedID) + fmt.Sprintf("#%t", strict)
		if seen[key] {
			continue
		}
		seen[key] = true
		mappings = append(mappings, sectionMapping{SourceID: sourceID, DerivedID: derivedID, Strict: strict})
	}
	return mappings, errors
}

func formatSingleSourceSectionMapping(mapping sectionMapping) string {
	src := strings.TrimSpace(mapping.SourceID)
	dst := strings.TrimSpace(mapping.DerivedID)
	if src == "" || dst == "" {
		return ""
	}
	if strings.EqualFold(src, dst) && mapping.Strict {
		return src
	}
	if mapping.Strict {
		return fmt.Sprintf("%s=>%s", src, dst)
	}
	return fmt.Sprintf("%s->%s", src, dst)
}

func formatSourceSectionMappings(mappings []sectionMapping) string {
	parts := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		token := formatSingleSourceSectionMapping(mapping)
		if token != "" {
			parts = append(parts, token)
		}
	}
	return strings.Join(parts, ",")
}

func suggestSourceSectionMappings(sourceDoc, derivedDoc *agd.Document) []sectionMapping {
	return mappingsFromSuggestions(suggestSourceSectionMappingsDetailed(sourceDoc, derivedDoc, false))
}

func mappingsFromSuggestions(suggestions []sectionSuggestion) []sectionMapping {
	out := make([]sectionMapping, 0, len(suggestions))
	for _, suggestion := range suggestions {
		out = append(out, suggestion.Mapping)
	}
	return out
}

func suggestSourceSectionMappingsDetailed(sourceDoc, derivedDoc *agd.Document, strict bool) []sectionSuggestion {
	if sourceDoc == nil || derivedDoc == nil {
		return nil
	}

	sourceByID := make(map[string]*agd.Section)
	sourceBySuffix := make(map[string][]*agd.Section)
	sourceByTitle := make(map[string][]*agd.Section)
	for _, src := range sourceDoc.Sections {
		if src == nil {
			continue
		}
		id := strings.TrimSpace(src.ID)
		if id == "" {
			continue
		}
		sourceByID[strings.ToUpper(id)] = src
		if suffix := idNumericSuffix(id); suffix != "" {
			sourceBySuffix[suffix] = append(sourceBySuffix[suffix], src)
		}
		titleKey := normalizeComparableText(src.Title)
		if titleKey != "" {
			sourceByTitle[titleKey] = append(sourceByTitle[titleKey], src)
		}
	}

	usedSource := make(map[string]bool)
	usedDerived := make(map[string]bool)
	suggestions := make([]sectionSuggestion, 0)

	// Pass 1: exact ID match.
	for _, dst := range derivedDoc.Sections {
		if dst == nil {
			continue
		}
		dstID := strings.TrimSpace(dst.ID)
		if dstID == "" {
			continue
		}
		src := sourceByID[strings.ToUpper(dstID)]
		if src == nil {
			continue
		}
		srcID := strings.TrimSpace(src.ID)
		usedSource[strings.ToUpper(srcID)] = true
		usedDerived[strings.ToUpper(dstID)] = true
		suggestions = append(suggestions, sectionSuggestion{
			Mapping: sectionMapping{
				SourceID:  srcID,
				DerivedID: dstID,
				Strict:    strict,
			},
			Rule: "id",
		})
	}

	// Pass 2: unique numeric suffix match (example: SYS-020 -> FP-020).
	for _, dst := range derivedDoc.Sections {
		if dst == nil {
			continue
		}
		dstID := strings.TrimSpace(dst.ID)
		if dstID == "" || usedDerived[strings.ToUpper(dstID)] {
			continue
		}
		suffix := idNumericSuffix(dstID)
		if suffix == "" {
			continue
		}

		candidates := make([]*agd.Section, 0, 1)
		for _, src := range sourceBySuffix[suffix] {
			if src == nil {
				continue
			}
			srcID := strings.TrimSpace(src.ID)
			if srcID == "" || usedSource[strings.ToUpper(srcID)] {
				continue
			}
			candidates = append(candidates, src)
		}
		if len(candidates) != 1 {
			continue
		}

		srcID := strings.TrimSpace(candidates[0].ID)
		usedSource[strings.ToUpper(srcID)] = true
		usedDerived[strings.ToUpper(dstID)] = true
		suggestions = append(suggestions, sectionSuggestion{
			Mapping: sectionMapping{
				SourceID:  srcID,
				DerivedID: dstID,
				Strict:    strict,
			},
			Rule: "suffix",
		})
	}

	// Pass 3: unique normalized title match.
	for _, dst := range derivedDoc.Sections {
		if dst == nil {
			continue
		}
		dstID := strings.TrimSpace(dst.ID)
		if dstID == "" || usedDerived[strings.ToUpper(dstID)] {
			continue
		}
		titleKey := normalizeComparableText(dst.Title)
		if titleKey == "" {
			continue
		}

		candidates := make([]*agd.Section, 0, 1)
		for _, src := range sourceByTitle[titleKey] {
			if src == nil {
				continue
			}
			srcID := strings.TrimSpace(src.ID)
			if srcID == "" || usedSource[strings.ToUpper(srcID)] {
				continue
			}
			candidates = append(candidates, src)
		}
		if len(candidates) != 1 {
			continue
		}

		srcID := strings.TrimSpace(candidates[0].ID)
		usedSource[strings.ToUpper(srcID)] = true
		usedDerived[strings.ToUpper(dstID)] = true
		suggestions = append(suggestions, sectionSuggestion{
			Mapping: sectionMapping{
				SourceID:  srcID,
				DerivedID: dstID,
				Strict:    strict,
			},
			Rule: "title",
		})
	}

	return suggestions
}

func applySmartAutoRelaxation(suggestions []sectionSuggestion, sourceDoc, derivedDoc *agd.Document) []sectionSuggestion {
	if sourceDoc == nil || derivedDoc == nil {
		return suggestions
	}
	out := make([]sectionSuggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		if !suggestion.Mapping.Strict {
			out = append(out, suggestion)
			continue
		}
		src := sourceDoc.SectionByID(suggestion.Mapping.SourceID)
		dst := derivedDoc.SectionByID(suggestion.Mapping.DerivedID)
		if src == nil || dst == nil {
			out = append(out, suggestion)
			continue
		}
		sameTitle := normalizeComparableText(src.Title) == normalizeComparableText(dst.Title)
		sameSummary := normalizeComparableText(src.Summary) == normalizeComparableText(dst.Summary)
		if sameTitle && sameSummary {
			out = append(out, suggestion)
			continue
		}
		relaxed := suggestion
		relaxed.Mapping.Strict = false
		relaxed.Relaxed = true
		out = append(out, relaxed)
	}
	return out
}

func suggestionRuleLabel(rule string) string {
	switch strings.ToLower(strings.TrimSpace(rule)) {
	case "id":
		return text("exact ID match", "exact ID match")
	case "suffix":
		return text("numeric suffix match", "numeric suffix match")
	case "title":
		return text("normalized title match", "normalized title match")
	default:
		return text("heuristic match", "heuristic match")
	}
}

func idNumericSuffix(sectionID string) string {
	id := strings.TrimSpace(sectionID)
	if id == "" {
		return ""
	}
	i := len(id) - 1
	for i >= 0 {
		ch := id[i]
		if ch < '0' || ch > '9' {
			break
		}
		i--
	}
	if i == len(id)-1 {
		return ""
	}
	return id[i+1:]
}
