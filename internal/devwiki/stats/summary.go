package stats

import (
	"path/filepath"
	"sort"
)

// SummaryCounts holds cumulative API counters rebuilt from query logs.
type SummaryCounts struct {
	SearchCount int
	ReadCount   int
}

// SummaryCountsFromLogs rebuilds cumulative search/read counts from queries-*.jsonl.
func SummaryCountsFromLogs(statsDir string) (SummaryCounts, error) {
	matches, err := filepath.Glob(filepath.Join(statsDir, "queries-*.jsonl"))
	if err != nil {
		return SummaryCounts{}, err
	}
	sort.Strings(matches)
	counts := SummaryCounts{}
	for _, path := range matches {
		events, err := ReadEvents(path)
		if err != nil {
			return SummaryCounts{}, err
		}
		for _, event := range events {
			switch event.Endpoint {
			case "search":
				counts.SearchCount++
			case "read":
				counts.ReadCount++
			}
		}
	}
	return counts, nil
}

func mergeSummaryCounts(file summaryFile, logs SummaryCounts) summaryFile {
	out := file
	if logs.SearchCount > out.TotalSearchCount {
		out.TotalSearchCount = logs.SearchCount
	}
	if logs.ReadCount > out.TotalReadCount {
		out.TotalReadCount = logs.ReadCount
	}
	return out
}
