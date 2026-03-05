// kit_role_graph.go handles kit generation and source/derived role graph output.
// kit_role_graph.go는 kit 생성과 source/derived 역할 그래프 출력을 처리합니다.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"agd/internal/agd"
	"agd/internal/agdtemplates"
)

type kitDocSpec struct {
	Key            string
	DocType        string
	Subdir         string
	FileSuffix     string
	TitleEN        string
	TitleKO        string
	Role           string
	SourceKey      string
	SourceSections string
}

type roleGraphRecord struct {
	Path           string
	Rel            string
	Title          string
	Authority      string
	SourceDoc      string
	SourceSections string
	IncidentSource string
	IncidentSecID  string
	MaintSource    string
	MaintSecID     string
	LoadError      string
	ParseErrors    []string
}

type roleGraphEdge struct {
	SourcePath  string
	DerivedPath string
	Mapping     string
}

type unresolvedRoleLink struct {
	DerivedPath  string
	SourceDocRef string
	Reason       string
}

const kitProfileUsage = "starter-kit|new-project|maintenance|incident"

type roleGraphScope string

const (
	roleGraphScopeAll                        roleGraphScope = "all"
	roleGraphScopeMaintenanceOnly            roleGraphScope = "maintenance"
	roleGraphScopeIncidentOnly               roleGraphScope = "incident"
	roleGraphScopeMaintenanceIncidentOnly    roleGraphScope = "maintenance-incident"
	roleGraphScopeExcludeMaintenanceIncident roleGraphScope = "exclude-maintenance-incident"
)

func roleGraphUsageText() string {
	return text(
		"usage: agd role-graph [root] [--format text|mermaid] [--out file] [--include-archive] [--scope all|maintenance|incident|maintenance-incident|exclude-maintenance-incident]",
		"usage: agd role-graph [root] [--format text|mermaid] [--out file] [--include-archive] [--scope all|maintenance|incident|maintenance-incident|exclude-maintenance-incident]",
	)
}

func parseRoleGraphScope(raw string) (roleGraphScope, bool) {
	scope := strings.ToLower(strings.TrimSpace(raw))
	scope = strings.ReplaceAll(scope, "_", "-")
	scope = strings.ReplaceAll(scope, " ", "-")

	switch scope {
	case "", "all":
		return roleGraphScopeAll, true
	case "maintenance", "maint", "maintenance-only":
		return roleGraphScopeMaintenanceOnly, true
	case "incident", "error", "incident-only", "errfix":
		return roleGraphScopeIncidentOnly, true
	case "maintenance-incident", "maintenance+incident", "both":
		return roleGraphScopeMaintenanceIncidentOnly, true
	case "exclude-maintenance-incident", "exclude-maintenance+incident", "exclude-both":
		return roleGraphScopeExcludeMaintenanceIncident, true
	default:
		return "", false
	}
}

