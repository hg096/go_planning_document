// encoding_guard_test.go checks core CLI sources for UTF-8 and mojibake issues.
// encoding_guard_test.go는 핵심 CLI 소스의 UTF-8/깨짐 문자를 검증합니다.
package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestNoMojibakeInCoreCLI(t *testing.T) {
	files, err := collectEncodingGuardFiles()
	if err != nil {
		t.Fatalf("collect files failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no files found for encoding guard")
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read failed (%s): %v", file, err)
		}
		if !utf8.Valid(data) {
			t.Fatalf("invalid utf-8 bytes in %s", file)
		}

		text := string(data)
		if strings.ContainsRune(text, unicode.ReplacementChar) {
			t.Fatalf("replacement char found in %s", file)
		}

		for _, r := range text {
			if unicode.Is(unicode.Han, r) {
				t.Fatalf("han rune %q found in %s (possible mojibake)", r, file)
			}
		}
	}
}

func collectEncodingGuardFiles() ([]string, error) {
	var files []string

	roots := []struct {
		dir  string
		exts map[string]struct{}
	}{
		{
			dir: ".",
			exts: map[string]struct{}{
				".go": {},
			},
		},
		{
			dir: filepath.Join("..", "..", "internal", "agdtemplates", "files"),
			exts: map[string]struct{}{
				".agd": {},
			},
		},
		{
			dir: filepath.Join("..", "..", "docs"),
			exts: map[string]struct{}{
				".md": {},
			},
		},
		{
			dir: filepath.Join("..", "..", "examples"),
			exts: map[string]struct{}{
				".agd": {},
				".md":  {},
			},
		},
	}

	for _, root := range roots {
		if _, err := os.Stat(root.dir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		err := filepath.WalkDir(root.dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if _, ok := root.exts[strings.ToLower(filepath.Ext(path))]; !ok {
				return nil
			}
			files = append(files, filepath.Clean(path))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i]) < strings.ToLower(files[j])
	})
	return files, nil
}
