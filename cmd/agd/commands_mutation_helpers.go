// commands_mutation_helpers.go provides shared helpers for mutation command handlers.
// commands_mutation_helpers.go는 변경 명령에서 공통으로 쓰는 보조 함수를 제공합니다.
package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"agd/internal/agd"
)

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func nextChangeID(doc *agd.Document) string {
	date := time.Now().Format("2006-01-02")
	prefix := "CHG-" + date + "-"
	maxSeq := 0

	for _, change := range doc.Changes {
		id := strings.TrimSpace(change.ID)
		if !strings.HasPrefix(strings.ToUpper(id), strings.ToUpper(prefix)) {
			continue
		}
		suffix := strings.TrimSpace(strings.TrimPrefix(id, prefix))
		n, err := strconv.Atoi(suffix)
		if err != nil {
			continue
		}
		if n > maxSeq {
			maxSeq = n
		}
	}

	for seq := maxSeq + 1; ; seq++ {
		candidate := fmt.Sprintf("%s%02d", prefix, seq)
		if !changeExists(doc, candidate) {
			return candidate
		}
	}
}

func changeExists(doc *agd.Document, id string) bool {
	for _, change := range doc.Changes {
		if strings.EqualFold(strings.TrimSpace(change.ID), strings.TrimSpace(id)) {
			return true
		}
	}
	return false
}