func commandKitEasy(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, text(
			fmt.Sprintf("usage: agd kit <%s> <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
			fmt.Sprintf("usage: agd kit <%s> <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
		))
		fmt.Fprintln(os.Stderr, text(
			fmt.Sprintf("   or: agd %s <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
			fmt.Sprintf("   or: agd %s <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
		))
		return 2
	}

	commandToken := strings.ToLower(strings.TrimSpace(args[0]))
	profileRaw := ""
	rest := args[1:]
	switch commandToken {
	case "starter-kit":
		profileRaw = "starter-kit"
	case "new-project":
		profileRaw = "new-project"
	case "maintenance":
		profileRaw = "maintenance"
	case "incident":
		profileRaw = "incident"
	case "kit":
		if len(rest) == 0 {
			fmt.Fprintln(os.Stderr, text(
				fmt.Sprintf("kit error: profile is required (%s)", kitProfileUsage),
				fmt.Sprintf("kit error: profile is required (%s)", kitProfileUsage),
			))
			return 2
		}
		profileRaw = rest[0]
		rest = rest[1:]
	default:
		profileRaw = commandToken
	}

	profile, ok := normalizeKitProfile(profileRaw)
	if !ok {
		fmt.Fprintf(os.Stderr, text(
			"kit error: unknown profile %q\n", "kit error: unknown profile %q\n",
		), profileRaw)
		return 2
	}

	owner := ""
	root := docsRootDir
	lang := activeLang()
	force := false
	showGraph := true
	featureTag := ""
	tagSourceDocRaw := ""
	tagSectionInput := ""
	projectRaw := ""

	for i := 0; i < len(rest); i++ {
		token := strings.TrimSpace(rest[i])
		switch strings.ToLower(token) {
		case "":
			continue
		case "--owner":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, text("kit error: --owner requires value", "kit error: --owner requires value"))
				return 2
			}
			owner = strings.TrimSpace(rest[i+1])
			i++
		case "--root":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, text("kit error: --root requires value", "kit error: --root requires value"))
				return 2
			}
			root = strings.TrimSpace(rest[i+1])
			i++
		case "--lang":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, text("kit error: --lang requires value", "kit error: --lang requires value"))
				return 2
			}
			langValue := strings.ToLower(strings.TrimSpace(rest[i+1]))
			if langValue != "ko" && langValue != "en" {
				fmt.Fprintln(os.Stderr, text("kit error: --lang must be ko or en", "kit error: --lang must be ko or en"))
				return 2
			}
			lang = langValue
			i++
		case "--feature-tag":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, text("kit error: --feature-tag requires value", "kit error: --feature-tag requires value"))
				return 2
			}
			featureTag = strings.TrimSpace(rest[i+1])
			i++
		case "--tag-source":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, text("kit error: --tag-source requires value", "kit error: --tag-source requires value"))
				return 2
			}
			tagSourceDocRaw = strings.TrimSpace(rest[i+1])
			i++
		case "--tag-section":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, text("kit error: --tag-section requires value", "kit error: --tag-section requires value"))
				return 2
			}
			tagSectionInput = strings.TrimSpace(rest[i+1])
			i++
		case "--force":
			force = true
		case "--no-graph":
			showGraph = false
		default:
			if strings.HasPrefix(token, "-") {
				fmt.Fprintln(os.Stderr, text(
					fmt.Sprintf("usage: agd kit <%s> <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
					fmt.Sprintf("usage: agd kit <%s> <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
				))
				fmt.Fprintln(os.Stderr, text(
					fmt.Sprintf("   or: agd %s <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
					fmt.Sprintf("   or: agd %s <project-key> [--owner name] [--root dir] [--lang ko|en] [--feature-tag TAG] [--tag-source source.agd --tag-section SEC-ID] [--force] [--no-graph]", kitProfileUsage),
				))
				return 2
			}
			if projectRaw != "" {
				fmt.Fprintln(os.Stderr, text("kit error: too many positional arguments", "kit error: too many positional arguments"))
				return 2
			}
			projectRaw = token
		}
	}

	incidentSourceDocPath := ""
	incidentSourceSectionID := ""

	if strings.TrimSpace(projectRaw) == "" {
		fmt.Fprintln(os.Stderr, text("kit error: <project-key> is required", "kit error: <project-key> is required"))
		return 2
	}
	projectKey := sanitizeProjectKey(projectRaw)
	if projectKey == "" {
		fmt.Fprintln(os.Stderr, text(
			"kit error: project-key is invalid after normalization",
			"kit error: project-key is invalid after normalization",
		))
		return 2
	}
	useRootTagging := profile == "incident" || profile == "maintenance"
	if useRootTagging {
		rawFeatureTag := strings.TrimSpace(featureTag)
		featureTag = ""
		if rawFeatureTag != "" {
			featureTag = normalizeIncidentFeatureTag(rawFeatureTag, "")
			if featureTag == "" {
				fmt.Fprintln(os.Stderr, text(
					"kit error: invalid --feature-tag format",
					"kit error: invalid --feature-tag format",
				))
				return 2
			}
		}
		if profile == "incident" && featureTag == "" && strings.TrimSpace(tagSourceDocRaw) == "" {
			fmt.Fprintln(os.Stderr, text(
				"kit error: incident requires --feature-tag (ex: FT-CHECKOUT)",
				"kit error: incident requires --feature-tag (ex: FT-CHECKOUT)",
			))
			return 2
		}
		sourceProvided := strings.TrimSpace(tagSourceDocRaw) != ""
		sectionProvided := strings.TrimSpace(tagSectionInput) != ""
		if sourceProvided != sectionProvided {
			fmt.Fprintln(os.Stderr, text(
				"kit error: --tag-source and --tag-section must be used together",
				"kit error: --tag-source and --tag-section must be used together",
			))
			return 2
		}
		if sourceProvided {
			resolvedSource, resolveErr := resolveAGDPathEasy(tagSourceDocRaw, true)
			if resolveErr != nil {
				fmt.Fprintf(os.Stderr, "%s\n", resolveErr.Error())
				return 1
			}
			sourceDoc, parseErrs, readErr := agd.LoadFile(resolvedSource)
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "kit error: source read failed: %v\n", readErr)
				return 1
			}
			if len(parseErrs) > 0 {
				fmt.Fprintf(os.Stderr, "kit error: source parse error: %s\n", strings.Join(parseErrs, "; "))
				return 1
			}
			sectionID, sectionErr := resolveSectionIDChoice(sourceDoc, tagSectionInput)
			if sectionErr != nil {
				fmt.Fprintf(os.Stderr, "kit error: %v\n", sectionErr)
				return 2
			}
			sectionTitle := ""
			if section := sourceDoc.SectionByID(sectionID); section != nil {
				sectionTitle = strings.TrimSpace(section.Title)
			}
			incidentSourceDocPath = resolvedSource
			incidentSourceSectionID = sectionID
			if featureTag == "" {
				featureTag = defaultIncidentFeatureTagFromSection(sectionID, sectionTitle)
			}
			if featureTag == "" {
				fmt.Fprintln(os.Stderr, text(
					"kit error: cannot infer feature tag from --tag-source/--tag-section; provide --feature-tag",
					"kit error: cannot infer feature tag from --tag-source/--tag-section; provide --feature-tag",
				))
				return 2
			}
		}
	} else if strings.TrimSpace(featureTag) != "" {
		fmt.Fprintln(os.Stderr, text(
			"kit warning: --feature-tag is only used for incident/maintenance (ignored)", "kit warning: --feature-tag is only used for incident/maintenance (ignored)",
		))
		featureTag = ""
		if strings.TrimSpace(tagSourceDocRaw) != "" || strings.TrimSpace(tagSectionInput) != "" {
			fmt.Fprintln(os.Stderr, text(
				"kit warning: --tag-source/--tag-section are only used for incident/maintenance (ignored)", "kit warning: --tag-source/--tag-section are only used for incident/maintenance (ignored)",
			))
		}
	}

	if err := ensureKitRoot(root, lang); err != nil {
		fmt.Fprintf(os.Stderr, "kit error: %v\n", err)
		return 1
	}

	specs := kitProfileSpecs(profile)
	if len(specs) == 0 {
		fmt.Fprintf(os.Stderr, text(
			"kit error: profile %s has no document specs\n", "kit error: profile %s has no document specs\n",
		), profile)
		return 1
	}

	keyToPath := make(map[string]string, len(specs))
	nowDate := time.Now().Format("2006-01-02")
	version := "0.1"

	for _, spec := range specs {
		templateName, ok := easyTypeToTemplate(spec.DocType + "-" + lang)
		if !ok {
			fmt.Fprintf(os.Stderr, text(
				"kit error: cannot resolve template for doc type %s\n", "kit error: cannot resolve template for doc type %s\n",
			), spec.DocType)
			return 1
		}

		titleTemplate := spec.TitleEN
		if lang == "ko" {
			titleTemplate = spec.TitleKO
		}
		title := strings.TrimSpace(fmt.Sprintf(titleTemplate, projectRaw))
		outPath := filepath.Join(
			root,
			kitProjectScopedSubdir(profile, spec.Subdir, projectKey),
			ensureAGDExt(projectKey+"_"+spec.FileSuffix),
		)

		if err := createDocFromTemplate(templateName, outPath, title, owner, version, nowDate, force); err != nil {
			fmt.Fprintf(os.Stderr, "kit error: %v\n", err)
			return 1
		}
		keyToPath[spec.Key] = filepath.Clean(outPath)
	}

	for _, spec := range specs {
		if spec.Role == "" {
			continue
		}
		docPath, ok := keyToPath[spec.Key]
		if !ok {
			fmt.Fprintf(os.Stderr, "kit error: internal key not found: %s\n", spec.Key)
			return 1
		}
		switch spec.Role {
		case "source":
			if err := applyRoleToDoc(docPath, "source", "", "", "ai-agent"); err != nil {
				fmt.Fprintf(os.Stderr, "kit error: %v\n", err)
				return 1
			}
		case "derived":
			sourcePath, ok := keyToPath[spec.SourceKey]
			if !ok {
				fmt.Fprintf(os.Stderr, text(
					"kit error: source key not found for %s -> %s\n", "kit error: source key not found for %s -> %s\n",
				), spec.Key, spec.SourceKey)
				return 1
			}
			if err := applyRoleToDoc(docPath, "derived", sourcePath, spec.SourceSections, "ai-agent"); err != nil {
				fmt.Fprintf(os.Stderr, "kit error: %v\n", err)
				return 1
			}
		default:
			continue
		}
	}
	if profile == "incident" {
		if err := applyIncidentFeatureTagRooting(keyToPath, featureTag, lang, "ai-agent", incidentSourceDocPath, incidentSourceSectionID); err != nil {
			fmt.Fprintf(os.Stderr, "kit error: %v\n", err)
			return 1
		}
	}
	if profile == "maintenance" && strings.TrimSpace(featureTag) != "" {
		if err := applyMaintenanceFeatureTagRooting(keyToPath, featureTag, lang, "ai-agent", incidentSourceDocPath, incidentSourceSectionID); err != nil {
			fmt.Fprintf(os.Stderr, "kit error: %v\n", err)
			return 1
		}
	}

	fmt.Printf(text(
		"Created kit profile=%s project=%s root=%s\n", "Created kit profile=%s project=%s root=%s\n",
	), profile, projectKey, filepath.Clean(root))
	for _, spec := range specs {
		path := keyToPath[spec.Key]
		rel := relativePathForReport(root, path)
		roleLabel := spec.Role
		if strings.TrimSpace(roleLabel) == "" {
			roleLabel = "shared"
		}
		fmt.Printf("  - %s [%s, %s]\n", rel, spec.DocType, roleLabel)
	}
	if profile == "incident" {
		traceChain := incidentTraceChainForKeyMap(featureTag, keyToPath)
		fmt.Printf(text("Incident feature root tag: %s\n", "Incident feature root tag: %s\n"), featureTag)
		fmt.Printf(text("Root trace chain: %s\n", "Root trace chain: %s\n"), traceChain)
	}
	if profile == "maintenance" && strings.TrimSpace(featureTag) != "" {
		fmt.Printf(text("Maintenance root tag: %s\n", "Maintenance root tag: %s\n"), featureTag)
		if strings.TrimSpace(incidentSourceDocPath) != "" && strings.TrimSpace(incidentSourceSectionID) != "" {
			fmt.Printf(text(
				"Maintenance source section: %s | %s\n", "Maintenance source section: %s | %s\n",
			), incidentSourceDocPath, incidentSourceSectionID)
		}
	}

	if showGraph {
		graphText, err := renderRoleGraph(root, false, "text", roleGraphScopeAll)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kit warning: cannot render role graph: %v\n", err)
			return 0
		}
		fmt.Println("")
		fmt.Println(text("Role graph after kit creation:", "Role graph after kit creation:"))
		fmt.Println(graphText)
	}
	return 0
}

func kitProjectScopedSubdir(profile, baseSubdir, projectKey string) string {
	base := strings.TrimSpace(baseSubdir)
	key := strings.TrimSpace(projectKey)
	if base == "" {
		return base
	}
	prof := strings.ToLower(strings.TrimSpace(profile))
	if prof == "starter-kit" || prof == "incident" {
		return base
	}
	if prof == "new-project" {
		if key == "" {
			return base
		}
		return filepath.Join(base, key)
	}
	if prof == "maintenance" {
		maintenanceBase := filepath.Join("30_shared", "maintenance")
		if filepath.Clean(base) == filepath.Clean(maintenanceBase) {
			return base
		}
		return base
	}
	if key == "" {
		return base
	}
	return filepath.Join(base, key)
}

