// readme_generation_test.go verifies docs root scaffolding does not auto-create README files.
// readme_generation_test.go는 docs 루트 스캐폴드 시 README 자동 생성을 하지 않는지 검증합니다.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDocsRootDoesNotGenerateReadme(t *testing.T) {
	oldDocsRoot := docsRootDir
	docsRootDir = filepath.Join(t.TempDir(), "agd_docs")
	t.Cleanup(func() { docsRootDir = oldDocsRoot })

	if err := ensureDocsRoot(); err != nil {
		t.Fatalf("ensureDocsRoot failed: %v", err)
	}

	readmePath := filepath.Join(docsRootDir, "README.md")
	if _, err := os.Stat(readmePath); !os.IsNotExist(err) {
		t.Fatalf("README must not be auto-generated at docs root: %s", readmePath)
	}

	if info, err := os.Stat(filepath.Join(docsRootDir, "10_source", "product")); err != nil || !info.IsDir() {
		t.Fatalf("docs scaffold directory should exist: %v", err)
	}
}

func TestEnsureKitRootDoesNotGenerateReadme(t *testing.T) {
	oldDocsRoot := docsRootDir
	docsRootDir = filepath.Join(t.TempDir(), "default_docs_root")
	t.Cleanup(func() { docsRootDir = oldDocsRoot })

	kitRoot := filepath.Join(t.TempDir(), "kit_docs_root")
	if err := ensureKitRoot(kitRoot, "ko"); err != nil {
		t.Fatalf("ensureKitRoot failed: %v", err)
	}

	readmePath := filepath.Join(kitRoot, "README.md")
	if _, err := os.Stat(readmePath); !os.IsNotExist(err) {
		t.Fatalf("README must not be auto-generated at kit root: %s", readmePath)
	}

	if info, err := os.Stat(filepath.Join(kitRoot, "10_source", "service")); err != nil || !info.IsDir() {
		t.Fatalf("kit scaffold directory should exist: %v", err)
	}
}
