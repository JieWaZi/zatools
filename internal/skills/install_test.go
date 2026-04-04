package skills

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadLockAndSaveLockRoundTrip(t *testing.T) {
	t.Parallel()

	lockPath := filepath.Join(t.TempDir(), "nested", LockFileName)
	empty, err := LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(missing) error = %v", err)
	}
	if len(empty.Skills) != 0 {
		t.Fatalf("LoadLock(missing) skills = %#v, want empty", empty.Skills)
	}

	want := LockFile{
		Skills: map[string]InstalledSkill{
			"demo": {Name: "demo", Description: "desc", Path: "/tmp/demo"},
		},
	}
	if err := SaveLock(lockPath, want); err != nil {
		t.Fatalf("SaveLock error = %v", err)
	}

	got, err := LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(saved) error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadLock(saved) = %#v, want %#v", got, want)
	}
}

func TestCopyDirHashDirInstallAndSync(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "src")
	mustMkdirAll(t, filepath.Join(src, "nested"))
	mustWriteFile(t, filepath.Join(src, "nested", "data.txt"), "hello")
	mustWriteFile(t, filepath.Join(src, "SKILL.md"), "---\nname: Demo Skill\ndescription: Demo\n---\n")

	dst := filepath.Join(t.TempDir(), "dst")
	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir error = %v", err)
	}

	gotHash, err := HashDir(dst)
	if err != nil {
		t.Fatalf("HashDir(dst) error = %v", err)
	}
	if gotHash == "" {
		t.Fatal("HashDir(dst) returned empty hash")
	}

	source := Source{Original: "owner/repo", Type: "github", Subpath: "skills"}
	skill := Skill{
		Name:        "Demo Skill",
		Description: "Demo",
		Dir:         src,
		RelativeDir: "nested-skill",
	}
	installRoot := filepath.Join(t.TempDir(), "install-root")
	entry, err := InstallSkill(installRoot, source, skill)
	if err != nil {
		t.Fatalf("InstallSkill error = %v", err)
	}

	if entry.Name != "Demo Skill" || entry.SourceSubdir != "skills/nested-skill" {
		t.Fatalf("InstallSkill entry = %#v", entry)
	}
	if _, err := os.Stat(filepath.Join(entry.Path, "nested", "data.txt")); err != nil {
		t.Fatalf("installed file missing: %v", err)
	}

	synced, err := SyncInstalledSkill(entry.Path, entry.Name, map[string]string{
		"cursor": filepath.Join(t.TempDir(), "cursor"),
		"claude": filepath.Join(t.TempDir(), "claude"),
	})
	if err != nil {
		t.Fatalf("SyncInstalledSkill error = %v", err)
	}
	if len(synced) != 2 {
		t.Fatalf("SyncInstalledSkill paths = %#v, want 2 entries", synced)
	}
	for _, path := range synced {
		if _, err := os.Stat(filepath.Join(path, "nested", "data.txt")); err != nil {
			t.Fatalf("synced file missing in %q: %v", path, err)
		}
	}
}

func TestHashDirChangesWhenContentsChange(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.txt"), "one")
	hashA, err := HashDir(dir)
	if err != nil {
		t.Fatalf("HashDir(first) error = %v", err)
	}

	mustWriteFile(t, filepath.Join(dir, "a.txt"), "two")
	hashB, err := HashDir(dir)
	if err != nil {
		t.Fatalf("HashDir(second) error = %v", err)
	}

	if hashA == hashB {
		t.Fatal("HashDir did not change after file content update")
	}
}

func TestSanitizeName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"My Skill":      "my-skill",
		"already-good":  "already-good",
		"..///":         "unnamed-skill",
		"Hello.World_1": "hello.world_1",
	}

	for input, want := range tests {
		if got := SanitizeName(input); got != want {
			t.Fatalf("SanitizeName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCopyFileCopiesContents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	mustWriteFile(t, src, "content")

	if err := copyFile(src, dst, 0o644); err != nil {
		t.Fatalf("copyFile error = %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile(dst) error = %v", err)
	}
	if got := string(data); got != "content" {
		t.Fatalf("copied content = %q, want %q", got, "content")
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func TestLoadLockRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	lockPath := filepath.Join(t.TempDir(), LockFileName)
	mustWriteFile(t, lockPath, "{bad json")

	_, err := LoadLock(lockPath)
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("LoadLock invalid JSON error = %v, want parse failure", err)
	}
}
