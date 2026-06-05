package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const maxKeywords = 100

// UpdateKeywordsOptions controls keyword aggregation from daily query logs.
type UpdateKeywordsOptions struct {
	Root string
	Now  time.Time
}

// UpdateKeywordsResult reports keyword aggregation output.
type UpdateKeywordsResult struct {
	GeneratedAt time.Time      `json:"generated_at"`
	LastEventAt time.Time      `json:"last_event_at"`
	Date        string         `json:"date"`
	Keywords    []KeywordEntry `json:"keywords"`
	AddedTerms  int            `json:"added_terms"`
	OutputPath  string         `json:"output_path"`
}

// UpdateKeywords merges search terms from queries-*.jsonl into keywords.json incrementally.
func UpdateKeywords(opts UpdateKeywordsOptions) (UpdateKeywordsResult, error) {
	root := filepath.Clean(opts.Root)
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	statsDir := filepath.Join(root, filepath.FromSlash(DirName))
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		return UpdateKeywordsResult{}, err
	}

	weights := map[string]int{}
	lastEventAt := time.Time{}
	keywordsPath := filepath.Join(statsDir, "keywords.json")
	if data, err := os.ReadFile(keywordsPath); err == nil {
		var existing keywordsFile
		if json.Unmarshal(data, &existing) == nil {
			for _, item := range existing.Keywords {
				text := strings.TrimSpace(item.Text)
				if text == "" {
					continue
				}
				weights[text] = item.Weight
			}
			lastEventAt = existing.LastEventAt
		}
	}

	addedTerms := 0
	latestEventAt := lastEventAt
	matches, err := filepath.Glob(filepath.Join(statsDir, "queries-*.jsonl"))
	if err != nil {
		return UpdateKeywordsResult{}, err
	}
	sort.Strings(matches)
	for _, path := range matches {
		events, err := ReadEvents(path)
		if err != nil {
			return UpdateKeywordsResult{}, err
		}
		for _, event := range events {
			if event.Endpoint != "search" {
				continue
			}
			if !lastEventAt.IsZero() && !event.Timestamp.After(lastEventAt) {
				continue
			}
			for _, term := range tokenizeQueries(event.Queries) {
				weights[term]++
				addedTerms++
			}
			if event.Timestamp.After(latestEventAt) {
				latestEventAt = event.Timestamp
			}
		}
	}

	keywords := topKeywordEntries(weights, maxKeywords)
	payload := keywordsFile{
		GeneratedAt: now,
		LastEventAt: latestEventAt,
		Date:        now.Format("2006-01-02"),
		Keywords:    keywords,
	}
	if err := writeJSONAtomic(keywordsPath, payload); err != nil {
		return UpdateKeywordsResult{}, err
	}
	return UpdateKeywordsResult{
		GeneratedAt: payload.GeneratedAt,
		LastEventAt: payload.LastEventAt,
		Date:        payload.Date,
		Keywords:    payload.Keywords,
		AddedTerms:  addedTerms,
		OutputPath:  keywordsPath,
	}, nil
}

func tokenizeQueries(queries []string) []string {
	terms := make([]string, 0, len(queries))
	seen := map[string]struct{}{}
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		addTerm(seen, &terms, query)
		var builder strings.Builder
		for _, r := range query {
			if unicode.IsSpace(r) || unicode.IsPunct(r) {
				if builder.Len() > 0 {
					addTerm(seen, &terms, builder.String())
					builder.Reset()
				}
				continue
			}
			builder.WriteRune(r)
		}
		if builder.Len() > 0 {
			addTerm(seen, &terms, builder.String())
		}
	}
	return terms
}

func addTerm(seen map[string]struct{}, terms *[]string, raw string) {
	term := strings.TrimSpace(raw)
	if term == "" {
		return
	}
	if _, ok := seen[term]; ok {
		return
	}
	seen[term] = struct{}{}
	*terms = append(*terms, term)
}

func topKeywordEntries(weights map[string]int, limit int) []KeywordEntry {
	if limit <= 0 {
		limit = maxKeywords
	}
	items := make([]KeywordEntry, 0, len(weights))
	for text, weight := range weights {
		if weight <= 0 {
			continue
		}
		items = append(items, KeywordEntry{Text: text, Weight: weight})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Weight == items[j].Weight {
			return items[i].Text < items[j].Text
		}
		return items[i].Weight > items[j].Weight
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}
