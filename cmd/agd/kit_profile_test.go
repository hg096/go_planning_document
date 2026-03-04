package main

import "testing"

func TestNormalizeKitProfile(t *testing.T) {
	cases := map[string]string{
		"starter-kit": "starter-kit",
		"new-project": "new-project",
		"maintenance": "maintenance",
		"incident":    "incident",
	}

	for input, want := range cases {
		got, ok := normalizeKitProfile(input)
		if !ok {
			t.Fatalf("expected profile %q to be valid", input)
		}
		if got != want {
			t.Fatalf("unexpected profile normalization for %q: got=%q want=%q", input, got, want)
		}
	}

	if got, ok := normalizeKitProfile("not-a-profile"); ok || got != "" {
		t.Fatalf("invalid profile should fail normalization, got=%q ok=%v", got, ok)
	}

	removedAliases := []string{
		"starter",
		"change-flow",
		"incident-lifecycle",
		"incident-response",
		"maintenance-kit",
		"new-project-kit",
		"incident-kit",
		"bridge-lite",
		"quality-gate",
	}
	for _, alias := range removedAliases {
		if got, ok := normalizeKitProfile(alias); ok || got != "" {
			t.Fatalf("alias should be removed: %s got=%q ok=%v", alias, got, ok)
		}
	}
}

func TestNormalizeKitProfileAcceptsWhitespaceAndUnderscore(t *testing.T) {
	cases := map[string]string{
		" starter-kit ": "starter-kit",
		"new_project":   "new-project",
		"maintenance ":  "maintenance",
		" INCIDENT ":    "incident",
	}
	for in, want := range cases {
		got, ok := normalizeKitProfile(in)
		if !ok || got != want {
			t.Fatalf("normalization mismatch input=%q got=%q ok=%v want=%q", in, got, ok, want)
		}
	}
}
