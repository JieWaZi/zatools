package skills

import (
	"path/filepath"
	"testing"
)

func TestWorkspaceProjectDetectionAndLockPaths(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "go.mod"), "module example\n")
	nested := filepath.Join(root, "nested", "deep")
	mustMkdirAll(t, nested)
	t.Setenv("HOME", filepath.Join(root, "home"))

	workspace := NewWorkspace(nested)
	if workspace.ProjectRoot != root {
		t.Fatalf("ProjectRoot = %q, want %q", workspace.ProjectRoot, root)
	}
	if workspace.ProjectDir() != root {
		t.Fatalf("ProjectDir = %q, want %q", workspace.ProjectDir(), root)
	}

	projectLock, err := workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath(project) error = %v", err)
	}
	if want := filepath.Join(root, LockFileName); projectLock != want {
		t.Fatalf("LockFilePath(project) = %q, want %q", projectLock, want)
	}

	globalLock, err := workspace.LockFilePath(true)
	if err != nil {
		t.Fatalf("LockFilePath(global) error = %v", err)
	}
	if want := filepath.Join(root, "home", ".agents", LockFileName); globalLock != want {
		t.Fatalf("LockFilePath(global) = %q, want %q", globalLock, want)
	}
}

func TestResolveProjectRootAndMarkers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	mustMkdirAll(t, nested)

	if hasProjectMarker(root) {
		t.Fatal("hasProjectMarker(root) = true before markers created")
	}

	mustWriteFile(t, filepath.Join(root, "package.json"), "{}")
	if !hasProjectMarker(root) {
		t.Fatal("hasProjectMarker(root) = false, want true")
	}

	if got := resolveProjectRoot(nested); got != root {
		t.Fatalf("resolveProjectRoot(nested) = %q, want %q", got, root)
	}
}
