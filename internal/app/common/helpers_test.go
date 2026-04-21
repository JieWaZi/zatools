package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitignoreEntryForProjectPathCollapsesTopLevelHiddenDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testCases := []struct {
		name string
		path string
		want string
	}{
		{name: "agents skills", path: filepath.Join(root, ".agents", "skills"), want: ".agents"},
		{name: "cursor rules", path: filepath.Join(root, ".cursor", "rules"), want: ".cursor"},
		{name: "cache dir", path: filepath.Join(root, ".cache"), want: ".cache"},
		{name: "lock file", path: filepath.Join(root, ".zatools-lock.json"), want: ".zatools-lock.json"},
		{name: "plain relative file", path: "devwiki-sample/README.md", want: "devwiki-sample/README.md"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := GitignoreEntryForProjectPath(root, tc.path)
			if err != nil {
				t.Fatalf("GitignoreEntryForProjectPath error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("GitignoreEntryForProjectPath = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEnsureProjectGitignoreCreatesFileAndAvoidsDuplicates(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gitignorePath := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("vendor/\n.agents/\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(.gitignore) error = %v", err)
	}

	err := EnsureProjectGitignore(
		root,
		filepath.Join(root, ".agents", "skills"),
		filepath.Join(root, ".cursor", "skills"),
		filepath.Join(root, ".cache"),
		filepath.Join(root, ".zatools-lock.json"),
	)
	if err != nil {
		t.Fatalf("EnsureProjectGitignore error = %v", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	content := string(data)

	for _, want := range []string{"vendor/", ".agents/", ".cursor", ".cache", ".zatools-lock.json"} {
		if !strings.Contains(content, want) {
			t.Fatalf(".gitignore missing %q:\n%s", want, content)
		}
	}
	if strings.Count(content, ".agents") != 1 {
		t.Fatalf(".gitignore should not duplicate .agents:\n%s", content)
	}
}
