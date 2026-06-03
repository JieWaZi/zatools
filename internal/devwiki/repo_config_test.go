package devwiki

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSaveAndLoadRepoConfigRemoteSource(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	cfg := RepoConfig{
		ProjectName:  "huawei-zddi",
		ProjectSlug:  "huawei-zddi",
		Language:     "zh",
		ActiveSource: RepoSourceRemote,
		Sources: RepoSources{
			Remote: &RepoSource{
				Type: RepoSourceRemote,
				URL:  "http://devwiki.example.com:5697",
			},
		},
	}
	if err := SaveRepoConfig(cfg); err != nil {
		t.Fatalf("SaveRepoConfig() error = %v", err)
	}

	got, err := LoadRepoConfig("huawei-zddi")
	if err != nil {
		t.Fatalf("LoadRepoConfig() error = %v", err)
	}
	if got.ProjectName != cfg.ProjectName || got.ProjectSlug != cfg.ProjectSlug || got.Language != cfg.Language {
		t.Fatalf("config metadata = %#v, want %#v", got, cfg)
	}
	if got.ActiveSource != RepoSourceRemote {
		t.Fatalf("active source = %q", got.ActiveSource)
	}
	if got.Sources.Remote == nil || got.Sources.Remote.URL != "http://devwiki.example.com:5697" {
		t.Fatalf("loaded remote source = %#v", got.Sources.Remote)
	}
	data, err := os.ReadFile(filepath.Join(configRoot, "devwiki", "huawei-zddi", "config.yaml"))
	if err != nil {
		t.Fatalf("ReadFile(config.yaml) error = %v", err)
	}
	if strings.Contains(string(data), "\nsource:") {
		t.Fatalf("config.yaml should not contain duplicate source field:\n%s", string(data))
	}
	if _, err := os.Stat(filepath.Join(configRoot, "devwiki", "huawei-zddi", "config.yaml")); err != nil {
		t.Fatalf("config.yaml not written under XDG devwiki dir: %v", err)
	}
}

func TestSaveAndLoadRepoConfigPreservesInactiveSources(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := RepoConfig{
		ProjectName:  "huawei-zddi",
		ProjectSlug:  "huawei-zddi",
		Language:     "zh",
		ActiveSource: RepoSourceRemote,
		Sources: RepoSources{
			Local:  &RepoSource{Type: RepoSourceLocal, Path: "/tmp/huawei-zddi-wiki"},
			Remote: &RepoSource{Type: RepoSourceRemote, URL: "http://devwiki.example.com:5697"},
		},
	}
	if err := SaveRepoConfig(cfg); err != nil {
		t.Fatalf("SaveRepoConfig() error = %v", err)
	}

	got, err := LoadRepoConfig("huawei-zddi")
	if err != nil {
		t.Fatalf("LoadRepoConfig() error = %v", err)
	}
	if got.ActiveSource != RepoSourceRemote {
		t.Fatalf("active source = %q", got.ActiveSource)
	}
	if got.Sources.Local == nil || got.Sources.Local.Path != "/tmp/huawei-zddi-wiki" {
		t.Fatalf("local source = %#v", got.Sources.Local)
	}
	if got.Sources.Remote == nil || got.Sources.Remote.URL != "http://devwiki.example.com:5697" {
		t.Fatalf("remote source = %#v", got.Sources.Remote)
	}
}

func TestListRepoProjectSlugsReturnsConfiguredProjects(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	for _, slug := range []string{"zeta", "alpha"} {
		if err := SaveRepoConfig(RepoConfig{
			ProjectName:  slug,
			ProjectSlug:  slug,
			Language:     "zh",
			ActiveSource: RepoSourceRemote,
			Sources: RepoSources{
				Remote: &RepoSource{
					Type: RepoSourceRemote,
					URL:  "http://devwiki.example.com:5697/" + slug,
				},
			},
		}); err != nil {
			t.Fatalf("SaveRepoConfig(%s) error = %v", slug, err)
		}
	}

	got, err := ListRepoProjectSlugs()
	if err != nil {
		t.Fatalf("ListRepoProjectSlugs() error = %v", err)
	}
	want := []string{"alpha", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("project slugs = %#v, want %#v", got, want)
	}
}

func TestListRepoProjectSlugsReturnsEmptyWhenConfigRootMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, err := ListRepoProjectSlugs()
	if err != nil {
		t.Fatalf("ListRepoProjectSlugs() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("project slugs = %#v, want empty", got)
	}
}

func TestSaveRepoConfigRejectsMismatchedSourceFields(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := SaveRepoConfig(RepoConfig{
		ProjectName:  "sample",
		ProjectSlug:  "sample",
		Language:     "zh",
		ActiveSource: RepoSourceLocal,
		Sources: RepoSources{
			Local: &RepoSource{
				Type: RepoSourceLocal,
				URL:  "http://devwiki.example.com:5697",
			},
		},
	})
	if err == nil {
		t.Fatal("SaveRepoConfig() error = nil, want local source without path rejected")
	}
}
