// commands_code_plan_check.go validates code-change and planning-doc linkage.
// commands_code_plan_check.go는 코드 변경과 기획 문서 연결 검증을 처리합니다.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agd/internal/agd"
)

var (
	codePlanDefaultCacheRelPath       = filepath.Join("00_agd", ".cache", "code_plan_validation.json")
	codePlanDefaultScopePolicyRelPath = filepath.Join("00_agd", "policy", "code_plan_scope.txt")
	codePlanDefaultRelationPolicyPath = filepath.Join("00_agd", "policy", "code_plan_relations.txt")
)

type codePlanMode string

const (
	codePlanModeAuto  codePlanMode = "auto"
	codePlanModeGit   codePlanMode = "git"
	codePlanModeCache codePlanMode = "cache"
)

type codePlanGitSource string

const (
	codePlanGitSourceWorktree codePlanGitSource = "worktree"
	codePlanGitSourceStaged   codePlanGitSource = "staged"
)

type codePlanScopePolicy struct {
	AllowedExt      map[string]bool
	IncludePrefixes []string
	ExcludePrefixes []string
}

type codePlanDocInfo struct {
	AbsPath          string
	RelPath          string
	RefPaths         []string
	SectionIDs       map[string]bool
	HasTag           bool
	ChangeDigest     string
	LatestChangeID   string
	LatestChangeDate string
}

type codePlanDocIssue struct {
	RelPath string
	Message string
}

type codePlanDocResult struct {
	DocPath       string
	DocKey        string
	HasTag        bool
	ChangeUpdated bool
	Passed        bool
}

type codePlanFileResult struct {
	ChangedPath string
	MatchedDocs []codePlanDocResult
	MissingDoc  bool
}

type codePlanGitStatusEntry struct {
	Status string
	Path   string
}

type codePlanRelationPolicy struct {
	Rules []codePlanRelationRule
}

type codePlanRelationRule struct {
	LineNo int
	Raw    string
	Left   codePlanRelationEndpoint
	Right  codePlanRelationEndpoint
}

type codePlanRelationEndpoint struct {
	Raw       string
	DocPath   string
	DocKey    string
	SectionID string
}

type codePlanRelationResult struct {
	ChangedPath    string
	TriggerDocPath string
	TriggerSection string
	RequiredDoc    string
	RequiredSec    string
	ChangeUpdated  bool
	Passed         bool
	Message        string
}

func printCodePlanCheckUsage() {
	fmt.Fprintln(os.Stderr, text(
		"usage: agd code-plan-check [root] [--mode auto|git|cache] [--git-source staged|worktree] [--cache <file>] [--scope-policy <file>] [--relation-policy <file>] [--strict-mapping] [--strict-relation]",
		"usage: agd code-plan-check [root] [--mode auto|git|cache] [--git-source staged|worktree] [--cache <file>] [--scope-policy <file>] [--relation-policy <file>] [--strict-mapping] [--strict-relation]",
	))
}

