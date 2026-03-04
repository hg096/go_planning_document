// commands_export_init.go handles view/patch/export/init command flows.
// commands_export_init.go는 view/patch/export/init 명령 흐름을 처리합니다.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agd/internal/agd"
	"agd/internal/agdtemplates"
)

func commandViewEasy(args []string) int {
	if len(args) == 0 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, "usage: agd view <file.agd> [out.md]")
		return 2
	}

	inputPath, err := resolveAGDPathEasy(args[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	if len(args) == 2 {
		return commandExport([]string{inputPath, "--format", "md", "--out", args[1]})
	}

	outPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".md"
	return commandExport([]string{inputPath, "--format", "md", "--out", outPath})
}

func commandPatch(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: agd patch <file.agd> --section <SEC-ID> [...flags]\n")
		return 2
	}
	inputPath := args[0]

	fs := flag.NewFlagSet("patch", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	sectionID := fs.String("section", "", "target section id")
	title := fs.String("title", "", "new section title")
	summary := fs.String("summary", "", "new section summary")
	setContent := fs.String("set-content", "", "replace full section content")
	appendContent := fs.String("append-content", "", "append text to section content")
	links := fs.String("links", "", "comma-separated links")
	reason := fs.String("reason", "", "create change log reason")
	note := fs.String("note", "", "create change log note")
	impact := fs.String("impact", "", "change impact summary")
	author := fs.String("author", "ai-agent", "change author")
	outPath := fs.String("out", "", "output file path")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "usage: agd patch <file.agd> --section <SEC-ID> [...flags]\n")
		return 2
	}
	if len(fs.Args()) != 0 {
		fmt.Fprintf(os.Stderr, "usage: agd patch <file.agd> --section <SEC-ID> [...flags]\n")
		return 2
	}
	if strings.TrimSpace(*sectionID) == "" {
		fmt.Fprintln(os.Stderr, "patch error: --section is required")
		return 2
	}
	hasSectionMutation := strings.TrimSpace(*title) != "" ||
		strings.TrimSpace(*summary) != "" ||
		strings.TrimSpace(*setContent) != "" ||
		strings.TrimSpace(*appendContent) != "" ||
		strings.TrimSpace(*links) != ""

	reasonValue := strings.TrimSpace(*reason)
	noteValue := strings.TrimSpace(*note)
	if reasonValue != "" && noteValue != "" {
		fmt.Fprintln(os.Stderr, "patch error: use either --reason or --note, not both")
		return 2
	}
	if reasonValue == "" {
		reasonValue = noteValue
	}
	impactValue := strings.TrimSpace(*impact)

	if !hasSectionMutation && reasonValue == "" && impactValue == "" {
		fmt.Fprintln(os.Stderr, "patch error: no mutation flags provided")
		return 2
	}
	if reasonValue == "" {
		fmt.Fprintln(os.Stderr, "patch error: --reason is required for all mutations")
		return 2
	}
	if impactValue == "" {
		fmt.Fprintln(os.Stderr, "patch error: --impact is required for all mutations")
		return 2
	}

	doc, parseErrs, err := agd.LoadFile(inputPath)
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

	section := doc.SectionByID(*sectionID)
	if section == nil {
		fmt.Fprintf(os.Stderr, "patch error: section %s not found\n", *sectionID)
		return 1
	}

	if strings.TrimSpace(*title) != "" {
		section.Title = *title
	}
	if strings.TrimSpace(*summary) != "" {
		section.Summary = *summary
	}
	if strings.TrimSpace(*setContent) != "" {
		section.Content = *setContent
	}
	if strings.TrimSpace(*appendContent) != "" {
		if strings.TrimSpace(section.Content) == "" {
			section.Content = *appendContent
		} else {
			section.Content = strings.TrimRight(section.Content, "\n") + "\n" + *appendContent
		}
	}
	if strings.TrimSpace(*links) != "" {
		section.Links = splitCSV(*links)
	}

	mapEntry := doc.MapByID(section.ID)
	if mapEntry != nil {
		if strings.TrimSpace(*title) != "" {
			mapEntry.Title = section.Title
		}
		if strings.TrimSpace(*summary) != "" {
			mapEntry.Summary = section.Summary
		}
	}

	if reasonValue != "" {
		change := &agd.Change{
			ID:     nextChangeID(doc),
			Target: section.ID,
			Date:   time.Now().Format("2006-01-02"),
			Author: strings.TrimSpace(*author),
			Reason: reasonValue,
			Impact: impactValue,
		}
		doc.Changes = append(doc.Changes, change)
	}

	outputPath := inputPath
	if strings.TrimSpace(*outPath) != "" {
		outputPath = *outPath
	}
	ensureDocPathMetadata(doc, outputPath)

	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "lint error: %s\n", e)
		}
		return 1
	}

	if err := os.WriteFile(outputPath, []byte(agd.Serialize(doc)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		return 1
	}

	fmt.Printf("Patched section %s -> %s\n", section.ID, outputPath)
	return 0
}

