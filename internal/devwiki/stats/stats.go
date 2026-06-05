package stats

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// DirName is the stats directory under a DevWiki root.
	DirName = ".devwiki/stats"
	// DefaultFlushInterval is how often in-memory counters are written to disk.
	DefaultFlushInterval = 30 * time.Second
	// DefaultKeywordUpdateInterval is how often keywords.json is refreshed by the server.
	DefaultKeywordUpdateInterval = time.Hour
)

// DocumentEntry tracks read counts for one wiki page.
type DocumentEntry struct {
	Kind      string `json:"kind"`
	ReadCount int    `json:"read_count"`
}

// SearchHit is one compact search result stored in daily events.
type SearchHit struct {
	Slug  string `json:"slug,omitempty"`
	Score string `json:"score,omitempty"`
	Type  string `json:"type,omitempty"`
}

// Event is one append-only daily log record.
type Event struct {
	Timestamp   time.Time   `json:"ts"`
	Endpoint    string      `json:"endpoint"`
	Kind        string      `json:"kind,omitempty"`
	Queries     []string    `json:"queries,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	View        string      `json:"view,omitempty"`
	ResultCount int         `json:"result_count,omitempty"`
	Empty       bool        `json:"empty,omitempty"`
	Results     []SearchHit `json:"results,omitempty"`
}

type documentsFile struct {
	UpdatedAt time.Time                `json:"updated_at"`
	Documents map[string]DocumentEntry `json:"documents"`
}

type summaryFile struct {
	UpdatedAt        time.Time `json:"updated_at"`
	Today            string    `json:"today"`
	TodaySearchCount int       `json:"today_search_count"`
	TodayReadCount   int       `json:"today_read_count"`
	TotalSearchCount int       `json:"total_search_count"`
	TotalReadCount   int       `json:"total_read_count"`
}

// Recorder records DevWiki API usage to files under .devwiki/stats.
type Recorder struct {
	root            string
	mu              sync.Mutex
	today           string
	todaySearch     int
	todayRead       int
	totalSearch     int
	totalRead       int
	documents       map[string]DocumentEntry
	dirtyDocs       bool
	dirtySummary    bool
	flushInterval   time.Duration
	keywordInterval time.Duration
	now             func() time.Time
}

// NewRecorder loads existing stats from disk and returns a ready recorder.
func NewRecorder(root string) *Recorder {
	root = filepath.Clean(root)
	today := time.Now().Format("2006-01-02")
	rec := &Recorder{
		root:            root,
		today:           today,
		documents:       map[string]DocumentEntry{},
		flushInterval:   DefaultFlushInterval,
		keywordInterval: DefaultKeywordUpdateInterval,
		now:             time.Now,
	}
	rec.loadExisting()
	return rec
}

// Start runs the periodic flush loop until ctx is canceled, then flushes once more.
func (r *Recorder) Start(ctx context.Context) {
	r.StartKeywordUpdates(ctx)
	go func() {
		ticker := time.NewTicker(r.flushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				r.Flush()
				return
			case <-ticker.C:
				r.Flush()
			}
		}
	}()
}

// StartKeywordUpdates runs keyword aggregation immediately and then periodically until ctx is canceled.
func (r *Recorder) StartKeywordUpdates(ctx context.Context) {
	if r.keywordInterval <= 0 {
		return
	}
	go func() {
		r.updateKeywords()
		ticker := time.NewTicker(r.keywordInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.updateKeywords()
			}
		}
	}()
}

func (r *Recorder) updateKeywords() {
	_, _ = UpdateKeywords(UpdateKeywordsOptions{
		Root: r.root,
		Now:  r.now(),
	})
}

// RecordSearch increments API search counters and appends a search event.
func (r *Recorder) RecordSearch(kind string, queries []string, results []SearchHit, resultCount int) {
	now := r.now()
	r.mu.Lock()
	r.ensureTodayLocked(now)
	r.todaySearch++
	r.totalSearch++
	r.dirtySummary = true
	r.mu.Unlock()

	event := Event{
		Timestamp:   now,
		Endpoint:    "search",
		Kind:        kind,
		Queries:     append([]string(nil), queries...),
		ResultCount: resultCount,
		Empty:       resultCount == 0,
		Results:     append([]SearchHit(nil), results...),
	}
	_ = r.appendEvent(event)
}

// RecordRead increments read counters and appends a read event.
func (r *Recorder) RecordRead(kind, slug, view string) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return
	}
	now := r.now()
	r.mu.Lock()
	r.ensureTodayLocked(now)
	r.todayRead++
	r.totalRead++
	entry := r.documents[slug]
	entry.Kind = kind
	entry.ReadCount++
	r.documents[slug] = entry
	r.dirtyDocs = true
	r.dirtySummary = true
	r.mu.Unlock()

	event := Event{
		Timestamp: now,
		Endpoint:  "read",
		Kind:      kind,
		Slug:      slug,
		View:      view,
	}
	_ = r.appendEvent(event)
}

// Flush writes dirty documents and summary files to disk.
func (r *Recorder) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	r.ensureTodayLocked(now)

	if r.dirtyDocs {
		payload := documentsFile{
			UpdatedAt: now,
			Documents: cloneDocuments(r.documents),
		}
		if err := writeJSONAtomic(filepath.Join(r.statsDir(), "documents.json"), payload); err == nil {
			r.dirtyDocs = false
		}
	}

	if r.dirtySummary {
		payload := summaryFile{
			UpdatedAt:        now,
			Today:            r.today,
			TodaySearchCount: r.todaySearch,
			TodayReadCount:   r.todayRead,
			TotalSearchCount: r.totalSearch,
			TotalReadCount:   r.totalRead,
		}
		if err := writeJSONAtomic(filepath.Join(r.statsDir(), "summary.json"), payload); err == nil {
			r.dirtySummary = false
		}
	}
}

func (r *Recorder) loadExisting() {
	if data, err := os.ReadFile(filepath.Join(r.statsDir(), "documents.json")); err == nil {
		var loaded documentsFile
		if json.Unmarshal(data, &loaded) == nil && loaded.Documents != nil {
			r.documents = loaded.Documents
		}
	}
	if data, err := os.ReadFile(filepath.Join(r.statsDir(), "summary.json")); err == nil {
		var loaded summaryFile
		if json.Unmarshal(data, &loaded) == nil {
			r.totalSearch = loaded.TotalSearchCount
			r.totalRead = loaded.TotalReadCount
			if loaded.Today == r.today {
				r.todaySearch = loaded.TodaySearchCount
				r.todayRead = loaded.TodayReadCount
			}
		}
	}
}

func (r *Recorder) ensureTodayLocked(now time.Time) {
	today := now.Format("2006-01-02")
	if today == r.today {
		return
	}
	r.today = today
	r.todaySearch = 0
	r.todayRead = 0
	r.dirtySummary = true
}

func (r *Recorder) appendEvent(event Event) error {
	if err := os.MkdirAll(r.statsDir(), 0o755); err != nil {
		return err
	}
	path := filepath.Join(r.statsDir(), dailyLogName(event.Timestamp))
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(event)
}

func (r *Recorder) statsDir() string {
	return filepath.Join(r.root, filepath.FromSlash(DirName))
}

func dailyLogName(ts time.Time) string {
	return fmt.Sprintf("queries-%s.jsonl", ts.Format("2006-01-02"))
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func cloneDocuments(src map[string]DocumentEntry) map[string]DocumentEntry {
	out := make(map[string]DocumentEntry, len(src))
	for slug, entry := range src {
		out[slug] = entry
	}
	return out
}

// ReadEvents loads all events from a daily JSONL file.
func ReadEvents(path string) ([]Event, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	events := make([]Event, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}

// TopDocuments returns up to limit documents sorted by read_count descending.
func TopDocuments(documents map[string]DocumentEntry, limit int) []DashboardDocument {
	if limit <= 0 {
		limit = 10
	}
	items := make([]DashboardDocument, 0, len(documents))
	for slug, entry := range documents {
		items = append(items, DashboardDocument{
			Slug:      slug,
			Kind:      entry.Kind,
			ReadCount: entry.ReadCount,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].ReadCount == items[j].ReadCount {
			return items[i].Slug < items[j].Slug
		}
		return items[i].ReadCount > items[j].ReadCount
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}
