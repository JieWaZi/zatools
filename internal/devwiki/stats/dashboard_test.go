package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDashboardAggregatesDerivedMetrics(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatalf("MkdirAll(config) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "project.yaml"), []byte("project_name: Huawei ZDDI\nproject_slug: huawei-zddi\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(project.yaml) error = %v", err)
	}

	now := time.Date(2026, 6, 5, 15, 0, 0, 0, time.UTC)
	writeJSON(t, filepath.Join(statsDir, "summary.json"), summaryFile{
		UpdatedAt:        now,
		Today:            "2026-06-05",
		TodaySearchCount: 3,
		TotalSearchCount: 30,
	})
	writeJSON(t, filepath.Join(statsDir, "documents.json"), documentsFile{
		UpdatedAt: now,
		Documents: map[string]DocumentEntry{
			"vip": {Kind: "topic", ReadCount: 8},
		},
	})

	logPath := filepath.Join(statsDir, dailyLogName(now))
	file, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	encoder := json.NewEncoder(file)
	for _, event := range []Event{
		{Timestamp: now, Endpoint: "search", Kind: "index", Queries: []string{"VIP"}, ResultCount: 1, Empty: false},
		{Timestamp: now, Endpoint: "search", Kind: "topic", Queries: []string{"missing"}, ResultCount: 0, Empty: true},
		{Timestamp: now, Endpoint: "read", Kind: "topic", Slug: "vip", View: "card"},
	} {
		if err := encoder.Encode(event); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}
	file.Close()

	dashboard, err := LoadDashboard(root, now)
	if err != nil {
		t.Fatalf("LoadDashboard() error = %v", err)
	}
	if dashboard.ProjectSlug != "huawei-zddi" || dashboard.ProjectName != "Huawei ZDDI" {
		t.Fatalf("project = %#v", dashboard)
	}
	if dashboard.TodaySearchCount != 2 || dashboard.TodayReadCount != 1 {
		t.Fatalf("today counts = search %d read %d", dashboard.TodaySearchCount, dashboard.TodayReadCount)
	}
	if dashboard.TodayAPICount != 3 {
		t.Fatalf("today api count = %d, want 3", dashboard.TodayAPICount)
	}
	if dashboard.TodayEmptySearchCount != 1 {
		t.Fatalf("empty search count = %d, want 1", dashboard.TodayEmptySearchCount)
	}
	if len(dashboard.TodayEvents) != 3 {
		t.Fatalf("today events = %d, want 3", len(dashboard.TodayEvents))
	}
	if dashboard.TodayLogFile != "queries-2026-06-05.jsonl" {
		t.Fatalf("log file = %q", dashboard.TodayLogFile)
	}
}

func TestLoadDashboardAggregatesFiles(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	now := time.Date(2026, 6, 5, 15, 0, 0, 0, time.UTC)
	writeJSON(t, filepath.Join(statsDir, "summary.json"), summaryFile{
		UpdatedAt:        now,
		Today:            "2026-06-05",
		TodaySearchCount: 3,
		TotalSearchCount: 30,
	})
	writeJSON(t, filepath.Join(statsDir, "documents.json"), documentsFile{
		UpdatedAt: now,
		Documents: map[string]DocumentEntry{
			"vip": {Kind: "topic", ReadCount: 8},
			"wf":  {Kind: "workflow", ReadCount: 2},
		},
	})

	event := Event{
		Timestamp:   now,
		Endpoint:    "search",
		Kind:        "index",
		Queries:     []string{"VIP"},
		ResultCount: 1,
		Empty:       false,
	}
	file, err := os.Create(filepath.Join(statsDir, dailyLogName(now)))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(event); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	file.Close()

	dashboard, err := LoadDashboard(root, now)
	if err != nil {
		t.Fatalf("LoadDashboard() error = %v", err)
	}
	if dashboard.TodaySearchCount != 1 || dashboard.TotalSearchCount != 30 {
		t.Fatalf("counts = search today %d total %d", dashboard.TodaySearchCount, dashboard.TotalSearchCount)
	}
	if dashboard.TodayAPICount != 1 || dashboard.TotalAPICount != 30 {
		t.Fatalf("api counts = today %d total %d", dashboard.TodayAPICount, dashboard.TotalAPICount)
	}
	if len(dashboard.TopDocuments) != 2 || dashboard.TopDocuments[0].Slug != "vip" {
		t.Fatalf("top documents = %#v", dashboard.TopDocuments)
	}
	if dashboard.TrackedDocumentCount != 2 {
		t.Fatalf("tracked documents = %d, want 2", dashboard.TrackedDocumentCount)
	}
	if len(dashboard.TodayEvents) != 1 {
		t.Fatalf("today events = %#v", dashboard.TodayEvents)
	}
	if dashboard.KeywordsAvailable {
		t.Fatal("keywords should be unavailable")
	}
}

func TestLoadDashboardLoadsKeywordsWhenPresent(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	now := time.Date(2026, 6, 5, 16, 0, 0, 0, time.UTC)
	writeJSON(t, filepath.Join(statsDir, "keywords.json"), keywordsFile{
		GeneratedAt: now,
		Date:        "2026-06-05",
		Keywords:    []KeywordEntry{{Text: "脑裂", Weight: 5}},
	})

	dashboard, err := LoadDashboard(root, now)
	if err != nil {
		t.Fatalf("LoadDashboard() error = %v", err)
	}
	if !dashboard.KeywordsAvailable || len(dashboard.Keywords) != 1 {
		t.Fatalf("keywords = %#v available=%v", dashboard.Keywords, dashboard.KeywordsAvailable)
	}
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