func commandExport(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "usage: agd export <file.agd> --format md [--out file.md]\n")
		return 2
	}
	inputPath := args[0]

	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	format := fs.String("format", "md", "export format (md)")
	outPath := fs.String("out", "", "output file path")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprint(os.Stderr, "usage: agd export <file.agd> --format md [--out file.md]\n")
		return 2
	}

	if len(fs.Args()) != 0 {
		fmt.Fprint(os.Stderr, "usage: agd export <file.agd> --format md [--out file.md]\n")
		return 2
	}

	doc, parseErrs, err := agd.LoadFile(inputPath)
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

	if strings.ToLower(strings.TrimSpace(*format)) != "md" {
		fmt.Fprintf(os.Stderr, "export error: unsupported format %q (supported: md)\n", *format)
		return 2
	}

	rendered := agd.ToMarkdown(doc)
	if strings.TrimSpace(*outPath) == "" {
		fmt.Print(rendered)
		return 0
	}

	if err := os.WriteFile(*outPath, []byte(rendered), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		return 1
	}
	fmt.Printf("Exported markdown -> %s\n", *outPath)
	return 0
}

func commandInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	templateName := fs.String("template", "", "template name")
	outPath := fs.String("out", "", "output file path")
	title := fs.String("title", "", "document title")
	owner := fs.String("owner", "", "document owner")
	version := fs.String("version", "0.1", "document version")
	list := fs.Bool("list", false, "list available templates")
	force := fs.Bool("force", false, "overwrite output if file exists")
	allowedTemplates := allowedCLITemplates()

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "usage: agd init --template <name> --out <file.agd> [--title ...] [--owner ...] [--version ...] [--force]")
		fmt.Fprintln(os.Stderr, "   or: agd init --list")
		return 2
	}
	if len(fs.Args()) != 0 {
		fmt.Fprintln(os.Stderr, "usage: agd init --template <name> --out <file.agd> [--title ...] [--owner ...] [--version ...] [--force]")
		fmt.Fprintln(os.Stderr, "   or: agd init --list")
		return 2
	}

	if *list {
		fmt.Println("Available templates:")
		for _, name := range allowedTemplates {
			fmt.Printf("- %s\n", name)
		}
		return 0
	}

	if strings.TrimSpace(*templateName) == "" {
		fmt.Fprintln(os.Stderr, "init error: --template is required")
		return 2
	}
	resolvedTemplateName := strings.ToLower(strings.TrimSpace(*templateName))
	if !isAllowedCLITemplate(resolvedTemplateName) {
		fmt.Fprintf(os.Stderr, "init error: template %q is not allowed in CLI\n", strings.TrimSpace(*templateName))
		fmt.Fprintln(os.Stderr, "allowed templates:")
		for _, name := range allowedTemplates {
			fmt.Fprintf(os.Stderr, "- %s\n", name)
		}
		return 2
	}
	if strings.TrimSpace(*outPath) == "" {
		fmt.Fprintln(os.Stderr, "init error: --out is required")
		return 2
	}

	rendered, err := agdtemplates.Render(resolvedTemplateName, agdtemplates.RenderOptions{
		Title:   *title,
		Owner:   *owner,
		Version: *version,
		Date:    time.Now().Format("2006-01-02"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "init error: %v\n", err)
		return 1
	}

	doc, parseErrs := agd.Parse(rendered)
	if len(parseErrs) > 0 {
		for _, e := range parseErrs {
			fmt.Fprintf(os.Stderr, "template parse error: %s\n", e)
		}
		return 1
	}
	target := strings.TrimSpace(*outPath)
	ensureDocPathMetadata(doc, target)

	validationErrors := agd.Validate(doc)
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "template lint error: %s\n", e)
		}
		return 1
	}

	if _, err := os.Stat(target); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "init error: %s already exists (use --force to overwrite)\n", target)
		return 1
	}

	dir := filepath.Dir(target)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "init error: cannot create directory %s: %v\n", dir, err)
			return 1
		}
	}

	if err := os.WriteFile(target, []byte(agd.Serialize(doc)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "init error: write failed: %v\n", err)
		return 1
	}

	fmt.Printf("Created %s from template %s\n", target, resolvedTemplateName)
	return 0
}

func easyTypeToTemplate(kind string) (string, bool) {
	raw := strings.ToLower(strings.TrimSpace(kind))
	if raw == "" {
		return "", false
	}

	lang := "en"
	if isKO() {
		lang = "ko"
	}
	if strings.HasSuffix(raw, "-ko") {
		lang = "ko"
		raw = strings.TrimSuffix(raw, "-ko")
	}
	if strings.HasSuffix(raw, "-en") {
		lang = "en"
		raw = strings.TrimSuffix(raw, "-en")
	}

	switch raw {
	case "prd", "adr", "guide", "ai-guide", "ai-planning-guide", "aiguide":
		return "core-spec-" + lang, true
	case "roadmap":
		return "roadmap-" + lang, true
	case "handoff":
		return "handoff-" + lang, true
	case "meeting", "minutes":
		return "meeting-" + lang, true
	case "policy":
		return "policy-" + lang, true
	case "experiment":
		return "experiment-" + lang, true
	case "service", "system", "service-logic":
		return "core-spec-" + lang, true
	case "frontend", "frontend-page", "page-logic", "qa", "qa-plan", "qaplan", "runbook":
		return "delivery-plan-" + lang, true
	case "incident", "incident-case", "incident-root", "incident-workspace":
		return "incident-case-" + lang, true
	case "maintenance", "maintenance-case", "maintenance-workspace":
		return "maintenance-case-" + lang, true
	case "core", "core-spec":
		return "core-spec-" + lang, true
	case "delivery", "delivery-plan":
		return "delivery-plan-" + lang, true
	default:
		return "", false
	}
}
