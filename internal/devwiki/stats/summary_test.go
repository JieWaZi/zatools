package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSummaryCountsFromLogsRebuildsTotals(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	now := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	appendEvent(t, statsDir, Event{Timestamp: now, Endpoint: "search", Queries: []string{"VIP"}, ResultCount: 1})
	appendEvent(t, statsDir, Event{Timestamp: now, Endpoint: "search", Queries: []string{"HA"}, ResultCount: 0, Empty: true})
	appendEvent(t, statsDir, Event{Timestamp: now, Endpoint: "read", Kind: "topic", Slug: "vip", View: "card"})

	counts, err := SummaryCountsFromLogs(statsDir)
	if err != nil {
		t.Fatalf("SummaryCountsFromLogs() error = %v", err)
	}
	if counts.SearchCount != 2 || counts.ReadCount != 1 {
		t.Fatalf("counts = %#v", counts)
	}
}

func TestLoadDashboardUsesLogTotalsWhenSummaryMissing(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	now := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)
	appendEvent(t, statsDir, Event{Timestamp: now, Endpoint: "search", Queries: []string{"VIP"}, ResultCount: 1})
	appendEvent(t, statsDir, Event{Timestamp: now, Endpoint: "read", Kind: "topic", Slug: "vip", View: "core"})

	dashboard, err := LoadDashboard(root, now)
	if err != nil {
		t.Fatalf("LoadDashboard() error = %v", err)
	}
	if dashboard.TodayAPICount != 2 || dashboard.TotalAPICount != 2 {
		t.Fatalf("api counts = today %d total %d", dashboard.TodayAPICount, dashboard.TotalAPICount)
	}
}