func commandCodePlanCheckEasy(args []string) int {
	root := docsRootDir
	mode := codePlanModeAuto
	gitSource := codePlanGitSourceWorktree
	cachePathInput := codePlanDefaultCacheRelPath
	scopePolicyInput := codePlanDefaultScopePolicyRelPath
	relationPolicyInput := codePlanDefaultRelationPolicyPath
	strictMapping := false
	strictRelation := false
	rootProvided := false

	for i := 0; i < len(args); i++ {
		token := strings.TrimSpace(args[i])
		switch strings.ToLower(token) {
		case "":
			continue
		case "--mode":
			if i+1 >= len(args) {
				printCodePlanCheckUsage()
				return 2
			}
			parsed, ok := parseCodePlanMode(args[i+1])
			if !ok {
				fmt.Fprintln(os.Stderr, text(
					"code-plan-check error: --mode must be auto|git|cache",
					"code-plan-check error: --mode must be auto|git|cache",
				))
				return 2
			}
			mode = parsed
			i++
		case "--git-source":
			if i+1 >= len(args) {
				printCodePlanCheckUsage()
				return 2
			}
			parsed, ok := parseCodePlanGitSource(args[i+1])
			if !ok {
				fmt.Fprintln(os.Stderr, text(
					"code-plan-check error: --git-source must be staged|worktree",
					"code-plan-check error: --git-source must be staged|worktree",
				))
				return 2
			}
			gitSource = parsed
			i++
		case "--cache":
			if i+1 >= len(args) {
				printCodePlanCheckUsage()
				return 2
			}
			cachePathInput = strings.TrimSpace(args[i+1])
			i++
		case "--scope-policy":
			if i+1 >= len(args) {
				printCodePlanCheckUsage()
				return 2
			}
			scopePolicyInput = strings.TrimSpace(args[i+1])
			i++
		case "--relation-policy":
			if i+1 >= len(args) {
				printCodePlanCheckUsage()
				return 2
			}
			relationPolicyInput = strings.TrimSpace(args[i+1])
			i++
		case "--strict-mapping":
			strictMapping = true
		case "--strict-relation":
			strictRelation = true
		default:
			if strings.HasPrefix(token, "-") || rootProvided {
				printCodePlanCheckUsage()
				return 2
			}
			root = token
			rootProvided = true
		}
	}

	if strings.EqualFold(filepath.Clean(root), filepath.Clean(docsRootDir)) {
		if err := ensureDocsRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "code-plan-check failed: %v\n", err)
			return 1
		}
	}

	repoRoot := detectCodePlanRepoRoot()
	scanRoot := resolveCodePlanPath(repoRoot, root)
	cachePath := resolveCodePlanPath(repoRoot, cachePathInput)
	scopePolicyPath := resolveCodePlanPath(repoRoot, scopePolicyInput)
	relationPolicyPath := resolveCodePlanPath(repoRoot, relationPolicyInput)

	scopePolicy, scopeWarnings, err := loadCodePlanScopePolicy(scopePolicyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "code-plan-check failed: %v\n", err)
		return 1
	}

	cache, err := loadCodePlanCache(cachePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "code-plan-check failed: %v\n", err)
		return 1
	}

	changedFiles, modeUsed, modeWarning, err := collectChangedCodeFiles(repoRoot, mode, gitSource, scopePolicy, cache)
	if err != nil {
		fmt.Fprintf(os.Stderr, "code-plan-check failed: %v\n", err)
		return 1
	}

	docs, docIssues, err := collectCodePlanDocs(scanRoot, repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "code-plan-check failed: %v\n", err)
		return 1
	}

	relationPolicy, relationWarnings, relationEnabled, err := loadCodePlanRelationPolicy(relationPolicyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "code-plan-check failed: %v\n", err)
		return 1
	}

	results := evaluateCodePlanCheck(changedFiles, docs, cache)
	relationResults := make([]codePlanRelationResult, 0)
	if relationEnabled {
		relationResults = evaluateCodePlanRelations(results, docs, relationPolicy, cache)
	}

	failed := 0
	missingDoc := 0
	warned := 0
	for _, item := range results {
		if item.MissingDoc {
			missingDoc++
			if strictMapping {
				failed++
			} else {
				warned++
			}
			continue
		}
		for _, docResult := range item.MatchedDocs {
			if !docResult.Passed {
				failed++
			}
		}
	}

	relationWarned := 0
	relationFailed := 0
	for _, relation := range relationResults {
		if relation.Passed {
			continue
		}
		if strictRelation {
			relationFailed++
		} else {
			relationWarned++
		}
	}

	relationByChanged := make(map[string][]codePlanRelationResult)
	for _, relation := range relationResults {
		key := codePlanNormalizeRelPath(relation.ChangedPath)
		relationByChanged[key] = append(relationByChanged[key], relation)
	}

	fmt.Println(text("CODE PLAN VALIDATION", "코드 기획 검증"))
	fmt.Printf(text("Docs root: %s\n", "문서 루트: %s\n"), codePlanDisplayPath(repoRoot, scanRoot))
	fmt.Printf(text("Mode: requested=%s, used=%s\n", "모드: 요청=%s, 사용=%s\n"), mode, modeUsed)
	fmt.Printf(text("Git source: %s\n", "Git 소스: %s\n"), gitSource)
	fmt.Printf(text("Scope policy: %s\n", "범위 정책: %s\n"), codePlanDisplayPath(repoRoot, scopePolicyPath))
	if relationEnabled {
		fmt.Printf(text("Relation policy: %s (enabled)\n", "연관 정책: %s (활성)\n"), codePlanDisplayPath(repoRoot, relationPolicyPath))
	} else {
		fmt.Printf(text("Relation policy: %s (disabled: file not found)\n", "연관 정책: %s (비활성: 파일 없음)\n"), codePlanDisplayPath(repoRoot, relationPolicyPath))
	}
	fmt.Printf(text("Cache path: %s\n", "캐시 경로: %s\n"), codePlanDisplayPath(repoRoot, cachePath))
	fmt.Printf(text("Changed code files (in scope): %d\n", "변경 코드 파일(범위 적용): %d\n"), len(changedFiles))

	if strings.TrimSpace(modeWarning) != "" {
		fmt.Printf(text("Mode note: %s\n", "모드 참고: %s\n"), strings.TrimSpace(modeWarning))
	}
	for _, warning := range scopeWarnings {
		fmt.Printf("WARN: %s\n", warning)
	}
	for _, warning := range relationWarnings {
		fmt.Printf("WARN: %s\n", warning)
	}
	for _, issue := range docIssues {
		fmt.Printf("WARN: %s (%s)\n", issue.Message, issue.RelPath)
	}

	fmt.Println("")
	if len(results) == 0 {
		fmt.Println(text(
			"No changed code files in scope. Validation passed.",
			"검사 대상 코드 변경이 없습니다. 검증 통과입니다.",
		))
	} else {
		for _, item := range results {
			fmt.Printf("CHANGED: %s\n", item.ChangedPath)
			if item.MissingDoc {
				if strictMapping {
					fmt.Println(text(
						"  FAIL: no matching AGD doc (doc_base_path/section.path overlap not found)",
						"  FAIL: 매칭 AGD 문서 없음 (doc_base_path/section.path 겹침 미검출)",
					))
				} else {
					fmt.Println(text(
						"  WARN: no matching AGD doc (doc_base_path/section.path overlap not found)",
						"  WARN: 매칭 AGD 문서 없음 (doc_base_path/section.path 겹침 미검출)",
					))
				}
				continue
			}
			for _, docResult := range item.MatchedDocs {
				status := "PASS"
				if !docResult.Passed {
					status = "FAIL"
				}
				fmt.Printf("  %s: %s (tag=%s, changeUpdated=%s)\n",
					status,
					docResult.DocPath,
					boolToken(docResult.HasTag),
					boolToken(docResult.ChangeUpdated),
				)
			}
			if relationEnabled {
				relItems := relationByChanged[codePlanNormalizeRelPath(item.ChangedPath)]
				if len(relItems) == 0 {
					fmt.Println(text(
						"  RELATION: no relation-policy match",
						"  RELATION: 연관 정책 매칭 없음",
					))
				} else {
					for _, relation := range relItems {
						status := "PASS"
						if !relation.Passed {
							if strictRelation {
								status = "FAIL"
							} else {
								status = "WARN"
							}
						}
						fmt.Printf("  RELATION %s: %s#%s -> %s#%s (changeUpdated=%s)\n",
							status,
							relation.TriggerDocPath,
							relation.TriggerSection,
							relation.RequiredDoc,
							relation.RequiredSec,
							boolToken(relation.ChangeUpdated),
						)
						if strings.TrimSpace(relation.Message) != "" {
							fmt.Printf("    note: %s\n", strings.TrimSpace(relation.Message))
						}
					}
				}
			}
		}
	}

	fmt.Println("")
	fmt.Printf(text(
		"Summary: changed=%d, docs=%d, warned=%d, failed=%d, missing-doc=%d, relation-checked=%d, relation-warned=%d, relation-failed=%d\n",
		"요약: 변경=%d, 문서=%d, 경고=%d, 실패=%d, 문서미매칭=%d, 연관검사=%d, 연관경고=%d, 연관실패=%d\n",
	), len(changedFiles), len(docs), warned, failed, missingDoc, len(relationResults), relationWarned, relationFailed)

	if failed > 0 || relationFailed > 0 {
		fmt.Println(text("Fix guide:", "수정 가이드:"))
		fmt.Println(text(
			"1) Set meta.doc_base_path and/or section.path to overlap changed code paths.",
			"1) 변경 코드 경로와 겹치도록 meta.doc_base_path 또는 section.path를 설정하세요.",
		))
		fmt.Println(text(
			"2) Add maintenance/incident tag metadata or tag block when flow requires it.",
			"2) 흐름에 맞게 유지보수/오류 태그 메타 또는 태그 블록을 추가하세요.",
		))
		fmt.Println(text(
			"3) Record @change with reason/impact (ex: agd logic-log <doc> <SEC-ID> --reason \"...\" --impact \"...\").",
			"3) @change(reason/impact)를 기록하세요 (예: agd logic-log <doc> <SEC-ID> --reason \"...\" --impact \"...\").",
		))
		fmt.Println(text(
			"4) For relation-policy matches, update required linked docs and append @change there.",
			"4) 연관 정책에 매칭된 반대편 문서도 갱신하고 해당 문서에 @change를 기록하세요.",
		))
		return 1
	}

	fileSnapshot, err := buildCodePlanFileSnapshot(repoRoot, scopePolicy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "code-plan-check failed: cannot update cache snapshot: %v\n", err)
		return 1
	}
	docSnapshot := buildCodePlanDocSnapshot(docs)
	nextCache := &codePlanCache{
		Version:      codePlanCacheVersion,
		UpdatedAt:    time.Now().Format(time.RFC3339),
		RepoRoot:     filepath.Clean(repoRoot),
		FileSnapshot: fileSnapshot,
		DocSnapshot:  docSnapshot,
	}
	if err := saveCodePlanCache(cachePath, nextCache); err != nil {
		fmt.Fprintf(os.Stderr, "code-plan-check failed: cache update error: %v\n", err)
		return 1
	}
	fmt.Printf(text("Cache updated: %s\n", "캐시 업데이트: %s\n"), codePlanDisplayPath(repoRoot, cachePath))
	return 0
}

