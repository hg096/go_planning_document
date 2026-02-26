// wizard.go provides the interactive AGD wizard workflows.
// wizard.go는 대화형 AGD 위자드 흐름을 제공합니다.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"agd/internal/agd"
)

var errWizardBack = errors.New("wizard: back")

func commandWizard(_ []string) int {
	reader := bufio.NewReader(os.Stdin)
	selectedDoc := ""
	fmt.Println(text(
		"Guided flow is on for edit actions: reason+impact required -> map-sync -> check.", "Guided flow is on for edit actions: reason+impact required -> map-sync -> check.",
	))

	fmt.Println(text("AGD Wizard", "AGD Wizard"))
	fmt.Println(text("Type a number to choose an action.", "Type a number to choose an action."))

	for {
		fmt.Println("")
		fmt.Println(text("[1] Select document", "[1] Select document"))
		fmt.Println(text("[2] Generate doc kit (starter/maintenance/new/incident)", "[2] Generate doc kit (starter/maintenance/new/incident)"))
		fmt.Println(text("[3] Validate whole docs tree", "[3] Validate whole docs tree"))
		fmt.Println(text("[4] New document", "[4] New document"))
		fmt.Println(text("[5] Show source/derived relation graph", "[5] Show source/derived relation graph"))
		fmt.Println(text("[0] Exit", "[0] Exit"))

		choice, err := promptRequired(reader, text("Select", "Select"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
			return 1
		}

		switch strings.TrimSpace(choice) {
		case "1":
			code := wizardSelectDoc(reader, &selectedDoc)
			if code == 2 {
				continue
			}
			if code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
				continue
			}
			if code := wizardDocActionLoop(reader, &selectedDoc); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "2":
			if code := wizardKit(reader); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "3":
			if code := wizardCheckAll(reader); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "4":
			if code := wizardNew(reader); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "5":
			if code := wizardRoleGraph(reader); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "0", "q", "quit", "exit":
			fmt.Println(text("Bye.", "Bye."))
			return 0
		default:
			fmt.Println(text("Invalid choice. Use 0-5.", "Invalid choice. Use 0-5."))
		}
	}
}

func wizardDocActionLoop(reader *bufio.Reader, selectedDoc *string) int {
	for {
		filePath, code := wizardRequireSelectedDoc(selectedDoc)
		if code != 0 {
			return code
		}
		displayName := filepath.Base(filePath)
		if strings.TrimSpace(displayName) == "" || displayName == "." || displayName == string(os.PathSeparator) {
			displayName = filePath
		}
		fmt.Println("")
		fmt.Printf(text("Selected document: %s\n", "Selected document: %s\n"), displayName)
		fmt.Println(text("[1] Search keyword", "[1] Search keyword"))
		fmt.Println(text("[2] Add new section (also updates map)", "[2] Add new section (also updates map)"))
		fmt.Println(text("[3] Add logic-change record", "[3] Add logic-change record"))
		fmt.Println(text("[4] Sync map only", "[4] Sync map only"))
		fmt.Println(text("[5] Check document", "[5] Check document"))
		fmt.Println(text("[6] Export to markdown", "[6] Export to markdown"))
		fmt.Println(text("[7] Set source/derived role", "[7] Set source/derived role"))
		fmt.Println(text("[0] Back", "[0] 뒤로가기"))

		choice, err := promptRequired(reader, text("Select", "Select"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
			return 1
		}

		switch strings.TrimSpace(choice) {
		case "1":
			if code := wizardSearchOnDoc(reader, filePath); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "2":
			if code := wizardSectionAddOnDoc(reader, filePath); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "3":
			if code := wizardLogicLogOnDoc(reader, filePath); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "4":
			if code := wizardMapSyncOnDoc(reader, filePath); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "5":
			if code := wizardCheckOnDoc(filePath); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "6":
			if code := wizardViewOnDoc(reader, filePath); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "7":
			if code := wizardRoleSetOnDoc(reader, filePath); code != 0 {
				fmt.Printf("failed (code=%d)\n", code)
			}
		case "0":
			return 0
		default:
			fmt.Println(text("Invalid choice. Use 0-7.", "Invalid choice. Use 0-7."))
		}
	}
}

func wizardSelectDoc(reader *bufio.Reader, selectedDoc *string) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 2
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	*selectedDoc = filePath
	fmt.Printf(text("Selected document -> %s\n", "Selected document -> %s\n"), filePath)
	return 0
}

func wizardRequireSelectedDoc(selectedDoc *string) (string, int) {
	if strings.TrimSpace(*selectedDoc) == "" {
		fmt.Println(text(
			"No document selected. Choose [1] Select document first.", "No document selected. Choose [1] Select document first.",
		))
		return "", 2
	}
	return strings.TrimSpace(*selectedDoc), 0
}

