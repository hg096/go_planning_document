// commands_search_new.go handles search/find and new-document command entry points.
// commands_search_new.go는 search/find와 새 문서 생성 진입 명령을 처리합니다.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agd/internal/agd"
)

func commandSearchEasy(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, text("usage: agd search <file.agd> <keyword>", "usage: agd search <file.agd> <keyword>"))
		return 2
	}
	path, err := resolveAGDPathEasy(args[0], true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	keyword := strings.Join(args[1:], " ")
	return commandFind([]string{path, keyword})
}
func commandFind(args []string) int {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, "usage: agd find <file.agd> <query>\n")
		return 2
	}
	path := args[0]
	query := strings.TrimSpace(strings.Join(args[1:], " "))

	doc, parseErrs, err := agd.LoadFile(path)
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

	hits := agd.Search(doc, query)
	if len(hits) == 0 {
		fmt.Println("No matches.")
		return 0
	}

	for _, hit := range hits {
		fmt.Printf("[%s %s] %s: %s\n", hit.Kind, hit.ID, hit.Field, hit.Snippet)
	}
	return 0
}

func commandNewEasy(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, text("usage: agd new <type> <out.agd> [title] [owner]", "usage: agd new <type> <out.agd> [title] [owner]"))
		fmt.Fprintln(os.Stderr, text("types:", "types:"))
		printAllowedCreateDocTypeCatalog(os.Stderr, false)
		return 2
	}
	if strings.EqualFold(strings.TrimSpace(args[0]), "list") {
		fmt.Println(text("types:", "types:"))
		printAllowedCreateDocTypeCatalog(os.Stdout, false)
		return 0
	}
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, text("usage: agd new <type> <out.agd> [title] [owner]", "usage: agd new <type> <out.agd> [title] [owner]"))
		return 2
	}
	if !isAllowedCreateDocType(args[0]) {
		fmt.Fprintf(os.Stderr, text("new error: unknown type %q\n", "new error: unknown type %q\n"), args[0])
		fmt.Fprintln(os.Stderr, text("types:", "types:"))
		printAllowedCreateDocTypeCatalog(os.Stderr, false)
		return 2
	}

	templateName, ok := easyTypeToTemplate(args[0])
	if !ok {
		fmt.Fprintf(os.Stderr, text("new error: unknown type %q\n", "new error: unknown type %q\n"), args[0])
		fmt.Fprintln(os.Stderr, text("types:", "types:"))
		printAllowedCreateDocTypeCatalog(os.Stderr, false)
		return 2
	}

	outPath, err := resolveNewDocPathByType(args[0], args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	rawOut := strings.TrimSpace(args[1])
	if !isExplicitPath(rawOut) && !hasPathSeparator(rawOut) && !isUnderDocsRoot(rawOut) {
		autoTarget := filepath.Clean(filepath.Join(docsRootDir, defaultSubdirForDocType(args[0]), ensureAGDExt(rawOut)))
		if strings.EqualFold(filepath.Clean(outPath), autoTarget) {
			fmt.Printf(text(
				"Auto-routed by document type -> %s\n", "Auto-routed by document type -> %s\n",
			), outPath)
		} else {
			fmt.Printf(text(
				"Resolved existing file path -> %s\n", "Resolved existing file path -> %s\n",
			), outPath)
		}
	}
	title := ""
	owner := ""
	if len(args) >= 3 {
		title = args[2]
	}
	if len(args) >= 4 {
		owner = args[3]
	}

	initArgs := []string{
		"--template", templateName,
		"--out", outPath,
	}
	if strings.TrimSpace(title) != "" {
		initArgs = append(initArgs, "--title", title)
	}
	if strings.TrimSpace(owner) != "" {
		initArgs = append(initArgs, "--owner", owner)
	}
	return commandInit(initArgs)
}