func parseCodePlanMode(raw string) (codePlanMode, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto", "1":
		return codePlanModeAuto, true
	case "git", "2":
		return codePlanModeGit, true
	case "cache", "3":
		return codePlanModeCache, true
	default:
		return "", false
	}
}

func parseCodePlanGitSource(raw string) (codePlanGitSource, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "worktree", "1":
		return codePlanGitSourceWorktree, true
	case "staged", "2":
		return codePlanGitSourceStaged, true
	default:
		return "", false
	}
}

func detectCodePlanRepoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return filepath.Clean(cwd)
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return filepath.Clean(cwd)
	}
	return filepath.Clean(root)
}

func resolveCodePlanPath(repoRoot, value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return filepath.Clean(repoRoot)
	}
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw)
	}
	return filepath.Clean(filepath.Join(repoRoot, raw))
}

func collectChangedCodeFiles(repoRoot string, mode codePlanMode, gitSource codePlanGitSource, scope codePlanScopePolicy, cache *codePlanCache) ([]string, codePlanMode, string, error) {
	switch mode {
	case codePlanModeGit:
		files, err := collectChangedCodeFilesFromGit(repoRoot, scope, gitSource)
		return files, codePlanModeGit, "", err
	case codePlanModeCache:
		files, err := collectChangedCodeFilesFromCache(repoRoot, scope, cache)
		return files, codePlanModeCache, "", err
	default:
		files, err := collectChangedCodeFilesFromGit(repoRoot, scope, gitSource)
		if err == nil {
			return files, codePlanModeGit, "", nil
		}
		cacheFiles, cacheErr := collectChangedCodeFilesFromCache(repoRoot, scope, cache)
		if cacheErr == nil {
			return cacheFiles, codePlanModeCache, text(
				"git status failed; used cache fallback",
				"git status 실패로 cache fallback 사용",
			), nil
		}
		return nil, codePlanModeAuto, "", fmt.Errorf("auto mode failed (git: %v, cache: %v)", err, cacheErr)
	}
}