func normalizeKitProfile(raw string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.ReplaceAll(key, "_", "-")
	key = strings.ReplaceAll(key, " ", "-")

	switch key {
	case "starter-kit":
		return "starter-kit", true
	case "new-project":
		return "new-project", true
	case "maintenance":
		return "maintenance", true
	case "incident":
		return "incident", true
	default:
		return "", false
	}
}

func sanitizeProjectKey(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var out []rune
	lastSep := false
	for _, r := range trimmed {
		lower := unicode.ToLower(r)
		if unicode.IsLetter(lower) || unicode.IsDigit(lower) {
			out = append(out, lower)
			lastSep = false
			continue
		}
		if r == '-' || r == '_' || r == ' ' || r == '/' || r == '\\' || r == '.' {
			if !lastSep {
				out = append(out, '_')
				lastSep = true
			}
		}
	}
	normalized := strings.Trim(string(out), "_-")
	return normalized
}

func normalizeIncidentFeatureTag(raw, _ string) string {
	tag := strings.TrimSpace(raw)
	if tag == "" {
		return ""
	}

	out := make([]rune, 0, len(tag))
	lastSep := false
	for _, r := range tag {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			out = append(out, unicode.ToUpper(r))
			lastSep = false
		case r == '-' || r == '_' || r == ' ' || r == '/' || r == '\\' || r == '.':
			if !lastSep {
				out = append(out, '-')
				lastSep = true
			}
		}
	}

	normalized := strings.Trim(string(out), "-_")
	if normalized == "" {
		return ""
	}
	return normalized
}

func incidentTraceChainForKeyMap(featureTag string, keyToPath map[string]string) string {
	tag := strings.TrimSpace(featureTag)
	if tag == "" {
		return ""
	}
	if _, ok := keyToPath["incident_case"]; ok {
		return fmt.Sprintf("%s -> INC-001 -> INC-010 -> INC-020 -> INC-030 -> INC-040 -> INC-050 -> INC-060", tag)
	}
	return fmt.Sprintf("%s -> SYS-020 -> RUN-020 -> QA-020 -> PM-020", tag)
}

type incidentTraceSpec struct {
	SectionID string
	MarkerEN  string
	MarkerKO  string
	BlockEN   string
	BlockKO   string
}

func incidentTraceSpecForKey(key, featureTag string) (incidentTraceSpec, bool) {
	traceChain := fmt.Sprintf("%s -> SYS-020 -> RUN-020 -> QA-020 -> PM-020", featureTag)
	switch strings.TrimSpace(key) {
	case "incident_case":
		singleChain := fmt.Sprintf("%s -> INC-001 -> INC-010 -> INC-020 -> INC-030 -> INC-040 -> INC-050 -> INC-060", featureTag)
		return incidentTraceSpec{
			SectionID: "INC-001",
			MarkerEN:  "[Incident Root Flow by Feature Tag]",
			MarkerKO:  "[기능 태그 기반 장애 루트 흐름]",
			BlockEN: fmt.Sprintf(`[Incident Root Flow by Feature Tag]
- feature_tag: %s
- root_trace_chain: %s
- execution_flow:
  1) tag affected section in INC-001
  2) write bug situation in INC-010
  3) draft quick RCA in INC-020
  4) choose fix direction in INC-030
  5) handoff concrete AI task in INC-040
  6) summarize AI output in INC-050
  7) close or continue in INC-060`,
				featureTag, singleChain),
			BlockKO: fmt.Sprintf(`[기능 태그 기반 장애 루트 흐름]
- feature_tag: %s
- root_trace_chain: %s
- execution_flow:
  1) INC-001에서 영향 섹션 태깅
  2) INC-010에 버그 상황 기록
  3) INC-020에 빠른 RCA 초안 작성
  4) INC-030에서 수정 방향 결정
  5) INC-040에 AI 작업 요청 구체화
  6) INC-050에 AI 결과 요약
  7) INC-060에서 종료/추가 대응 결정`,
				featureTag, singleChain),
		}, true
	case "ai_guide":
		return incidentTraceSpec{
			SectionID: "AIG-020",
			MarkerEN:  "[Incident Prompt Contract by Feature Tag]",
			MarkerKO:  "[기능 태그 기반 장애 프롬프트 계약]",
			BlockEN: fmt.Sprintf(`[Incident Prompt Contract by Feature Tag]
- required_input.feature_tag: %s
- required_output.root_trace: %s
- if feature_tag is missing, ask clarification before editing docs.`,
				featureTag, traceChain),
			BlockKO: fmt.Sprintf(`[기능 태그 기반 장애 프롬프트 계약]
- required_input.feature_tag: %s
- required_output.root_trace: %s
- feature_tag가 없으면 문서 수정 전에 질문으로 확인한다.`,
				featureTag, traceChain),
		}, true
	case "service":
		return incidentTraceSpec{
			SectionID: "SYS-020",
			MarkerEN:  "[Incident Feature Root Trace]",
			MarkerKO:  "[장애 기능 루트 추적]",
			BlockEN: fmt.Sprintf(`[Incident Feature Root Trace]
- feature_tag: %s
- trace_root_chain: %s
- classify this incident under the tagged feature before deep dive
- root-down steps:
  1) confirm user impact and traffic slice for %s
  2) trace state/event path and dependency chain for %s
  3) connect mitigation and validation artifacts with the same tag`,
				featureTag, traceChain, featureTag, featureTag),
			BlockKO: fmt.Sprintf(`[장애 기능 루트 추적]
- feature_tag: %s
- trace_root_chain: %s
- 상세 분석 전에 장애를 해당 기능 태그 하위로 분류한다
- root-down steps:
  1) %s 기능의 사용자 영향/트래픽 구간 확인
  2) %s 기능의 상태/이벤트 경로와 의존 체인 추적
  3) 완화/검증 산출물을 동일 태그로 연결`,
				featureTag, traceChain, featureTag, featureTag),
		}, true
	case "runbook":
		return incidentTraceSpec{
			SectionID: "RUN-020",
			MarkerEN:  "[Feature-tag Immediate Response Root]",
			MarkerKO:  "[기능 태그 즉시 대응 루트]",
			BlockEN: fmt.Sprintf(`[Feature-tag Immediate Response Root]
- feature_tag: %s
- response_path: %s -> RUN-020 -> RUN-030 -> RUN-040
- first-10-minute actions must be aligned by this feature tag`,
				featureTag, featureTag),
			BlockKO: fmt.Sprintf(`[기능 태그 즉시 대응 루트]
- feature_tag: %s
- response_path: %s -> RUN-020 -> RUN-030 -> RUN-040
- 최초 10분 대응은 이 기능 태그 기준으로 정렬한다`,
				featureTag, featureTag),
		}, true
	case "qa":
		return incidentTraceSpec{
			SectionID: "QA-020",
			MarkerEN:  "[Feature-tag Reproduction Root]",
			MarkerKO:  "[기능 태그 재현 루트]",
			BlockEN: fmt.Sprintf(`[Feature-tag Reproduction Root]
- feature_tag: %s
- repro_path: %s -> QA-020 -> QA-040
- record reproduction and validation evidence by feature tag`,
				featureTag, featureTag),
			BlockKO: fmt.Sprintf(`[기능 태그 재현 루트]
- feature_tag: %s
- repro_path: %s -> QA-020 -> QA-040
- 재현/검증 증거를 기능 태그 기준으로 기록한다`,
				featureTag, featureTag),
		}, true
	case "postmortem":
		return incidentTraceSpec{
			SectionID: "PM-020",
			MarkerEN:  "[Feature-tag Root Cause Tree]",
			MarkerKO:  "[기능 태그 원인 트리]",
			BlockEN: fmt.Sprintf(`[Feature-tag Root Cause Tree]
- feature_tag: %s
- root_tree: %s -> PM-020 -> PM-040 -> PM-050
- connect causes, actions, and source feedback with the same feature tag`,
				featureTag, featureTag),
			BlockKO: fmt.Sprintf(`[기능 태그 원인 트리]
- feature_tag: %s
- root_tree: %s -> PM-020 -> PM-040 -> PM-050
- 원인, 조치, 원본 문서 피드백을 동일 기능 태그로 연결한다`,
				featureTag, featureTag),
		}, true
	case "meeting":
		return incidentTraceSpec{
			SectionID: "MTG-001",
			MarkerEN:  "[Feature-tag Incident Review Agenda]",
			MarkerKO:  "[기능 태그 장애 리뷰 아젠다]",
			BlockEN: fmt.Sprintf(`[Feature-tag Incident Review Agenda]
- focus_feature_tag: %s
- keep this meeting focused on trace, mitigation, and prevention for the tagged feature`,
				featureTag),
			BlockKO: fmt.Sprintf(`[기능 태그 장애 리뷰 아젠다]
- focus_feature_tag: %s
- 이 회의는 해당 기능의 추적, 완화, 재발 방지에 집중한다`,
				featureTag),
		}, true
	case "policy":
		return incidentTraceSpec{
			SectionID: "POL-030",
			MarkerEN:  "[Feature-tag Exception Control]",
			MarkerKO:  "[기능 태그 예외 통제]",
			BlockEN: fmt.Sprintf(`[Feature-tag Exception Control]
- required_feature_tag: %s
- hotfix/exception approvals must include this tag and trace links`,
				featureTag),
			BlockKO: fmt.Sprintf(`[기능 태그 예외 통제]
- required_feature_tag: %s
- hotfix/예외 승인에는 이 태그와 추적 링크를 반드시 포함한다`,
				featureTag),
		}, true
	default:
		return incidentTraceSpec{}, false
	}
}