func wizardDocTypeOptions() []docTypeOption {
	return allowedCreateDocTypeOptions()
}

func printDocTypeCatalogWithOptions(w io.Writer, options []docTypeOption, withNumber bool) {
	for i, opt := range options {
		if withNumber {
			fmt.Fprintln(w, text(
				fmt.Sprintf("  [%d] %-10s - %s", i+1, opt.Key, opt.EN),
				fmt.Sprintf("  [%d] %-10s - %s", i+1, opt.Key, opt.KO),
			))
			continue
		}
		fmt.Fprintln(w, text(
			fmt.Sprintf("  %-10s - %s", opt.Key, opt.EN),
			fmt.Sprintf("  %-10s - %s", opt.Key, opt.KO),
		))
	}
}

func wizardNew(reader *bufio.Reader) int {
	fmt.Println("")
	fmt.Println(text("Document type:", "문서 타입:"))
	options := wizardDocTypeOptions()
	printDocTypeCatalogWithOptions(os.Stdout, options, true)
	fmt.Println(text("[0] Back", "[0] 뒤로가기"))

	typeChoice, err := promptRequired(reader, text(
		fmt.Sprintf("Type (1-%d)", len(options)),
		fmt.Sprintf("타입 (1-%d)", len(options)),
	))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	normalized := strings.ToLower(strings.TrimSpace(typeChoice))
	if normalized == "0" {
		return 0
	}
	docType := normalized
	if idx, convErr := strconv.Atoi(normalized); convErr == nil {
		if idx < 1 || idx > len(options) {
			fmt.Println(text("Invalid type.", "Invalid type."))
			return 2
		}
		docType = options[idx-1].Key
	}
	allowed := false
	for _, opt := range options {
		if strings.EqualFold(opt.Key, docType) {
			allowed = true
			break
		}
	}
	if !allowed {
		fmt.Println(text("Invalid type.", "Invalid type."))
		return 2
	}
	if _, ok := easyTypeToTemplate(docType); !ok {
		fmt.Println(text("Invalid type.", "Invalid type."))
		return 2
	}

	folderPath, err := promptDocFolder(reader, docType)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	fileName, err := promptRequired(reader, text("File name", "파일 이름"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	title, err := promptRequired(reader, text("Title", "제목"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	owner, err := promptRequired(reader, text("Owner", "담당자"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	outPath := filepath.Join(folderPath, ensureAGDExt(fileName))
	if code := commandNewEasy([]string{docType, outPath, title, owner}); code != 0 {
		return code
	}

	createdPath, resolveErr := resolveAGDPathEasy(outPath, true)
	if resolveErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", resolveErr.Error())
		return 1
	}

	if code := wizardPromptRoleAfterNew(reader, createdPath); code != 0 {
		return code
	}
	return runGuidedFlowAfterMutation(createdPath)
}
func wizardCheck(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return wizardCheckOnDoc(filePath)
}

func wizardCheckOnDoc(filePath string) int {
	return commandLint([]string{filePath})
}

func wizardCheckAll(reader *bufio.Reader) int {
	root, err := promptOptional(reader, text("Root folder (Enter for agd_docs)", "Root folder (Enter for agd_docs)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	strictAnswer, err := promptOptional(reader, text("Strict mode? (y/N)", "Strict mode? (y/N)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	includeArchiveAnswer, err := promptOptional(reader, text("Include 90_archive? (y/N)", "Include 90_archive? (y/N)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	args := make([]string, 0, 3)
	if isYesInput(strictAnswer) {
		args = append(args, "--strict")
	}
	if isYesInput(includeArchiveAnswer) {
		args = append(args, "--include-archive")
	}
	if strings.TrimSpace(root) != "" {
		args = append(args, root)
	}
	return commandCheckAllEasy(args)
}

func wizardServiceGate(reader *bufio.Reader) int {
	serviceDoc, err := promptOptional(reader, text(
		"Service doc (Enter for auto-select)", "Service doc (Enter for auto-select)",
	))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	maxAgeDays, err := promptOptional(reader, text(
		"Max age days (default 0 = must update today)", "Max age days (default 0 = must update today)",
	))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	allowMissingImpactAnswer, err := promptOptional(reader, text(
		"Allow missing impact? (y/N)", "Allow missing impact? (y/N)",
	))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	args := make([]string, 0, 4)
	if strings.TrimSpace(serviceDoc) != "" {
		args = append(args, serviceDoc)
	}
	if strings.TrimSpace(maxAgeDays) != "" {
		args = append(args, "--max-age-days", strings.TrimSpace(maxAgeDays))
	}
	if isYesInput(allowMissingImpactAnswer) {
		args = append(args, "--allow-missing-impact")
	}
	return commandServiceGateEasy(args)
}

func wizardKit(reader *bufio.Reader) int {
	fmt.Println("")
	fmt.Println(text("Kit profile:", "Kit profile:"))
	fmt.Println(text("  [1] starter-kit     - initial source baseline", "  [1] starter-kit     - initial source baseline"))
	fmt.Println(text("  [2] maintenance     - single-file maintenance flow", "  [2] maintenance     - single-file maintenance flow"))
	fmt.Println(text("  [3] new-project     - new feature expansion flow", "  [3] new-project     - new feature expansion flow"))
	fmt.Println(text("  [4] incident-response - single-file incident trace flow", "  [4] incident-response - single-file incident trace flow"))
	fmt.Println(text("  [0] back", "  [0] back"))

	profileInput, err := promptRequired(reader, text("Profile (1-4 or name)", "Profile (1-4 or name)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	profileRaw := strings.TrimSpace(profileInput)
	if profileRaw == "0" {
		return 0
	}
	switch strings.ToLower(profileRaw) {
	case "1":
		profileRaw = "starter-kit"
	case "2":
		profileRaw = "maintenance"
	case "3":
		profileRaw = "new-project"
	case "4":
		profileRaw = "incident-response"
	}
	profile, ok := normalizeKitProfile(profileRaw)
	if !ok {
		fmt.Println(text("Invalid profile.", "Invalid profile."))
		return 2
	}

	projectKey, err := promptRequired(reader, text("Project key (ex: checkout)", "Project key (ex: checkout)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	owner, err := promptOptional(reader, text("Owner (Enter to leave empty)", "Owner (Enter to leave empty)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	root, err := promptOptional(reader, text("Root folder (Enter for agd_docs)", "Root folder (Enter for agd_docs)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	featureTag := ""
	incidentSourceDoc := ""
	incidentSourceSection := ""
	if profile == "incident-response" || profile == "maintenance" {
		selected, autoTag, err := promptIncidentSectionCandidate(reader, strings.TrimSpace(root))
		if err != nil {
			if errors.Is(err, errWizardBack) {
				return 0
			}
			fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
			return 1
		}
		featureTag = strings.TrimSpace(autoTag)
		incidentSourceDoc = strings.TrimSpace(selected.DocPath)
		incidentSourceSection = strings.TrimSpace(selected.SectionID)
		fmt.Printf(text(
			"Selected: %s | %s | %s\nFeature tag auto-set: %s\n", "Selected: %s | %s | %s\nFeature tag auto-set: %s\n",
		), selected.DocRel, incidentSourceSection, strings.TrimSpace(selected.SectionTitle), featureTag)
	}
	fmt.Printf(text(
		"Language auto-selected by build: %s\n", "Language auto-selected by build: %s\n",
	), activeLang())
	forceAnswer, err := promptOptional(reader, text("Overwrite existing files? (y/N)", "Overwrite existing files? (y/N)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	showGraphAnswer, err := promptOptional(reader, text("Show role graph after creation? (Y/n)", "Show role graph after creation? (Y/n)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	args := []string{"kit", profile, strings.TrimSpace(projectKey)}
	if strings.TrimSpace(owner) != "" {
		args = append(args, "--owner", strings.TrimSpace(owner))
	}
	if strings.TrimSpace(featureTag) != "" {
		args = append(args, "--feature-tag", strings.TrimSpace(featureTag))
	}
	if strings.TrimSpace(incidentSourceDoc) != "" && strings.TrimSpace(incidentSourceSection) != "" {
		args = append(args, "--tag-source", strings.TrimSpace(incidentSourceDoc), "--tag-section", strings.TrimSpace(incidentSourceSection))
	}
	if strings.TrimSpace(root) != "" {
		args = append(args, "--root", strings.TrimSpace(root))
	}
	if isYesInput(forceAnswer) {
		args = append(args, "--force")
	}
	if isNoInput(showGraphAnswer) {
		args = append(args, "--no-graph")
	}
	return commandKitEasy(args)
}
func wizardRoleGraph(reader *bufio.Reader) int {
	root, err := promptOptional(reader, text("Root (Enter for agd_docs)", "Root (Enter for agd_docs)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	fmt.Println(text("Format: [1] text, [2] mermaid, [0] back", "Format: [1] text, [2] mermaid, [0] back"))
	formatInput, err := promptOptional(reader, text("Format (Enter=1)", "Format (Enter=1)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	if strings.TrimSpace(formatInput) == "0" {
		return 0
	}
	includeArchiveAnswer, err := promptOptional(reader, text("Include 90_archive? (y/N)", "Include 90_archive? (y/N)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	outPath, err := promptOptional(reader, text("Output file (Enter to print on screen)", "Output file (Enter to print on screen)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	format := strings.ToLower(strings.TrimSpace(formatInput))
	switch format {
	case "", "1", "text":
		format = "text"
	case "2", "mermaid", "mmd":
		format = "mermaid"
	default:
		fmt.Println(text("Invalid format. Use text or mermaid.", "Invalid format. Use text or mermaid."))
		return 2
	}

	args := make([]string, 0, 7)
	if strings.TrimSpace(root) != "" {
		args = append(args, strings.TrimSpace(root))
	}
	if format != "text" {
		args = append(args, "--format", format)
	}
	if isYesInput(includeArchiveAnswer) {
		args = append(args, "--include-archive")
	}
	if strings.TrimSpace(outPath) != "" {
		args = append(args, "--out", strings.TrimSpace(outPath))
	}
	return commandRoleGraphEasy(args)
}
func wizardSearch(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return wizardSearchOnDoc(reader, filePath)
}

func wizardSearchOnDoc(reader *bufio.Reader, filePath string) int {
	keyword, err := promptRequired(reader, text("Keyword", "Keyword"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return commandFind([]string{filePath, keyword})
}

func wizardEdit(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	sectionID, err := promptRequired(reader, text("Section ID (ex: PRD-020)", "Section ID (ex: PRD-020)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	summary, err := promptRequired(reader, text("New summary", "New summary"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	reason, err := promptRequired(reader, text("Change reason", "Change reason"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	impact, err := promptRequired(reader, text("Change impact", "Change impact"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

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

func wizardAdd(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	sectionID, err := promptRequired(reader, text("Section ID (ex: PRD-020)", "Section ID (ex: PRD-020)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	line, err := promptRequired(reader, text("Line to append", "Line to append"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	reason, err := promptRequired(reader, text("Change reason", "Change reason"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	impact, err := promptRequired(reader, text("Change impact", "Change impact"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	if code := commandPatch([]string{
		filePath,
		"--section", sectionID,
		"--append-content", line,
		"--reason", reason,
		"--impact", impact,
	}); code != 0 {
		return code
	}
	return runGuidedFlowAfterMutation(filePath)
}

func wizardView(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return wizardViewOnDoc(reader, filePath)
}

func wizardViewOnDoc(reader *bufio.Reader, filePath string) int {
	outPath, err := promptOptional(reader, text("Output .md path (Enter for auto)", "Output .md path (Enter for auto)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	if strings.TrimSpace(outPath) == "" {
		return commandViewEasy([]string{filePath})
	}
	return commandViewEasy([]string{filePath, outPath})
}

func wizardSectionAdd(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return wizardSectionAddOnDoc(reader, filePath)
}

func wizardSectionAddOnDoc(reader *bufio.Reader, filePath string) int {
	sectionID, err := promptRequired(reader, text("Section ID (ex: PRD-050)", "Section ID (ex: PRD-050)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	title, err := promptRequired(reader, text("Section title", "Section title"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	summary, err := promptRequired(reader, text("Section summary", "Section summary"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	links, err := promptOptional(reader, text("Links (comma-separated, optional)", "Links (comma-separated, optional)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	content, err := promptOptional(reader, text("First content line (optional)", "First content line (optional)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	reason, err := promptRequired(reader, text("Change reason", "Change reason"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	impact, err := promptRequired(reader, text("Change impact", "Change impact"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	if code := addSectionWithMapUpdate(filePath, sectionID, title, summary, links, content, "ai-agent", reason, impact); code != 0 {
		return code
	}
	return runGuidedFlowAfterMutation(filePath)
}

func wizardMapSync(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return wizardMapSyncOnDoc(reader, filePath)
}

func wizardMapSyncOnDoc(reader *bufio.Reader, filePath string) int {
	reason, err := promptRequired(reader, text("Map sync reason", "Map sync reason"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	impact, err := promptRequired(reader, text("Map sync impact", "Map sync impact"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return commandMapSyncEasy([]string{filePath, "--reason", reason, "--impact", impact})
}

func wizardLogicLog(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return wizardLogicLogOnDoc(reader, filePath)
}

func wizardLogicLogOnDoc(reader *bufio.Reader, filePath string) int {
	sectionID, err := promptRequired(reader, text("Section ID (ex: PRD-020)", "Section ID (ex: PRD-020)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	reason, err := promptRequired(reader, text("Change reason", "Change reason"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	impact, err := promptRequired(reader, text("Change impact", "Change impact"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return commandLogicLogEasy([]string{filePath, sectionID, "--reason", reason, "--impact", impact})
}

func wizardRoleSet(reader *bufio.Reader) int {
	filePath, err := promptExistingDocPath(reader)
	if err != nil {
		if errors.Is(err, errWizardBack) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}
	return wizardRoleSetOnDoc(reader, filePath)
}

func wizardRoleSetOnDoc(reader *bufio.Reader, filePath string) int {
	fmt.Println(text("[1] source (authoritative)", "[1] source (authoritative)"))
	fmt.Println(text("[2] derived (references source)", "[2] derived (references source)"))
	fmt.Println(text("[0] Back", "[0] 뒤로가기"))
	roleChoice, err := promptRequired(reader, text("Role (1-2)", "Role (1-2)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	role := strings.ToLower(strings.TrimSpace(roleChoice))
	switch role {
	case "0":
		return 0
	case "1", "source":
		return commandRoleSetEasy([]string{filePath, "source"})
	case "2", "derived":
		sourceDoc, err := promptRequired(reader, text("Source document (name/path)", "Source document (name/path)"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
			return 1
		}
		sourceSections, err := promptOptional(reader, text(
			"Source sections mapping CSV (optional, ex: SYS-020->FP-020,SYS-030=>FP-030, auto, strict-auto, smart-auto)", "Source sections mapping CSV (optional, ex: SYS-020->FP-020,SYS-030=>FP-030, auto, strict-auto, smart-auto)",
		))
		if err != nil {
			fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
			return 1
		}
		args := []string{filePath, "derived", sourceDoc}
		if strings.TrimSpace(sourceSections) != "" {
			args = append(args, sourceSections)
		}
		return commandRoleSetEasy(args)
	default:
		fmt.Println(text("Invalid role.", "Invalid role."))
		return 2
	}
}
func wizardPromptRoleAfterNew(reader *bufio.Reader, filePath string) int {
	fmt.Println("")
	fmt.Println(text(
		"Set document role now (recommended):", "Set document role now (recommended):",
	))
	fmt.Println(text("[1] source (authoritative)", "[1] source (authoritative)"))
	fmt.Println(text("[2] derived (references source)", "[2] derived (references source)"))
	fmt.Println(text("[3] later (skip for now)", "[3] later (skip for now)"))
	fmt.Println(text("[0] Back", "[0] 뒤로가기"))

	choice, err := promptRequired(reader, text("Role (1-3)", "Role (1-3)"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		return 1
	}

	role := strings.ToLower(strings.TrimSpace(choice))
	switch role {
	case "0":
		return 0
	case "1", "source":
		return commandRoleSetEasy([]string{filePath, "source"})
	case "2", "derived":
		sourceDoc, err := promptRequired(reader, text("Source document (name/path)", "Source document (name/path)"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
			return 1
		}
		sourceSections, err := promptOptional(reader, text(
			"Source sections mapping CSV (optional, ex: SYS-020->FP-020,SYS-030=>FP-030, auto, strict-auto, smart-auto)", "Source sections mapping CSV (optional, ex: SYS-020->FP-020,SYS-030=>FP-030, auto, strict-auto, smart-auto)",
		))
		if err != nil {
			fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
			return 1
		}
		args := []string{filePath, "derived", sourceDoc}
		if strings.TrimSpace(sourceSections) != "" {
			args = append(args, sourceSections)
		}
		return commandRoleSetEasy(args)
	case "3", "later", "skip":
		fmt.Println(text(
			"Skipped role setup. Set it soon with: agd role-set <file.agd> <source|derived> ...", "Skipped role setup. Set it soon with: agd role-set <file.agd> <source|derived> ...",
		))
		return 0
	default:
		fmt.Println(text("Invalid role.", "Invalid role."))
		return 2
	}
}
func runGuidedFlowAfterMutation(filePath string) int {
	fmt.Println(text("Guided flow: sync map", "Guided flow: sync map"))
	if code := commandMapSyncEasy([]string{
		filePath,
		"--reason", "guided flow post-mutation map sync",
		"--impact", "map index aligned to latest section state",
	}); code != 0 {
		return code
	}
	fmt.Println(text("Guided flow: check document", "Guided flow: check document"))
	return commandLint([]string{filePath})
}
func folderDescription(rel string) string {
	norm := filepath.ToSlash(strings.TrimSpace(rel))
	norm = strings.TrimPrefix(norm, "./")
	norm = strings.Trim(norm, "/")
	normLower := strings.ToLower(norm)

	switch normLower {
	case "":
		return text("all documents root", "전체 문서 루트")
	case "00_inbox":
		return text("temporary drafts and triage", "임시 초안 및 분류")
	case "10_source":
		return text("source-of-truth documents", "기준(source) 문서")
	case "10_source/product":
		return text("product requirements source docs", "제품 요구사항 source 문서")
	case "10_source/service":
		return text("service-wide logic source docs", "서비스 전역 로직 source 문서")
	case "10_source/policy":
		return text("policy/guideline source docs", "정책/가이드 source 문서")
	case "10_source/architecture":
		return text("architecture decision source docs", "아키텍처 의사결정 source 문서")
	case "20_derived":
		return text("documents derived from source", "source 기반 파생 문서")
	case "20_derived/frontend":
		return text("frontend page logic derived docs", "프런트엔드 페이지 로직 파생 문서")
	case "20_derived/ops":
		return text("operations/runbook derived docs", "운영/runbook 파생 문서")
	case "20_derived/qa":
		return text("QA/test derived docs", "QA/test 파생 문서")
	case "30_shared":
		return text("shared collaboration documents", "공유 협업 문서")
	case "30_shared/meeting":
		return text("meeting notes and decisions", "회의 기록 및 의사결정")
	case "30_shared/handoff":
		return text("cross-team handoff docs", "팀 간 인수인계 문서")
	case "30_shared/postmortem":
		return text("incident retrospectives", "오류,버그 회고 문서")
	case "30_shared/maintenance":
		return text("single-file maintenance workspaces", "단일 파일 유지보수 워크스페이스")
	case "30_shared/errfix":
		return text("single-file incident trace/fix workspaces", "단일 파일 오류,버그 추적/수정 워크스페이스")
	case "30_shared/roadmap":
		return text("roadmaps and milestones", "로드맵 및 마일스톤")
	case "30_shared/experiment":
		return text("experiments and hypotheses", "실험 및 가설")
	case "90_archive":
		return text("archived/retired documents", "보관/종료 문서")
	}

	switch {
	case strings.HasPrefix(normLower, "10_source/"):
		return text("custom source subfolder", "사용자 지정 source 하위 폴더")
	case strings.HasPrefix(normLower, "20_derived/"):
		return text("custom derived subfolder", "사용자 지정 derived 하위 폴더")
	case strings.HasPrefix(normLower, "30_shared/"):
		return text("custom shared subfolder", "사용자 지정 shared 하위 폴더")
	default:
		return text("custom folder", "사용자 지정 폴더")
	}
}

func isPathWithinBase(base, target string) bool {
	baseClean := filepath.Clean(base)
	targetClean := filepath.Clean(target)

	rel, err := filepath.Rel(baseClean, targetClean)
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
	if strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	return !filepath.IsAbs(rel)
}

func promptDocFolder(reader *bufio.Reader, docType string) (string, error) {
	if err := ensureDocsRoot(); err != nil {
		return "", err
	}
	baseRel := defaultSubdirForDocType(docType)
	basePath := filepath.Clean(filepath.Join(docsRootDir, baseRel))
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return "", err
	}

	subdirs, err := listSubfolders(basePath)
	if err != nil {
		return "", err
	}

	fmt.Println(text(
		"Available folders for this document type:",
		"이 문서 타입에서 선택 가능한 폴더:",
	))
	fmt.Printf("  [1] %s - %s\n", basePath, folderDescription(baseRel))
	for i, dir := range subdirs {
		fullPath := filepath.Join(basePath, dir)
		relToDocs, relErr := filepath.Rel(docsRootDir, fullPath)
		if relErr != nil {
			relToDocs = fullPath
		}
		fmt.Printf("  [%d] %s - %s\n", i+2, fullPath, folderDescription(relToDocs))
	}
	fmt.Println(text("  [0] Back", "  [0] 뒤로가기"))
	fmt.Println(text(
		"Only this type's allowed path can be selected.",
		"이 타입에서 허용된 경로만 선택할 수 있습니다.",
	))

	for {
		answer, err := promptOptional(reader, text(
			"Folder number or new subfolder (Enter for default)",
			"폴더 번호 또는 새 하위 폴더 (기본값은 Enter)",
		))
		if err != nil {
			return "", err
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			return basePath, nil
		}
		if answer == "0" {
			return "", errWizardBack
		}

		if idx, convErr := strconv.Atoi(answer); convErr == nil {
			switch {
			case idx == 1:
				return basePath, nil
			case idx >= 2 && idx <= len(subdirs)+1:
				return filepath.Join(basePath, subdirs[idx-2]), nil
			default:
				fmt.Println(text("Invalid folder number.", "Invalid folder number."))
				continue
			}
		}

		var candidate string
		switch {
		case filepath.IsAbs(answer):
			candidate = filepath.Clean(answer)
		case isUnderDocsRoot(answer):
			candidate = filepath.Clean(answer)
		default:
			candidate = filepath.Clean(filepath.Join(basePath, answer))
		}

		if !isPathWithinBase(basePath, candidate) {
			fmt.Println(text(
				"That path is outside the allowed folder for this type.",
				"That path is outside the allowed folder for this type.",
			))
			continue
		}
		return candidate, nil
	}
}
func promptExistingDocPath(reader *bufio.Reader) (string, error) {
	if err := ensureDocsRoot(); err != nil {
		return "", err
	}

	files, err := listAGDFiles(docsRootDir)
	if err != nil {
		return "", err
	}
	if len(files) > 0 {
		fmt.Println(text("Available AGD files:", "Available AGD files:"))
		for i, file := range files {
			rel, relErr := filepath.Rel(docsRootDir, file)
			if relErr != nil {
				rel = file
			}
			fmt.Printf("  [%d] %s\n", i+1, filepath.Join(docsRootDir, rel))
		}
		fmt.Println(text("  [0] Back", "  [0] Back"))
	}

	for {
		answer, err := promptRequired(reader, text("File (number or path/name)", "File (number or path/name)"))
		if err != nil {
			return "", err
		}
		answer = strings.TrimSpace(answer)
		if answer == "0" {
			return "", errWizardBack
		}

		if idx, convErr := strconv.Atoi(answer); convErr == nil {
			if idx >= 1 && idx <= len(files) {
				return files[idx-1], nil
			}
			fmt.Println(text("Invalid file number.", "Invalid file number."))
			continue
		}

		resolved, resolveErr := resolveAGDPathEasy(answer, true)
		if resolveErr != nil {
			fmt.Println(resolveErr.Error())
			continue
		}
		return resolved, nil
	}
}

type incidentSectionCandidate struct {
	Domain       string
	DocPath      string
	DocRel       string
	SectionID    string
	SectionTitle string
}

func collectIncidentSectionCandidates(root string) ([]incidentSectionCandidate, error) {
	baseRoot := strings.TrimSpace(root)
	if baseRoot == "" {
		baseRoot = docsRootDir
	}
	if _, err := os.Stat(baseRoot); err != nil {
		if os.IsNotExist(err) {
			return []incidentSectionCandidate{}, nil
		}
		return nil, err
	}

	targets := []struct {
		domain string
		rel    string
	}{
		{domain: "service", rel: filepath.Join("10_source", "service")},
		{domain: "frontend", rel: filepath.Join("20_derived", "frontend")},
	}

	out := make([]incidentSectionCandidate, 0)
	for _, target := range targets {
		targetDir := filepath.Join(baseRoot, target.rel)
		if _, err := os.Stat(targetDir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		files, err := listAGDFiles(targetDir)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			doc, parseErrs, loadErr := agd.LoadFile(file)
			if loadErr != nil || len(parseErrs) > 0 {
				continue
			}
			docPath := filepath.Clean(file)
			docRel := relativePathForReport(baseRoot, docPath)
			for _, section := range doc.Sections {
				sectionID := strings.TrimSpace(section.ID)
				if sectionID == "" {
					continue
				}
				out = append(out, incidentSectionCandidate{
					Domain:       target.domain,
					DocPath:      docPath,
					DocRel:       docRel,
					SectionID:    sectionID,
					SectionTitle: strings.TrimSpace(section.Title),
				})
			}
		}
	}

	rank := map[string]int{"service": 0, "frontend": 1}
	sort.Slice(out, func(i, j int) bool {
		ri := rank[out[i].Domain]
		rj := rank[out[j].Domain]
		if ri != rj {
			return ri < rj
		}
		li := strings.ToLower(out[i].DocRel)
		lj := strings.ToLower(out[j].DocRel)
		if li != lj {
			return li < lj
		}
		return strings.ToUpper(out[i].SectionID) < strings.ToUpper(out[j].SectionID)
	})
	return out, nil
}

func defaultIncidentFeatureTagFromSection(sectionID, sectionTitle string) string {
	idSeed := strings.TrimSpace(sectionID)
	if idSeed != "" {
		if tag := normalizeIncidentFeatureTag("FT-"+idSeed, ""); tag != "" {
			return tag
		}
	}
	titleSeed := strings.TrimSpace(sectionTitle)
	if titleSeed != "" {
		return normalizeIncidentFeatureTag("FT-"+titleSeed, "")
	}
	return ""
}

func promptIncidentSectionCandidate(reader *bufio.Reader, root string) (incidentSectionCandidate, string, error) {
	scanRoot := strings.TrimSpace(root)
	if scanRoot == "" {
		scanRoot = docsRootDir
	}
	candidates, err := collectIncidentSectionCandidates(scanRoot)
	if err != nil {
		return incidentSectionCandidate{}, "", err
	}
	if len(candidates) == 0 && !strings.EqualFold(filepath.Clean(scanRoot), filepath.Clean(docsRootDir)) {
		candidates, err = collectIncidentSectionCandidates(docsRootDir)
		if err != nil {
			return incidentSectionCandidate{}, "", err
		}
	}
	if len(candidates) == 0 {
		return incidentSectionCandidate{}, "", fmt.Errorf("%s", text(
			"no sections found under agd_docs/10_source/service or agd_docs/20_derived/frontend",
			"no sections found under agd_docs/10_source/service or agd_docs/20_derived/frontend",
		))
	}

	fmt.Println(text(
		"Select issue-root section (scanned from service/frontend):", "Select issue-root section (scanned from service/frontend):",
	))
	for i, c := range candidates {
		title := strings.TrimSpace(c.SectionTitle)
		if title == "" {
			title = text("(untitled)", "(untitled)")
		}
		fmt.Printf("  [%d] [%s] %s | %s | %s\n", i+1, c.Domain, c.DocRel, c.SectionID, title)
	}
	fmt.Println(text("  [0] Back", "  [0] Back"))

	for {
		answer, promptErr := promptRequired(reader, text(
			"Section number or SECTION-ID", "Section number or SECTION-ID",
		))
		if promptErr != nil {
			return incidentSectionCandidate{}, "", promptErr
		}
		answer = strings.TrimSpace(answer)
		if answer == "0" {
			return incidentSectionCandidate{}, "", errWizardBack
		}

		if idx, convErr := strconv.Atoi(answer); convErr == nil {
			if idx >= 1 && idx <= len(candidates) {
				selected := candidates[idx-1]
				tag := defaultIncidentFeatureTagFromSection(selected.SectionID, selected.SectionTitle)
				if tag == "" {
					fmt.Println(text("Cannot build feature tag from selected section.", "Cannot build feature tag from selected section."))
					continue
				}
				return selected, tag, nil
			}
			fmt.Println(text("Invalid section number.", "Invalid section number."))
			continue
		}

		matches := make([]incidentSectionCandidate, 0, 1)
		for _, c := range candidates {
			if strings.EqualFold(strings.TrimSpace(c.SectionID), answer) {
				matches = append(matches, c)
			}
		}
		if len(matches) == 1 {
			selected := matches[0]
			tag := defaultIncidentFeatureTagFromSection(selected.SectionID, selected.SectionTitle)
			if tag == "" {
				fmt.Println(text("Cannot build feature tag from selected section.", "Cannot build feature tag from selected section."))
				continue
			}
			return selected, tag, nil
		}
		if len(matches) > 1 {
			fmt.Println(text(
				"Multiple sections share that SECTION-ID. Choose by number.", "Multiple sections share that SECTION-ID. Choose by number.",
			))
			continue
		}

		fmt.Println(text("Section not found. Enter a valid number or SECTION-ID.", "Section not found. Enter a valid number or SECTION-ID."))
	}
}
func resolveSectionIDChoice(doc *agd.Document, choice string) (string, error) {
	if doc == nil {
		return "", fmt.Errorf("%s", text("document is empty", "document is empty"))
	}
	if len(doc.Sections) == 0 {
		return "", fmt.Errorf("%s", text("document has no sections", "document has no sections"))
	}
	raw := strings.TrimSpace(choice)
	if raw == "" {
		return "", fmt.Errorf("%s", text("section choice is required", "section choice is required"))
	}
	if idx, err := strconv.Atoi(raw); err == nil {
		if idx < 1 || idx > len(doc.Sections) {
			return "", fmt.Errorf("%s", text("invalid section number", "invalid section number"))
		}
		return strings.TrimSpace(doc.Sections[idx-1].ID), nil
	}
	section := doc.SectionByID(raw)
	if section == nil {
		return "", fmt.Errorf(text("section not found: %s", "section not found: %s"), raw)
	}
	return strings.TrimSpace(section.ID), nil
}

func promptSectionIDChoiceFromDoc(reader *bufio.Reader, docPath string, doc *agd.Document) (string, error) {
	if doc == nil {
		return "", fmt.Errorf("%s", text("document is empty", "document is empty"))
	}
	if len(doc.Sections) == 0 {
		return "", fmt.Errorf("%s", text("document has no sections", "document has no sections"))
	}

	fmt.Printf(text("Section IDs in %s:\n", "Section IDs in %s:\n"), filepath.Clean(docPath))
	for i, section := range doc.Sections {
		fmt.Printf("  [%d] %s | %s\n", i+1, strings.TrimSpace(section.ID), strings.TrimSpace(section.Title))
	}
	fmt.Println(text(
		"Choose a number or type SECTION-ID directly.", "Choose a number or type SECTION-ID directly.",
	))

	for {
		answer, err := promptRequired(reader, text("Section number or ID", "Section number or ID"))
		if err != nil {
			return "", err
		}
		sectionID, resolveErr := resolveSectionIDChoice(doc, answer)
		if resolveErr != nil {
			fmt.Println(resolveErr.Error())
			continue
		}
		return sectionID, nil
	}
}

func promptRequired(reader *bufio.Reader, label string) (string, error) {
	for {
		value, err := promptOptional(reader, label)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value), nil
		}
		fmt.Println(text("This field is required.", "This field is required."))
	}
}

func promptOptional(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	if err == io.EOF && strings.TrimSpace(line) == "" {
		return "", io.EOF
	}
	return strings.TrimSpace(line), nil
}

func isYesInput(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes", "1", "true", "on", "ye", "yeah":
		return true
	default:
		return false
	}
}

func isNoInput(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "n", "no", "0", "false", "off":
		return true
	default:
		return false
	}
}