func collectChangedCodeFilesFromGit(repoRoot string, scope codePlanScopePolicy, gitSource codePlanGitSource) ([]string, error) {
	entries, err := collectGitStatusEntries(repoRoot, gitSource)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, entry := range entries {
		status := strings.TrimSpace(entry.Status)
		if status == "" || status == "!!" || codePlanIsDeletedStatus(entry.Status) {
			continue
		}
		rel := codePlanNormalizeRelPath(entry.Path)
		if rel == "" {
			continue
		}
		if !isCodePlanFileInScope(rel, scope) {
			continue
		}
		seen[rel] = struct{}{}
	}
	return codePlanSortedKeys(seen), nil
}

func collectGitStatusEntries(repoRoot string, gitSource codePlanGitSource) ([]codePlanGitStatusEntry, error) {
	source := gitSource
	if source == "" {
		source = codePlanGitSourceWorktree
	}

	var cmd *exec.Cmd
	switch source {
	case codePlanGitSourceStaged:
		cmd = exec.Command("git", "diff", "--cached", "--name-only", "-z", "--diff-filter=ACMR")
	default:
		cmd = exec.Command("git", "status", "--porcelain", "-z", "--untracked-files=all")
	}
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		cmdName := "git status"
		if source == codePlanGitSourceStaged {
			cmdName = "git diff --cached"
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if stderr == "" {
				stderr = strings.TrimSpace(exitErr.Error())
			}
			return nil, fmt.Errorf("%s failed: %s", cmdName, stderr)
		}
		return nil, fmt.Errorf("%s failed: %v", cmdName, err)
	}

	if source == codePlanGitSourceStaged {
		return parseGitNameOnlyZ(out), nil
	}
	return parseGitStatusPorcelainZ(out), nil
}

func parseGitNameOnlyZ(raw []byte) []codePlanGitStatusEntry {
	parts := bytes.Split(raw, []byte{0})
	out := make([]codePlanGitStatusEntry, 0, len(parts))
	for _, chunk := range parts {
		path := strings.TrimSpace(string(chunk))
		if path == "" {
			continue
		}
		out = append(out, codePlanGitStatusEntry{
			Status: "M ",
			Path:   path,
		})
	}
	return out
}

func collectChangedCodeFilesFromCache(repoRoot string, scope codePlanScopePolicy, cache *codePlanCache) ([]string, error) {
	if cache == nil || len(cache.FileSnapshot) == 0 {
		return nil, fmt.Errorf("cache snapshot is missing (run code-plan-check --mode git once)")
	}
	current, err := buildCodePlanFileSnapshot(repoRoot, scope)
	if err != nil {
		return nil, err
	}
	changed := make(map[string]struct{})
	for rel, snap := range current {
		prev, ok := cache.FileSnapshot[codePlanPathKey(rel)]
		if !ok || prev != snap {
			changed[rel] = struct{}{}
		}
	}
	return codePlanSortedKeys(changed), nil
}

func parseGitStatusPorcelainZ(raw []byte) []codePlanGitStatusEntry {
	parts := bytes.Split(raw, []byte{0})
	out := make([]codePlanGitStatusEntry, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		chunk := parts[i]
		if len(chunk) == 0 {
			continue
		}
		line := string(chunk)
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		path := ""
		if len(line) > 3 {
			path = line[3:]
		}
		isRenameCopy := strings.Contains(status, "R") || strings.Contains(status, "C")
		if isRenameCopy && i+1 < len(parts) && len(parts[i+1]) > 0 {
			i++
			path = string(parts[i])
		}
		out = append(out, codePlanGitStatusEntry{
			Status: status,
			Path:   path,
		})
	}
	return out
}

func codePlanIsDeletedStatus(status string) bool {
	value := status
	if len(value) < 2 {
		return false
	}
	return value[0] == 'D' || value[1] == 'D'
}

func loadCodePlanScopePolicy(path string) (codePlanScopePolicy, []string, error) {
	policy := codePlanScopePolicy{
		AllowedExt:      defaultCodePlanAllowedExt(),
		IncludePrefixes: []string{},
		ExcludePrefixes: []string{},
	}
	target := strings.TrimSpace(path)
	if target == "" {
		return policy, nil, nil
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return policy, nil, nil
		}
		return policy, nil, fmt.Errorf("cannot read scope policy: %w", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	warnings := parseCodePlanScopePolicyLines(lines, &policy)
	return policy, warnings, nil
}

func parseCodePlanScopePolicyLines(lines []string, policy *codePlanScopePolicy) []string {
	if policy == nil {
		return nil
	}
	if policy.AllowedExt == nil {
		policy.AllowedExt = defaultCodePlanAllowedExt()
	}
	warnings := make([]string, 0)
	for idx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(strings.ToLower(line), "ext:"):
			value := strings.TrimSpace(line[4:])
			if value == "" {
				warnings = append(warnings, fmt.Sprintf("scope policy line %d: empty ext entry", idx+1))
				continue
			}
			if !strings.HasPrefix(value, ".") {
				value = "." + value
			}
			policy.AllowedExt[strings.ToLower(value)] = true
		case strings.HasPrefix(strings.ToLower(line), "include:"):
			value := strings.TrimSpace(line[len("include:"):])
			value = codePlanNormalizeRelPath(value)
			if value == "" {
				warnings = append(warnings, fmt.Sprintf("scope policy line %d: empty include entry", idx+1))
				continue
			}
			policy.IncludePrefixes = append(policy.IncludePrefixes, value)
		case strings.HasPrefix(strings.ToLower(line), "exclude:"):
			value := strings.TrimSpace(line[len("exclude:"):])
			value = codePlanNormalizeRelPath(value)
			if value == "" {
				warnings = append(warnings, fmt.Sprintf("scope policy line %d: empty exclude entry", idx+1))
				continue
			}
			policy.ExcludePrefixes = append(policy.ExcludePrefixes, value)
		default:
			warnings = append(warnings, fmt.Sprintf("scope policy line %d: unknown token (%s)", idx+1, line))
		}
	}
	sort.Strings(policy.IncludePrefixes)
	sort.Strings(policy.ExcludePrefixes)
	return warnings
}

