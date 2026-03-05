// doc_types.go defines document type catalogs and creation allowlists.
// doc_types.go는 문서 타입 카탈로그와 생성 허용 목록을 정의합니다.
package main

import (
	"fmt"
	"io"
	"strings"
)

type docTypeOption struct {
	Key string
	EN  string
	KO  string
}

func docTypeOptions() []docTypeOption {
	return []docTypeOption{
		{Key: "roadmap", EN: "Quarter/half roadmap plan", KO: "분기/반기 로드맵 계획"},
		{Key: "handoff", EN: "Cross-team handoff document", KO: "팀 간 인수인계 문서"},
		{Key: "meeting", EN: "Meeting notes and decisions", KO: "회의 기록 및 의사결정"},
		{Key: "policy", EN: "Policy and guideline document", KO: "정책 및 가이드 문서"},
		{Key: "experiment", EN: "Experiment and A/B test plan", KO: "실험 및 A/B 테스트 계획"},
		{Key: "service", EN: "Service logic source spec", KO: "서비스 로직 source 스펙"},
		{Key: "core-spec", EN: "Unified core planning spec", KO: "통합 코어 기획 스펙"},
		{Key: "delivery-plan", EN: "Unified frontend delivery plan", KO: "통합 프런트엔드 전달 계획"},
	}
}

func printDocTypeCatalog(w io.Writer, withNumber bool) {
	options := docTypeOptions()
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

func allowedCreateDocTypeKeys() []string {
	return []string{
		"service",
		"core-spec",
		"delivery-plan",
		"meeting",
		"experiment",
		"roadmap",
		"handoff",
		"policy",
	}
}

func allowedCreateDocTypeOptions() []docTypeOption {
	base := docTypeOptions()
	byKey := make(map[string]docTypeOption, len(base))
	for _, opt := range base {
		byKey[opt.Key] = opt
	}

	keys := allowedCreateDocTypeKeys()
	out := make([]docTypeOption, 0, len(keys))
	for _, key := range keys {
		if opt, ok := byKey[key]; ok {
			out = append(out, opt)
		}
	}
	return out
}

func printAllowedCreateDocTypeCatalog(w io.Writer, withNumber bool) {
	options := allowedCreateDocTypeOptions()
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

func isAllowedCreateDocType(kind string) bool {
	target := normalizeDocTypeKey(kind)
	for _, key := range allowedCreateDocTypeKeys() {
		if target == key {
			return true
		}
	}
	return false
}

func allowedCLITemplates() []string {
	keys := allowedCreateDocTypeKeys()
	out := make([]string, 0, len(keys)*2)
	seen := make(map[string]struct{}, len(keys)*2)
	for _, key := range keys {
		for _, lang := range []string{"ko", "en"} {
			templateName, ok := easyTypeToTemplate(key + "-" + lang)
			if !ok {
				continue
			}
			if _, exists := seen[templateName]; exists {
				continue
			}
			seen[templateName] = struct{}{}
			out = append(out, templateName)
		}
	}
	return out
}

func isAllowedCLITemplate(name string) bool {
	target := strings.ToLower(strings.TrimSpace(name))
	for _, templateName := range allowedCLITemplates() {
		if target == templateName {
			return true
		}
	}
	return false
}
