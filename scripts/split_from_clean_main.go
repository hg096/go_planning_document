package main

import (
	"fmt"
	"os"
	"strings"
)

func find(lines []string, marker string) int {
	for i, line := range lines {
		if strings.HasPrefix(line, marker) {
			return i
		}
	}
	panic("marker not found: " + marker)
}

func write(path string, lines []string) {
	data := strings.Join(lines, "\n")
	if !strings.HasSuffix(data, "\n") {
		data += "\n"
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		panic(err)
	}
}

func main() {
	src, err := os.ReadFile("tmp_extract/cmd/agd/main.go")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n")

	idxPathStart := find(lines, "func ensureDocsRoot() error {")
	idxWizardStart := find(lines, "func commandWizard(_ []string) int {")
	idxFindStart := find(lines, "func commandFind(args []string) int {")
	idxKitStart := find(lines, "type kitDocSpec struct {")
	idxAddSectionStart := find(lines, "func addSectionWithMapUpdate(")

	pathBlock := lines[idxPathStart:idxWizardStart]
	wizardBlock := lines[idxWizardStart:idxFindStart]
	kitBlock := lines[idxKitStart:idxAddSectionStart]

	pathFile := append([]string{
		"package main",
		"",
		"import (",
		"\t\"fmt\"",
		"\t\"io/fs\"",
		"\t\"os\"",
		"\t\"path/filepath\"",
		"\t\"sort\"",
		"\t\"strings\"",
		")",
		"",
	}, pathBlock...)

	wizardFile := append([]string{
		"package main",
		"",
		"import (",
		"\t\"bufio\"",
		"\t\"fmt\"",
		"\t\"io\"",
		"\t\"os\"",
		"\t\"path/filepath\"",
		"\t\"sort\"",
		"\t\"strconv\"",
		"\t\"strings\"",
		"",
		"\t\"agd/internal/agd\"",
		")",
		"",
	}, wizardBlock...)

	kitFile := append([]string{
		"package main",
		"",
		"import (",
		"\t\"fmt\"",
		"\t\"os\"",
		"\t\"path/filepath\"",
		"\t\"sort\"",
		"\t\"strings\"",
		"\t\"time\"",
		"\t\"unicode\"",
		"",
		"\t\"agd/internal/agd\"",
		"\t\"agd/internal/agdtemplates\"",
		")",
		"",
	}, kitBlock...)

	newMain := make([]string, 0, len(lines)-len(pathBlock)-len(wizardBlock)-len(kitBlock))
	for i, line := range lines {
		drop := (i >= idxPathStart && i < idxWizardStart) ||
			(i >= idxWizardStart && i < idxFindStart) ||
			(i >= idxKitStart && i < idxAddSectionStart)
		if !drop {
			newMain = append(newMain, line)
		}
	}

	write("cmd/agd/main.go", newMain)
	write("cmd/agd/path_resolver.go", pathFile)
	write("cmd/agd/wizard.go", wizardFile)
	write("cmd/agd/kit_role_graph.go", kitFile)
	fmt.Println("split complete")
}
