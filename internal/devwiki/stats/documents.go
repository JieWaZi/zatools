package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func loadDocumentsFile(path string) map[string]DocumentEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]DocumentEntry{}
	}
	var loaded documentsFile
	if json.Unmarshal(data, &loaded) != nil || loaded.Documents == nil {
		return map[string]DocumentEntry{}
	}
	return loaded.Documents
}

// DocumentCountsFromLogs rebuilds read counts by replaying read events in queries-*.jsonl.
func DocumentCountsFromLogs(statsDir string) (map[string]DocumentEntry, error) {
	matches, err := filepath.Glob(filepath.Join(statsDir, "queries-*.jsonl"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	documents := map[string]DocumentEntry{}
	for _, path := range matches {
		events, err := ReadEvents(path)
		if err != nil {
			return nil, err
		}
		for _, event := range events {
			if event.Endpoint != "read" {
				continue
			}
			slug := strings.TrimSpace(event.Slug)
			if slug == "" {
				continue
			}
			entry := documents[slug]
			if event.Kind != "" {
				entry.Kind = event.Kind
			}
			entry.ReadCount++
			documents[slug] = entry
		}
	}
	return documents, nil
}

func mergeDocumentCounts(base, extra map[string]DocumentEntry) map[string]DocumentEntry {
	if len(base) == 0 && len(extra) == 0 {
		return map[string]DocumentEntry{}
	}
	out := cloneDocuments(base)
	for slug, entry := range extra {
		current := out[slug]
		if entry.ReadCount > current.ReadCount {
			current.ReadCount = entry.ReadCount
		}
		if entry.Kind != "" {
			current.Kind = entry.Kind
		}
		out[slug] = current
	}
	return out
}
