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
		"adr-en",
		"prd-ko",
		"prd-en",
		"runbook-ko",
		"runbook-en",
		"postmortem-ko",
		"postmortem-en",
		"roadmap-ko",
		"roadmap-en",
		"handoff-ko",
		"handoff-en",
		"meeting-ko",
		"meeting-en",
		"adr-ko",
		"policy-ko",
		"policy-en",
		"experiment-ko",
		"experiment-en",
		"qa-plan-ko",
		"qa-plan-en",
		"service-logic-ko",
		"service-logic-en",
		"frontend-page-ko",
		"frontend-page-en",
		"ai-guide-ko",
		"ai-guide-en",
		"incident-case-ko",
		"incident-case-en",
		"maintenance-case-ko",
		"maintenance-case-en",
		"core-spec-ko",
		"core-spec-en",
		"delivery-plan-ko",
		"delivery-plan-en",
		"change-log-ko",
		"change-log-en",
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
	case "prd-ko":
		return "files/prd_ko.agd", nil
	case "prd-en":
		return "files/prd_en.agd", nil
	case "runbook-ko":
		return "files/runbook_ko.agd", nil
	case "runbook-en":
		return "files/runbook_en.agd", nil
	case "postmortem-ko":
		return "files/postmortem_ko.agd", nil
	case "postmortem-en":
		return "files/postmortem_en.agd", nil
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
	case "adr-ko":
		return "files/adr_ko.agd", nil
	case "adr-en":
		return "files/adr_en.agd", nil
	case "policy-ko":
		return "files/policy_ko.agd", nil
	case "policy-en":
		return "files/policy_en.agd", nil
	case "experiment-ko":
		return "files/experiment_ko.agd", nil
	case "experiment-en":
		return "files/experiment_en.agd", nil
	case "qa-plan-ko":
		return "files/qa_plan_ko.agd", nil
	case "qa-plan-en":
		return "files/qa_plan_en.agd", nil
	case "service-logic-ko":
		return "files/service_logic_ko.agd", nil
	case "service-logic-en":
		return "files/service_logic_en.agd", nil
	case "frontend-page-ko":
		return "files/frontend_page_ko.agd", nil
	case "frontend-page-en":
		return "files/frontend_page_en.agd", nil
	case "ai-guide-ko":
		return "files/ai_planning_guide_ko.agd", nil
	case "ai-guide-en":
		return "files/ai_planning_guide_en.agd", nil
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
	case "change-log-ko":
		return "files/change_log_ko.agd", nil
	case "change-log-en":
		return "files/change_log_en.agd", nil
	default:
		return "", fmt.Errorf("unknown template %q", name)
	}
}

func defaultTitle(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "prd-ko", "prd-en":
		return "New PRD Draft"
	case "runbook-ko", "runbook-en":
		return "New Runbook Draft"
	case "postmortem-ko", "postmortem-en":
		return "New Postmortem Draft"
	case "roadmap-ko", "roadmap-en":
		return "New Roadmap Draft"
	case "handoff-ko", "handoff-en":
		return "New Handoff Draft"
	case "meeting-ko", "meeting-en":
		return "New Meeting Notes Draft"
	case "adr-ko", "adr-en":
		return "New ADR Draft"
	case "policy-ko", "policy-en":
		return "New Policy Draft"
	case "experiment-ko", "experiment-en":
		return "New Experiment Plan Draft"
	case "qa-plan-ko", "qa-plan-en":
		return "New QA Plan Draft"
	case "service-logic-ko", "service-logic-en":
		return "New Service Logic Draft"
	case "frontend-page-ko", "frontend-page-en":
		return "New Frontend Page Logic Draft"
	case "ai-guide-ko", "ai-guide-en":
		return "New AI Planning Guide Draft"
	case "incident-case-ko", "incident-case-en":
		return "New Incident Root Trace Case"
	case "maintenance-case-ko", "maintenance-case-en":
		return "New Maintenance Unified Case"
	case "core-spec-ko", "core-spec-en":
		return "New Core Spec"
	case "delivery-plan-ko", "delivery-plan-en":
		return "New Delivery Plan"
	case "change-log-ko", "change-log-en":
		return "New Change Log"
	default:
		return "New AGD Document"
	}
}
