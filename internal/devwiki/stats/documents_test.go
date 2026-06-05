package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDocumentCountsFromLogsRebuildsReadCounts(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	logPath := filepath.Join(statsDir, dailyLogName(now))
	file, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	encoder := json.NewEncoder(file)
	for _, event := range []Event{
		{Timestamp: now, Endpoint: "read", Kind: "topic", Slug: "vip", View: "card"},
		{Timestamp: now, Endpoint: "read", Kind: "topic", Slug: "vip", View: "core"},
		{Timestamp: now, Endpoint: "read", Kind: "workflow", Slug: "wf", View: "explain"},
	} {
		if err := encoder.Encode(event); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}
	file.Close()

	counts, err := DocumentCountsFromLogs(statsDir)
	if err != nil {
		t.Fatalf("DocumentCountsFromLogs() error = %v", err)
	}
	if counts["vip"].ReadCount != 2 || counts["vip"].Kind != "topic" {
		t.Fatalf("vip = %#v", counts["vip"])
	}
	if counts["wf"].ReadCount != 1 {
		t.Fatalf("wf = %#v", counts["wf"])
	}
}

func TestMergeDocumentCountsPrefersHigherReadCount(t *testing.T) {
	base := map[string]DocumentEntry{
		"vip": {Kind: "topic", ReadCount: 5},
	}
	logs := map[string]DocumentEntry{
		"vip": {Kind: "topic", ReadCount: 8},
		"wf":  {Kind: "workflow", ReadCount: 2},
	}
	merged := mergeDocumentCounts(base, logs)
	if merged["vip"].ReadCount != 8 {
		t.Fatalf("vip read count = %d, want 8", merged["vip"].ReadCount)
	}
	if merged["wf"].ReadCount != 2 {
		t.Fatalf("wf read count = %d, want 2", merged["wf"].ReadCount)
	}
}

func TestLoadDashboardTopDocumentsFromLogsWhenDocumentsFileMissing(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	now := time.Date(2026, 6, 5, 13, 0, 0, 0, time.UTC)
	appendEvent(t, statsDir, Event{Timestamp: now, Endpoint: "read", Kind: "topic", Slug: "vip", View: "card"})
	appendEvent(t, statsDir, Event{Timestamp: now, Endpoint: "read", Kind: "topic", Slug: "vip", View: "core"})

	dashboard, err := LoadDashboard(root, now)
	if err != nil {
		t.Fatalf("LoadDashboard() error = %v", err)
	}
	if len(dashboard.TopDocuments) != 1 || dashboard.TopDocuments[0].Slug != "vip" {
		t.Fatalf("top documents = %#v", dashboard.TopDocuments)
	}
	if dashboard.TopDocuments[0].ReadCount != 2 {
		t.Fatalf("vip read count = %d, want 2", dashboard.TopDocuments[0].ReadCount)
	}
}
