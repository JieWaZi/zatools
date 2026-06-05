package stats

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecorderRecordSearchAppendsDailyJSONL(t *testing.T) {
	root := t.TempDir()
	fixed := time.Date(2026, 6, 5, 10, 30, 0, 0, time.FixedZone("CST", 8*3600))
	rec := NewRecorder(root)
	rec.now = func() time.Time { return fixed }

	rec.RecordSearch("workflow", []string{"防脑裂"}, []SearchHit{{Slug: "workflow-ha", Score: "100%"}}, 1)

	deadline := time.After(2 * time.Second)
	for {
		path := filepath.Join(root, DirName, "queries-2026-06-05.jsonl")
		if _, err := os.Stat(path); err == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for daily JSONL")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	events, err := ReadEvents(filepath.Join(root, DirName, "queries-2026-06-05.jsonl"))
	if err != nil {
		t.Fatalf("ReadEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].Endpoint != "search" || events[0].ResultCount != 1 || events[0].Empty {
		t.Fatalf("event = %#v", events[0])
	}

	rec.mu.Lock()
	if rec.todaySearch != 1 || rec.totalSearch != 1 {
		t.Fatalf("counts = today %d total %d, want 1/1", rec.todaySearch, rec.totalSearch)
	}
	rec.mu.Unlock()
}

func TestRecorderRecordReadIncrementsMemoryAndFlush(t *testing.T) {
	root := t.TempDir()
	fixed := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)
	rec := NewRecorder(root)
	rec.now = func() time.Time { return fixed }

	rec.RecordRead("topic", "vip", "card")
	rec.RecordRead("topic", "vip", "core")
	rec.mu.Lock()
	if rec.todayRead != 2 || rec.totalRead != 2 {
		t.Fatalf("read counts = today %d total %d, want 2/2", rec.todayRead, rec.totalRead)
	}
	rec.mu.Unlock()

	rec.Flush()

	data, err := os.ReadFile(filepath.Join(root, DirName, "summary.json"))
	if err != nil {
		t.Fatalf("ReadFile(summary.json) error = %v", err)
	}
	var summary summaryFile
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatalf("Unmarshal(summary) error = %v", err)
	}
	if summary.TodayReadCount != 2 || summary.TotalReadCount != 2 {
		t.Fatalf("summary read counts = %#v", summary)
	}

	data, err = os.ReadFile(filepath.Join(root, DirName, "documents.json"))
	if err != nil {
		t.Fatalf("ReadFile(documents.json) error = %v", err)
	}
	var loaded documentsFile
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if loaded.Documents["vip"].ReadCount != 2 || loaded.Documents["vip"].Kind != "topic" {
		t.Fatalf("documents = %#v", loaded.Documents)
	}
}

func TestRecorderFlushWritesSummary(t *testing.T) {
	root := t.TempDir()
	fixed := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	rec := NewRecorder(root)
	rec.now = func() time.Time { return fixed }

	rec.RecordSearch("index", []string{"VIP"}, nil, 0)
	rec.Flush()

	data, err := os.ReadFile(filepath.Join(root, DirName, "summary.json"))
	if err != nil {
		t.Fatalf("ReadFile(summary.json) error = %v", err)
	}
	var loaded summaryFile
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if loaded.TodaySearchCount != 1 || loaded.TotalSearchCount != 1 || loaded.Today != "2026-06-05" {
		t.Fatalf("summary = %#v", loaded)
	}
}

func TestRecorderLoadsExistingSummaryOnStart(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	existing := summaryFile{
		UpdatedAt:        time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC),
		Today:            "2026-06-04",
		TodaySearchCount: 9,
		TotalSearchCount: 99,
	}
	payload, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(statsDir, "summary.json"), payload, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	fixed := time.Date(2026, 6, 5, 8, 0, 0, 0, time.UTC)
	rec := NewRecorder(root)
	rec.now = func() time.Time { return fixed }
	rec.RecordSearch("topic", []string{"x"}, nil, 0)

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.totalSearch != 100 {
		t.Fatalf("totalSearch = %d, want 100", rec.totalSearch)
	}
	if rec.todaySearch != 1 {
		t.Fatalf("todaySearch = %d, want 1", rec.todaySearch)
	}
}

func TestRecorderStartUpdatesKeywordsImmediatelyAndHourly(t *testing.T) {
	if DefaultKeywordUpdateInterval != time.Hour {
		t.Fatalf("DefaultKeywordUpdateInterval = %s, want 1h", DefaultKeywordUpdateInterval)
	}

	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	fixed := time.Date(2026, 6, 5, 13, 0, 0, 0, time.UTC)
	appendEvent(t, statsDir, Event{
		Timestamp:   fixed,
		Endpoint:    "search",
		Kind:        "topic",
		Queries:     []string{"VIP 脑裂"},
		ResultCount: 1,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rec := NewRecorder(root)
	rec.now = func() time.Time { return fixed.Add(time.Minute) }
	rec.keywordInterval = time.Hour
	rec.Start(ctx)

	keywordsPath := filepath.Join(statsDir, "keywords.json")
	deadline := time.After(2 * time.Second)
	for {
		data, err := os.ReadFile(keywordsPath)
		if err == nil && len(data) > 0 {
			var loaded keywordsFile
			if err := json.Unmarshal(data, &loaded); err != nil {
				t.Fatalf("Unmarshal(keywords.json) error = %v", err)
			}
			if len(loaded.Keywords) == 0 {
				t.Fatalf("keywords = %#v, want non-empty", loaded.Keywords)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for automatic keywords.json")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestTopDocumentsSortsByReadCount(t *testing.T) {
	docs := map[string]DocumentEntry{
		"b": {Kind: "topic", ReadCount: 2},
		"a": {Kind: "workflow", ReadCount: 5},
		"c": {Kind: "topic", ReadCount: 5},
	}
	top := TopDocuments(docs, 2)
	if len(top) != 2 {
		t.Fatalf("top len = %d, want 2", len(top))
	}
	if top[0].Slug != "a" || top[1].Slug != "c" {
		t.Fatalf("top order = %#v", top)
	}
}
