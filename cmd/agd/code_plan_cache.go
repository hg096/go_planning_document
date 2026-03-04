// code_plan_cache.go stores cache snapshots for code-plan validation.
// code_plan_cache.go는 코드 기획 검증용 캐시 스냅샷을 저장합니다.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const codePlanCacheVersion = 1

type codePlanCache struct {
	Version      int                             `json:"version"`
	UpdatedAt    string                          `json:"updated_at"`
	RepoRoot     string                          `json:"repo_root"`
	FileSnapshot map[string]codePlanFileSnapshot `json:"file_snapshot"`
	DocSnapshot  map[string]codePlanDocSnapshot  `json:"doc_snapshot"`
}

type codePlanFileSnapshot struct {
	Size    int64 `json:"size"`
	ModUnix int64 `json:"mod_unix"`
}

type codePlanDocSnapshot struct {
	ChangeDigest   string `json:"change_digest"`
	LatestChangeID string `json:"latest_change_id"`
	LatestChangeAt string `json:"latest_change_date"`
}

func loadCodePlanCache(path string) (*codePlanCache, error) {
	target := strings.TrimSpace(path)
	if target == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("code-plan cache read failed: %w", err)
	}
	var cache codePlanCache
	if err := json.Unmarshal(raw, &cache); err != nil {
		return nil, fmt.Errorf("code-plan cache parse failed: %w", err)
	}
	normalizeCodePlanCache(&cache)
	return &cache, nil
}

func saveCodePlanCache(path string, cache *codePlanCache) error {
	if cache == nil {
		return fmt.Errorf("code-plan cache is nil")
	}
	target := strings.TrimSpace(path)
	if target == "" {
		return fmt.Errorf("code-plan cache path is empty")
	}
	if cache.Version <= 0 {
		cache.Version = codePlanCacheVersion
	}
	if strings.TrimSpace(cache.UpdatedAt) == "" {
		cache.UpdatedAt = time.Now().Format(time.RFC3339)
	}
	normalizeCodePlanCache(cache)

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("code-plan cache directory create failed: %w", err)
	}

	encoded, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("code-plan cache encode failed: %w", err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(target, encoded, 0o644); err != nil {
		return fmt.Errorf("code-plan cache write failed: %w", err)
	}
	return nil
}

func normalizeCodePlanCache(cache *codePlanCache) {
	if cache == nil {
		return
	}
	if cache.FileSnapshot == nil {
		cache.FileSnapshot = map[string]codePlanFileSnapshot{}
	}
	if cache.DocSnapshot == nil {
		cache.DocSnapshot = map[string]codePlanDocSnapshot{}
	}

	normalizedFiles := make(map[string]codePlanFileSnapshot, len(cache.FileSnapshot))
	for rawPath, snap := range cache.FileSnapshot {
		key := codePlanPathKey(rawPath)
		if key == "" {
			continue
		}
		normalizedFiles[key] = snap
	}
	cache.FileSnapshot = normalizedFiles

	normalizedDocs := make(map[string]codePlanDocSnapshot, len(cache.DocSnapshot))
	for rawPath, snap := range cache.DocSnapshot {
		key := codePlanPathKey(rawPath)
		if key == "" {
			continue
		}
		normalizedDocs[key] = snap
	}
	cache.DocSnapshot = normalizedDocs
}
