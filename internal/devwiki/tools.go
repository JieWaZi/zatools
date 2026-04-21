package devwiki

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	indexTemplate = "# Wiki Index\n"
	logTemplate   = "# Wiki Log\n\n> Append-only chronological log.\n"
)

var (
	resetScopes = []string{"wiki", "raw", "log", "checkpoints"}
	rawDirNames = []string{"requirements", "designs", "features", "tests"}
)

// ResetPlan describes which files would be removed or rewritten for a reset run.
type ResetPlan struct {
	Root   string   `json:"root"`
	Scopes []string `json:"scopes"`
	Delete []string `json:"delete"`
	Reset  []string `json:"reset"`
}

// ResetResult reports the files touched by an applied reset.
type ResetResult struct {
	Scopes       []string `json:"scopes"`
	DeletedCount int      `json:"deleted_count"`
	ResetCount   int      `json:"reset_count"`
	DeletedFiles []string `json:"deleted_files"`
	ResetFiles   []string `json:"reset_files"`
}

// CodeRef describes a persisted code reference path.
type CodeRef struct {
	Path string `json:"path" yaml:"path"`
}

// CodeCandidate is a scored code search hit.
type CodeCandidate struct {
	Path  string `json:"path"`
	Score int    `json:"score"`
}

var docTypes = map[string]string{
	"requirements": "requirement",
	"designs":      "design",
	"features":     "feature",
	"tests":        "test",
}

// ParseResetScopes parses a reset scope list and expands `all`.
func ParseResetScopes(raw string) ([]string, error) {
	tokens := strings.Split(raw, ",")
	scopes := make([]string, 0, len(tokens))
	seen := make(map[string]struct{}, len(resetScopes))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if token == "all" {
			for _, scope := range resetScopes {
				if _, ok := seen[scope]; ok {
					continue
				}
				seen[scope] = struct{}{}
				scopes = append(scopes, scope)
			}
			continue
		}
		if !containsResetScope(token) {
			valid := append(append([]string{}, resetScopes...), "all")
			return nil, fmt.Errorf("unknown scope: %s. valid scopes: %s", token, strings.Join(valid, ", "))
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		scopes = append(scopes, token)
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("scope is required")
	}
	return scopes, nil
}

// BuildResetPlan returns the delete/reset plan for the requested scopes.
func BuildResetPlan(root string, scopes []string) (ResetPlan, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ResetPlan{}, err
	}

	deleteSet := map[string]struct{}{}
	resetSet := map[string]struct{}{}

	if containsScope(scopes, "wiki") {
		for _, rel := range []string{
			filepath.Join("wiki", "capabilities"),
			filepath.Join("wiki", "features"),
			filepath.Join("wiki", "outputs"),
			filepath.Join("wiki", "graph"),
		} {
			for _, path := range resetCandidates(filepath.Join(absRoot, rel)) {
				deleteSet[path] = struct{}{}
			}
		}
		resetSet[filepath.Join(absRoot, "wiki", "index.md")] = struct{}{}
	}

	if containsScope(scopes, "raw") {
		for _, name := range rawDirNames {
			for _, path := range resetCandidates(filepath.Join(absRoot, "raw", name)) {
				deleteSet[path] = struct{}{}
			}
		}
	}

	if containsScope(scopes, "log") {
		resetSet[filepath.Join(absRoot, "wiki", "log.md")] = struct{}{}
	}

	if containsScope(scopes, "checkpoints") {
		for _, path := range resetCandidates(filepath.Join(absRoot, "wiki", ".checkpoints")) {
			deleteSet[path] = struct{}{}
		}
	}

	return ResetPlan{
		Root:   absRoot,
		Scopes: append([]string(nil), scopes...),
		Delete: sortedKeys(deleteSet),
		Reset:  sortedKeys(resetSet),
	}, nil
}

// ApplyResetPlan deletes the planned files and restores index/log templates.
func ApplyResetPlan(plan ResetPlan) (ResetResult, error) {
	deleted := make([]string, 0, len(plan.Delete))
	for _, target := range plan.Delete {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return ResetResult{}, err
		}
		if err := os.Remove(target); err != nil {
			return ResetResult{}, err
		}
		deleted = append(deleted, target)
	}

	resetFiles := make([]string, 0, len(plan.Reset))
	for _, target := range plan.Reset {
		if err := rewriteResetFile(target); err != nil {
			return ResetResult{}, err
		}
		resetFiles = append(resetFiles, target)
	}

	return ResetResult{
		Scopes:       append([]string(nil), plan.Scopes...),
		DeletedCount: len(deleted),
		ResetCount:   len(resetFiles),
		DeletedFiles: deleted,
		ResetFiles:   resetFiles,
	}, nil
}

// AppendLog appends one dated entry to `wiki/log.md`.
func AppendLog(wikiRoot string, message string, now time.Time) error {
	logPath := filepath.Join(wikiRoot, "log.md")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "## [%s] %s\n", now.Format("2006-01-02"), message)
	return err
}

// ClassifyDocType infers the canonical doc type from the raw document's parent directory.
func ClassifyDocType(path string) (string, error) {
	docType, ok := docTypes[filepath.Base(filepath.Dir(path))]
	if !ok {
		return "", fmt.Errorf("unknown doc type for %s", path)
	}
	return docType, nil
}

// ExtractTitle extracts the first markdown H1, falling back to the file stem.
func ExtractTitle(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# ")), nil
		}
	}
	name := filepath.Base(path)
	return strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name))), nil
}

// SourceHash returns the SHA-1 digest of the file contents.
func SourceHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:]), nil
}

// SearchCandidates finds scored file hits for the given terms.
func SearchCandidates(codeRoot string, terms []string, limit int) ([]CodeCandidate, error) {
	if limit <= 0 {
		limit = 10
	}

	results := make([]CodeCandidate, 0, limit)
	err := filepath.WalkDir(codeRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		score := scoreText(filepath.Base(path), terms)*3 + scoreText(string(data), terms)
		if score <= 0 {
			return nil
		}
		results = append(results, CodeCandidate{Path: path, Score: score})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Path < results[j].Path
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

// FindMissingCodePaths returns paths that do not exist on disk.
func FindMissingCodePaths(refs []CodeRef) []string {
	missing := make([]string, 0, len(refs))
	for _, ref := range refs {
		if _, err := os.Stat(ref.Path); os.IsNotExist(err) {
			missing = append(missing, ref.Path)
		}
	}
	return missing
}

func containsResetScope(scope string) bool {
	for _, candidate := range resetScopes {
		if candidate == scope {
			return true
		}
	}
	return false
}

func containsScope(scopes []string, target string) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func resetCandidates(base string) []string {
	info, err := os.Stat(base)
	if err != nil || !info.IsDir() {
		return nil
	}

	results := make([]string, 0)
	_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() || d.Name() == ".gitkeep" {
			return nil
		}
		results = append(results, path)
		return nil
	})
	sort.Strings(results)
	return results
}

func rewriteResetFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content := ""
	switch filepath.Base(path) {
	case "index.md":
		content = indexTemplate
	case "log.md":
		content = logTemplate
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func scoreText(text string, terms []string) int {
	lower := strings.ToLower(text)
	score := 0
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term == "" {
			continue
		}
		score += strings.Count(lower, term)
	}
	return score
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
