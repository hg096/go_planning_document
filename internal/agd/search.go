package agd

import "strings"

func Search(doc *Document, query string) []SearchHit {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return []SearchHit{}
	}

	hits := make([]SearchHit, 0)

	for _, row := range doc.Map {
		if containsCI(row.ID, q) {
			hits = append(hits, SearchHit{Kind: "map", ID: row.ID, Field: "id", Snippet: row.ID})
		}
		if containsCI(row.Title, q) {
			hits = append(hits, SearchHit{Kind: "map", ID: row.ID, Field: "title", Snippet: row.Title})
		}
		if containsCI(row.Summary, q) {
			hits = append(hits, SearchHit{Kind: "map", ID: row.ID, Field: "summary", Snippet: row.Summary})
		}
	}

	for _, section := range doc.Sections {
		if containsCI(section.ID, q) {
			hits = append(hits, SearchHit{Kind: "section", ID: section.ID, Field: "id", Snippet: section.ID})
		}
		if containsCI(section.Path, q) {
			hits = append(hits, SearchHit{Kind: "section", ID: section.ID, Field: "path", Snippet: section.Path})
		}
		if containsCI(section.Title, q) {
			hits = append(hits, SearchHit{Kind: "section", ID: section.ID, Field: "title", Snippet: section.Title})
		}
		if containsCI(section.Summary, q) {
			hits = append(hits, SearchHit{Kind: "section", ID: section.ID, Field: "summary", Snippet: section.Summary})
		}
		if containsCI(strings.Join(section.Links, ", "), q) {
			hits = append(hits, SearchHit{Kind: "section", ID: section.ID, Field: "links", Snippet: strings.Join(section.Links, ", ")})
		}
		if containsCI(section.Content, q) {
			hits = append(hits, SearchHit{
				Kind:    "section",
				ID:      section.ID,
				Field:   "content",
				Snippet: snippetFor(section.Content, q, 96),
			})
		}
	}

	for _, change := range doc.Changes {
		if containsCI(change.ID, q) {
			hits = append(hits, SearchHit{Kind: "change", ID: change.ID, Field: "id", Snippet: change.ID})
		}
		if containsCI(change.Target, q) {
			hits = append(hits, SearchHit{Kind: "change", ID: change.ID, Field: "target", Snippet: change.Target})
		}
		if containsCI(change.Reason, q) {
			hits = append(hits, SearchHit{Kind: "change", ID: change.ID, Field: "reason", Snippet: change.Reason})
		}
		if containsCI(change.Impact, q) {
			hits = append(hits, SearchHit{Kind: "change", ID: change.ID, Field: "impact", Snippet: change.Impact})
		}
	}

	return hits
}

func containsCI(value string, loweredNeedle string) bool {
	return strings.Contains(strings.ToLower(value), loweredNeedle)
}

func snippetFor(value string, loweredNeedle string, window int) string {
	lowered := strings.ToLower(value)
	idx := strings.Index(lowered, loweredNeedle)
	if idx < 0 {
		return value
	}

	left := idx - window/2
	if left < 0 {
		left = 0
	}
	right := left + window
	if right > len(value) {
		right = len(value)
	}
	snippet := strings.TrimSpace(value[left:right])
	if left > 0 {
		snippet = "..." + snippet
	}
	if right < len(value) {
		snippet = snippet + "..."
	}
	return snippet
}
