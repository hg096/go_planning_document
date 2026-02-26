// i18n_test.go verifies localization fallback and language selection behavior.
// i18n_test.go는 현지화 fallback과 언어 선택 동작을 검증합니다.
package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestTextFallbackForBrokenKorean(t *testing.T) {
	t.Setenv("AGD_LANG", "ko")

	got := text("Select", strings.Repeat("?", 8))
	if got != koFallbackText("Select") {
		t.Fatalf("expected Korean fallback, got %q", got)
	}
}

func TestTextFallbackForEnglishMirror(t *testing.T) {
	t.Setenv("AGD_LANG", "ko")

	got := text("AGD Wizard", "AGD Wizard")
	if got != koFallbackText("AGD Wizard") {
		t.Fatalf("expected translated fallback, got %q", got)
	}
}

func TestTextKeepsValidKorean(t *testing.T) {
	t.Setenv("AGD_LANG", "ko")

	ko := "\uC81C\uBAA9"
	got := text("Title", ko)
	if got != ko {
		t.Fatalf("expected valid Korean to pass through, got %q", got)
	}
}

func TestBrokenKOTextSingleQuestionMarkIsNotBroken(t *testing.T) {
	if isBrokenKOText("a?b") {
		t.Fatalf("single question mark should not be treated as broken")
	}
}

func TestBrokenKOTextReplacementRuneIsBroken(t *testing.T) {
	if !isBrokenKOText("x\uFFFDy") {
		t.Fatalf("replacement rune should be treated as broken")
	}
}

func TestDocsScaffoldReadmeByLangKOHasFallback(t *testing.T) {
	got := docsScaffoldReadmeByLang("ko")
	if strings.TrimSpace(got) == "" {
		t.Fatalf("ko readme should not be empty")
	}
}

func TestActiveLangUsesBuildDefaultWhenEnvUnset(t *testing.T) {
	old := defaultLang
	defer func() { defaultLang = old }()

	t.Setenv("AGD_LANG", "")
	defaultLang = "ko"
	if got := activeLang(); got != "ko" {
		t.Fatalf("expected ko from defaultLang, got %q", got)
	}

	defaultLang = "en"
	if got := activeLang(); got != "en" {
		t.Fatalf("expected en from defaultLang, got %q", got)
	}
}

func TestUsageHasNoBrokenKorean(t *testing.T) {
	t.Setenv("AGD_LANG", "ko")
	got := usage()
	if strings.Contains(got, "??") {
		t.Fatalf("usage contains broken Korean token")
	}
}

func TestHelpHasNoBrokenKorean(t *testing.T) {
	t.Setenv("AGD_LANG", "ko")

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe create failed: %v", err)
	}
	os.Stdout = w
	code := commandHelp(nil)
	_ = w.Close()
	os.Stdout = oldStdout

	body, _ := io.ReadAll(r)
	_ = r.Close()

	if code != 0 {
		t.Fatalf("help returned non-zero: %d", code)
	}
	if strings.Contains(string(body), "??") {
		t.Fatalf("help output contains broken Korean token")
	}
}