func loadCodePlanRelationPolicy(path string) (codePlanRelationPolicy, []string, bool, error) {
	policy := codePlanRelationPolicy{
		Rules: make([]codePlanRelationRule, 0),
	}
	target := strings.TrimSpace(path)
	if target == "" {
		return policy, nil, false, nil
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return policy, nil, false, nil
		}
		return policy, nil, false, fmt.Errorf("cannot read relation policy: %w", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	warnings := parseCodePlanRelationPolicyLines(lines, &policy)
	return policy, warnings, true, nil
}

func parseCodePlanRelationPolicyLines(lines []string, policy *codePlanRelationPolicy) []string {
	if policy == nil {
		return nil
	}
	if policy.Rules == nil {
		policy.Rules = make([]codePlanRelationRule, 0)
	}
	warnings := make([]string, 0)
	for idx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "<->", 2)
		if len(parts) != 2 {
			warnings = append(warnings, fmt.Sprintf("relation policy line %d: missing <-> separator", idx+1))
			continue
		}
		left, leftErr := parseCodePlanRelationEndpoint(strings.TrimSpace(parts[0]))
		if leftErr != nil {
			warnings = append(warnings, fmt.Sprintf("relation policy line %d: %v", idx+1, leftErr))
			continue
		}
		right, rightErr := parseCodePlanRelationEndpoint(strings.TrimSpace(parts[1]))
		if rightErr != nil {
			warnings = append(warnings, fmt.Sprintf("relation policy line %d: %v", idx+1, rightErr))
			continue
		}
		policy.Rules = append(policy.Rules, codePlanRelationRule{
			LineNo: idx + 1,
			Raw:    line,
			Left:   left,
			Right:  right,
		})
	}
	return warnings
}

func parseCodePlanRelationEndpoint(raw string) (codePlanRelationEndpoint, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return codePlanRelationEndpoint{}, fmt.Errorf("empty endpoint")
	}
	idx := strings.LastIndex(value, "#")
	if idx <= 0 || idx >= len(value)-1 {
		return codePlanRelationEndpoint{}, fmt.Errorf("endpoint must be <doc.agd>#<SECTION-ID> (got %s)", value)
	}
	docPath := codePlanNormalizeRelPath(value[:idx])
	sectionID := strings.ToUpper(strings.TrimSpace(value[idx+1:]))
	if docPath == "" || sectionID == "" {
		return codePlanRelationEndpoint{}, fmt.Errorf("endpoint must include both doc path and section id (got %s)", value)
	}
	if !strings.EqualFold(filepath.Ext(docPath), ".agd") {
		return codePlanRelationEndpoint{}, fmt.Errorf("endpoint doc path should end with .agd (got %s)", value)
	}
	return codePlanRelationEndpoint{
		Raw:       value,
		DocPath:   docPath,
		DocKey:    codePlanPathKey(docPath),
		SectionID: sectionID,
	}, nil
}

func defaultCodePlanAllowedExt() map[string]bool {
	base := []string{
		".go", ".js", ".jsx", ".ts", ".tsx",
		".py", ".java", ".kt", ".cs",
		".c", ".cpp", ".h", ".hpp",
		".rs", ".swift", ".rb", ".php",
		".sh", ".bash", ".ps1", ".psm1",
		".sql",
	}
	out := make(map[string]bool, len(base))
	for _, ext := range base {
		out[ext] = true
	}
	return out
}

