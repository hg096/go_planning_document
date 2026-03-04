package agdtemplates

import (
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"
)

//go:embed files/*.agd
var templateFS embed.FS

type RenderOptions struct {
	Title   string
	Owner   string
	Version string
	Date    string
}

func Names() []string {
	names := []string{
		"roadmap-ko",
		"roadmap-en",
		"handoff-ko",
		"handoff-en",
		"meeting-ko",
		"meeting-en",
		"policy-ko",
		"policy-en",
		"experiment-ko",
		"experiment-en",
		"incident-case-ko",
		"incident-case-en",
		"maintenance-case-ko",
		"maintenance-case-en",
		"core-spec-ko",
		"core-spec-en",
		"delivery-plan-ko",
		"delivery-plan-en",
	}
	sort.Strings(names)
	return names
}

func Load(name string) (string, error) {
	path, err := nameToPath(name)
	if err != nil {
		return "", err
	}
	raw, err := templateFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func Render(name string, opt RenderOptions) (string, error) {
	raw, err := Load(name)
	if err != nil {
		return "", err
	}

	title := strings.TrimSpace(opt.Title)
	if title == "" {
		title = defaultTitle(name)
	}
	owner := strings.TrimSpace(opt.Owner)
	version := strings.TrimSpace(opt.Version)
	if version == "" {
		version = "0.1"
	}
	date := strings.TrimSpace(opt.Date)
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	replacer := strings.NewReplacer(
		"{{TITLE}}", title,
		"{{OWNER}}", owner,
		"{{VERSION}}", version,
		"{{DATE}}", date,
	)
	return replacer.Replace(raw), nil
}

func nameToPath(name string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "roadmap-ko":
		return "files/roadmap_ko.agd", nil
	case "roadmap-en":
		return "files/roadmap_en.agd", nil
	case "handoff-ko":
		return "files/handoff_ko.agd", nil
	case "handoff-en":
		return "files/handoff_en.agd", nil
	case "meeting-ko":
		return "files/meeting_ko.agd", nil
	case "meeting-en":
		return "files/meeting_en.agd", nil
	case "policy-ko":
		return "files/policy_ko.agd", nil
	case "policy-en":
		return "files/policy_en.agd", nil
	case "experiment-ko":
		return "files/experiment_ko.agd", nil
	case "experiment-en":
		return "files/experiment_en.agd", nil
	case "incident-case-ko":
		return "files/incident_case_ko.agd", nil
	case "incident-case-en":
		return "files/incident_case_en.agd", nil
	case "maintenance-case-ko":
		return "files/maintenance_case_ko.agd", nil
	case "maintenance-case-en":
		return "files/maintenance_case_en.agd", nil
	case "core-spec-ko":
		return "files/core_spec_ko.agd", nil
	case "core-spec-en":
		return "files/core_spec_en.agd", nil
	case "delivery-plan-ko":
		return "files/delivery_plan_ko.agd", nil
	case "delivery-plan-en":
		return "files/delivery_plan_en.agd", nil
	default:
		return "", fmt.Errorf("unknown template %q", name)
	}
}

func defaultTitle(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "roadmap-ko", "roadmap-en":
		return "New Roadmap Draft"
	case "handoff-ko", "handoff-en":
		return "New Handoff Draft"
	case "meeting-ko", "meeting-en":
		return "New Meeting Notes Draft"
	case "policy-ko", "policy-en":
		return "New Policy Draft"
	case "experiment-ko", "experiment-en":
		return "New Experiment Plan Draft"
	case "incident-case-ko", "incident-case-en":
		return "New Incident Root Trace Case"
	case "maintenance-case-ko", "maintenance-case-en":
		return "New Maintenance Unified Case"
	case "core-spec-ko", "core-spec-en":
		return "New Core Spec"
	case "delivery-plan-ko", "delivery-plan-en":
		return "New Delivery Plan"
	default:
		return "New AGD Document"
	}
}
