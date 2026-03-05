// commands_service_gate.go handles service freshness gate and tree check commands.
// commands_service_gate.go는 서비스 변경 게이트와 트리 점검 명령을 처리합니다.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"agd/internal/agd"
)

type serviceGateDecision struct {
	LatestChange *agd.Change
	LatestDate   time.Time
	Errors       []string
}

func commandCheckAllEasy(args []string) int {
	root := docsRootDir
	strict := false
	includeArchive := false
	rootProvided := false
	for _, raw := range args {
		token := strings.TrimSpace(raw)
		switch strings.ToLower(token) {
		case "":
			continue
		case "--strict":
			strict = true
		case "--include-archive":
			includeArchive = true
		default:
			if strings.HasPrefix(token, "-") || rootProvided {
				fmt.Fprintln(os.Stderr, text(
					"usage: agd check-all [root] [--strict] [--include-archive]", "usage: agd check-all [root] [--strict] [--include-archive]",
				))
				return 2
			}
			root = token
			rootProvided = true
		}
	}
	if root == "" {
		root = docsRootDir
	}

	if strings.EqualFold(filepath.Clean(root), filepath.Clean(docsRootDir)) {
		if err := ensureDocsRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "check-all failed: %v\n", err)
			return 1
		}
	}

	files, err := listAGDFiles(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "check-all failed: %v\n", err)
		return 1
	}

	filtered := make([]string, 0, len(files))
	for _, file := range files {
		if !includeArchive && isUnderArchiveDir(root, file) {
			continue
		}
		filtered = append(filtered, file)
	}
	if len(filtered) == 0 {
		fmt.Println(text(
			fmt.Sprintf("No AGD files found under %s", root),
			fmt.Sprintf("No AGD files found under %s", root),
		))
		return 0
	}

	results := make([]treeCheckResult, 0, len(filtered))
	passed := 0
	warned := 0
	failed := 0

	for _, file := range filtered {
		result := lintAndAuditFile(root, file)
		results = append(results, result)
		if len(result.Errors) > 0 {
			failed++
			continue
		}
		if len(result.Warnings) > 0 {
			warned++
			continue
		}
		passed++
	}

	for _, result := range results {
		rel := relativePathForReport(root, result.Path)
		if len(result.Errors) > 0 {
			fmt.Printf("FAIL: %s\n", rel)
			for _, item := range result.Errors {
				fmt.Printf("  - %s\n", item)
			}
			for _, item := range result.Warnings {
				fmt.Printf("  ~ %s\n", item)
			}
			continue
		}
		if len(result.Warnings) > 0 {
			fmt.Printf("WARN: %s\n", rel)
			for _, item := range result.Warnings {
				fmt.Printf("  ~ %s\n", item)
			}
			continue
		}
		fmt.Printf("OK:   %s\n", rel)
	}

	fmt.Println("")
	fmt.Printf(text(
		"Summary: checked=%d, passed=%d, warned=%d, failed=%d\n", "Summary: checked=%d, passed=%d, warned=%d, failed=%d\n",
	), len(results), passed, warned, failed)

	if failed > 0 {
		return 1
	}
	if strict && warned > 0 {
		fmt.Println(text(
			"Strict mode: warnings are treated as failures.", "Strict mode: warnings are treated as failures.",
		))
		return 1
	}
	return 0
}

func printServiceGateUsage() {
	fmt.Fprintln(os.Stderr, text(
		"usage: agd service-gate [service-doc.agd] [--max-age-days N] [--allow-missing-impact]", "usage: agd service-gate [service-doc.agd] [--max-age-days N] [--allow-missing-impact]",
	))
}