func upsertMetaValue(meta map[string]string, key, value string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "" {
		return false
	}
	v := strings.TrimSpace(value)
	if current, ok := meta[k]; ok && strings.TrimSpace(current) == v {
		return false
	}
	meta[k] = v
	return true
}

func appendSectionBlockIfMissing(content, marker, block string) (string, bool) {
	m := strings.TrimSpace(marker)
	if m == "" {
		return content, false
	}
	if strings.Contains(content, m) {
		return content, false
	}
	b := strings.TrimSpace(block)
	if b == "" {
		return content, false
	}
	if strings.TrimSpace(content) == "" {
		return b, true
	}
	return strings.TrimRight(content, "\n") + "\n\n" + b, true
}

func applyIncidentFeatureTagRooting(
	keyToPath map[string]string,
	featureTag, lang, author string,
	sourceDocPath, sourceSectionID string,
) error {
	tag := strings.TrimSpace(featureTag)
	if tag == "" {
		return fmt.Errorf("incident feature tag is empty")
	}
	if strings.TrimSpace(author) == "" {
		author = "ai-agent"
	}
	isKorean := strings.EqualFold(strings.TrimSpace(lang), "ko")
	traceChain := incidentTraceChainForKeyMap(tag, keyToPath)
	keys := []string{"ai_guide", "service", "runbook", "qa", "postmortem", "meeting", "policy"}
	if _, ok := keyToPath["incident_case"]; ok {
		keys = append([]string{"incident_case"}, keys...)
	}

	for _, key := range keys {
		path, ok := keyToPath[key]
		if !ok {
			continue
		}

		doc, parseErrs, err := agd.LoadFile(path)
		if err != nil {
			return fmt.Errorf("incident feature-root: read failed for %s: %w", path, err)
		}
		if len(parseErrs) > 0 {
			return fmt.Errorf("incident feature-root: parse error for %s: %s", path, strings.Join(parseErrs, "; "))
		}

		changed := false
		changed = upsertMetaValue(doc.Meta, "incident_mode", "feature-root-trace") || changed
		changed = upsertMetaValue(doc.Meta, "incident_feature_tag", tag) || changed
		changed = upsertMetaValue(doc.Meta, "incident_trace_root", traceChain) || changed
		if key == "incident_case" {
			sourceDoc := strings.TrimSpace(sourceDocPath)
			sectionID := strings.TrimSpace(sourceSectionID)
			if sourceDoc != "" && sectionID != "" {
				resolvedSource := filepath.Clean(sourceDoc)
				if !filepath.IsAbs(resolvedSource) {
					if absSource, absErr := filepath.Abs(resolvedSource); absErr == nil {
						resolvedSource = filepath.Clean(absSource)
					}
				}
				storedSource := filepath.Clean(sourceDoc)
				if rel, relErr := filepath.Rel(filepath.Dir(path), resolvedSource); relErr == nil {
					storedSource = filepath.Clean(rel)
				}
				changed = upsertMetaValue(doc.Meta, "incident_source_doc", storedSource) || changed
				changed = upsertMetaValue(doc.Meta, "incident_source_section", sectionID) || changed
				if section := doc.SectionByID("INC-001"); section != nil {
					block := incidentProblemTagBlock(tag, storedSource, sectionID, "")
					nextContent, updated := upsertDelimitedContentBlock(
						section.Content,
						"[INCIDENT-PROBLEM-TAG]",
						"[/INCIDENT-PROBLEM-TAG]",
						block,
					)
					if updated {
						section.Content = nextContent
						changed = true
					}
				}
			}
		}

		targetSectionID := ""
		spec, hasSpec := incidentTraceSpecForKey(key, tag)
		if hasSpec {
			targetSectionID = spec.SectionID
			section := doc.SectionByID(spec.SectionID)
			if section != nil {
				marker := spec.MarkerEN
				block := spec.BlockEN
				if isKorean {
					marker = spec.MarkerKO
					block = spec.BlockKO
				}
				nextContent, updated := appendSectionBlockIfMissing(section.Content, marker, block)
				if updated {
					section.Content = nextContent
					changed = true
				}
			}
		}
		if !changed {
			continue
		}

		if strings.TrimSpace(targetSectionID) == "" || doc.SectionByID(targetSectionID) == nil {
			if len(doc.Sections) == 0 {
				return fmt.Errorf("incident feature-root: document has no section target: %s", path)
			}
			targetSectionID = strings.TrimSpace(doc.Sections[0].ID)
		}

		reason := fmt.Sprintf("incident feature-root trace initialized (%s)", tag)
		impact := fmt.Sprintf("incident triage now starts from feature tag %s and follows rooted trace chain", tag)
		if isKorean {
			reason = fmt.Sprintf("장애 기능 태그 루트 추적 체인을 초기화했습니다 (%s)", tag)
			impact = fmt.Sprintf("장애 대응이 기능 태그 %s를 기준으로 루트 체인에 맞춰 진행됩니다", tag)
		}
		doc.Changes = append(doc.Changes, &agd.Change{
			ID:     nextChangeID(doc),
			Target: targetSectionID,
			Date:   time.Now().Format("2006-01-02"),
			Author: strings.TrimSpace(author),
			Reason: reason,
			Impact: impact,
		})
		ensureDocPathMetadata(doc, path)

		validationErrors := agd.Validate(doc)
		if len(validationErrors) > 0 {
			return fmt.Errorf("incident feature-root: lint error for %s: %s", path, strings.Join(validationErrors, "; "))
		}
		if err := os.WriteFile(path, []byte(agd.Serialize(doc)), 0o644); err != nil {
			return fmt.Errorf("incident feature-root: write failed for %s: %w", path, err)
		}
	}

	return nil
}

