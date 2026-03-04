// path_defaults.go resolves runtime defaults for docs root and display hints.
// path_defaults.go는 문서 루트 기본값과 표시용 힌트를 런타임에 계산합니다.
package main

import (
	"os"
	"path/filepath"
	"strings"
)

var docsRootDir = detectDefaultDocsRootDir()

func detectDefaultDocsRootDir() string {
	if fromExe := detectDocsRootFromExecutable(); fromExe != "" {
		return filepath.Clean(fromExe)
	}

	// Prefer package folders when running from project root.
	candidates := []string{
		filepath.Join("00_agd", docsRootLeafDir),
		filepath.Join("agd", docsRootLeafDir),
		docsRootLeafDir,
	}
	for _, candidate := range candidates {
		if isDir(candidate) {
			return filepath.Clean(candidate)
		}
	}

	// Even if docs root is not created yet, keep 00_agd as the default target.
	if isDir("00_agd") || pathExists("00_agd") {
		return filepath.Clean(filepath.Join("00_agd", docsRootLeafDir))
	}
	if isDir("agd") || pathExists("agd") {
		return filepath.Clean(filepath.Join("agd", docsRootLeafDir))
	}
	return filepath.Clean(docsRootLeafDir)
}

func detectDocsRootFromExecutable() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	execDir := filepath.Clean(filepath.Dir(execPath))
	execBase := strings.ToLower(strings.TrimSpace(filepath.Base(execDir)))

	if execBase == "00_agd" || execBase == "agd" {
		return filepath.Join(execDir, docsRootLeafDir)
	}

	// If executable is outside package dir, still respect nearby packaged layout.
	parent := execDir
	nearby := []string{
		filepath.Join(parent, "00_agd", docsRootLeafDir),
		filepath.Join(parent, "agd", docsRootLeafDir),
	}
	for _, candidate := range nearby {
		if isDir(candidate) {
			return filepath.Clean(candidate)
		}
	}
	return ""
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func docsRootDisplayPath() string {
	root := filepath.Clean(docsRootDir)
	rel, err := filepath.Rel(".", root)
	if err != nil {
		return filepath.ToSlash(root)
	}
	rel = filepath.Clean(rel)
	if rel == "." || rel == "" || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return filepath.ToSlash(root)
	}
	return filepath.ToSlash(rel)
}

func docsRootInputPrompt() string {
	return fmtPathHint(text(
		"Root folder (Enter for %s)",
		"루트 폴더 (기본값: %s)",
	))
}

func docsRootInputPromptShort() string {
	return fmtPathHint(text(
		"Root (Enter for %s)",
		"루트 (기본값: %s)",
	))
}

func docsRootSectionScanHint() string {
	service := filepath.ToSlash(filepath.Join(docsRootDisplayPath(), "10_source", "service"))
	frontend := filepath.ToSlash(filepath.Join(docsRootDisplayPath(), "20_derived", "frontend"))
	return text(
		"no sections found under "+service+" or "+frontend,
		service+" 또는 "+frontend+" 경로에서 섹션을 찾지 못했습니다",
	)
}

func fmtPathHint(format string) string {
	return strings.Replace(format, "%s", docsRootDisplayPath(), 1)
}