func commandServiceGateEasy(args []string) int {
	docInput := ""
	maxAgeDays := 0
	allowMissingImpact := false

	for i := 0; i < len(args); i++ {
		token := strings.TrimSpace(args[i])
		switch strings.ToLower(token) {
		case "":
			continue
		case "--allow-missing-impact":
			allowMissingImpact = true
		case "--max-age-days":
			if i+1 >= len(args) {
				printServiceGateUsage()
				return 2
			}
			parsed, err := strconv.Atoi(strings.TrimSpace(args[i+1]))
			if err != nil || parsed < 0 {
				fmt.Fprintln(os.Stderr, text(
					"service-gate error: --max-age-days must be a non-negative integer",
					"service-gate error: --max-age-days must be a non-negative integer",
				))
				return 2
			}
			maxAgeDays = parsed
			i++
		default:
			if strings.HasPrefix(token, "-") || docInput != "" {
				printServiceGateUsage()
				return 2
			}
			docInput = token
		}
	}

	serviceDocPath, err := resolveServiceDocForGate(docInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}

	doc, parseErrs, err := agd.LoadFile(serviceDocPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service-gate failed: %v\n", err)
		return 1
	}
	if len(parseErrs) > 0 {
		fmt.Println(text("SERVICE CHANGE BLOCKED", "SERVICE CHANGE BLOCKED"))
		fmt.Printf(text("Target doc: %s\n", "Target doc: %s\n"), serviceDocPath)
		for _, parseErr := range parseErrs {
			fmt.Printf("  - parse error: %s\n", parseErr)
		}
		return 1
	}

	now := time.Now()
	guardErrors := make([]string, 0)
	for _, validationErr := range agd.Validate(doc) {
		guardErrors = append(guardErrors, fmt.Sprintf("lint error: %s", validationErr))
	}
	for _, crossErr := range lintAuthorityAndConflicts(serviceDocPath, doc) {
		guardErrors = append(guardErrors, fmt.Sprintf("lint error: %s", crossErr))
	}
	decision := evaluateServiceGate(doc, now, maxAgeDays, allowMissingImpact)
	guardErrors = append(guardErrors, decision.Errors...)

	if len(guardErrors) > 0 {
		fmt.Println(text("SERVICE CHANGE BLOCKED", "SERVICE CHANGE BLOCKED"))
		fmt.Printf(text("Target doc: %s\n", "Target doc: %s\n"), serviceDocPath)
		for _, guardErr := range guardErrors {
			fmt.Printf("  - %s\n", guardErr)
		}
		fmt.Println(text(
			"Action: update service-logic doc first (@section + @change reason/impact), then run service-gate again.", "Action: update service-logic doc first (@section + @change reason/impact), then run service-gate again.",
		))
		return 1
	}

	latestID := ""
	if decision.LatestChange != nil {
		latestID = strings.TrimSpace(decision.LatestChange.ID)
	}
	if latestID == "" {
		latestID = "(no-id)"
	}
	ageDays := dayDistance(now, decision.LatestDate)
	fmt.Println(text("SERVICE CHANGE ALLOWED", "SERVICE CHANGE ALLOWED"))
	fmt.Printf(text("Target doc: %s\n", "Target doc: %s\n"), serviceDocPath)
	fmt.Printf(text(
		"Latest change: %s (%s, age=%d day(s), limit=%d day(s))\n",
		"Latest change: %s (%s, age=%d day(s), limit=%d day(s))\n",
	), latestID, decision.LatestDate.Format("2006-01-02"), ageDays, maxAgeDays)
	fmt.Println(text(
		"Guard passed. You can proceed with service-code edits.",
		"Guard passed. You can proceed with service-code edits.",
	))
	return 0
}

