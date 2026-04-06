package skills

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSourceVariants(t *testing.T) {
	t.Parallel()

	localDir := t.TempDir()

	tests := []struct {
		name   string
		input  string
		assert func(t *testing.T, got Source)
	}{
		{
			name:  "local path",
			input: localDir,
			assert: func(t *testing.T, got Source) {
				t.Helper()
				if got.Type != "local" || got.LocalDir == "" {
					t.Fatalf("ParseSource(local) = %#v", got)
				}
			},
		},
		{
			name:  "github shorthand",
			input: "owner/repo/skills/demo",
			assert: func(t *testing.T, got Source) {
				t.Helper()
				if got.Type != "github" || got.Subpath != "skills/demo" {
					t.Fatalf("ParseSource(github shorthand) = %#v", got)
				}
			},
		},
		{
			name:  "github tree url",
			input: "https://github.com/owner/repo/tree/main/skills/demo",
			assert: func(t *testing.T, got Source) {
				t.Helper()
				if got.Type != "github" || got.Ref != "main" || got.Subpath != "skills/demo" {
					t.Fatalf("ParseSource(github tree url) = %#v", got)
				}
			},
		},
		{
			name:  "gitlab tree url",
			input: "https://gitlab.com/group/repo/-/tree/main/skills/demo",
			assert: func(t *testing.T, got Source) {
				t.Helper()
				if got.Type != "gitlab" || got.Ref != "main" || got.Subpath != "skills/demo" {
					t.Fatalf("ParseSource(gitlab tree url) = %#v", got)
				}
			},
		},
		{
			name:  "alias",
			input: "coinbase/agentWallet",
			assert: func(t *testing.T, got Source) {
				t.Helper()
				if got.RepoURL != "https://github.com/coinbase/agentic-wallet-skills.git" {
					t.Fatalf("ParseSource(alias) repo = %q", got.RepoURL)
				}
			},
		},
		{
			name:  "direct git url",
			input: "https://example.com/repo.git",
			assert: func(t *testing.T, got Source) {
				t.Helper()
				if got.Type != "git" || got.RepoURL != "https://example.com/repo.git" {
					t.Fatalf("ParseSource(direct git) = %#v", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSource(tt.input)
			if err != nil {
				t.Fatalf("ParseSource(%q) error = %v", tt.input, err)
			}
			tt.assert(t, got)
		})
	}
}

func TestParseSourceRejectsUnsupported(t *testing.T) {
	t.Parallel()

	if _, err := ParseSource("://bad"); err == nil {
		t.Fatal("expected unsupported source error")
	}
}

func TestDiscoverAndParseSkillFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "top", "SKILL.md"), "---\nname: top\ndescription: top skill\n---\n")
	mustWriteFile(t, filepath.Join(root, "top", "nested", "SKILL.md"), "---\nname: nested\ndescription: nested skill\n---\n")
	mustWriteFile(t, filepath.Join(root, "vendor", "ignored", "SKILL.md"), "---\nname: ignored\ndescription: ignored skill\n---\n")

	found, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover error = %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("Discover found %d skills, want 1", len(found))
	}
	if found[0].Name != "top" || found[0].RelativeDir != "top" {
		t.Fatalf("Discover found = %#v", found[0])
	}

	parsed, err := ParseSkillFile(filepath.Join(root, "top", "SKILL.md"))
	if err != nil {
		t.Fatalf("ParseSkillFile error = %v", err)
	}
	if parsed.Name != "top" || parsed.Description != "top skill" {
		t.Fatalf("ParseSkillFile = %#v", parsed)
	}
}

func TestParseSkillFileRejectsMissingFrontmatter(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "SKILL.md")
	mustWriteFile(t, path, "# no frontmatter\n")

	if _, err := ParseSkillFile(path); err == nil {
		t.Fatal("expected missing frontmatter error")
	}
}

func TestParseSkillFileFallsBackToDirectoryNameWhenNameMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "skill-stocktake", "SKILL.md")
	mustWriteFile(t, path, "---\ndescription: stocktake skill\n---\n")

	got, err := ParseSkillFile(path)
	if err != nil {
		t.Fatalf("ParseSkillFile error = %v", err)
	}
	if got.Name != "skill-stocktake" {
		t.Fatalf("ParseSkillFile name = %q, want %q", got.Name, "skill-stocktake")
	}
	if got.Description != "stocktake skill" {
		t.Fatalf("ParseSkillFile description = %q, want %q", got.Description, "stocktake skill")
	}
}

