// path_resolver.go resolves AGD paths and scans/selects document files.
// path_resolver.go는 AGD 경로 해석과 문서 파일 스캔/선택을 처리합니다.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ensureDocsRoot() error {
	if err := os.MkdirAll(docsRootDir, 0o755); err != nil {
		return err
	}
	for _, rel := range docsScaffoldFolders {
		if err := os.MkdirAll(filepath.Join(docsRootDir, rel), 0o755); err != nil {
			return err
		}
	}
	return writeScaffoldReadmeIfMissing(docsRootDir, activeLang())
}

func writeScaffoldReadmeIfMissing(root, lang string) error {
	readmePath := filepath.Join(root, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	content := docsScaffoldReadmeByLang(lang)
	return os.WriteFile(readmePath, []byte(content), 0o644)
}

func hasPathSeparator(value string) bool {
	return strings.Contains(value, `\`) || strings.Contains(value, `/`)
}

func ensureAGDExt(value string) string {
	if strings.EqualFold(filepath.Ext(value), ".agd") {
		return value
	}
	return value + ".agd"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func normalizeDocTypeKey(kind string) string {
	raw := strings.ToLower(strings.TrimSpace(kind))
	raw = strings.TrimSuffix(raw, "-ko")
	raw = strings.TrimSuffix(raw, "-en")
	switch raw {
	case "service", "system", "service-logic":
		return "service"
	case "prd", "adr", "guide", "ai", "ai-planning-guide", "ai-guide":
		return "core-spec"
	case "qa", "qaplan", "qa-plan", "frontend", "page-logic", "frontend-page", "runbook":
		return "delivery-plan"
	case "minutes":
		return "meeting"
	case "incident", "incident-case", "incident-root", "incident-workspace":
		return "incident-case"
	case "maintenance", "maintenance-case", "maintenance-workspace":
		return "maintenance-case"
	case "core", "core-spec":
		return "core-spec"
	case "delivery", "delivery-plan":
		return "delivery-plan"
	default:
		return raw
	}
}

func defaultSubdirForDocType(kind string) string {
	switch normalizeDocTypeKey(kind) {
	case "service":
		return filepath.Join("10_source", "service")
	case "policy":
		return filepath.Join("10_source", "policy")
	case "incident-case":
		return filepath.Join("30_shared", "errFix")
	case "maintenance-case":
		return filepath.Join("30_shared", "maintenance")
	case "core-spec":
		return filepath.Join("10_source", "product")
	case "delivery-plan":
		return filepath.Join("20_derived", "frontend")
	case "meeting":
		return filepath.Join("30_shared", "meeting")
	case "handoff":
		return filepath.Join("30_shared", "handoff")
	case "roadmap":
		return filepath.Join("30_shared", "roadmap")
	case "experiment":
		return filepath.Join("30_shared", "experiment")
	default:
		return "00_inbox"
	}
}

func resolveNewDocPathByType(kind, out string) (string, error) {
	raw := strings.TrimSpace(out)
	if raw == "" {
		return "", fmt.Errorf("%s", text("path is required", "path is required"))
	}
	if err := ensureDocsRoot(); err != nil {
		return "", err
	}

	if fileExists(raw) || isExplicitPath(raw) || hasPathSeparator(raw) || isUnderDocsRoot(raw) {
		return resolveAGDPathEasy(raw, false)
	}

	baseName := ensureAGDExt(raw)
	matches, err := findByBaseName(docsRootDir, baseName)
	if err != nil {
		return "", err
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf(text(
			"multiple files matched %s: %s (use explicit path)", "multiple files matched %s: %s (use explicit path)",
		), baseName, strings.Join(matches, ", "))
	}

	subdir := defaultSubdirForDocType(kind)
	return filepath.Clean(filepath.Join(docsRootDir, subdir, ensureAGDExt(raw))), nil
}

func isExplicitPath(path string) bool {
	if filepath.IsAbs(path) {
		return true
	}
	return strings.HasPrefix(path, `.\`) ||
		strings.HasPrefix(path, `./`) ||
		strings.HasPrefix(path, `..\`) ||
		strings.HasPrefix(path, `../`)
}

func isUnderDocsRoot(path string) bool {
	cleanDocs := strings.ToLower(filepath.Clean(docsRootDir))
	cleanPath := strings.ToLower(filepath.Clean(path))
	if cleanPath == cleanDocs {
		return true
	}
	return strings.HasPrefix(cleanPath, cleanDocs+string(os.PathSeparator))
}

func listAGDFiles(root string) ([]string, error) {
	out := make([]string, 0)
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".agd") {
			if shouldSkipEndedFlowDoc(path) {
				return nil
			}
			out = append(out, filepath.Clean(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

func shouldSkipEndedFlowDoc(path string) bool {
	base := strings.ToLower(strings.TrimSpace(filepath.Base(path)))
	if !strings.HasPrefix(base, "end__") {
		return false
	}
	clean := strings.ToLower(filepath.ToSlash(filepath.Clean(path)))
	isMaintenanceDoc := strings.Contains(base, "maintenance_case") && strings.Contains(clean, "/30_shared/maintenance/")
	isIncidentDoc := strings.Contains(base, "incident_case") && strings.Contains(clean, "/30_shared/errfix/")
	return isMaintenanceDoc || isIncidentDoc
}

func listSubfolders(root string) ([]string, error) {
	out := make([]string, 0)
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || filepath.Clean(path) == filepath.Clean(root) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.Clean(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

func findByBaseName(root, baseName string) ([]string, error) {
	files, err := listAGDFiles(root)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, path := range files {
		if strings.EqualFold(filepath.Base(path), baseName) {
			out = append(out, path)
		}
	}
	sort.Strings(out)
	return out, nil
}

func resolveAGDPathEasy(input string, mustExist bool) (string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return "", fmt.Errorf("%s", text("path is required", "path is required"))
	}

	if fileExists(raw) {
		return filepath.Clean(raw), nil
	}

	if err := ensureDocsRoot(); err != nil {
		return "", err
	}

	if isExplicitPath(raw) {
		path := ensureAGDExt(raw)
		if mustExist && !fileExists(path) {
			return "", fmt.Errorf(text("file not found: %s", "file not found: %s"), path)
		}
		return filepath.Clean(path), nil
	}

	if !hasPathSeparator(raw) {
		baseName := ensureAGDExt(raw)
		matches, err := findByBaseName(docsRootDir, baseName)
		if err != nil {
			return "", err
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 && mustExist {
			msg := strings.Join(matches, ", ")
			return "", fmt.Errorf(text("multiple files matched %s: %s", "multiple files matched %s: %s"), baseName, msg)
		}
		candidate := filepath.Join(docsRootDir, baseName)
		if mustExist && !fileExists(candidate) {
			return "", fmt.Errorf(text("file not found: %s", "file not found: %s"), candidate)
		}
		return filepath.Clean(candidate), nil
	}

	asIs := ensureAGDExt(raw)
	if isUnderDocsRoot(raw) {
		if mustExist && !fileExists(asIs) {
			return "", fmt.Errorf(text("file not found: %s", "file not found: %s"), asIs)
		}
		return filepath.Clean(asIs), nil
	}
	if fileExists(asIs) {
		return filepath.Clean(asIs), nil
	}

	inDocs := ensureAGDExt(filepath.Join(docsRootDir, raw))
	if mustExist {
		if fileExists(inDocs) {
			return filepath.Clean(inDocs), nil
		}
		return "", fmt.Errorf(text("file not found: %s", "file not found: %s"), inDocs)
	}

	return filepath.Clean(inDocs), nil
}
