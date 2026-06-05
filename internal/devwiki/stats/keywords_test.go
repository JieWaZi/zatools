package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpdateKeywordsAggregatesFromQueriesLogs(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	now := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	logPath := filepath.Join(statsDir, dailyLogName(now))
	file, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	encoder := json.NewEncoder(file)
	for _, event := range []Event{
		{Timestamp: now, Endpoint: "search", Kind: "index", Queries: []string{"VIP 脑裂"}, ResultCount: 1},
		{Timestamp: now.Add(time.Minute), Endpoint: "search", Kind: "topic", Queries: []string{"VIP", "脑裂"}, ResultCount: 2},
		{Timestamp: now.Add(2 * time.Minute), Endpoint: "read", Kind: "topic", Slug: "vip", View: "card"},
	} {
		if err := encoder.Encode(event); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}
	file.Close()

	result, err := UpdateKeywords(UpdateKeywordsOptions{Root: root, Now: now.Add(3 * time.Minute)})
	if err != nil {
		t.Fatalf("UpdateKeywords() error = %v", err)
	}
	if len(result.Keywords) == 0 {
		t.Fatal("expected keywords")
	}
	if result.AddedTerms == 0 {
		t.Fatalf("added terms = %d, want > 0", result.AddedTerms)
	}

	data, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var loaded keywordsFile
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(loaded.Keywords) > maxKeywords {
		t.Fatalf("keywords len = %d, want <= %d", len(loaded.Keywords), maxKeywords)
	}

	second, err := UpdateKeywords(UpdateKeywordsOptions{Root: root, Now: now.Add(4 * time.Minute)})
	if err != nil {
		t.Fatalf("UpdateKeywords(second) error = %v", err)
	}
	if second.AddedTerms != 0 {
		t.Fatalf("second added terms = %d, want 0", second.AddedTerms)
	}
}

func TestUpdateKeywordsIncrementsIncrementally(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	firstAt := time.Date(2026, 6, 5, 9, 0, 0, 0, time.UTC)
	secondAt := firstAt.Add(time.Hour)
	appendEvent(t, statsDir, Event{Timestamp: firstAt, Endpoint: "search", Queries: []string{"alpha"}, ResultCount: 1})

	first, err := UpdateKeywords(UpdateKeywordsOptions{Root: root, Now: firstAt.Add(time.Minute)})
	if err != nil {
		t.Fatalf("UpdateKeywords(first) error = %v", err)
	}
	if first.AddedTerms == 0 {
		t.Fatal("expected first pass to add terms")
	}

	appendEvent(t, statsDir, Event{Timestamp: secondAt, Endpoint: "search", Queries: []string{"alpha", "beta"}, ResultCount: 2})

	second, err := UpdateKeywords(UpdateKeywordsOptions{Root: root, Now: secondAt.Add(time.Minute)})
	if err != nil {
		t.Fatalf("UpdateKeywords(second) error = %v", err)
	}
	if second.AddedTerms == 0 {
		t.Fatal("expected second pass to add new terms")
	}

	weights := map[string]int{}
	for _, item := range second.Keywords {
		weights[item.Text] = item.Weight
	}
	if weights["alpha"] < 2 {
		t.Fatalf("alpha weight = %d, want >= 2", weights["alpha"])
	}
	if weights["beta"] != 1 {
		t.Fatalf("beta weight = %d, want 1", weights["beta"])
	}
}

func appendEvent(t *testing.T, statsDir string, event Event) {
	t.Helper()
	path := filepath.Join(statsDir, dailyLogName(event.Timestamp))
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(event); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}
