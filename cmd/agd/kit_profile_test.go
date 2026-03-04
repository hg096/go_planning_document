package main

import "testing"

func TestNormalizeKitProfile(t *testing.T) {
	cases := map[string]string{
		"starter-kit":        "starter-kit",
		"bridge-lite":        "bridge-lite",
		"change-flow":        "change-flow",
		"incident-lifecycle": "incident-lifecycle",
		"quality-gate":       "quality-gate",
		"bridge":             "bridge-lite",
		"minimal":            "bridge-lite",
		"maintenance":        "change-flow",
		"new-project":        "change-flow",
		"incident-response":  "incident-lifecycle",
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
}
