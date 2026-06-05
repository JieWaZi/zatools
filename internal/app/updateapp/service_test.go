package updateapp

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUpdateDownloadsVerifiesAndReplacesExecutable(t *testing.T) {
	exePath := filepath.Join(t.TempDir(), "zatools")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	if err := os.WriteFile(exePath, []byte("old"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	archive := makeArchive(t, binaryName(runtime.GOOS), []byte("new"))
	sum := sha256.Sum256(archive)
	asset := fmt.Sprintf("zatools_v9.9.9_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	requests := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		switch r.URL.Path {
		case "/checksums.txt":
			fmt.Fprintf(w, "%x  ./%s\n", sum, asset)
		case "/" + asset:
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	service := NewService(ServiceOptions{
		BaseURL:        server.URL,
		ExecutablePath: exePath,
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
	})
	result, err := service.Update(context.Background())
	if err != nil {
		t.Fatalf("Update error = %v", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read updated executable: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("updated executable content = %q, want new", string(data))
	}
	if result.Asset != asset {
		t.Fatalf("result Asset = %q, want %q", result.Asset, asset)
	}
	if result.Path != exePath {
		t.Fatalf("result Path = %q, want %q", result.Path, exePath)
	}
	if strings.Join(requests, ",") != "/checksums.txt,/"+asset {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestUpdateRejectsChecksumMismatch(t *testing.T) {
	exePath := filepath.Join(t.TempDir(), "zatools")
	if err := os.WriteFile(exePath, []byte("old"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	archive := makeArchive(t, "zatools", []byte("new"))
	asset := "zatools_v9.9.9_linux_amd64.tar.gz"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/checksums.txt":
			fmt.Fprintf(w, "%064x  ./%s\n", 0, asset)
		case "/" + asset:
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	service := NewService(ServiceOptions{
		BaseURL:        server.URL,
		ExecutablePath: exePath,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if _, err := service.Update(context.Background()); err == nil || !strings.Contains(err.Error(), "checksum verification failed") {
		t.Fatalf("Update error = %v, want checksum failure", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read executable: %v", err)
	}
	if string(data) != "old" {
		t.Fatalf("executable changed after checksum failure: %q", string(data))
	}
}

func TestSelectAssetRejectsAmbiguousMatches(t *testing.T) {
	checksums := strings.Join([]string{
		"abc  ./zatools_v1_linux_amd64.tar.gz",
		"def  ./zatools_v2_linux_amd64.tar.gz",
	}, "\n")

	_, _, err := selectAsset(checksums, "linux", "amd64")
	if err == nil || !strings.Contains(err.Error(), "multiple matching release assets") {
		t.Fatalf("selectAsset error = %v, want ambiguous match error", err)
	}
}

func TestWindowsUpdateDefersReplacement(t *testing.T) {
	root := t.TempDir()
	exePath := filepath.Join(root, "zatools.exe")
	if err := os.WriteFile(exePath, []byte("old"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	archive := makeArchive(t, "zatools.exe", []byte("new"))
	sum := sha256.Sum256(archive)
	asset := "zatools_v9.9.9_windows_amd64.tar.gz"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/checksums.txt":
			fmt.Fprintf(w, "%x  ./%s\n", sum, asset)
		case "/" + asset:
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	var commands [][]string
	service := NewService(ServiceOptions{
		BaseURL:        server.URL,
		ExecutablePath: exePath,
		GOOS:           "windows",
		GOARCH:         "amd64",
		RunDetached: func(name string, args ...string) error {
			commands = append(commands, append([]string{name}, args...))
			return nil
		},
	})
	result, err := service.Update(context.Background())
	if err != nil {
		t.Fatalf("Update error = %v", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read executable: %v", err)
	}
	if string(data) != "old" {
		t.Fatalf("windows update should defer replacement while process is running, got %q", string(data))
	}
	if !result.Deferred {
		t.Fatal("result Deferred = false, want true")
	}
	if len(commands) != 1 || commands[0][0] != "cmd" {
		t.Fatalf("detached commands = %#v", commands)
	}
}

func makeArchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	header := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}