func resolveServiceDocForGate(raw string) (string, error) {
	if err := ensureDocsRoot(); err != nil {
		return "", err
	}
	serviceRoot := filepath.Join(docsRootDir, "10_source", "service")
	docInput := strings.TrimSpace(raw)
	if docInput != "" {
		resolved, err := resolveAGDPathEasy(docInput, true)
		if err != nil {
			return "", err
		}
		if !isPathUnderRoot(serviceRoot, resolved) {
			return "", fmt.Errorf(text(
				"service-gate error: document must be under %s",
				"service-gate error: document must be under %s",
			), serviceRoot)
		}
		return filepath.Clean(resolved), nil
	}

	files, err := listAGDFiles(serviceRoot)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf(text(
			"service-gate error: no .agd file found under %s",
			"service-gate error: no .agd file found under %s",
		), serviceRoot)
	}
	if len(files) > 1 {
		candidates := make([]string, 0, len(files))
		for _, file := range files {
			if rel, relErr := filepath.Rel(docsRootDir, file); relErr == nil {
				candidates = append(candidates, filepath.Join(docsRootDir, rel))
				continue
			}
			candidates = append(candidates, file)
		}
		return "", fmt.Errorf(text(
			"service-gate error: multiple service docs found; specify one: %s", "service-gate error: multiple service docs found; specify one: %s",
		), strings.Join(candidates, ", "))
	}
	return filepath.Clean(files[0]), nil
}

func isPathUnderRoot(root, target string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(target))
	if err != nil {
		return false
	}
	rel = filepath.Clean(rel)
	if rel == "." {
		return true
	}
	if rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func evaluateServiceGate(doc *agd.Document, now time.Time, maxAgeDays int, allowMissingImpact bool) serviceGateDecision {
	result := serviceGateDecision{
		Errors: make([]string, 0),
	}
	if doc == nil {
		result.Errors = append(result.Errors, "document is nil")
		return result
	}
	if strings.ToLower(strings.TrimSpace(doc.Meta["authority"])) != "source" {
		result.Errors = append(result.Errors, "meta.authority must be source")
	}
	if len(doc.Changes) == 0 {
		result.Errors = append(result.Errors, "at least one @change entry is required")
		return result
	}

	hasValidDate := false
	latestIndex := -1
	for i, change := range doc.Changes {
		if change == nil {
			continue
		}
		parsedDate, err := parseChangeDateFlexible(change.Date)
		if err != nil {
			changeID := strings.TrimSpace(change.ID)
			if changeID == "" {
				changeID = fmt.Sprintf("#%d", i+1)
			}
			result.Errors = append(result.Errors, fmt.Sprintf(
				"change %s has invalid date %q (use YYYY-MM-DD)",
				changeID,
				change.Date,
			))
			continue
		}
		if !hasValidDate || parsedDate.After(result.LatestDate) || (parsedDate.Equal(result.LatestDate) && i > latestIndex) {
			hasValidDate = true
			latestIndex = i
			result.LatestDate = parsedDate
			result.LatestChange = change
		}
	}
	if !hasValidDate {
		result.Errors = append(result.Errors, "no valid @change date found")
		return result
	}

	ageDays := dayDistance(now, result.LatestDate)
	if ageDays < 0 {
		result.Errors = append(result.Errors, fmt.Sprintf(
			"latest @change date is in the future: %s",
			result.LatestDate.Format("2006-01-02"),
		))
	} else if ageDays > maxAgeDays {
		result.Errors = append(result.Errors, fmt.Sprintf(
			"latest @change is %d day(s) old; max allowed is %d day(s)",
			ageDays,
			maxAgeDays,
		))
	}

	if result.LatestChange != nil && strings.TrimSpace(result.LatestChange.Reason) == "" {
		result.Errors = append(result.Errors, "latest @change is missing reason")
	}
	if !allowMissingImpact && result.LatestChange != nil && strings.TrimSpace(result.LatestChange.Impact) == "" {
		result.Errors = append(result.Errors, "latest @change is missing impact")
	}
	return result
}

func parseChangeDateFlexible(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"20060102",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if layout == time.RFC3339 {
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return parsed.In(time.Local), nil
			}
			continue
		}
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func dayDistance(now, then time.Time) int {
	loc := now.Location()
	nowLocal := now.In(loc)
	thenLocal := then.In(loc)

	nowDate := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, time.UTC)
	thenDate := time.Date(thenLocal.Year(), thenLocal.Month(), thenLocal.Day(), 0, 0, 0, 0, time.UTC)
	return int(nowDate.Sub(thenDate).Hours() / 24)
}