func applyMaintenanceFeatureTagRooting(
	keyToPath map[string]string,
	featureTag, lang, author string,
	sourceDocPath, sourceSectionID string,
) error {
	tag := strings.TrimSpace(featureTag)
	if tag == "" {
		return nil
	}
	path, ok := keyToPath["maintenance_case"]
	if !ok {
		return fmt.Errorf("maintenance feature-root: maintenance_case document is missing")
	}
	if strings.TrimSpace(author) == "" {
		author = "ai-agent"
	}
	isKorean := strings.EqualFold(strings.TrimSpace(lang), "ko")

	doc, parseErrs, err := agd.LoadFile(path)
	if err != nil {
		return fmt.Errorf("maintenance feature-root: read failed for %s: %w", path, err)
	}
	if len(parseErrs) > 0 {
		return fmt.Errorf("maintenance feature-root: parse error for %s: %s", path, strings.Join(parseErrs, "; "))
	}
	if doc.Meta == nil {
		doc.Meta = map[string]string{}
	}

	changed := false
	changed = upsertMetaValue(doc.Meta, "maintenance_mode", "single-file-rooted") || changed
	changed = upsertMetaValue(doc.Meta, "maintenance_feature_tag", tag) || changed

	storedSource := ""
	sectionID := strings.TrimSpace(sourceSectionID)
	if strings.TrimSpace(sourceDocPath) != "" && sectionID != "" {
		resolvedSource := filepath.Clean(sourceDocPath)
		if !filepath.IsAbs(resolvedSource) {
			if absSource, absErr := filepath.Abs(resolvedSource); absErr == nil {
				resolvedSource = filepath.Clean(absSource)
			}
		}
		storedSource = filepath.Clean(sourceDocPath)
		if rel, relErr := filepath.Rel(filepath.Dir(path), resolvedSource); relErr == nil {
			storedSource = filepath.Clean(rel)
		}
		changed = upsertMetaValue(doc.Meta, "maintenance_source_doc", storedSource) || changed
		changed = upsertMetaValue(doc.Meta, "maintenance_source_section", sectionID) || changed
	}

	targetSectionID := ""
	targetSection := doc.SectionByID("MNT-001")
	if targetSection == nil && len(doc.Sections) > 0 {
		targetSection = doc.Sections[0]
	}
	if targetSection != nil {
		targetSectionID = strings.TrimSpace(targetSection.ID)
		block := maintenanceRootTagBlock(tag, storedSource, sectionID)
		nextContent, updated := upsertDelimitedContentBlock(
			targetSection.Content,
			"[MAINTENANCE-ROOT-TAG]",
			"[/MAINTENANCE-ROOT-TAG]",
			block,
		)
		if updated {
			targetSection.Content = nextContent
			changed = true
		}
	}

	if !changed {
		return nil
	}
	if strings.TrimSpace(targetSectionID) == "" {
		if len(doc.Sections) == 0 {
			return fmt.Errorf("maintenance feature-root: document has no section target: %s", path)
		}
		targetSectionID = strings.TrimSpace(doc.Sections[0].ID)
	}

	reason := fmt.Sprintf("maintenance root section tagged (%s)", tag)
	impact := fmt.Sprintf("maintenance flow is anchored on feature tag %s", tag)
	if isKorean {
		reason = fmt.Sprintf("유지보수 루트 섹션 태깅을 기록했습니다 (%s)", tag)
		impact = fmt.Sprintf("유지보수 흐름이 기능 태그 %s 기준으로 고정되었습니다", tag)
	}
	doc.Changes = append(doc.Changes, &agd.Change{
		ID:     nextChangeID(doc),
		Target: targetSectionID,
		Date:   time.Now().Format("2006-01-02"),
		Author: strings.TrimSpace(author),
		Reason: reason,
		Impact: impact,
	})
	ensureDocPathMetadata(doc, path)

	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		return fmt.Errorf("maintenance feature-root: lint error for %s: %s", path, strings.Join(validationErrors, "; "))
	}
	if err := os.WriteFile(path, []byte(agd.Serialize(doc)), 0o644); err != nil {
		return fmt.Errorf("maintenance feature-root: write failed for %s: %w", path, err)
	}
	return nil
}

func ensureKitRoot(root, lang string) error {
	cleanRoot := filepath.Clean(root)
	if strings.EqualFold(cleanRoot, filepath.Clean(docsRootDir)) {
		return ensureDocsRoot()
	}
	if err := os.MkdirAll(cleanRoot, 0o755); err != nil {
		return err
	}
	for _, rel := range docsScaffoldFolders {
		if err := os.MkdirAll(filepath.Join(cleanRoot, rel), 0o755); err != nil {
			return err
		}
	}
	return writeScaffoldReadmeIfMissing(cleanRoot, lang)
}

func createDocFromTemplate(templateName, outPath, title, owner, version, date string, force bool) error {
	rendered, err := agdtemplates.Render(templateName, agdtemplates.RenderOptions{
		Title:   title,
		Owner:   owner,
		Version: version,
		Date:    date,
	})
	if err != nil {
		return fmt.Errorf("cannot render template %s: %w", templateName, err)
	}

	doc, parseErrs := agd.Parse(rendered)
	if len(parseErrs) > 0 {
		return fmt.Errorf("template parse error: %s", strings.Join(parseErrs, "; "))
	}
	ensureDocPathMetadata(doc, outPath)
	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		return fmt.Errorf("template lint error: %s", strings.Join(validationErrors, "; "))
	}

	target := filepath.Clean(strings.TrimSpace(outPath))
	if _, err := os.Stat(target); err == nil && !force {
		return fmt.Errorf("%s already exists (use --force)", target)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", filepath.Dir(target), err)
	}
	if err := os.WriteFile(target, []byte(agd.Serialize(doc)), 0o644); err != nil {
		return fmt.Errorf("write failed for %s: %w", target, err)
	}
	return nil
}

func applyRoleToDoc(filePath, role, sourcePath, sourceSections, author string) error {
	doc, parseErrs, err := agd.LoadFile(filePath)
	if err != nil {
		return fmt.Errorf("read failed for %s: %w", filePath, err)
	}
	if len(parseErrs) > 0 {
		return fmt.Errorf("parse error for %s: %s", filePath, strings.Join(parseErrs, "; "))
	}

	doc.Meta["authority"] = strings.TrimSpace(role)
	delete(doc.Meta, "source_doc")
	delete(doc.Meta, "source_sections")

	reason := ""
	impact := ""
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "source":
		reason = "document authority set to source by kit generator"
		impact = "document becomes authority source for downstream docs generated by this kit"
	case "derived":
		if strings.TrimSpace(sourcePath) == "" {
			return fmt.Errorf("derived role requires source path for %s", filePath)
		}
		storedSource := filepath.Clean(sourcePath)
		if rel, relErr := filepath.Rel(filepath.Dir(filePath), sourcePath); relErr == nil {
			storedSource = filepath.Clean(rel)
		}
		doc.Meta["source_doc"] = storedSource
		if strings.TrimSpace(sourceSections) != "" {
			doc.Meta["source_sections"] = strings.TrimSpace(sourceSections)
		}
		reason = fmt.Sprintf("document authority set to derived by kit generator (source=%s)", storedSource)
		impact = fmt.Sprintf("derived linkage established to source_doc=%s", storedSource)
	default:
		return fmt.Errorf("invalid role %q", role)
	}

	if len(doc.Sections) == 0 {
		return fmt.Errorf("document %s has no section to attach @change target", filePath)
	}
	doc.Changes = append(doc.Changes, &agd.Change{
		ID:     nextChangeID(doc),
		Target: strings.TrimSpace(doc.Sections[0].ID),
		Date:   time.Now().Format("2006-01-02"),
		Author: strings.TrimSpace(author),
		Reason: reason,
		Impact: impact,
	})
	ensureDocPathMetadata(doc, filePath)

	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		return fmt.Errorf("lint error for %s: %s", filePath, strings.Join(validationErrors, "; "))
	}
	crossDocErrors := lintAuthorityAndConflicts(filePath, doc)
	if len(crossDocErrors) > 0 {
		return fmt.Errorf("cross-doc lint error for %s: %s", filePath, strings.Join(crossDocErrors, "; "))
	}

	if err := os.WriteFile(filePath, []byte(agd.Serialize(doc)), 0o644); err != nil {
		return fmt.Errorf("write failed for %s: %w", filePath, err)
	}
	return nil
}

