package devwikiapp

import (
	"zatools/internal/devwiki/retrieval"
	"zatools/internal/devwiki/stats"
)

func recordSearchStats(root, kind string, queries []string, resultCount int, hits []stats.SearchHit) {
	rec := stats.NewRecorder(root)
	rec.RecordSearch(kind, queries, hits, resultCount)
	rec.Flush()
}

func recordReadStats(root, kind, slug, view string) {
	rec := stats.NewRecorder(root)
	rec.RecordRead(kind, slug, view)
	rec.Flush()
}

func indexSearchHits(results []retrieval.IndexSearchResult) []stats.SearchHit {
	hits := make([]stats.SearchHit, 0, len(results))
	for _, result := range results {
		hits = append(hits, stats.SearchHit{
			Slug: result.Slug,
			Type: result.Type,
		})
	}
	return hits
}

func glossarySearchHits(results []retrieval.GlossarySearchResult) []stats.SearchHit {
	hits := make([]stats.SearchHit, 0, len(results))
	for _, result := range results {
		hits = append(hits, stats.SearchHit{
			Slug: result.Slug,
			Type: result.Type,
		})
	}
	return hits
}

func pageSearchHits(results []retrieval.SearchResult) []stats.SearchHit {
	hits := make([]stats.SearchHit, 0, len(results))
	for _, result := range results {
		hits = append(hits, stats.SearchHit{
			Slug:  result.Slug,
			Score: result.Score,
		})
	}
	return hits
}