func buildCodePlanFileSnapshot(repoRoot string, scope codePlanScopePolicy) (map[string]codePlanFileSnapshot, error) {
	snapshot := make(map[string]codePlanFileSnapshot)
	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filepath.Clean(path) == filepath.Clean(repoRoot) {
			return nil
		}
		rel, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		rel = codePlanNormalizeRelPath(rel)
		if rel == "" {
			return nil
		}
		if d.IsDir() {
			if shouldSkipCodePlanDir(rel, scope) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isCodePlanFileInScope(rel, scope) {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		snapshot[codePlanPathKey(rel)] = codePlanFileSnapshot{
			Size:    info.Size(),
			ModUnix: info.ModTime().UTC().Unix(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func shouldSkipCodePlanDir(rel string, scope codePlanScopePolicy) bool {
	if codePlanPathHasPrefix(rel, ".git") ||
		codePlanPathHasPrefix(rel, "00_agd") ||
		codePlanPathHasPrefix(rel, "agd") {
		return true
	}
	for _, excluded := range scope.ExcludePrefixes {
		if codePlanPathHasPrefix(rel, excluded) {
			return true
		}
	}
	return false
}

func isCodePlanFileInScope(rel string, scope codePlanScopePolicy) bool {
	clean := codePlanNormalizeRelPath(rel)
	if clean == "" {
		return false
	}
	if codePlanPathHasPrefix(clean, ".git") ||
		codePlanPathHasPrefix(clean, "00_agd") ||
		codePlanPathHasPrefix(clean, "agd") {
		return false
	}
	if strings.EqualFold(filepath.Ext(clean), ".agd") {
		return false
	}
	for _, excluded := range scope.ExcludePrefixes {
		if codePlanPathHasPrefix(clean, excluded) {
			return false
		}
	}
	for _, included := range scope.IncludePrefixes {
		if codePlanPathHasPrefix(clean, included) {
			return true
		}
	}
	ext := strings.ToLower(filepath.Ext(clean))
	return scope.AllowedExt[ext]
}

func collectCodePlanDocs(scanRoot, repoRoot string) ([]codePlanDocInfo, []codePlanDocIssue, error) {
	files, err := listAGDFiles(scanRoot)
	if err != nil {
		return nil, nil, err
	}
	docs := make([]codePlanDocInfo, 0, len(files))
	issues := make([]codePlanDocIssue, 0)

	for _, file := range files {
		absPath := filepath.Clean(file)
		if !filepath.IsAbs(absPath) {
			if absolute, absErr := filepath.Abs(file); absErr == nil {
				absPath = filepath.Clean(absolute)
			}
		}
		relPath := codePlanRelPathFromRoot(repoRoot, absPath)

		doc, parseErrs, loadErr := agd.LoadFile(file)
		if loadErr != nil {
			issues = append(issues, codePlanDocIssue{
				RelPath: relPath,
				Message: fmt.Sprintf("doc read skipped: %v", loadErr),
			})
			continue
		}
		if len(parseErrs) > 0 {
			issues = append(issues, codePlanDocIssue{
				RelPath: relPath,
				Message: "doc parse skipped",
			})
			continue
		}

		changeDigest, latestID, latestDate := summarizeCodePlanChanges(doc)
		docs = append(docs, codePlanDocInfo{
			AbsPath:          absPath,
			RelPath:          relPath,
			RefPaths:         collectDocReferencePaths(doc, absPath, repoRoot),
			SectionIDs:       collectDocSectionIDs(doc),
			HasTag:           hasCodePlanTag(doc),
			ChangeDigest:     changeDigest,
			LatestChangeID:   latestID,
			LatestChangeDate: latestDate,
		})
	}

	sort.Slice(docs, func(i, j int) bool {
		return strings.ToLower(docs[i].RelPath) < strings.ToLower(docs[j].RelPath)
	})
	return docs, issues, nil
}

func collectDocSectionIDs(doc *agd.Document) map[string]bool {
	sections := make(map[string]bool)
	if doc == nil {
		return sections
	}
	for _, section := range doc.Sections {
		if section == nil {
			continue
		}
		id := strings.ToUpper(strings.TrimSpace(section.ID))
		if id == "" {
			continue
		}
		sections[id] = true
	}
	return sections
}

func collectDocReferencePaths(doc *agd.Document, docAbsPath, repoRoot string) []string {
	if doc == nil {
		return nil
	}
	out := make([]string, 0, len(doc.Sections)+1)
	seen := make(map[string]struct{})
	appendRef := func(raw string) {
		normalized := resolveDocReferenceToRepoPath(raw, docAbsPath, repoRoot)
		key := codePlanPathKey(normalized)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}

	appendRef(strings.TrimSpace(doc.Meta["doc_base_path"]))
	for _, section := range doc.Sections {
		if section == nil {
			continue
		}
		pathValue := strings.TrimSpace(section.Path)
		if idx := strings.Index(pathValue, "#"); idx >= 0 {
			pathValue = strings.TrimSpace(pathValue[:idx])
		}
		appendRef(pathValue)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i]) < strings.ToLower(out[j])
	})
	return out
}

func resolveDocReferenceToRepoPath(raw, docAbsPath, repoRoot string) string {
	value := strings.TrimSpace(strings.Trim(raw, `"`))
	if value == "" {
		return ""
	}
	docDir := filepath.Dir(docAbsPath)

	var absCandidate string
	switch {
	case filepath.IsAbs(value):
		absCandidate = filepath.Clean(value)
	case strings.HasPrefix(value, "./"), strings.HasPrefix(value, `.\`),
		strings.HasPrefix(value, "../"), strings.HasPrefix(value, `..\`):
		absCandidate = filepath.Clean(filepath.Join(docDir, value))
	default:
		repoCandidate := filepath.Clean(filepath.Join(repoRoot, value))
		docCandidate := filepath.Clean(filepath.Join(docDir, value))
		if pathExists(repoCandidate) {
			absCandidate = repoCandidate
		} else if pathExists(docCandidate) {
			absCandidate = docCandidate
		} else {
			absCandidate = repoCandidate
		}
	}
	return codePlanRelPathFromRoot(repoRoot, absCandidate)
}

func hasCodePlanTag(doc *agd.Document) bool {
	if doc == nil {
		return false
	}
	for key, value := range doc.Meta {
		trimmedKey := strings.ToLower(strings.TrimSpace(key))
		if strings.HasPrefix(trimmedKey, "incident_") || strings.HasPrefix(trimmedKey, "maintenance_") {
			if strings.TrimSpace(value) != "" {
				return true
			}
		}
	}

	for _, section := range doc.Sections {
		if section == nil {
			continue
		}
		content := strings.ToUpper(section.Content)
		if (strings.Contains(content, "[INCIDENT-PROBLEM-TAG]") && strings.Contains(content, "[/INCIDENT-PROBLEM-TAG]")) ||
			(strings.Contains(content, "[MAINTENANCE-ROOT-TAG]") && strings.Contains(content, "[/MAINTENANCE-ROOT-TAG]")) {
			return true
		}
	}
	return false
}

func summarizeCodePlanChanges(doc *agd.Document) (digest, latestID, latestDate string) {
	if doc == nil {
		return "", "", ""
	}
	h := sha256.New()
	for _, change := range doc.Changes {
		if change == nil {
			continue
		}
		_, _ = h.Write([]byte(strings.TrimSpace(change.ID)))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(strings.TrimSpace(change.Target)))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(strings.TrimSpace(change.Date)))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(strings.TrimSpace(change.Author)))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(strings.TrimSpace(change.Reason)))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(strings.TrimSpace(change.Impact)))
		_, _ = h.Write([]byte{0})
	}
	digest = hex.EncodeToString(h.Sum(nil))

	var latestParsed time.Time
	hasParsed := false
	for i, change := range doc.Changes {
		if change == nil {
			continue
		}
		if i == len(doc.Changes)-1 {
			latestID = strings.TrimSpace(change.ID)
			latestDate = strings.TrimSpace(change.Date)
		}
		parsed, err := parseChangeDateFlexible(change.Date)
		if err != nil {
			continue
		}
		if !hasParsed || parsed.After(latestParsed) {
			hasParsed = true
			latestParsed = parsed
			latestID = strings.TrimSpace(change.ID)
			latestDate = strings.TrimSpace(change.Date)
		}
	}
	return digest, latestID, latestDate
}

func evaluateCodePlanCheck(changedFiles []string, docs []codePlanDocInfo, cache *codePlanCache) []codePlanFileResult {
	results := make([]codePlanFileResult, 0, len(changedFiles))
	today := time.Now().Format("2006-01-02")
	for _, changed := range changedFiles {
		item := codePlanFileResult{
			ChangedPath: changed,
			MatchedDocs: make([]codePlanDocResult, 0),
		}
		for _, doc := range docs {
			if !docOverlapsChangedPath(changed, doc.RefPaths) {
				continue
			}
			changeUpdated := isCodePlanChangeUpdated(doc, cache, today)
			passed := doc.HasTag || changeUpdated
			item.MatchedDocs = append(item.MatchedDocs, codePlanDocResult{
				DocPath:       doc.RelPath,
				DocKey:        codePlanPathKey(doc.RelPath),
				HasTag:        doc.HasTag,
				ChangeUpdated: changeUpdated,
				Passed:        passed,
			})
		}
		if len(item.MatchedDocs) == 0 {
			item.MissingDoc = true
		}
		sort.Slice(item.MatchedDocs, func(i, j int) bool {
			return strings.ToLower(item.MatchedDocs[i].DocPath) < strings.ToLower(item.MatchedDocs[j].DocPath)
		})
		results = append(results, item)
	}
	return results
}

func evaluateCodePlanRelations(results []codePlanFileResult, docs []codePlanDocInfo, policy codePlanRelationPolicy, cache *codePlanCache) []codePlanRelationResult {
	if len(policy.Rules) == 0 || len(results) == 0 {
		return []codePlanRelationResult{}
	}
	today := time.Now().Format("2006-01-02")
	docsByKey := make(map[string]codePlanDocInfo, len(docs))
	for _, doc := range docs {
		key := codePlanPathKey(doc.RelPath)
		if key == "" {
			continue
		}
		docsByKey[key] = doc
	}

	seen := make(map[string]codePlanRelationResult)
	addResult := func(item codePlanRelationResult) {
		key := strings.Join([]string{
			codePlanPathKey(item.ChangedPath),
			codePlanPathKey(item.TriggerDocPath),
			strings.ToUpper(strings.TrimSpace(item.TriggerSection)),
			codePlanPathKey(item.RequiredDoc),
			strings.ToUpper(strings.TrimSpace(item.RequiredSec)),
		}, "|")
		prev, ok := seen[key]
		if !ok {
			seen[key] = item
			return
		}
		if prev.Passed && !item.Passed {
			seen[key] = item
		}
	}

	for _, changedItem := range results {
		for _, matchedDoc := range changedItem.MatchedDocs {
			triggerDoc, ok := docsByKey[matchedDoc.DocKey]
			if !ok {
				continue
			}
			for _, rule := range policy.Rules {
				switch matchedDoc.DocKey {
				case rule.Left.DocKey:
					addResult(evaluateCodePlanRelationPair(changedItem.ChangedPath, triggerDoc, rule.Left, rule.Right, docsByKey, cache, today))
				case rule.Right.DocKey:
					addResult(evaluateCodePlanRelationPair(changedItem.ChangedPath, triggerDoc, rule.Right, rule.Left, docsByKey, cache, today))
				}
			}
		}
	}

	out := make([]codePlanRelationResult, 0, len(seen))
	for _, item := range seen {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if strings.ToLower(out[i].ChangedPath) != strings.ToLower(out[j].ChangedPath) {
			return strings.ToLower(out[i].ChangedPath) < strings.ToLower(out[j].ChangedPath)
		}
		if strings.ToLower(out[i].TriggerDocPath) != strings.ToLower(out[j].TriggerDocPath) {
			return strings.ToLower(out[i].TriggerDocPath) < strings.ToLower(out[j].TriggerDocPath)
		}
		if strings.ToUpper(out[i].TriggerSection) != strings.ToUpper(out[j].TriggerSection) {
			return strings.ToUpper(out[i].TriggerSection) < strings.ToUpper(out[j].TriggerSection)
		}
		if strings.ToLower(out[i].RequiredDoc) != strings.ToLower(out[j].RequiredDoc) {
			return strings.ToLower(out[i].RequiredDoc) < strings.ToLower(out[j].RequiredDoc)
		}
		return strings.ToUpper(out[i].RequiredSec) < strings.ToUpper(out[j].RequiredSec)
	})
	return out
}

func evaluateCodePlanRelationPair(changedPath string, triggerDoc codePlanDocInfo, triggerEndpoint, requiredEndpoint codePlanRelationEndpoint, docsByKey map[string]codePlanDocInfo, cache *codePlanCache, today string) codePlanRelationResult {
	base := codePlanRelationResult{
		ChangedPath:    changedPath,
		TriggerDocPath: triggerDoc.RelPath,
		TriggerSection: triggerEndpoint.SectionID,
		RequiredDoc:    requiredEndpoint.DocPath,
		RequiredSec:    requiredEndpoint.SectionID,
		ChangeUpdated:  false,
		Passed:         false,
	}

	if !docHasSectionID(triggerDoc, triggerEndpoint.SectionID) {
		base.Message = "trigger endpoint section not found in matched doc"
		return base
	}

	requiredDoc, ok := docsByKey[requiredEndpoint.DocKey]
	if !ok {
		base.Message = "required endpoint doc not found"
		return base
	}
	base.RequiredDoc = requiredDoc.RelPath

	if !docHasSectionID(requiredDoc, requiredEndpoint.SectionID) {
		base.Message = "required endpoint section not found"
		return base
	}

	base.ChangeUpdated = isCodePlanChangeUpdated(requiredDoc, cache, today)
	base.Passed = base.ChangeUpdated
	if !base.Passed {
		base.Message = "required linked doc has no @change update"
	}
	return base
}

func docHasSectionID(doc codePlanDocInfo, sectionID string) bool {
	needle := strings.ToUpper(strings.TrimSpace(sectionID))
	if needle == "" {
		return false
	}
	if len(doc.SectionIDs) == 0 {
		return false
	}
	return doc.SectionIDs[needle]
}

func docOverlapsChangedPath(changedPath string, refs []string) bool {
	for _, ref := range refs {
		if codePlanPathsOverlap(changedPath, ref) {
			return true
		}
	}
	return false
}

func codePlanPathsOverlap(changedPath, docPath string) bool {
	changed := codePlanPathKey(changedPath)
	ref := codePlanPathKey(docPath)
	if changed == "" || ref == "" {
		return false
	}
	if changed == ref {
		return true
	}
	return codePlanPathHasPrefix(changed, ref)
}

func isCodePlanChangeUpdated(doc codePlanDocInfo, cache *codePlanCache, today string) bool {
	key := codePlanPathKey(doc.RelPath)
	if cache == nil {
		return isCodePlanTodayChange(doc.LatestChangeDate, today)
	}
	prev, ok := cache.DocSnapshot[key]
	if !ok {
		return isCodePlanTodayChange(doc.LatestChangeDate, today)
	}
	return strings.TrimSpace(prev.ChangeDigest) != strings.TrimSpace(doc.ChangeDigest)
}

func isCodePlanTodayChange(rawDate, today string) bool {
	trimmed := strings.TrimSpace(rawDate)
	if trimmed == "" {
		return false
	}
	parsed, err := parseChangeDateFlexible(trimmed)
	if err != nil {
		return strings.EqualFold(trimmed, today)
	}
	return parsed.Format("2006-01-02") == today
}

func buildCodePlanDocSnapshot(docs []codePlanDocInfo) map[string]codePlanDocSnapshot {
	out := make(map[string]codePlanDocSnapshot, len(docs))
	for _, doc := range docs {
		key := codePlanPathKey(doc.RelPath)
		if key == "" {
			continue
		}
		out[key] = codePlanDocSnapshot{
			ChangeDigest:   strings.TrimSpace(doc.ChangeDigest),
			LatestChangeID: strings.TrimSpace(doc.LatestChangeID),
			LatestChangeAt: strings.TrimSpace(doc.LatestChangeDate),
		}
	}
	return out
}

func codePlanDisplayPath(repoRoot, path string) string {
	rel := codePlanRelPathFromRoot(repoRoot, path)
	if rel == "" {
		return codePlanNormalizeRelPath(path)
	}
	return rel
}

func codePlanRelPathFromRoot(repoRoot, path string) string {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "" {
		return ""
	}
	if !filepath.IsAbs(cleanPath) {
		return codePlanNormalizeRelPath(cleanPath)
	}
	rel, err := filepath.Rel(filepath.Clean(repoRoot), cleanPath)
	if err != nil {
		return codePlanNormalizeRelPath(cleanPath)
	}
	rel = filepath.Clean(rel)
	if rel == "." {
		return "."
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return codePlanNormalizeRelPath(cleanPath)
	}
	return codePlanNormalizeRelPath(rel)
}

func codePlanNormalizeRelPath(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return ""
	}
	clean = strings.Trim(clean, `"`)
	clean = strings.ReplaceAll(clean, `\`, "/")
	for strings.HasPrefix(clean, "./") {
		clean = strings.TrimPrefix(clean, "./")
	}
	clean = strings.TrimPrefix(clean, "/")
	clean = strings.TrimSpace(clean)
	if clean == "" || clean == "." {
		return ""
	}
	return clean
}

func codePlanPathKey(value string) string {
	return strings.ToLower(codePlanNormalizeRelPath(value))
}

func codePlanPathHasPrefix(pathValue, prefixValue string) bool {
	path := codePlanPathKey(pathValue)
	prefix := codePlanPathKey(prefixValue)
	if path == "" || prefix == "" {
		return false
	}
	prefix = strings.TrimSuffix(prefix, "/")
	if path == prefix {
		return true
	}
	return strings.HasPrefix(path, prefix+"/")
}

func codePlanSortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, codePlanNormalizeRelPath(key))
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i]) < strings.ToLower(out[j])
	})
	return out
}

func boolToken(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