func TestParseSkillFileFallsBackWhenDescriptionMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "skill-stocktake", "SKILL.md")
	mustWriteFile(t, path, "---\nname: skill-stocktake\n---\n")

	got, err := ParseSkillFile(path)
	if err != nil {
		t.Fatalf("ParseSkillFile error = %v", err)
	}
	if got.Name != "skill-stocktake" {
		t.Fatalf("ParseSkillFile name = %q, want %q", got.Name, "skill-stocktake")
	}
	if got.Description != fallbackSkillDescription {
		t.Fatalf("ParseSkillFile description = %q, want %q", got.Description, fallbackSkillDescription)
	}
}

func TestResolveSourceLocalAndSearchRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := Source{Type: "local", LocalDir: root, Subpath: "skills/demo"}
	mustMkdirAll(t, filepath.Join(root, "skills", "demo"))

	resolved, err := ResolveSource(context.Background(), source)
	if err != nil {
		t.Fatalf("ResolveSource(local) error = %v", err)
	}
	defer func() {
		if err := resolved.Cleanup(); err != nil {
			t.Fatalf("Cleanup error = %v", err)
		}
	}()

	searchRoot, err := resolved.SearchRoot()
	if err != nil {
		t.Fatalf("SearchRoot error = %v", err)
	}
	if want := filepath.Join(root, "skills", "demo"); searchRoot != want {
		t.Fatalf("SearchRoot = %q, want %q", searchRoot, want)
	}
}

func TestResolveSourceRejectsMissingLocalDir(t *testing.T) {
	t.Parallel()

	_, err := ResolveSource(context.Background(), Source{Type: "local", LocalDir: filepath.Join(t.TempDir(), "missing")})
	if err == nil || !strings.Contains(err.Error(), "stat local source") {
		t.Fatalf("ResolveSource missing dir error = %v, want stat error", err)
	}
}

func TestCleanupWithoutTempDirIsNoop(t *testing.T) {
	t.Parallel()

	if err := (ResolvedSource{}).Cleanup(); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}
}

func TestParseFragmentAndHelpers(t *testing.T) {
	t.Parallel()

	base, ref := parseFragmentRef("github:owner/repo#main")
	if base != "github:owner/repo" || ref != "main" {
		t.Fatalf("parseFragmentRef = %q, %q", base, ref)
	}
	if got := appendFragmentRef("owner/repo", "main"); got != "owner/repo#main" {
		t.Fatalf("appendFragmentRef = %q", got)
	}
	if !looksLikeGitSource("owner/repo") {
		t.Fatal("looksLikeGitSource(owner/repo) = false")
	}
	if !isDirectGitURL("git@example.com:repo.git") {
		t.Fatal("isDirectGitURL(git@...) = false")
	}
	if got := decodeFragmentValue("feature%2Fmain"); got != "feature/main" {
		t.Fatalf("decodeFragmentValue = %q", got)
	}
	if got := shouldSkipDir("vendor"); !got {
		t.Fatal("shouldSkipDir(vendor) = false")
	}
}

func TestSanitizeSubpathAndCommitHashHelpers(t *testing.T) {
	t.Parallel()

	got, err := sanitizeSubpath(`skills\demo`)
	if err != nil {
		t.Fatalf("sanitizeSubpath error = %v", err)
	}
	if got != "skills/demo" {
		t.Fatalf("sanitizeSubpath = %q, want %q", got, "skills/demo")
	}

	if _, err := sanitizeSubpath("../escape"); err == nil {
		t.Fatal("expected sanitizeSubpath traversal error")
	}
	if !looksLikeCommitHash("abcdef1") {
		t.Fatal("looksLikeCommitHash(abcdef1) = false")
	}
	if looksLikeCommitHash("not-a-hash") {
		t.Fatal("looksLikeCommitHash(not-a-hash) = true")
	}
}

func TestResolveSourceNilContextFallsBack(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	resolved, err := ResolveSource(nil, Source{Type: "local", LocalDir: root})
	if err != nil {
		t.Fatalf("ResolveSource(nil ctx) error = %v", err)
	}
	if resolved.RootDir != root {
		t.Fatalf("ResolveSource(nil ctx) root = %q, want %q", resolved.RootDir, root)
	}
}

func TestIsLocalPathDetectsExistingRelativePath(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd error = %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir error = %v", err)
	}
	mustWriteFile(t, "example.txt", "data")

	if !isLocalPath("example.txt") {
		t.Fatal("isLocalPath(existing relative file) = false")
	}
}