func kitProfileSpecs(profile string) []kitDocSpec {
	switch profile {
	case "starter-kit":
		return []kitDocSpec{
			{Key: "service", DocType: "service", Subdir: filepath.Join("10_source", "service"), FileSuffix: "service", TitleEN: "%s - Service Logic Source", TitleKO: "%s - 서비스 로직 Source", Role: "source"},
			{Key: "policy", DocType: "policy", Subdir: filepath.Join("10_source", "policy"), FileSuffix: "policy", TitleEN: "%s - Release Policy", TitleKO: "%s - Release Policy", Role: "source"},
			{Key: "roadmap", DocType: "roadmap", Subdir: filepath.Join("30_shared", "roadmap"), FileSuffix: "roadmap", TitleEN: "%s - Roadmap", TitleKO: "%s - Roadmap"},
		}
	case "new-project":
		return []kitDocSpec{
			{Key: "core_spec", DocType: "core-spec", Subdir: filepath.Join("10_source", "product"), FileSuffix: "core_spec", TitleEN: "%s - New Project Core Spec", TitleKO: "%s - New Project Core Spec", Role: "source"},
			{Key: "policy", DocType: "policy", Subdir: filepath.Join("10_source", "product"), FileSuffix: "policy", TitleEN: "%s - New Project Policy", TitleKO: "%s - New Project Policy", Role: "source"},
			{Key: "delivery_plan", DocType: "delivery-plan", Subdir: filepath.Join("10_source", "product"), FileSuffix: "delivery_plan", TitleEN: "%s - New Project Delivery Plan", TitleKO: "%s - New Project Delivery Plan", Role: "derived", SourceKey: "core_spec", SourceSections: "CORE-010->DEL-001,CORE-030->DEL-020"},
			{Key: "roadmap", DocType: "roadmap", Subdir: filepath.Join("10_source", "product"), FileSuffix: "roadmap", TitleEN: "%s - New Project Roadmap", TitleKO: "%s - New Project Roadmap"},
		}
	case "maintenance":
		return []kitDocSpec{
			{Key: "maintenance_case", DocType: "maintenance-case", Subdir: filepath.Join("30_shared", "maintenance"), FileSuffix: "maintenance_case", TitleEN: "%s - Maintenance Case", TitleKO: "%s - Maintenance Case", Role: "source"},
		}
	case "incident":
		return []kitDocSpec{
			{Key: "incident_case", DocType: "incident-case", Subdir: filepath.Join("30_shared", "errFix"), FileSuffix: "incident_case", TitleEN: "%s - Incident Case", TitleKO: "%s - Incident Case", Role: "source"},
		}
	default:
		return nil
	}
}
func commandRoleGraphEasy(args []string) int {
	root := docsRootDir
	format := "text"
	outPath := ""
	includeArchive := false
	scope := roleGraphScopeAll
	rootProvided := false

	for i := 0; i < len(args); i++ {
		token := strings.TrimSpace(args[i])
		switch strings.ToLower(token) {
		case "":
			continue
		case "--format":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, roleGraphUsageText())
				return 2
			}
			format = strings.ToLower(strings.TrimSpace(args[i+1]))
			i++
		case "--out":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, roleGraphUsageText())
				return 2
			}
			outPath = strings.TrimSpace(args[i+1])
			i++
		case "--include-archive":
			includeArchive = true
		case "--scope":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, roleGraphUsageText())
				return 2
			}
			parsed, ok := parseRoleGraphScope(args[i+1])
			if !ok {
				fmt.Fprintln(os.Stderr, text(
					"role-graph error: --scope must be one of all|maintenance|incident|maintenance-incident|exclude-maintenance-incident",
					"role-graph error: --scope must be one of all|maintenance|incident|maintenance-incident|exclude-maintenance-incident",
				))
				return 2
			}
			scope = parsed
			i++
		default:
			if strings.HasPrefix(token, "-") || rootProvided {
				fmt.Fprintln(os.Stderr, roleGraphUsageText())
				return 2
			}
			root = token
			rootProvided = true
		}
	}

	if format != "text" && format != "mermaid" {
		fmt.Fprintln(os.Stderr, text(
			"role-graph error: --format must be text or mermaid",
			"role-graph error: --format must be text or mermaid",
		))
		return 2
	}

	if strings.EqualFold(filepath.Clean(root), filepath.Clean(docsRootDir)) {
		if err := ensureDocsRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "role-graph error: %v\n", err)
			return 1
		}
	}

	output, err := renderRoleGraph(root, includeArchive, format, scope)
	if err != nil {
		fmt.Fprintf(os.Stderr, "role-graph error: %v\n", err)
		return 1
	}

	if strings.TrimSpace(outPath) == "" {
		fmt.Println(output)
		return 0
	}

	target := filepath.Clean(outPath)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "role-graph error: cannot create output directory: %v\n", err)
		return 1
	}
	if err := os.WriteFile(target, []byte(output), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "role-graph error: cannot write output: %v\n", err)
		return 1
	}
	fmt.Printf(text("Exported role graph -> %s\n", "Exported role graph -> %s\n"), target)
	return 0
}

func renderRoleGraph(root string, includeArchive bool, format string, scope roleGraphScope) (string, error) {
	records, edges, unresolved, unassigned, _, _, err := collectRoleGraph(root, includeArchive)
	if err != nil {
		return "", err
	}
	records, edges, unresolved, unassigned = applyRoleGraphScopeFilter(scope, records, edges, unresolved, unassigned)
	parseFailures, loadFailures := buildRoleGraphFailureSummaries(records)
	if format == "mermaid" {
		return formatRoleGraphMermaid(root, records, edges, unresolved, parseFailures, loadFailures), nil
	}
	return formatRoleGraphText(root, records, edges, unresolved, unassigned, parseFailures, loadFailures), nil
}

func roleGraphClassifyRel(rel string) (maintenance bool, incident bool) {
	norm := strings.ToLower(filepath.ToSlash(filepath.Clean(strings.TrimSpace(rel))))
	norm = strings.TrimPrefix(norm, "./")
	maintenance = norm == "30_shared/maintenance" || strings.HasPrefix(norm, "30_shared/maintenance/")
	incident = norm == "30_shared/errfix" || strings.HasPrefix(norm, "30_shared/errfix/")
	return maintenance, incident
}

func roleGraphIsFocusedScope(scope roleGraphScope) bool {
	switch scope {
	case roleGraphScopeMaintenanceOnly, roleGraphScopeIncidentOnly, roleGraphScopeMaintenanceIncidentOnly:
		return true
	default:
		return false
	}
}

func roleGraphMatchPrimary(scope roleGraphScope, rel string) bool {
	isMaintenance, isIncident := roleGraphClassifyRel(rel)
	switch scope {
	case roleGraphScopeAll:
		return true
	case roleGraphScopeMaintenanceOnly:
		return isMaintenance
	case roleGraphScopeIncidentOnly:
		return isIncident
	case roleGraphScopeMaintenanceIncidentOnly:
		return isMaintenance || isIncident
	case roleGraphScopeExcludeMaintenanceIncident:
		return !(isMaintenance || isIncident)
	default:
		return true
	}
}

func applyRoleGraphScopeFilter(
	scope roleGraphScope,
	records map[string]roleGraphRecord,
	edges []roleGraphEdge,
	unresolved []unresolvedRoleLink,
	unassigned []string,
) (
	map[string]roleGraphRecord,
	[]roleGraphEdge,
	[]unresolvedRoleLink,
	[]string,
) {
	if scope == roleGraphScopeAll {
		return records, edges, unresolved, unassigned
	}

	primarySet := make(map[string]struct{})
	for path, rec := range records {
		if roleGraphMatchPrimary(scope, rec.Rel) {
			primarySet[path] = struct{}{}
		}
	}

	includeSet := make(map[string]struct{}, len(primarySet))
	for path := range primarySet {
		includeSet[path] = struct{}{}
	}

	focused := roleGraphIsFocusedScope(scope)
	if focused {
		for _, edge := range edges {
			_, sourcePrimary := primarySet[edge.SourcePath]
			_, derivedPrimary := primarySet[edge.DerivedPath]
			if !sourcePrimary && !derivedPrimary {
				continue
			}
			includeSet[edge.SourcePath] = struct{}{}
			includeSet[edge.DerivedPath] = struct{}{}
		}
	}

	filteredRecords := make(map[string]roleGraphRecord, len(includeSet))
	for path := range includeSet {
		if rec, ok := records[path]; ok {
			filteredRecords[path] = rec
		}
	}

	filteredEdges := make([]roleGraphEdge, 0, len(edges))
	for _, edge := range edges {
		_, sourceIncluded := filteredRecords[edge.SourcePath]
		_, derivedIncluded := filteredRecords[edge.DerivedPath]
		if !sourceIncluded || !derivedIncluded {
			continue
		}
		if focused {
			_, sourcePrimary := primarySet[edge.SourcePath]
			_, derivedPrimary := primarySet[edge.DerivedPath]
			if !sourcePrimary && !derivedPrimary {
				continue
			}
		}
		filteredEdges = append(filteredEdges, edge)
	}

	filteredUnresolved := make([]unresolvedRoleLink, 0, len(unresolved))
	for _, item := range unresolved {
		if _, ok := filteredRecords[item.DerivedPath]; !ok {
			continue
		}
		if focused {
			if _, primary := primarySet[item.DerivedPath]; !primary {
				continue
			}
		}
		filteredUnresolved = append(filteredUnresolved, item)
	}

	filteredUnassigned := make([]string, 0, len(unassigned))
	for _, path := range unassigned {
		if _, ok := filteredRecords[path]; !ok {
			continue
		}
		if focused {
			if _, primary := primarySet[path]; !primary {
				continue
			}
		}
		filteredUnassigned = append(filteredUnassigned, path)
	}

	return filteredRecords, filteredEdges, filteredUnresolved, filteredUnassigned
}

