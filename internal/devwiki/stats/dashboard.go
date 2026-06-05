package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// KeywordEntry is one weighted keyword for the stats dashboard word cloud.
type KeywordEntry struct {
	Text   string `json:"text"`
	Weight int    `json:"weight"`
}

type keywordsFile struct {
	GeneratedAt time.Time      `json:"generated_at"`
	LastEventAt time.Time      `json:"last_event_at,omitempty"`
	Date        string         `json:"date"`
	Keywords    []KeywordEntry `json:"keywords"`
}

// DashboardDocument is one ranked document for the stats page.
type DashboardDocument struct {
	Slug      string `json:"slug"`
	Kind      string `json:"kind"`
	ReadCount int    `json:"read_count"`
}

// Dashboard is the aggregated stats payload for the graph stats page.
type Dashboard struct {
	ProjectSlug           string              `json:"project_slug"`
	ProjectName           string              `json:"project_name"`
	UpdatedAt             time.Time           `json:"updated_at"`
	Today                 string              `json:"today"`
	TodayLogFile          string              `json:"today_log_file"`
	TodaySearchCount      int                 `json:"today_search_count"`
	TodayReadCount        int                 `json:"today_read_count"`
	TodayAPICount         int                 `json:"today_api_count"`
	TotalSearchCount      int                 `json:"total_search_count"`
	TotalReadCount        int                 `json:"total_read_count"`
	TotalAPICount         int                 `json:"total_api_count"`
	TodayEmptySearchCount int                 `json:"today_empty_search_count"`
	TrackedDocumentCount  int                 `json:"tracked_document_count"`
	TopDocuments          []DashboardDocument `json:"top_documents"`
	TodayEvents           []Event             `json:"today_events"`
	Keywords              []KeywordEntry      `json:"keywords"`
	KeywordsAvailable     bool                `json:"keywords_available"`
	DocumentsUpdatedAt    time.Time           `json:"documents_updated_at,omitempty"`
}

// LoadDashboard reads stats files from a DevWiki root and returns dashboard data.
func LoadDashboard(root string, now time.Time) (Dashboard, error) {
	root = filepath.Clean(root)
	statsDir := filepath.Join(root, filepath.FromSlash(DirName))
	today := now.Format("2006-01-02")

	dashboard := Dashboard{
		Today:             today,
		TodayLogFile:      dailyLogName(now),
		TopDocuments:      []DashboardDocument{},
		TodayEvents:       []Event{},
		Keywords:          []KeywordEntry{},
		KeywordsAvailable: false,
		UpdatedAt:         now,
	}
	dashboard.ProjectSlug, dashboard.ProjectName = loadProjectInfo(root)

	summaryPath := filepath.Join(statsDir, "summary.json")
	summaryFromFile := summaryFile{}
	if data, err := os.ReadFile(summaryPath); err == nil {
		if json.Unmarshal(data, &summaryFromFile) == nil {
			dashboard.UpdatedAt = summaryFromFile.UpdatedAt
		}
	}
	logTotals, err := SummaryCountsFromLogs(statsDir)
	if err != nil {
		return dashboard, err
	}
	mergedSummary := mergeSummaryCounts(summaryFromFile, logTotals)
	dashboard.TotalSearchCount = mergedSummary.TotalSearchCount
	dashboard.TotalReadCount = mergedSummary.TotalReadCount
	documentsPath := filepath.Join(statsDir, "documents.json")
	documentsFromFile := loadDocumentsFile(documentsPath)
	if data, err := os.ReadFile(documentsPath); err == nil {
		var documents documentsFile
		if json.Unmarshal(data, &documents) == nil {
			dashboard.DocumentsUpdatedAt = documents.UpdatedAt
		}
	}
	documentsFromLogs, err := DocumentCountsFromLogs(statsDir)
	if err != nil {
		return dashboard, err
	}
	mergedDocuments := mergeDocumentCounts(documentsFromFile, documentsFromLogs)
	dashboard.TrackedDocumentCount = len(mergedDocuments)
	dashboard.TopDocuments = TopDocuments(mergedDocuments, 10)

	events, err := ReadEvents(filepath.Join(statsDir, dailyLogName(now)))
	if err != nil {
		return dashboard, err
	}
	dashboard.TodaySearchCount = 0
	dashboard.TodayReadCount = 0
	dashboard.TodayEmptySearchCount = 0
	for _, event := range events {
		dashboard.TodayEvents = append(dashboard.TodayEvents, event)
		switch event.Endpoint {
		case "search":
			dashboard.TodaySearchCount++
			if event.Empty {
				dashboard.TodayEmptySearchCount++
			}
		case "read":
			dashboard.TodayReadCount++
		}
	}
	dashboard.TodayAPICount = dashboard.TodaySearchCount + dashboard.TodayReadCount
	dashboard.TotalAPICount = dashboard.TotalSearchCount + dashboard.TotalReadCount

	keywordsPath := filepath.Join(statsDir, "keywords.json")
	if data, err := os.ReadFile(keywordsPath); err == nil {
		var keywords keywordsFile
		if json.Unmarshal(data, &keywords) == nil && len(keywords.Keywords) > 0 {
			dashboard.Keywords = keywords.Keywords
			dashboard.KeywordsAvailable = true
		}
	}

	return dashboard, nil
}

func loadProjectInfo(root string) (string, string) {
	slug := filepath.Base(root)
	name := slug
	data, err := os.ReadFile(filepath.Join(root, "config", "project.yaml"))
	if err != nil {
		return slug, name
	}
	var parsed struct {
		ProjectSlug string `yaml:"project_slug"`
		ProjectName string `yaml:"project_name"`
	}
	if yaml.Unmarshal(data, &parsed) != nil {
		return slug, name
	}
	if strings.TrimSpace(parsed.ProjectSlug) != "" {
		slug = strings.TrimSpace(parsed.ProjectSlug)
	}
	if strings.TrimSpace(parsed.ProjectName) != "" {
		name = strings.TrimSpace(parsed.ProjectName)
	}
	return slug, name
}
