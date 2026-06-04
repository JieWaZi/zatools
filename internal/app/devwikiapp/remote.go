package devwikiapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zatools/internal/devwiki"
	devwikigraph "zatools/internal/devwiki/graph"
)

type remoteReadRequest struct {
	Kind string `json:"kind"`
	Slug string `json:"slug"`
	View string `json:"view"`
}

type remoteSearchRequest struct {
	Kind  string   `json:"kind"`
	Query []string `json:"query"`
}

type remoteTextResponse struct {
	Text string `json:"text"`
}

func (s *Service) readRemote(ctx context.Context, source devwiki.RepoSource, opts ReadOptions) error {
	if strings.TrimSpace(opts.Format) != "" && opts.Format != "text" {
		return fmt.Errorf("unsupported devwiki read format %q; only text is supported", opts.Format)
	}
	var response remoteTextResponse
	if err := postRemoteJSON(ctx, source.URL, "/api/devwiki/read", remoteReadRequest{
		Kind: opts.Kind,
		Slug: opts.Slug,
		View: opts.View,
	}, &response); err != nil {
		return err
	}
	_, err := fmt.Fprintln(opts.Stdout, response.Text)
	return err
}

func (s *Service) searchRemote(ctx context.Context, source devwiki.RepoSource, opts SearchOptions) error {
	var raw json.RawMessage
	if err := postRemoteJSON(ctx, source.URL, "/api/devwiki/search", remoteSearchRequest{
		Kind:  opts.Kind,
		Query: normalizeSearchQueries(opts),
	}, &raw); err != nil {
		return err
	}
	switch strings.TrimSpace(opts.Kind) {
	case "index":
		var results []IndexSearchResult
		if err := json.Unmarshal(raw, &results); err != nil {
			return err
		}
		return writeIndexSearchTable(opts.Stdout, results)
	case "glossary":
		var results []GlossarySearchResult
		if err := json.Unmarshal(raw, &results); err != nil {
			return err
		}
		return writeGlossarySearchTable(opts.Stdout, results)
	case "topic", "workflow":
		var results []SearchResult
		if err := json.Unmarshal(raw, &results); err != nil {
			return err
		}
		return writePageSearchTable(opts.Stdout, results)
	default:
		return fmt.Errorf("unsupported devwiki search kind %q", opts.Kind)
	}
}

func (s *Service) glossaryKeywordsRemote(ctx context.Context, source devwiki.RepoSource, opts GlossaryKeywordsOptions) error {
	var response remoteTextResponse
	if err := postRemoteJSON(ctx, source.URL, "/api/devwiki/glossary/keywords", struct{}{}, &response); err != nil {
		return err
	}
	_, err := fmt.Fprint(opts.Stdout, response.Text)
	return err
}

func postRemoteJSON(ctx context.Context, baseURL string, path string, request any, response any) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	url := strings.TrimRight(baseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(devwikigraph.DefaultAPIUsername, devwikigraph.DefaultAPIPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("remote devwiki request failed: %s", resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(response)
}