func buildRoleGraphFailureSummaries(records map[string]roleGraphRecord) ([]string, []string) {
	paths := make([]string, 0, len(records))
	for path := range records {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		return strings.ToLower(paths[i]) < strings.ToLower(paths[j])
	})

	parseFailures := make([]string, 0)
	loadFailures := make([]string, 0)
	for _, path := range paths {
		rec := records[path]
		if rec.LoadError != "" {
			loadFailures = append(loadFailures, fmt.Sprintf("%s: %s", rec.Rel, rec.LoadError))
		}
		if len(rec.ParseErrors) > 0 {
			parseFailures = append(parseFailures, fmt.Sprintf("%s: %s", rec.Rel, strings.Join(rec.ParseErrors, "; ")))
		}
	}
	return parseFailures, loadFailures
}

func collectRoleGraph(root string, includeArchive bool) (
	map[string]roleGraphRecord,
	[]roleGraphEdge,
	[]unresolvedRoleLink,
	[]string,
	[]string,
	[]string,
	error,
) {
	files, err := listAGDFiles(root)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	records := make(map[string]roleGraphRecord, len(files))
	parseFailures := make([]string, 0)
	loadFailures := make([]string, 0)

	for _, file := range files {
		if !includeArchive && isUnderArchiveDir(root, file) {
			continue
		}
		rec := roleGraphRecord{
			Path:        filepath.Clean(file),
			Rel:         relativePathForReport(root, file),
			ParseErrors: make([]string, 0),
		}
		doc, parseErrs, loadErr := agd.LoadFile(file)
		if loadErr != nil {
			rec.LoadError = loadErr.Error()
			loadFailures = append(loadFailures, fmt.Sprintf("%s: %s", rec.Rel, rec.LoadError))
			records[rec.Path] = rec
			continue
		}
		if len(parseErrs) > 0 {
			rec.ParseErrors = append(rec.ParseErrors, parseErrs...)
			parseFailures = append(parseFailures, fmt.Sprintf("%s: %s", rec.Rel, strings.Join(parseErrs, "; ")))
			records[rec.Path] = rec
			continue
		}
		rec.Title = strings.TrimSpace(doc.Meta["title"])
		rec.Authority = strings.ToLower(strings.TrimSpace(doc.Meta["authority"]))
		rec.SourceDoc = strings.TrimSpace(doc.Meta["source_doc"])
		rec.SourceSections = strings.TrimSpace(doc.Meta["source_sections"])
		rec.IncidentSource = strings.TrimSpace(doc.Meta["incident_source_doc"])
		rec.IncidentSecID = strings.TrimSpace(doc.Meta["incident_source_section"])
		rec.MaintSource = strings.TrimSpace(doc.Meta["maintenance_source_doc"])
		rec.MaintSecID = strings.TrimSpace(doc.Meta["maintenance_source_section"])
		records[rec.Path] = rec
	}

	unresolved := make([]unresolvedRoleLink, 0)
	edges := make([]roleGraphEdge, 0)
	unassigned := make([]string, 0)
	sourceDocCache := make(map[string]*agd.Document)

	loadSourceDoc := func(path string) *agd.Document {
		clean := filepath.Clean(path)
		if cached, ok := sourceDocCache[clean]; ok {
			return cached
		}
		doc, parseErrs, err := agd.LoadFile(clean)
		if err != nil || len(parseErrs) > 0 {
			sourceDocCache[clean] = nil
			return nil
		}
		sourceDocCache[clean] = doc
		return doc
	}

	incidentMappingLabel := func(sourcePath, sectionID string) string {
		if sectionID == "" {
			return ""
		}
		sourceDoc := loadSourceDoc(sourcePath)
		if sourceDoc == nil {
			return sectionID
		}
		section := sourceDoc.SectionByID(sectionID)
		if section == nil {
			return sectionID
		}
		sectionTitle := strings.TrimSpace(section.Title)
		if sectionTitle == "" {
			return sectionID
		}
		return fmt.Sprintf("%s | %s", sectionID, sectionTitle)
	}

	for _, rec := range records {
		if rec.LoadError != "" || len(rec.ParseErrors) > 0 {
			continue
		}
		incidentLinked := strings.TrimSpace(rec.SourceDoc) == "" && strings.TrimSpace(rec.IncidentSource) != ""
		maintenanceLinked := strings.TrimSpace(rec.SourceDoc) == "" && strings.TrimSpace(rec.MaintSource) != ""
		linkedByTag := incidentLinked || maintenanceLinked
		isDerived := rec.Authority == "derived" || strings.TrimSpace(rec.SourceDoc) != "" || linkedByTag
		isSource := rec.Authority == "source" && !linkedByTag
		if !isSource && !isDerived {
			unassigned = append(unassigned, rec.Path)
			continue
		}
		if !isDerived {
			continue
		}
		sourceRef := strings.TrimSpace(rec.SourceDoc)
		if sourceRef == "" {
			sourceRef = strings.TrimSpace(rec.IncidentSource)
		}
		if sourceRef == "" {
			sourceRef = strings.TrimSpace(rec.MaintSource)
		}
		if sourceRef == "" {
			unresolved = append(unresolved, unresolvedRoleLink{
				DerivedPath:  rec.Path,
				SourceDocRef: "",
				Reason:       "source_doc is empty",
			})
			continue
		}
		resolved, resolveErr := resolveSourceDocPathForLint(rec.Path, sourceRef)
		if resolveErr != nil {
			unresolved = append(unresolved, unresolvedRoleLink{
				DerivedPath:  rec.Path,
				SourceDocRef: sourceRef,
				Reason:       resolveErr.Error(),
			})
			continue
		}
		sourcePath := filepath.Clean(resolved)
		if _, ok := records[sourcePath]; !ok {
			unresolved = append(unresolved, unresolvedRoleLink{
				DerivedPath:  rec.Path,
				SourceDocRef: sourceRef,
				Reason:       "source document is outside current root",
			})
			continue
		}
		mapping := strings.TrimSpace(rec.SourceSections)
		if incidentLinked {
			mapping = incidentMappingLabel(sourcePath, strings.TrimSpace(rec.IncidentSecID))
		} else if maintenanceLinked {
			mapping = incidentMappingLabel(sourcePath, strings.TrimSpace(rec.MaintSecID))
		}
		edges = append(edges, roleGraphEdge{
			SourcePath:  sourcePath,
			DerivedPath: rec.Path,
			Mapping:     mapping,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		left := edges[i].SourcePath + "->" + edges[i].DerivedPath
		right := edges[j].SourcePath + "->" + edges[j].DerivedPath
		return strings.ToLower(left) < strings.ToLower(right)
	})
	sort.Slice(unresolved, func(i, j int) bool {
		return strings.ToLower(unresolved[i].DerivedPath) < strings.ToLower(unresolved[j].DerivedPath)
	})
	sort.Slice(unassigned, func(i, j int) bool {
		return strings.ToLower(unassigned[i]) < strings.ToLower(unassigned[j])
	})

	return records, edges, unresolved, unassigned, parseFailures, loadFailures, nil
}

func formatRoleGraphText(
	root string,
	records map[string]roleGraphRecord,
	edges []roleGraphEdge,
	unresolved []unresolvedRoleLink,
	unassigned []string,
	parseFailures []string,
	loadFailures []string,
) string {
	sourcePaths := make([]string, 0)
	sourceSet := make(map[string]struct{})
	hasIncoming := make(map[string]bool)
	for _, edge := range edges {
		hasIncoming[edge.DerivedPath] = true
	}
	for path, rec := range records {
		if strings.EqualFold(rec.Authority, "source") && !hasIncoming[path] {
			sourceSet[path] = struct{}{}
		}
	}
	for _, edge := range edges {
		sourceSet[edge.SourcePath] = struct{}{}
	}
	for path := range sourceSet {
		sourcePaths = append(sourcePaths, path)
	}
	sort.Slice(sourcePaths, func(i, j int) bool {
		return strings.ToLower(sourcePaths[i]) < strings.ToLower(sourcePaths[j])
	})

	edgeBySource := make(map[string][]roleGraphEdge)
	for _, edge := range edges {
		edgeBySource[edge.SourcePath] = append(edgeBySource[edge.SourcePath], edge)
	}

	displayTitle := func(raw string) string {
		title := strings.TrimSpace(raw)
		if title == "" {
			return text("(untitled)", "(untitled)")
		}
		return title
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(text("Role graph root: %s\n", "Role graph root: %s\n"), filepath.Clean(root)))
	b.WriteString(fmt.Sprintf(text(
		"Summary: sources=%d, derived-links=%d, unresolved=%d, unassigned=%d\n", "Summary: sources=%d, derived-links=%d, unresolved=%d, unassigned=%d\n",
	),
		len(sourcePaths), len(edges), len(unresolved), len(unassigned)))

	if len(sourcePaths) == 0 {
		b.WriteString("\n")
		b.WriteString(text("Sources:\n", "Sources:\n"))
		b.WriteString(text("  (none)\n", "  (none)\n"))
	} else {
		b.WriteString("\n")
		b.WriteString(text("Sources:\n", "Sources:\n"))
		for sourceIdx, sourcePath := range sourcePaths {
			rec := records[sourcePath]
			rel := relativePathForReport(root, sourcePath)
			title := displayTitle(rec.Title)
			outgoing := edgeBySource[sourcePath]
			sort.Slice(outgoing, func(i, j int) bool {
				return strings.ToLower(outgoing[i].DerivedPath) < strings.ToLower(outgoing[j].DerivedPath)
			})

			b.WriteString(fmt.Sprintf("[%d] %s\n", sourceIdx+1, rel))
			b.WriteString(fmt.Sprintf("    %s %s\n", text("title:", "title:"), title))
			b.WriteString(fmt.Sprintf("    %s %d\n", text("derived docs:", "derived docs:"), len(outgoing)))

			if len(outgoing) == 0 {
				b.WriteString(fmt.Sprintf("    %s\n", text("links: (none)", "links: (none)")))
				b.WriteString("\n")
				continue
			}
			for linkIdx, edge := range outgoing {
				dstRel := relativePathForReport(root, edge.DerivedPath)
				dstTitle := displayTitle(records[edge.DerivedPath].Title)
				b.WriteString(fmt.Sprintf("    [%d] %s\n", linkIdx+1, dstRel))
				b.WriteString(fmt.Sprintf("         %s %s\n", text("title:", "title:"), dstTitle))

				mappingTokens := splitCSV(edge.Mapping)
				if len(mappingTokens) == 0 {
					b.WriteString(fmt.Sprintf("         %s\n", text("source_sections: (none)", "source_sections: (none)")))
					continue
				}
				b.WriteString(fmt.Sprintf(
					"         %s %d\n",
					text("source_sections count:", "source_sections count:"),
					len(mappingTokens),
				))
				for _, token := range mappingTokens {
					b.WriteString(fmt.Sprintf("           - %s\n", token))
				}
			}
			b.WriteString("\n")
		}
	}

	if len(unresolved) > 0 {
		b.WriteString("\n")
		b.WriteString(text("Unresolved derived docs:\n", "Unresolved derived docs:\n"))
		for _, item := range unresolved {
			rel := relativePathForReport(root, item.DerivedPath)
			ref := strings.TrimSpace(item.SourceDocRef)
			if ref == "" {
				ref = text("(empty)", "(empty)")
			}
			b.WriteString(fmt.Sprintf("- %s\n", rel))
			b.WriteString(fmt.Sprintf("    %s %s\n", text("source_doc:", "source_doc:"), ref))
			b.WriteString(fmt.Sprintf("    %s %s\n", text("reason:", "reason:"), item.Reason))
		}
	}

	if len(unassigned) > 0 {
		b.WriteString("\n")
		b.WriteString(text("Unassigned docs (no source/derived role):\n", "Unassigned docs (no source/derived role):\n"))
		for _, path := range unassigned {
			rel := relativePathForReport(root, path)
			title := displayTitle(records[path].Title)
			b.WriteString(fmt.Sprintf("- %s | %s\n", rel, title))
		}
	}

	if len(parseFailures) > 0 {
		b.WriteString("\n")
		b.WriteString(text("Parse failures:\n", "Parse failures:\n"))
		for _, item := range parseFailures {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	if len(loadFailures) > 0 {
		b.WriteString("\n")
		b.WriteString(text("Read failures:\n", "Read failures:\n"))
		for _, item := range loadFailures {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func formatRoleGraphMermaid(
	root string,
	records map[string]roleGraphRecord,
	edges []roleGraphEdge,
	unresolved []unresolvedRoleLink,
	parseFailures []string,
	loadFailures []string,
) string {
	nodeID := make(map[string]string)
	nodeOrder := make([]string, 0)

	addNode := func(path string) string {
		if id, ok := nodeID[path]; ok {
			return id
		}
		id := fmt.Sprintf("N%03d", len(nodeOrder)+1)
		nodeID[path] = id
		nodeOrder = append(nodeOrder, path)
		return id
	}

	for _, edge := range edges {
		addNode(edge.SourcePath)
		addNode(edge.DerivedPath)
	}

	var b strings.Builder
	b.WriteString("graph TD\n")
	b.WriteString(fmt.Sprintf("  %% root: %s\n", filepath.Clean(root)))

	sort.Slice(nodeOrder, func(i, j int) bool {
		return strings.ToLower(nodeOrder[i]) < strings.ToLower(nodeOrder[j])
	})
	for _, path := range nodeOrder {
		rec := records[path]
		labelRole := strings.TrimSpace(rec.Authority)
		if labelRole == "" {
			labelRole = "doc"
		}
		label := fmt.Sprintf("%s\\n(%s)", relativePathForReport(root, path), labelRole)
		id := nodeID[path]
		b.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", id, escapeMermaidLabel(label)))
	}

	for _, edge := range edges {
		sourceID := nodeID[edge.SourcePath]
		derivedID := nodeID[edge.DerivedPath]
		label := strings.TrimSpace(edge.Mapping)
		if label == "" {
			label = "source_doc"
		}
		b.WriteString(fmt.Sprintf("  %s -->|%s| %s\n", sourceID, escapeMermaidLabel(label), derivedID))
	}

	for _, item := range unresolved {
		b.WriteString(fmt.Sprintf(
			"  %% unresolved: %s -> %s (%s)\n",
			relativePathForReport(root, item.DerivedPath),
			item.SourceDocRef,
			item.Reason,
		))
	}
	for _, item := range parseFailures {
		b.WriteString(fmt.Sprintf("  %% parse-failure: %s\n", escapeMermaidLabel(item)))
	}
	for _, item := range loadFailures {
		b.WriteString(fmt.Sprintf("  %% read-failure: %s\n", escapeMermaidLabel(item)))
	}

	return strings.TrimRight(b.String(), "\n")
}

func escapeMermaidLabel(raw string) string {
	value := strings.ReplaceAll(raw, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
